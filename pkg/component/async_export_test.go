package component

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestAsyncEndpointExportsAndImportedEffects(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, name := range []string{"async-stream-exports", "async-future-exports"} {
				var kept Value
				imports := Imports{Functions: map[string]HostFunction{
					"keep": func(_ *Caller, args []Value) (*Value, error) { kept = args[0]; return nil, nil },
					"drop": func(c *Caller, args []Value) (*Value, error) { return nil, c.DropAsync(args[0].Resource) },
				}}
				in := asyncInstance(t, backend, name, imports)
				makeFn, _ := in.Function("make")
				write, _ := in.Function("write")
				drop, _ := in.Function("drop")
				other := asyncInstance(t, backend, name, imports)
				for _, hostDrop := range []bool{true, false} {
					value, err := makeFn.Call(nil, nil, CallOptions{})
					if err != nil || value.Value == nil {
						t.Fatal(value, err)
					}
					ref := value.Value.Resource
					if _, err := other.DropAsync(nil, ref, CallOptions{}); !errors.Is(err, vm.ErrReference) {
						t.Fatal("foreign endpoint", err)
					}
					if hostDrop {
						_, err = in.DropAsync(nil, ref, CallOptions{})
					} else {
						_, err = drop.Call(nil, []Value{*value.Value}, CallOptions{})
					}
					if err != nil {
						t.Fatal(err)
					}
					if _, err := in.DropAsync(nil, ref, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("stale endpoint", err)
					}
					result, err := write.Call(nil, nil, CallOptions{})
					if err != nil || result.Value == nil || result.Value.Bits != 1 {
						t.Fatal("peer drop", result, err)
					}
				}
				if _, err := in.DropAsync(nil, nil, CallOptions{}); err == nil {
					t.Fatal("nil endpoint")
				}
				if err := new(Caller).DropAsync(nil); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal(err)
				}
				// A successful imported ownership effect survives a later consumer trap.
				leak, _ := in.Function("leak")
				if _, err := leak.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal(err)
				}
				if _, err := in.DropAsync(nil, kept.Resource, CallOptions{}); err != nil {
					t.Fatal("imported effect rolled back", err)
				}
				probe := canonical.NewResourceTable(in.resourceBudget)
				for range in.config.MaxResources {
					if _, err := probe.AddOpaque("quota probe"); err != nil {
						t.Fatal("failed transfer retained quota", err)
					}
				}
				probe.Close()
			}
		})
	}
}

func TestAsyncEndpointGenerativeIdentity(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in := asyncInstance(t, backend, "async-future-generative", Imports{})
		makeOne, _ := in.Function("one", "make")
		takeOne, _ := in.Function("one", "take")
		takeTwo, _ := in.Function("two", "take")
		v, err := makeOne.Call(nil, nil, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := takeTwo.Call(nil, []Value{*v.Value}, CallOptions{}); !errors.Is(err, vm.ErrType) {
			t.Fatal("generative future identity escaped", err)
		}
		if _, err := takeOne.Call(nil, []Value{*v.Value}, CallOptions{}); err != nil {
			t.Fatal("rejected transfer consumed owner", err)
		}
	}
}
