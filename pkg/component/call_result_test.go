package component

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestCallWithResultConsumesBeforeProducerExit(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"host-result-future", "host-result-stream"} {
			in := asyncInstance(t, backend, name, Imports{})
			run, _ := in.Function("run")
			var escaped *Caller
			calls := 0
			fuel, err := run.CallWithResult(nil, nil, CallOptions{Metered: true, Fuel: 1000}, func(c *Caller, v *Value) error {
				calls++
				escaped = c
				r, err := c.ReadAsync(v.Resource, 1)
				if err != nil || r.Count != 1 || len(r.Values) != 1 || r.Values[0].Bits != 42 {
					t.Fatal(name, r, err)
				}
				return c.DropAsync(v.Resource)
			})
			if err != nil || calls != 1 || fuel == 0 || fuel >= 1000 {
				t.Fatal(name, fuel, calls, err)
			}
			if _, err := escaped.ReadAsync(nil, 1); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("result Caller retained", err)
			}
			if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal("result consumer retained state")
			}
			probe := canonical.NewResourceTable(in.resourceBudget)
			for range in.config.MaxResources {
				if _, err := probe.AddOpaque(1); err != nil {
					t.Fatal(err)
				}
			}
			probe.Close()
			if _, err := run.CallWithResult(nil, nil, CallOptions{}, nil); !errors.Is(err, vm.ErrType) {
				t.Fatal(err)
			}
			if _, err := new(Function).CallWithResult(nil, nil, CallOptions{}, func(*Caller, *Value) error { return nil }); !errors.Is(err, vm.ErrReference) {
				t.Fatal(err)
			}
		}
	}
}
func TestCallWithResultAbandonsUnconsumedEndpoints(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, panicCallback := range []bool{false, true} {
			in := asyncInstance(t, backend, "host-result-stream", Imports{})
			run, _ := in.Function("run")
			var escaped *Value
			invoke := func() (err error) {
				defer func() {
					if recover() != nil {
						err = wasm.ErrTrap
					}
				}()
				_, err = run.CallWithResult(nil, nil, CallOptions{}, func(_ *Caller, v *Value) error {
					escaped = v
					if panicCallback {
						panic("consume")
					}
					return vm.ErrLink
				})
				return err
			}
			if err := invoke(); err == nil {
				t.Fatal("failed consumer accepted")
			}
			if _, err := in.DropAsync(nil, escaped.Resource, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("unpublished result survived", err)
			}
			if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal("failed callback retained state")
			}
			probe := canonical.NewResourceTable(in.resourceBudget)
			for range in.config.MaxResources {
				if _, err := probe.AddOpaque(1); err != nil {
					t.Fatal("failed callback retained endpoint", err)
				}
			}
			probe.Close()
		}
	}
}
