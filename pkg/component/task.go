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
	runtime    *componentTask
}

func (in *Instance) taskBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	n.taskOps = true
	if n.task == nil && n.plan != nil {
		// Core start functions run in an initialization thread, before any
		// exported canonical call creates its own task and zeroed storage.
		n.task = &callTask{}
	}
	if plan.Definition.Op == 0x0c {
		return in.yieldBuiltin(caller, n, plan)
	}
	if plan.Definition.Op == 0x0a || plan.Definition.Op == 0x0b {
		return in.contextBuiltin(caller, n, plan)
	}
	fn, err := caller.NewScalarFunction(plan.Signature, func(c *vm.Caller, words []uint64) error {
		task := n.task
		if n.failed || task == nil {
			return wasm.ErrTrap
		}
		if n.blocked || task.plan == nil || !task.plan.Async || task.resolved || task.plan.Encoding != plan.Encoding {
			return wasm.ErrTrap
		}
		// A return without a memory option needs no result memory under static
		// validation. The official cross-ABI corpus and Wasmtime permit this
		// when the lift's memory was needed only for spilled parameters.
		if plan.Memory != nil && (task.plan.Memory == nil || n.items[model.CoreMemorySort][*plan.Memory].core != n.items[model.CoreMemorySort][*task.plan.Memory].core) {
			return wasm.ErrTrap
		}
		equal, err := in.matchValue(plan.Type.Result, task.plan.Type.Result, n, n)
		if err != nil {
			return err
		}
		if !equal {
			return wasm.ErrTrap
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
				ptr := words[0]
				value, err = codec.Load(plan.Type.Result, ptr)
			} else {
				var local [canonical.MaxFlatParams]canonical.Slot
				if len(plan.Signature.Params) > len(local) {
					return vm.ErrType
				}
				for i, typ := range plan.Signature.Params {
					local[i] = canonical.Slot{Type: typ, Bits: words[i]}
				}
				value, err = codec.LiftFlat(plan.Type.Result, local[:len(plan.Signature.Params)])
			}
			if err != nil {
				return err
			}
			task.result = &value
		}
		if err := returnBorrows(task.parameters); err != nil {
			return err
		}
		task.resolved = true
		if task.runtime != nil {
			return task.runtime.resolve(c, task.result, false)
		}
		return nil
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}

func (in *Instance) contextBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	if n.contextStorage == nil {
		storage, err := vm.NewScalarStorage(2, func() ([]uint64, bool) {
			if n.failed || n.task == nil {
				return nil, false
			}
			return n.task.storage[:], true
		})
		if err != nil {
			return err
		}
		n.contextStorage = storage
	}
	fn, err := caller.NewScalarStorageFunction(n.contextStorage, plan.Definition.Index, plan.Definition.CoreValue, plan.Definition.Op == 0x0b)
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}

func (in *Instance) yieldBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	fn, err := caller.NewSuspendableFunction(plan.Signature, func(c *vm.Caller, _ []vm.Value) (vm.HostResult, error) {
		if n.failed || n.blocked || n.task == nil {
			return vm.HostResult{}, wasm.ErrTrap
		}
		t := n.task.runtime
		if t == nil {
			return vm.HostResult{Values: []vm.Value{vm.I32(0)}}, nil
		}
		return vm.HostResult{Pending: &vm.HostSuspension{Ready: func() bool { return true }, Resume: func(*vm.Caller) (vm.HostResult, error) {
			return vm.HostResult{Values: []vm.Value{vm.I32(0)}}, nil
		}}}, nil
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}
