package codegen

import (
	"bytes"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	enginememory "github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestDirectMemoryOperandsEveryColdFrameAndEffect(t *testing.T) {
	for _, pair := range [][2]uint32{{0x36, 0x28}, {0x37, 0x29}, {0x38, 0x2a}, {0x39, 0x2b}, {0x3a, 0x2c}, {0x3b, 0x2e}, {0x3c, 0x30}, {0x3d, 0x32}, {0x3e, 0x34}} {
		typ := wasm.I32
		if pair[0] == 0x37 || pair[0] >= 0x3c {
			typ = wasm.I64
		} else if pair[0] == 0x38 {
			typ = wasm.F32
		} else if pair[0] == 0x39 {
			typ = wasm.F64
		}
		for _, constant := range []bool{false, true} {
			for _, shared := range []bool{false, true} {
				for _, offset := range []uint32{0, 127, 128, 4095} {
					f := &slots.Function{LocalCount: 3, Slots: []wasm.ValueType{wasm.I32, typ, wasm.I32, wasm.I32, wasm.I32, wasm.I32, typ, typ, wasm.I32}}
					for s, typ := range f.Slots {
						f.Args = append(f.Args, slots.Slot(s))
						f.Signature.Results = append(f.Signature.Results, typ)
					}
					emit := func(op uint32, dst, a, b slots.Slot, local uint32, bits uint64) {
						source := uint32(len(f.Sources))
						f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, B: offset, Bits: bits, Offset: uint64(source) + 100}})
						f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
					}
					emit(0x20, 3, slots.NoSlot, slots.NoSlot, 0, 0)
					emit(0x41, 4, slots.NoSlot, slots.NoSlot, 0, 255)
					if constant {
						emit(0x41, 5, slots.NoSlot, slots.NoSlot, 0, 3)
					} else {
						emit(0x71, 5, 3, 4, 0, 0)
					}
					emit(0x22, 5, 5, slots.NoSlot, 2, 0)
					emit(0x20, 6, slots.NoSlot, slots.NoSlot, 1, 0)
					emit(pair[0], slots.NoSlot, 5, 6, 0, 0)
					emit(0x20, 8, slots.NoSlot, slots.NoSlot, 2, 0)
					emit(pair[1], 7, 8, slots.NoSlot, 0, 0)
					emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
					f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
					config := Config{Unmetered: true, MemoryMinimumPages: 1, NativeThreads: true, MemoryCount: 1, SharedMemories: []bool{shared}}
					for _, arch := range []string{"amd64", "arm64"} {
						c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
						if err != nil {
							t.Fatal(arch, pair, offset, constant, err)
						}
						if arch == "amd64" && c.DirectMemoryOperands == 0 || arch == "arm64" && c.DirectMemoryOperands != 0 {
							t.Fatal(arch, pair, offset, constant, "wrong direct selection", c.DirectMemoryOperands)
						}
					}
					if !native.Available {
						continue
					}
					_, lease, ctx := configuredContext(t, f, config)
					var mem *native.Memory
					var view uint64
					if shared {
						mem, view = bindSharedTestMemory(t, ctx, 65536, 65536)
					} else {
						mem = bindMemory(t, ctx, 65536, 65536)
					}
					ref, err := enginememory.New(module.Limits{Min: 1}, 1)
					if err != nil {
						t.Fatal(err)
					}
					initial := make([]byte, 65536)
					for i := range initial {
						initial[i] = byte(i*29 + 131)
					}
					for start := range f.Code {
						for _, bits := range []uint64{0, 0xffffffff, 1 << 63, 0xfff123458001abcd} {
							mem.Write(0, initial)
							ref.Write(0, initial)
							words := make([]value.Value, len(f.Slots))
							for s, typ := range f.Slots {
								words[s] = value.Value{Lo: bits ^ uint64(s), Hi: 0x9876543200000000 | uint64(s)}
								if typ == wasm.F32 {
									words[s].Lo = uint64(uint32(words[s].Lo))
								}
								writeSlot(t, ctx, 16, slots.Slot(s), words[s])
							}
							trap, pc := wasm.NoTrap, start
							for ; pc < len(f.Code); pc++ {
								in := f.Code[pc]
								src := f.Sources[in.Source].Wasm
								switch in.Op {
								case 0x20:
									words[in.Dst].Lo = words[src.A].Lo
								case 0x22:
									words[src.A].Lo = words[in.A].Lo
									words[in.Dst].Lo = words[in.A].Lo
								case 0x41:
									words[in.Dst].Lo = uint64(uint32(src.Bits))
								case 0x71:
									words[in.Dst].Lo = uint64(uint32(words[in.A].Lo) & uint32(words[in.B].Lo))
								default:
									if isMemoryAccess(in.Op) {
										address := uint64(uint32(words[in.A].Lo))
										if load, _ := memoryAccess(in.Op); load {
											v, failed := ref.Load(in.Op, address, src.MemoryOffset(), 0, value.Value{})
											trap = failed
											if trap == wasm.NoTrap {
												words[in.Dst].Lo = v.Lo
											}
										} else {
											trap = ref.Store(in.Op, address, src.MemoryOffset(), 0, words[in.B])
										}
									}
								}
								if trap != wasm.NoTrap || in.Op >= slots.ControlBase {
									break
								}
							}
							state := native.State{Used: 16, Depth: 1, Continuation: uint64(start), Fuel: 42, Budget: 37}
							state.Payload[2] = view
							if err := ctx.SetState(state); err != nil {
								t.Fatal(err)
							}
							exit, err := ctx.Enter(lease, 0)
							state, stateErr := ctx.State()
							if err != nil || stateErr != nil || state.Fuel != 42 || state.Budget != 37 || trap == wasm.NoTrap && (exit != native.Return || state.Continuation != uint64(pc) || state.Payload[0] != uint64(pc)) || trap != wasm.NoTrap && (exit != native.Trap || state.Source != f.Sources[f.Code[pc].Source].Wasm.Offset || state.Payload[0] != uint64(trap)) {
								t.Fatal(pair, offset, constant, shared, start, bits, exit, state, err, stateErr)
							}
							for s, want := range words {
								if got := readSlot(t, ctx, 16, slots.Slot(s)); got != want {
									t.Fatal(pair, offset, constant, shared, start, bits, "word", s, got, want)
								}
							}
							got, want := make([]byte, 65536), make([]byte, 65536)
							mem.Read(0, got)
							ref.Read(0, want)
							if !bytes.Equal(got, want) {
								t.Fatal(pair, offset, constant, shared, start, bits, "memory effect")
							}
						}
					}
					ctx.Close()
					lease.Close()
					mem.Close()
				}
			}
		}
	}
}
