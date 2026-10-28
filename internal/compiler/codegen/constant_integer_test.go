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

func TestConstantIntegerAliasesAndEntries(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, operation := range []uint32{0, 1, 2, 7, 8, 9, 10, 11, 12, 13, 14} {
			for _, bits := range []uint64{0, 127, 128, 0x80000000, 0xffffffff, 1 << 32, ^uint64(0), ^uint64(128)} {
				for dst := slots.Slot(0); dst < 3; dst++ {
					t.Run(fmt.Sprintf("wide=%v/op=%d/bits=%x/dst=%d", wide, operation, bits, dst), func(t *testing.T) {
						typ, literal, first, constant := wasm.I32, uint32(0x41), uint32(0x6a), bits
						if wide {
							typ, literal, first = wasm.I64, 0x42, 0x7c
						} else {
							constant = uint64(uint32(bits))
						}
						f := &slots.Function{Slots: []wasm.ValueType{typ, typ, typ}, Sources: []ir.Source{{Wasm: module.Instr{Bits: constant}}, {}}, Code: []slots.Instruction{
							{Op: literal, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
							{Op: first + operation, Dst: dst, A: 0, B: 1, C: slots.NoSlot},
							{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
						}}
						// Alternate live constants and dead temporaries; returning
						// every backing slot checks both optimized and fallback paths.
						if dst == 2 {
							f.Code = append(f.Code[:2], slots.Instruction{Op: literal, Source: 1, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}, f.Code[2])
						}
						compiled, err := CompileUnmeteredFunction("amd64", f, DefaultLimits())
						fused := !wide || operation >= 10 || uint64(int64(int32(bits))) == bits
						if err != nil || (compiled.FusedIntegers == 1) != fused {
							t.Fatal(compiled, err)
						}
						if native.Available {
							checkLinearIntegerEntries(t, f)
						}
					})
				}
			}
		}
	}
}

func TestIntegerConstantPollBoundary(t *testing.T) {
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}, Sources: []ir.Source{{Wasm: module.Instr{Bits: 7}}}}
	for range 255 {
		f.Code = append(f.Code, slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	f.Code = append(f.Code,
		slots.Instruction{Op: 0x41, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		slots.Instruction{Op: 0x6c, Dst: 0, A: 0, B: 1, C: slots.NoSlot},
		slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	for _, compile := range []func(string, *slots.Function, Limits) (*CompiledFunction, error){CompileFunction, CompileUnmeteredFunction} {
		c, err := compile("amd64", f, DefaultLimits())
		if err != nil || c.FusedIntegers != 0 {
			t.Fatal(c, err)
		}
	}
}

func TestConstantIntegerLocalForwarding(t *testing.T) {
	if !native.Available {
		return
	}
	for _, wide := range []bool{false, true} {
		for _, operation := range []uint32{0, 1, 2, 7, 9, 13} {
			for local := uint32(0); local < 4; local++ {
				for _, prefix := range []bool{false, true} {
					for _, dead := range []bool{false, true} {
						t.Run(fmt.Sprintf("wide=%v/op=%d/local=%d/prefix=%v/dead=%v", wide, operation, local, prefix, dead), func(t *testing.T) {
							typ, literal, first := wasm.I32, uint32(0x41), uint32(0x6a)
							if wide {
								typ, literal, first = wasm.I64, 0x42, 0x7c
							}
							f := &slots.Function{LocalCount: 4, Slots: []wasm.ValueType{typ, typ, typ, typ, typ}, Sources: []ir.Source{
								{Wasm: module.Instr{Bits: 0x12345678}}, {Wasm: module.Instr{A: local}}, {}, {},
							}, Code: []slots.Instruction{
								{Op: literal, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
								{Op: first + operation, Dst: 2, A: 0, B: 1, C: slots.NoSlot},
								{Op: 0x22, Source: 1, Dst: 4, A: 2, B: slots.NoSlot, C: slots.NoSlot},
							}}
							if prefix {
								f.Code[1].A = 2
								f.Code = append([]slots.Instruction{{Op: 0x20, Source: 3, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}, f.Code...)
							}
							if dead {
								f.Code = append(f.Code, slots.Instruction{Op: literal, Source: 2, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
							}
							f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
							checkLinearIntegerEntries(t, f)
						})
					}
				}
			}
		}
	}
}
