package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// AwaitRead starts a native endpoint read without retaining the callback stack.
// The returned completion invokes consume with a fresh Caller when the read or
// operation context completes. consume owns the returned payload. The caller
// must keep the endpoint alive until completion, for example with AsyncWork.Owned.
func (c *Caller) AwaitRead(ctx context.Context, ref *Resource, limit uint32, consume func(*Caller, ReadResult, error) (*Value, error)) (*HostCompletion, error) {
	return c.awaitRead(ctx, ref, limit, nil, consume)
}

// AwaitReadChecked validates each byte batch before acknowledging its writer.
// A rejected batch closes the reader without consuming those bytes; consume
// receives the validation error. check must be prompt and must not enter the VM.
func (c *Caller) AwaitReadChecked(ctx context.Context, ref *Resource, limit uint32, check func(ReadResult) error, consume func(*Caller, ReadResult, error) (*Value, error)) (*HostCompletion, error) {
	if check == nil {
		return nil, vm.ErrType
	}
	return c.awaitRead(ctx, ref, limit, check, consume)
}
func (c *Caller) awaitRead(ctx context.Context, ref *Resource, limit uint32, check func(ReadResult) error, consume func(*Caller, ReadResult, error) (*Value, error)) (*HostCompletion, error) {
	if err := c.enter(); err != nil {
		return nil, err
	}
	defer c.mu.Unlock()
	if consume == nil {
		return nil, vm.ErrType
	}
	if ctx == nil {
		ctx = c.Context()
	}
	parent := c.Context()
	in := c.instance
	e, err := in.readable(ref)
	if err != nil {
		return nil, err
	}
	if check != nil && (e.typ.Kind != KindStream || e.typ.Element == nil || e.typ.Element.Kind != KindU8) {
		return nil, vm.ErrType
	}
	if e.state != endIdle || limit >= 1<<28 || e.typ.Kind == KindFuture && limit != 1 {
		return nil, wasm.ErrTrap
	}
	batch, err := in.bufferType(e.typ.Element)
	if err != nil {
		return nil, err
	}
	state := &pendingHostRead{in: in, end: e, ctx: ctx, parent: parent, consume: consume}
	b := in.initHostReadBuffer(&state.storage, e, batch, limit, check, nil)
	state.buffer = b
	e.waitable.syncWaiter = true
	e.waitable.hostRead = state
	if err = e.copy(c.core, b); err != nil {
		state.abort(c.core)
		return nil, err
	}
	c.addRead(state)
	if e.waitable.pending {
		// The resolver checks both cancellation scopes before accepting this
		// already available event. No cancellation watcher is needed to wake it.
		state.resolve()
		return &state.completion, nil
	}
	fn := state.resolve
	state.stopOperation = context.AfterFunc(ctx, fn)
	if parent != ctx {
		state.stopParent = context.AfterFunc(parent, fn)
	}
	return &state.completion, nil
}

// AwaitSourceRead forwards an owned readable endpoint through an AsyncSource
// without retaining a callback stack. transform runs with a fresh Caller and
// owns the read payload, under the same rules as AwaitRead's consume callback.
func (c *Caller) AwaitSourceRead(ctx context.Context, ref *Resource, limit uint32, transform func(*Caller, ReadResult, error) (ReadResult, error)) (*ReadCompletion, error) {
	if transform == nil {
		return nil, vm.ErrType
	}
	result := &ReadCompletion{}
	completion, err := c.AwaitRead(ctx, ref, limit, func(fresh *Caller, value ReadResult, failed error) (*Value, error) {
		value, failed = transform(fresh, value, failed)
		if failed != nil {
			return nil, failed
		}
		return result.encode(value)
	})
	if err != nil {
		return nil, err
	}
	result.completion = completion
	return result, nil
}

type pendingHostRead struct {
	in            *Instance
	end           *asyncEnd
	buffer        *guestBuffer
	storage       hostReadBuffer
	completion    HostCompletion
	ctx, parent   context.Context
	consume       func(*Caller, ReadResult, error) (*Value, error)
	finished      bool
	stopOperation func() bool
	stopParent    func() bool
}

func (r *pendingHostRead) resolve() {
	_ = r.completion.Resolve(r.complete)
}

func (r *pendingHostRead) complete(fresh *Caller) (*Value, error) {
	failed := r.ctx.Err()
	if failed == nil {
		failed = r.parent.Err()
	}
	if failed == nil {
		failed = fresh.Context().Err()
	}
	result, err := r.take(fresh.core, failed)
	if err == nil {
		canonical.VisitReferences(r.buffer.batch, &r.buffer.data, func(ref *Resource) { fresh.addCreated(ref) })
	}
	return r.consume(fresh, result, err)
}

func (r *pendingHostRead) detach() {
	r.finished = true
	r.end.waitable.syncWaiter = false
	r.end.waitable.hostRead = nil
	if r.stopOperation != nil {
		r.stopOperation()
		r.stopOperation = nil
	}
	if r.stopParent != nil {
		r.stopParent()
		r.stopParent = nil
	}
}
func (r *pendingHostRead) take(c *vm.Caller, failed error) (ReadResult, error) {
	if r.finished {
		return ReadResult{}, wasm.ErrClosed
	}
	if failed != nil {
		r.abort(c)
		return ReadResult{}, failed
	}
	if !r.end.waitable.pending {
		return ReadResult{}, wasm.ErrTrap
	}
	event, err := r.end.waitable.poll(c)
	if err != nil {
		r.abort(c)
		return ReadResult{}, err
	}
	r.detach()
	b := r.buffer
	if b.rejected != nil {
		return ReadResult{}, b.rejected
	}
	return ReadResult{Values: b.data.Elements, Bytes: b.data.Text, Count: b.progress, Closed: r.end.typ.Kind == KindFuture || event[2]&15 == 1}, nil
}
func (r *pendingHostRead) abort(c *vm.Caller) {
	if r.finished {
		return
	}
	r.detach()
	e, b := r.end, r.buffer
	guard := cleanupGuard{}
	if job := e.sourceJob; job != nil {
		guard.run(func() { job.finish(c, wasm.ErrTrap) })
	}
	if e.waitable.pending {
		guard.run(func() { e.waitable.poll(c) })
	} else if e.state != endDone {
		e.state = endIdle
	}
	e.buffer = nil
	canonical.VisitReferences(b.batch, &b.data, func(ref *Resource) { guard.run(func() { r.in.cleanupResource(c, ref) }) })
	guard.finish()
}
