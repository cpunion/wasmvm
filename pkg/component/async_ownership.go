package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func asyncValueType(t *model.Type) bool { return t.Kind == model.Future || t.Kind == model.Stream }
func (e *asyncEnd) matches(t *model.Type, n *namespace) error {
	equal, err := validate.EqualValue(e.typ, t, e.in.config.Validate.MaxNodes, func(a, b uint64) bool {
		left, right := e.scope.resources[a], n.resources[b]
		return left != nil && right != nil && left.identity == right.identity
	})
	if err != nil {
		return err
	}
	if !equal {
		return vm.ErrType
	}
	return nil
}
func (r *resourceCodec) asyncReference(t *model.Type, v Value) (*asyncEnd, error) {
	table, _, id, owner, borrow, err := canonical.ReferenceInfo(v.Resource)
	if err != nil {
		return nil, err
	}
	e, ok := owner.(*asyncEnd)
	if !ok || e.in != r.in || table != r.in.hostResources || borrow || e.n != nil || e.id != id || !e.read || e.state != endIdle || e.waitable.set != nil {
		return nil, vm.ErrType
	}
	if stored, err := table.Opaque(id); err != nil || stored != e {
		return nil, wasm.ErrTrap
	}
	if err := e.matches(t, r.n); err != nil {
		return nil, err
	}
	return e, nil
}
func (r *resourceCodec) liftAsync(t *model.Type, id uint32) (Value, error) {
	e, err := r.in.end(r.n, id, t, true)
	if err != nil {
		return Value{}, err
	}
	if e.state != endIdle || e.waitable.set != nil {
		return Value{}, wasm.ErrTrap
	}
	next, err := r.n.handles.MoveOpaque(id, r.in.hostResources)
	if err != nil {
		return Value{}, err
	}
	e.n, e.id = nil, next
	ref := canonical.NewReference(r.in.hostResources, nil, next, e, nil)
	r.created = append(r.created, ref)
	return Value{Resource: ref}, nil
}
func (r *resourceCodec) lowerAsync(t *model.Type, v Value) (uint32, error) {
	e, err := r.asyncReference(t, v)
	if err != nil {
		return 0, err
	}
	r.in.resourceTable(r.n)
	next, err := r.in.hostResources.MoveOpaque(e.id, r.n.handles)
	if err != nil {
		return 0, err
	}
	e.n, e.id = r.n, next
	canonical.InvalidateReference(v.Resource)
	r.lowered = append(r.lowered, ownedHandle{table: r.n.handles, id: next, end: e})
	return next, nil
}
func (r *resourceCodec) acceptAsync(t *model.Type, v Value, commit bool) (Value, error) {
	if _, err := r.asyncReference(t, v); err != nil {
		return Value{}, err
	}
	if !commit {
		if r.seen == nil {
			r.seen = map[any]model.Kind{}
		}
		key := canonical.ReferenceKey(v.Resource)
		if r.seen[key] != 0 {
			return Value{}, vm.ErrType
		}
		r.seen[key] = model.Own
		return v, nil
	}
	ref, err := canonical.MoveReference(v.Resource)
	if err != nil {
		return Value{}, err
	}
	r.created = append(r.created, ref)
	return Value{Resource: ref}, nil
}
