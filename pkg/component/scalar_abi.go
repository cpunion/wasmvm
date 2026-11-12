package component

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// Core canonical signatures already describe scalar carriers. Retain canonical
// validation and ownership, while avoiding public VM Value boxing at this edge.
func liftScalarArguments(codec *canonical.Codec, plan *validate.CanonicalFunction, words []uint64) ([]Value, error) {
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
	return liftFlatArguments(codec, plan, slots[:len(plan.Signature.Params)])
}

func lowerScalarResult(codec *canonical.Codec, plan *validate.CanonicalFunction, words []uint64, v *Value) error {
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
	slots, err := codec.LowerFlat(plan.Type.Result, *v)
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
