package component

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHostSources(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, name := range []string{"host-source-future-u32", "host-source-stream-u32", "host-source-stream-u8", "host-source-future-string", "host-source-stream-unit"} {
				for _, mode := range []uint64{0, 1, 3, 4} {
					var typ ValueType
					reads, closes := 0, 0
					var original *Caller
					var workContext context.Context
					workers := make(chan struct{}, 1)
					source := AsyncSource{Close: func() { closes++ }, Read: func(c *Caller, limit uint32) (*ReadCompletion, error) {
						reads++
						original = c
						workContext = c.Context()
						if limit != 1 {
							return nil, vm.ErrType
						}
						completion := NewReadCompletion()
						if mode == 3 || mode == 4 {
							go func() { defer func() { workers <- struct{}{} }(); <-workContext.Done(); completion.Cancelled() }()
						} else {
							go func() {
								defer func() { workers <- struct{}{} }()
								completion.Resolve(func(fresh *Caller) (ReadResult, error) {
									if fresh == original {
										return ReadResult{}, vm.ErrType
									}
									if _, err := original.Call(nil, nil); !errors.Is(err, wasm.ErrClosed) {
										return ReadResult{}, vm.ErrType
									}
									result := ReadResult{Values: []Value{{Bits: 42}}, Closed: true}
									if strings.HasSuffix(name, "-string") {
										result.Values = []Value{{Text: "雪😀"}}
									}
									if strings.HasSuffix(name, "-u8") {
										result.Values = nil
										result.Bytes = "*"
									}
									if strings.HasSuffix(name, "-unit") {
										result.Values = nil
										result.Count = 1
									}
									return result, nil
								})
							}()
						}
						return completion, nil
					}}
					imports := Imports{FunctionChecks: map[string]FunctionCheck{"make": func(info FunctionInfo) error { typ = info.Result(); return nil }}, Functions: map[string]HostFunction{
						"make":    func(c *Caller, _ []Value) (*Value, error) { v, err := c.NewSource(typ, source); return &v, err },
						"started": func(*Caller, []Value) (*Value, error) { return &Value{Bits: uint64(reads)}, nil },
					}}
					in := asyncInstance(t, backend, name, imports)
					run, _ := in.Function("run")
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					v, err := run.Call(ctx, []Value{{Bits: mode}}, CallOptions{})
					cancel()
					if mode == 4 {
						if !errors.Is(err, wasm.ErrTrap) {
							t.Fatal(name, mode, err)
						}
					} else if err != nil || v.Value == nil {
						t.Fatal(name, mode, v, err)
					} else if mode == 0 {
						if strings.HasSuffix(name, "-string") {
							if v.Value.Text != "雪😀" {
								t.Fatal(v)
							}
						} else {
							want := uint64(42)
							if strings.HasSuffix(name, "-unit") {
								want = 17
							}
							if v.Value.Bits != want {
								t.Fatal(name, v)
							}
						}
					}
					wantReads := 1
					if mode == 1 {
						wantReads = 0
					}
					if reads != wantReads || closes != 1 {
						t.Fatal(name, mode, "source lifecycle", reads, closes)
					}
					if wantReads != 0 {
						select {
						case <-workers:
						case <-time.After(time.Second):
							t.Fatal("external worker remained blocked")
						}
						if workContext.Err() == nil {
							t.Fatal("completed work context retained")
						}
					}
					if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
						t.Fatal("source retained task or payload")
					}
					probe := canonical.NewResourceTable(in.resourceBudget)
					for range in.config.MaxResources {
						if _, err := probe.AddOpaque(1); err != nil {
							t.Fatal(name, mode, "source retained endpoint", err)
						}
					}
					probe.Close()
				}
			}
		})
	}
}

