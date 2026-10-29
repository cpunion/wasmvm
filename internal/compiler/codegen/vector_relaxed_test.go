package codegen

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var relaxedVectorEdges = []value.Value{
	{Lo: 0x8080808080808080, Hi: 0x8080808080808080},
	{Lo: 0x7f7f7f7f7f7f7f7f, Hi: 0x8000800080008000},
	{Lo: 0x800000007fffffff, Hi: 0x7fffffff80000000},
	{Lo: 0x7fa123457f800000, Hi: 0xff80000000000001},
	{Lo: 0x8000000000000000, Hi: 0x7ff123456789abcd},
	{Lo: 0x3f8000013f800001, Hi: 0x3ff0000000000001},
	{Lo: 0x3f7ffffe3f7ffffe, Hi: 0x3feffffffffffffe},
	{Lo: 0xbf800000bf800000, Hi: 0xbff0000000000000},
	{Lo: 0x3f8000003f800000, Hi: 0x3ff0000000000000},
	{Lo: 0x0000000100000000, Hi: 0xffffffffffffffff},
}

func TestCompleteRelaxedVectorInventoryAndLimits(t *testing.T) {
	for op := uint32(0x100); op <= 0x12f; op++ {
		want := op <= 0x113
		if nativeRelaxedVectorOpcode(0xfd0000|op) != want || nativeVectorOpcode(0xfd0000|op) != want || nativeRelaxedVectorOpcode(op) {
			t.Fatal("relaxed SIMD grammar", op)
		}
		if !want {
			continue
		}
		f, _ := vectorFixture(op, 0, 3, 12)
		for _, arch := range []string{"arm64", "amd64"} {
			profiles := []X86Features{0}
			if arch == "amd64" {
				profiles = append(profiles, 1, 2, 3, 4, 7, 31)
			}
			for _, profile := range profiles {
				for _, fast := range []bool{false, true} {
					config := Config{Unmetered: fast, X86Features: profile}
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || c.Helpers != 0 || c.NativeVectors != 1 {
						t.Fatal(op, arch, profile, fast, err)
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
}
