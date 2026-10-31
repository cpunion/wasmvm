package component

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostComponentsExecuteAndRetainCode(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			for _, name := range []string{"host-components", "host-components-nested"} {
				t.Run(name, func(t *testing.T) {
					code, err := e.Compile(nil, fixture(t, "host-component-guest"))
					if err != nil {
						t.Fatal(err)
					}
					m, err := e.Compile(nil, fixture(t, name))
					if err != nil {
						t.Fatal(err)
					}
					imports := Imports{Components: map[string]*Compiled{"code": code}}
					if name == "host-components-nested" {
						imports = Imports{Instances: map[string]Imports{"host": imports}}
					}
					in, err := m.Instantiate(nil, imports, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer in.Close()
					code.Close()
					m.Close()
					if err := in.store.Collect(); err != nil {
						t.Fatal(err)
					}
					f, err := in.Function("add")
					if err != nil {
						t.Fatal(err)
					}
					for i := uint64(0); i < 3; i++ {
						got, err := f.Call(nil, []Value{{Bits: i}, {Bits: 5}}, CallOptions{})
						if err != nil || got.Value.Bits != i+5 {
							t.Fatal(got, err)
						}
					}
					if name == "host-components" {
						f, err := in.Function("twice")
						if err != nil {
							t.Fatal(err)
						}
						got, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
						if err != nil || got.Value.Bits != 10 {
							t.Fatal(got, err)
						}
						if _, err := in.Function("api", "extra"); !errors.Is(err, vm.ErrLink) {
							t.Fatal("wider component export escaped", err)
						}
					}
					if _, err := m.Instantiate(nil, imports, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal(err)
					}
				})
			}
		})
	}
}

func TestHostComponentResourceGenerativity(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			code, err := e.Compile(nil, fixture(t, "ownership"))
			if err != nil {
				t.Fatal(err)
			}
			defer code.Close()
			m, err := e.Compile(nil, fixture(t, "host-components-resources"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{Components: map[string]*Compiled{"code": code}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			call := func(side, name string, args ...Value) (CallResult, error) {
				f, err := in.Function(side, name)
				if err != nil {
					return CallResult{}, err
				}
				return f.Call(nil, args, CallOptions{})
			}
			left, err := call("left", "make", Value{Bits: 42})
			if err != nil {
				t.Fatal(err)
			}
			right, err := call("right", "make", Value{Bits: 7})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := call("right", "get", *left.Value); !errors.Is(err, vm.ErrType) {
				t.Fatal("generative resource crossed instances", err)
			}
			for _, v := range []struct {
				side  string
				value Value
				want  uint64
			}{{"left", *left.Value, 42}, {"right", *right.Value, 7}} {
				got, err := call(v.side, "get", v.value)
				if err != nil || got.Value.Bits != v.want {
					t.Fatal(got, err)
				}
				if _, err := in.DropResource(nil, v.value.Resource, CallOptions{}); err != nil {
					t.Fatal(err)
				}
				got, err = call(v.side, "total")
				if err != nil || got.Value.Bits != v.want {
					t.Fatal(got, err)
				}
			}
		})
	}
}

func TestHostComponentBindingFailures(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	m, err := e.Compile(nil, fixture(t, "host-components"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	otherEngine := resourceValuesEngine(t, "interpreter")
	foreign, err := otherEngine.Compile(nil, fixture(t, "host-component-guest"))
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.Close()
	closed, err := e.Compile(nil, fixture(t, "host-component-guest"))
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	wrong, err := e.Compile(nil, []byte("\x00asm\x0d\x00\x01\x00"))
	if err != nil {
		t.Fatal(err)
	}
	defer wrong.Close()
	for _, test := range []struct {
		code *Compiled
		want error
	}{{nil, vm.ErrLink}, {new(Compiled), vm.ErrReference}, {foreign, vm.ErrReference}, {closed, wasm.ErrClosed}, {wrong, vm.ErrType}} {
		if in, err := m.Instantiate(nil, Imports{Components: map[string]*Compiled{"code": test.code}}, CallOptions{}); in != nil || !errors.Is(err, test.want) {
			t.Fatal(in, err, test.want)
		}
	}
	// A component import now compiles, even when unused, and still requires an
	// actual binding before any guest initialization can occur.
	empty, err := e.Compile(nil, fixture(t, "component-import"))
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if in, err := empty.Instantiate(nil, Imports{}, CallOptions{}); in != nil || !errors.Is(err, vm.ErrLink) {
		t.Fatal(in, err)
	}
	new(Compiled).Close()
}

func TestHostComponentCoreImportsPreserveGCIdentities(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			ops, err := e.CompileCore(nil, fixture(t, "host-component-ops"))
			if err != nil {
				t.Fatal(err)
			}
			code, err := e.Compile(nil, fixture(t, "host-component-core"))
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, fixture(t, "host-components-core"))
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"ops": ops}, Components: map[string]*Compiled{"code": code}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			ops.Close()
			code.Close()
			m.Close()
			f, err := in.Function("add")
			if err != nil {
				t.Fatal(err)
			}
			got, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
			if err != nil || got.Value.Bits != 9 {
				t.Fatal(got, err)
			}
			if err := in.store.Collect(); err != nil {
				t.Fatal(err)
			}
			got, err = f.Call(nil, []Value{{Bits: 7}, {Bits: 5}}, CallOptions{})
			if err != nil || got.Value.Bits != 12 {
				t.Fatal(got, err)
			}
		})
	}
}

