package component

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	bytes, err := os.ReadFile("testdata/" + name + ".wasm")
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestCompileAndInstantiationFailureBoundaries(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(c *Config) { c.MaxBytes = 0 },
		func(c *Config) { c.MaxValues = 0 },
		func(c *Config) { c.MaxNesting = 0 },
		func(c *Config) { c.MaxResources = 0 },
	} {
		c := DefaultConfig()
		mutate(&c)
		if e, err := NewEngine(c); e != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(e, err)
		}
	}
	config := DefaultConfig()
	config.Core.Backend = "missing"
	if _, err := NewEngine(config); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Compile(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := e.Compile(nil, nil); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal(err)
	}
	if _, err := e.Compile(nil, fixture(t, "invalid-index")); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}

	m, err := e.Compile(nil, fixture(t, "void"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Instantiate(nil, Imports{}, vm.CallOptions{}); !errors.Is(err, vm.ErrLink) {
		t.Fatal(err)
	}
	m.Close()
	if _, err := m.Instantiate(nil, Imports{}, vm.CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	e.Close()
	if _, err := e.Compile(nil, []byte("\x00asm\x0d\x00\x01\x00")); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	config = DefaultConfig()
	config.Core.Profile, config.Core.Features = "core2", 0
	limited, _ := NewEngine(config)
	defer limited.Close()
	if _, err := limited.Compile(nil, fixture(t, "strings64")); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal("Core 2 limits grammar accepted a Memory64 module", err)
	}
}

func TestRuntimeViewsAndOuterClosures(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	for _, name := range []string{"ascription", "outer"} {
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		path := []string{"add"}
		if name == "ascription" {
			path = []string{"library", "public"}
			if _, err := in.Function("library", "hidden"); !errors.Is(err, vm.ErrLink) {
				t.Fatal("ascription leaked hidden export", err)
			}
		}
		f, err := in.Function(path...)
		if err != nil {
			t.Fatal(err)
		}
		v, err := f.Call(nil, []Value{{Bits: 4}, {Bits: 5}}, vm.CallOptions{})
		if err != nil || v.Value.Bits != 9 {
			t.Fatal(v, err)
		}
		if _, err := in.Function(); !errors.Is(err, vm.ErrLink) {
			t.Fatal(err)
		}
		if _, err := in.Function(append(append([]string(nil), path...), "absent")...); !errors.Is(err, vm.ErrType) {
			t.Fatal(err)
		}
		in.Close()
		if _, err := in.Function(path...); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		m.Close()
	}
}

func TestInstanceExpansionAndHostReentryAreBounded(t *testing.T) {
	for _, change := range []func(*Config){
		func(c *Config) { c.Store.Limits.MaxInstances = 8 },
		func(c *Config) { c.MaxNesting = 2 },
		func(c *Config) { c.Validate.MaxNodes = 200 },
	} {
		config := DefaultConfig()
		change(&config)
		e, _ := NewEngine(config)
		m, err := e.Compile(nil, fixture(t, "expansion"))
		if err != nil {
			t.Fatal(err)
		}
		if result, err := m.Instantiate(nil, Imports{}, vm.CallOptions{}); result != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(result, err)
		}
		m.Close()
		e.Close()
	}
	config := DefaultConfig()
	config.Store.MaxReentry = 2
	e, _ := NewEngine(config)
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "void"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var run *Function
	in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{
		"host": func(c *Caller, _ []Value) (*Value, error) { return c.Call(run, nil) },
	}}, vm.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	run, _ = in.Function("run")
	if _, err := run.Call(nil, nil, vm.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func TestTrapsAndHostFailuresLockCanonicalState(t *testing.T) {
	sentinel := errors.New("host failure")
	for _, name := range []string{"strings", "post-trap", "core-start-trap"} {
		e, _ := NewEngine(DefaultConfig())
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{
			"echo": func(_ *Caller, args []Value) (*Value, error) {
				if name == "strings" {
					return nil, sentinel
				}
				return &args[0], nil
			},
		}}, vm.CallOptions{})
		if name == "core-start-trap" {
			if in != nil || !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(in, err)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			echo, _ := in.Function("echo")
			_, err = echo.Call(nil, []Value{{Text: "hello"}}, vm.CallOptions{})
			if name == "strings" && !errors.Is(err, sentinel) || name == "post-trap" && !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(name, err)
			}
			count, _ := in.Function("calls")
			if _, err := count.Call(nil, nil, vm.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("canonical trap did not lock the instance", err)
			}
			in.Close()
		}
		m.Close()
		e.Close()
	}
}

func TestFloatingPointCoreBoundaryAndCancellation(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	for _, kind := range []string{"f32", "f64"} {
		m, err := e.Compile(nil, fixture(t, kind))
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := m.Instantiate(ctx, Imports{}, vm.CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		f, _ := in.Function("add")
		a, b, want := uint64(math.Float32bits(1.5)), uint64(math.Float32bits(2.75)), uint64(math.Float32bits(4.25))
		if kind == "f64" {
			a, b, want = math.Float64bits(1.5), math.Float64bits(2.75), math.Float64bits(4.25)
		}
		result, err := f.Call(nil, []Value{{Bits: a}, {Bits: b}}, vm.CallOptions{})
		if err != nil || result.Value.Bits != want {
			t.Fatal(result, err)
		}
		if _, err := f.Call(ctx, nil, vm.CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		in.Close()
		m.Close()
	}
	var nilFunction *Function
	if _, err := nilFunction.Call(nil, nil, vm.CallOptions{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
}

func TestVoidCallsRejectUnexpectedHostResults(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "void"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, result := range []*Value{nil, {Bits: 1}} {
		in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{
			"host": func(*Caller, []Value) (*Value, error) { return result, nil },
		}}, vm.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		run, _ := in.Function("run")
		value, err := run.Call(nil, nil, vm.CallOptions{})
		if result == nil && (err != nil || value.Value != nil) || result != nil && !errors.Is(err, vm.ErrType) {
			t.Fatal(value, err)
		}
		in.Close()
	}
}
