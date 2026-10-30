package component

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostCoreModuleDomainsTypesAndPublication(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.CompileCore(nil, fixture(t, "host-core-math"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			info, err := m.Info()
			if err != nil || info.Backend != backend || len(info.Exports) != 1 {
				t.Fatal(info, err)
			}
			info.Exports[0].Function.Params[0] = wasm.F64
			if again, err := m.Info(); err != nil || again.Exports[0].Function.Params[0] != wasm.I32 {
				t.Fatal("metadata mutated publication", again, err)
			}
			component, err := e.Compile(nil, fixture(t, "host-module"))
			if err != nil {
				t.Fatal(err)
			}
			defer component.Close()
			starts := 0
			imports := Imports{CoreModules: map[string]*CoreModule{"math": m}, Functions: map[string]HostFunction{"mark": func(*Caller, []Value) (*Value, error) { starts++; m.Close(); return nil, nil }}}
			in, err := component.Instantiate(nil, imports, CallOptions{})
			if err != nil || starts != 1 {
				t.Fatal(in, err, starts)
			}
			defer in.Close()
			if _, err := m.Info(); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
			if again, err := component.Instantiate(nil, imports, CallOptions{}); again != nil || !errors.Is(err, wasm.ErrClosed) || starts != 1 {
				t.Fatal(again, err, starts)
			}
			component.Close()
			e.Close()
			runtime.GC()
			f, err := in.Function("add")
			if err != nil {
				t.Fatal(err)
			}
			if v, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 9}}, CallOptions{Metered: true, Fuel: 100}); err != nil || v.Value.Bits != 13 {
				t.Fatal(v, err)
			}
			in.Close()
			in.Close()
			m.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := in.WaitClosed(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := f.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
			m.mu.Lock()
			defer m.mu.Unlock()
			if m.refs != 0 || m.compiled != nil {
				t.Fatal("module binding retained code", m.refs)
			}
		})
	}
}

func TestHostCoreModuleMatchingAndFailedInstantiation(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "host-module-rich"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			for _, name := range []string{"rich", "wrong-memory32", "wrong-memory-small", "wrong-memory-max", "wrong-table", "wrong-global", "wrong-tag", "wrong-function", "wrong-missing", "wrong-import"} {
				actual, err := e.CompileCore(nil, fixture(t, "host-core-"+name))
				if err != nil {
					t.Fatal(name, err)
				}
				in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"module": actual}}, CallOptions{})
				if name == "rich" {
					if err != nil {
						t.Fatal(err)
					}
					f, err := in.Function("apply")
					if err != nil {
						t.Fatal(err)
					}
					if v, err := f.Call(nil, []Value{{Bits: 41}}, CallOptions{}); err != nil || v.Value.Bits != 42 {
						t.Fatal(v, err)
					}
					in.Close()
				} else if in != nil || !errors.Is(err, vm.ErrType) {
					t.Fatal(name, in, err)
				}
				if _, err := actual.Info(); err != nil {
					t.Fatal("failed bind closed host owner", name, err)
				}
				actual.Close()
			}
			if in, err := m.Instantiate(nil, Imports{}, CallOptions{}); in != nil || !errors.Is(err, vm.ErrLink) {
				t.Fatal(in, err)
			}
			foreign, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer foreign.Close()
			actual, err := foreign.CompileCore(nil, fixture(t, "host-core-rich"))
			if err != nil {
				t.Fatal(err)
			}
			defer actual.Close()
			if in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"module": actual}}, CallOptions{}); in != nil || !errors.Is(err, vm.ErrReference) {
				t.Fatal(in, err)
			}
			if in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"module": {}}}, CallOptions{}); in != nil || !errors.Is(err, vm.ErrReference) {
				t.Fatal(in, err)
			}
			start, err := e.Compile(nil, fixture(t, "host-module"))
			if err != nil {
				t.Fatal(err)
			}
			defer start.Close()
			wrong, err := e.CompileCore(nil, fixture(t, "host-core-rich"))
			if err != nil {
				t.Fatal(err)
			}
			marks := 0
			if in, err := start.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"math": wrong}, Functions: map[string]HostFunction{"mark": func(*Caller, []Value) (*Value, error) { marks++; return nil, nil }}}, CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) || marks != 0 {
				t.Fatal("bad type executed an earlier Core start", in, err, marks)
			}
			wrong.Close()
			trap, err := e.CompileCore(nil, fixture(t, "host-core-start-trap"))
			if err != nil {
				t.Fatal(err)
			}
			defer trap.Close()
			if in, err := start.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"math": trap}, Functions: map[string]HostFunction{"mark": func(*Caller, []Value) (*Value, error) { return nil, nil }}}, CallOptions{}); in != nil || !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(in, err)
			}
			trap.mu.Lock()
			if trap.refs != 1 {
				t.Fatal("failed start retained module lease", trap.refs)
			}
			trap.mu.Unlock()
		})
	}
}

