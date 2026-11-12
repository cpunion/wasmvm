package component

import (
	"context"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (in *Instance) readable(ref *Resource) (*asyncEnd, error) {
	table, _, id, owner, borrow, err := canonical.ReferenceInfo(ref)
	if err != nil {
		return nil, err
	}
	e, ok := owner.(*asyncEnd)
	if !ok || e.in != in || borrow || table != in.hostResources || e.n != nil || e.id != id || !e.read {
		return nil, vm.ErrReference
	}
	if stored, err := table.Opaque(id); err != nil || stored != e {
		return nil, wasm.ErrTrap
	}
	return e, nil
}

// ReadAsync waits for one bounded batch from a host-owned readable endpoint,
// driving the existing component tasks in the same Store lease. Futures require
// limit=1; unit endpoints return Count without allocating Values.
func (in *Instance) ReadAsync(ctx context.Context, ref *Resource, limit uint32, options CallOptions) (result ReadResult, err error) {
	return in.readAsyncCall(ctx, ref, limit, options, nil)
}

// ByteDestination supplies storage for a ReadAsyncInto batch after its actual
// size is known. Repeated calls request the cumulative size and must preserve
// the already written prefix. The callback runs under the Store lease and must
// not reenter the VM. It and its storage are not retained after the read returns.
type ByteDestination func(uint32) ([]byte, error)

// ReadAsyncInto reads a stream<u8> batch into destination's caller-owned buffer.
// Count and Closed describe the result; Bytes and Values are empty. No buffer is
// requested for an empty batch. On failure previously supplied storage may have
// changed. Logical copy limits still apply.
func (in *Instance) ReadAsyncInto(ctx context.Context, ref *Resource, limit uint32, destination ByteDestination, options CallOptions) (ReadResult, error) {
	if destination == nil {
		return ReadResult{}, vm.ErrType
	}
	return in.readAsyncCall(ctx, ref, limit, options, destination)
}

func (in *Instance) readAsyncCall(ctx context.Context, ref *Resource, limit uint32, options CallOptions, into ByteDestination) (result ReadResult, err error) {
	var buffer *guestBuffer
	_, err = in.store.WithCallerCleanup(ctx, options, func(c *vm.Caller) (err error) {
		complete := false
		defer in.finishTasks(c, &err, &complete)
		result, buffer, err = in.readAsync(c, ref, limit, c.Context(), into)
		complete = err == nil
		return err
	}, func(c *vm.Caller) error {
		if buffer != nil {
			canonical.VisitReferences(buffer.batch, &buffer.data, func(r *Resource) { in.cleanupResource(c, r) })
		}
		return nil
	})
	if err != nil {
		return ReadResult{}, err
	}
	return result, nil
}
func (c *Caller) ReadAsync(ref *Resource, limit uint32) (ReadResult, error) {
	return c.ReadAsyncContext(nil, ref, limit)
}

// ReadAsyncContext adds a cancellable operation context without extending the
// callback lifetime. A nil context uses the callback context.
func (c *Caller) ReadAsyncContext(ctx context.Context, ref *Resource, limit uint32) (ReadResult, error) {
	return c.readAsyncContext(ctx, ref, limit, nil)
}

// ReadAsyncInto uses the callback's Store lease to read stream<u8> into storage
// supplied by destination. A nil context uses the callback context. It shares
// the size, ownership and cancellation contracts of Instance.ReadAsyncInto.
func (c *Caller) ReadAsyncInto(ctx context.Context, ref *Resource, limit uint32, destination ByteDestination) (ReadResult, error) {
	if destination == nil {
		return ReadResult{}, vm.ErrType
	}
	return c.readAsyncContext(ctx, ref, limit, destination)
}

func (c *Caller) readAsyncContext(ctx context.Context, ref *Resource, limit uint32, into ByteDestination) (ReadResult, error) {
	if err := c.enter(); err != nil {
		return ReadResult{}, err
	}
	defer c.mu.Unlock()
	if c.prompt {
		return ReadResult{}, vm.ErrType
	}
	parent := c.Context()
	if ctx == nil {
		ctx, parent = parent, nil
	}
	result, b, err := c.instance.readAsyncContexts(c.core, ref, limit, ctx, parent, into)
	// Both contexts are checked synchronously at publication. No asynchronous
	// cancellation bridge is needed: the scheduler waits on both Done channels.
	if err == nil {
		err = ctx.Err()
	}
	if err == nil && parent != nil {
		err = parent.Err()
	}
	if b != nil {
		canonical.VisitReferences(b.batch, &b.data, func(ref *Resource) {
			if err != nil {
				c.instance.cleanupResource(c.core, ref)
			} else {
				c.created = append(c.created, ref)
			}
		})
	}
	if err != nil {
		return ReadResult{}, err
	}
	return result, nil
}
func (in *Instance) readAsync(c *vm.Caller, ref *Resource, limit uint32, ctx context.Context, into ByteDestination) (result ReadResult, b *guestBuffer, err error) {
	return in.readAsyncContexts(c, ref, limit, ctx, nil, into)
}

func (in *Instance) readAsyncContexts(c *vm.Caller, ref *Resource, limit uint32, ctx, parent context.Context, into ByteDestination) (result ReadResult, b *guestBuffer, err error) {
	e, err := in.readable(ref)
	if err != nil {
		return result, nil, err
	}
	if into != nil && (e.typ.Kind != KindStream || e.typ.Element == nil || e.typ.Element.Kind != KindU8) {
		return result, nil, vm.ErrType
	}
	if e.state != endIdle || limit >= 1<<28 || e.typ.Kind == KindFuture && limit != 1 {
		return result, nil, wasm.ErrTrap
	}
	batch, err := in.bufferType(e.typ.Element)
	if err != nil {
		return result, nil, err
	}
	b = in.hostReadBuffer(e, batch, limit, nil, into)
	defer func() {
		if b.hooks != nil {
			b.hooks.into = nil
		}
	}()
	complete := false
	e.waitable.syncWaiter = true
	defer func() {
		guard := cleanupGuard{panicked: recover()}
		e.waitable.syncWaiter = false
		if !complete {
			if e.sourceJob != nil {
				job := e.sourceJob
				failed := err
				if failed == nil {
					failed = wasm.ErrTrap
				}
				guard.run(func() { job.finish(c, failed) })
			}
			if e.waitable.pending {
				guard.run(func() { e.waitable.poll(c) })
			} else if e.state != endDone {
				e.state = endIdle
			}
			e.buffer = nil
			canonical.VisitReferences(b.batch, &b.data, func(ref *Resource) { guard.run(func() { in.cleanupResource(c, ref) }) })
		}
		guard.finish()
	}()

	if err := e.copy(c, b); err != nil {
		return result, b, err
	}
	if err := in.awaitTasksContexts(c, ctx, parent, func() bool { return e.waitable.pending }); err != nil {
		return result, b, err
	}
	event, err := e.waitable.poll(c)
	if err != nil {
		return result, b, err
	}
	result = ReadResult{Values: b.data.Elements, Bytes: b.data.Text, Count: b.progress, Closed: event[2]&15 == 1}
	if e.typ.Kind == KindFuture {
		result.Closed = true
	}
	complete = true
	return result, b, nil
}
