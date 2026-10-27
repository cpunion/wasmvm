package codegen

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMultiplyAddConstantOperands(t *testing.T) {
	if !native.Available {
		return
	}
	for _, wide := range []bool{false, true} {
		for _, swap := range []bool{false, true} {
			for product := slots.Slot(0); product < 3; product++ {
				for result := slots.Slot(0); result < 5; result++ {
					for constant := 0; constant < 5; constant++ {
						t.Run(fmt.Sprintf("wide=%v/swap=%v/product=%d/result=%d/constant=%d", wide, swap, product, result, constant), func(t *testing.T) {
							f := multiplyAddFixture(wide, swap, product, result, slots.Slot(constant), constant)
							f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Bits: 0x12345678}})
							operand := slots.Slot(0)
							if swap {
								operand = 1
							}
							if result == 4 {
								f.Code[0].A, f.Code[0].B = operand, operand
							}
							f.Code = append([]slots.Instruction{{Op: f.Sources[1].Wasm.Op, Source: 2, Dst: operand, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}, f.Code...)
							checkLinearIntegerEntries(t, f)
						})
					}
				}
			}
		}
	}
}

func multiplyAddFixture(wide, swap bool, product, result, addend slots.Slot, constant int) *slots.Function {
	mul, literal, bits := uint32(0x6c), uint32(0x41), uint64(0xfedcba98)
	if wide {
		mul, literal, bits = 0x7e, 0x42, 0xfedcba9876543210
	}
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64},
		Args: []slots.Slot{0, 1, 2, 3, 4}, Sources: []ir.Source{{}, {Wasm: module.Instr{Op: literal, Bits: bits}}}}
	f.Code = append(f.Code, slots.Instruction{Op: mul, Dst: product, A: 0, B: 1, C: slots.NoSlot})
	if constant >= 0 {
		f.Code = append(f.Code, slots.Instruction{Op: literal, Source: 1, Dst: slots.Slot(constant), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	l, r := product, addend
	if swap {
		l, r = r, l
	}
	f.Code = append(f.Code, slots.Instruction{Op: mul - 2, Dst: result, A: l, B: r, C: slots.NoSlot},
		slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: 5, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	return f
}

func TestMultiplyAddSlotsAndContinuations(t *testing.T) {
	// Cover destination/input aliases, commuted adds, constant clobbers, and
	// deliberate non-fusion cases. Every slot remains observable, not just sum.
	aliases := [][4]int{{2, 3, 4, -1}, {2, 2, 4, -1}, {0, 3, 4, -1}, {1, 3, 4, -1},
		{2, 4, 4, -1}, {2, 3, 0, -1}, {2, 3, 1, -1}, {2, 3, 1, 1}, {2, 3, 0, 0},
		{2, 3, 4, 4}, {2, 3, 4, 2}, {2, 3, 2, -1}, {0, 0, 1, 1}, {1, 1, 0, 0}, {0, 1, 4, 1}}
	for _, wide := range []bool{false, true} {
		for _, swap := range []bool{false, true} {
			for _, alias := range aliases {
				for _, overwrite := range []bool{false, true} {
					p, d, c, k := alias[0], alias[1], alias[2], alias[3]
					f := multiplyAddFixture(wide, swap, slots.Slot(p), slots.Slot(d), slots.Slot(c), k)
					if overwrite {
						last := f.Code[len(f.Code)-1]
						f.Code[len(f.Code)-1] = slots.Instruction{Op: f.Sources[1].Wasm.Op, Source: 1, Dst: slots.Slot(p), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
						f.Code = append(f.Code, last)
					}
					compiled, err := CompileUnmeteredFunction("arm64", f, DefaultLimits())
					wantFused := p != c && p != k
					if err != nil || (compiled.FusedIntegers == 1) != wantFused {
						t.Fatal(alias, compiled, err)
					}
					if !native.Available {
						continue
					}
					_, lease, context := functionContext(t, f, true)
					for _, seed := range []uint64{0, 1, 0xffffffff, 1 << 63, ^uint64(0), 0x123456789abcdef0} {
						for start := 0; start < len(f.Code); start++ {
							var want [5]uint64
							for i := range want {
								want[i] = seed + uint64(i)*0x123456789abcdef
								writeSlot(t, context, 16, slots.Slot(i), value.Value{Lo: want[i]})
							}
							for _, in := range f.Code[start : len(f.Code)-1] {
								if in.Op == 0x41 || in.Op == 0x42 {
									want[in.Dst] = f.Sources[in.Source].Wasm.Bits
								} else {
									want[in.Dst], _ = numeric.Eval(in.Op, want[in.A], want[in.B])
								}
							}
							if err := context.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start)}); err != nil {
								t.Fatal(err)
							}
							if exit, err := context.Enter(lease, 0); err != nil || exit != native.Return {
								t.Fatal(runtime.GOARCH, exit, err)
							}
							for i, bits := range want {
								if got := readSlot(t, context, 16, slots.Slot(i)); got != (value.Value{Lo: bits}) {
									t.Fatalf("wide=%v swap=%v alias=%v start=%d seed=%x slot=%d: %x want=%x", wide, swap, alias, start, seed, i, got, bits)
								}
							}
						}
					}
					context.Close()
					lease.Close()
				}
			}
		}
	}
}

func TestMultiplyAddRetainsProductAcrossTrap(t *testing.T) {
	if !native.Available {
		return
	}
	f := multiplyAddFixture(false, false, 2, 3, 4, 4)
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x41}})
	last := f.Code[len(f.Code)-1]
	f.Code = append(f.Code[:len(f.Code)-1],
		slots.Instruction{Op: 0x41, Source: 2, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		slots.Instruction{Op: 0x6d, Dst: 2, A: 1, B: 0, C: slots.NoSlot}, last)
	compiled, lease, c := functionContext(t, f, true)
	_, exit, state := runFunction(t, compiled, lease, c, []value.Value{{Lo: 6}, {Lo: 7}}, 0, false, 64)
	if exit != native.Trap || state.Payload[0] != uint64(wasm.TrapIntegerDivideByZero) || readSlot(t, c, 16, 2).Lo != 42 {
		t.Fatal("product lost before trapping overwrite", exit, state, readSlot(t, c, 16, 2))
	}
}

