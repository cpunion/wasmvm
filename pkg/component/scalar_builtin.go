package component

import (
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Both edges share one builtin implementation. Only operations which can
// actually park need public VM values and the existing suspension protocol.
type scalarOperands struct {
	words  []uint64
	values []vm.Value
}

func (a scalarOperands) at(index int) uint64 {
	if a.words != nil {
		return a.words[index]
	}
	bits, _, _ := a.values[index].Bits()
	return bits
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
		return c.NewSuspendableFunction(signature, func(c *vm.Caller, values []vm.Value) (vm.HostResult, error) {
			step, err := run(c, scalarOperands{values: values})
			return scalarHostResult(signature, step), err
		})
	}
	return c.NewScalarFunction(signature, func(c *vm.Caller, words []uint64) error {
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
