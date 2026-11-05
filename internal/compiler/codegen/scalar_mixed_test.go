package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMixedScalarCopyAndResultSelection(t *testing.T) {
	for _, op := range []uint32{0x6a, 0x6b, 0x6c, 0x71, 0x72, 0x73, 0x7c, 0x7d, 0x7e, 0x83, 0x84, 0x85} {
		for _, result := range []bool{false, true} {
			f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64}}
			emit := func(op uint32, dst, a, b slots.Slot, local uint32) {
				source := uint32(len(f.Sources))
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local}})
				f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
			}
			if result {
				emit(op, 2, 0, 1, 0)
				emit(0x21, slots.NoSlot, 2, slots.NoSlot, 3)
			} else {
				emit(0x20, 3, slots.NoSlot, slots.NoSlot, 0)
				emit(0x20, 4, slots.NoSlot, slots.NoSlot, 1)
				emit(op, 2, 3, 4, 0)
			}
			e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, unmetered: true, registers: map[slots.Slot]byte{0: 3}}
			if e.canRegisterInteger(op, 2, 0, 1) || !e.canRegisterScalar(slots.Instruction{Op: op, Dst: 2, A: 0, B: 1}, 0) {
				t.Fatal(op, result, "mixed form missing from shared eligibility")
			}
			if result {
				if !e.forwardScalarResult(0) {
					t.Fatal(op, "result forwarding rejected")
				}
			} else if end, ok := e.combineScalarCopies(0); !ok || end != 2 {
				t.Fatal(op, "copy forwarding rejected", end, ok)
			}
			if e.mixedIntegers != 1 || len(e.a.code) == 0 {
				t.Fatal(op, result, "shared selection did not emit its mixed form", e.mixedIntegers)
			}
		}
	}
}
