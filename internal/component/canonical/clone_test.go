package canonical

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostCopiesPreserveOnlyDeclaredShape(t *testing.T) {
	list := &model.Type{Kind: model.List, Element: typ(model.String)}
	mapType := &model.Type{Kind: model.Map, Key: typ(model.U8), Element: list}
	for _, tc := range []struct {
		typ   *model.Type
		value Value
	}{
		{typ(model.U8), Value{Bits: 42}},
		{typ(model.String), Value{Text: "hello"}},
		{&model.Type{Kind: model.Enum, Labels: []string{"a"}}, Value{}},
		{&model.Type{Kind: model.Flags, Labels: []string{"a"}}, Value{Bits: 1}},
		{list, Value{Elements: []Value{{Text: "hello"}}}},
		{&model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: 1}, Value{Elements: []Value{{Bits: 42}}}},
		{&model.Type{Kind: model.Record, Fields: []model.Field{{Name: "text", Type: list}}}, Value{Elements: []Value{{Elements: []Value{{Text: "hello"}}}}}},
		{&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: list}}}, Value{Elements: []Value{{Elements: []Value{{Text: "hello"}}}}}},
		{&model.Type{Kind: model.Option, Element: list}, Value{}},
		{&model.Type{Kind: model.Option, Element: list}, Value{Case: 1, Elements: []Value{{Elements: []Value{{Text: "hello"}}}}}},
		{mapType, Value{Elements: []Value{{Elements: []Value{{Bits: 42}, {Elements: []Value{{Text: "hello"}}}}}}}},
	} {
		c := valueCodec(false)
		out, err := c.Clone(tc.typ, tc.value)
		if err != nil {
			t.Fatal(tc.typ.Kind, err)
		}
		if len(out.Elements) != 0 {
			out.Elements[0].Text = "changed"
			if tc.value.Elements[0].Text == "changed" {
				t.Fatal("copy retained aggregate slice", tc.typ.Kind)
			}
		}
	}
	// Irrelevant host fields may contain cycles. A scalar copy must strip them.
	cycle := Value{Bits: 42, Elements: make([]Value, 1)}
	cycle.Elements[0] = cycle
	copy, err := valueCodec(false).Clone(typ(model.U8), cycle)
	if err != nil || copy.Bits != 42 || copy.Elements != nil {
		t.Fatal(copy, err)
	}
}

func TestHostCopiesEnforceTypesAndAggregateQuotas(t *testing.T) {
	fixed := &model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: 2}
	if _, err := valueCodec(false).Clone(fixed, Value{Elements: []Value{{Bits: 1}}}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	mapType := &model.Type{Kind: model.Map, Key: typ(model.U8), Element: typ(model.String)}
	if _, err := valueCodec(false).Clone(mapType, Value{Elements: []Value{{}}}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	value := Value{Elements: []Value{{Elements: []Value{{Bits: 1}, {Text: "hello"}}}}}
	for _, budget := range []uint64{0, 63, 64, 191, 192} {
		c := valueCodec(false)
		c.Transfer.MaxBytes = budget
		if _, err := c.Clone(mapType, value); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(budget, err)
		}
	}
	c := valueCodec(false)
	c.MaxValues = 1
	if _, err := c.Clone(mapType, value); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
