package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestErrorContextHostAndComponentBoundaries(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Core.Backend = backend
			cfg.MaxResources = 4
			e, err := NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			guest, err := e.Compile(nil, fixture(t, "error-context"))
			if err != nil {
				t.Fatal(err)
			}
			defer guest.Close()
			consumer, err := e.Compile(nil, fixture(t, "error-context-components"))
			if err != nil {
				t.Fatal(err)
			}
			defer consumer.Close()
			in, err := consumer.Instantiate(nil, Imports{Components: map[string]*Compiled{"code": guest}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			makeContext, _ := in.Function("make")
			inspect, _ := in.Function("inspect")
			release, _ := in.Function("release")
			v, err := makeContext.Call(nil, []Value{{Text: "shared"}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := release.Call(nil, nil, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			for range 10 {
				out, err := inspect.Call(nil, []Value{*v.Value}, CallOptions{})
				if err != nil || out.Value.Text != "shared" {
					t.Fatal(out, err)
				}
			}
			in.Close()
			if msg, err := ErrorContextMessage(*v.Value); err != nil || msg != "shared" {
				t.Fatal(msg, err)
			}
			host, err := e.Compile(nil, fixture(t, "error-context-host"))
			if err != nil {
				t.Fatal(err)
			}
			defer host.Close()
			for _, mode := range []string{"copy", "error", "panic", "cancel"} {
				t.Run(mode, func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					imports := Imports{Functions: map[string]HostFunction{"host": func(_ *Caller, args []Value) (*Value, error) {
						switch mode {
						case "error":
							return nil, wasm.ErrTrap
						case "panic":
							panic("error-context host")
						case "cancel":
							cancel()
						}
						copy := args[0]
						return &copy, nil
					}}, FunctionChecks: map[string]FunctionCheck{"host": func(info FunctionInfo) error {
						_, param := info.Parameter(0)
						if param.Kind() != KindErrorContext || info.Result().Kind() != KindErrorContext {
							return wasm.ErrInvalid
						}
						return nil
					}}}
					instance, err := host.Instantiate(nil, imports, CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer instance.Close()
					function, _ := instance.Function("call")
					if mode == "panic" {
						func() {
							defer func() {
								if recover() == nil {
									t.Error("missing host panic")
								}
							}()
							function.Call(ctx, []Value{*v.Value}, CallOptions{})
						}()
					} else {
						out, err := function.Call(ctx, []Value{*v.Value}, CallOptions{})
						if mode == "copy" {
							if err != nil {
								t.Fatal(err)
							}
							message, err := ErrorContextMessage(*out.Value)
							if err != nil || message != "shared" {
								t.Fatal(message, err)
							}
							cleanup, _ := instance.Function("release")
							if _, err := cleanup.Call(nil, nil, CallOptions{}); err != nil {
								t.Fatal(err)
							}
						} else if mode == "cancel" {
							if !errors.Is(err, context.Canceled) {
								t.Fatal(err)
							}
						} else if !errors.Is(err, wasm.ErrTrap) {
							t.Fatal(err)
						}
					}
					// Failed transfers release their unpublished lowered handles. A
					// fresh namespace with the same Store budget can still allocate.
					probe := canonical.NewResourceTable(instance.resourceBudget)
					for range cfg.MaxResources {
						if _, err := probe.AddErrorContext(*v.Value); err != nil {
							t.Fatal("unpublished handles retained quota", err)
						}
					}
					probe.Close()
				})
			}
		})
	}
}

func TestErrorContextBudgetsDiscardAndGuards(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, discard := range []bool{false, true} {
			cfg := DefaultConfig()
			cfg.Core.Backend = backend
			cfg.MaxBytes = 24
			cfg.MaxResources = 4
			cfg.DiscardErrorContextMessages = discard
			e, err := NewEngine(cfg)
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, fixture(t, "error-context"))
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			newRaw, _ := in.Function("raw-new")
			drop, _ := in.Function("raw-drop")
			id, err := newRaw.Call(nil, []Value{{Bits: 0}, {Bits: 10}}, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !discard {
				if _, err := newRaw.Call(nil, []Value{{Bits: 0}, {Bits: 10}}, CallOptions{}); err != nil {
					t.Fatal(err)
				}
				if _, err := newRaw.Call(nil, []Value{{Bits: 0}, {Bits: 10}}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(err)
				}
			} else {
				if _, err := drop.Call(nil, []Value{*id.Value}, CallOptions{}); err != nil {
					t.Fatal(err)
				}
				// Discard policy deliberately does not read/validate the input range.
				id, err = newRaw.Call(nil, []Value{{Bits: 0xffffffff}, {Bits: 0xffffffff}}, CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
			}
			in.Close()
			m.Close()
			e.Close()
		}
		cfg := DefaultConfig()
		cfg.Core.Backend = backend
		e, _ := NewEngine(cfg)
		m, err := e.Compile(nil, fixture(t, "error-context-post-return"))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		function, _ := in.Function("make")
		if _, err := function.Call(nil, []Value{{Text: "message"}}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("post-return illegally left the instance", err)
		}
		in.Close()
		m.Close()
		e.Close()
	}
}

func TestErrorContextStandaloneValuesAndCrossNamespace(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		cfg := DefaultConfig()
		cfg.Core.Backend = backend
		cfg.MaxResources = 8
		e, err := NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, fixture(t, "error-context-cross"))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, Imports{}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		run, _ := in.Function("run")
		out, err := run.Call(nil, []Value{{Bits: 5}}, CallOptions{})
		if err != nil || out.Value.Bits != 5 {
			t.Fatal(out, err)
		}
		in.Close()
		m.Close()
		m, err = e.Compile(nil, fixture(t, "error-context-values"))
		if err != nil {
			t.Fatal(err)
		}
		first, _ := NewErrorContext("same")
		copyRef := *first.Resource
		equal := Value{Resource: &copyRef}
		different, _ := NewErrorContext("same")
		for _, second := range []Value{equal, different} {
			in, err := m.Instantiate(nil, Imports{Values: map[string]Value{"first": first, "same": second}}, CallOptions{})
			if second.Resource == different.Resource {
				if in != nil || err == nil {
					t.Fatal("different contexts satisfied an equality bound", err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			for range 3 {
				value, err := in.Value(nil, "first")
				if err != nil {
					t.Fatal(err)
				}
				message, err := ErrorContextMessage(value)
				if err != nil || message != "same" {
					t.Fatal(message, err)
				}
			}
			in.Close()
		}
		in, err = m.Instantiate(nil, Imports{ValueFactories: map[string]ValueFactory{"first": func(*Caller) (Value, error) { return first, nil }, "same": func(*Caller) (Value, error) { return equal, nil }}}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		in.Close()
		m.Close()
		e.Close()
	}
}
