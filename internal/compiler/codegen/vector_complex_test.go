package codegen

import (
	"encoding/binary"
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompleteNativeSIMDInventory(t *testing.T) {
	// Independently pinned Core 2.0 grammar, including every reserved hole.
	mask := [4]uint64{0xffffffffffffffff, 0xffffffffffffffff, 0xf7e27f9bfbffffff, 0xffffbffbffe27f9b}
	count, complex := 0, 0
	for op := uint32(0); op < 256; op++ {
		full := uint32(0xfd0000) + op
		got := nativeVectorOpcode(full) || nativeVectorMemoryOpcode(full) || op == 12
		want := mask[op/64]>>(op%64)&1 != 0
		if got != want {
			t.Fatal("native SIMD grammar mismatch", op, got, want)
		}
		if got {
			count++
		}
		if nativeVectorComplexOpcode(full) {
			complex++
		}
		if !nativeVectorOpcode(full) {
			continue
		}
		f, _ := vectorFixture(op, 0, 3, 0)
		for _, arch := range []string{"arm64", "amd64"} {
			profiles := []X86Features{0}
			if arch == "amd64" {
				profiles = append(profiles, X86SSSE3, X86SSE41, X86SSSE3|X86SSE41)
			}
			for _, profile := range profiles {
				for _, fast := range []bool{false, true} {
					config := Config{Unmetered: fast, X86Features: profile}
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || c.Helpers != 0 || c.NativeVectors != 1 {
						t.Fatal(op, arch, fast, c, err)
					}
					limit := DefaultLimits()
					limit.MaxCodeBytes = uint64(len(c.Entry.Code))
					if _, err := CompileFunctionWithConfig(arch, f, limit, config); err != nil {
						t.Fatal(err)
					}
					limit.MaxCodeBytes--
					if _, err := CompileFunctionWithConfig(arch, f, limit, config); !errors.Is(err, wasm.ErrLimit) {
						t.Fatal(op, arch, fast, err)
					}
				}
			}
		}
	}
	if count != 236 || complex != 39 || nativeVectorComplexOpcode(13) {
		t.Fatal(count, complex)
	}
}

// Execute only profiles advertised by the actual host; emission tests above
// independently exercise all masks even when running on a different ISA.
func vectorTestProfiles(op uint32) []X86Features {
	profiles := []X86Features{0}
	if !native.Available || runtime.GOARCH != "amd64" {
		return profiles
	}
	affected := X86Features(0)
	if op == 13 || op == 14 || op == 98 || op == 130 {
		affected |= X86SSSE3
	}
	if op == 134 || op >= 135 && op <= 138 || op >= 156 && op <= 159 ||
		op >= 167 && op <= 170 || op == 181 || op >= 199 && op <= 202 ||
		op == 220 || op == 221 || op >= 103 && op <= 106 || op == 116 || op == 117 || op == 122 || op == 148 {
		affected |= X86SSE41
	}
	host := X86Features(native.CPUFeatures())
	for _, profile := range []X86Features{X86SSSE3, X86SSE41, X86SSSE3 | X86SSE41} {
		if profile&^host == 0 && profile&affected != 0 {
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

func TestNativeComplexSIMDMixedEdges(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	edges := []uint32{0, 1, 7, 8, 15, 16, 31, 32, 127, 128, 255, 256,
		32767, 32768, 32769, 65535, 65536, 0x7fffffff, 0x80000000, 0x80000001,
		0xffff8000, 0xffff7fff, 0xffffffff}
	var inputs []value.Value
	for i, edge := range edges {
		// Combine different lane widths and opposite saturation boundaries.
		var b [16]byte
		for lane := 0; lane < 4; lane++ {
			binary.LittleEndian.PutUint32(b[lane*4:], edges[(i+lane*7)%len(edges)])
		}
		inputs = append(inputs, value.Value{Lo: binary.LittleEndian.Uint64(b[:8]), Hi: binary.LittleEndian.Uint64(b[8:])},
			value.Value{Lo: uint64(edge) * 0x100000001, Hi: uint64(edge) * 0x100000001})
	}
	seed := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 { seed ^= seed << 13; seed ^= seed >> 7; seed ^= seed << 17; return seed }
	for i := 0; i < 128; i++ {
		inputs = append(inputs, value.Value{Lo: next(), Hi: next()})
	}
	for op := uint32(13); op < 256; op++ {
		if !nativeVectorComplexOpcode(0xfd0000 + op) {
			continue
		}
		patterns := 1
		if op == 13 {
			patterns = 12
		}
		for pattern := 0; pattern < patterns; pattern++ {
			for _, profile := range vectorTestProfiles(op) {
				for _, fast := range []bool{false, true} {
					f, pc := vectorFixture(op, 0, pattern%3, 0)
					if op == 13 {
						for i := range f.Sources[f.Code[pc].Source].Wasm.Vector {
							n := byte(next() & 31)
							switch pattern {
							case 0, 1:
								n = byte(i + pattern*16)
							case 2, 3:
								n = byte(15 - i + (pattern-2)*16)
							case 4, 5:
								n = byte(12 - i/4*4 + i%4 + (pattern-4)*16)
							case 6:
								n = 31
							}
							f.Sources[f.Code[pc].Source].Wasm.Vector[i] = n
						}
					}
					_, lease, ctx := configuredContext(t, f, Config{Unmetered: fast, X86Features: profile})
					for sample, a := range inputs {
						b := inputs[(sample+13)%len(inputs)]
						if sample%3 == 0 {
							b = a // Includes MIN*MIN Q15/dot overflow and duplicate inputs.
						}
						in := f.Code[pc]
						initial := make([]value.Value, len(f.Slots))
						initial[0] = a
						if in.B != slots.NoSlot {
							initial[in.B] = b
						}
						for i, v := range initial {
							writeSlot(t, ctx, 16, slots.Slot(i), v)
						}
						if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1000, Fuel: 1000}); err != nil {
							t.Fatal(err)
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(op, pattern, sample, fast, exit, err)
						}
						initial[in.Dst] = simd.Eval(f.Sources[in.Source].Wasm, a, b, value.Value{})
						for i, want := range initial {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
								t.Fatalf("op=%d pattern=%d sample=%d fast=%v slot=%d got=%x want=%x", op, pattern, sample, fast, i, got, want)
							}
						}
					}
					ctx.Close()
					lease.Close()
				}
			}
		}
	}
}
