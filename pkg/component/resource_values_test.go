package component

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func resourceValuesEngine(t *testing.T, backend string) *Engine {
	t.Helper()
	cfg := DefaultConfig()
	cfg.Core.Backend = backend
	e, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	return e
}

func TestResourceValuesAndStarts(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			for _, name := range []string{"resource-value-import", "resource-value-start", "resource-value-child", "resource-value-pair", "resource-value-empty", "resource-value-nested"} {
				t.Run(name, func(t *testing.T) {
					m, err := e.Compile(nil, fixture(t, name))
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close()
					drops := uint32(0)
					rt := NewHostResourceType(func(c *Caller, rep uint32) error { drops += rep; return nil })
					var old Value
					var escaped *Caller
					passCalls := 0
					imports := Imports{
						Resources: map[string]*HostResourceType{"r": rt},
						ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) {
							escaped = c
							a, err := c.NewResource(rt, 42)
							if err != nil {
								return Value{}, err
							}
							old = a
							if name == "resource-value-pair" {
								b, err := c.NewResource(rt, 7)
								return Value{Elements: []Value{a, b}}, err
							}
							if name == "resource-value-empty" {
								if err := c.DropResource(a.Resource); err != nil {
									return Value{}, err
								}
								return Value{}, nil
							}
							return a, nil
						}},
						ValueChecks: map[string]func(ValueType) error{"input": func(v ValueType) error {
							if !v.Valid() {
								return vm.ErrType
							}
							return nil
						}},
						Functions: map[string]HostFunction{"pass": func(c *Caller, args []Value) (*Value, error) {
							passCalls++
							v := args[0]
							return &v, nil
						}},
					}
					if name == "resource-value-nested" {
						imports = Imports{Instances: map[string]Imports{"box": imports}}
					}
					in, err := m.Instantiate(nil, imports, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer in.Close()
					if name == "resource-value-start" && passCalls != 1 {
						t.Fatal(passCalls)
					}
					if _, err := escaped.NewResource(rt, 1); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal(err)
					}
					if _, err := in.ResourceRep(nil, rt, old.Resource); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("old owner survived", err)
					}
					path := []string{"output"}
					if name == "resource-value-nested" {
						path = []string{"box", "input"}
					}
					v, err := in.Value(nil, path...)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := in.Value(nil, path...); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("value read twice", err)
					}
					if name == "resource-value-empty" {
						if drops != 42 {
							t.Fatal(drops)
						}
						return
					}
					values := []Value{v}
					if name == "resource-value-pair" {
						values = v.Elements
					}
					for _, v := range values {
						rep, err := in.ResourceRep(nil, rt, v.Resource)
						if err != nil || rep != 42 && rep != 7 {
							t.Fatal(rep, err)
						}
						if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
							t.Fatal(err)
						}
					}
					want := uint32(42)
					if name == "resource-value-pair" {
						want = 49
					}
					if drops != want {
						t.Fatal(drops, want)
					}
				})
			}
			m, err := e.Compile(nil, fixture(t, "resource-value-guest-start"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			m.Close()
			if err := in.store.Collect(); err != nil {
				t.Fatal(err)
			}
			v, err := in.Value(nil, "output")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			total, _ := in.Function("total")
			got, err := total.Call(nil, nil, CallOptions{})
			if err != nil || got.Value.Bits != 42 {
				t.Fatal(got, err)
			}
		})
	}
}

func TestResourceValueBorrowInitializationLifetime(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	m, err := e.Compile(nil, fixture(t, "resource-value-borrow"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	rt := NewHostResourceType(nil)
	var lender, borrowed Value
	in, err := m.Instantiate(nil, Imports{
		Resources:      map[string]*HostResourceType{"r": rt},
		ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) { var err error; lender, err = c.NewResource(rt, 42); return lender, err }},
		Functions: map[string]HostFunction{"get": func(c *Caller, args []Value) (*Value, error) {
			borrowed = args[0]
			if err := c.DropResource(lender.Resource); !errors.Is(err, wasm.ErrTrap) {
				return nil, fmt.Errorf("unpinned lender: %v", err)
			}
			rep, err := c.ResourceRep(rt, borrowed.Resource)
			return &Value{Bits: uint64(rep)}, err
		}},
	}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	answer, err := in.Value(nil, "answer")
	if err != nil || answer.Bits != 42 {
		t.Fatal(answer, err)
	}
	if _, err := in.ResourceRep(nil, rt, borrowed.Resource); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("borrow escaped", err)
	}
	if _, err := in.DropResource(nil, lender.Resource, CallOptions{}); err != nil {
		t.Fatal("lender stayed pinned", err)
	}
	if _, err := e.Compile(nil, fixture(t, "resource-value-borrow-export")); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("borrow export accepted", err)
	}
}

