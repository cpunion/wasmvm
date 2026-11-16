package component

import (
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/component/validate"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestBorrowedParametersPreserveReentryAndPublicSnapshots(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			var run, host *Function
			var expired []*Caller
			var retained [][]Value
			nested, public := false, false
			in := asyncInstance(t, backend, "canonical-frames", Imports{
				BorrowedParameters: map[string]bool{"step": true},
				Functions: map[string]HostFunction{"step": func(c *Caller, args []Value) (*Value, error) {
					host = c.function
					before := args[0].Bits
					copy := *c
					expired = append(expired, c, &copy)
					if public {
						retained = append(retained, args)
						args[0].Bits = 99
					} else if !nested {
						nested = true
						v, err := c.Call(run, []Value{{Bits: 1}, {Bits: 29}})
						nested = false
						runtime.GC()
						if err != nil || v.Bits != 29*17+5 || args[0].Bits != before {
							t.Fatal("reentry changed borrowed arguments", v, err)
						}
					}
					return &Value{Bits: uint64(uint32(before)*17 + 5)}, nil
				}},
			})
			run, _ = in.Function("run")
			for range 4 {
				got, err := run.Call(nil, []Value{{Bits: 3}, {Bits: 7}}, CallOptions{})
				want := uint32(7)
				for range 3 {
					want = want*17 + 5
				}
				if err != nil || got.Value.Bits != uint64(want) || checkCanonicalFramesCleared(t, in) != 2 {
					t.Fatal(got, err)
				}
			}
			// A public call to the same annotated import must still copy input.
			public = true
			args := []Value{{Bits: 23}}
			for range 2 {
				got, err := host.Call(nil, args, CallOptions{})
				if err != nil || got.Value.Bits != 23*17+5 || args[0].Bits != 23 {
					t.Fatal("public call borrowed caller storage", got, args, err)
				}
			}
			retained[1][0].Bits = 101
			if retained[0][0].Bits != 99 {
				t.Fatal("public calls share their parameter snapshots")
			}
			for _, c := range expired {
				if _, err := c.Call(run, nil); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("borrowed callback Caller revived", err)
				}
			}
		})
	}
}

func TestBorrowedAggregateParametersKeepResultAndQuotaOwnership(t *testing.T) {
	in := asyncInstance(t, "interpreter", "pending-callback", Imports{})
	element := &model.Type{Kind: model.U32}
	typ := &model.Type{Kind: model.List, Element: element}
	f := &Function{instance: in, typ: &validate.Type{Kind: model.Function, Parameters: []model.Field{{Type: typ}}, Result: typ}}
	args := []Value{{Elements: []Value{{Bits: 17}, {Bits: 23}}}}
	_, err := in.store.WithCaller(nil, CallOptions{}, func(core *vm.Caller) error {
		owned, err := f.hostParameters(core, args, false, nil)
		if err != nil {
			return err
		}
		values, bytes := owned.values, owned.bytes
		owned.args[0].Elements[0].Bits = 41
		if args[0].Elements[0].Bits != 17 {
			t.Fatal("owned callback aliases input aggregate")
		}
		if err := owned.finish(core, true, nil); err != nil {
			return err
		}
		borrowed, err := f.hostParameters(core, args, true, nil)
		if err != nil {
			return err
		}
		if borrowed.values != values || borrowed.bytes != bytes || &borrowed.args[0] != &args[0] || &borrowed.args[0].Elements[0] != &args[0].Elements[0] {
			t.Fatal("borrowed retention changed storage or logical accounting")
		}
		result, err := borrowed.accept(&borrowed.args[0])
		if err != nil {
			return err
		}
		if err := borrowed.finish(core, true, nil); err != nil {
			return err
		}
		clear(args[0].Elements)
		if result.Elements[0].Bits != 17 || result.Elements[1].Bits != 23 {
			t.Fatal("result retained borrowed aggregate storage")
		}
		in.config.MaxBytes = bytes - 1
		if _, err := f.hostParameters(core, args, true, nil); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("borrowed arguments bypassed transfer limit", err)
		}
		if in.heldValues != 0 || in.heldBytes != 0 {
			t.Fatal("failed borrowed retention leaked budget")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBorrowedParametersRejectAsyncBindings(t *testing.T) {
	e, err := NewEngine(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "async-host-completion"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	_, err = m.Instantiate(nil, Imports{BorrowedParameters: map[string]bool{"add": true}}, CallOptions{})
	if !errors.Is(err, vm.ErrType) {
		t.Fatal("async type accepted callback-scoped parameters", err)
	}
}
