package codegen

import (
	"fmt"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func logicalShiftFixture(wide bool, logical, shift uint32, count uint64, intermediate, result slots.Slot, preserve bool) *slots.Function {
	typ, literal, first := wasm.I32, uint32(0x41), uint32(0x71)
	if wide {
		typ, literal, first = wasm.I64, 0x42, 0x83
	} else {
		count = uint64(uint32(count))
	}
	f := &slots.Function{Slots: []wasm.ValueType{typ, typ, typ, typ, typ}, Sources: []ir.Source{
		{Wasm: module.Instr{Bits: count}}, {},
	}}
	f.Code = []slots.Instruction{
		{Op: first + logical, Dst: intermediate, A: 0, B: 1, C: slots.NoSlot},
		{Op: literal, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		{Op: first + 3 + shift, Dst: result, A: intermediate, B: 3, C: slots.NoSlot},
	}
	if !preserve && intermediate != result {
		f.Code = append(f.Code, slots.Instruction{Op: literal, Source: 1, Dst: intermediate, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	return f
}

func TestLogicalShiftIndependentEntries(t *testing.T) {
	if !native.Available {
		return
	}
	for _, wide := range []bool{false, true} {
		for logical := uint32(0); logical < 3; logical++ {
			for shift := uint32(0); shift < 5; shift++ {
				for _, count := range []uint64{0, 1, 5, 31, 32, 63, 64, ^uint64(0)} {
					t.Run(fmt.Sprintf("wide=%v/op=%d/shift=%d/count=%x", wide, logical, shift, count), func(t *testing.T) {
						f := logicalShiftFixture(wide, logical, shift, count, 2, 4, false)
						compiled, err := CompileUnmeteredFunction("arm64", f, DefaultLimits())
						if err != nil || compiled.FusedIntegers != 1 {
							t.Fatal(compiled, err)
						}
						checkLinearIntegerEntries(t, f)
					})
				}
			}
		}
	}
}

func TestLogicalShiftAliasesAndCopies(t *testing.T) {
	if !native.Available {
		return
	}
	for _, wide := range []bool{false, true} {
		for intermediate := slots.Slot(0); intermediate < 3; intermediate++ {
			for result := slots.Slot(0); result < 5; result++ {
				for _, prefix := range []bool{false, true} {
					for _, preserve := range []bool{false, true} {
						t.Run(fmt.Sprintf("wide=%v/intermediate=%d/result=%d/copies=%v/live=%v", wide, intermediate, result, prefix, preserve), func(t *testing.T) {
							f := logicalShiftFixture(wide, 2, 3, 5, intermediate, result, preserve)
							if prefix {
								f.LocalCount = 1
								// Forward through a temporary that the following constant
								// overwrites. The logical inputs still use their old value.
								f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{A: 0}})
								f.Code[0].A = 3
								f.Code = append([]slots.Instruction{{Op: 0x20, Source: 2, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}, f.Code...)
							}
							checkLinearIntegerEntries(t, f)
						})
					}
				}
			}
		}
	}
}

func TestLogicalShiftDeadConstant(t *testing.T) {
	if !native.Available {
		return
	}
	f := logicalShiftFixture(true, 0, 1, 63, 2, 2, false)
	f.Code = append(f.Code[:3],
		slots.Instruction{Op: 0x42, Source: 1, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}, f.Code[3])
	checkLinearIntegerEntries(t, f)
}

func TestLogicalShiftX86RightAlias(t *testing.T) {
	f := logicalShiftFixture(false, 2, 3, 5, 2, 1, false)
	compiled, err := CompileUnmeteredFunction("amd64", f, DefaultLimits())
	if err != nil || compiled.FusedIntegers != 1 {
		t.Fatal("right-alias combination not selected", compiled, err)
	}
	// All aliases and original cold entries are executed on each host target
	// by TestLogicalShiftAliasesAndCopies.
}

func TestLogicalShiftRejectsUnsafeCombinations(t *testing.T) {
	for _, kind := range []string{"metered", "short", "poll", "opcode", "constant", "shift", "input", "count", "clobber", "live", "spill"} {
		t.Run(kind, func(t *testing.T) {
			f := logicalShiftFixture(false, 2, 3, 5, 2, 4, false)
			e := functionEmitter{f: f, unmetered: true, a: &assembler{arch: "arm64"}, registers: map[slots.Slot]byte{0: 19, 1: 20, 2: 21, 3: 22, 4: 23}}
			pc := 0
			switch kind {
			case "metered":
				e.unmetered = false
			case "short":
				f.Code = f.Code[:2]
			case "poll":
				f.Code = append(make([]slots.Instruction, 254), f.Code...)
				pc = 254
			case "opcode":
				f.Code[0].Op = 0x6a
			case "constant":
				f.Code[1].Op = 0x42
			case "shift":
				f.Code[2].Op = 0x6a
			case "input":
				f.Code[2].A = 0
			case "count":
				f.Code[2].B = 1
			case "clobber":
				f.Code[1].Dst, f.Code[2].B = 2, 2
			case "live":
				f.Code = append(f.Code[:3], f.Code[4])
			case "spill":
				delete(e.registers, 4)
			}
			if e.logicalShift(pc, f.Code[pc]) || len(e.a.code) != 0 {
				t.Fatal("unsafe combination emitted code")
			}
		})
	}
}
