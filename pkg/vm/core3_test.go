package vm

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCore3ProfileInitializationAndComposition(t *testing.T) {
	input, err := os.ReadFile("testdata/core3.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		config := newEngine().config
		legacy, err := NewEngineWithConfig(config)
		if err != nil {
			t.Fatal(err)
		}
		defer legacy.Close()
		if err := legacy.Validate(nil, input); err == nil {
			t.Fatal("Core 3.0 enabled by default")
		}
		config.Features = wasm.Core3Features &^ wasm.Core3Rules
		proposals, err := NewEngineWithConfig(config)
		if err != nil {
			t.Fatal(err)
		}
		defer proposals.Close()
		if err := proposals.Validate(nil, input); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("archived proposal changed", err)
		}
		config.Features = 0
		config.Profile = "core3"
		e, err := NewEngineWithConfig(config)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		info, err := e.Inspect(nil, input)
		if err != nil || info.Profile != "core3" || info.Features != wasm.Core3Features {
			t.Fatal(info, err)
		}
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		info, err = m.Info()
		if err != nil || info.Profile != "core3" || info.Features != wasm.Core3Features {
			t.Fatal(info, err)
		}
		for _, seed := range []uint32{8, 12} {
			sc := DefaultStoreConfig()
			sc.CollectEveryAllocation = true
			sc.Limits.MaxHeapObjects = 2
			sc.Execution.MaxFrames = 2
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			g, err := s.NewGlobal(GlobalType{Value: wasm.I32}, I32(seed))
			if err != nil {
				t.Fatal(err)
			}
			l := NewLinker(s)
			if err := l.Define("seed", "value", g); err != nil {
				t.Fatal(err)
			}
			in := instantiate(t, l, m)
			var actual [1]byte
			if !in.instance.Memories[1].Read(uint64(seed), actual[:]) || actual[0] != 'C' {
				t.Fatal("defined-global data offset", actual)
			}
			if in.instance.Tables[0].Values[0] == 0 {
				t.Fatal("defined-global element initializer")
			}
			for _, metered := range []bool{false, true} {
				r, err := export(t, in, "run").Call(nil, nil, CallOptions{Metered: metered, Fuel: 200})
				if err != nil || bits(t, r.Values[0]) != uint64(60+seed) {
					t.Fatal(r, err)
				}
			}
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			usage, err := s.Usage()
			if err != nil || usage.Core.HeapObjects != 1 {
				t.Fatal(usage, err)
			}
			s.Close()
			if err := s.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		}
		sc := DefaultStoreConfig()
		sc.Limits.MaxHeapObjects = 0
		s, err := e.NewStore(sc)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		g, err := s.NewGlobal(GlobalType{Value: wasm.I32}, I32(8))
		if err != nil {
			t.Fatal(err)
		}
		l := NewLinker(s)
		if err := l.Define("seed", "value", g); err != nil {
			t.Fatal(err)
		}
		if _, err := l.Instantiate(nil, m, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		usage, err := s.Usage()
		if err != nil || usage.Core.Instances != 0 || usage.Core.HeapObjects != 0 {
			t.Fatal("failed initializer leaked", usage, err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := l.Instantiate(ctx, m, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})
}
