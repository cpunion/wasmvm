package vm

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCustomPagesMemoryKindsQuotasAndMatching(t *testing.T) {
	input, err := os.ReadFile("testdata/custom-pages.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		blocked, _ := base.NewStore(DefaultStoreConfig())
		defer blocked.Close()
		if _, err := blocked.NewMemory(MemoryType{PageSize: 1}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal("host custom pages enabled by default", err)
		}
		for _, metered := range []bool{false, true} {
			cfg := base.config
			cfg.Profile, cfg.Features = "core3", wasm.Threads|wasm.CustomPageSizes
			e, err := NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, input)
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			sc := DefaultStoreConfig()
			sc.Limits.MaxMemoryPages = 5
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			linker := NewLinker(s)
			mems := map[string]Extern{}
			for _, kind := range []string{"small", "wide", "shared"} {
				memory, err := s.NewMemory(MemoryType{Min: 8, Max: 131072, HasMax: true, PageSize: 1, Index64: kind == "wide", Shared: kind == "shared"})
				if err != nil {
					t.Fatal(kind, err)
				}
				mems[kind] = memory
				if err := linker.Define("h", kind, memory); err != nil {
					t.Fatal(err)
				}
			}
			in := instantiate(t, linker, m)
			call := func(name string, args ...Value) uint64 {
				t.Helper()
				r, err := export(t, in, name).Call(nil, args, CallOptions{Metered: metered, Fuel: 100})
				if err != nil || len(r.Values) != 1 {
					t.Fatal(name, err, r)
				}
				return bits(t, r.Values[0])
			}
			if call("small-size") != 8 || call("wide-size") != 8 || call("shared-size") != 8 || call("default-size") != 1 {
				t.Fatal("page size leaked between memory indices")
			}
			// Native bounds-elision must use eight bytes, not eight 64-KiB pages.
			if _, err := export(t, in, "small-constant-oob").Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("incorrect constant memory guard", err)
			}
			if call("small-grow", I32(1)) != 8 || call("wide-grow", I64(1)) != 8 || call("shared-grow", I32(1)) != 8 {
				t.Fatal("guest growth unit")
			}
			if call("small-load", I32(8)) != 0 || call("wide-load", I64(8)) != 0 || call("shared-add", I32(0)) != 0 || call("shared-add", I32(0)) != 1 || call("shared-wait", I32(0)) != 1 {
				t.Fatal("growth zeroing or atomic access")
			}
			for _, name := range []string{"small-load", "wide-load", "shared-add", "shared-wait"} {
				arg := I32(9)
				if name == "wide-load" {
					arg = I64(9)
				}
				if _, err := export(t, in, name).Call(nil, []Value{arg}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("logical byte boundary", name, err)
				}
			}
			// Shared maximum reserves two host quota pages; each private small
			// memory uses one and the default memory uses one. Grow inside a
			// charged host page, then reject the first additional quota unit.
			if call("small-grow", I32(65527)) != 9 || call("small-size") != 65536 || call("small-grow", I32(1)) != uint64(^uint32(0)) {
				t.Fatal("host byte quota")
			}
			if call("wide-grow", I64(^uint64(0))) != ^uint64(0) || call("wide-size") != 9 {
				t.Fatal("memory64 growth overflow")
			}
			usage, err := s.Usage()
			if err != nil || usage.Core.MemoryPages != 4 || usage.Core.ReservedSharedMemoryPages != 2 || usage.Core.SharedMemoryPages != 1 {
				t.Fatal("host quota usage", usage, err)
			}
			// All count bounds match; only the static page size differs.
			wrong := compileDefinition(t, e, &module.Module{Imports: []module.Import{{Module: "h", Name: "small", Kind: 2, Memory: module.Limits{Min: 8}}}})
			if _, err := linker.Instantiate(nil, wrong, CallOptions{}); !errors.Is(err, ErrLink) {
				t.Fatal("page-size mismatch linked", err)
			}
			capability, err := mems["shared"].ShareMemory()
			if err != nil {
				t.Fatal(err)
			}
			defer capability.Close()
			childConfig := DefaultStoreConfig()
			childConfig.Limits.MaxMemoryPages = 1
			low, err := e.NewStore(childConfig)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := low.ImportMemory(capability); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("shared maximum escaped byte quota", err)
			}
			low.Close()
			childConfig.Limits.MaxMemoryPages = 2
			child, err := e.NewStore(childConfig)
			if err != nil {
				t.Fatal(err)
			}
			defer child.Close()
			shared, err := child.ImportMemory(capability)
			if err != nil {
				t.Fatal(err)
			}
			s.Close()
			if err := s.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
			if s.state.Usage().MemoryPages != 0 || s.state.Usage().ReservedSharedMemoryPages != 0 {
				t.Fatal("custom-page quota leaked at close", s.state.Usage())
			}
			if old, ok, err := shared.GrowMemory(65527); err != nil || !ok || old != 9 {
				t.Fatal("shared lease after creator close", old, ok, err)
			}
			child.Close()
			if err := child.WaitClosed(nil); err != nil || child.state.Usage().ReservedSharedMemoryPages != 0 {
				t.Fatal("shared quota leaked", err)
			}
		}
	})
}