func TestHostCoreModuleCloseAndCancellationDuringBinding(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "host-module"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			for _, canceled := range []bool{false, true} {
				actual, err := e.CompileCore(nil, fixture(t, "host-core-math"))
				if err != nil {
					t.Fatal(err)
				}
				bound, resume := make(chan struct{}), make(chan struct{})
				starts := 0
				imports := Imports{CoreModules: map[string]*CoreModule{"math": actual}, Functions: map[string]HostFunction{"mark": func(*Caller, []Value) (*Value, error) { starts++; return nil, nil }}, FunctionChecks: map[string]FunctionCheck{"mark": func(FunctionInfo) error { close(bound); <-resume; return nil }}}
				ctx, cancel := context.WithCancel(context.Background())
				type outcome struct {
					in  *Instance
					err error
				}
				done := make(chan outcome, 1)
				go func() { in, err := m.Instantiate(ctx, imports, CallOptions{}); done <- outcome{in, err} }()
				select {
				case <-bound:
				case <-time.After(time.Second):
					t.Fatal("binding was not reached")
				}
				if canceled {
					cancel()
				} else {
					actual.Close()
				}
				runtime.GC()
				close(resume)
				var result outcome
				select {
				case result = <-done:
				case <-time.After(time.Second):
					t.Fatal("binding did not leave")
				}
				cancel()
				if canceled {
					if result.in != nil || !errors.Is(result.err, context.Canceled) || starts != 0 {
						t.Fatal(result, starts)
					}
					if _, err := actual.Info(); err != nil {
						t.Fatal("cancellation retired the owner's module", err)
					}
				} else {
					if result.err != nil || starts != 1 {
						t.Fatal(result, starts)
					}
					f, err := result.in.Function("add")
					if err != nil {
						t.Fatal(err)
					}
					if v, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 9}}, CallOptions{}); err != nil || v.Value.Bits != 13 {
						t.Fatal(v, err)
					}
					result.in.Close()
				}
				actual.Close()
				actual.mu.Lock()
				if actual.refs != 0 {
					t.Fatal("binding lease leaked", actual.refs)
				}
				actual.mu.Unlock()
			}
		})
	}
}

func TestHostCoreModuleLifecycleAndBindingBudget(t *testing.T) {
	for _, m := range []*CoreModule{nil, {}} {
		m.Close()
		if _, err := m.Info(); !errors.Is(err, vm.ErrReference) {
			t.Fatal(err)
		}
	}
	config := DefaultConfig()
	config.Validate.MaxNodes = 128
	e, err := NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if m, err := e.CompileCore(ctx, fixture(t, "host-core-math")); m != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(m, err)
	}
	if m, err := e.CompileCore(nil, nil); m != nil || !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal(m, err)
	}
	m, err := e.Compile(nil, fixture(t, "host-module"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	actual, err := e.CompileCore(nil, fixture(t, "host-core-many"))
	if err != nil {
		t.Fatal(err)
	}
	defer actual.Close()
	starts := 0
	if in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"math": actual}, Functions: map[string]HostFunction{"mark": func(*Caller, []Value) (*Value, error) { starts++; return nil, nil }}}, CallOptions{}); in != nil || !errors.Is(err, wasm.ErrLimit) || starts != 0 {
		t.Fatal(in, err, starts)
	}
	actual.mu.Lock()
	if actual.refs != 1 {
		t.Fatal(actual.refs)
	}
	actual.mu.Unlock()
	e.Close()
	if m, err := e.CompileCore(nil, fixture(t, "host-core-math")); m != nil || !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(m, err)
	}
}

