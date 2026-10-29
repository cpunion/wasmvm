package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

// Compare every slot after entering at every original PC. Inputs deliberately
// differ at cold continuations, so no preceding copy/constant may be assumed.
func checkLinearIntegerEntries(t *testing.T, f *slots.Function) {
	t.Helper()
	_, lease, c := functionContext(t, f, true)
	for _, seed := range []uint64{0, 1, 0xffffffff, 1 << 63, ^uint64(0), 0x123456789abcdef0} {
		for start := range f.Code {
			want := make([]uint64, len(f.Slots))
			for i := range want {
				want[i] = seed + uint64(i)*0x123456789abcdef
				writeSlot(t, c, 16, slots.Slot(i), value.Value{Lo: want[i]})
			}
			for _, in := range f.Code[start : len(f.Code)-1] {
				switch in.Op {
				case 1:
				case 0x20:
					want[in.Dst] = want[f.Sources[in.Source].Wasm.A]
				case 0x21, 0x22:
					v := want[in.A]
					want[f.Sources[in.Source].Wasm.A] = v
					if in.Dst != slots.NoSlot {
						want[in.Dst] = v
					}
				case 0x41, 0x42:
					want[in.Dst] = f.Sources[in.Source].Wasm.Bits
				default:
					var b uint64
					if in.B != slots.NoSlot {
						b = want[in.B]
					}
					v, trap := numeric.Eval(in.Op, want[in.A], b)
					if trap != wasm.NoTrap {
						t.Fatal(trap)
					}
					want[in.Dst] = v
				}
			}
			if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start)}); err != nil {
				t.Fatal(err)
			}
			if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
				t.Fatal(exit, err)
			}
			for i, bits := range want {
				if got := readSlot(t, c, 16, slots.Slot(i)); got != (value.Value{Lo: bits}) {
					t.Fatalf("start=%d seed=%x slot=%d: %x want %x code=%+v", start, seed, i, got, bits, f.Code)
				}
			}
		}
	}
	c.Close()
	lease.Close()
}

func TestIntegerCopyForwarding(t *testing.T) {
	if !native.Available {
		return
	}
	for _, op := range integerOps() {
		for _, clobber := range []bool{false, true} {
			for _, dst := range []slots.Slot{0, 1, 3, 4, 5} {
				f := &slots.Function{LocalCount: 3, Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64},
					Sources: []ir.Source{{Wasm: module.Instr{A: 0}}, {Wasm: module.Instr{A: 1}}, {Wasm: module.Instr{A: 2}}, {Wasm: module.Instr{Bits: 0x123456789abcdef}}}}
				e := functionEmitter{a: &assembler{arch: "arm64"}, registers: map[slots.Slot]byte{0: 19, 1: 20, 2: 21, 3: 22, 4: 23, 5: 24}}
				if !e.canRegisterInteger(op, 5, 3, 4) {
					continue
				}
				f.Code = []slots.Instruction{
					{Op: 0x20, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
					{Op: 0x20, Source: 1, Dst: 4, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
				}
				if clobber {
					f.Code = append(f.Code, slots.Instruction{Op: 0x42, Source: 3, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
				}
				f.Code = append(f.Code,
					slots.Instruction{Op: op, Dst: dst, A: 3, B: 4, C: slots.NoSlot},
					slots.Instruction{Op: 0x22, Source: 2, Dst: 6, A: dst, B: slots.NoSlot, C: slots.NoSlot},
					slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
				compiled, err := CompileUnmeteredFunction("arm64", f, DefaultLimits())
				if err != nil || compiled.ForwardedScalars != 1 {
					t.Fatal(compiled, err)
				}
				checkLinearIntegerEntries(t, f)
			}
		}
	}
}

func TestMultiplyAddForwardsLocalResult(t *testing.T) {
	if !native.Available {
		return
	}
	for _, wide := range []bool{false, true} {
		for _, swap := range []bool{false, true} {
			for local := slots.Slot(0); local < 5; local++ {
				f := multiplyAddFixture(wide, swap, 2, 3, 4, 4)
				f.LocalCount = 5
				// A shorter copy/constant group must not consume the multiply before
				// the longer multiply-add selector can see it.
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{A: 0}})
				f.Code = append([]slots.Instruction{
					{Op: 0x20, Source: 2, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
					{Op: f.Sources[1].Wasm.Op, Source: 1, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
				}, f.Code...)
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{A: uint32(local)}})
				last := f.Code[len(f.Code)-1]
				f.Code[len(f.Code)-1] = slots.Instruction{Op: 0x21, Source: 3, Dst: slots.NoSlot, A: 3, B: slots.NoSlot, C: slots.NoSlot}
				f.Code = append(f.Code, last)
				compiled, err := CompileUnmeteredFunction("arm64", f, DefaultLimits())
				if err != nil || compiled.FusedIntegers != 1 {
					t.Fatal("copy pass suppressed multiply-add", compiled, err)
				}
				checkLinearIntegerEntries(t, f)
			}
		}
	}
}
