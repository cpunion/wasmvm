package vm

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarConstantResultBitWidths(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, test := range []struct {
			name string
			typ  wasm.ValueType
			op   uint32
			bits uint64
			want uint64
		}{
			{"i32-negative-one", wasm.I32, 0x41, ^uint64(0), 0xffffffff},
			{"i32-minimum", wasm.I32, 0x41, 0xffffffff80000000, 0x80000000},
			{"i64-full-word", wasm.I64, 0x42, 0x123456789abcdef0, 0x123456789abcdef0},
			{"f32-nan-payload", wasm.F32, 0x43, 0xff800001, 0xff800001},
			{"f64-nan-payload", wasm.F64, 0x44, 0xfff0000000000001, 0xfff0000000000001},
		} {
			t.Run(test.name, func(t *testing.T) {
				e, _, linker := setupWithEngine(t, newEngine)
				definition := &module.Module{
					Types:     []module.FuncType{{Results: []wasm.ValueType{test.typ}}},
					Functions: []module.Function{{Body: module.Expr{{Op: test.op, Bits: test.bits}, {Op: 0xb}}}},
					Exports:   []module.Export{{Name: "run", Kind: 0}},
				}
				fn := export(t, instantiate(t, linker, compileDefinition(t, e, definition)), "run")
				for _, metered := range []bool{false, true} {
					result, err := fn.Call(nil, nil, CallOptions{Metered: metered, Fuel: 10})
					if err != nil {
						t.Fatal(err)
					}
					lo, hi, err := result.Values[0].Bits()
					if err != nil || lo != test.want || hi != 0 || result.Values[0].Type() != test.typ {
						t.Errorf("metered=%v: bits=%x/%x type=%v, want %x/0 type=%v: %v", metered, lo, hi, result.Values[0].Type(), test.want, test.typ, err)
					}
				}
			})
		}
	})
}
