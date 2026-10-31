package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncLowerResourceTransfersAndRollback(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"async-lower-resources", "async-lower-resources64", "async-lower-resource-sync", "async-lower-resource-sync64"} {
			for _, mode := range []string{"success", "unaligned", "range", "error", "panic", "cancel", "shape"} {
				t.Run(backend+"/"+name+"/"+mode, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					var drops []uint32
					typ := NewHostResourceType(func(_ *Caller, rep uint32) error { drops = append(drops, rep); return nil })
					var saved *Caller
					var borrowed Value
					called := 0
					imports := Imports{Resources: map[string]*HostResourceType{"r": typ}, AsyncFunctions: map[string]HostFunction{"make": func(c *Caller, args []Value) (*Value, error) {
						called++
						saved, borrowed = c, args[0]
						rep, err := c.ResourceRep(typ, args[0].Resource)
						if err != nil {
							return nil, err
						}
						value, err := c.NewResource(typ, rep+1)
						if err != nil {
							return nil, err
						}
						switch mode {
						case "error":
							return nil, wasm.ErrTrap
						case "panic":
							panic("async lower host")
						case "cancel":
							cancel()
						case "shape":
							return &Value{Elements: []Value{value}}, nil
						}
						return &Value{Elements: []Value{value, {Bits: 42}}}, nil
					}}}
					in := asyncInstance(t, backend, name, imports)
					arg, err := in.NewResource(nil, typ, 7)
					if err != nil {
						t.Fatal(err)
					}
					f, _ := in.Function("call")
					ptr := uint64(32)
					if mode == "unaligned" {
						ptr = 2
					}
					if mode == "range" {
						ptr = 65532
						if name == "async-lower-resources64" || name == "async-lower-resource-sync64" {
							ptr = 1 << 32
						}
					}
					var result CallResult
					if mode == "panic" {
						func() {
							defer func() {
								if recover() == nil {
									t.Error("missing host panic")
								}
							}()
							result, err = f.Call(ctx, []Value{arg, {Bits: ptr}}, CallOptions{})
						}()
					} else {
						result, err = f.Call(ctx, []Value{arg, {Bits: ptr}}, CallOptions{})
						if (err == nil) != (mode == "success") {
							t.Fatal(result, err)
						}
						if mode == "cancel" && !errors.Is(err, context.Canceled) {
							t.Fatal(err)
						}
						if (mode == "error" || mode == "range" || mode == "unaligned") && !errors.Is(err, wasm.ErrTrap) {
							t.Fatal(err)
						}
					}
					if called != 1 {
						t.Fatal("host effect lost", called)
					}
					if mode == "success" {
						if len(drops) != 0 || result.Value.Elements[1].Bits != 42 {
							t.Fatal(result, drops)
						}
						if rep, err := in.ResourceRep(nil, typ, result.Value.Elements[0].Resource); err != nil || rep != 8 {
							t.Fatal(rep, err)
						}
						if _, err := in.DropResource(nil, result.Value.Elements[0].Resource, CallOptions{}); err != nil {
							t.Fatal(err)
						}
					}
					if len(drops) != 1 || drops[0] != 8 {
						t.Fatal("unpublished ownership retained", drops)
					}
					if _, err := saved.ResourceRep(typ, borrowed.Resource); err == nil {
						t.Fatal("expired callback remained active")
					}
					if _, err := in.ResourceRep(nil, typ, borrowed.Resource); err == nil {
						t.Fatal("borrowed value escaped")
					}
					if rep, err := in.ResourceRep(nil, typ, arg.Resource); err != nil || rep != 7 {
						t.Fatal("source loan lost", rep, err)
					}
					if _, err := in.DropResource(nil, arg.Resource, CallOptions{}); err != nil {
						t.Fatal("source remained lent", err)
					}
					table := canonical.NewResourceTable(in.resourceBudget)
					for range in.config.MaxResources {
						if _, err := table.New(&canonical.ResourceType{}, 1); err != nil {
							t.Fatal("handle quota leaked", err)
						}
					}
					table.Close()
				})
			}
		}
	}
}
