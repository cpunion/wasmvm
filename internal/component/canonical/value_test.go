package canonical

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func valueCodec(memory64 bool) *Codec {
	x, _ := textTransfer(UTF8, memory64)
	return &Codec{Transfer: x, Layouts: NewBuilder(memory64, DefaultLimits()), MaxValues: 1024}
}

func TestNestedMemoryValues(t *testing.T) {
	for _, wide := range []bool{false, true} {
		variant := &model.Type{Kind: model.Variant, Cases: []model.Field{{Name: "empty"}, {Name: "text", Type: typ(model.String)}}}
		list := &model.Type{Kind: model.List, Element: variant}
		record := &model.Type{Kind: model.Record, Fields: []model.Field{{Name: "code", Type: typ(model.U8)}, {Name: "entries", Type: list}}}
		option := &model.Type{Kind: model.Option, Element: typ(model.U64)}
		result := &model.Type{Kind: model.Result, Cases: []model.Field{{Type: typ(model.String)}, {}}}
		for _, row := range []struct {
			tpe *model.Type
			v   Value
		}{
			{typ(model.U64), Value{Bits: 42}},
			{typ(model.String), Value{Text: "中文😀"}},
			{record, Value{Elements: []Value{{Bits: 7}, {Elements: []Value{{Case: 0}, {Case: 1, Elements: []Value{{Text: "one"}}}, {Case: 1, Elements: []Value{{Text: "二"}}}}}}}},
			{option, Value{}}, {option, Value{Case: 1, Elements: []Value{{Bits: 0xffffffffffffffff}}}},
			{result, Value{Case: 1}}, {result, Value{Elements: []Value{{Text: "ok"}}}},
			{&model.Type{Kind: model.Flags, Labels: []string{"a", "b", "c"}}, Value{Bits: 5}},
			{&model.Type{Kind: model.Enum, Labels: []string{"a", "b"}}, Value{Bits: 1}},
			{&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: typ(model.S8)}, {Type: typ(model.F64)}}}, Value{Elements: []Value{{Bits: ^uint64(0)}, {Bits: 0x3ff0000000000000}}}},
			{list, Value{Elements: []Value{}}},
		} {
			c := valueCodec(wide)
			if err := c.Store(row.tpe, 0, row.v); err != nil {
				t.Fatal(row.tpe.Kind, err)
			}
			got, err := c.Load(row.tpe, 0)
			if err != nil || !reflect.DeepEqual(row.v, got) {
				t.Fatal(wide, row.tpe.Kind, row.v, got, err)
			}
		}
	}
	c := valueCodec(false)
	flags := &model.Type{Kind: model.Flags, Labels: []string{"a", "b"}}
	c.Transfer.Memory.Write(0, []byte{0xff})
	if v, err := c.Load(flags, 0); err != nil || v.Bits != 3 {
		t.Fatal("unused flag bits", v, err)
	}
}

func TestHostAggregateValidationPrecedesEffects(t *testing.T) {
	scalar := typ(model.U8)
	variant := &model.Type{Kind: model.Variant, Cases: []model.Field{{}, {Type: scalar}}}
	for _, row := range []struct {
		tpe *model.Type
		v   Value
	}{
		{scalar, Value{Bits: 256}},
		{&model.Type{Kind: model.Record, Fields: []model.Field{{Type: scalar}}}, Value{}},
		{&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: scalar}}}, Value{Elements: []Value{{Bits: 256}}}},
		{&model.Type{Kind: model.List, Element: scalar}, Value{Elements: []Value{{Bits: 256}}}},
		{&model.Type{Kind: model.Flags, Labels: []string{"a"}}, Value{Bits: 2}},
		{&model.Type{Kind: model.Enum, Labels: []string{"a"}}, Value{Bits: 1}},
		{variant, Value{Case: 2}}, {variant, Value{Elements: []Value{{}}}}, {variant, Value{Case: 1}},
		{variant, Value{Case: 1, Elements: []Value{{Bits: 256}}}},
		{typ(model.String), Value{Text: string([]byte{0xff})}},
	} {
		c := valueCodec(false)
		calls := 0
		c.Transfer.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { calls++; return 64, nil }
		if err := c.Store(row.tpe, 0, row.v); !errors.Is(err, wasm.ErrInvalid) || calls != 0 {
			t.Fatal(row, err, calls)
		}
	}
	resource := &model.Type{Kind: model.Own, ResourceID: 1}
	c := valueCodec(false)
	if err := c.Store(resource, 0, Value{}); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := c.Load(resource, 0); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestInvalidGuestDiscriminantsPointersAndBudgets(t *testing.T) {
	for _, tpe := range []*model.Type{{Kind: model.Enum, Labels: []string{"a"}}, {Kind: model.Variant, Cases: []model.Field{{}}}, {Kind: model.Option, Element: typ(model.U8)}} {
		c := valueCodec(false)
		c.Transfer.Memory.Write(0, []byte{255})
		if _, err := c.Load(tpe, 0); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
	}
	list := &model.Type{Kind: model.List, Element: typ(model.U8)}
	for _, row := range [][2]uint64{{4096, 1}, {0, MaxStringBytes + 1}, {0, 1024}} {
		c := valueCodec(false)
		c.storePair(0, row[0], row[1])
		if _, err := c.Load(list, 0); err == nil {
			t.Fatal(row)
		}
	}
	c := valueCodec(false)
	c.MaxValues = 0
	if _, err := c.Load(typ(model.U8), 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.MaxValues = 1
	if err := c.Store(list, 0, Value{Elements: []Value{{}}}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.MaxBytes = 1
	c.storePair(0, 64, 1)
	if _, err := c.Load(list, 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Realloc = nil
	if err := c.Store(list, 0, Value{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return 4096, nil }
	if err := c.Store(list, 0, Value{Elements: []Value{{}}}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.Realloc = func(uint64, uint64, uint64, uint64) (uint64, error) { return 0, context.Canceled }
	if err := c.Store(list, 0, Value{Elements: []Value{{}}}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	c.Transfer.MaxBytes = 0
	if err := c.Store(list, 0, Value{Elements: []Value{{}}}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c = valueCodec(false)
	if _, err := c.Load(typ(model.U64), 1); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := c.Load(nil, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := c.Store(nil, 0, Value{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.Transfer.Context = ctx
	if err := c.Store(list, 0, Value{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
