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
	// ReadReady optionally supplies an already available typed batch. It must
	// return promptly. Its fresh, expiring Caller permits resource creation but
	// rejects guest reentry and blocking endpoint reads. False readiness must
	// return no data; the required Read/ReadOwned fallback then starts normally.
	// ReadReady and ReadBytes are mutually exclusive.
	ReadReady func(*Caller, uint32) (ReadResult, bool, error)
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
	if source.ReadBytes != nil && source.ReadReady != nil {
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
	owned := source.Owned
	n := uint64(len(owned))
	bytes := n * 128
	if n > uint64(in.config.MaxResources) || n > in.config.MaxValues-in.heldValues || bytes > in.config.MaxBytes-in.heldBytes {
		return Value{}, wasm.ErrLimit
	}
	if err := in.checkUniqueOwned(owned); err != nil {
		return Value{}, err
	}
	source.Owned = nil
	ends := new([2]asyncEnd)
	scope := &namespace{resources: t.resources, handles: in.hostResources}
	r, w, err := in.initEnds(ends, nil, scope, t.typ)
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
	if len(owned) != 0 {
		if len(owned) <= len(state.inlineOwned) {
			state.owned = state.inlineOwned[:0]
		} else {
			state.owned = make([]*Resource, 0, len(owned))
		}
		for _, ref := range owned {
			moved, err := in.references.Move(ref)
			if err != nil {
				in.mu.Unlock()
				r.abort()
				w.abort()
				return Value{}, err
			}
			state.owned = append(state.owned, moved)
		}
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

	return Value{Resource: in.references.New(in.hostResources, nil, r.id, r, nil)}, nil
}
func (c *Caller) NewSource(t ValueType, source AsyncSource) (Value, error) {
	v, _, err := c.newOwnedSource(t, source, false)
	return v, err
}

// NewOwnedSource also returns the source's moved capabilities for a provider
// that needs a private ownership anchor. Moving a returned reference onward
// removes it from the source's eventual cleanup. Use only in this Store.
func (c *Caller) NewOwnedSource(t ValueType, source AsyncSource) (v Value, owned []*Resource, err error) {
	return c.newOwnedSource(t, source, true)
}

func (c *Caller) newOwnedSource(t ValueType, source AsyncSource, returnOwned bool) (v Value, owned []*Resource, err error) {
	if err := c.enter(); err != nil {
		return Value{}, nil, err
	}
	defer c.mu.Unlock()
	defer c.instance.finishSourceCleanup(c.core, &err)
	v, err = c.instance.newSource(t, source)
	if err != nil {
		return v, nil, err
	}
	c.addCreated(v.Resource)
	if !returnOwned {
		return v, nil, nil
	}
	end, err := c.instance.readable(v.Resource)
	if err != nil {
		return Value{}, nil, err
	}
	if len(end.other.source.owned) == 0 {
		return v, nil, nil
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

func (in *Instance) tryImmediateSourceRead(c *vm.Caller, r, w *asyncEnd, b *guestBuffer, batch *model.Type) (bool, error) {
	if handled, err := in.readSourceBytes(c, r, w, b, batch); handled {
		return true, err
	}
	return in.readReadySource(c, r, w, b, batch)
}

func (in *Instance) startSourceRead(c *vm.Caller, r, w *asyncEnd, b *guestBuffer) error {
	batch, err := in.bufferType(r.typ.Element)
	if err != nil {
		return err
	}
	if handled, err := in.tryImmediateSourceRead(c, r, w, b, batch); handled || err != nil {
		return err
	}
	return in.startSourceJob(c, r, w, b, batch)
}

type sourceReadTask struct {
	task       componentTask
	function   Function
	host       resourceHostCall
	r, w       *asyncEnd
	b          *guestBuffer
	completion *ReadCompletion
	unit       bool
	aborted    bool
}

func (sr *sourceReadTask) start(caller *Caller) (*HostCompletion, error) {
	in := sr.task.in
	if in.callDepth >= in.config.Store.MaxReentry {
		return nil, wasm.ErrLimit
	}
	in.callDepth++
	defer func() { in.callDepth-- }()
	completion, err := sr.w.source.start(caller, sr.b.remain())
	if err != nil {
		return nil, err
	}
	if completion == nil || completion.completion == nil {
		return nil, vm.ErrType
	}
	completion.unit = sr.unit
	sr.completion = completion
	return completion.completion, nil
}

func (sr *sourceReadTask) cleanup() {
	if sr.aborted {
		return
	}
	sr.aborted = true
	if sr.completion != nil {
		sr.completion.result = ReadResult{}
	}
	sr.r.sourceJob = nil
	sr.r.buffer = nil
	sr.r.state = endDone
	sr.w.closeSource()
}

func (sr *sourceReadTask) resolve(current *vm.Caller, value *Value, cancelled bool) (err error) {
	in, r, w, b := sr.task.in, sr.r, sr.w, sr.b
	frame := in.acquireFrame()
	defer in.releaseCanonicalFrame(frame)
	var output *canonical.Codec
	complete := false
	defer func() {
		guard := cleanupGuard{panicked: recover()}
		if output != nil {
			err = guard.finishCodec(output, complete, err)
		}
		guard.finish()
	}()
	if cancelled {
		r.notify(0)
	} else {
		meta := sr.completion.result
		sr.completion.result = ReadResult{}
		count := uint64(len(value.Elements))
		if sr.unit {
			count = value.Bits
		} else if r.typ.Element.Kind == model.U8 && value.Elements == nil {
			count = uint64(len(value.Text))
		}
		if count > uint64(b.remain()) || meta.Count != 0 && uint64(meta.Count) != count || sr.unit && (len(meta.Values) != 0 || meta.Bytes != "") || r.typ.Kind == model.Future && count != 1 || count == 0 && !meta.Closed && b.remain() != 0 {
			return vm.ErrType
		}
		if !sr.unit && count != 0 {
			output, err = b.storeCodec(current, *value, frame.aux())
			if err != nil {
				if !r.rejectCopy(err) {
					return err
				}
				r.sourceJob = nil
				sr.aborted = true
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
	sr.aborted = true
	complete = true
	return nil
}

func (in *Instance) startSourceJob(c *vm.Caller, r, w *asyncEnd, b *guestBuffer, batch *model.Type) error {
	r.buffer = b
	plan := w.source.ensureReadPlan(batch)
	sr := in.acquireSourceReadTask()
	sr.function = Function{instance: in, scope: w.scope, typ: &plan.typ, values: &plan.values}
	sr.r, sr.w, sr.b, sr.unit = r, w, b, r.typ.Element == nil
	t := &sr.task
	t.readPool = sr
	if err := in.attachTask(t, w.scope, nil, vm.Extern{}, nil, nil); err != nil {
		t.readPool = nil
		in.recycleTask(nil, nil, sr)
		return err
	}
	t.hostFunction = &sr.function
	t.sourceRead = sr
	r.sourceJob = t
	if w.source.source.Prompt {
		var err error
		t.host, err = sr.function.hostParameters(c, nil, false, &sr.host)
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
	inlineOwned         [2]*Resource
	owned               []*Resource
	values, bytes       uint64
	cleanupNext         *sourceState
	writer              *asyncEnd
	queued, closeWriter bool
	controlAccepted     bool
	hasOwned            bool
	source              AsyncSource
	readPlan            *sourceReadPlan
	once                sync.Once
	closed              atomic.Bool
}

// One source has a fixed batch schema and resource namespace. Bind that
// immutable signature/layout policy once; each read still has a fresh Function,
// callback, codec, ownership transaction and cancellation scope.
type sourceReadPlan struct {
	typ    validate.Type
	values functionValues
}

func (in *Instance) sourceReadPlan(batch *model.Type) *sourceReadPlan {
	if plan := in.sourceReadPlans[batch]; plan != nil {
		return plan
	}
	plan := &sourceReadPlan{
		typ:    validate.Type{Kind: model.AsyncFunction, Result: batch},
		values: functionValues{result: canonical.NewValuePlan(batch)},
	}
	if in.sourceReadPlans == nil {
		in.sourceReadPlans = map[*model.Type]*sourceReadPlan{}
	}
	in.sourceReadPlans[batch] = plan
	return plan
}

func (s *sourceState) ensureReadPlan(batch *model.Type) *sourceReadPlan {
	if s.readPlan == nil {
		s.readPlan = s.in.sourceReadPlan(batch)
	}
	return s.readPlan
}

func (s *sourceState) readFunction(batch *model.Type) *Function {
	plan := s.ensureReadPlan(batch)
	return &Function{instance: s.in, scope: s.writer.scope, typ: &plan.typ, values: &plan.values}
}

func (s *sourceState) start(c *Caller, limit uint32) (*ReadCompletion, error) {
	if s.closed.Load() {
		return nil, wasm.ErrClosed
	}
	if s.source.ReadOwned != nil {
		var owned []*Resource
		if len(s.owned) != 0 {
			owned = append([]*Resource(nil), s.owned...)
		}
		return s.source.ReadOwned(c, owned, limit)
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
