package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"sync"
	"sync/atomic"
)

// ReadResult supplies one bounded batch. Bytes is the packed stream<u8> form;
// Values carries other typed elements. Unit endpoints use Count without data.
// Closed reports final data/EOF; a future must produce exactly one element.
type ReadResult struct {
	Values []Value
	Bytes  string
	Count  uint32
	Closed bool
}

// AsyncSource starts one read after the initiating guest turn has parked.
// Read may return an unresolved completion; its Caller expires on return.
// Close runs once after EOF, failure, reader drop or instance close. It must
// be nonblocking and must not enter or wait for the VM.
type AsyncSource struct {
	// Prompt permits an immediate readiness probe. Read must return promptly,
	// and its Caller rejects guest reentry and blocking endpoint reads.
	Prompt bool
	Read   func(*Caller, uint32) (*ReadCompletion, error)
	// ReadBytes optionally probes packed stream<u8> data without an asynchronous
	// invocation. It must not block or enter the VM. A false readiness result
	// supplies no data and delegates to Read/ReadOwned, which remains required.
	// Ready batches obey the same count, budget, copy and close rules as Read.
	ReadBytes func(uint32) (ReadResult, bool, error)
	// Owned capabilities move on successful construction. ReadOwned receives
	// their moved references; source closure reclaims them in the Store session.
	Owned     []*Resource
	Cancel    *Cancellation
	ReadOwned func(*Caller, []*Resource, uint32) (*ReadCompletion, error)
	Close     func()
}

// ReadCompletion shares the one-shot, thread-safe HostCompletion protocol.
// Resolve's function runs inside the Store with a fresh Caller.
type ReadCompletion struct {
	completion *HostCompletion
	result     ReadResult
	unit       bool
}

func NewReadCompletion() *ReadCompletion { return &ReadCompletion{completion: NewHostCompletion()} }

// NewReadCompletionWithCleanup releases external batch leases after completion,
// cancellation or abandonment, under the HostCompletion cleanup contract.
func NewReadCompletionWithCleanup(cleanup func()) *ReadCompletion {
	return &ReadCompletion{completion: NewHostCompletionWithCleanup(cleanup)}
}

func (r *ReadCompletion) Resolve(fn func(*Caller) (ReadResult, error)) error {
	if r == nil || r.completion == nil || fn == nil {
		return vm.ErrType
	}
	return r.completion.Resolve(func(c *Caller) (*Value, error) {
		value, err := fn(c)
		if err != nil {
			return nil, err
		}
		return r.encode(value)
	})
}
func (r *ReadCompletion) encode(value ReadResult) (*Value, error) {
	r.result = value
	if r.unit {
		return &Value{Bits: uint64(value.Count)}, nil
	}
	if value.Values != nil && value.Bytes != "" {
		return nil, vm.ErrType
	}
	return &Value{Elements: value.Values, Text: value.Bytes}, nil
}
func (r *ReadCompletion) Cancelled() error {
	if r == nil || r.completion == nil {
		return vm.ErrType
	}
	return r.completion.Cancelled()
}

