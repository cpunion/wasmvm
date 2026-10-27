package codegen

import (
	"encoding/binary"
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestPackedVectorMaskedShiftCounts(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	inputs := []value.Value{{}, {Lo: ^uint64(0), Hi: ^uint64(0)},
		{Lo: 0x8001800180018001, Hi: 0x7fff7fff7fff7fff},
		{Lo: 0x807f01ff807f01ff, Hi: 0x010080ff010080ff},
		{Lo: 0x80000000ffffffff, Hi: 0x7fffffff00000001},
		{Lo: 0x8000000000000001, Hi: 0x7fffffffffffffff}}
	for size := uint32(0); size < 4; size++ {
		for kind := uint32(11); kind <= 13; kind++ {
			op := 96 + size*32 + kind
			for _, fast := range []bool{false, true} {
				for _, profile := range vectorTestProfiles(op) {
					f, pc := vectorFixture(op, 0, 0, 12)
					_, lease, c := configuredContext(t, f, Config{Unmetered: fast, X86Features: profile})
					in := f.Code[pc]
					for _, a := range inputs {
						for _, count := range []uint64{0, 1, 7, 8, 15, 16, 31, 32, 63, 64, 127, 128, 255, 256, 0x7fffffff, 0x80000000, 0xffffffff} {
							b := value.Value{Lo: count}
							writeSlot(t, c, 16, in.A, a)
							writeSlot(t, c, 16, in.B, b)
							if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Fuel: 1000, Budget: 1000}); err != nil {
								t.Fatal(err)
							}
							if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
								t.Fatal(op, fast, count, exit, err)
							}
							want := simd.Eval(f.Sources[in.Source].Wasm, a, b, value.Value{})
							if got := readSlot(t, c, 16, in.Dst); got != want || readSlot(t, c, 16, in.B) != b {
								t.Fatal(op, fast, count, a, got, want)
							}
						}
					}
					c.Close()
					lease.Close()
				}
			}
		}
	}
}

func TestPackedVectorReductionLaneOrder(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for size := uint32(0); size < 4; size++ {
		for _, kind := range []uint32{3, 4} {
			op := 96 + size*32 + kind
			for _, fast := range []bool{false, true} {
				f, pc := vectorFixture(op, 0, 3, 12)
				_, lease, c := functionContext(t, f, fast)
				in := f.Code[pc]
				for lane := uint32(0); lane < 16>>size; lane++ {
					for _, negative := range []bool{false, true} {
						var raw [16]byte
						for i := range raw {
							raw[i] = 1
						}
						for i := uint32(0); i < 1<<size; i++ {
							raw[(lane<<size)+i] = 0 // Exactly one zero lane.
						}
						if negative {
							raw[((lane+1)<<size)-1] = 0x80 // Exactly one signed lane.
						}
						a := value.Value{Lo: binary.LittleEndian.Uint64(raw[:8]), Hi: binary.LittleEndian.Uint64(raw[8:])}
						writeSlot(t, c, 16, in.A, a)
						if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Fuel: 1000, Budget: 1000}); err != nil {
							t.Fatal(err)
						}
						if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(op, lane, fast, exit, err)
						}
						want := simd.Eval(f.Sources[in.Source].Wasm, a, value.Value{}, value.Value{})
						if got := readSlot(t, c, 16, in.Dst); got != want || readSlot(t, c, 16, in.A) != a {
							t.Fatal(op, lane, fast, negative, got, want)
						}
					}
				}
			}
		}
	}
}