func TestHostSourceFailures(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, mode := range []string{"nil", "zero", "start-error", "start-panic", "result-error", "result-panic", "count", "empty", "invalid", "close-panic", "pending"} {
				var typ ValueType
				var closed atomic.Uint32
				source := AsyncSource{Close: func() {
					closed.Add(1)
					if mode == "close-panic" {
						panic(mode)
					}
				}, Read: func(*Caller, uint32) (*ReadCompletion, error) {
					switch mode {
					case "nil":
						return nil, nil
					case "zero":
						return new(ReadCompletion), nil
					case "start-error":
						return nil, vm.ErrLink
					case "start-panic":
						panic(mode)
					}
					p := NewReadCompletion()
					if mode == "pending" {
						return p, nil
					}
					p.Resolve(func(*Caller) (ReadResult, error) {
						if mode == "result-error" {
							return ReadResult{}, vm.ErrLink
						}
						if mode == "result-panic" {
							panic(mode)
						}
						r := ReadResult{Values: []Value{{Bits: 42}}, Closed: true}
						switch mode {
						case "count":
							r.Count = 2
						case "empty":
							r.Values = nil
						case "invalid":
							r.Values[0].Bits = 1 << 32
						}
						return r, nil
					})
					return p, nil
				}}
				in := asyncInstance(t, backend, "host-source-future-u32", Imports{FunctionChecks: map[string]FunctionCheck{"make": func(info FunctionInfo) error { typ = info.Result(); return nil }}, Functions: map[string]HostFunction{
					"make": func(c *Caller, _ []Value) (*Value, error) { v, err := c.NewSource(typ, source); return &v, err }, "started": func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
				}})
				run, _ := in.Function("run")
				invoke := func() (err error) {
					defer func() {
						if recover() != nil {
							err = wasm.ErrTrap
						}
					}()
					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
					defer cancel()
					_, err = run.Call(ctx, []Value{{}}, CallOptions{})
					return err
				}
				if err := invoke(); err == nil {
					t.Fatal(mode, "failed source accepted")
				}
				if closed.Load() != 1 || in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal(mode, "failed source retained state", closed.Load())
				}
				probe := canonical.NewResourceTable(in.resourceBudget)
				for range in.config.MaxResources {
					if _, err := probe.AddOpaque(1); err != nil {
						t.Fatal(mode, "endpoint quota leaked", err)
					}
				}
				probe.Close()
			}
		})
	}
}

func TestHostSourceOwnedPayloadTransactions(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"guest", "host", "invalid", "close-panic"} {
			drops := uint32(0)
			rt := NewHostResourceType(func(*Caller, uint32) error { drops += 42; return nil })
			var typ ValueType
			source := AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
				p := NewReadCompletion()
				p.Resolve(func(c *Caller) (ReadResult, error) {
					value, err := c.NewResource(rt, 42)
					if err != nil {
						return ReadResult{}, err
					}
					r := ReadResult{Values: []Value{value}, Closed: true}
					if mode == "invalid" {
						r.Count = 2
					}
					return r, nil
				})
				return p, nil
			}, Close: func() {
				if mode == "close-panic" {
					panic("close")
				}
			}}
			in := asyncInstance(t, backend, "host-source-future-owned", Imports{Resources: map[string]*HostResourceType{"r": rt}, FunctionChecks: map[string]FunctionCheck{"make": func(info FunctionInfo) error { typ = info.Result(); return nil }}, Functions: map[string]HostFunction{
				"make": func(c *Caller, _ []Value) (*Value, error) { value, err := c.NewSource(typ, source); return &value, err }, "started": func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
			}})
			if mode == "host" {
				value, err := in.NewSource(nil, typ, source)
				if err != nil {
					t.Fatal(err)
				}
				r, err := in.ReadAsync(nil, value.Resource, 1, CallOptions{})
				if err != nil || len(r.Values) != 1 || drops != 0 {
					t.Fatal(r, drops, err)
				}
				if _, err := in.DropResource(nil, r.Values[0].Resource, CallOptions{}); err != nil {
					t.Fatal(err)
				}
				in.DropAsync(nil, value.Resource, CallOptions{})
			} else {
				run, _ := in.Function("run")
				invoke := func() (err error) {
					defer func() {
						if p := recover(); p != nil {
							if p != "close" {
								t.Fatal(p)
							}
							err = wasm.ErrTrap
						}
					}()
					_, err = run.Call(nil, []Value{{}}, CallOptions{})
					return err
				}
				err := invoke()
				if (mode == "guest") != (err == nil) {
					t.Fatal(mode, err)
				}
			}
			if drops != 42 {
				t.Fatal(mode, "owned payload cleanup", drops)
			}
			probe := canonical.NewResourceTable(in.resourceBudget)
			for range in.config.MaxResources {
				if _, err := probe.AddOpaque(1); err != nil {
					t.Fatal(mode, "owned payload stranded", err)
				}
			}
			probe.Close()
		}
	}
}

func TestHostSourceCleanupPreservesExecutionPanic(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	closed := 0
	var caught any
	func() {
		defer func() { caught = recover() }()
		in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			for range 2 {
				if _, err := c.NewSource(typ, AsyncSource{
					Read:  func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink },
					Close: func() { closed++; panic("cleanup") },
				}); err != nil {
					return err
				}
			}
			panic("execution")
		})
	}()
	if caught != "execution" || closed != 2 || in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
		t.Fatal("cleanup lost original panic or retained state", caught, closed)
	}
	probe := canonical.NewResourceTable(in.resourceBudget)
	defer probe.Close()
	for range in.config.MaxResources {
		if _, err := probe.AddOpaque(1); err != nil {
			t.Fatal("panic retained endpoint quota", err)
		}
	}
}
