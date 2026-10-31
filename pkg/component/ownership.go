package component

import (
	"context"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Resource is opaque. Copying its Go representation copies the capability,
// not ownership; a move or drop invalidates all copies of that capability.
type Resource = canonical.Reference

type resourceCodec struct {
	in      *Instance
	core    *vm.Caller
	n       *namespace
	scope   *canonical.BorrowScope
	created []*Resource
	seen    map[any]model.Kind
	borrows uint32
	lowered []ownedHandle
}

type ownedHandle struct {
	table *canonical.Resources
	rt    *resourceType
	id    uint32
	end   *asyncEnd
}

func (r *resourceCodec) declaration(t *model.Type) (*resourceType, error) {
	rt := r.n.resources[t.ResourceID]
	if rt == nil || rt.instance != r.in {
		return nil, vm.ErrType
	}
	if t.Kind == model.Borrow && rt.representation == wasm.I64 && rt.owner == r.n {
		return nil, unsupported("cm64 borrow conversion")
	}
	return rt, nil
}

func (r *resourceCodec) borrowScope() *canonical.BorrowScope {
	if r.scope == nil {
		r.scope = canonical.NewBorrowScope(r.in.config.MaxResources)
	}
	return r.scope
}

func (r *resourceCodec) inspect(t *model.Type, v Value) (*resourceType, *canonical.Resources, uint32, error) {
	rt, err := r.declaration(t)
	if err != nil {
		return nil, nil, 0, err
	}
	table, identity, id, owner, borrow, err := canonical.ReferenceInfo(v.Resource)
	if err != nil {
		return nil, nil, 0, err
	}
	if owner != rt || identity != rt.identity {
		return nil, nil, 0, vm.ErrType
	}
	if t.Kind == model.Own {
		if borrow || table != r.in.hostResources {
			return nil, nil, 0, vm.ErrType
		}
		err = table.CheckOwn(identity, id)
	} else {
		_, err = table.Rep(identity, id)
	}
	return rt, table, id, err
}

func (r *resourceCodec) Validate(t *model.Type, v Value) error {
	if asyncValueType(t) {
		_, err := r.asyncReference(t, v)
		return err
	}
	_, _, _, err := r.inspect(t, v)
	return err
}

func (r *resourceCodec) Lift(t *model.Type, id uint32) (Value, error) {
	if asyncValueType(t) {
		return r.liftAsync(t, id)
	}
	if t.Kind == model.ErrorContext {
		return r.n.handles.ErrorContext(id)
	}
	rt, err := r.declaration(t)
	if err != nil {
		return Value{}, err
	}
	if t.Kind == model.Borrow {
		_, err = r.borrowScope().Lift(r.n.handles, rt.identity, id)
		if err != nil {
			return Value{}, err
		}
		return Value{Resource: canonical.NewReference(r.n.handles, rt.identity, id, rt, r.scope)}, nil
	}
	rep, err := r.n.handles.Take(rt.identity, id)
	if err != nil {
		return Value{}, err
	}
	id, err = r.in.hostResources.New(rt.identity, rep)
	if err != nil {
		return Value{}, err
	}
	ref := canonical.NewReference(r.in.hostResources, rt.identity, id, rt, nil)
	r.created = append(r.created, ref)
	return Value{Resource: ref}, nil
}

func (r *resourceCodec) Lower(t *model.Type, v Value) (uint32, error) {
	if asyncValueType(t) {
		return r.lowerAsync(t, v)
	}
	if t.Kind == model.ErrorContext {
		id, err := r.n.handles.AddErrorContext(v)
		if err == nil {
			r.lowered = append(r.lowered, ownedHandle{table: r.n.handles, id: id})
		}
		return id, err
	}
	rt, table, id, err := r.inspect(t, v)
	if err != nil {
		return 0, err
	}
	if t.Kind == model.Borrow {
		rep, err := r.borrowScope().Lift(table, rt.identity, id)
		if err != nil {
			return 0, err
		}
		if rt.owner == r.n {
			return uint32(rep), nil
		}
		return r.scope.Lower(r.n.handles, rt.identity, rep)
	}
	rep, err := table.Take(rt.identity, id)
	if err != nil {
		return 0, err
	}
	canonical.InvalidateReference(v.Resource)
	id, err = r.n.handles.New(rt.identity, rep)
	if err == nil {
		r.lowered = append(r.lowered, ownedHandle{table: r.n.handles, rt: rt, id: id})
	}
	return id, err
}

func (r *resourceCodec) Accept(t *model.Type, v Value, commit bool) (Value, error) {
	if asyncValueType(t) {
		return r.acceptAsync(t, v, commit)
	}
	rt, table, id, err := r.inspect(t, v)
	if err != nil {
		return Value{}, err
	}
	if !commit {
		if t.Kind == model.Own {
			if r.seen == nil {
				r.seen = map[any]model.Kind{}
			}
			key := canonical.ReferenceKey(v.Resource)
			if r.seen[key] != 0 {
				return Value{}, vm.ErrType
			}
			r.seen[key] = model.Own
		} else {
			key := canonical.ReferenceKey(v.Resource)
			if r.seen == nil {
				r.seen = map[any]model.Kind{}
			}
			if r.seen[key] == model.Own {
				return Value{}, vm.ErrType
			}
			r.seen[key] = model.Borrow
			if r.borrows >= r.in.config.MaxResources {
				return Value{}, wasm.ErrLimit
			}
			r.borrows++
		}
		return v, nil
	}
	if t.Kind == model.Borrow {
		if _, err := r.borrowScope().Lift(table, rt.identity, id); err != nil {
			return Value{}, err
		}
		return Value{Resource: canonical.NewReference(table, rt.identity, id, rt, r.scope)}, nil
	}
	ref, err := canonical.MoveReference(v.Resource)
	if err != nil {
		return Value{}, err
	}
	r.created = append(r.created, ref)
	return Value{Resource: ref}, nil
}

func (r *resourceCodec) Finish(err error) error {
	if r.scope != nil {
		if ended := r.scope.End(); err == nil {
			err = ended
		}
	}
	if err != nil {
		for _, ref := range r.created {
			r.in.cleanupResource(r.core, ref)
		}
		for _, h := range r.lowered {
			if h.end != nil {
				if h.end.table() == h.table && h.end.id == h.id {
					h.end.abort()
				}
				continue
			}
			if h.rt == nil {
				h.table.DropErrorContext(h.id)
				continue
			}
			if rep, failed := h.table.Take(h.rt.identity, h.id); failed == nil {
				r.in.cleanupRepresentation(r.core, h.rt, rep)
			}
		}
	}
	if err == nil {
		for _, ref := range r.created {
			table, _, _, owner, _, failed := canonical.ReferenceInfo(ref)
			if end, ok := owner.(*asyncEnd); failed == nil && ok && table == r.in.hostResources && end.n == nil {
				end.unlink()
			}
		}
	}
	r.created = nil
	r.lowered = nil
	r.seen = nil
	return err
}

func (in *Instance) discardResource(ref *Resource) {
	table, identity, id, owner, borrow, err := canonical.ReferenceInfo(ref)
	if err != nil || borrow || table != in.hostResources {
		return
	}
	rt, ok := owner.(*resourceType)
	if !ok || rt.instance != in {
		return
	}
	if _, err := table.Take(identity, id); err == nil {
		canonical.InvalidateReference(ref)
	}
}

func (in *Instance) dropResource(c *vm.Caller, ref *Resource) error {
	table, identity, id, owner, borrow, err := canonical.ReferenceInfo(ref)
	if err != nil {
		return err
	}
	rt, ok := owner.(*resourceType)
	if !ok || rt.instance != in || table != in.hostResources || borrow {
		return vm.ErrReference
	}
	if err := table.CheckOwn(identity, id); err != nil {
		return err
	}
	return table.DropWith(c.Context(), identity, id, func(_ context.Context, rep uint64) error {
		canonical.InvalidateReference(ref)
		return rt.destroy(c, rep, nil)
	})
}

// DropResource consumes host ownership and invokes its defining destructor in
// the same Store. Use Caller.DropResource while a host callback is active.
func (in *Instance) DropResource(ctx context.Context, ref *Resource, options CallOptions) (uint64, error) {
	return in.store.WithCaller(ctx, options, func(c *vm.Caller) error { return in.dropResource(c, ref) })
}

func (c *Caller) DropResource(ref *Resource) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	return c.instance.dropResource(c.core, ref)
}

