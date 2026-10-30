//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func httpOwner(t *testing.T, config HTTPConfig) (*HTTP, *Network, *IO) {
	t.Helper()
	network, process := networkOwner(t, NetworkConfig{AllowNameLookup: true})
	p, err := NewHTTP(network, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
	})
	return p, network, process
}
func httpInstance(t *testing.T, p *HTTP, backend, name string) *component.Instance {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	engine, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(engine.Close)
	module, err := engine.Compile(nil, fixture(t, name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(module.Close)
	in, err := module.Instantiate(nil, httpWorldImports(t, p), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}
func TestHTTPFieldsGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, _, _ := httpOwner(t, HTTPConfig{})
			in := httpInstance(t, p, backend, "http-fields")
			for _, name := range []string{"probe", "fields-batch", "body-batch"} {
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				var args []component.Value
				want := uint64(42)
				if name != "probe" {
					args = []component.Value{{Bits: 128}}
					want = 128
				}
				result, err := f.Call(nil, args, component.CallOptions{Metered: true, Fuel: 10000000})
				if err != nil || result.Value.Bits != want {
					t.Fatal(name, result, err)
				}
			}
			p.mu.Lock()
			left := len(p.entries)
			p.mu.Unlock()
			if left != 0 {
				t.Fatal("field resource leak", left)
			}
		})
	}
}
func TestHTTPContractRejections(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for _, name := range []string{"tuple-order", "value-width", "header-error", "clone-result", "lookup-result", "request-option", "request-path", "http-error"} {
		module, err := engine.Compile(nil, fixture(t, "http-contract-"+name))
		if err != nil {
			t.Fatal(name, err)
		}
		in, err := module.Instantiate(nil, p.Imports(), component.CallOptions{})
		if in != nil {
			in.Close()
		}
		module.Close()
		if !errors.Is(err, vm.ErrType) {
			t.Fatal(name, err)
		}
	}
}

func httpWorldImports(t *testing.T, p *HTTP) component.Imports {
	t.Helper()
	out := p.Imports()
	clock, err := NewWallClock(WallClockConfig{})
	if err != nil {
		t.Fatal(err)
	}
	random, err := NewRandom(RandomConfig{})
	if err != nil {
		t.Fatal(err)
	}
	for _, imports := range []component.Imports{clock.Imports(), random.Imports()} {
		for name, value := range imports.Instances {
			out.Instances[name] = value
		}
	}
	return out
}