func TestHostComponentBindingLeaseCacheAndEarlyOwnerClose(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	code, err := e.Compile(nil, []byte("\x00asm\x0d\x00\x01\x00"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(nil, fixture(t, "host-components-cache"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	in, err := m.Instantiate(nil, Imports{Components: map[string]*Compiled{"first": code, "second": code}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if code.refs != 2 {
		t.Fatal("duplicate compilation leases", code.refs)
	}
	code.Close()
	in.Close()
	in.Close()
	if code.refs != 0 || code.plan != nil || code.cores != nil {
		t.Fatal("compilation not retired")
	}
	code, err = e.Compile(nil, fixture(t, "host-component-guest"))
	if err != nil {
		t.Fatal(err)
	}
	before, err := e.Compile(nil, fixture(t, "host-components-before"))
	if err != nil {
		t.Fatal(err)
	}
	effects := 0
	in, err = before.Instantiate(nil, Imports{Components: map[string]*Compiled{"code": code}, Functions: map[string]HostFunction{"before": func(*Caller, []Value) (*Value, error) { effects++; return nil, nil }}, FunctionChecks: map[string]FunctionCheck{"before": func(FunctionInfo) error { code.Close(); before.Close(); return nil }}}, CallOptions{})
	if err != nil {
		t.Fatal("in-flight binding did not retain code", err)
	}
	defer in.Close()
	if effects != 0 {
		t.Fatal("unused host import executed", effects)
	}
	f, err := in.Function("add")
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
	if err != nil || got.Value.Bits != 9 {
		t.Fatal(got, err)
	}
}

func TestHostComponentBindingRollbackAndLimits(t *testing.T) {
	for _, mode := range []string{"error", "panic", "cancel", "instances", "depth"} {
		t.Run(mode, func(t *testing.T) {
			cfg := DefaultConfig()
			if mode == "instances" {
				cfg.Store.Limits.MaxInstances = 1
			}
			if mode == "depth" {
				cfg.MaxNesting = 1
			}
			e, err := NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			code, err := e.Compile(nil, fixture(t, "host-component-guest"))
			if err != nil {
				t.Fatal(err)
			}
			defer code.Close()
			m, err := e.Compile(nil, fixture(t, "host-components-before"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			imports := Imports{Components: map[string]*Compiled{"code": code}, Functions: map[string]HostFunction{"before": func(*Caller, []Value) (*Value, error) { t.Error("unexpected effect"); return nil, nil }}, FunctionChecks: map[string]FunctionCheck{"before": func(FunctionInfo) error {
				if mode == "error" {
					return vm.ErrType
				}
				if mode == "panic" {
					panic("binding check")
				}
				if mode == "cancel" {
					cancel()
				}
				return nil
			}}}
			func() {
				defer func() {
					r := recover()
					if mode == "panic" && r != "binding check" {
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
			if code.refs != 1 || m.refs != 1 {
				t.Fatal("failed binding retained leases", code.refs, m.refs)
			}
		})
	}
}

func TestConcurrentHostComponentBindingsSurviveOwnerClose(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	code, err := e.Compile(nil, fixture(t, "host-component-guest"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(nil, fixture(t, "host-components-before"))
	if err != nil {
		t.Fatal(err)
	}
	const count = 6
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer code.Close()
	defer m.Close()
	ready := make(chan struct{}, count)
	release := make(chan struct{})
	done := make(chan error, count)
	imports := Imports{Components: map[string]*Compiled{"code": code}, Functions: map[string]HostFunction{"before": func(*Caller, []Value) (*Value, error) { return nil, nil }}, FunctionChecks: map[string]FunctionCheck{"before": func(FunctionInfo) error {
		select {
		case ready <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}}}
	for i := 0; i < count; i++ {
		go func() {
			err := func() error {
				in, err := m.Instantiate(ctx, imports, CallOptions{})
				if err != nil {
					return err
				}
				defer in.Close()
				var f *Function
				f, err = in.Function("twice")
				if err == nil {
					var got CallResult
					got, err = f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
					if err == nil && got.Value.Bits != 10 {
						err = vm.ErrType
					}
				}
				return err
			}()
			done <- err
		}()
	}
	for i := 0; i < count; i++ {
		select {
		case <-ready:
		case err := <-done:
			t.Fatal("binding failed before barrier", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	code.Close()
	m.Close()
	close(release)
	for i := 0; i < count; i++ {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}
	if code.refs != 0 || m.refs != 0 {
		t.Fatal("bindings did not retire", code.refs, m.refs)
	}
}

func TestImportedComponentKeepsExistingStoreAfterEngineClose(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	code, err := e.Compile(nil, fixture(t, "host-component-guest"))
	if err != nil {
		t.Fatal(err)
	}
	defer code.Close()
	m, err := e.Compile(nil, fixture(t, "host-components"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	in, err := m.Instantiate(nil, Imports{Components: map[string]*Compiled{"code": code}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	e.Close()
	f, _ := in.Function("add")
	got, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
	if err != nil || got.Value.Bits != 9 {
		t.Fatal(got, err)
	}
	if result, err := m.Instantiate(nil, Imports{Components: map[string]*Compiled{"code": code}}, CallOptions{}); result != nil || !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(result, err)
	}
	if m.refs != 2 {
		t.Fatal("failed Store allocation leaked a lease", m.refs)
	}
}

func TestImportedComponentValuesAndFailedConsumerCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			for _, fail := range []bool{false, true} {
				t.Run(fmt.Sprint(fail), func(t *testing.T) {
					code, err := e.Compile(nil, fixture(t, "host-component-observed"))
					if err != nil {
						t.Fatal(err)
					}
					defer code.Close()
					name := "host-components-values"
					if fail {
						name = "host-components-start-failure"
					}
					m, err := e.Compile(nil, fixture(t, name))
					if err != nil {
						t.Fatal(err)
					}
					defer m.Close()
					drops := uint64(0)
					imports := Imports{Components: map[string]*Compiled{"code": code}, Functions: map[string]HostFunction{
						"observe": func(c *Caller, args []Value) (*Value, error) { drops += args[0].Bits; return nil, nil },
						"fail":    func(*Caller, []Value) (*Value, error) { return nil, vm.ErrType },
					}}
					in, err := m.Instantiate(nil, imports, CallOptions{})
					if fail {
						if in != nil || !errors.Is(err, vm.ErrType) {
							t.Fatal(in, err)
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						defer in.Close()
						code.Close()
						m.Close()
						v, err := in.Value(nil, "values", "output")
						if err != nil {
							t.Fatal(err)
						}
						if _, err := in.DropResource(nil, v.Resource, CallOptions{}); err != nil {
							t.Fatal(err)
						}
					}
					if drops != 42 {
						t.Fatal("healthy imported destructor could not call its host provider", drops)
					}
				})
			}
		})
	}
}

func TestHostFunctionAliasesRetainProviderState(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			e := resourceValuesEngine(t, backend)
			m, err := e.Compile(nil, fixture(t, "host-component-host-alias"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{"host": func(c *Caller, args []Value) (*Value, error) { return &Value{Bits: args[0].Bits + 5}, nil }}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			trap, _ := in.Function("trap")
			if _, err := trap.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			host, _ := in.Function("host")
			got, err := host.Call(nil, []Value{{Bits: 4}}, CallOptions{})
			if err != nil || got.Value.Bits != 9 {
				t.Fatal("host alias inherited consumer failure", got, err)
			}
		})
	}
}

func TestImportedComponentsAcceptComponentArguments(t *testing.T) {
	e := resourceValuesEngine(t, "interpreter")
	inner, err := e.Compile(nil, fixture(t, "host-component-guest"))
	if err != nil {
		t.Fatal(err)
	}
	outer, err := e.Compile(nil, fixture(t, "host-component-delegate"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(nil, fixture(t, "host-components-delegate"))
	if err != nil {
		t.Fatal(err)
	}
	in, err := m.Instantiate(nil, Imports{Components: map[string]*Compiled{"inner": inner, "outer": outer}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	inner.Close()
	outer.Close()
	m.Close()
	f, _ := in.Function("add")
	got, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, CallOptions{})
	if err != nil || got.Value.Bits != 9 {
		t.Fatal(got, err)
	}
}