func TestHostCoreModuleGCTypeIdentityAndCovariance(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "host-module-gc"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			for _, name := range []string{"host-core-gc", "host-core-wrong-gc"} {
				actual, err := e.CompileCore(nil, fixture(t, name))
				if err != nil {
					t.Fatal(name, err)
				}
				in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"module": actual}}, CallOptions{})
				if name == "host-core-wrong-gc" {
					if in != nil || !errors.Is(err, vm.ErrType) {
						t.Fatal(in, err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					actual.Close()
					runtime.GC()
					f, err := in.Function("get")
					if err != nil {
						t.Fatal(err)
					}
					if v, err := f.Call(nil, nil, CallOptions{Metered: true, Fuel: 100}); err != nil || v.Value.Bits != 42 {
						t.Fatal(v, err)
					}
					in.Close()
				}
				actual.Close()
			}
		})
	}
}

func TestHostCoreModuleConcurrentStores(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "host-module-gc"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			host, err := e.CompileCore(nil, fixture(t, "host-core-gc"))
			if err != nil {
				t.Fatal(err)
			}
			defer host.Close()
			const count = 8
			ready, done := make(chan error, count), make(chan error, count)
			resume := make(chan struct{})
			for i := 0; i < count; i++ {
				go func() {
					in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"module": host}}, CallOptions{})
					ready <- err
					if err != nil {
						done <- err
						return
					}
					defer in.Close()
					<-resume
					f, err := in.Function("get")
					if err == nil {
						var result CallResult
						result, err = f.Call(nil, nil, CallOptions{Metered: true, Fuel: 100})
						if err == nil && result.Value.Bits != 42 {
							err = errors.New("incorrect result from retained shared module")
						}
					}
					in.Close()
					done <- err
				}()
			}
			for i := 0; i < count; i++ {
				if err := <-ready; err != nil {
					t.Error(err)
				}
			}
			host.Close()
			m.Close()
			e.Close()
			runtime.GC()
			close(resume)
			for i := 0; i < count; i++ {
				if err := <-done; err != nil {
					t.Error(err)
				}
			}
			host.mu.Lock()
			defer host.mu.Unlock()
			if host.refs != 0 || host.compiled != nil {
				t.Fatal("concurrent Stores retained module leases", host.refs)
			}
		})
	}
}

func TestHostCoreModuleNestedBindingsAndImportVariance(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			for _, componentName := range []string{"host-module-nested", "host-module-imports"} {
				m, err := e.Compile(nil, fixture(t, componentName))
				if err != nil {
					t.Fatal(err)
				}
				for _, coreName := range []string{"host-core-math", "host-core-imports", "host-core-import-wrong-type"} {
					actual, err := e.CompileCore(nil, fixture(t, coreName))
					if err != nil {
						t.Fatal(err)
					}
					imports := Imports{CoreModules: map[string]*CoreModule{"math": actual}}
					if componentName == "host-module-nested" {
						imports = Imports{Instances: map[string]Imports{"host": imports}}
					}
					in, err := m.Instantiate(nil, imports, CallOptions{})
					valid := coreName == "host-core-math" || componentName == "host-module-imports" && coreName == "host-core-imports"
					if valid {
						if err != nil {
							t.Fatal(componentName, coreName, err)
						}
						f, err := in.Function("add")
						if err != nil {
							t.Fatal(err)
						}
						want := uint64(13)
						if coreName == "host-core-imports" {
							want += 10
						}
						if v, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 9}}, CallOptions{}); err != nil || v.Value.Bits != want {
							t.Fatal(v, err)
						}
						in.Close()
					} else if in != nil || !errors.Is(err, vm.ErrType) {
						t.Fatal(componentName, coreName, in, err)
					}
					actual.Close()
				}
				m.Close()
			}
		})
	}
}
