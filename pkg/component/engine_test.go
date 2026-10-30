package component

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestOwnCoreComponentComposition(t *testing.T) {
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
			for _, fixture := range []string{"../../internal/component/validate/testdata/canonical-math.wasm", "../../internal/component/decode/testdata/composition.wasm"} {
				input, err := os.ReadFile(fixture)
				if err != nil {
					t.Fatal(err)
				}
				compiled, err := e.Compile(nil, input)
				if err != nil {
					t.Fatal(err)
				}
				in, err := compiled.Instantiate(nil, Imports{}, vm.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				f, err := in.Function("add")
				if err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 10; i++ {
					result, err := f.Call(nil, []Value{{Bits: uint64(i)}, {Bits: 5}}, vm.CallOptions{Metered: true, Fuel: 3})
					if err != nil || result.Value.Bits != uint64(i+5) || result.FuelRemaining != 0 {
						t.Fatal(result, err)
					}
				}
				if _, err := f.Call(nil, []Value{{Bits: 1}}, vm.CallOptions{}); !errors.Is(err, vm.ErrType) {
					t.Fatal(err)
				}
				compiled.Close()
				if _, err := f.Call(nil, []Value{{Bits: 1}, {Bits: 2}}, vm.CallOptions{}); err != nil {
					t.Fatal("compiled close invalidated an existing instance", err)
				}
				in.Close()
				if _, err := f.Call(nil, nil, vm.CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal(err)
				}
				if err := in.WaitClosed(nil); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestCanonicalStringHostCallAndPostReturn(t *testing.T) {
	for _, fixture := range []string{"strings", "strings-utf16", "strings-latin1", "strings64"} {
		for _, backend := range []string{"interpreter", "native"} {
			if backend == "native" && !native.Available {
				continue
			}
			t.Run(backend+"/"+fixture, func(t *testing.T) {
				config := DefaultConfig()
				config.Core.Backend = backend
				e, err := NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				defer e.Close()
				bytes, err := os.ReadFile("testdata/" + fixture + ".wasm")
				if err != nil {
					t.Fatal(err)
				}
				compiled, err := e.Compile(nil, bytes)
				if err != nil {
					t.Fatal(err)
				}
				defer compiled.Close()
				var escaped *Caller
				in, err := compiled.Instantiate(nil, Imports{Functions: map[string]HostFunction{
					"echo": func(c *Caller, args []Value) (*Value, error) {
						escaped = c
						return &Value{Text: strings.ToUpper(args[0].Text)}, nil
					},
				}}, vm.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				echo, _ := in.Function("echo")
				calls, _ := in.Function("calls")
				for i, text := range []string{"hello", "界🙂", ""} {
					result, err := echo.Call(nil, []Value{{Text: text}}, vm.CallOptions{})
					if err != nil || result.Value.Text != strings.ToUpper(text) {
						t.Fatal(result, err)
					}
					count, err := calls.Call(nil, nil, vm.CallOptions{})
					if err != nil || count.Value.Bits != uint64(i+1) {
						t.Fatal("post-return did not execute", count, err)
					}
					if _, err := escaped.Call(calls, nil); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal("escaped host capability", err)
					}
				}
				if _, err := echo.Call(nil, []Value{{Text: string([]byte{255})}}, vm.CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
					t.Fatal(err)
				}
				if _, err := echo.Call(nil, []Value{{Text: "still usable"}}, vm.CallOptions{}); err != nil {
					t.Fatal("invalid host input poisoned the instance", err)
				}
			})
		}
	}
}

func TestComponentStartActuallyExecutesAndValuesAreOwned(t *testing.T) {
	e, err := NewEngine(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	bytes, err := os.ReadFile("../../internal/component/validate/testdata/value-start.wasm")
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := e.Compile(nil, bytes)
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	called := 0
	in, err := compiled.Instantiate(nil, Imports{Functions: map[string]HostFunction{
		"convert": func(c *Caller, args []Value) (*Value, error) {
			if err := c.Context().Err(); err != nil {
				return nil, err
			}
			called++
			return &Value{Bits: args[0].Bits * 2}, nil
		},
	}}, vm.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	answer, err := in.Value(nil, "answer")
	if err != nil || answer.Bits != 84 || called != 1 {
		t.Fatal(answer, called, err)
	}
	answer.Bits = 0
	if again, err := in.Value(nil, "answer"); err != nil || again.Bits != 84 {
		t.Fatal("exported value aliases instance state", again, err)
	}
	if _, err := in.Value(nil, "missing"); !errors.Is(err, vm.ErrLink) {
		t.Fatal(err)
	}
}

func TestCanonicalParameterAndResultSpilling(t *testing.T) {
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
			bytes, err := os.ReadFile("testdata/record-spill.wasm")
			if err != nil {
				t.Fatal(err)
			}
			compiled, err := e.Compile(nil, bytes)
			if err != nil {
				t.Fatal(err)
			}
			defer compiled.Close()
			if _, err := compiled.Instantiate(nil, Imports{}, vm.CallOptions{}); !errors.Is(err, vm.ErrLink) {
				t.Fatal("missing host interface accepted", err)
			}
			if _, err := compiled.Instantiate(nil, Imports{Instances: map[string]Imports{"host": {}}}, vm.CallOptions{}); !errors.Is(err, vm.ErrLink) {
				t.Fatal("missing nested host function accepted", err)
			}
			in, err := compiled.Instantiate(nil, Imports{Instances: map[string]Imports{
				"host": {Functions: map[string]HostFunction{"echo": func(_ *Caller, args []Value) (*Value, error) {
					var sum uint64
					for _, field := range args[0].Elements {
						sum += field.Bits
					}
					return &Value{Elements: []Value{{Bits: sum}, {Bits: 17}, {Bits: 42}}}, nil
				}}},
			}}, vm.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			fn, _ := in.Function("echo")
			args := Value{Elements: make([]Value, 17)}
			for i := range args.Elements {
				args.Elements[i].Bits = uint64(i)
			}
			result, err := fn.Call(nil, []Value{args}, vm.CallOptions{})
			if err != nil || result.Value.Elements[0].Bits != 136 || result.Value.Elements[1].Bits != 17 || result.Value.Elements[2].Bits != 42 {
				t.Fatal(result, err)
			}
		})
	}
}
