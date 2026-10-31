package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAsyncLiftValues(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"async-values", "async-values-utf16", "async-values-compact", "async-values64"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				in := asyncInstance(t, backend, name, Imports{})
				for range 5 {
					f, _ := in.Function("echo")
					v, err := f.Call(nil, []Value{{Text: "retained λ 😃"}}, CallOptions{})
					if err != nil || v.Value.Text != "retained λ 😃" {
						t.Fatal(v, err)
					}
					f, _ = in.Function("wide")
					v, err = f.Call(nil, nil, CallOptions{})
					if err != nil || len(v.Value.Elements) != 17 {
						t.Fatal(v, err)
					}
					for i, x := range v.Value.Elements {
						if x.Bits != uint64(i+1) {
							t.Fatal(i, x)
						}
					}
					f, _ = in.Function("spill")
					v, err = f.Call(nil, []Value{*v.Value}, CallOptions{})
					if err != nil || v.Value.Bits != 17 {
						t.Fatal(v, err)
					}
					f, _ = in.Function("empty")
					v, err = f.Call(nil, nil, CallOptions{})
					if err != nil || v.Value != nil {
						t.Fatal(v, err)
					}
				}
				for _, entry := range []string{"bad-pointer", "bad-memory", "bad-encoding"} {
					t.Run(entry, func(t *testing.T) {
						fresh := asyncInstance(t, backend, name, Imports{})
						f, _ := fresh.Function(entry)
						if _, err := f.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
							t.Fatal(err)
						}
					})
				}
			})
		}
	}
}

func TestAsyncHostContractsAndBorrowReturn(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			drops := 0
			typ := NewHostResourceType(func(*Caller, uint32) error { drops++; return nil })
			imports := Imports{Resources: map[string]*HostResourceType{"r": typ}, AsyncFunctions: map[string]HostFunction{"host": func(*Caller, []Value) (*Value, error) { return &Value{Bits: 42}, nil }}, Functions: map[string]HostFunction{"hook": func(*Caller, []Value) (*Value, error) { return nil, nil }}}
			in := asyncInstance(t, backend, "async-host", imports)
			f, _ := in.Function("host")
			if v, err := f.Call(nil, nil, CallOptions{}); err != nil || v.Value.Bits != 42 {
				t.Fatal(v, err)
			}
			owned, err := in.NewResource(nil, typ, 7)
			if err != nil {
				t.Fatal(err)
			}
			f, _ = in.Function("borrow")
			for range 5 {
				if _, err := f.Call(nil, []Value{owned}, CallOptions{}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := in.DropResource(nil, owned.Resource, CallOptions{}); err != nil || drops != 1 {
				t.Fatal(drops, err)
			}
			fresh := asyncInstance(t, backend, "async-host", imports)
			owned, _ = fresh.NewResource(nil, typ, 9)
			f, _ = fresh.Function("escaped")
			if _, err := f.Call(nil, []Value{owned}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			if _, err := fresh.DropResource(nil, owned.Resource, CallOptions{}); err != nil || drops != 2 {
				t.Fatal("lender not released", drops, err)
			}
			afterImports := imports
			var original Value
			afterImports.Functions = map[string]HostFunction{"hook": func(c *Caller, _ []Value) (*Value, error) { return nil, c.DropResource(original.Resource) }}
			after := asyncInstance(t, backend, "async-host", afterImports)
			original, err = after.NewResource(nil, typ, 11)
			if err != nil {
				t.Fatal(err)
			}
			f, _ = after.Function("borrow-after")
			if _, err := f.Call(nil, []Value{original}, CallOptions{}); err != nil || drops != 3 {
				t.Fatal("task.return retained a lender", drops, err)
			}
			for _, mode := range []string{"error", "panic", "cancel", "close"} {
				t.Run(mode, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					args := imports
					args.Functions = map[string]HostFunction{"hook": func(c *Caller, _ []Value) (*Value, error) {
						switch mode {
						case "error":
							return nil, vm.ErrType
						case "panic":
							panic("after task.return")
						case "cancel":
							cancel()
						case "close":
							c.CloseInstance()
						}
						return nil, nil
					}}
					fresh := asyncInstance(t, backend, "async-host", args)
					f, _ := fresh.Function("after")
					if mode == "panic" {
						func() {
							defer func() {
								if recover() == nil {
									t.Error("missing host panic")
								}
							}()
							f.Call(ctx, nil, CallOptions{})
						}()
					} else if _, err := f.Call(ctx, nil, CallOptions{}); err == nil {
						t.Fatal("published result after failure")
					}
				})
			}
			e, _ := NewEngine(DefaultConfig())
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "async-host"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			bad := imports
			bad.AsyncFunctions = nil
			if _, err := m.Instantiate(nil, bad, CallOptions{}); !errors.Is(err, vm.ErrLink) {
				t.Fatal("sync binding accepted for async type", err)
			}
			bad = imports
			bad.Functions = nil
			if _, err := m.Instantiate(nil, bad, CallOptions{}); !errors.Is(err, vm.ErrLink) {
				t.Fatal("async binding accepted for sync type", err)
			}
		})
	}
}

func TestAsyncOwnedResults(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in := asyncInstance(t, backend, "async-resources", Imports{})
		makeValue, _ := in.Function("make")
		consume, _ := in.Function("consume")
		for range 5 {
			v, err := makeValue.Call(nil, []Value{{Bits: 7}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			v, err = consume.Call(nil, []Value{*v.Value}, CallOptions{})
			if err != nil || v.Value.Bits != 7 {
				t.Fatal(v, err)
			}
		}
		f, _ := in.Function("total")
		if v, err := f.Call(nil, nil, CallOptions{}); err != nil || v.Value.Bits != 35 {
			t.Fatal(v, err)
		}
		f, _ = in.Function("bad")
		if _, err := f.Call(nil, []Value{{Bits: 9}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		// A trapped owner cannot execute guest destructors; unpublished host
		// ownership must still release every handle and its shared quota.
		table := canonical.NewResourceTable(in.resourceBudget)
		identity := &canonical.ResourceType{}
		for range in.config.MaxResources {
			if _, err := table.New(identity, 1); err != nil {
				t.Fatal("unpublished handle retained", err)
			}
		}
		if _, err := table.New(identity, 1); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		table.Close()
	}
}
