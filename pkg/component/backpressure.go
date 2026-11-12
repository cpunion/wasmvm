package component

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (in *Instance) backpressureBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	fn, err := caller.NewScalarFunction(plan.Signature, func(*vm.Caller, []uint64) error {
		if n.failed {
			return wasm.ErrTrap
		}
		if plan.Definition.Op == 0x24 {
			n.backpressure++
			if n.backpressure == 1<<16 {
				return wasm.ErrTrap
			}
		} else {
			if n.backpressure == 0 {
				return wasm.ErrTrap
			}
			n.backpressure--
		}
		return nil
	})
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}
