package codegen

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	enginememory "github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func memoryForwardFixture(store, load uint32, delta uint32, copies bool) *slots.Function {
	source, result := wasm.I32, wasm.I32
	if store == 0x37 || store >= 0x3c {
		source = wasm.I64
	}
	if load == 0x29 || load >= 0x30 {
		result = wasm.I64
	}
	f := &slots.Function{LocalCount: 2, Slots: []wasm.ValueType{wasm.I32, source, wasm.I32, source, wasm.I32, result, wasm.I32, wasm.I32}}
	for i, typ := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
		f.Signature.Results = append(f.Signature.Results, typ)
	}
	emit := func(op uint32, dst, a, b slots.Slot, local, offset uint32) {
		s := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, B: offset, Offset: uint64(100 + s)}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: s, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	address, data := slots.Slot(0), slots.Slot(1)
	emit(0x41, 2, slots.NoSlot, slots.NoSlot, 0, 0)
	f.Sources[len(f.Sources)-1].Wasm.Bits = 63
	emit(0x71, 0, 0, 2, 0, 0)
	if copies {
		emit(0x20, 4, slots.NoSlot, slots.NoSlot, 0, 0)
		emit(0x20, 3, slots.NoSlot, slots.NoSlot, 1, 0)
		address, data = 4, 3
	}
	emit(store, slots.NoSlot, address, data, 0, 3)
	if copies {
		emit(0x20, 7, slots.NoSlot, slots.NoSlot, 0, 0)
		emit(1, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
		emit(0x1a, slots.NoSlot, 7, slots.NoSlot, 0, 0)
		emit(0x20, 0, slots.NoSlot, slots.NoSlot, 0, 0)
		emit(0x20, 6, slots.NoSlot, slots.NoSlot, 0, 0)
		address = 6
	}
	emit(load, 5, address, slots.NoSlot, 0, 3+delta)
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
	return f
}

func TestPrivateMemoryForwardSpilledOperands(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for mask := 0; mask < 4; mask++ {
			a := assembler{arch: arch, labels: map[int]int{}}
			f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I32}}
			e := functionEmitter{a: &a, f: f, unmetered: true, exit: a.label()}
			e.contextLoad(0, nativeabi.ArenaOffset)
			frame, homes := byte(6), []byte{3, 12}
			if arch == "arm64" {
				frame, homes = 12, []byte{19, 20}
			}
			e.registerMove(frame, e.operandRegister(0), true)
			plan := &registerPlan{homes: map[slots.Slot]byte{}}
			for s, reg := range homes {
				if mask&(1<<s) != 0 {
					plan.homes[slots.Slot(s)] = reg
					plan.slots = append(plan.slots, slots.Slot(s))
				}
			}
			e.transitionRegisters(plan)
			e.forwardMemoryValue(0, 1, 40, 1, true)
			e.leave(nativeabi.Return, 1, 1)
			a.mark(e.exit)
			body, err := a.finish()
			if err != nil {
				t.Fatal(err)
			}
			entry, err := nativeabi.Wrap(arch, body, 1<<16)
			if err != nil {
				t.Fatal(err)
			}
			if !native.Available || runtime.GOARCH != arch {
				continue
			}
			code, err := native.Publish(entry.Code, []uint32{0}, 1<<16)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(code.Close)
			lease, err := code.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			ctx, err := native.NewContext(32, 32, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			for _, bits := range []uint64{0, ^uint64(0), 1 << 63, 0x0000800000000000, 0x00007f0000000000, 0x0123abcd567889ef} {
				writeSlot(t, ctx, 0, 0, value.Value{Lo: bits, Hi: 123})
				writeSlot(t, ctx, 0, 1, value.Value{Lo: ^bits, Hi: 456})
				if err := ctx.SetState(native.State{Depth: 1}); err != nil {
					t.Fatal(err)
				}
				exit, err := ctx.Enter(lease, 0)
				want := uint64(uint32(int32(int8(bits >> 40))))
				if err != nil || exit != native.Return || readSlot(t, ctx, 0, 0) != (value.Value{Lo: bits, Hi: 123}) || readSlot(t, ctx, 0, 1) != (value.Value{Lo: want, Hi: 456}) {
					t.Fatal(arch, mask, bits, exit, err, readSlot(t, ctx, 0, 0), readSlot(t, ctx, 0, 1))
				}
			}
		}
	}
}

