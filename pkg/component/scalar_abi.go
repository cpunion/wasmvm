package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// Types are bound once, while layout validation remains at its first use under
// the active codec policy. Mutable budgets and ownership stay in the frame.
type canonicalCallPlan struct {
	definition *validate.CanonicalFunction
	values     *functionValues
}

func newCanonicalCallPlan(plan *validate.CanonicalFunction) *canonicalCallPlan {
	bound := &canonicalCallPlan{definition: plan}
	if plan.Type != nil {
		bound.values = newFunctionValues(plan.Type)
	}
	return bound
}

func (bound *canonicalCallPlan) liftScalarArguments(codec *canonical.Codec, words []uint64, dst []Value) ([]Value, error) {
	plan := bound.definition
	if plan.ABI.ParameterMemory != nil {
		v, err := codec.Load(plan.ABI.ParameterMemory, words[0])
		return v.Elements, err
	}
	// A direct call has at most sixteen parameter slots and one result pointer.
	var slots [canonical.MaxFlatParams + 1]canonical.Slot
	if len(plan.Signature.Params) > len(slots) {
		return nil, vm.ErrType
	}
	for i, typ := range plan.Signature.Params {
		slots[i] = canonical.Slot{Type: typ, Bits: words[i]}
	}
	flat := slots[:len(plan.Signature.Params)]
	if dst == nil {
		dst = make([]Value, 0, min(len(plan.Type.Parameters), canonical.MaxFlatParams))
	}
	for i := range bound.values.parameters {
		parameter := &bound.values.parameters[i]
		count, err := parameter.FlatCount(codec)
		if err != nil {
			return nil, err
		}
		if count > uint64(len(flat)) {
			return nil, vm.ErrType
		}
		value, err := parameter.LiftFlat(codec, flat[:count])
		if err != nil {
			return nil, err
		}
		dst = append(dst, value)
		flat = flat[count:]
	}
	return dst, nil
}

func (bound *canonicalCallPlan) lowerScalarResult(codec *canonical.Codec, words []uint64, v *Value, scratch []canonical.Slot) error {
	plan := bound.definition
	if v == nil {
		if !plan.Async && len(plan.Signature.Results) != 0 {
			return vm.ErrType
		}
		return nil
	}
	if codec.Resources != nil {
		if err := codec.AcceptResult(plan.Type.Result, v); err != nil {
			return err
		}
	}
	if plan.ABI.ResultMemory != nil {
		return codec.Store(plan.Type.Result, words[len(plan.Signature.Params)-1], *v)
	}
	slots, err := bound.values.result.LowerFlatInto(codec, *v, scratch)
	if err != nil || plan.Async {
		return err
	}
	if len(slots) != len(plan.Signature.Results) {
		return vm.ErrType
	}
	for i, slot := range slots {
		if slot.Type != plan.Signature.Results[i] {
			return vm.ErrType
		}
		words[i] = slot.Bits
	}
	return nil
}
