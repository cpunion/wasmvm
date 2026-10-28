package slots

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLiteralBitsAndTypes(t *testing.T) {
	for _, test := range []struct {
		name  string
		types []wasm.ValueType
		body  module.Expr
	}{
		{"integer types", []wasm.ValueType{wasm.I32, wasm.I64}, module.Expr{{Op: 0x41, Bits: 7}, {Op: 0x42, Bits: 7}}},
		{"integer high bits", []wasm.ValueType{wasm.I64, wasm.I64}, module.Expr{{Op: 0x42, Bits: 7}, {Op: 0x42, Bits: 1<<63 | 7}}},
		{"float types", []wasm.ValueType{wasm.F32, wasm.I32}, module.Expr{{Op: 0x43, Bits: 7}, {Op: 0x41, Bits: 7}}},
		{"f32 zero signs", []wasm.ValueType{wasm.F32, wasm.F32}, module.Expr{{Op: 0x43}, {Op: 0x43, Bits: 1 << 31}}},
		{"f64 zero signs", []wasm.ValueType{wasm.F64, wasm.F64}, module.Expr{{Op: 0x44}, {Op: 0x44, Bits: 1 << 63}}},
		{"f32 NaN payloads", []wasm.ValueType{wasm.F32, wasm.F32}, module.Expr{{Op: 0x43, Bits: 0x7fc00001}, {Op: 0x43, Bits: 0x7fc00002}}},
		{"f64 NaN payloads", []wasm.ValueType{wasm.F64, wasm.F64}, module.Expr{{Op: 0x44, Bits: 0x7ff8000000000001}, {Op: 0x44, Bits: 0x7ff8000000000002}}},
		{"vector high bits", []wasm.ValueType{wasm.V128, wasm.V128}, module.Expr{{Op: 0xfd000c, Vector: [16]byte{1}}, {Op: 0xfd000c, Vector: [16]byte{1, 8: 2}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Each literal has two overlapping definitions. Equal bits/types
			// share storage, while the two distinct literals remain independent.
			body := module.Expr{test.body[0], test.body[1], test.body[0], test.body[1], {Op: 0xb}}
			results := append(append([]wasm.ValueType(nil), test.types...), test.types...)
			m := makeIR(t, &module.Module{Types: []module.FuncType{{Results: results}}, Functions: []module.Function{{Body: body}}})
			compiled, err := Compile(m, DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			out := compiled.Functions[0]
			checkSymbolic(t, test.name, m.Functions[0], out)
			code := out.Code[:4]
			if code[0].Dst == code[1].Dst || code[0].Dst != code[2].Dst || code[1].Dst != code[3].Dst {
				t.Fatal("incorrect typed literal sharing", code)
			}
		})
	}
}

func TestLiteralLastAliasReleaseAndBoundedFrames(t *testing.T) {
	// The older 7 must survive after the second 7 is consumed, even when a
	// different literal reuses another slot. Zero-use values also release once.
	body := module.Expr{{Op: 0x41, Bits: 7}, {Op: 0x41, Bits: 7}, {Op: 0x41, Bits: 9}, {Op: 0x1a}, {Op: 0x1a}, {Op: 0x41, Bits: 11}, {Op: 0x1a}}
	for i := 0; i < 1000; i++ {
		body = append(body, module.Instr{Op: 0x41, Bits: uint64(i)}, module.Instr{Op: 0x41, Bits: uint64(i)}, module.Instr{Op: 0x6a}, module.Instr{Op: 0x1a})
	}
	body = append(body, module.Instr{Op: 0xb})
	m := makeIR(t, &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{{Body: body}}})
	compiled, err := Compile(m, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	out := compiled.Functions[0]
	checkSymbolic(t, "last alias", m.Functions[0], out)
	if len(out.Slots) > 4 {
		t.Fatal("inactive literal groups retained frame storage", len(out.Slots))
	}
}
