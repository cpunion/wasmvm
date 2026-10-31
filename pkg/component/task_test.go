package component

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func asyncInstance(t *testing.T, backend, name string, imports Imports) *Instance {
	t.Helper()
	config := DefaultConfig()
	config.Core.Backend = backend
	config.MaxResources = 16
	e, err := NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	m, err := e.Compile(nil, fixture(t, name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	in, err := m.Instantiate(nil, imports, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}

func TestAsyncLiftTasks(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			wide := asyncInstance(t, backend, "async-context64", Imports{})
			contextValue, _ := wide.Function("wide-context")
			for _, bits := range []uint64{0, 1 << 63, 0xfedcba9876543210, ^uint64(0)} {
				v, err := contextValue.Call(nil, []Value{{Bits: bits}}, CallOptions{})
				if err != nil || v.Value.Bits != bits {
					t.Fatal("context width or thread isolation", v, err)
				}
			}
			stackful, _ := wide.Function("stackful")
			if v, err := stackful.Call(nil, []Value{{Bits: 7}}, CallOptions{}); err != nil || v.Value.Bits != 49 {
				t.Fatal(v, err)
			}
			in := asyncInstance(t, backend, "async-lift", Imports{})
			for range 8 {
				for _, name := range []string{"run", "stackful", "direct", "sync"} {
					f, _ := in.Function(name)
					if f.Info().Async() != (name != "sync") {
						t.Fatal("function kind", name)
					}
					v, err := f.Call(nil, []Value{{Bits: 7}}, CallOptions{})
					want := uint64(49)
					if name == "run" {
						want = 42
					}
					if name == "sync" {
						want = 7
					}
					if err != nil || v.Value == nil || v.Value.Bits != want {
						t.Fatal(name, v, err)
					}
				}
			}
			for _, name := range []string{"missing", "twice", "wrong", "bad-code", "wait", "after-return", "sync-return"} {
				t.Run(name, func(t *testing.T) {
					fresh := asyncInstance(t, backend, "async-lift", Imports{})
					f, _ := fresh.Function(name)
					if _, err := f.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal(err)
					}
					f, _ = fresh.Function("run")
					if _, err := f.Call(nil, []Value{{}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("failed namespace reentered", err)
					}
				})
			}
			for _, fuel := range []uint64{0, 40} {
				fresh := asyncInstance(t, backend, "async-lift", Imports{})
				f, _ := fresh.Function("forever")
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				_, err := f.Call(ctx, []Value{{}}, CallOptions{Metered: fuel != 0, Fuel: fuel})
				cancel()
				if fuel == 0 && !errors.Is(err, context.DeadlineExceeded) || fuel != 0 && !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("callback limit", fuel, err)
				}
			}
		})
	}
	if (FunctionInfo{}).Async() {
		t.Fatal("nil metadata")
	}
}
