package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"sync"
	"sync/atomic"
)

// Cancellation is a one-job, goroutine-safe cancellation signal. It may be
// requested before scheduling, including from a source's nonblocking Close.
type Cancellation struct {
	mu        sync.Mutex
	requested atomic.Bool
	used      bool
	cancel    func()
	wake      chan<- struct{}
}

func NewCancellation() *Cancellation { return &Cancellation{} }
func (s *Cancellation) Cancel() {
	if s == nil {
		return
	}
	s.requested.Store(true)
	s.mu.Lock()
	cancel, wake := s.cancel, s.wake
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if wake != nil {
		select {
		case wake <- struct{}{}:
		default:
		}
	}
}
func (s *Cancellation) attach(wake chan<- struct{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.used {
		return vm.ErrType
	}
	s.used = true
	s.wake = wake
	return nil
}
func (s *Cancellation) bind(cancel func()) {
	s.mu.Lock()
	s.cancel = cancel
	requested := s.requested.Load()
	s.mu.Unlock()
	if requested {
		cancel()
	}
}
func (s *Cancellation) release() { s.mu.Lock(); s.cancel = nil; s.wake = nil; s.mu.Unlock() }

// AsyncWork starts after the current guest turn parks or returns. It shares
// task/fuel/cancellation state with the component scheduler. Start resolves a
// void HostCompletion; Done runs with a fresh Caller on success/cancellation.
// Owned capabilities move at scheduling and are discarded when work finishes
// unless Start/Done moved or dropped them. Borrowed capabilities are rejected.
type AsyncWork struct {
	// Prompt starts without waiting for a guest yield. Start and Done must return
	// promptly; their Callers reject guest reentry and blocking endpoint reads.
	Prompt bool
	Owned  []*Resource
	Start  func(*Caller, []*Resource) (*HostCompletion, error)
	Done   func(*Caller, bool) error
	Cancel *Cancellation
	// Cleanup runs once after accepted work finishes or is abandoned, including
	// before Start. It must be nonblocking and must not enter the VM.
	Cleanup func()
}
type workState struct {
	prompt        bool
	control       *Cancellation
	refs          []*Resource
	start         func(*Caller, []*Resource) (*HostCompletion, error)
	done          func(*Caller, bool) error
	values, bytes uint64
	cleanup       func()
}

type scheduledTask struct {
	task       componentTask
	scope      namespace
	function   Function
	work       workState
	control    Cancellation
	host       resourceHostCall
	inlineRefs [2]*Resource
}

func (s *workState) startWork(caller *Caller) (*HostCompletion, error) {
	var args []*Resource
	if len(s.refs) != 0 {
		args = append([]*Resource(nil), s.refs...)
	}
	return s.start(caller, args)
}

func (in *Instance) checkUniqueOwned(refs []*Resource) error {
	if len(refs) <= 8 {
		var seen [8]any
		for i, ref := range refs {
			if err := in.checkOwned(ref); err != nil {
				return err
			}
			key := canonical.ReferenceKey(ref)
			for j := range i {
				if seen[j] == key {
					return vm.ErrType
				}
			}
			seen[i] = key
		}
		return nil
	}
	seen := make(map[any]bool, len(refs))
	for _, ref := range refs {
		if err := in.checkOwned(ref); err != nil {
			return err
		}
		key := canonical.ReferenceKey(ref)
		if seen[key] {
			return vm.ErrType
		}
		seen[key] = true
	}
	return nil
}

func (in *Instance) checkOwned(ref *Resource) error {
	table, typ, id, owner, borrow, err := canonical.ReferenceInfo(ref)
	if err != nil {
		return err
	}
	if borrow || table != in.hostResources {
		return vm.ErrReference
	}
	switch x := owner.(type) {
	case *resourceType:
		if x.instance != in || typ != x.identity {
			return vm.ErrReference
		}
		return table.CheckOwn(typ, id)
	case *asyncEnd:
		if x.in != in || x.n != nil || x.id != id || !x.read || x.state != endIdle && x.state != endDone {
			return vm.ErrReference
		}
		if v, err := table.Opaque(id); err != nil || v != x {
			return vm.ErrReference
		}
		return nil
	}
	return vm.ErrReference
}

// Scheduled work has no canonical parameters or result. These private descriptors
// are immutable and carry no instance, ownership or callback state.
var (
	scheduledWorkType   = validate.Type{Kind: model.AsyncFunction}
	scheduledWorkValues = functionValues{}
)

func (c *Caller) Schedule(work AsyncWork) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	in := c.instance
	if err := in.drainPromptWork(c.core); err != nil {
		return err
	}
	if work.Start == nil {
		return vm.ErrType
	}
	n := uint64(len(work.Owned))
	bytes := n * 128
	if n > uint64(in.config.MaxResources) || n > in.config.MaxValues-in.heldValues || bytes > in.config.MaxBytes-in.heldBytes {
		return wasm.ErrLimit
	}
	if err := in.checkUniqueOwned(work.Owned); err != nil {
		return err
	}
	alloc := in.acquireScheduledTask()
	alloc.scope = namespace{handles: in.hostResources}
	alloc.work = workState{control: work.Cancel, prompt: work.Prompt, start: work.Start, done: work.Done}
	alloc.function = Function{instance: in, scope: &alloc.scope, typ: &scheduledWorkType, values: &scheduledWorkValues}
	t := &alloc.task
	t.schedPool = alloc
	if err := in.attachTask(t, &alloc.scope, nil, vm.Extern{}, nil, nil); err != nil {
		t.schedPool = nil
		in.recycleTask(nil, alloc, nil)
		return err
	}
	state := &alloc.work
	if state.control == nil {
		state.control = &alloc.control
	}
	if err := state.control.attach(in.scheduler().wake); err != nil {
		t.finish(c.core, err)
		return err
	}
	t.work = state
	state.values, state.bytes = n, bytes
	in.heldValues += n
	in.heldBytes += bytes
	if len(work.Owned) <= len(alloc.inlineRefs) {
		state.refs = alloc.inlineRefs[:len(work.Owned)]
	} else {
		state.refs = make([]*Resource, len(work.Owned))
	}
	for i, ref := range work.Owned {
		moved, err := in.references.Move(ref)
		if err != nil {
			t.finish(c.core, err)
			return err
		}
		state.refs[i] = moved
	}
	f := &alloc.function
	t.hostFunction = f
	host, err := f.hostParameters(c.core, nil, false, &alloc.host)
	t.host = host
	if err != nil {
		t.finish(c.core, err)
		return err
	}
	t.host.prompt = work.Prompt
	state.cleanup = work.Cleanup
	if work.Prompt {
		return t.resume(c.core)
	}
	return nil
}

