package component

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (n *namespace) addScalarLeaf(c *vm.Caller, signature vm.FunctionType, run func(uint64, uint64) (uint64, error)) error {
	fn, err := c.NewScalarLeafFunction(signature, state.NewScalarLeaf(run))
	if err == nil {
		n.add(item{sort: model.CoreFunctionSort, core: fn})
	}
	return err
}

// Both edges borrow scalar carriers and share one builtin implementation.
// Parked results use the existing owned continuation protocol.
type scalarOperands struct {
	words []uint64
}

func (a scalarOperands) at(index int) uint64 {
	return a.words[index]
}

type scalarStep struct {
	bits    uint64
	pending *vm.HostSuspension
}

func newScalarBuiltin(c *vm.Caller, signature vm.FunctionType, suspends bool, run func(*vm.Caller, scalarOperands) (scalarStep, error)) (vm.Extern, error) {
	if run == nil || len(signature.Results) > 1 {
		return vm.Extern{}, vm.ErrType
	}
	if suspends {
		return c.NewBorrowedSuspendableScalarFunction(signature, func(c *vm.Caller, words []uint64) (*vm.HostSuspension, error) {
			step, err := run(c, scalarOperands{words: words})
			if err == nil && step.pending == nil && len(signature.Results) != 0 {
				words[0] = step.bits
			}
			return step.pending, err
		})
	}
	return c.NewBorrowedScalarFunction(signature, func(c *vm.Caller, words []uint64) error {
		step, err := run(c, scalarOperands{words: words})
		if err != nil {
			return err
		}
		if step.pending != nil {
			return vm.ErrType
		}
		if len(signature.Results) != 0 {
			words[0] = step.bits
		}
		return nil
	})
}

func scalarHostResult(signature vm.FunctionType, step scalarStep) vm.HostResult {
	if step.pending != nil {
		return vm.HostResult{Pending: step.pending}
	}
	if len(signature.Results) == 0 {
		return vm.HostResult{}
	}
	var result vm.Value
	switch signature.Results[0] {
	case wasm.I32:
		result = vm.I32(uint32(step.bits))
	case wasm.I64:
		result = vm.I64(step.bits)
	case wasm.F32:
		result = vm.F32Bits(uint32(step.bits))
	case wasm.F64:
		result = vm.F64Bits(step.bits)
	}
	return vm.HostResult{Values: []vm.Value{result}}
}
