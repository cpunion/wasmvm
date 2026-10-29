package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func shiftedIntegerFixture(wide bool, shift, consumer uint32, count uint64, intermediate, destination, left slots.Slot, swapped, preserve bool) *slots.Function {
	typ, literal, first, add := wasm.I32, uint32(0x41), uint32(0x74), uint32(0x6a)
	if wide {
		typ, literal, first, add = wasm.I64, 0x42, 0x86, 0x7c
	} else {
		count = uint64(uint32(count))
	}
	f := &slots.Function{Slots: []wasm.ValueType{typ, typ, typ, typ, typ, typ}, Sources: []ir.Source{{Wasm: module.Instr{Bits: count}}, {Wasm: module.Instr{Bits: 7}}}}
	a, b := left, intermediate
	if swapped {
		a, b = b, a
	}
	f.Code = []slots.Instruction{
		{Op: literal, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		{Op: first + shift, Dst: intermediate, A: 0, B: 2, C: slots.NoSlot},
		{Op: add + consumer, Dst: destination, A: a, B: b, C: slots.NoSlot},
	}
	if !preserve && intermediate != destination {
		f.Code = append(f.Code, slots.Instruction{Op: literal, Source: 1, Dst: intermediate, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	for i := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
	}
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: uint32(len(f.Args)), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	return f
}

func TestShiftedIntegerWidthsCountsAliasesAndColdEntries(t *testing.T) {
	aliases := [][3]slots.Slot{{3, 4, 1}, {0, 4, 1}, {3, 0, 1}, {3, 1, 1}, {3, 2, 2}, {2, 4, 1}, {3, 3, 1}, {3, 4, 2}, {1, 4, 1}}
	for _, wide := range []bool{false, true} {
		for shift := uint32(0); shift < 5; shift++ {
			for _, consumer := range []uint32{0, 1, 7, 8, 9} {
				for _, count := range []uint64{0, 1, 31, 32, 63, 64, ^uint64(0)} {
					// Rotate through alias layouts without multiplying all dimensions.
					alias := aliases[(uint64(shift*5+consumer)+count)%uint64(len(aliases))]
					for _, swapped := range []bool{false, true} {
						for _, preserve := range []bool{false, true} {
							f := shiftedIntegerFixture(wide, shift, consumer, count, alias[0], alias[1], alias[2], swapped, preserve)
							e := functionEmitter{f: f, a: &assembler{arch: "arm64"}, unmetered: true, registers: map[slots.Slot]byte{0: 19, 1: 20, 2: 21, 3: 22, 4: 23, 5: 24}}
							_, selected := e.combineShiftedInteger(0)
							want := (consumer >= 7 || shift < 3) && (consumer != 1 || !swapped) && alias[2] != alias[0] && (!preserve || alias[0] == alias[1])
							if selected != want {
								t.Fatal(wide, shift, consumer, count, alias, swapped, preserve, selected, want)
							}
							if native.Available {
								checkLinearIntegerEntries(t, f)
							}
						}
					}
				}
			}
		}
	}
}

func TestShiftedIntegerSelectionBoundaries(t *testing.T) {
	for _, kind := range []string{"metered", "target", "short", "poll", "constant", "shift", "count", "source", "consumer", "input", "spill", "trap", "guard", "region"} {
		f := shiftedIntegerFixture(false, 0, 0, 5, 3, 4, 1, false, false)
		e := functionEmitter{f: f, a: &assembler{arch: "arm64"}, unmetered: true, registers: map[slots.Slot]byte{0: 19, 1: 20, 2: 21, 3: 22, 4: 23}}
		pc := 0
		switch kind {
		case "metered":
			e.unmetered = false
		case "target":
			e.a.arch = "amd64"
		case "short":
			f.Code = f.Code[:2]
		case "poll":
			f.Code = append(make([]slots.Instruction, 254), f.Code...)
			pc = 254
		case "constant":
			f.Code[0].Op = 0x44
		case "shift":
			f.Code[1].Op = 0x6c
		case "count":
			f.Code[1].B = 1
		case "source":
			f.Code[1].A = 2
		case "consumer":
			f.Code[2].Op = 0x6c
		case "input":
			f.Code[2].B = 0
		case "spill":
			delete(e.registers, 4)
		case "trap":
			f.Code = append(f.Code[:3], append([]slots.Instruction{{Op: 0x6d, Dst: 5, A: 0, B: 1, C: slots.NoSlot}}, f.Code[3:]...)...)
		case "guard":
			e.guardPCs = map[int]memoryGuard{3: {address: 0, after: 1}}
		case "region":
			e.regionLimit = 2
		}
		if _, ok := e.combineShiftedInteger(pc); ok || len(e.a.code) != 0 {
			t.Fatal("unsafe selection", kind)
		}
	}
}
