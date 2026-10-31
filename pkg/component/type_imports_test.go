package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func TestEqualityBoundTypeImportsExecute(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			for _, name := range []string{"type-imports", "type-imports-nested"} {
				t.Run(name, func(t *testing.T) {
					m, err := e.Compile(nil, fixture(t, name))
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close()
					imports := Imports{}
					if name == "type-imports-nested" {
						imports.Instances = map[string]Imports{"types": {}}
					}
					in, err := m.Instantiate(nil, imports, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer in.Close()
					paths := [][]string{{"add"}}
					if name == "type-imports" {
						paths = append(paths, []string{"api", "add"})
					}
					for _, path := range paths {
						f, err := in.Function(path...)
						if err != nil {
							t.Fatal(err)
						}
						got, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
						if err != nil || got.Value.Bits != 9 {
							t.Fatal(got, err)
						}
					}
				})
			}
		})
	}
}

func TestEqualityBoundFunctionTypeRetainsNominalResource(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	m, err := e.Compile(nil, fixture(t, "type-imports-resources"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	r := NewHostResourceType(nil)
	calls := 0
	imports := Imports{Resources: map[string]*HostResourceType{"r": r}, Functions: map[string]HostFunction{
		"get": func(c *Caller, args []Value) (*Value, error) {
			calls++
			rep, err := c.ResourceRep(r, args[0].Resource)
			return &Value{Bits: uint64(rep)}, err
		},
		"consume": func(c *Caller, args []Value) (*Value, error) {
			rep, err := c.ResourceRep(r, args[0].Resource)
			if err == nil {
				err = c.DropResource(args[0].Resource)
			}
			return &Value{Bits: uint64(rep)}, err
		},
		"create": func(c *Caller, args []Value) (*Value, error) { v, err := c.NewResource(r, 7); return &v, err },
	}, FunctionChecks: map[string]FunctionCheck{"get": func(f FunctionInfo) error {
		_, v := f.Parameter(0)
		if f.ParameterCount() != 1 || !v.IsResource(r) {
			return vm.ErrType
		}
		return nil
	}}}
	in, err := m.Instantiate(nil, imports, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	v, err := in.NewResource(nil, r, 42)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := in.Function("get")
	got, err := f.Call(nil, []Value{v}, CallOptions{})
	if err != nil || got.Value.Bits != 42 || calls != 1 {
		t.Fatal(got, calls, err)
	}
	if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
		t.Fatal(err)
	}
	imports.FunctionChecks["get"] = func(FunctionInfo) error { return vm.ErrType }
	if in, err := m.Instantiate(nil, imports, CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
		t.Fatal(in, err)
	}
	if calls != 1 {
		t.Fatal("provider check had guest effects", calls)
	}
}
