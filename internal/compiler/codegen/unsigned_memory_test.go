package codegen

import (
	"bytes"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"testing"
)

func unsignedMemoryFixture(first, reversed bool, stride uint32) *slots.Function {
	f := memoryLoopFixture(12, first, reversed)
	f.Sources[7].Wasm.Bits = uint64(stride)
	for pc := range f.Code {
		in := &f.Code[pc]
		if in.Op != 0x46 && in.Op != 0x47 {
			continue
		}
		op := uint32(0x49)
		if reversed {
			op = 0x4f
		}
		if first {
			if reversed {
				op = 0x4d
			} else {
				op = 0x4b
			}
		}
		in.Op = op
		f.Sources[in.Source].Wasm.Op = op
	}
	return f
}

func TestUnsignedMemoryLoopKeepsColdFramesAndPartialEffects(t *testing.T) {
	for _, first := range []bool{false, true} {
		for _, reversed := range []bool{false, true} {
			for _, stride := range []uint32{1, 3, 4} {
				f := unsignedMemoryFixture(first, reversed, stride)
				config := Config{Unmetered: true, MemoryMinimumPages: 1}
				for _, arch := range []string{"amd64", "arm64"} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || c.LoopGuards != 1 {
						t.Fatal("unsigned loop not guarded", arch, first, reversed, stride, err)
					}
				}
				if !native.Available {
					continue
				}
				_, lease, ctx := configuredContext(t, f, config)
				_, reference, ref := functionContext(t, f, true)
				mem := bindMemory(t, ctx, 65536, 65536)
				refmem := bindMemory(t, ref, 65536, 65536)
				initial := bytes.Repeat([]byte{0xa5}, 65536)
				for pc := range f.Code {
					for _, input := range [][2]uint64{{0, 0}, {0, 1}, {0, 11}, {0, 12}, {0, 13}, {65500, 0}, {65501, 0}, {65502, 0}, {0xffffffff, 0}, {0, 0xffffffff}, {1 << 32, 1<<63 | 1}} {
						if err := mem.Write(0, initial); err != nil {
							t.Fatal(err)
						}
						if err := refmem.Write(0, initial); err != nil {
							t.Fatal(err)
						}
						for s := range f.Slots {
							v := value.Value{Lo: 1<<63 | uint64(s+7), Hi: ^uint64(0) - uint64(s)}
							if s < 2 {
								v.Lo = input[s]
							}
							writeSlot(t, ctx, 16, slots.Slot(s), v)
							writeSlot(t, ref, 16, slots.Slot(s), v)
						}
						state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
						ctx.SetState(state)
						ref.SetState(state)
						got, err := ctx.Enter(lease, 0)
						want, re := ref.Enter(reference, 0)
						gs, _ := ctx.State()
						ws, _ := ref.State()
						if err != nil || re != nil || got != want || gs != ws {
							t.Fatal("state/trap", first, reversed, stride, pc, input, got, want, gs, ws, err, re)
						}
						for s := range f.Slots {
							g, w := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s))
							if g != w {
								t.Fatal("raw frame", first, reversed, stride, pc, input, s, g, w)
							}
						}
						gb, wb := make([]byte, 65536), make([]byte, 65536)
						mem.Read(0, gb)
						refmem.Read(0, wb)
						if !bytes.Equal(gb, wb) {
							t.Fatal("partial effects", first, reversed, stride, pc, input)
						}
					}
				}
			}
		}
	}
}

func TestUnsignedMemoryLoopRejectsSignedAndOppositeOrder(t *testing.T) {
	for _, first := range []bool{false, true} {
		for _, reversed := range []bool{false, true} {
			for op := uint32(0x48); op <= 0x4f; op++ {
				accepted := uint32(0x49)
				if reversed {
					accepted = 0x4f
				}
				if first {
					if reversed {
						accepted = 0x4d
					} else {
						accepted = 0x4b
					}
				}
				if op == accepted {
					continue
				}
				f := unsignedMemoryFixture(first, reversed, 4)
				for pc := range f.Code {
					if f.Code[pc].Source == 11 {
						f.Code[pc].Op = op
						f.Sources[11].Wasm.Op = op
					}
				}
				for _, arch := range []string{"amd64", "arm64"} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
					if err != nil || c.LoopGuards != 0 {
						t.Fatal("unsafe order selected", arch, first, reversed, op, err)
					}
				}
			}
		}
	}
}
