package canonical

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncLowerFourSlotAndIndirectResultABI(t *testing.T) {
	for _, memory64 := range []bool{false, true} {
		b := NewBuilder(memory64, DefaultLimits())
		address := wasm.I32
		if memory64 {
			address = wasm.I64
		}
		params := []*model.Type{typ(model.String), typ(model.U32), typ(model.U64)}
		layout, err := b.AsyncLower(params, typ(model.U32))
		if err != nil || !reflect.DeepEqual(layout.Params, []wasm.ValueType{address, address, wasm.I32, wasm.I64, address}) || !reflect.DeepEqual(layout.Results, []wasm.ValueType{wasm.I32}) || layout.ParameterMemory != nil || layout.ResultMemory.Kind != model.U32 {
			t.Fatal(layout, err)
		}
		params = append(params, typ(model.Bool))
		layout, err = b.AsyncLower(params, typ(model.U32))
		if err != nil || !reflect.DeepEqual(layout.Params, []wasm.ValueType{address, address}) || layout.ParameterMemory == nil || layout.ResultMemory == nil {
			t.Fatal(layout, err)
		}
		layout, err = b.AsyncLower(params, nil)
		if err != nil || !reflect.DeepEqual(layout.Params, []wasm.ValueType{address}) || layout.ResultMemory != nil {
			t.Fatal(layout, err)
		}
		layout, err = b.AsyncLower(nil, nil)
		if err != nil || len(layout.Params) != 0 || !reflect.DeepEqual(layout.Results, []wasm.ValueType{wasm.I32}) {
			t.Fatal(layout, err)
		}
		if _, err := b.AsyncLower(nil, &model.Type{Kind: model.Function}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	b := NewBuilder(false, Limits{MaxNodes: 0, MaxDepth: 128, MaxBytes: 1024})
	if _, err := b.AsyncLower([]*model.Type{typ(model.U32)}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
