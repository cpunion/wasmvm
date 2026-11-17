package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestReadyTypedSourcesCopyWithoutTasksAndKeepFallback(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, name := range []string{"host-source-future-u32", "host-source-future-string", "host-source-stream-u32", "host-source-stream-unit", "host-source-stream-u8"} {
			t.Run(backend+"/"+name, func(t *testing.T) {
				in, typ := sourceType(t, backend, name)
				for _, ready := range []bool{true, false} {
					probes, starts, closes := 0, 0, 0
					want := ReadResult{Values: []Value{{Bits: 42}}, Closed: true}
					if name == "host-source-future-u32" || name == "host-source-future-string" {
						want.Closed = false // A single future value also implies EOF.
					}
					switch name {
					case "host-source-future-string":
						want.Values[0] = Value{Text: "雪😀"}
					case "host-source-stream-unit":
						want.Values, want.Count = nil, 1
					case "host-source-stream-u8":
						want.Values, want.Bytes = nil, "\x00\xff"
					}
					limit := uint32(1)
					if want.Bytes != "" {
						limit = 2
					}
					v, err := in.NewSource(nil, typ, AsyncSource{ReadReady: func(_ *Caller, n uint32) (ReadResult, bool, error) {
						probes++
						if n != limit {
							t.Fatal("read limit changed", n)
						}
						if ready {
							return want, true, nil
						}
						return ReadResult{}, false, nil
					}, Read: func(*Caller, uint32) (*ReadCompletion, error) {
						starts++
						p := NewReadCompletion()
						p.Resolve(func(*Caller) (ReadResult, error) { return want, nil })
						return p, nil
					}, Close: func() { closes++ }})
					if err != nil {
						t.Fatal(err)
					}
					got, err := in.ReadAsync(nil, v.Resource, limit, CallOptions{})
					if err != nil || !got.Closed || got.Count != limit || got.Bytes != want.Bytes || len(got.Values) != len(want.Values) || probes != 1 || ready && starts != 0 || !ready && starts != 1 || closes != 1 {
						t.Fatal(ready, got, err, probes, starts, closes)
					}
					if len(got.Values) != 0 {
						if got.Values[0].Bits != want.Values[0].Bits || got.Values[0].Text != want.Values[0].Text {
							t.Fatal("typed payload changed", got, want)
						}
						got.Values[0] = Value{Bits: 99}
						if want.Values[0].Bits == 99 {
							t.Fatal("ready source exposed producer storage")
						}
					}
					if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 || in.callDepth != 0 {
						t.Fatal("completed source retained invocation state")
					}
					if _, err := in.DropAsync(nil, v.Resource, CallOptions{}); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func TestReadySourceProbeContracts(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	probe := func(*Caller, uint32) (ReadResult, bool, error) {
		t.Fatal("invalid source started")
		return ReadResult{}, false, nil
	}
	read := func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }
	for _, source := range []AsyncSource{{ReadReady: probe}, {ReadReady: probe, ReadBytes: func(uint32) (ReadResult, bool, error) {
		t.Fatal("invalid byte probe started")
		return ReadResult{}, false, nil
	}, Read: read}} {
		if _, err := in.NewSource(nil, typ, source); !errors.Is(err, vm.ErrType) {
			t.Fatal("invalid ready source admitted", err)
		}
	}
}

func TestReadySourceKeepsTaskReservationDuringCallback(t *testing.T) {
	for _, ready := range []bool{true, false} {
		in, typ := sourceType(t, "interpreter", "host-source-future-u32")
		in.config.MaxTasks = 1
		v, err := in.NewSource(nil, typ, AsyncSource{ReadReady: func(c *Caller, _ uint32) (ReadResult, bool, error) {
			if err := c.Schedule(AsyncWork{Start: func(*Caller, []*Resource) (*HostCompletion, error) {
				t.Fatal("over-quota work started")
				return nil, nil
			}}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("ready callback lost its logical task reservation", err)
			}
			if !ready {
				return ReadResult{}, false, nil
			}
			return ReadResult{Values: []Value{{Bits: 42}}}, true, nil
		}, Read: func(*Caller, uint32) (*ReadCompletion, error) {
			p := NewReadCompletion()
			p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Values: []Value{{Bits: 42}}}, nil })
			return p, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		if got, err := in.ReadAsync(nil, v.Resource, 1, CallOptions{}); err != nil || got.Values[0].Bits != 42 || in.tasks != nil {
			t.Fatal(got, err)
		}
		if _, err := in.DropAsync(nil, v.Resource, CallOptions{}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadyTypedSourceFailuresAndAdmission(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, mode := range []string{"count", "bounds", "future-empty", "invalid-value", "mixed", "not-ready-data", "error", "panic", "close-panic", "values", "bytes", "cancel", "tasks", "reentry"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				in, typ := sourceType(t, backend, "host-source-future-u32")
				closes := 0
				admission := mode == "cancel" || mode == "tasks" || mode == "reentry"
				v, err := in.NewSource(nil, typ, AsyncSource{ReadReady: func(*Caller, uint32) (ReadResult, bool, error) {
					if admission {
						t.Fatal("rejected read invoked producer")
					}
					r := ReadResult{Values: []Value{{Bits: 42}}, Closed: true}
					switch mode {
					case "count":
						r.Count = 2
					case "bounds":
						r.Values = append(r.Values, Value{})
					case "future-empty":
						r.Values = nil
					case "invalid-value":
						r.Values[0].Bits = 1 << 32
					case "mixed":
						r.Bytes = "x"
					case "not-ready-data":
						return r, false, nil
					case "error":
						return ReadResult{}, false, vm.ErrLink
					case "panic":
						panic(mode)
					}
					return r, true, nil
				}, Read: func(*Caller, uint32) (*ReadCompletion, error) {
					t.Fatal("failed ready source reached deferred path")
					return nil, nil
				}, Close: func() {
					closes++
					if mode == "close-panic" {
						panic(mode)
					}
				}})
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				switch mode {
				case "values":
					in.config.MaxValues = 1
				case "bytes":
					in.config.MaxBytes = 1
				case "cancel":
					cancel()
				case "tasks":
					in.config.MaxTasks = 0
				case "reentry":
					in.config.Store.MaxReentry = 0
				}
				invoke := func() (err error) {
					defer func() {
						if p := recover(); p != nil {
							if p != mode {
								t.Fatal("original panic replaced", p)
							}
							err = wasm.ErrTrap
						}
					}()
					_, err = in.ReadAsync(ctx, v.Resource, 1, CallOptions{})
					return err
				}
				if err := invoke(); err == nil || mode == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatal(mode, err)
				}
				if !admission && closes != 1 || in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 || in.callDepth != 0 {
					t.Fatal("failed source retained transaction state", closes, in.heldValues, in.heldBytes)
				}
				in.config.MaxTasks, in.config.Store.MaxReentry = 100, 100
				if _, err := in.DropAsync(nil, v.Resource, CallOptions{}); err != nil || closes != 1 {
					t.Fatal("failed source did not close exactly once", closes, err)
				}
			})
		}
	}
}

func TestReadyTypedSourceOwnedTransfers(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, mode := range []string{"guest", "host", "count", "close-panic", "unread", "fallback"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				drops, closes := 0, 0
				rt := NewHostResourceType(func(*Caller, uint32) error { drops++; return nil })
				var typ ValueType
				var expired []*Caller
				makeSource := func(c *Caller, _ []Value) (*Value, error) {
					resource, err := c.NewResource(rt, 42)
					if err != nil {
						return nil, err
					}
					var owned []*Resource
					value, moved, err := c.NewOwnedSource(typ, AsyncSource{Owned: []*Resource{resource.Resource}, ReadReady: func(fresh *Caller, _ uint32) (ReadResult, bool, error) {
						alias := *fresh
						expired = append(expired, fresh, &alias)
						if _, err := fresh.Call(nil, nil); !errors.Is(err, vm.ErrType) {
							t.Fatal("ready callback permitted guest reentry", err)
						}
						if mode == "fallback" {
							_, err := fresh.NewResource(rt, 99)
							return ReadResult{}, false, err
						}
						r := ReadResult{Values: []Value{{Resource: owned[0]}}, Closed: true}
						if mode == "count" {
							r.Count = 2
						}
						return r, true, nil
					}, Read: func(*Caller, uint32) (*ReadCompletion, error) {
						if mode != "fallback" || drops != 1 {
							t.Fatal("probe did not roll back speculative ownership", mode, drops)
						}
						p := NewReadCompletion()
						p.Resolve(func(*Caller) (ReadResult, error) {
							return ReadResult{Values: []Value{{Resource: owned[0]}}, Closed: true}, nil
						})
						return p, nil
					}, Close: func() {
						closes++
						if mode == "close-panic" {
							panic(mode)
						}
					}})
					owned = moved
					if err == nil {
						if _, _, _, _, _, err := canonical.ReferenceInfo(resource.Resource); err == nil {
							t.Fatal("source construction kept sender ownership")
						}
					}
					return &value, err
				}
				in := asyncInstance(t, backend, "host-source-future-owned", Imports{Resources: map[string]*HostResourceType{"r": rt}, FunctionChecks: map[string]FunctionCheck{"make": func(info FunctionInfo) error { typ = info.Result(); return nil }}, Functions: map[string]HostFunction{"make": makeSource, "started": func(*Caller, []Value) (*Value, error) { return &Value{}, nil }}})
				if mode == "host" || mode == "unread" {
					_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
						value, err := makeSource(c, nil)
						if err != nil {
							return err
						}
						if mode == "host" {
							r, err := c.ReadAsync(value.Resource, 1)
							if err != nil || len(r.Values) != 1 || drops != 0 {
								t.Fatal(r, err, drops)
							}
							if err := c.DropResource(r.Values[0].Resource); err != nil {
								return err
							}
						}
						return c.DropAsync(value.Resource)
					})
					if err != nil {
						t.Fatal(err)
					}
				} else {
					run, _ := in.Function("run")
					invoke := func() (err error) {
						defer func() {
							if p := recover(); p != nil {
								if p != mode {
									t.Fatal(p)
								}
								err = wasm.ErrTrap
							}
						}()
						_, err = run.Call(nil, []Value{{}}, CallOptions{})
						return err
					}
					if err := invoke(); (mode == "guest" || mode == "fallback") != (err == nil) {
						t.Fatal(mode, err)
					}
				}
				wantDrops := 1
				if mode == "fallback" {
					wantDrops++
				}
				for _, old := range expired {
					if _, err := old.Call(nil, nil); !errors.Is(err, wasm.ErrClosed) {
						t.Fatal("probe Caller survived return", err)
					}
				}
				if drops != wantDrops || closes != 1 || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal("owned source leaked or duplicated cleanup", drops, closes, in.heldValues, in.heldBytes)
				}
				probe := canonical.NewResourceTable(in.resourceBudget)
				for range in.config.MaxResources {
					if _, err := probe.AddOpaque(1); err != nil {
						t.Fatal("source retained resource quota", err)
					}
				}
				probe.Close()
			})
		}
	}
}