func TestResourceValueFailuresAreAtomic(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	m, err := e.Compile(nil, fixture(t, "resource-value-pair"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, mode := range []string{"duplicate", "error", "panic", "cancel", "shape", "foreign", "check", "missing", "ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			drops := 0
			rt := NewHostResourceType(func(c *Caller, rep uint32) error { drops++; return nil })
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			created := false
			var escaped *Caller
			imports := Imports{Resources: map[string]*HostResourceType{"r": rt}, ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) {
				created = true
				escaped = c
				v, err := c.NewResource(rt, 42)
				if err != nil {
					return Value{}, err
				}
				switch mode {
				case "error":
					return Value{}, vm.ErrType
				case "panic":
					panic("factory failure")
				case "cancel":
					cancel()
				case "shape":
					return v, nil
				}
				return Value{Elements: []Value{v, v}}, nil
			}}}
			switch mode {
			case "check":
				imports.ValueChecks = map[string]func(ValueType) error{"input": func(ValueType) error { return vm.ErrType }}
			case "missing":
				imports.ValueFactories = nil
			case "ambiguous":
				imports.Values = map[string]Value{"input": {}}
			case "foreign":
				other, err := e.Compile(nil, fixture(t, "resource-value-import"))
				if err != nil {
					t.Fatal(err)
				}
				defer other.Close()
				foreign, err := other.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": rt}, ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) { return c.NewResource(rt, 7) }}}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer foreign.Close()
				v, err := foreign.Value(nil, "output")
				if err != nil {
					t.Fatal(err)
				}
				imports.ValueFactories = nil
				imports.Values = map[string]Value{"input": {Elements: []Value{v, v}}}
				defer func() {
					if _, err := foreign.ResourceRep(nil, rt, v.Resource); err != nil {
						t.Fatal("foreign owner consumed", err)
					}
				}()
			}
			func() {
				defer func() {
					r := recover()
					if mode == "panic" && r == nil {
						t.Error("panic disappeared")
					}
					if mode != "panic" && r != nil {
						panic(r)
					}
				}()
				in, err := m.Instantiate(ctx, imports, CallOptions{})
				if in != nil || err == nil {
					t.Fatal("failed import published", in, err)
				}
			}()
			want := 0
			if created {
				want = 1
			}
			if drops != want {
				t.Fatal("cleanup", drops, want)
			}
			if escaped != nil {
				if _, err := escaped.NewResource(rt, 1); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestResourceValueStartFailureCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			for _, mode := range []string{"error", "cancel", "panic", "destructor-error"} {
				t.Run(mode, func(t *testing.T) {
					m, err := e.Compile(nil, fixture(t, "resource-value-host-start-failure"))
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close()
					drops := 0
					rt := NewHostResourceType(func(c *Caller, rep uint32) error {
						if rep != 42 {
							t.Error(rep)
						}
						drops++
						if mode == "destructor-error" {
							return wasm.ErrTrap
						}
						return nil
					})
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					imports := Imports{Resources: map[string]*HostResourceType{"r": rt}, ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) { return c.NewResource(rt, 42) }}, Functions: map[string]HostFunction{
						"pass": func(c *Caller, args []Value) (*Value, error) { v := args[0]; return &v, nil },
						"fail": func(c *Caller, args []Value) (*Value, error) {
							if mode == "cancel" {
								cancel()
								return nil, nil
							}
							if mode == "panic" {
								panic("later start")
							}
							return nil, vm.ErrType
						},
					}}
					func() {
						defer func() {
							r := recover()
							if mode == "panic" && r != "later start" {
								t.Error(r)
							}
							if mode != "panic" && r != nil {
								panic(r)
							}
						}()
						in, err := m.Instantiate(ctx, imports, CallOptions{})
						if in != nil || err == nil {
							t.Fatal(in, err)
						}
					}()
					if drops != 1 {
						t.Fatal("unpublished result destructor", drops)
					}
				})
			}
		})
	}
}

