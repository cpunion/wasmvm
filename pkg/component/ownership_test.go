package component

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/pkg/vm"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestOwnedResourcesMoveBorrowAndDrop(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, _ := NewEngine(config)
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "ownership"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			call := func(name string, args ...Value) (CallResult, error) {
				f, err := in.Function(name)
				if err != nil {
					return CallResult{}, err
				}
				return f.Call(nil, args, CallOptions{})
			}
			v, err := call("make", Value{Bits: 42})
			if err != nil || v.Value == nil || v.Value.Resource == nil {
				t.Fatal(v, err)
			}
			m.Close()
			if err := in.store.Collect(); err != nil {
				t.Fatal(err)
			}
			owned := *v.Value
			alias := *owned.Resource
			for i := 0; i < 2; i++ {
				get, err := call("get", owned)
				if err != nil || get.Value.Bits != 42 {
					t.Fatal(get, err)
				}
			}
			moved, err := call("pass", owned)
			if err != nil || moved.Value.Resource == nil {
				t.Fatal(moved, err)
			}
			for _, stale := range []Value{owned, {Resource: &alias}} {
				if _, err := call("get", stale); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("moved capability survived", err)
				}
			}
			v, err = call("consume", *moved.Value)
			if err != nil || v.Value.Bits != 42 {
				t.Fatal(v, err)
			}
			total, err := call("total")
			if err != nil || total.Value.Bits != 42 {
				t.Fatal("destructor did not execute", total, err)
			}
			v, err = call("make", Value{Bits: 7})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := in.DropResource(nil, v.Value.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			if _, err := in.DropResource(nil, v.Value.Resource, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("double drop", err)
			}
			total, err = call("total")
			if err != nil || total.Value.Bits != 49 {
				t.Fatal(total, err)
			}
			if _, err := call("get", Value{Bits: 1}); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal("integer forged ownership", err)
			}
			if _, err := call("get", Value{Resource: new(Resource)}); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal("zero token forged ownership", err)
			}
		})
	}
}

