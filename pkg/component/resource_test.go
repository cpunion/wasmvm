package component

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestResourceInstructions(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"resources", "resources64"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				config := DefaultConfig()
				config.Core.Backend = backend
				e, _ := NewEngine(config)
				defer e.Close()
				m, err := e.Compile(nil, fixture(t, name))
				if err != nil {
					t.Fatal(err)
				}
				defer m.Close()
				instantiate := func() *Instance {
					in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(in.Close)
					return in
				}
				in := instantiate()
				call := func(in *Instance, name string, args ...Value) (CallResult, error) {
					f, err := in.Function(name)
					if err != nil {
						return CallResult{}, err
					}
					return f.Call(nil, args, vm.CallOptions{})
				}
				rep := uint64(0xfedcba98)
				if name == "resources64" {
					rep |= 0x98765432 << 32
				}
				id, err := call(in, "new", Value{Bits: rep})
				if err != nil || id.Value.Bits != 1 {
					t.Fatal(id, err)
				}
				v, err := call(in, "rep", *id.Value)
				if err != nil || v.Value.Bits != rep {
					t.Fatal(v, err)
				}
				if _, err := call(in, "drop", *id.Value); err != nil {
					t.Fatal(err)
				}
				v, err = call(in, "total")
				if err != nil || v.Value.Bits != rep {
					t.Fatal("destructor did not execute", v, err)
				}
				v, err = call(in, "churn", Value{Bits: 3})
				if err != nil || v.Value.Bits != 3 {
					t.Fatal(v, err)
				}
				// Each failed call locks down its component, so use independent
				// instances for each negative assertion.
				for _, bad := range []string{"rep", "drop", "wrong"} {
					in := instantiate()
					id, err := call(in, "new", Value{Bits: 7})
					if err != nil {
						t.Fatal(err)
					}
					if bad != "wrong" {
						if _, err := call(in, "drop", *id.Value); err != nil {
							t.Fatal(err)
						}
					}
					if _, err := call(in, bad, *id.Value); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal(bad, "accepted stale or wrong-type handle", err)
					}
					if _, err := call(in, "total"); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("trap did not lock down component", err)
					}
				}
				in = instantiate()
				id, err = call(in, "new", Value{Bits: 13})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := call(in, "drop", *id.Value); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("destructor trap disappeared", err)
				}
				for _, table := range in.resourceTables {
					if err := table.Drain(nil); err != nil {
						t.Fatal("failed destructor resurrected handle", err)
					}
				}
			})
		}
	}
}

func TestResourceRoundTripAndLimits(t *testing.T) {
	config := DefaultConfig()
	config.MaxResources = 1
	e, _ := NewEngine(config)
	defer e.Close()
	for _, name := range []string{"resource-new", "resource-roundtrip", "resources"} {
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		m.Close() // Runtime declarations retain the instantiated Core code.
		switch name {
		case "resource-roundtrip":
			f, _ := in.Function("run")
			v, err := f.Call(nil, []Value{{Bits: 42}}, vm.CallOptions{})
			if err != nil || v.Value.Bits != 42 {
				t.Fatal(v, err)
			}
		case "resources":
			f, _ := in.Function("new")
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := f.Call(ctx, []Value{{Bits: 1}}, vm.CallOptions{}); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			v, err := f.Call(nil, []Value{{Bits: 1}}, vm.CallOptions{})
			if err != nil || v.Value.Bits != 1 {
				t.Fatal("cancellation created a handle", v, err)
			}
			if _, err := f.Call(nil, []Value{{Bits: 2}}, vm.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("resource limit ignored", err)
			}
		}
		in.Close()
		in.Close()
		if err := in.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerativeResourcesAndSharedLimits(t *testing.T) {
	for _, limit := range []uint32{1, 2} {
		config := DefaultConfig()
		config.MaxResources = limit
		e, _ := NewEngine(config)
		m, err := e.Compile(nil, fixture(t, "resource-generative"))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		a := in.exports["a"].instance["r"].resource
		b := in.exports["b"].instance["r"].resource
		if a.owner.plan != nil || b.owner.plan != nil {
			t.Fatal("runtime resource retained its construction plan")
		}
		if a == b || a.identity == b.identity || a.owner == b.owner {
			t.Fatal("repeated instantiation shared a resource identity")
		}
		for i, name := range []string{"a", "b"} {
			f, _ := in.Function(name, "new")
			id, err := f.Call(nil, []Value{{Bits: uint64(i + 7)}}, vm.CallOptions{})
			if i == 1 && limit == 1 {
				if !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("nested instance bypassed the aggregate quota", err)
				}
				continue
			}
			if err != nil || id.Value.Bits != 1 {
				t.Fatal("handle domains were not independent", id, err)
			}
			f, _ = in.Function(name, "rep")
			v, err := f.Call(nil, []Value{*id.Value}, vm.CallOptions{})
			if err != nil || v.Value.Bits != uint64(i+7) {
				t.Fatal(v, err)
			}
		}
		in.Close()
		for _, table := range in.resourceTables {
			if _, err := table.Rep(a.identity, 1); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("shutdown retained a handle domain", err)
			}
		}
		m.Close()
		e.Close()
	}
}

func TestResourcePostReturnGuards(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "resource-post-return"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, name := range []string{"allow", "reject-new", "reject-drop"} {
		in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f, _ := in.Function(name)
		v, err := f.Call(nil, []Value{{Bits: 7}}, vm.CallOptions{})
		if name == "allow" {
			if err != nil || v.Value.Bits != 1 {
				t.Fatal("resource.rep must remain available during post-return", v, err)
			}
		} else if !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("post-return acquired resource effects", name, err)
		}
		in.Close()
	}
}

func TestResourceInitializationFailureClosesStore(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, _ := NewEngine(config)
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "resource-start-failure"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			for _, mode := range []string{"error", "panic", "cancel"} {
				ctx, cancel := context.WithCancel(context.Background())
				var store *vm.Store
				var escaped *Caller
				var result *Instance
				var gotPanic any
				func() {
					defer func() { gotPanic = recover() }()
					result, err = m.Instantiate(ctx, Imports{Functions: map[string]HostFunction{
						"fail": func(c *Caller, _ []Value) (*Value, error) {
							store, escaped = c.core.Store(), c
							if mode == "panic" {
								panic("initialization panic")
							}
							if mode == "cancel" {
								cancel()
								return nil, nil
							}
							return nil, wasm.ErrTrap
						},
					}}, vm.CallOptions{})
				}()
				cancel()
				if result != nil || store == nil || escaped == nil {
					t.Fatal("failed initialization published an instance or missed the callback", mode)
				}
				if mode == "panic" {
					if gotPanic != "initialization panic" {
						t.Fatal(gotPanic)
					}
				} else if gotPanic != nil || err == nil {
					t.Fatal("initialization failure disappeared", mode, err, gotPanic)
				}
				wait, stop := context.WithTimeout(context.Background(), time.Second)
				if err := store.WaitClosed(wait); err != nil {
					t.Fatal("unpublished Store survived initialization failure", mode, err)
				}
				stop()
				if _, err := store.Usage(); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("failed initialization retained its Store", mode, err)
				}
				if _, err := escaped.Call(nil, nil); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("failed initialization retained an active Caller", mode, err)
				}
			}
		})
	}
}
