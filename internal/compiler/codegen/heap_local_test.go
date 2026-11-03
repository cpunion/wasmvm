package codegen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeHeapForwardedLocalAndTee(t *testing.T) {
	fields := []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I32, Packed: 16}, {Value: wasm.I32}, {Value: wasm.I64}, {Value: wasm.F32}, {Value: wasm.F64}, {Value: wasm.AnyRef}, {Value: wasm.V128}}
	for _, field := range fields {
		for _, array := range []bool{false, true} {
			for _, signed := range []bool{false, true} {
				for _, local := range []slots.Slot{2, 4, 6} {
					for _, tee := range []bool{false, true} {
						for _, dead := range []bool{false, true} {
							t.Run(fmt.Sprintf("%v/array=%v/signed=%v/local=%d/tee=%v/dead=%v", field, array, signed, local, tee, dead), func(t *testing.T) {
								f, config := heapFixture(field, array, signed)
								f.Slots = append(f.Slots, field.Unpacked().Storage(), field.Unpacked().Storage())
								f.LocalCount = uint32(len(f.Slots))
								op, dst := uint32(0x21), slots.NoSlot
								if tee {
									op, dst = 0x22, 7
								}
								source := uint32(len(f.Sources))
								f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: uint32(local)}})
								suffix := []slots.Instruction{{Op: op, Source: source, A: 4, B: slots.NoSlot, C: slots.NoSlot, Dst: dst}}
								if dead {
									source = uint32(len(f.Sources))
									f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(local)}})
									suffix = append(suffix, slots.Instruction{Op: 0x20, Source: source, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: 4})
								}
								f.Code = append(f.Code[:3:3], append(suffix, f.Code[3:]...)...)
								for _, fast := range []bool{false, true} {
									config.Unmetered = fast
									for _, arch := range []string{"amd64", "arm64"} {
										c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
										if err != nil || (c.ForwardedHeap == 1) != fast {
											t.Fatal(arch, c, err)
										}
									}
									if !native.Available {
										continue
									}
									_, lease, ctx := configuredContext(t, f, config)
									storage, err := native.NewMemory(256, 256)
									if err != nil {
										t.Fatal(err)
									}
									defer storage.Close()
									buffer, _ := storage.Buffer()
									address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
									words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
									ref := uint64(0x8000000100000001)
									words[0], words[1], words[6], words[7], words[8] = 2, address+16, ref, address+128, 2
									width, offset := field.HeapBytes(), 9
									if array {
										offset = width
									}
									for start := range f.Code {
										for i := range buffer[128:] {
											buffer[128+i] = 0xa5
										}
										binary.LittleEndian.PutUint64(buffer[128+offset:], 0x89abcdef87658081)
										binary.LittleEndian.PutUint64(buffer[136+offset:], 0x123456789abcdef0)
										expected := append([]byte(nil), buffer[128:]...)
										want := []value.Value{{Lo: ref}, {Lo: 0x100000001}, {Lo: 0xffffbeef00008081}, {Lo: 33}, {Lo: 44}, {Lo: 55}, {Lo: 66}, {Lo: 77}}
										if field.Value == wasm.F32 {
											want[2].Lo = 0xffc08081 // Canonical-width NaN payload.
										}
										if field.Value == wasm.V128 {
											for i := 2; i < len(want); i++ {
												if i != 5 {
													want[i].Hi = uint64(i) * 0x123456789abcdef
												}
											}
										}
										for i, v := range want {
											writeSlot(t, ctx, 16, slots.Slot(i), v)
										}
										for _, in := range f.Code[start:] {
											switch in.Op {
											case 0xfb0005, 0xfb000e:
												var raw [16]byte
												binary.LittleEndian.PutUint64(raw[:], want[2].Lo)
												binary.LittleEndian.PutUint64(raw[8:], want[2].Hi)
												copy(expected[offset:offset+width], raw[:width])
											case 0xfb0002, 0xfb0003, 0xfb0004, 0xfb000b, 0xfb000c, 0xfb000d:
												v := value.Value{Lo: binary.LittleEndian.Uint64(expected[offset:])}
												if width < 8 {
													v.Lo &= (uint64(1) << uint(width*8)) - 1
												}
												if width == 16 {
													v.Hi = binary.LittleEndian.Uint64(expected[offset+8:])
												}
												if signed && field.Packed != 0 {
													v.Lo = uint64(uint32(int32(uint32(v.Lo)<<(32-field.Packed)) >> (32 - field.Packed)))
												}
												want[in.Dst] = v
											case 0xfb000f:
												want[in.Dst] = value.Value{Lo: 2}
											case 0x21, 0x22:
												v := want[in.A]
												want[f.Sources[in.Source].Wasm.A] = v
												if in.Dst != slots.NoSlot {
													want[in.Dst] = v
												}
											case 0x20:
												want[in.Dst] = want[f.Sources[in.Source].Wasm.A]
											}
										}
										state := native.State{Used: 16, Depth: 1, Budget: 64, Fuel: 1000, Continuation: uint64(start)}
										state.Payload[2] = address
										if !fast {
											state.Payload[3] = 1
										}
										if err := ctx.SetState(state); err != nil {
											t.Fatal(err)
										}
										if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
											t.Fatal(fast, start, exit, err)
										}
										for i, v := range want {
											if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
												t.Fatal("frame", fast, start, i, got, v)
											}
										}
										if !bytes.Equal(buffer[128:], expected) {
											t.Fatal("payload", fast, start, buffer[128:], expected)
										}
									}
									// A failing store must precede every forwarded local/tee effect.
									writeSlot(t, ctx, 16, 0, value.Value{})
									writeSlot(t, ctx, 16, local, value.Value{Lo: 0x9999})
									state := native.State{Used: 16, Depth: 1, Budget: 64, Fuel: 1000, Continuation: 1}
									state.Payload[2] = address
									ctx.SetState(state)
									if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Trap {
										t.Fatal(exit, err)
									}
									state, err = ctx.State()
									if err != nil || state.Source != 101 || state.Payload[0] != uint64(wasm.TrapNullReference) || readSlot(t, ctx, 16, local).Lo != 0x9999 {
										t.Fatal("store trap reordered local", state, err)
									}
								}
							})
						}
					}
				}
			}
		}
	}
}
