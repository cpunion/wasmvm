package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestRegisterPriorityIgnoresSelfCopiesAndIncludesTuples(t *testing.T) {
	f := &slots.Function{LocalCount: 13, Args: []slots.Slot{11, 12}}
	for i := 0; i < 13; i++ {
		f.Slots = append(f.Slots, wasm.I64)
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(i)}})
		for j := 0; j < 100; j++ {
			f.Code = append(f.Code, slots.Instruction{Op: 0x20, Source: uint32(i), Dst: slots.Slot(i), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
	}
	f.Code = append(f.Code, slots.Instruction{Op: 0x7c, Dst: 12, A: 10, B: 10, C: slots.NoSlot})
	for _, arch := range []string{"amd64", "arm64"} {
		e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: true}
		e.prepareRegisters()
		if len(e.registers) != 3 {
			t.Fatal("eliminated copies consumed registers", arch, e.registers)
		}
		for _, s := range []slots.Slot{10, 11, 12} {
			if _, ok := e.registers[s]; !ok {
				t.Fatal("real operand or tuple omitted", arch, s, e.registers)
			}
		}
	}
}

func TestSharedLiteralsPreserveColdEntries(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I64}}, 0, module.Expr{
		{Op: 0x42, Bits: 0x123456789abcdef0}, {Op: 0x42, Bits: 0x123456789abcdef0}, {Op: 0x85},
		{Op: 0x42, Bits: 7}, {Op: 0x42, Bits: 7}, {Op: 0x7c}, {Op: 0x7c}, {Op: 0xb},
	})
	f := functionModule(t, m).Functions[0]
	if f.Code[0].Dst != f.Code[1].Dst || f.Code[3].Dst != f.Code[4].Dst {
		t.Fatal("fixture does not exercise shared literal storage", f.Code)
	}
	// Keep the compiled straight-line slot layout and remove only its final
	// result-edge transfer, allowing the independent integer checker to enter
	// at every original operation with arbitrary, unequal initial slot bits.
	f.Code = append(f.Code[:len(f.Code)-2], f.Code[len(f.Code)-1])
	f.Edges, f.Moves = nil, nil
	checkLinearIntegerEntries(t, f)
}