// Retire ready prompt host work before admitting more work. The bounded sweep
// never starts ordinary deferred work, waits for I/O, or resumes a guest task.
func (in *Instance) drainPromptWork(c *vm.Caller) error {
	if in.drainingPrompt || in.tasks == nil {
		return nil
	}
	in.drainingPrompt = true
	defer func() { in.drainingPrompt = false }()
	limit := in.tasks.live
	for range limit {
		var ready *componentTask
		for t := in.tasks.head; t != nil; t = t.next {
			if t.work != nil && t.work.prompt && t.started && t.ready() {
				ready = t
				break
			}
		}
		if ready == nil {
			break
		}
		if err := ready.resume(c); err != nil {
			return err
		}
	}
	return nil
}
func (t *componentTask) syncCancellation() {
	if t.work != nil && t.work.control.requested.Load() {
		t.cancelRequested = true
		if t.host != nil && t.host.cancel != nil {
			t.host.cancel()
		}
	}
}
func (s *workState) finish(c *vm.Caller, in *Instance) {
	defer func() {
		in.heldValues -= s.values
		in.heldBytes -= s.bytes
		s.values, s.bytes = 0, 0
		s.control.release()
	}()
	refs := s.refs
	s.refs = nil
	s.start, s.done = nil, nil
	guard := cleanupGuard{}
	for _, ref := range refs {
		guard.run(func() { in.cleanupResource(c, ref) })
	}
	clear(refs)
	cleanup := s.cleanup
	s.cleanup = nil
	if cleanup != nil {
		guard.run(cleanup)
	}
	guard.finish()
}
