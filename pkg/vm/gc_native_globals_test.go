package vm

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGCGlobalsPublishBeforeCollectionAndHostReentry(t *testing.T) {
	input, err := os.ReadFile("testdata/gc-native-globals.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Features = wasm.GC
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
		for _, options := range []CallOptions{{}, {Fuel: 2000}} {
			sc := DefaultStoreConfig()
			sc.CollectEveryAllocation = true
			sc.Limits.MaxHeapObjects, sc.Limits.MaxHeapBytes = 8, 1024
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			var in *Instance
			host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
				if err := c.Collect(); err != nil {
					return nil, err
				}
				load, err := c.Export(in, "load")
				if err != nil {
					return nil, err
				}
				r, err := c.Call(load, nil)
				if err != nil || bits(t, r.Values[0]) != 77 {
					t.Fatal("host did not see published global", r, err)
				}
				allocate, err := c.Export(in, "save-allocate")
				if err != nil {
					return nil, err
				}
				_, err = c.Call(allocate, nil)
				return nil, err
			})
			if err != nil {
				t.Fatal(err)
			}
			linker := NewLinker(s)
			if err := linker.Define("h", "collect", host); err != nil {
				t.Fatal(err)
			}
			in = instantiate(t, linker, m)
			r, err := export(t, in, "save-host").Call(nil, nil, options)
			if err != nil || bits(t, r.Values[0]) != 88 {
				t.Fatal("resumed entry retained stale global", r, err)
			}
			if _, err := export(t, in, "save-trap").Call(nil, nil, options); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			r, err = export(t, in, "load").Call(nil, nil, options)
			if err != nil || bits(t, r.Values[0]) != 99 {
				t.Fatal("trap lost published root", r, err)
			}
		}
	})
}
