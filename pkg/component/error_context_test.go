package component

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestErrorContextExecution(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, name := range []string{"error-context", "error-context-utf16", "error-context-compact", "error-context64"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				config := DefaultConfig()
				config.Core.Backend = backend
				config.MaxResources = 4
				e, err := NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				defer e.Close()
				m, err := e.Compile(nil, fixture(t, name))
				if err != nil {
					t.Fatal(err)
				}
				defer m.Close()
				in, err := m.Instantiate(nil, Imports{}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				makeContext, _ := in.Function("make")
				echo, _ := in.Function("echo")
				inspect, _ := in.Function("inspect")
				record, _ := in.Function("record")
				for range 10 {
					result, err := makeContext.Call(nil, []Value{{Text: "error λ 😃"}}, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					value := *result.Value
					if msg, err := ErrorContextMessage(value); err != nil || msg != "error λ 😃" {
						t.Fatal(msg, err)
					}
					result, err = echo.Call(nil, []Value{value}, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					result, err = inspect.Call(nil, []Value{*result.Value}, CallOptions{})
					if err != nil || result.Value.Text != "error λ 😃" {
						t.Fatal(result, err)
					}
					result, err = record.Call(nil, []Value{{Elements: []Value{value, {Bits: 42}}}}, CallOptions{})
					if err != nil || result.Value.Elements[1].Bits != 42 {
						t.Fatal(result, err)
					}
				}
				churn, _ := in.Function("churn")
				if result, err := churn.Call(nil, []Value{{Bits: 1000}}, CallOptions{}); err != nil || result.Value.Bits != 1000 {
					t.Fatal(result, err)
				}
				bad, _ := NewErrorContext("bad")
				if _, err := inspect.Call(nil, []Value{{}}, CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
					t.Fatal(err)
				}
				if result, err := inspect.Call(nil, []Value{bad}, CallOptions{}); err != nil || result.Value.Text != "bad" {
					t.Fatal(result, err)
				}
			})
		}
	}
}
