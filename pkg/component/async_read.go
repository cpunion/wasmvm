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
	var buffer *guestBuffer
	_, err = in.store.WithCallerCleanup(ctx, options, func(c *vm.Caller) (err error) {
		complete := false
		defer in.finishTasks(c, &err, &complete)
		result, buffer, err = in.readAsync(c, ref, limit, c.Context())
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
	if err := c.enter(); err != nil {
		return ReadResult{}, err
	}
	defer c.mu.Unlock()
	result, b, err := c.instance.readAsync(c.core, ref, limit, c.Context())
	if err == nil {
		err = c.Context().Err()
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
func (in *Instance) readAsync(c *vm.Caller, ref *Resource, limit uint32, ctx context.Context) (result ReadResult, b *guestBuffer, err error) {
	e, err := in.readable(ref)
	if err != nil {
		return result, nil, err
	}
	if e.state != endIdle || limit >= 1<<28 || e.typ.Kind == KindFuture && limit != 1 {
		return result, nil, wasm.ErrTrap
	}
	batch, err := in.bufferType(e.typ.Element)
	if err != nil {
		return result, nil, err
	}
	b = &guestBuffer{in: in, scope: e.scope, typ: e.typ.Element, batch: batch, length: limit}
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
	if err := in.awaitTasksContext(c, ctx, func() bool { return e.waitable.pending }); err != nil {
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
