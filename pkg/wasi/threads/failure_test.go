//go:build linux || darwin

package threads

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGroupRejectsInvalidContractsAndRollsBack(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		for _, test := range []struct {
			name string
			want error
		}{
			{"empty", vm.ErrLink}, {"missing-memory", vm.ErrLink},
			{"missing-worker", vm.ErrType}, {"bad-spawn", vm.ErrLink},
			{"private-memory", vm.ErrLink}, {"foreign-import", vm.ErrLink},
			{"alias-limits", vm.ErrLink},
		} {
			b, err := os.ReadFile("testdata/invalid/" + test.name + ".wasm")
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, b)
			if err != nil {
				t.Fatal(err)
			}
			p, _ := preview1.NewProcess(preview1.DefaultConfig())
			_, err = New(nil, e, m, p, Config{})
			if !errors.Is(err, test.want) || p.Context().Err() != nil {
				t.Fatal(test.name, err)
			}
			m.Close()
			p.Close()
		}
		b, _ := os.ReadFile("testdata/group.wasm")
		m, err := e.Compile(nil, b)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		p, _ := preview1.NewProcess(preview1.DefaultConfig())
		defer p.Close()
		if _, err := New(nil, nil, m, p, Config{}); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
		for _, limit := range []uint32{0, maxTID} {
			cfg := DefaultConfig()
			cfg.MaxThreads = limit
			if _, err := New(nil, e, m, p, cfg); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := New(ctx, e, m, p, Config{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		cfg := DefaultConfig()
		cfg.Store.Limits.MaxMemoryPages = 0
		if _, err := New(nil, e, m, p, cfg); !errors.Is(err, wasm.ErrLimit) || p.Context().Err() != nil {
			t.Fatal("memory rollback", err)
		}
		p.Close()
		if _, err := New(nil, e, m, p, Config{}); !errors.Is(err, context.Canceled) {
			t.Fatal("closed process", err)
		}
		m.Close()
		if _, err := New(nil, e, m, p, Config{}); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("closed module", err)
		}
		p2, _ := preview1.NewProcess(preview1.DefaultConfig())
		defer p2.Close()
		cfgEngine := vm.DefaultEngineConfig()
		e2, _ := vm.NewEngineWithConfig(cfgEngine)
		defer e2.Close()
		empty, _ := os.ReadFile("testdata/invalid/empty.wasm")
		m2, err := e2.Compile(nil, empty)
		if err != nil {
			t.Fatal(err)
		}
		defer m2.Close()
		if _, err := New(nil, e2, m2, p2, Config{}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal("missing Threads", err)
		}
		// A valid module on a closed Engine cannot allocate the shared backing.
		m3, err := e.Compile(nil, b)
		if err != nil {
			t.Fatal(err)
		}
		defer m3.Close()
		e.Close()
		if _, err := New(nil, e, m3, p2, Config{}); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("closed engine", err)
		}
	})
}

func TestGroupSpawnFailureReservations(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		for _, failure := range []string{"id-exhausted", "closed-module", "store-limit", "host-functions-limit", "spawn-function-limit", "closed-engine"} {
			cfg := DefaultConfig()
			cfg.Execution = options
			g, _ := newGroup(t, e, nil, cfg, 0)
			switch failure {
			case "id-exhausted":
				g.nextTID = maxTID
			case "closed-module":
				g.module.Close()
			case "store-limit":
				g.config.Store.Limits.MaxMemoryPages = 0
			case "host-functions-limit":
				g.config.Store.Limits.MaxFunctions = 0
			case "spawn-function-limit":
				g.config.Store.Limits.MaxFunctions = 46
			case "closed-engine":
				e.Close()
			}
			id, err := g.spawn(context.Background(), 16)
			if err != nil || int32(id) != -int32(preview1.ErrAgain) {
				t.Fatal(failure, id, err)
			}
			if g.active != 0 || len(g.stores) != 0 {
				t.Fatal("failed spawn leaked reservation")
			}
			g.Close()
			g.WaitClosed(nil)
			if _, err := g.spawn(context.Background(), 0); !errors.Is(err, context.Canceled) {
				t.Fatal("closed spawn", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := g.spawn(ctx, 0); !errors.Is(err, context.Canceled) {
				t.Fatal("canceled spawn", err)
			}
			if failure != "closed-engine" {
				if store, _, err := g.instantiate(0); store != nil || !errors.Is(err, context.Canceled) {
					t.Fatal("closing group published a new Store", err)
				}
			}
		}
	})
}

func TestGroupImportTypeFailureReclaimsProcess(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		b, err := os.ReadFile("testdata/invalid/bad-p1.wasm")
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, b)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		p, _ := preview1.NewProcess(preview1.DefaultConfig())
		g, err := New(nil, e, m, p, Config{})
		if err != nil {
			t.Fatal(err)
		}
		if err := g.Run(); !errors.Is(err, vm.ErrLink) {
			t.Fatal("mismatched P1 import", err)
		}
		if err := p.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGroupSharedImportAliasesIntersectLimits(t *testing.T) {
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		b, err := os.ReadFile("testdata/alias-compatible.wasm")
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, b)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		p, _ := preview1.NewProcess(preview1.DefaultConfig())
		defer p.Close()
		g, err := New(nil, e, m, p, Config{})
		if err != nil {
			t.Fatal(err)
		}
		if len(g.memories) != 1 {
			t.Fatal("duplicate shared backing")
		}
		if err := g.Run(); err != nil {
			t.Fatal(err)
		}
	})
}
