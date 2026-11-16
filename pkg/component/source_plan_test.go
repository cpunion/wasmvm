package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSourceReadPlansKeepIndependentValuesAndAggregateQuotas(t *testing.T) {
	in := asyncInstance(t, "interpreter", "pending-callback", Imports{})
	in.config.MaxValues = 3
	source := &sourceState{in: in, writer: &asyncEnd{scope: &namespace{}}}
	batch := &model.Type{Kind: model.List, Element: &model.Type{Kind: model.U32}}
	first, second := source.readFunction(batch), source.readFunction(batch)
	if first == second || first.typ != second.typ || first.values != second.values {
		t.Fatal("source reads did not share only their type and value plans")
	}
	_, err := in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
		a, err := first.hostCall(core, nil)
		if err != nil {
			return err
		}
		b, err := second.hostCall(core, nil)
		if err != nil {
			return err
		}
		value := Value{Elements: []Value{{Bits: 42}}}
		one, err := a.accept(&value)
		if err != nil || in.heldValues != 2 {
			t.Fatal(one, err, in.heldValues)
		}
		if _, err := b.accept(&value); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("shared plan lost the outstanding read's quota", err)
		}
		if err := a.finish(core, true, nil); err != nil {
			return err
		}
		c, err := source.readFunction(batch).hostCall(core, nil)
		if err != nil {
			return err
		}
		two, err := c.accept(&value)
		if err != nil {
			return err
		}
		two.Elements[0].Bits = 99
		if one.Elements[0].Bits != 42 || value.Elements[0].Bits != 42 {
			t.Fatal("source plan cached mutable result storage")
		}
		if err := b.finish(core, false, wasm.ErrLimit); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("failed transaction did not preserve its error", err)
		}
		return c.finish(core, true, nil)
	})
	if err != nil || in.heldValues != 0 || in.heldBytes != 0 {
		t.Fatal("source plan retained transaction accounting", err, in.heldValues, in.heldBytes)
	}
}
