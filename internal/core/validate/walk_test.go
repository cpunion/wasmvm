package validate

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestWalkSignatures(t *testing.T) {
	m := validFunction(module.Expr{{Op: 0x41}, {Op: 0x42}, {Op: 0x10}, {Op: 0x1a}, {Op: 0x1a}, {Op: 0xb}})
	m.Types = append(m.Types, module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I64}, Results: []wasm.ValueType{wasm.F64, wasm.F32}})
	m.Imports = []module.Import{{Kind: 0, Type: 1}}
	checked, err := Check(m, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	wantIn := [][]wasm.ValueType{nil, nil, {wasm.I32, wasm.I64}, {wasm.F32}, {wasm.F64}, nil}
	wantOut := [][]wasm.ValueType{{wasm.I32}, {wasm.I64}, {wasm.F64, wasm.F32}, nil, nil, nil}
	// Replay independently: the visitor must not retain mutable checker state.
	for round := 0; round < 2; round++ {
		calls := 0
		err := checked.WalkFunction(0, func(pc int, in module.Instr, inputs, outputs []wasm.ValueType) error {
			if pc != calls || in.Op != m.Functions[0].Body[pc].Op {
				t.Fatalf("callback order: %d %d", pc, calls)
			}
			if !reflect.DeepEqual(append([]wasm.ValueType(nil), inputs...), wantIn[pc]) || !reflect.DeepEqual(append([]wasm.ValueType(nil), outputs...), wantOut[pc]) {
				t.Fatalf("pc=%d: inputs=%v outputs=%v", pc, inputs, outputs)
			}
			calls++
			return nil
		})
		if err != nil || calls != len(wantIn) {
			t.Fatal(calls, err)
		}
	}
	stop := errors.New("visitor stopped")
	calls := 0
	err = checked.WalkFunction(0, func(int, module.Instr, []wasm.ValueType, []wasm.ValueType) error { calls++; return stop })
	if !errors.Is(err, stop) || calls != 1 {
		t.Fatal(calls, err)
	}
	if err := checked.WalkFunction(1, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := checked.WalkFunction(0, nil); err != nil {
		t.Fatal(err)
	}
}

func TestWalkUnreachableSignatures(t *testing.T) {
	m := validFunction(module.Expr{{Op: 0}, {Op: 0x6a}, {Op: 0x1a}, {Op: 0x1a}, {Op: 0xb}})
	checked, err := Check(m, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	err = checked.WalkFunction(0, func(pc int, _ module.Instr, inputs, outputs []wasm.ValueType) error {
		if pc == 1 && (!reflect.DeepEqual(inputs, []wasm.ValueType{wasm.I32, wasm.I32}) || !reflect.DeepEqual(outputs, []wasm.ValueType{wasm.I32})) {
			t.Fatal(inputs, outputs)
		}
		if pc == 3 && (len(inputs) != 1 || inputs[0] != 0) {
			t.Fatalf("untyped bottom: %v", inputs)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
