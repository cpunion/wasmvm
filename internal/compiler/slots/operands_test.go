package slots

import (
	"fmt"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLastUseIntegerOperands(t *testing.T) {
	checked := 0
	for op := uint32(0x45); op <= 0xc4; op++ {
		if !(op <= 0x5a || op >= 0x67 && op <= 0x8a || op == 0xa7 || op == 0xac || op == 0xad || op >= 0xc0) || ir.EffectsOf(op)&ir.MayTrap != 0 {
			continue
		}
		checked++
		t.Run(fmt.Sprintf("%x", op), func(t *testing.T) {
			typ, constant := wasm.I32, uint32(0x41)
			if op >= 0x50 && op <= 0x5a || op >= 0x79 && op <= 0x8a || op == 0xa7 || op >= 0xc2 {
				typ, constant = wasm.I64, 0x42
			}
			result := typ
			if op <= 0x5a || op == 0xa7 {
				result = wasm.I32
			} else if op == 0xac || op == 0xad {
				result = wasm.I64
			}
			body := module.Expr{{Op: constant, Bits: 7}}
			if op != 0x45 && op != 0x50 && !(op >= 0x67 && op <= 0x69 || op >= 0x79 && op <= 0x7b || op >= 0xa7) {
				body = append(body, module.Instr{Op: constant, Bits: 3})
			}
			body = append(body, module.Instr{Op: op}, module.Instr{Op: 0xb})
			m := makeIR(t, &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{result}}}, Functions: []module.Function{{Body: body}}})
			compiled, err := Compile(m, DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			out := compiled.Functions[0]
			checkSymbolic(t, "last use", m.Functions[0], out)
			in := out.Code[len(body)-2]
			if (in.Dst == in.A) != (result == typ) {
				t.Fatal("last-use preference or typed storage violated", in, out.Slots)
			}
		})
	}
	if checked != 58 {
		t.Fatal("incomplete nontrapping integer scope", checked)
	}
}

func TestRepeatedSSAOperandReleasedOnce(t *testing.T) {
	f := &ir.Function{Signature: module.FuncType{Results: []wasm.ValueType{wasm.I32}}, Values: []wasm.ValueType{wasm.I32, wasm.I32},
		Sources: []ir.Source{
			{Wasm: module.Instr{Op: 0x41, Bits: 7}, Outputs: []wasm.ValueType{wasm.I32}},
			{Wasm: module.Instr{Op: 0x6a}, Inputs: []wasm.ValueType{wasm.I32, wasm.I32}, Outputs: []wasm.ValueType{wasm.I32}},
			{Wasm: module.Instr{Op: 0xb}},
		}, Blocks: []ir.Block{{Ops: []ir.Operation{{Source: 0, Results: []ir.Value{0}}, {Source: 1, Args: []ir.Value{0, 0}, Results: []ir.Value{1}}}, Term: ir.Terminator{Kind: ir.Return, Source: 2, Args: []ir.Value{1}}}}}
	compiled, err := Compile(&lower.Module{Functions: []*ir.Function{f}}, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	out := compiled.Functions[0]
	checkSymbolic(t, "duplicate operand", f, out)
	if len(out.Slots) != 1 {
		t.Fatal("repeated operand allocated extra storage", out.Slots)
	}
}
