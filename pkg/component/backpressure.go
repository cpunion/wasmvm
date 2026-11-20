package component

import (
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (in *Instance) backpressureBuiltin(caller *vm.Caller, n *namespace, plan *validate.CanonicalFunction) error {
	return n.addScalarLeaf(caller, plan.Signature, func(_, _ uint64) (uint64, error) {
		if n.failed {
			return 0, wasm.ErrTrap
		}
		if plan.Definition.Op == 0x24 {
			n.backpressure++
			if n.backpressure == 1<<16 {
				return 0, wasm.ErrTrap
			}
		} else {
			if n.backpressure == 0 {
				return 0, wasm.ErrTrap
			}
			n.backpressure--
		}
		return 0, nil
	})
}
