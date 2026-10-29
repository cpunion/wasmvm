package canonical

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func typ(kind model.Kind) *model.Type { return &model.Type{Kind: kind} }

func TestPrimitiveAndPointerLayouts(t *testing.T) {
	for _, memory64 := range []bool{false, true} {
		builder := NewBuilder(memory64, DefaultLimits())
		for _, row := range []struct {
			kind model.Kind
			size uint64
			flat wasm.ValueType
		}{
			{model.Bool, 1, wasm.I32}, {model.S8, 1, wasm.I32}, {model.U8, 1, wasm.I32},
			{model.S16, 2, wasm.I32}, {model.U16, 2, wasm.I32},
			{model.S32, 4, wasm.I32}, {model.U32, 4, wasm.I32}, {model.Char, 4, wasm.I32},
			{model.S64, 8, wasm.I64}, {model.U64, 8, wasm.I64},
			{model.F32, 4, wasm.F32}, {model.F64, 8, wasm.F64},
		} {
			layout, err := builder.Build(typ(row.kind))
			if err != nil || layout.Size != row.size || layout.Align != row.size || layout.FlatCount != 1 || !reflect.DeepEqual(layout.Flat, []wasm.ValueType{row.flat}) {
				t.Fatal(row, layout, err)
			}
		}
		for _, kind := range []model.Kind{model.String, model.List} {
			layout, err := builder.Build(&model.Type{Kind: kind, Element: typ(model.U8)})
			pointer, flat := uint64(4), wasm.I32
			if memory64 {
				pointer, flat = 8, wasm.I64
			}
			if err != nil || layout.Size != 2*pointer || layout.Align != pointer || !reflect.DeepEqual(layout.Flat, []wasm.ValueType{flat, flat}) {
				t.Fatal(layout, err)
			}
		}
		for _, kind := range []model.Kind{model.Own, model.Borrow} {
			layout, err := builder.Build(&model.Type{Kind: kind, ResourceID: 7})
			if err != nil || layout.Size != 4 || layout.Align != 4 || layout.Flat[0] != wasm.I32 {
				t.Fatal(layout, err)
			}
		}
	}
}

func TestAggregatePaddingVariantsAndFlattening(t *testing.T) {
	b := NewBuilder(false, DefaultLimits())
	record := &model.Type{Kind: model.Record, Fields: []model.Field{{Name: "small", Type: typ(model.U8)}, {Name: "wide", Type: typ(model.U64)}, {Name: "end", Type: typ(model.U16)}}}
	l, err := b.Build(record)
	if err != nil || l.Size != 24 || l.Align != 8 || !reflect.DeepEqual(l.Offsets, []uint64{0, 8, 16}) {
		t.Fatal(l, err)
	}
	if again, err := b.Build(record); again != l || err != nil {
		t.Fatal("layout cache", again, err)
	}
	tuple := &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: record}, {Type: typ(model.U8)}}}
	l, err = b.Build(tuple)
	if err != nil || l.Size != 32 || l.Offsets[1] != 24 {
		t.Fatal(l, err)
	}
	v := &model.Type{Kind: model.Variant, Cases: []model.Field{{}, {Type: typ(model.F32)}, {Type: typ(model.U32)}, {Type: typ(model.F64)}}}
	l, err = b.Build(v)
	if err != nil || l.Size != 16 || l.Align != 8 || l.PayloadOffset != 8 || !reflect.DeepEqual(l.Flat, []wasm.ValueType{wasm.I32, wasm.I64}) {
		t.Fatal(l, err)
	}
	for _, cases := range [][]model.Field{
		{{Type: typ(model.F32)}, {Type: typ(model.F32)}},
		{{Type: typ(model.U32)}, {Type: typ(model.F32)}},
		{{}, {Type: typ(model.String)}},
	} {
		l, err = b.Build(&model.Type{Kind: model.Variant, Cases: cases})
		if err != nil || l.FlatCount < 2 {
			t.Fatal(l, err)
		}
	}
	l, err = b.Build(&model.Type{Kind: model.Option, Element: typ(model.U16)})
	if err != nil || l.Size != 4 || l.PayloadOffset != 2 {
		t.Fatal(l, err)
	}
	l, err = b.Build(&model.Type{Kind: model.Result, Cases: []model.Field{{}, {}}})
	if err != nil || l.Size != 1 || l.FlatCount != 1 {
		t.Fatal(l, err)
	}
	for _, n := range []int{1, 8, 9, 16, 17, 32} {
		l, err = b.Build(&model.Type{Kind: model.Flags, Labels: make([]string, n)})
		size := uint64(4)
		if n <= 8 {
			size = 1
		} else if n <= 16 {
			size = 2
		}
		if err != nil || l.Size != size || l.FlatCount != 1 {
			t.Fatal(n, l, err)
		}
	}
	for _, row := range []struct {
		n    int
		size uint64
	}{{1, 1}, {256, 1}, {257, 2}, {65536, 2}, {65537, 4}} {
		l, err = b.Build(&model.Type{Kind: model.Enum, Labels: make([]string, row.n)})
		if err != nil || l.Size != row.size {
			t.Fatal(row, l, err)
		}
	}
}