func TestResourceImportEqualityAndStoreAuthority(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "ownership-equality"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	a, b := NewHostResourceType(nil), NewHostResourceType(nil)
	for _, imports := range []Imports{{}, {Resources: map[string]*HostResourceType{"a": new(HostResourceType), "b": a}}, {Resources: map[string]*HostResourceType{"a": a, "b": b}}} {
		if in, err := m.Instantiate(nil, imports, CallOptions{}); err == nil || in != nil {
			t.Fatal("invalid resource binding published", in, err)
		}
	}
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"a": a, "b": a}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	other, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"a": a, "b": a}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	for _, typ := range []*HostResourceType{nil, b, new(HostResourceType)} {
		if _, err := in.NewResource(nil, typ, 1); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
	}
	v, err := in.NewResource(nil, a, 42)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.ResourceRep(nil, a, v.Resource); !errors.Is(err, vm.ErrReference) {
		t.Fatal("resource crossed Store", err)
	}
	if _, err := other.DropResource(nil, v.Resource, CallOptions{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if _, err := in.ResourceRep(nil, b, v.Resource); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := in.DropResource(ctx, v.Resource, CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if rep, err := in.ResourceRep(nil, a, v.Resource); err != nil || rep != 42 {
		t.Fatal("canceled drop consumed ownership", rep, err)
	}
	if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
		t.Fatal(err)
	}
	in.Close()
	if _, err := in.NewResource(nil, a, 1); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestHostDestructorReentryIsBounded(t *testing.T) {
	config := DefaultConfig()
	config.Store.MaxReentry = 0
	config.MaxResources = 1
	e, _ := NewEngine(config)
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "resource-import"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var typ *HostResourceType
	calls := 0
	typ = NewHostResourceType(func(c *Caller, rep uint32) error {
		calls++
		v, err := c.NewResource(typ, rep+1)
		if err != nil {
			return err
		}
		return c.DropResource(v.Resource)
	})
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"resource": typ}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	v, err := in.NewResource(nil, typ, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.DropResource(nil, v.Resource, CallOptions{}); !errors.Is(err, wasm.ErrLimit) || calls != 1 {
		t.Fatal("unbounded destructor reentry", err, calls)
	}
	if _, err := in.NewResource(nil, typ, 1); err != nil {
		t.Fatal("reentry failure retained handles", err)
	}
}

func TestEscapedGuestBorrowIsInvalidated(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "ownership-borrow-escape"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	typ := NewHostResourceType(nil)
	var escaped Value
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{
		"get":     func(c *Caller, args []Value) (*Value, error) { escaped = args[0]; return &Value{Bits: 42}, nil },
		"consume": func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
		"create":  func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
	}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	v, err := in.NewResource(nil, typ, 42)
	if err != nil {
		t.Fatal(err)
	}
	get, _ := in.Function("get")
	if _, err := get.Call(nil, []Value{v}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("guest borrow escaped", err)
	}
	if _, err := in.ResourceRep(nil, typ, escaped.Resource); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if rep, err := in.ResourceRep(nil, typ, v.Resource); err != nil || rep != 42 {
		t.Fatal(rep, err)
	}
	if _, err := get.Call(nil, []Value{v}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("escape did not lock down component", err)
	}
	if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestResourceCompositionAndFailedResultCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, _ := NewEngine(config)
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "ownership-composition"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			makeLibrary, _ := in.Function("library", "make")
			v, err := makeLibrary.Call(nil, []Value{{Bits: 7}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			get, _ := in.Function("get")
			got, err := get.Call(nil, []Value{*v.Value}, CallOptions{})
			if err != nil || got.Value.Bits != 7 {
				t.Fatal(got, err)
			}
			if _, err := in.DropResource(nil, v.Value.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			makeRoot, _ := in.Function("make")
			if _, err := makeRoot.Call(nil, []Value{{Bits: 42}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("post-return did not trap", err)
			}
			total, _ := in.Function("total")
			got, err = total.Call(nil, nil, CallOptions{})
			if err != nil || got.Value.Bits != 49 {
				t.Fatal("unpublished result missed its defining destructor", got, err)
			}
			if _, err := makeRoot.Call(nil, []Value{{Bits: 1}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("failed wrapper remained enterable", err)
			}
		})
	}
}

func TestOwnedCM64Representation(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		config := DefaultConfig()
		config.Core.Backend = backend
		e, _ := NewEngine(config)
		m, err := e.Compile(nil, fixture(t, "ownership64"))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		makeResource, _ := in.Function("make-owned")
		consume, _ := in.Function("consume-owned")
		rep := uint64(0xfedcba9876543210)
		v, err := makeResource.Call(nil, []Value{{Bits: rep}}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		result, err := consume.Call(nil, []Value{*v.Value}, CallOptions{})
		if err != nil || result.Value.Bits != rep {
			t.Fatal(result, err)
		}
		if _, err := e.Compile(nil, fixture(t, "ownership64-borrow")); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal("unsupported local cm64 borrow reached instantiation", err)
		}
		in.Close()
		m.Close()
		e.Close()
	}
}

func TestResourceCallerSerializesAndRejectsComponentReentry(t *testing.T) {
	var zero Caller
	if _, err := zero.NewResource(nil, 0); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	var absent *Caller
	if _, err := absent.Call(nil, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "ownership-host"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	started, release := make(chan struct{}), make(chan struct{})
	typ := NewHostResourceType(func(c *Caller, rep uint32) error {
		if rep == 99 {
			close(started)
			<-release
		}
		return nil
	})
	var blocked Value
	var get *Function
	var escaped *Caller
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{
		"get": func(c *Caller, args []Value) (*Value, error) {
			escaped = c
			if _, err := c.Call(get, args); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("entered a suspended component", err)
			}
			done := make(chan error, 1)
			go func() { done <- c.DropResource(blocked.Resource) }()
			<-started
			if _, err := c.NewResource(typ, 1); !errors.Is(err, vm.ErrBusy) {
				t.Fatal(err)
			}
			if _, err := c.ResourceRep(typ, args[0].Resource); !errors.Is(err, vm.ErrBusy) {
				t.Fatal(err)
			}
			if err := c.DropResource(nil); !errors.Is(err, vm.ErrBusy) {
				t.Fatal(err)
			}
			if _, err := c.Call(nil, nil); !errors.Is(err, vm.ErrBusy) {
				t.Fatal(err)
			}
			close(release)
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			rep, err := c.ResourceRep(typ, args[0].Resource)
			return &Value{Bits: uint64(rep)}, err
		},
		"consume": func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
		"create":  func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
	}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	get, _ = in.Function("get")
	blocked, err = in.NewResource(nil, typ, 99)
	if err != nil {
		t.Fatal(err)
	}
	v, err := in.NewResource(nil, typ, 42)
	if err != nil {
		t.Fatal(err)
	}
	result, err := get.Call(nil, []Value{v}, CallOptions{})
	if err != nil || result.Value.Bits != 42 {
		t.Fatal(result, err)
	}
	if _, err := escaped.ResourceRep(typ, v.Resource); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := escaped.DropResource(v.Resource); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestTransferredInputTrapReleasesQuotaAndRunsHostDestructor(t *testing.T) {
	config := DefaultConfig()
	config.MaxResources = 1
	e, _ := NewEngine(config)
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "ownership-input-trap"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var dropped uint32
	typ := NewHostResourceType(func(c *Caller, rep uint32) error { dropped += rep; return nil })
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{
		"get":     func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
		"consume": func(*Caller, []Value) (*Value, error) { t.Fatal("trapping guest called host"); return nil, nil },
		"create":  func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
	}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	v, err := in.NewResource(nil, typ, 42)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.NewResource(nil, typ, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("quota not enforced", err)
	}
	f, _ := in.Function("consume")
	if _, err := f.Call(nil, []Value{v}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if dropped != 42 {
		t.Fatal("transferred input was stranded", dropped)
	}
	if _, err := in.NewResource(nil, typ, 1); err != nil {
		t.Fatal("failure retained quota", err)
	}
}

func TestPureHostReentryHasSharedDepthLimit(t *testing.T) {
	config := DefaultConfig()
	config.Store.MaxReentry = 2
	e, _ := NewEngine(config)
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "host-reexport"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var run *Function
	in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{"run": func(c *Caller, _ []Value) (*Value, error) { return c.Call(run, nil) }}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	run, _ = in.Function("run")
	if _, err := run.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func TestResourceFailureCleanup(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "ownership-host"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, operation := range []string{"get", "consume", "create"} {
		for _, mode := range []string{"error", "cancel", "panic", "wrong-result"} {
			t.Run(operation+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var dropped uint32
				typ := NewHostResourceType(func(c *Caller, rep uint32) error {
					dropped += rep
					if mode == "panic" && operation != "get" {
						panic("cleanup panic")
					}
					return nil
				})
				var escaped Value
				callback := func(c *Caller, args []Value) (*Value, error) {
					if operation == "create" {
						var err error
						escaped, err = c.NewResource(typ, 7)
						if err != nil {
							return nil, err
						}
					} else {
						escaped = args[0]
					}
					switch mode {
					case "error":
						return nil, wasm.ErrTrap
					case "panic":
						panic("resource panic")
					case "cancel":
						cancel()
						if _, err := c.NewResource(typ, 1); !errors.Is(err, context.Canceled) {
							t.Fatal(err)
						}
						if _, err := c.ResourceRep(typ, escaped.Resource); !errors.Is(err, context.Canceled) {
							t.Fatal(err)
						}
						if err := c.DropResource(escaped.Resource); !errors.Is(err, context.Canceled) {
							t.Fatal(err)
						}
					}
					return nil, nil
				}
				in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{"get": callback, "consume": callback, "create": callback}}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				var args []Value
				if operation != "create" {
					v, err := in.NewResource(nil, typ, 42)
					if err != nil {
						t.Fatal(err)
					}
					args = []Value{v}
				}
				fn, _ := in.Function(operation)
				var caught any
				func() { defer func() { caught = recover() }(); _, err = fn.Call(ctx, args, CallOptions{}) }()
				if mode == "panic" {
					if caught != "resource panic" {
						t.Fatal(caught)
					}
				} else if caught != nil || err == nil {
					t.Fatal("failure disappeared", caught, err)
				}
				if operation == "get" {
					if rep, err := in.ResourceRep(nil, typ, args[0].Resource); err != nil || rep != 42 {
						t.Fatal("borrow failure consumed ownership", rep, err)
					}
					if dropped != 0 {
						t.Fatal(dropped)
					}
					if _, err := in.DropResource(nil, args[0].Resource, CallOptions{}); err != nil {
						t.Fatal(err)
					}
				} else {
					want := uint32(42)
					if operation == "create" {
						want = 7
					}
					if dropped != want {
						t.Fatal("unpublished ownership missed its host destructor", dropped, want)
					}
				}
				if _, err := in.ResourceRep(nil, typ, escaped.Resource); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("failed call kept an escaped capability", err)
				}
			})
		}
	}
}

func TestHostResourceBindingsAndBorrowLifetimes(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, _ := NewEngine(config)
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "ownership-host"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			var dropped uint32
			var destructorCaller *Caller
			typ := NewHostResourceType(func(c *Caller, rep uint32) error { destructorCaller = c; dropped += rep; return nil })
			var escaped Value
			var escapedCaller *Caller
			imports := Imports{Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{
				"get": func(c *Caller, args []Value) (*Value, error) {
					escaped = args[0]
					copied := *c
					escapedCaller = &copied
					rep, err := c.ResourceRep(typ, args[0].Resource)
					if err != nil {
						return nil, err
					}
					if err := c.DropResource(args[0].Resource); err == nil {
						t.Fatal("borrow became owned")
					}
					return &Value{Bits: uint64(rep)}, nil
				},
				"consume": func(c *Caller, args []Value) (*Value, error) {
					rep, err := c.ResourceRep(typ, args[0].Resource)
					if err != nil {
						return nil, err
					}
					err = c.DropResource(args[0].Resource)
					return &Value{Bits: uint64(rep)}, err
				},
				"create": func(c *Caller, args []Value) (*Value, error) { v, err := c.NewResource(typ, 7); return &v, err },
			}}
			in, err := m.Instantiate(nil, imports, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			v, err := in.NewResource(nil, typ, 42)
			if err != nil {
				t.Fatal(err)
			}
			get, _ := in.Function("get")
			for i := 0; i < 2; i++ {
				res, err := get.Call(nil, []Value{v}, CallOptions{})
				if err != nil || res.Value.Bits != 42 {
					t.Fatal(res, err)
				}
			}
			if _, err := in.ResourceRep(nil, typ, escaped.Resource); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("borrow escaped callback", err)
			}
			if _, err := escapedCaller.NewResource(typ, 1); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
			consume, _ := in.Function("consume")
			res, err := consume.Call(nil, []Value{v}, CallOptions{})
			if err != nil || res.Value.Bits != 42 || dropped != 42 {
				t.Fatal(res, err, dropped)
			}
			if _, err := destructorCaller.NewResource(typ, 1); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("destructor Caller escaped", err)
			}
			create, _ := in.Function("create")
			res, err = create.Call(nil, nil, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if rep, err := in.ResourceRep(nil, typ, res.Value.Resource); err != nil || rep != 7 {
				t.Fatal(rep, err)
			}
			if _, err := in.DropResource(nil, res.Value.Resource, CallOptions{}); err != nil || dropped != 49 {
				t.Fatal(err, dropped)
			}
			copyType := *typ
			v, err = in.NewResource(nil, &copyType, 1)
			if err != nil {
				t.Fatal("copy changed nominal identity", err)
			}
			if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResourceContainersAndDuplicateArguments(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, _ := NewEngine(config)
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "ownership-containers"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			typ := NewHostResourceType(nil)
			in, err := m.Instantiate(nil, Imports{Instances: map[string]Imports{"host": {
				Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{
					"echo": func(c *Caller, args []Value) (*Value, error) { return &args[0], nil },
				},
			}}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			refs := make([]Value, 13)
			for i := range refs {
				refs[i], err = in.NewResource(nil, typ, uint32(i+1))
				if err != nil {
					t.Fatal(err)
				}
			}
			input := Value{Elements: []Value{
				{Elements: []Value{refs[0]}},
				{Elements: []Value{{Elements: []Value{{Bits: 5}, refs[1]}}}},
				{Case: 1, Elements: []Value{refs[2]}},
				{Case: 1, Elements: []Value{refs[3]}},
				{Elements: append([]Value(nil), refs[4:12]...)},
				{Elements: []Value{refs[12]}},
			}}
			echo, _ := in.Function("echo")
			input.Elements[4].Elements[1] = refs[4]
			if _, err := echo.Call(nil, []Value{input}, CallOptions{}); err == nil {
				t.Fatal("duplicate ownership accepted")
			}
			for i, v := range refs {
				if rep, err := in.ResourceRep(nil, typ, v.Resource); err != nil || rep != uint32(i+1) {
					t.Fatal("validation consumed an argument", rep, err)
				}
			}
			input.Elements[4].Elements[1] = refs[5]
			result, err := echo.Call(nil, []Value{input}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for _, old := range refs {
				if _, err := in.ResourceRep(nil, typ, old.Resource); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("aggregate move kept source authority", err)
				}
			}
			out := result.Value.Elements
			got := []Value{out[0].Elements[0], out[1].Elements[0].Elements[1], out[2].Elements[0], out[3].Elements[0]}
			got = append(got, out[4].Elements...)
			got = append(got, out[5].Elements[0])
			for i, v := range got {
				if rep, err := in.ResourceRep(nil, typ, v.Resource); err != nil || rep != uint32(i+1) {
					t.Fatal(rep, err)
				}
				if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
