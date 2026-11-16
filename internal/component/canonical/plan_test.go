package canonical

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestValuePlansPreserveOwnedConversionsAndBudgets(t *testing.T) {
	rows := []struct {
		typ *model.Type
		v   Value
	}{
		{typ(model.Bool), Value{Bits: 1}},
		{typ(model.S8), Value{Bits: ^uint64(0)}},
		{typ(model.Char), Value{Bits: 0x10ffff}},
		{typ(model.F32), Value{Bits: 0xff800001}},
		{typ(model.F64), Value{Bits: 0xfff0000000000001}},
		{typ(model.String), Value{Text: "text😀"}},
		{&model.Type{Kind: model.List, Element: typ(model.U8)}, Value{Text: "bytes"}},
		{&model.Type{Kind: model.FixedList, Element: typ(model.U16), Length: 2}, Value{Elements: []Value{{Bits: 17}, {Bits: 23}}}},
		{&model.Type{Kind: model.Map, Key: typ(model.String), Element: typ(model.U32)}, Value{Elements: []Value{{Elements: []Value{{Text: "key"}, {Bits: 19}}}}}},
		{&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: typ(model.U8)}, {Type: typ(model.String)}}}, Value{Elements: []Value{{Bits: 7}, {Text: "tuple"}}}},
		{&model.Type{Kind: model.Option, Element: typ(model.U64)}, Value{}},
		{&model.Type{Kind: model.Result, Cases: []model.Field{{Type: typ(model.U8)}, {}}}, Value{Elements: []Value{{Bits: 11}}}},
		{&model.Type{Kind: model.Variant, Cases: []model.Field{{Type: typ(model.F32)}, {Type: typ(model.U64)}}}, Value{Elements: []Value{{Bits: 0x3f800000}}}},
		{&model.Type{Kind: model.Flags, Labels: []string{"a", "b"}}, Value{Bits: 3}},
		{&model.Type{Kind: model.Enum, Labels: []string{"a", "b"}}, Value{Bits: 1}},
	}
	for _, row := range rows {
		plan := NewValuePlan(row.typ)
		for _, wide := range []bool{false, true, false} {
			plain, bound := valueCodec(wide), valueCodec(wide)
			borrowed := valueCodec(wide)
			borrowedPlan := NewValuePlan(row.typ)
			for repeat := 0; repeat < 2; repeat++ {
				beforeValues, beforeBytes := plain.Usage()
				owned, err := plain.Clone(row.typ, row.v)
				if err != nil {
					t.Fatal(err)
				}
				afterValues, afterBytes := plain.Usage()
				borrowValues, borrowBytes := borrowed.Usage()
				if err := borrowedPlan.RetainBorrowed(borrowed, row.v); err != nil {
					t.Fatal(err)
				}
				nextValues, nextBytes := borrowed.Usage()
				if nextValues-borrowValues != afterValues-beforeValues || nextBytes-borrowBytes != afterBytes-beforeBytes {
					t.Fatal("borrowed retention changed logical clone accounting", row.typ.Kind)
				}
				copy, err := plan.Clone(bound, row.v)
				if err != nil || !reflect.DeepEqual(owned, copy) {
					t.Fatal(row.typ.Kind, copy, owned, err)
				}
				want, err := plain.LowerFlat(row.typ, row.v)
				if err != nil {
					t.Fatal(err)
				}
				var local [MaxFlatParams]Slot
				got, err := plan.LowerFlatInto(bound, row.v, local[:0])
				if err != nil || !reflect.DeepEqual(want, got) {
					t.Fatal(row.typ.Kind, want, got, err)
				}
				count, err := plan.FlatCount(bound)
				if err != nil || count != uint64(len(got)) {
					t.Fatal(count, err)
				}
				wantValue, err := plain.LiftFlat(row.typ, want)
				if err != nil {
					t.Fatal(err)
				}
				gotValue, err := plan.LiftFlat(bound, got)
				if err != nil || !reflect.DeepEqual(wantValue, gotValue) {
					t.Fatal(row.typ.Kind, gotValue, wantValue, err)
				}
				pv, pb := plain.Usage()
				bv, bb := bound.Usage()
				if pv != bv || pb != bb {
					t.Fatal("planned conversion changed budget charges", pv, pb, bv, bb)
				}
				if len(copy.Elements) != 0 {
					copy.Elements[0].Bits = 255
					if reflect.DeepEqual(copy, row.v) {
						t.Fatal("clone retained caller aggregate storage")
					}
				}
			}
		}
	}
}

func TestValuePlanRejectsInvalidRepresentationsAndKeepsErrorOrdering(t *testing.T) {
	plan := NewValuePlan(typ(model.U32))
	for _, slots := range [][]Slot{nil, {{Type: wasm.I64, Bits: 1}}, {{Type: wasm.I32, Bits: 1 << 32}}} {
		if _, err := plan.LiftFlat(valueCodec(false), slots); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, valueType := range []*model.Type{nil, {Kind: model.Function}} {
		p := NewValuePlan(valueType)
		if _, err := p.FlatCount(valueCodec(false)); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	large := &model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: MaxFlatParams + 1}
	p := NewValuePlan(large)
	if _, err := p.LowerFlatInto(valueCodec(false), Value{Text: "12345678901234567"}, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	for _, warm := range []bool{false, true} {
		c := valueCodec(false)
		p := NewValuePlan(typ(model.Bool))
		if warm {
			if _, err := p.Clone(c, Value{Bits: 1}); err != nil {
				t.Fatal(err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c.Transfer.Context = ctx
		if _, err := p.LowerFlatInto(c, Value{Bits: 2}, nil); !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation must precede invalid host representation", err)
		}
		if err := p.RetainBorrowed(c, Value{Bits: 2}); !errors.Is(err, context.Canceled) {
			t.Fatal("borrowed cancellation must precede invalid representation", err)
		}
		c.Transfer.Context = nil
		c.MaxValues = c.values
		if _, err := p.Clone(c, Value{Bits: 2}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("value quota must precede invalid host representation", err)
		}
		if err := p.RetainBorrowed(c, Value{Bits: 2}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("borrowed value quota must precede invalid representation", err)
		}
	}
	char := NewValuePlan(typ(model.Char))
	if _, err := char.LiftFlat(valueCodec(false), []Slot{{Type: wasm.I32, Bits: 0xd800}}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	for _, method := range []func(*Codec) error{
		func(c *Codec) error { _, err := plan.Clone(c, Value{Bits: 1 << 32}); return err },
		func(c *Codec) error { return plan.RetainBorrowed(c, Value{Bits: 1 << 32}) },
		func(c *Codec) error { _, err := plan.LowerFlatInto(c, Value{Bits: 1 << 32}, nil); return err },
	} {
		if err := method(valueCodec(false)); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
}