func TestPrivateMemoryForwardBarriers(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, tc := range []struct {
			name string
			edit func(*functionEmitter)
		}{
			{"metered", func(e *functionEmitter) { e.unmetered = false }},
			{"shared", func(e *functionEmitter) { e.sharedDefaultMemory = true }},
			{"views", func(e *functionEmitter) { e.memoryViews = true }},
			{"float-store", func(e *functionEmitter) { e.f.Code[2].Op = 0x38 }},
			{"float-load", func(e *functionEmitter) { e.f.Code[3].Op = 0x2a }},
			{"different-address", func(e *functionEmitter) { e.f.Code[3].A = 4 }},
			{"larger-load", func(e *functionEmitter) { e.f.Code[3].Op = 0x29 }},
			{"before-store", func(e *functionEmitter) { e.f.Sources[e.f.Code[3].Source].Wasm.B = 2 }},
			{"past-store", func(e *functionEmitter) { e.f.Sources[e.f.Code[3].Source].Wasm.B = 7 }},
			{"partial-overlap", func(e *functionEmitter) { e.f.Sources[e.f.Code[3].Source].Wasm.B = 4 }},
			{"high-offset", func(e *functionEmitter) { e.f.Sources[e.f.Code[3].Source].Wasm.MemoryOffsetHigh = 1 }},
			{"mutation", func(e *functionEmitter) { e.f.Code[3].Op = 0x36 }},
			{"trapping-operation", func(e *functionEmitter) { e.f.Code[3].Op = 0x6d }},
			{"helper", func(e *functionEmitter) { e.f.Code[3].Op = 0x40 }},
			{"region", func(e *functionEmitter) { e.regionLimit = 3 }},
			{"frame-copy", func(e *functionEmitter) { e.frameCopyLimit = 3 }},
			{"allocation", func(e *functionEmitter) { e.allocationLimit = 3 }},
			{"native-region", func(e *functionEmitter) { e.nativeRegionLimit = 3 }},
			{"poll", func(e *functionEmitter) {
				e.f.Code = append(make([]slots.Instruction, 253), e.f.Code[2:]...)
				e.start = 255
			}},
			{"bounded-scan", func(e *functionEmitter) {
				tail := append([]slots.Instruction(nil), e.f.Code[3:]...)
				e.f.Code = e.f.Code[:3]
				for range 8 {
					e.f.Code = append(e.f.Code, slots.Instruction{Op: 1})
				}
				e.f.Code = append(e.f.Code, tail...)
			}},
			{"source-reuse", func(e *functionEmitter) {
				e.registerPlan = &registerPlan{shared: true, slots: []slots.Slot{1, 3}, homes: map[slots.Slot]byte{1: 3, 3: 3}}
			}},
			{"aliased-source-reuse", func(e *functionEmitter) {
				e.frameAliases = map[slots.Slot]slots.Slot{1: 3}
				e.registerPlan = &registerPlan{shared: true, slots: []slots.Slot{3, 4}, homes: map[slots.Slot]byte{3: 3, 4: 3}}
			}},
			{"clobber-source", func(e *functionEmitter) {
				e.f.Code[3].Op, e.f.Code[3].Dst = 0x20, 1
				e.f.Sources[e.f.Code[3].Source].Wasm.A = 0
			}},
			{"clobber-address", func(e *functionEmitter) {
				e.f.Code[3].Op, e.f.Code[3].Dst = 0x20, 0
				e.f.Sources[e.f.Code[3].Source].Wasm.A = 1
			}},
			{"clobber-anchor", func(e *functionEmitter) {
				e.frameAliases = map[slots.Slot]slots.Slot{1: 3}
				e.f.Code[3].Op, e.f.Code[3].Dst = 0x20, 3
				e.f.Sources[e.f.Code[3].Source].Wasm.A = 0
			}},
		} {
			t.Run(fmt.Sprintf("%s/%s", arch, tc.name), func(t *testing.T) {
				a := assembler{arch: arch, labels: map[int]int{}}
				e := functionEmitter{a: &a, f: memoryForwardFixture(0x36, 0x28, 0, false), unmetered: true, start: 2}
				tc.edit(&e)
				if _, ok := e.forwardMemoryStoreLoad(e.start, e.f.Code[e.start]); ok || len(a.code) != 0 {
					t.Fatal("unsafe selection emitted code", ok, len(a.code))
				}
			})
		}
	}
}

