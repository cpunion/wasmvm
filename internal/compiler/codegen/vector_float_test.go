package codegen

import (
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func vectorLaneBits(v value.Value, wide bool, lane int) uint64 {
	if wide {
		if lane == 0 {
			return v.Lo
		}
		return v.Hi
	}
	bits := v.Lo
	if lane >= 2 {
		bits = v.Hi
	}
	return uint64(uint32(bits >> ((lane % 2) * 32)))
}

func vectorFloatMatches(op uint32, a, b, got, want value.Value) bool {
	if got == want {
		return true
	}
	if !nativeVectorFloatOpcode(0xfd0000 + op) {
		return false
	}
	if op == 94 || op == 95 {
		input, output := wasm.F64, wasm.F32
		if op == 95 {
			input, output = output, input
		} else if got.Hi != 0 {
			return false
		}
		for i := 0; i < 2; i++ {
			if !conversionMatches(input, output, vectorLaneBits(a, op == 94, i),
				vectorLaneBits(got, op == 95, i), vectorLaneBits(want, op == 95, i)) {
				return false
			}
		}
		return true
	}
	if op >= 248 {
		return false // Saturation, zero padding and integer-to-float rounding are exact.
	}
	scalar, wide, _ := vectorFloatOperation(op)
	if scalar >= 0x98 {
		return false // Pseudo-min/max preserve selected raw NaN and zero bits.
	}
	lanes := 4
	if wide {
		if scalar <= 0x60 {
			scalar += 6
		} else {
			scalar += 14
		}
		lanes = 2
	}
	for i := 0; i < lanes; i++ {
		if !floatResultMatches(scalar, vectorLaneBits(a, wide, i), vectorLaneBits(b, wide, i),
			vectorLaneBits(got, wide, i), vectorLaneBits(want, wide, i)) {
			return false
		}
	}
	return true
}

func TestVectorFloatInventoryAndCodeLimits(t *testing.T) {
	count := 0
	for op := uint32(0); op < 256; op++ {
		if !nativeVectorFloatOpcode(0xfd0000 + op) {
			continue
		}
		count++
		f, _ := vectorFixture(op, 0, 3, 0)
		for _, arch := range []string{"arm64", "amd64"} {
			profiles := []X86Features{0}
			if arch == "amd64" {
				profiles = []X86Features{0, 1, 2, 3, 4, 5, 6, 7}
			}
			for _, profile := range profiles {
				for _, fast := range []bool{false, true} {
					config := Config{Unmetered: fast, X86Features: profile}
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || c.Helpers != 0 || c.NativeVectors != 1 {
						t.Fatal(op, arch, fast, err)
					}
					limit := DefaultLimits()
					limit.MaxCodeBytes = uint64(len(c.Entry.Code))
					if _, err := CompileFunctionWithConfig(arch, f, limit, config); err != nil {
						t.Fatal(err)
					}
					limit.MaxCodeBytes--
					if _, err := CompileFunctionWithConfig(arch, f, limit, config); !errors.Is(err, wasm.ErrLimit) {
						t.Fatal(op, arch, err)
					}
				}
			}
		}
	}
	if count != 52 || nativeVectorFloatOpcode(65) || nativeVectorFloatOpcode(0xfd00ee) {
		t.Fatal("floating-point opcode inventory", count)
	}
}

func TestNativeVectorFloatMixedEdges(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for op := uint32(0); op < 256; op++ {
		if !nativeVectorFloatOpcode(0xfd0000 + op) {
			continue
		}
		typ := wasm.F32
		if op == 94 || op == 252 || op == 253 {
			typ = wasm.F64
		} else if op == 250 || op == 251 || op == 254 || op == 255 {
			typ = wasm.I32
		} else if op != 95 && op != 248 && op != 249 {
			if _, wide, _ := vectorFloatOperation(op); wide {
				typ = wasm.F64
			}
		}
		wide := typ == wasm.F64
		edges := conversionInputs(typ)
		for _, x := range []float64{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat32,
			-0.25, 0.25, -2.5, 2.5, -3.5, 3.5, -math.Ldexp(1, 23), math.Ldexp(1, 23),
			-math.Ldexp(1, 52), math.Ldexp(1, 52)} {
			bits := math.Float64bits(x)
			if !wide {
				bits = uint64(math.Float32bits(float32(x)))
			}
			edges = append(edges, bits)
		}
		for _, profile := range vectorTestProfiles(op) {
			for _, fast := range []bool{false, true} {
				f, pc := vectorFixture(op, 0, 3, 0)
				_, lease, c := configuredContext(t, f, Config{Unmetered: fast, X86Features: profile})
				in := f.Code[pc]
				pack := func(index int) value.Value {
					if wide {
						return value.Value{Lo: edges[index%len(edges)], Hi: edges[(index+1)%len(edges)]}
					}
					return value.Value{Lo: edges[index%len(edges)] | edges[(index+1)%len(edges)]<<32,
						Hi: edges[(index+2)%len(edges)] | edges[(index+3)%len(edges)]<<32}
				}
				for i := range edges {
					for _, offset := range []int{0, 1, 7, 31} {
						a, b := pack(i), pack(i+offset)
						writeSlot(t, c, 16, 0, a)
						if in.B != slots.NoSlot {
							writeSlot(t, c, 16, in.B, b)
						}
						if err := c.SetState(native.State{Used: 16, Depth: 1, Budget: 1000, Continuation: uint64(pc)}); err != nil {
							t.Fatal(err)
						}
						if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(op, exit, err)
						}
						got := readSlot(t, c, 16, in.Dst)
						want := simd.Eval(f.Sources[in.Source].Wasm, a, b, value.Value{})
						if !vectorFloatMatches(op, a, b, got, want) {
							t.Fatalf("op=%x fast=%v a=%x b=%x got=%x want=%x", op, fast, a, b, got, want)
						}
					}
				}
			}
		}
	}
}
