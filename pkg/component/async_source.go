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
	Read  func(*Caller, uint32) (*ReadCompletion, error)
	Close func()
}

// ReadCompletion shares the one-shot, thread-safe HostCompletion protocol.
// Resolve's function runs inside the Store with a fresh Caller.
type ReadCompletion struct {
	completion *HostCompletion
	result     ReadResult
	unit       bool
}

func NewReadCompletion() *ReadCompletion { return &ReadCompletion{completion: NewHostCompletion()} }
func (r *ReadCompletion) Resolve(fn func(*Caller) (ReadResult, error)) error {
	if r == nil || r.completion == nil || fn == nil {
		return vm.ErrType
	}
	return r.completion.Resolve(func(c *Caller) (*Value, error) {
		value, err := fn(c)
		if err != nil {
			return nil, err
		}
		r.result = value
		if r.unit {
			return &Value{Bits: uint64(value.Count)}, nil
		}
		if value.Values != nil && value.Bytes != "" {
			return nil, vm.ErrType
		}
		return &Value{Elements: value.Values, Text: value.Bytes}, nil
	})
}
func (r *ReadCompletion) Cancelled() error {
	if r == nil || r.completion == nil {
		return vm.ErrType
	}
	return r.completion.Cancelled()
}

func (in *Instance) newSource(t ValueType, source AsyncSource) (Value, error) {
	if !t.Valid() || !asyncValueType(t.typ) || source.Read == nil {
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
	scope := &namespace{resources: t.resources, handles: in.hostResources}
	r, w, err := in.allocateEnds(nil, scope, t.typ)
	if err != nil {
		return Value{}, err
	}
	state := &sourceState{source: source}
	in.mu.Lock()
	if in.closed {
		in.mu.Unlock()
		r.abort()
		w.abort()
		return Value{}, wasm.ErrClosed
	}
	w.source = state
	if in.sources == nil {
		in.sources = map[*sourceState]struct{}{}
	}
	in.sources[state] = struct{}{}
	in.mu.Unlock()
	return Value{Resource: canonical.NewReference(in.hostResources, nil, r.id, r, nil)}, nil
}
func (c *Caller) NewSource(t ValueType, source AsyncSource) (Value, error) {
	if err := c.enter(); err != nil {
		return Value{}, err
	}
	defer c.mu.Unlock()
	v, err := c.instance.newSource(t, source)
	if err == nil {
		c.created = append(c.created, v.Resource)
	}
	return v, err
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
		if r.sourceJob != nil && r.sourceJob.host != nil && r.sourceJob.host.cancel != nil {
			r.sourceJob.host.cancel()
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
				output, err = b.store(current, *value)
				if err != nil {
					return err
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
	source AsyncSource
	once   sync.Once
	closed atomic.Bool
}

func (s *sourceState) start(c *Caller, limit uint32) (*ReadCompletion, error) {
	if s.closed.Load() {
		return nil, wasm.ErrClosed
	}
	return s.source.Read(c, limit)
}
func (s *sourceState) close() (panicked any) {
	s.once.Do(func() {
		s.closed.Store(true)
		defer func() { panicked = recover() }()
		if s.source.Close != nil {
			s.source.Close()
		}
	})
	return panicked
}