func (in *Instance) newSource(t ValueType, source AsyncSource) (Value, error) {
	if !t.Valid() || !asyncValueType(t.typ) || (source.Read == nil) == (source.ReadOwned == nil) {
		return Value{}, vm.ErrType
	}
	if source.ReadBytes != nil && (t.typ.Kind != model.Stream || t.typ.Element == nil || t.typ.Element.Kind != model.U8) {
		return Value{}, vm.ErrType
	}
	for _, rt := range t.resources {
		if rt == nil || rt.instance != in {
			return Value{}, vm.ErrReference
		}
	}
	if _, err := in.layouts[0].Build(t.typ); err != nil {
		return Value{}, err
	}
	owned := append([]*Resource(nil), source.Owned...)
	n := uint64(len(owned))
	bytes := n * 128
	if n > uint64(in.config.MaxResources) || n > in.config.MaxValues-in.heldValues || bytes > in.config.MaxBytes-in.heldBytes {
		return Value{}, wasm.ErrLimit
	}
	seen := map[any]bool{}
	for _, ref := range owned {
		if err := in.checkOwned(ref); err != nil {
			return Value{}, err
		}
		key := canonical.ReferenceKey(ref)
		if seen[key] {
			return Value{}, vm.ErrType
		}
		seen[key] = true
	}
	source.Owned = nil
	scope := &namespace{resources: t.resources, handles: in.hostResources}
	r, w, err := in.allocateEnds(nil, scope, t.typ)
	if err != nil {
		return Value{}, err
	}
	state := &sourceState{source: source, in: in, values: n, bytes: bytes, writer: w, hasOwned: n != 0}
	if source.Cancel != nil {
		if err := source.Cancel.attach(in.scheduler().wake); err != nil {
			r.abort()
			w.abort()
			return Value{}, err
		}
		state.controlAccepted = true
	}
	in.mu.Lock()
	if in.closed {
		in.mu.Unlock()
		if state.controlAccepted {
			source.Cancel.release()
		}
		r.abort()
		w.abort()
		return Value{}, wasm.ErrClosed
	}
	w.source = state
	in.heldValues += n
	in.heldBytes += bytes
	for _, ref := range owned {
		moved, err := canonical.MoveReference(ref)
		if err != nil {
			in.mu.Unlock()
			r.abort()
			w.abort()
			return Value{}, err
		}
		state.owned = append(state.owned, moved)
	}
	if in.sources == nil {
		in.sources = map[*sourceState]struct{}{}
	}
	in.sources[state] = struct{}{}
	in.mu.Unlock()
	if state.controlAccepted {
		source.Cancel.mu.Lock()
		requested := false
		if !state.closed.Load() {
			source.Cancel.cancel = func() { state.queue(true) }
			requested = source.Cancel.requested.Load()
		}
		source.Cancel.mu.Unlock()
		if requested {
			state.queue(true)
		}
	}

	return Value{Resource: canonical.NewReference(in.hostResources, nil, r.id, r, nil)}, nil
}
func (c *Caller) NewSource(t ValueType, source AsyncSource) (Value, error) {
	v, _, err := c.NewOwnedSource(t, source)
	return v, err
}

// NewOwnedSource also returns the source's moved capabilities for a provider
// that needs a private ownership anchor. Moving a returned reference onward
// removes it from the source's eventual cleanup. Use only in this Store.
func (c *Caller) NewOwnedSource(t ValueType, source AsyncSource) (v Value, owned []*Resource, err error) {
	if err := c.enter(); err != nil {
		return Value{}, nil, err
	}
	defer c.mu.Unlock()
	defer c.instance.finishSourceCleanup(c.core, &err)
	v, err = c.instance.newSource(t, source)
	if err != nil {
		return v, nil, err
	}
	c.created = append(c.created, v.Resource)
	end, err := c.instance.readable(v.Resource)
	if err != nil {
		return Value{}, nil, err
	}
	return v, append([]*Resource(nil), end.other.source.owned...), nil
}
func (in *Instance) NewSource(ctx context.Context, t ValueType, source AsyncSource) (v Value, err error) {
	_, err = in.store.WithCallerCleanup(ctx, CallOptions{}, func(c *vm.Caller) (err error) {
		complete := false
		defer in.finishTasks(c, &err, &complete)
		v, err = in.newSource(t, source)
		complete = err == nil
		return err
	}, func(c *vm.Caller) error { return in.cleanupResource(c, v.Resource) })
	return v, err
}

func (w *asyncEnd) closeSource() {
	if w.source == nil {
		return
	}
	source := w.source
	w.source = nil
	w.in.mu.Lock()
	delete(w.in.sources, source)
	w.in.mu.Unlock()
	w.unlink()
	w.table().RemoveOpaque(w.id)
	w.state = endDone
	w.buffer = nil
	if r := w.other; r != nil {
		if job := r.sourceJob; job != nil {
			job.cancelRequested = true
			if job.host != nil && job.host.cancel != nil {
				job.host.cancel()
			}
		}
		r.other, w.other = nil, nil
		if r.state != endDone && !r.waitable.pending {
			r.notify(0)
		}
	}
	if p := source.close(); p != nil {
		panic(p)
	}
}

