package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// One implicit thread belongs to each active canonical lift. The existing
// Store session serializes execution; nested stackful calls save this state.
type callTask struct {
	plan       *validate.CanonicalFunction
	parameters *canonical.Codec
	returned   *canonical.Codec
	result     *Value
	resolved   bool
	storage    [2]uint64
}

func (in *Instance) taskBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	n.taskOps = true
	if n.task == nil && n.plan != nil {
		// Core start functions run in an initialization thread, before any
		// exported canonical call creates its own task and zeroed storage.
		n.task = &callTask{}
	}
	fn, err := caller.NewFunction(plan.Signature, func(c *vm.Caller, args []vm.Value) ([]vm.Value, error) {
		task := n.task
		if n.failed || task == nil {
			return nil, wasm.ErrTrap
		}
		switch plan.Definition.Op {
		case 0x0a:
			bits := task.storage[plan.Definition.Index]
			if plan.Definition.CoreValue == wasm.I64 {
				return []vm.Value{vm.I64(bits)}, nil
			}
			return []vm.Value{vm.I32(uint32(bits))}, nil
		case 0x0b:
			bits, _, _ := args[0].Bits()
			task.storage[plan.Definition.Index] = bits
			return nil, nil
		case 0x0c:
			if n.blocked {
				return nil, wasm.ErrTrap
			}
			// With no other ready threads, the pinned nondeterministic ABI
			// permits yielding directly back to the current implicit thread.
			return []vm.Value{vm.I32(0)}, nil
		}
		if n.blocked || task.plan == nil || !task.plan.Async || task.resolved || task.plan.Encoding != plan.Encoding {
			return nil, wasm.ErrTrap
		}
		if (task.plan.Memory == nil) != (plan.Memory == nil) || plan.Memory != nil && n.items[model.CoreMemorySort][*plan.Memory].core != n.items[model.CoreMemorySort][*task.plan.Memory].core {
			return nil, wasm.ErrTrap
		}
		equal, err := validate.EqualValue(plan.Type.Result, task.plan.Type.Result, in.config.Validate.MaxNodes)
		if err != nil {
			return nil, err
		}
		if !equal {
			return nil, wasm.ErrTrap
		}
		codec := in.codec(c, n, plan)
		if r, ok := codec.Resources.(*resourceCodec); ok {
			// Cleanup happens after this builtin's Caller expires. Retain the
			// outer canonical call's still-active session capability instead.
			r.core = task.parameters.Resources.(*resourceCodec).core
		}
		task.returned = codec
		if plan.Type.Result != nil {
			var value Value
			if plan.ABI.ParameterMemory != nil {
				ptr, _, _ := args[0].Bits()
				value, err = codec.Load(plan.Type.Result, ptr)
			} else {
				var slots []canonical.Slot
				slots, err = flatValues(args)
				if err == nil {
					value, err = codec.LiftFlat(plan.Type.Result, slots)
				}
			}
			if err != nil {
				return nil, err
			}
			task.result = &value
		}
		if r, ok := task.parameters.Resources.(*resourceCodec); ok && r.scope != nil {
			if err := r.scope.CheckReturned(); err != nil {
				return nil, err
			}
			if err := r.scope.End(); err != nil {
				return nil, err
			}
		}
		task.resolved = true
		return nil, nil
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}

func (in *Instance) asyncLift(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction, results []vm.Value) (*Value, error) {
	if plan.Callback != nil {
		callback := n.items[model.CoreFunctionSort][*plan.Callback].core
		for {
			packed, _, _ := results[0].Bits()
			switch packed & 0xf {
			case 0:
				goto exited
			case 1:
				// There are no suspended producers in this execution profile.
				// The current thread is ready and may be selected immediately.
			default:
				// WAIT requires a valid waitable-set handle; that producer
				// capability is rejected by preflight in this profile.
				return nil, wasm.ErrTrap
			}
			called, err := caller.Call(callback, []vm.Value{vm.I32(0), vm.I32(0), vm.I32(0)})
			if err != nil {
				return nil, err
			}
			results = called.Values
		}
	}
exited:
	if !n.task.resolved {
		return nil, wasm.ErrTrap
	}
	return n.task.result, nil
}
