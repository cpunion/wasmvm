package codegen

import (
	"slices"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLoopLiteralVersionsRequireUnobservedInvariantWords(t *testing.T) {
	for _, kind := range []string{"same", "different", "read-before", "arithmetic", "copy-write", "shared-home", "precise", "header-poll", "trap", "unproved"} {
		literal := slots.Instruction{Op: 0x41, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
		f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}, Sources: []ir.Source{{Wasm: module.Instr{Op: 0x41, Bits: 7}}, {Wasm: module.Instr{Op: 0x41, Bits: 7}}}, Code: []slots.Instruction{{Op: 1}, literal, literal}}
		f.Code[2].Source = 1
		start := 1
		e := functionEmitter{a: &assembler{arch: "arm64"}, f: f, unmetered: true}
		switch kind {
		case "different":
			f.Sources[1].Wasm.Bits = 9
		case "read-before":
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: 0}})
			f.Code = append(f.Code[:1:1], slots.Instruction{Op: 0x20, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Source: 2}, literal, literal)
		case "arithmetic", "trap", "unproved":
			op := uint32(0x6a)
			if kind == "trap" {
				op = 0x6d
			} else if kind == "unproved" {
				op = 0x28
			}
			f.Code[2] = slots.Instruction{Op: op, Dst: 0, A: 0, B: 0, C: slots.NoSlot}
		case "copy-write":
			f.Sources[1].Wasm = module.Instr{Op: 0x20, A: 0}
			f.Code[2].Op = 0x20
		case "shared-home":
			e.useRegisterPlan(&registerPlan{homes: map[slots.Slot]byte{0: 19, 1: 19}, slots: []slots.Slot{0, 1}, shared: true})
		case "precise":
			e.unmetered = false
		case "header-poll":
			start = 0
		}
		r := memoryRegion{end: len(f.Code) - 1, loop: &nativeLoop{}}
		got := e.loopLiteralDefinitions(e.loopValueVersions(start, r))
		if kind == "same" {
			if !slices.Equal(got, []int{1}) {
				t.Fatal(kind, got)
			}
		} else if len(got) != 0 {
			t.Fatal(kind, "unsafe literal motion", got)
		}
	}
}

func invariantLiteralLoopFixture() *slots.Function {
	f := coalescedValueLoopFixture(true)
	var literals []slots.Instruction
	for _, item := range []struct {
		typ  wasm.ValueType
		op   uint32
		bits uint64
	}{{wasm.I32, 0x41, 0xff001234}, {wasm.I64, 0x42, 0xfedcba9876543210}, {wasm.F32, 0x43, 0x80000000}, {wasm.F64, 0x44, 0xfff0000000000001}, {wasm.V128, 0xfd000c, 0}, {wasm.ExternRef, 0xd0, 0}} {
		dst := slots.Slot(len(f.Slots))
		f.Slots = append(f.Slots, item.typ)
		f.Args = append(f.Args, dst)
		source := uint32(len(f.Sources))
		in := module.Instr{Op: item.op, Bits: item.bits, Offset: uint64(3000 + source)}
		in.Vector = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10}
		f.Sources = append(f.Sources, ir.Source{Wasm: in})
		for i := 0; i < 2; i++ {
			literals = append(literals, slots.Instruction{Op: item.op, Dst: dst, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Source: source})
		}
	}
	f.Code = append(append(append([]slots.Instruction(nil), f.Code[:1]...), literals...), f.Code[1:]...)
	f.LocalCount = uint32(len(f.Slots))
	f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
	f.Edges[1].PC = uint32(len(f.Code) - 1)
	return f
}
