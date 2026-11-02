package codegen

import (
	"fmt"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestI31ForwardingSlotsAndIndependentEntries(t *testing.T) {
	for _, signed := range []bool{false, true} {
		for _, alias := range [][3]slots.Slot{{0, 1, 2}, {0, 0, 1}, {0, 1, 0}, {0, 1, 1}, {0, 0, 0}, {0, 31, 33}} {
			for _, overwrite := range []bool{false, true} {
				t.Run(fmt.Sprintf("signed=%v/slots=%v/overwrite=%v", signed, alias, overwrite), func(t *testing.T) {
					get := uint32(0xfb001e)
					if signed {
						get = 0xfb001d
					}
					f := &slots.Function{Slots: make([]wasm.ValueType, 34), Sources: []ir.Source{{Wasm: module.Instr{Offset: 77}}, {Wasm: module.Instr{Bits: 42}}}}
					for i := range f.Slots {
						f.Slots[i] = wasm.I64
						f.Args = append(f.Args, slots.Slot(i))
					}
					f.Code = []slots.Instruction{{Op: 0xfb001c, A: alias[0], Dst: alias[1], B: slots.NoSlot, C: slots.NoSlot}, {Op: get, A: alias[1], Dst: alias[2], B: slots.NoSlot, C: slots.NoSlot}}
					if overwrite {
						f.Code = append(f.Code, slots.Instruction{Op: 0x42, Source: 1, Dst: alias[1], A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
					}
					f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: uint32(len(f.Args)), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
					for _, arch := range []string{"amd64", "arm64"} {
						compiled, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
						if err != nil || compiled.FusedIntegers != 1 {
							t.Fatal(arch, compiled, err)
						}
					}
					if !native.Available {
						return
					}
					for _, fast := range []bool{false, true} {
						_, lease, c := functionContext(t, f, fast)
						for _, seed := range []uint64{0, 1, 0x3fffffff, 0x40000000, 0x7fffffff, 0x80000000, 0xffffffff, ^uint64(0)} {
							for start := range f.Code {
								want := make([]uint64, len(f.Slots))
								for i := range want {
									if seed != 0 {
										want[i] = seed + uint64(i)*0x123456789abcdef
									}
									writeSlot(t, c, 16, slots.Slot(i), value.Value{Lo: want[i]})
								}
								trap := false
								for _, in := range f.Code[start : len(f.Code)-1] {
									switch in.Op {
									case 0xfb001c:
										want[in.Dst] = 1<<62 | want[in.A]&0x7fffffff
									case 0xfb001d, 0xfb001e:
										if want[in.A] == 0 {
											trap = true
											break
										}
										bits := uint32(want[in.A]) & 0x7fffffff
										if in.Op == 0xfb001d {
											bits = uint32(int32(bits<<1) >> 1)
										}
										want[in.Dst] = uint64(bits)
									case 0x42:
										want[in.Dst] = 42
									}
									if trap {
										break
									}
								}
								if err := c.SetState(native.State{Used: 16, Depth: 1, Fuel: 1000, Budget: 64, Continuation: uint64(start)}); err != nil {
									t.Fatal(err)
								}
								exit, err := c.Enter(lease, 0)
								expected := native.Return
								if trap {
									expected = native.Trap
								}
								if err != nil || exit != expected {
									t.Fatal(fast, seed, start, exit, expected, err)
								}
								if trap {
									state, err := c.State()
									if err != nil || state.Source != 77 || state.Payload[0] != uint64(wasm.TrapNullReference) {
										t.Fatal("cold-entry null trap", state, err)
									}
								}
								for i, bits := range want {
									if got := readSlot(t, c, 16, slots.Slot(i)); got != (value.Value{Lo: bits}) {
										t.Fatalf("fast=%v seed=%x entry=%d slot=%d got=%x want=%x", fast, seed, start, i, got, bits)
									}
								}
							}
						}
						c.Close()
						lease.Close()
					}
				})
			}
		}
	}
}

func TestI31ForwardingBoundaries(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, crossPoll := range []bool{false, true} {
			for _, matching := range []bool{false, true} {
				f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I64, wasm.I32}, Sources: []ir.Source{{}}}
				if crossPoll {
					for range 255 {
						f.Code = append(f.Code, slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
					}
				}
				input := slots.Slot(1)
				if !matching {
					input = 0
				}
				f.Code = append(f.Code, slots.Instruction{Op: 0xfb001c, A: 0, Dst: 1, B: slots.NoSlot, C: slots.NoSlot}, slots.Instruction{Op: 0xfb001e, A: input, Dst: 2, B: slots.NoSlot, C: slots.NoSlot}, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
				for _, fast := range []bool{false, true} {
					compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: fast})
					want := uint32(0)
					if fast && matching && !crossPoll {
						want = 1
					}
					if err != nil || compiled.FusedIntegers != want {
						t.Fatal(arch, crossPoll, matching, fast, compiled, err)
					}
				}
			}
		}
	}
}