func (in *Instance) startSourceRead(c *vm.Caller, r, w *asyncEnd, b *guestBuffer) error {
	batch, err := in.bufferType(r.typ.Element)
	if err != nil {
		return err
	}
	if handled, err := in.readSourceBytes(c, r, w, b, batch); handled {
		return err
	}
	unit := r.typ.Element == nil
	var completion *ReadCompletion
	f := &Function{instance: in, scope: w.scope, typ: &validate.Type{Kind: model.AsyncFunction, Result: batch}}
	f.asyncHost = func(caller *Caller, _ []Value) (*HostCompletion, error) {
		var err error
		if in.callDepth >= in.config.Store.MaxReentry {
			return nil, wasm.ErrLimit
		}
		in.callDepth++
		defer func() { in.callDepth-- }()
		completion, err = w.source.start(caller, b.remain())
		if err != nil {
			return nil, err
		}
		if completion == nil || completion.completion == nil {
			return nil, vm.ErrType
		}
		completion.unit = unit
		return completion.completion, nil
	}
	t, err := in.makeTask(c, w.scope, nil, vm.Extern{}, func(*vm.Caller) ([]Value, error) { return nil, nil }, nil)
	if err != nil {
		return err
	}
	t.hostFunction = f
	r.sourceJob = t
	t.pendingCleanup = func() {
		if completion != nil {
			completion.result = ReadResult{}
		}
		r.sourceJob = nil
		r.buffer = nil
		r.state = endDone
		w.closeSource()
	}
	t.onResolve = func(current *vm.Caller, value *Value, cancelled bool) (err error) {
		var outputState canonical.Codec
		var output *canonical.Codec
		complete := false
		defer func() {
			guard := cleanupGuard{panicked: recover()}
			if output != nil {
				guard.run(func() { err = finishResources(output, complete, err) })
			}
			guard.finish()
		}()
		if cancelled {
			r.notify(0)
		} else {
			meta := completion.result
			completion.result = ReadResult{}
			count := uint64(len(value.Elements))
			if unit {
				count = value.Bits
			} else if r.typ.Element.Kind == model.U8 && value.Elements == nil {
				count = uint64(len(value.Text))
			}
			if count > uint64(b.remain()) || meta.Count != 0 && uint64(meta.Count) != count || unit && (len(meta.Values) != 0 || meta.Bytes != "") || r.typ.Kind == model.Future && count != 1 || count == 0 && !meta.Closed && b.remain() != 0 {
				return vm.ErrType
			}
			if !unit && count != 0 {
				output, err = b.storeCodec(current, *value, &outputState)
				if err != nil {
					if !r.rejectCopy(err) {
						return err
					}
					r.sourceJob = nil
					t.pendingCleanup = nil
					complete = true
					return nil
				}
			}
			b.progress += uint32(count)
			r.notify(b.progress)
			if meta.Closed || r.typ.Kind == model.Future {
				w.closeSource()
			}
		}
		r.sourceJob = nil
		t.pendingCleanup = nil
		complete = true
		return nil
	}
	if w.source.source.Prompt {
		t.host, err = f.hostCall(c, nil)
		if err != nil {
			return t.finish(c, err)
		}
		t.host.prompt = true
		return t.resume(c)
	}
	return nil
}

func (in *Instance) bufferType(element *model.Type) (*model.Type, error) {
	if t := in.bufferTypes[element]; t != nil {
		return t, nil
	}
	t := &model.Type{Kind: model.List, Element: element}
	if element == nil {
		t = &model.Type{Kind: model.U32}
	}
	if _, err := in.layouts[0].Build(t); err != nil {
		return nil, err
	}
	if in.bufferTypes == nil {
		in.bufferTypes = map[*model.Type]*model.Type{}
	}
	in.bufferTypes[element] = t
	return t, nil
}

type sourceState struct {
	in                  *Instance
	owned               []*Resource
	values, bytes       uint64
	cleanupNext         *sourceState
	writer              *asyncEnd
	queued, closeWriter bool
	controlAccepted     bool
	hasOwned            bool
	source              AsyncSource
	once                sync.Once
	closed              atomic.Bool
}

func (s *sourceState) start(c *Caller, limit uint32) (*ReadCompletion, error) {
	if s.closed.Load() {
		return nil, wasm.ErrClosed
	}
	if s.source.ReadOwned != nil {
		return s.source.ReadOwned(c, append([]*Resource(nil), s.owned...), limit)
	}
	return s.source.Read(c, limit)
}
func (s *sourceState) close() (panicked any) {
	s.once.Do(func() {
		s.closed.Store(true)
		if s.controlAccepted {
			s.source.Cancel.release()
		}
		defer func() {
			panicked = recover()
			if s.hasOwned {
				s.queue(false)
			}
		}()
		if s.source.Close != nil {
			s.source.Close()
		}
	})
	return panicked
}

func (s *sourceState) queue(closeWriter bool) {
	s.in.mu.Lock()
	if closeWriter && s.closed.Load() {
		s.in.mu.Unlock()
		return
	}
	if closeWriter {
		s.closeWriter = true
	}
	if !s.queued {
		s.queued = true
		s.cleanupNext = s.in.sourceCleanup
		s.in.sourceCleanup = s
	}
	s.in.mu.Unlock()
	if s.source.Cancel != nil {
		s.source.Cancel.mu.Lock()
		wake := s.source.Cancel.wake
		s.source.Cancel.mu.Unlock()
		if wake != nil {
			select {
			case wake <- struct{}{}:
			default:
			}
		}
	}
}
