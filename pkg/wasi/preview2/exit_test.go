package preview2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExitStopsEveryLinkedInstance(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			engine, _ := component.NewEngine(config)
			defer engine.Close()
			compiled, err := engine.Compile(nil, fixture(t, "exit"))
			if err != nil {
				t.Fatal(err)
			}
			defer compiled.Close()
			for _, test := range []struct {
				name string
				args []component.Value
				code uint32
			}{{"success", nil, 0}, {"failure", nil, 1}, {"with-code", []component.Value{{Bits: 42}}, 42}, {"with-code", []component.Value{{Bits: 255}}, 255}} {
				in, err := compiled.Instantiate(nil, ExitImports(), component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				left, _ := in.Function("left", "alive")
				right, _ := in.Function("right", "alive")
				for _, f := range []*component.Function{left, right} {
					v, err := f.Call(nil, nil, component.CallOptions{})
					if err != nil || v.Value.Bits != 7 {
						t.Fatal(v, err)
					}
				}
				run, _ := in.Function(test.name)
				_, err = run.Call(nil, test.args, component.CallOptions{Metered: true, Fuel: 100})
				var exit *ExitError
				if !errors.As(err, &exit) || exit.Code != test.code || exit.Error() == "" {
					t.Fatal(test.name, err, exit)
				}
				for _, f := range []*component.Function{left, right, run} {
					if _, err := f.Call(nil, nil, component.CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal("linked instance survived exit", err)
					}
				}
				wait, cancel := context.WithTimeout(context.Background(), time.Second)
				if err := in.WaitClosed(wait); err != nil {
					t.Fatal(err)
				}
				cancel()
				in.Close()
			}
			start, err := engine.Compile(nil, fixture(t, "exit-start"))
			if err != nil {
				t.Fatal(err)
			}
			defer start.Close()
			in, err := start.Instantiate(nil, ExitImports(), component.CallOptions{})
			var exit *ExitError
			if in != nil || !errors.As(err, &exit) || exit.Code != 42 {
				t.Fatal("initialization exit published an instance or hid its status", in, err)
			}
		})
	}
}

func TestExitRejectsSchemasAndInputBeforeStopping(t *testing.T) {
	e, _ := component.NewEngine(component.DefaultConfig())
	defer e.Close()
	for _, name := range []string{"exit-wrong-status", "exit-wrong-code"} {
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(err)
		}
		if in, err := m.Instantiate(nil, ExitImports(), component.CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal(in, err)
		}
		m.Close()
	}
	m, err := e.Compile(nil, fixture(t, "exit"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	in, err := m.Instantiate(nil, ExitImports(), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	run, _ := in.Function("with-code")
	if _, err := run.Call(nil, []component.Value{{Bits: 256}}, component.CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	alive, _ := in.Function("left", "alive")
	if _, err := alive.Call(nil, nil, component.CallOptions{}); err != nil {
		t.Fatal("invalid host input closed the Store", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := run.Call(ctx, []component.Value{{Bits: 1}}, component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := alive.Call(nil, nil, component.CallOptions{}); err != nil {
		t.Fatal(err)
	}
}
