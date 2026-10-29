package canonical

import (
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncMapAndFixedListLayouts(t *testing.T) {
	for _, memory64 := range []bool{false, true} {
		for _, d := range []*model.Type{
			{Kind: model.ErrorContext}, {Kind: model.Future}, {Kind: model.Future, Element: typ(model.String)},
			{Kind: model.Stream}, {Kind: model.Stream, Element: typ(model.U8)},
		} {
			l, err := NewBuilder(memory64, DefaultLimits()).Build(d)
			if err != nil || l.Size != 4 || l.Align != 4 || l.FlatCount != 1 || l.Flat[0] != wasm.I32 {
				t.Fatal(d, l, err)
			}
		}
		l, err := NewBuilder(memory64, DefaultLimits()).Build(&model.Type{Kind: model.Map, Key: typ(model.String), Element: typ(model.U64)})
		pointer := uint64(4)
		if memory64 {
			pointer = 8
		}
		if err != nil || l.Size != 2*pointer || l.FlatCount != 2 {
			t.Fatal(l, err)
		}
		for _, n := range []uint32{1, 16, 17, 1 << 20} {
			l, err := NewBuilder(memory64, DefaultLimits()).Build(&model.Type{Kind: model.FixedList, Element: typ(model.U8), Length: n})
			if err != nil || l.Size != uint64(n) || l.Align != 1 || l.FlatCount != uint64(n) || n > 16 && l.Flat != nil || n <= 16 && len(l.Flat) != int(n) {
				t.Fatal(n, l, err)
			}
		}
	}
	for _, d := range []*model.Type{
		{Kind: model.FixedList}, {Kind: model.FixedList, Length: 1},
		{Kind: model.Map, Key: typ(model.U8)}, {Kind: model.Map, Element: typ(model.U8)},
		{Kind: model.Future, Element: typ(model.Function)},
	} {
		if _, err := NewBuilder(false, DefaultLimits()).Build(d); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(d, err)
		}
	}
	if _, err := NewBuilder(false, Limits{MaxNodes: 10, MaxDepth: 10, MaxBytes: 7}).Build(&model.Type{Kind: model.FixedList, Element: typ(model.U64), Length: 2}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	// A compact repeated shape can overflow size/flattening arithmetic without
	// allocating an expanded vector.
	child := typ(model.U8)
	for i := 0; i < 62; i++ {
		child = &model.Type{Kind: model.Tuple, Fields: []model.Field{{Type: child}, {Type: child}}}
	}
	if _, err := NewBuilder(false, Limits{MaxNodes: 1000, MaxDepth: 128, MaxBytes: math.MaxUint64}).Build(&model.Type{Kind: model.FixedList, Element: child, Length: math.MaxUint32}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
