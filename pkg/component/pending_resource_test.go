package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPendingResourceCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"success", "unaligned", "range", "error", "panic", "cancel", "shape"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var drops []uint32
				typ := NewHostResourceType(func(_ *Caller, rep uint32) error { drops = append(drops, rep); return nil })
				var saved *Caller
				imports := Imports{Resources: map[string]*HostResourceType{"r": typ}, AsyncFunctions: map[string]HostFunction{"make": func(c *Caller, args []Value) (*Value, error) {
					saved = c
					rep, err := c.ResourceRep(typ, args[0].Resource)
					if err != nil {
						return nil, err
					}
					own, err := c.NewResource(typ, rep+1)
					if err != nil {
						return nil, err
					}
					switch mode {
					case "error":
						return nil, wasm.ErrTrap
					case "panic":
						panic("pending host")
					case "cancel":
						cancel()
					case "shape":
						return &Value{Elements: []Value{own}}, nil
					}
					return &Value{Elements: []Value{own, {Bits: 42}}}, nil
				}}}
				in := asyncInstance(t, backend, "pending-resources", imports)
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
				}
				var result CallResult
				if mode == "panic" {
					func() {
						defer func() {
							if recover() == nil {
								t.Error("missing panic")
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
				if _, err := saved.NewResource(typ, 1); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("expired caller", err)
				}
				if _, err := in.DropResource(nil, arg.Resource, CallOptions{}); err != nil {
					t.Fatal("loan retained", err)
				}
				table := canonical.NewResourceTable(in.resourceBudget)
				for range in.config.MaxResources {
					if _, err := table.New(&canonical.ResourceType{}, 1); err != nil {
						t.Fatal("quota leaked", err)
					}
				}
				table.Close()
				if in.tasks != nil {
					t.Fatal("continuation retained")
				}
			})
		}
	}
}
