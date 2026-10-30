package preview2

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type mutableIOError struct{ message string }

func (e *mutableIOError) Error() string { return e.message }

func TestIOErrorResourcesBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			cfg := component.DefaultConfig()
			cfg.Core.Backend = backend
			e, err := component.NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "io-errors"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			for _, message := range []string{"", "disk error", "界🙂\x00 detail", "\xffbroken", strings.Repeat("界", 20)} {
				provider, err := NewIOErrors(IOErrorConfig{MaxErrors: 2, MaxDebugBytes: 16})
				if err != nil {
					t.Fatal(err)
				}
				cause := &mutableIOError{message: message}
				var saved *component.Caller
				imports := provider.Imports()
				imports.Functions = map[string]component.HostFunction{"create": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
					saved = c
					value, err := provider.NewError(c, cause)
					if err != nil {
						return nil, err
					}
					info, err := provider.Describe(c, value.Resource)
					if err != nil {
						return nil, err
					}
					if info.Cause != cause {
						return nil, errors.New("host error cause changed")
					}
					return &value, nil
				}}
				in, err := m.Instantiate(nil, imports, component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				create, err := in.Function("create")
				if err != nil {
					t.Fatal(err)
				}
				value, err := create.Call(nil, nil, component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expected := message
				if len(expected) > 16 {
					expected = expected[:16]
				}
				expected = strings.ToValidUTF8(expected, "?")
				cause.message = "mutated after creation"
				debug, err := in.Function("debug")
				if err != nil {
					t.Fatal(err)
				}
				first, err := debug.Call(nil, []component.Value{*value.Value}, component.CallOptions{Metered: true, Fuel: 1000})
				if err != nil || first.Value.Text != expected || !utf8.ValidString(first.Value.Text) {
					t.Fatal(first, err, expected)
				}
				again, err := debug.Call(nil, []component.Value{*value.Value}, component.CallOptions{})
				if err != nil || again.Value.Text != first.Value.Text {
					t.Fatal(first, again, err)
				}
				if _, err := provider.Describe(saved, value.Value.Resource); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("expired Caller", err)
				}
				drop, err := in.Function("drop")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := drop.Call(nil, []component.Value{*value.Value}, component.CallOptions{}); err != nil {
					t.Fatal(err)
				}
				if _, err := debug.Call(nil, []component.Value{*value.Value}, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("stale error", err)
				}
				provider.mu.Lock()
				live := len(provider.entries)
				provider.mu.Unlock()
				if live != 0 {
					t.Fatal("error drop retained cause", live)
				}
				in.Close()
				provider.Close()
			}
		})
	}
}
