package component

import (
	"context"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// resourceType belongs to one runtime declaration, independent of the static
// ID that each importing, aliasing or instantiating scope uses to name it.
type resourceType struct {
	instance       *Instance
	host           *hostResourceDefinition
	identity       *canonical.ResourceType
	owner          *namespace
	representation wasm.ValueType
	destructor     *vm.Extern
}

func (in *Instance) bindResources(n *namespace, e item, t *validate.Type) error {
	if err := in.charge(1); err != nil {
		return err
	}
	if e.function != nil && (e.function.host != nil || e.function.asyncHost != nil) && e.function.scope == nil {
		// Imported host functions belong to their provider, independently of
		// the importing component's canonical failure/leave state. They share
		// its resolved nominal bindings, not its guest execution guards.
		if n.host == nil {
			n.host = &namespace{resources: n.resources}
		}
		e.function.scope = n.host
	}
	if e.sort == model.TypeSort && t.Kind == model.Resource {
		if e.resource == nil || n.resources[t.Resource] != nil && n.resources[t.Resource] != e.resource {
			return vm.ErrType
		}
		n.resources[t.Resource] = e.resource
	}
	if e.sort == model.InstanceSort {
		for _, ex := range t.Exports {
			if err := in.bindResources(n, e.instance[ex.Name.Text], ex.Type); err != nil {
				return err
			}
		}
	}
	return nil
}

func (in *Instance) resourceBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	rt := n.resources[plan.Type.Resource]
	if rt == nil {
		return vm.ErrType
	}
	in.resourceTable(n)
	fn, err := caller.NewFunction(plan.Signature, func(c *vm.Caller, args []vm.Value) ([]vm.Value, error) {
		if n.failed || n.blocked && plan.Definition.Op != 4 {
			return nil, wasm.ErrTrap
		}
		// The Core callback signature has already checked this scalar slot.
		bits, _, _ := args[0].Bits()
		switch plan.Definition.Op {
		case 2:
			id, err := n.handles.New(rt.identity, bits)
			if err != nil {
				return nil, err
			}
			return []vm.Value{vm.I32(id)}, nil
		case 3:
			err := n.handles.DropWith(c.Context(), rt.identity, uint32(bits), func(_ context.Context, rep uint64) error { return rt.destroy(c, rep, n) })
			return nil, err
		default:
			rep, err := n.handles.Rep(rt.identity, uint32(bits))
			if err != nil {
				return nil, err
			}
			return vmValues([]canonical.Slot{{Type: rt.representation, Bits: rep}}), nil
		}
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}

func (in *Instance) resourceTable(n *namespace) {
	if n.handles == nil {
		n.handles = canonical.NewResourceTable(in.resourceBudget)
		in.resourceTables = append(in.resourceTables, n.handles)
	}
}

func (rt *resourceType) destroy(c *vm.Caller, rep uint64, from *namespace) error {
	if rt.host != nil {
		return rt.destroyHost(c, rep)
	}
	if rt.destructor == nil {
		return nil
	}
	if rt.owner.failed || rt.owner.blocked || rt.owner.entered && rt.owner != from {
		return wasm.ErrTrap
	}
	if rt.owner != from {
		rt.owner.entered = true
		defer func() { rt.owner.entered = false }()
	}
	complete := false
	defer func() {
		if !complete {
			rt.owner.failed = true
		}
	}()
	_, err := c.Call(*rt.destructor, vmValues([]canonical.Slot{{Type: rt.representation, Bits: rep}}))
	complete = err == nil
	return err
}
