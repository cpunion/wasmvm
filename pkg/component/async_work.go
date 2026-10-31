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
	Owned  []*Resource
	Start  func(*Caller, []*Resource) (*HostCompletion, error)
	Done   func(*Caller, bool) error
	Cancel *Cancellation
}
type workState struct {
	control       *Cancellation
	refs          []*Resource
	values, bytes uint64
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
func (c *Caller) Schedule(work AsyncWork) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	in := c.instance
	if work.Start == nil {
		return vm.ErrType
	}
	n := uint64(len(work.Owned))
	bytes := n * 128
	if n > uint64(in.config.MaxResources) || n > in.config.MaxValues-in.heldValues || bytes > in.config.MaxBytes-in.heldBytes {
		return wasm.ErrLimit
	}
	seen := map[any]bool{}
	for _, ref := range work.Owned {
		if err := in.checkOwned(ref); err != nil {
			return err
		}
		key := canonical.ReferenceKey(ref)
		if seen[key] {
			return vm.ErrType
		}
		seen[key] = true
	}
	scope := &namespace{handles: in.hostResources}
	f := &Function{instance: in, scope: scope, typ: &validate.Type{Kind: model.AsyncFunction}}
	t, err := in.makeTask(c.core, scope, nil, vm.Extern{}, nil, nil)
	if err != nil {
		return err
	}
	state := &workState{control: work.Cancel}
	if state.control == nil {
		state.control = NewCancellation()
	}
	if err := state.control.attach(in.scheduler().wake); err != nil {
		t.finish(c.core, err)
		return err
	}
	t.work = state
	state.values, state.bytes = n, bytes
	in.heldValues += n
	in.heldBytes += bytes
	state.refs = make([]*Resource, len(work.Owned))
	for i, ref := range work.Owned {
		state.refs[i], err = canonical.MoveReference(ref)
		if err != nil {
			t.finish(c.core, err)
			return err
		}
	}
	t.hostFunction = f
	t.host, err = f.hostCall(c.core, nil)
	if err != nil {
		t.finish(c.core, err)
		return err
	}
	f.asyncHost = func(caller *Caller, _ []Value) (*HostCompletion, error) {
		return work.Start(caller, append([]*Resource(nil), state.refs...))
	}
	t.onResolve = func(core *vm.Caller, _ *Value, cancelled bool) error {
		if work.Done == nil {
			return nil
		}
		_, err := t.host.invoke(core, func(fresh *Caller) (*Value, error) { return nil, work.Done(fresh, cancelled) })
		return err
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
	guard := cleanupGuard{}
	for _, ref := range refs {
		guard.run(func() { in.cleanupResource(c, ref) })
	}
	guard.finish()
}
