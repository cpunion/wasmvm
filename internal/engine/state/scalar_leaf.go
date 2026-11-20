package state

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ScalarLeaf runs under the current Store lease without a Caller. A leaf may
// update its serialized host state. It must not reenter the VM, access guest
// memory or references, suspend, or invoke application callbacks.
type ScalarLeaf struct {
	call func(uint64, uint64) (uint64, error)
}

func NewScalarLeaf(call func(uint64, uint64) (uint64, error)) *ScalarLeaf {
	if call == nil {
		return nil
	}
	return &ScalarLeaf{call: call}
}

func (l *ScalarLeaf) Valid() bool { return l != nil && l.call != nil }

func narrowScalar(bits uint64, typ wasm.ValueType) uint64 {
	if typ == wasm.I32 || typ == wasm.F32 {
		return uint64(uint32(bits))
	}
	return bits
}

func (l *ScalarLeaf) Call(signature module.FuncType, a, b uint64) (value.Value, error) {
	if len(signature.Params) > 0 {
		a = narrowScalar(a, signature.Params[0])
	}
	if len(signature.Params) > 1 {
		b = narrowScalar(b, signature.Params[1])
	}
	bits, err := l.call(a, b)
	if err != nil || len(signature.Results) == 0 {
		return value.Value{}, err
	}
	return value.Value{Lo: narrowScalar(bits, signature.Results[0])}, nil
}

func (l *ScalarLeaf) CallSlots(signature module.FuncType, inputs, outputs []slots.Slot, values []value.Value) error {
	var a, b uint64
	if len(inputs) > 0 {
		a = values[inputs[0]].Lo
	}
	if len(inputs) > 1 {
		b = values[inputs[1]].Lo
	}
	v, err := l.Call(signature, a, b)
	if err == nil && len(outputs) != 0 {
		values[outputs[0]] = v
	}
	return err
}