func TestPrivateMemoryForwardWidthsColdEntriesAndFrames(t *testing.T) {
	for _, store := range []uint32{0x36, 0x37, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e} {
		_, stored := memoryAccess(store)
		for _, load := range []uint32{0x28, 0x29, 0x2c, 0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35} {
			_, size := memoryAccess(load)
			if size > stored {
				continue
			}
			for _, delta := range []uint32{0, stored - size} {
				for _, copies := range []bool{false, true} {
					for _, memory64 := range []bool{false, true} {
						for _, alias := range []bool{false, true} {
							f := memoryForwardFixture(store, load, delta, copies)
							if memory64 {
								f.Code = f.Code[2:]
								for _, slot := range []int{0, 4, 6, 7} {
									f.Slots[slot], f.Signature.Results[slot] = wasm.I64, wasm.I64
								}
							}
							if alias {
								result := f.Slots[5]
								if result == f.Slots[0] {
									f.Code[len(f.Code)-2].Dst = 0
								} else if result == f.Slots[1] {
									f.Code[len(f.Code)-2].Dst = 1
								} else {
									continue
								}
							}
							for _, arch := range []string{"amd64", "arm64"} {
								c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
								// Unknown memory64 addresses do not use prefix-copy
								// remapping, which requires an i32 bounds proof.
								wantForward := !memory64 || !copies
								if err != nil || (c.ForwardedMemory == 1) != wantForward {
									t.Fatal("selection", arch, store, load, delta, copies, memory64, c.ForwardedMemory, err)
								}
							}
							if !native.Available {
								continue
							}
							_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
							mem := bindMemory(t, ctx, 65536, 65536)
							ref, err := enginememory.New(module.Limits{Min: 1, Index64: memory64}, 1)
							if err != nil {
								t.Fatal(err)
							}
							initial := []byte{0x81, 0x7f, 0xff, 0x80, 0, 0x12, 0x45, 0xfe, 3, 5, 9, 13, 17, 0x80, 0xfe, 0xff}
							for pc := range f.Code {
								for _, address := range []uint64{0, 65535, 0xffffffff, 0xfeedabcd00000000} {
									for _, bits := range []uint64{0, 0xff80017f80ff7f81, 0x80000000ffffffff} {
										mem.Write(0, initial)
										ref.Write(0, initial)
										words := make([]value.Value, len(f.Slots))
										for i := range words {
											words[i] = value.Value{Lo: address, Hi: uint64(i + 1000)}
										}
										words[1].Lo, words[3].Lo = bits, bits
										for i, v := range words {
											writeSlot(t, ctx, 16, slots.Slot(i), v)
										}
										trap, stop := wasm.NoTrap, pc
										for ; stop < len(f.Code); stop++ {
											in := f.Code[stop]
											src := f.Sources[in.Source].Wasm
											if in.Op == 0x41 {
												words[in.Dst].Lo = uint64(uint32(src.Bits))
											} else if in.Op == 0x71 {
												words[in.Dst].Lo = uint64(uint32(words[in.A].Lo) & uint32(words[in.B].Lo))
											} else if in.Op == 0x20 {
												words[in.Dst].Lo = words[src.A].Lo
											} else if isMemoryAccess(in.Op) {
												if load, _ := memoryAccess(in.Op); load {
													v, failed := ref.Load(in.Op, effectiveAddress(f, in.A, words), src.MemoryOffset(), 0, value.Value{})
													trap = failed
													if trap == wasm.NoTrap {
														words[in.Dst].Lo = v.Lo
													}
												} else {
													trap = ref.Store(in.Op, effectiveAddress(f, in.A, words), src.MemoryOffset(), 0, words[in.B])
												}
											}
											if trap != wasm.NoTrap || in.Op >= slots.ControlBase {
												break
											}
										}
										ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Fuel: 99, Budget: 77})
										exit, err := ctx.Enter(lease, 0)
										state, stateErr := ctx.State()
										if err != nil || stateErr != nil || state.Fuel != 99 || state.Budget != 77 || trap == wasm.NoTrap && exit != native.Return || trap != wasm.NoTrap && (exit != native.Trap || state.Source != f.Sources[f.Code[stop].Source].Wasm.Offset || state.Payload[0] != uint64(trap)) {
											t.Fatal("exit", store, load, delta, copies, pc, address, bits, exit, state, err, stateErr)
										}
										for i, want := range words {
											if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
												t.Fatal("frame", store, load, delta, copies, pc, address, bits, i, got, want)
											}
										}
										got, want := make([]byte, 80), make([]byte, 80)
										mem.Read(0, got)
										ref.Read(0, want)
										if !bytes.Equal(got, want) {
											t.Fatal("memory", store, load, delta, pc, address, got, want)
										}
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
	}
}

func effectiveAddress(f *slots.Function, s slots.Slot, words []value.Value) uint64 {
	address := words[s].Lo
	if f.Slots[s] == wasm.I32 {
		address = uint64(uint32(address))
	}
	return address
}
