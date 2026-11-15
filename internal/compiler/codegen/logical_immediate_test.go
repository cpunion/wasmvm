package codegen

import (
	"math/bits"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Decode the architectural bitmask fields independently of the encoder, for
// every valid N:immr:imms combination at both instruction widths.
func TestARMLogicalImmediateEncoding(t *testing.T) {
	for _, width := range []int{32, 64} {
		for n := uint32(0); n < 2; n++ {
			for imms := uint32(0); imms < 64; imms++ {
				length := bits.Len32(n<<6|(^imms&63)) - 1
				if length < 1 || 1<<length > width {
					continue
				}
				levels := uint32(1<<length) - 1
				s := imms & levels
				if s == levels {
					continue
				}
				for immr := uint32(0); immr < 64; immr++ {
					element := uint64(0)
					for bit := uint32(0); bit <= s; bit++ {
						element |= uint64(1) << ((bit - (immr & levels)) & levels)
					}
					var value uint64
					for offset := 0; offset < width; offset += 1 << length {
						value |= element << offset
					}
					fields, ok := armLogicalImmediate(value, width)
					if !ok {
						t.Fatalf("not encoded width=%d value=%x", width, value)
					}
					nn, rr, ss := fields>>22&1, fields>>16&63, fields>>10&63
					ll := bits.Len32(nn<<6|(^ss&63)) - 1
					if ll < 1 || 1<<ll > width || ss&uint32(1<<ll-1) == uint32(1<<ll-1) {
						t.Fatal("invalid fields", fields)
					}
					var decoded uint64
					for offset := 0; offset < width; offset += 1 << ll {
						for bit := uint32(0); bit <= ss&uint32(1<<ll-1); bit++ {
							decoded |= uint64(1) << (uint32(offset) + ((bit - rr) & uint32(1<<ll-1)))
						}
					}
					if decoded != value {
						t.Fatalf("width=%d value=%x fields=%x decoded=%x", width, value, fields, decoded)
					}
				}
			}
		}
		for _, value := range []uint64{0, ^uint64(0), 0x1555, 0x12345678, 0x0102030405060708} {
			if _, ok := armLogicalImmediate(value, width); ok {
				t.Fatalf("invalid mask accepted width=%d value=%x", width, value)
			}
		}
	}
}

func TestARMLogicalImmediateColdFrames(t *testing.T) {
	for _, wide := range []bool{false, true} {
		typ, literal, first := wasm.I32, uint32(0x41), uint32(0x71)
		if wide {
			typ, literal, first = wasm.I64, 0x42, 0x83
		}
		masks := []uint64{0xff00ff, 0xff00ff00, 0x80000001, 0x55555555}
		if wide {
			masks = []uint64{0x00ff00ff00ff00ff, 0xff00ff00ff00ff00, 0x8000000000000001, 0x5555555555555555}
		}
		for _, mask := range masks {
			if !wide {
				mask = uint64(uint32(mask))
			}
			for op := first; op < first+3; op++ {
				for dst := slots.Slot(0); dst < 3; dst++ {
					f := &slots.Function{Slots: []wasm.ValueType{typ, typ, typ}, Sources: []ir.Source{{Wasm: module.Instr{Bits: mask}}}, Code: []slots.Instruction{
						{Op: literal, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
						{Op: op, Dst: dst, A: 0, B: 1, C: slots.NoSlot},
						{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
					}}
					c, err := CompileUnmeteredFunction("arm64", f, DefaultLimits())
					if err != nil || c.FusedIntegers != 1 {
						t.Fatalf("wide=%v mask=%x op=%x dst=%d fused=%d err=%v", wide, mask, op, dst, c.FusedIntegers, err)
					}
					if native.Available {
						checkLinearIntegerEntries(t, f)
					}
				}
			}
		}
	}
}
