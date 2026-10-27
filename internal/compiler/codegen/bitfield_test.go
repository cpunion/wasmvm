package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func maskShiftFixture(wide bool, mask, amount uint64, multiply, assign bool, result slots.Slot) *slots.Function {
	constant, and, shift, typ := uint32(0x41), uint32(0x71), uint32(0x74), wasm.I32
	if wide {
		constant, and, shift, typ = 0x42, 0x83, 0x86, wasm.I64
	}
	if multiply {
		shift -= 8
	}
	f := &slots.Function{LocalCount: 1, Slots: []wasm.ValueType{typ, typ, typ, typ, typ, typ}, Args: []slots.Slot{0, 1, 2, 3, 4, 5}}
	appendOp := func(op uint32, dst, a, b slots.Slot, n uint64) {
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: uint32(len(f.Sources)), Dst: dst, A: a, B: b, C: slots.NoSlot})
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: n}})
	}
	appendOp(constant, 1, slots.NoSlot, slots.NoSlot, mask)
	appendOp(and, 2, 0, 1, 0)
	appendOp(constant, 3, slots.NoSlot, slots.NoSlot, amount)
	appendOp(shift, result, 2, 3, 0)
	if assign {
		appendOp(0x22, 5, result, slots.NoSlot, 0)
	}
	// Overwrite the masked temporary unless it is the final shift result.
	// All other slots, including the two constants, remain observable.
	if result != 2 {
		appendOp(constant, 2, slots.NoSlot, slots.NoSlot, 19)
	}
	appendOp(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = 6
	return f
}

func TestMaskShiftFusionContinuations(t *testing.T) {
	for _, wide := range []bool{false, true} {
		width := uint(32)
		if wide {
			width = 64
		}
		for _, length := range []uint{1, 13, width - 1, width} {
			mask := ^uint64(0) >> (64 - length)
			for _, count := range []uint64{0, 2, uint64(width - 1), uint64(width), 127} {
				for _, multiply := range []bool{false, true} {
					amount := count
					if multiply {
						amount = uint64(1) << (count & uint64(width-1))
					}
					for _, result := range []slots.Slot{0, 1, 2, 3, 4} {
						f := maskShiftFixture(wide, mask, amount, multiply, result%2 == 0, result)
						c, err := CompileUnmeteredFunction("arm64", f, DefaultLimits())
						if err != nil || c.FusedIntegers != 1 {
							t.Fatal(wide, length, count, multiply, result, "missing fusion", err)
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

func TestMaskShiftProofBoundaries(t *testing.T) {
	for _, kind := range []string{"live-mask", "bad-mask", "bad-multiplier", "poll", "source-clobber", "amount-clobber"} {
		f := maskShiftFixture(false, 8191, 4, true, false, 4)
		switch kind {
		case "live-mask":
			f.Code = append(f.Code[:4], f.Code[len(f.Code)-1])
		case "bad-mask":
			f.Sources[0].Wasm.Bits = 0x1555
		case "bad-multiplier":
			f.Sources[2].Wasm.Bits = 3
		case "source-clobber":
			f.Code[1].A = f.Code[0].Dst
		case "amount-clobber":
			f.Code[2].Dst = f.Code[1].Dst
			f.Code[3].B = f.Code[2].Dst
		case "poll":
			prefix := make([]slots.Instruction, 254)
			for i := range prefix {
				prefix[i] = slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
			}
			f.Code = append(prefix, f.Code...)
		}
		for _, arch := range []string{"arm64", "amd64"} {
			c, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
			if err != nil || c.FusedIntegers != 0 {
				t.Fatal(kind, arch, err)
			}
		}
		if native.Available {
			checkLinearIntegerEntries(t, f)
		}
	}
}
