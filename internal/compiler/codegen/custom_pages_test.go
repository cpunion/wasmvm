package codegen

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeCustomPageConfiguration(t *testing.T) {
	m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32}}, 0, module.Expr{{Op: 0x3f}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	for _, config := range []Config{
		{MemoryCount: 2, MemoryPageSizes: []uint64{1}},
		{MemoryCount: 1, MemoryPageSizes: []uint64{2}},
	} {
		if _, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), config); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("invalid page-size inventory", err)
		}
	}
	for _, fast := range []bool{false, true} {
		config := Config{Unmetered: fast, MemoryCount: 1, MemoryPageSizes: []uint64{1}, MemoryMinimumPages: 65537}
		for _, arch := range []string{"amd64", "arm64"} {
			if _, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config); err != nil {
				t.Fatal("valid byte-sized minimum", arch, err)
			}
		}
		if !native.Available {
			continue
		}
		compiled, lease, ctx := configuredContext(t, f, config)
		bindMemory(t, ctx, 65537, 65537)
		got, exit, _ := runFunction(t, compiled, lease, ctx, nil, 100, !fast, 100)
		if exit != native.Return || len(got) != 1 || got[0].Lo != 65537 {
			t.Fatal(got, exit)
		}
	}
}