func finishResources(codec *canonical.Codec, complete bool, err error) error {
	if !complete && err == nil {
		return codec.Finish(wasm.ErrTrap)
	}
	return codec.Finish(err)
}

func returnBorrows(codec *canonical.Codec) error {
	r, ok := codec.Resources.(*resourceCodec)
	if !ok || r.scope == nil {
		return nil
	}
	if err := r.scope.CheckReturned(); err != nil {
		return err
	}
	if err := r.scope.End(); err != nil {
		return err
	}
	r.scope = nil
	return nil
}

// Cleanup always consumes unpublished ownership. Host destructors still get a
// cancellation context; failed/canceled guest code is never reentered.
func (in *Instance) cleanupResource(core *vm.Caller, ref *Resource) (err error) {
	table, identity, id, owner, borrow, err := canonical.ReferenceInfo(ref)
	if err != nil || borrow || table != in.hostResources {
		return nil
	}
	if end, ok := owner.(*asyncEnd); ok && end.in == in {
		end.abort()
		canonical.InvalidateReference(ref)
		return nil
	}
	rt, ok := owner.(*resourceType)
	if !ok || rt.instance != in {
		return nil
	}
	rep, err := table.Take(identity, id)
	if err != nil {
		return nil
	}
	canonical.InvalidateReference(ref)
	return in.cleanupRepresentation(core, rt, rep)
}

func (in *Instance) cleanupRepresentation(core *vm.Caller, rt *resourceType, rep uint64) (err error) {
	if core == nil {
		return nil
	}
	defer func() {
		if recover() != nil {
			err = wasm.ErrTrap
		}
	}()
	if rt.host != nil {
		return rt.destroyHost(core, rep)
	}
	if core.Context().Err() == nil && !rt.owner.failed && !rt.owner.blocked && !rt.owner.entered {
		return rt.destroy(core, rep, nil)
	}
	return nil
}