func TestResourceValueLimitsAndFactoryBounds(t *testing.T) {
	for _, mode := range []string{"resources", "bytes", "values", "literal", "equality", "equal-success", "raw-success", "check-cancel", "start-copy"} {
		t.Run(mode, func(t *testing.T) {
			cfg := DefaultConfig()
			fixtureName := "resource-value-pair"
			switch mode {
			case "resources":
				cfg.MaxResources = 1
			case "bytes":
				cfg.MaxBytes = 255
			case "values":
				cfg.MaxValues = 1
			case "literal":
				fixtureName = "value-constant-bound"
			case "equality", "equal-success", "raw-success":
				fixtureName = "value-imports"
			case "check-cancel":
				fixtureName = "resource-value-import"
			case "start-copy":
				fixtureName = "resource-value-start"
				cfg.MaxBytes = 255
			}
			e, err := NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, fixtureName))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			drops := 0
			rt := NewHostResourceType(func(*Caller, uint32) error { drops++; return nil })
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			imports := Imports{Resources: map[string]*HostResourceType{"r": rt}, ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) {
				a, err := c.NewResource(rt, 42)
				if err != nil {
					return Value{}, err
				}
				if mode == "start-copy" {
					return a, nil
				}
				b, err := c.NewResource(rt, 7)
				return Value{Elements: []Value{a, b}}, err
			}}, Functions: map[string]HostFunction{"pass": func(c *Caller, args []Value) (*Value, error) { v := args[0]; return &v, nil }}}
			if mode == "literal" {
				imports.ValueFactories["input"] = func(*Caller) (Value, error) { return Value{Bits: 43}, nil }
			}
			if mode == "equality" || mode == "equal-success" || mode == "raw-success" {
				imports.ValueFactories = map[string]ValueFactory{}
				makeValue := func(n uint64) Value { return Value{Elements: []Value{{Bits: n}, {Elements: []Value{{Text: "hello"}}}}} }
				n := uint64(8)
				if mode != "equality" {
					n = 7
				}
				imports.ValueFactories["first"] = func(*Caller) (Value, error) { return makeValue(7), nil }
				imports.ValueFactories["same"] = func(*Caller) (Value, error) { return makeValue(n), nil }
				if mode == "raw-success" {
					delete(imports.ValueFactories, "same")
					imports.Values = map[string]Value{"same": makeValue(7)}
				}
			}
			if mode == "check-cancel" {
				imports.ValueChecks = map[string]func(ValueType) error{"input": func(v ValueType) error { cancel(); return nil }}
			}
			in, err := m.Instantiate(ctx, imports, CallOptions{})
			if mode == "equal-success" || mode == "raw-success" {
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				v, err := in.Value(nil, "values", "first")
				if err != nil || v.Elements[0].Bits != 7 {
					t.Fatal(v, err)
				}
				return
			}
			if in != nil || err == nil {
				t.Fatal("limit/bound accepted", in, err)
			}
			want := 2
			if mode == "resources" || mode == "start-copy" {
				want = 1
			}
			if mode == "literal" || mode == "equality" || mode == "check-cancel" {
				want = 0
			}
			if drops != want {
				t.Fatal("cleanup", drops, want, err)
			}
		})
	}
}

func TestResourceValueGuestDestructorDuringFailedInitialization(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			m, err := e.Compile(nil, fixture(t, "resource-value-start-failure"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			drops := uint64(0)
			imports := Imports{Functions: map[string]HostFunction{
				"observe": func(c *Caller, args []Value) (*Value, error) { drops += args[0].Bits; return nil, nil },
				"fail":    func(c *Caller, args []Value) (*Value, error) { return nil, vm.ErrType },
			}}
			if in, err := m.Instantiate(nil, imports, CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
				t.Fatal(in, err)
			}
			if drops != 42 {
				t.Fatal("guest destructor skipped", drops)
			}
		})
	}
}

func TestResourceValueConcurrentExtraction(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	m, err := e.Compile(nil, fixture(t, "resource-value-import"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	rt := NewHostResourceType(nil)
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": rt}, ValueFactories: map[string]ValueFactory{"input": func(c *Caller) (Value, error) { return c.NewResource(rt, 42) }}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := in.Value(ctx, "output"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	type result struct {
		v   Value
		err error
	}
	gate := make(chan struct{})
	done := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() { <-gate; v, err := in.Value(nil, "output"); done <- result{v, err} }()
	}
	close(gate)
	success := 0
	var owned Value
	for i := 0; i < 2; i++ {
		r := <-done
		if r.err == nil {
			success++
			owned = r.v
		} else if !errors.Is(r.err, vm.ErrBusy) && !errors.Is(r.err, wasm.ErrTrap) {
			t.Fatal(r.err)
		}
	}
	if success != 1 {
		t.Fatal("duplicate extraction", success)
	}
	if _, err := in.DropResource(nil, owned.Resource, CallOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestResourceValueBorrowEqualityUsesCapabilityIdentity(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	m, err := e.Compile(nil, fixture(t, "resource-value-borrow-equality"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	rt := NewHostResourceType(nil)
	var lender Value
	factory := func(c *Caller) (Value, error) {
		if lender.Resource == nil {
			var err error
			lender, err = c.NewResource(rt, 42)
			return lender, err
		}
		return lender, nil
	}
	in, err := m.Instantiate(nil, Imports{Resources: map[string]*HostResourceType{"r": rt}, ValueFactories: map[string]ValueFactory{"first": factory, "same": factory}, Functions: map[string]HostFunction{"get": func(c *Caller, args []Value) (*Value, error) {
		rep, err := c.ResourceRep(rt, args[0].Resource)
		return &Value{Bits: uint64(rep)}, err
	}}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	for _, path := range []string{"first", "second"} {
		v, err := in.Value(nil, path)
		if err != nil || v.Bits != 42 {
			t.Fatal(v, err)
		}
	}
	if _, err := in.DropResource(nil, lender.Resource, CallOptions{}); err != nil {
		t.Fatal(err)
	}
}
