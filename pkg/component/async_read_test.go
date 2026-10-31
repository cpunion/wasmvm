package component

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
	"time"
)

func sourceType(t *testing.T, backend, name string) (*Instance, ValueType) {
	t.Helper()
	var typ ValueType
	in := asyncInstance(t, backend, name, Imports{FunctionChecks: map[string]FunctionCheck{"make": func(info FunctionInfo) error { typ = info.Result(); return nil }}, Functions: map[string]HostFunction{
		"make": func(*Caller, []Value) (*Value, error) { return nil, vm.ErrLink }, "started": func(*Caller, []Value) (*Value, error) { return &Value{}, nil },
	}})
	return in, typ
}
func TestHostSourceEmbeddingRead(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			calls, closes := 0, 0
			value, err := in.NewSource(nil, typ, AsyncSource{Close: func() { closes++ }, Read: func(c *Caller, n uint32) (*ReadCompletion, error) {
				calls++
				p := NewReadCompletion()
				p.Resolve(func(*Caller) (ReadResult, error) {
					if n == 0 {
						return ReadResult{}, nil
					}
					return ReadResult{Bytes: "abc"[:min(n, 3)], Closed: calls >= 4}, nil
				})
				return p, nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			for _, n := range []uint32{0, 1, 2, 3} {
				r, err := in.ReadAsync(nil, value.Resource, n, CallOptions{})
				if err != nil || r.Count != n || r.Bytes != "abc"[:n] || r.Closed != (n == 3) {
					t.Fatal(n, r, err)
				}
			}
			if closes != 1 {
				t.Fatal("EOF did not close source", closes)
			}
			if _, err := in.ReadAsync(nil, value.Resource, 1, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("read after DONE", err)
			}
			if _, err := in.DropAsync(nil, value.Resource, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"host-source-future-u32", "host-source-future-string", "host-source-stream-unit"} {
				other, typ := sourceType(t, backend, name)
				v, err := other.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
					p := NewReadCompletion()
					p.Resolve(func(*Caller) (ReadResult, error) {
						r := ReadResult{Values: []Value{{Bits: 42}}, Closed: true}
						if name == "host-source-future-string" {
							r.Values = []Value{{Text: "雪😀"}}
						}
						if name == "host-source-stream-unit" {
							r.Values = nil
							r.Count = 1
						}
						return r, nil
					})
					return p, nil
				}})
				if err != nil {
					t.Fatal(err)
				}
				r, err := other.ReadAsync(nil, v.Resource, 1, CallOptions{})
				if err != nil || !r.Closed || r.Count != 1 {
					t.Fatal(name, r, err)
				}
				if name == "host-source-future-u32" && r.Values[0].Bits != 42 || name == "host-source-future-string" && r.Values[0].Text != "雪😀" {
					t.Fatal(name, r)
				}
				if _, err := in.ReadAsync(nil, v.Resource, 1, CallOptions{}); !errors.Is(err, vm.ErrReference) {
					t.Fatal("foreign reader", err)
				}
				other.DropAsync(nil, v.Resource, CallOptions{})
			}
			if _, err := in.NewSource(nil, ValueType{}, AsyncSource{}); !errors.Is(err, vm.ErrType) {
				t.Fatal(err)
			}
			if _, err := in.NewSource(nil, typ, AsyncSource{}); !errors.Is(err, vm.ErrType) {
				t.Fatal(err)
			}
			if _, err := in.ReadAsync(nil, nil, 1, CallOptions{}); err == nil {
				t.Fatal("nil reader")
			}
			if _, err := new(Caller).ReadAsync(nil, 1); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
		})
	}
}

func TestHostSourceEmbeddingCancellationAndClose(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, closeInstance := range []bool{false, true} {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			started := make(chan context.Context, 1)
			closed := 0
			v, err := in.NewSource(nil, typ, AsyncSource{Close: func() { closed++ }, Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
				started <- c.Context()
				return NewReadCompletion(), nil
			}})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			done := make(chan error, 1)
			go func() { _, err := in.ReadAsync(ctx, v.Resource, 1, CallOptions{}); done <- err }()
			workCtx := <-started
			if closeInstance {
				in.Close()
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) && !errors.Is(err, wasm.ErrClosed) {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("suspended read survived shutdown")
			}
			cancel()
			if closed != 1 || workCtx.Err() == nil || in.tasks != nil || in.heldBytes != 0 || in.heldValues != 0 {
				t.Fatal("shutdown retained source", closed, workCtx.Err())
			}
			if !closeInstance {
				if _, err := in.DropAsync(nil, v.Resource, CallOptions{}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestHostSourceIdleShutdownClosesEverySource(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	closed := 0
	for range 3 {
		if _, err := in.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { t.Fatal("idle source started"); return nil, nil }, Close: func() { closed++ }}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := in.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	in.Close()
	in.Close()
	if err := in.WaitClosed(nil); err != nil || closed != 3 {
		t.Fatal("idle source cleanup", closed, err)
	}
}
func TestHostSourceShutdownContinuesAfterClosePanic(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	closed := 0
	for _, panics := range []bool{true, false, true} {
		if _, err := in.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }, Close: func() {
			closed++
			if panics {
				panic("close")
			}
		}}); err != nil {
			t.Fatal(err)
		}
	}
	func() {
		defer func() {
			if recover() != "close" {
				t.Fatal("close panic lost")
			}
		}()
		in.Close()
	}()
	if err := in.WaitClosed(nil); err != nil || closed != 3 {
		t.Fatal("partial close", closed, err)
	}
	in.Close()
}
