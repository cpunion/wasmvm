package canonical

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFlatAggregatesAndVariantBitJoining(t *testing.T) {
	for _, wide := range []bool{false, true} {
		variant := &model.Type{Kind: model.Variant, Cases: []model.Field{{}, {Type: typ(model.F32)}, {Type: typ(model.U64)}, {Type: typ(model.String)}}}
		list := &model.Type{Kind: model.List, Element: typ(model.U8)}
		tuple := &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: typ(model.S8)}, {Type: typ(model.String)}}}
		for _, row := range []struct {
			tpe *model.Type
			v   Value
		}{
			{typ(model.Bool), Value{Bits: 1}},
			{tuple, Value{Elements: []Value{{Bits: ^uint64(0)}, {Text: "中文😀"}}}},
			{list, Value{Elements: []Value{{Bits: 1}, {Bits: 2}}}},
			{&model.Type{Kind: model.Flags, Labels: []string{"a", "b"}}, Value{Bits: 3}},
			{&model.Type{Kind: model.Enum, Labels: []string{"a", "b"}}, Value{Bits: 1}},
			{variant, Value{}}, {variant, Value{Case: 1, Elements: []Value{{Bits: 0x3f800000}}}},
			{variant, Value{Case: 2, Elements: []Value{{Bits: ^uint64(0)}}}}, {variant, Value{Case: 3, Elements: []Value{{Text: "payload"}}}},
		} {
			c := valueCodec(wide)
			flat, err := c.LowerFlat(row.tpe, row.v)
			if err != nil {
				t.Fatal(row.tpe.Kind, err)
			}
			got, err := c.LiftFlat(row.tpe, flat)
			if err != nil || !reflect.DeepEqual(row.v, got) {
				t.Fatal(wide, row.v, flat, got, err)
			}
		}
		c := valueCodec(wide)
		l, _ := c.Layouts.Build(variant)
		slots := make([]Slot, len(l.Flat))
		for i, kind := range l.Flat {
			slots[i] = Slot{kind, 0}
		}
		slots[0].Bits = 1
		slots[1].Bits = 0xffffffff3f800000
		if !wide {
			slots[2].Bits = 0xffffffff
		}
		v, err := c.LiftFlat(variant, slots)
		if err != nil || v.Elements[0].Bits != 0x3f800000 {
			t.Fatal("joined payload truncation", v, err)
		}
	}
	// A scalar-only canonical function needs no memory option.
	c := valueCodec(false)
	c.Transfer.Memory = nil
	if v, err := c.LiftFlat(typ(model.U32), []Slot{{wasm.I32, 42}}); err != nil || v.Bits != 42 {
		t.Fatal(v, err)
	}
}

func TestFlatInvalidInputsAndBudgets(t *testing.T) {
	c := valueCodec(false)
	for _, slots := range [][]Slot{nil, {{wasm.I64, 1}}, {{wasm.I32, 1 << 32}}} {
		if _, err := c.LiftFlat(typ(model.U32), slots); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, tpe := range []*model.Type{nil, {Kind: model.Function}} {
		if _, err := c.LiftFlat(tpe, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := c.LowerFlat(typ(model.U8), Value{Bits: 256}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	variant := &model.Type{Kind: model.Option, Element: typ(model.U8)}
	if _, err := c.LiftFlat(variant, []Slot{{wasm.I32, 2}, {wasm.I32, 0}}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := c.LiftFlat(typ(model.Char), []Slot{{wasm.I32, 0xd800}}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	for _, kind := range []model.Kind{model.Enum, model.Flags} {
		tpe := &model.Type{Kind: kind, Labels: []string{"a"}}
		v, err := c.LiftFlat(tpe, []Slot{{wasm.I32, 3}})
		if kind == model.Enum && !errors.Is(err, wasm.ErrTrap) || kind == model.Flags && (err != nil || v.Bits != 1) {
			t.Fatal(kind, v, err)
		}
	}
	tpe := &model.Type{Kind: model.Tuple}
	v := Value{}
	for i := 0; i < 17; i++ {
		tpe.Fields = append(tpe.Fields, model.Field{Type: typ(model.U8)})
		v.Elements = append(v.Elements, Value{})
	}
	if _, err := c.LowerFlat(tpe, v); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := c.LiftFlat(tpe, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	res := &model.Type{Kind: model.Borrow, ResourceID: 1}
	if _, err := c.LiftFlat(res, []Slot{{wasm.I32, 1}}); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.MaxValues = 0
	if _, err := c.LiftFlat(typ(model.U8), []Slot{{wasm.I32, 1}}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Realloc = nil
	if _, err := c.LowerFlat(typ(model.String), Value{Text: "x"}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Memory.Write(0, []byte{0xff})
	if _, err := c.LiftFlat(typ(model.String), []Slot{{wasm.I32, 0}, {wasm.I32, 1}}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
}

func TestSynchronousFunctionSpillSignatures(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, lifting := range []bool{false, true} {
			b := NewBuilder(wide, DefaultLimits())
			ptr := wasm.I32
			if wide {
				ptr = wasm.I64
			}
			empty, err := b.Function(nil, nil, lifting)
			if err != nil || len(empty.Params) != 0 || len(empty.Results) != 0 {
				t.Fatal(empty, err)
			}
			direct, err := b.Function([]*model.Type{typ(model.U32)}, typ(model.U64), lifting)
			if err != nil || !reflect.DeepEqual(direct.Params, []wasm.ValueType{wasm.I32}) || !reflect.DeepEqual(direct.Results, []wasm.ValueType{wasm.I64}) {
				t.Fatal(direct, err)
			}
			params := make([]*model.Type, 17)
			for i := range params {
				params[i] = typ(model.U8)
			}
			spilled, err := b.Function(params, typ(model.String), lifting)
			if err != nil || spilled.ParameterMemory == nil || spilled.ResultMemory == nil {
				t.Fatal(spilled, err)
			}
			if lifting {
				if !reflect.DeepEqual(spilled.Params, []wasm.ValueType{ptr}) || !reflect.DeepEqual(spilled.Results, []wasm.ValueType{ptr}) {
					t.Fatal(spilled)
				}
			} else {
				if !reflect.DeepEqual(spilled.Params, []wasm.ValueType{ptr, ptr}) || len(spilled.Results) != 0 {
					t.Fatal(spilled)
				}
			}
			if _, err := b.Function([]*model.Type{nil}, nil, lifting); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(err)
			}
			if _, err := b.Function(nil, typ(model.Function), lifting); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(err)
			}
		}
	}
	if _, err := NewBuilder(false, Limits{}).Function([]*model.Type{typ(model.U8)}, nil, true); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
