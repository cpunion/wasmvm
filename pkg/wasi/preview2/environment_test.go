package preview2

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name + ".wasm")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestEnvironmentExecutesThroughBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			e, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "environment"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			cwd := "/sandbox"
			cfg := EnvironmentConfig{Args: []string{"guest", "", "界🙂"}, Env: []string{"A=1=2", "A=second", "EMPTY="}, InitialCWD: &cwd}
			environment, err := NewEnvironment(cfg)
			if err != nil {
				t.Fatal(err)
			}
			cfg.Args[0] = "mutated"
			cfg.Env[0] = "BAD=changed"
			cwd = "/changed"
			in, err := m.Instantiate(nil, environment.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			call := func(name string) component.Value {
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				r, err := f.Call(nil, nil, component.CallOptions{})
				if err != nil {
					t.Fatal(name, err)
				}
				return *r.Value
			}
			for i := 0; i < 10; i++ {
				args := call("arguments")
				got := []string{}
				for _, v := range args.Elements {
					got = append(got, v.Text)
				}
				if !reflect.DeepEqual(got, []string{"guest", "", "界🙂"}) {
					t.Fatal(got)
				}
				args.Elements[0].Text = "changed result"
				vars := call("variables")
				pairs := [][2]string{}
				for _, v := range vars.Elements {
					pairs = append(pairs, [2]string{v.Elements[0].Text, v.Elements[1].Text})
				}
				if !reflect.DeepEqual(pairs, [][2]string{{"A", "1=2"}, {"A", "second"}, {"EMPTY", ""}}) {
					t.Fatal(pairs)
				}
				gotCWD := call("cwd")
				if gotCWD.Case != 1 || gotCWD.Elements[0].Text != "/sandbox" {
					t.Fatal(gotCWD)
				}
				if call("argument-count").Bits != 3 || call("environment-count").Bits != 3 {
					t.Fatal("guest memory did not receive the lists")
				}
			}
			f, _ := in.Function("argument-count")
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := f.Call(ctx, nil, component.CallOptions{}); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			if call("argument-count").Bits != 3 {
				t.Fatal("cancellation changed immutable configuration")
			}
			empty, _ := NewEnvironment(EnvironmentConfig{})
			other, err := m.Instantiate(nil, empty.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer other.Close()
			for _, name := range []string{"arguments", "variables", "cwd"} {
				f, _ := other.Function(name)
				v, err := f.Call(nil, nil, component.CallOptions{})
				if err != nil || len(v.Value.Elements) != 0 || v.Value.Case != 0 {
					t.Fatal("implicitly inherited host configuration", v, err)
				}
			}
		})
	}
}

func TestEnvironmentRejectsInvalidConfiguration(t *testing.T) {
	badCWD := string([]byte{255})
	for _, cfg := range []EnvironmentConfig{
		{Args: []string{"x\x00y"}}, {Args: []string{string([]byte{255})}},
		{Env: []string{"MISSING"}}, {Env: []string{"=value"}}, {Env: []string{"A=bad\x00"}},
		{InitialCWD: &badCWD},
	} {
		if e, err := NewEnvironment(cfg); e != nil || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(cfg, e, err)
		}
	}
	for _, cfg := range []EnvironmentConfig{{MaxBytes: 2, Args: []string{"abc"}}, {MaxBytes: 2, Env: []string{"A=1"}}, {MaxEntries: 1, Args: []string{"a"}, Env: []string{"A=b"}}} {
		if e, err := NewEnvironment(cfg); e != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(e, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	e, _ := NewEnvironment(EnvironmentConfig{Args: []string{"x"}, Env: []string{"X=y"}})
	if _, err := e.arguments(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := e.variables(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestEnvironmentChecksSignaturesBeforeEffects(t *testing.T) {
	e, _ := component.NewEngine(component.DefaultConfig())
	defer e.Close()
	env, _ := NewEnvironment(EnvironmentConfig{})
	for _, name := range []string{"wrong-arguments", "wrong-result"} {
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(err)
		}
		if in, err := m.Instantiate(nil, env.Imports(), component.CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal("invalid standard interface linked", name, in, err)
		}
		m.Close()
	}
}