func TestSharedTypeDAGDoesNotExpandFlatVectors(t *testing.T) {
	typeGraph := typ(model.Bool)
	for i := 0; i < 40; i++ {
		typeGraph = &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: typeGraph}, {Type: typeGraph}}}
	}
	b := NewBuilder(true, Limits{MaxNodes: 256, MaxDepth: 64, MaxBytes: 1 << 50})
	l, err := b.Build(typeGraph)
	if err != nil || l.Size != 1<<40 || l.FlatCount != 1<<40 || l.Flat != nil {
		t.Fatal(l, err)
	}
	l, err = b.Build(&model.Type{Kind: model.Variant, Cases: []model.Field{{Type: typeGraph}, {Type: typ(model.U8)}}})
	if err != nil || l.FlatCount != 1+(1<<40) || l.Flat != nil {
		t.Fatal(l, err)
	}
	// Even when many fields share one type, offset allocation consumes budget.
	child := typ(model.U8)
	fields := []model.Field{{Type: child}, {Type: child}, {Type: child}}
	b = NewBuilder(false, Limits{MaxNodes: 6, MaxDepth: 8, MaxBytes: 100})
	if _, err = b.Build(&model.Type{Kind: model.Record, Fields: fields}); err != nil {
		t.Fatal(err)
	}
	if _, err = b.Build(&model.Type{Kind: model.Record, Fields: fields}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("unbounded repeated offsets", err)
	}
}

func TestInvalidGraphsAndExactBounds(t *testing.T) {
	for _, tpe := range []*model.Type{nil, typ(model.Function), typ(model.Resource), typ(model.Own), typ(model.Borrow), typ(model.List), typ(model.Option), typ(model.Record), typ(model.Tuple), typ(model.Variant), typ(model.Result), typ(model.Flags), typ(model.Enum), {Kind: model.Flags, Labels: make([]string, 33)}} {
		if _, err := NewBuilder(false, DefaultLimits()).Build(tpe); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(tpe, err)
		}
	}
	cycle := &model.Type{Kind: model.List}
	cycle.Element = cycle
	if _, err := NewBuilder(false, DefaultLimits()).Build(cycle); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := NewBuilder(false, Limits{}).Build(typ(model.U8)); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	for _, limit := range []Limits{{MaxNodes: 1, MaxDepth: 8, MaxBytes: 100}, {MaxNodes: 8, MaxDepth: 1, MaxBytes: 100}} {
		if _, err := NewBuilder(false, limit).Build(&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: typ(model.U8)}}}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	for _, size := range []uint64{7, 8} {
		b := NewBuilder(false, Limits{MaxNodes: 10, MaxDepth: 10, MaxBytes: size})
		l, err := b.Build(&model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: typ(model.U64)}}})
		if size == 7 && !errors.Is(err, wasm.ErrLimit) || size == 8 && (err != nil || l.Size != 8) {
			t.Fatal(size, l, err)
		}
	}
	tooLarge := typ(model.U64)
	for i := 0; i < 64; i++ {
		tooLarge = &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: tooLarge}, {Type: tooLarge}}}
	}
	if _, err := NewBuilder(true, Limits{MaxNodes: 1000, MaxDepth: 100, MaxBytes: ^uint64(0)}).Build(tooLarge); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewBuilder(false, Limits{MaxNodes: 100, MaxDepth: 10, MaxBytes: 8}).Build(&model.Type{Kind: model.Variant, Cases: []model.Field{{Type: typ(model.U64)}}}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