func TestDeadTemporaryBoundaries(t *testing.T) {
	constant := slots.Instruction{Op: 0x41, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
	for _, tc := range []struct {
		op          uint32
		read, write slots.Slot
		local       uint32
		want        bool
	}{
		{1, slots.NoSlot, slots.NoSlot, 0, true},
		{0x20, slots.NoSlot, 3, 2, false},
		{0x21, 3, slots.NoSlot, 2, true},
		{0x6a, 2, 2, 0, false},
		{0x6d, 3, 2, 0, false},
		{0x28, 3, 2, 0, false},
		{slots.ControlBase + uint32(ir.Return), slots.NoSlot, slots.NoSlot, 0, false},
	} {
		f := &slots.Function{Sources: []ir.Source{{Wasm: module.Instr{A: tc.local}}},
			Code: []slots.Instruction{{}, {Op: tc.op, Dst: tc.write, A: tc.read, B: slots.NoSlot, C: slots.NoSlot}, constant}}
		e := functionEmitter{f: f}
		if got := e.overwrittenBeforeExit(1, 2); got != tc.want || e.overwrittenBeforeExit(0, 2) {
			t.Fatal(tc, got)
		}
	}
}

func TestMultiplyAddPreservesPollAndFuelBoundaries(t *testing.T) {
	f := multiplyAddFixture(false, false, 2, 3, 4, 4)
	prefix := make([]slots.Instruction, 254)
	for i := range prefix {
		prefix[i] = slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
	}
	f.Code = append(prefix, f.Code...)
	for _, compile := range []func(string, *slots.Function, Limits) (*CompiledFunction, error){CompileFunction, CompileUnmeteredFunction} {
		compiled, err := compile("arm64", f, DefaultLimits())
		if err != nil || compiled.FusedIntegers != 0 {
			t.Fatal("crossed an instruction boundary", compiled, err)
		}
	}
}
