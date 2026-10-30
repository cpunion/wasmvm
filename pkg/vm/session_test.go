package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSessionSharesFuelAcrossCallsAndInstantiation(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		m, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		for _, fuel := range []uint64{5, 6} {
			remaining, err := s.WithCaller(nil, CallOptions{Metered: true, Fuel: fuel}, func(c *Caller) error {
				in, err := c.Instantiate(l, m)
				if err != nil {
					return err
				}
				fn, err := c.Export(in, "add")
				if err != nil {
					return err
				}
				for i := 0; i < 2; i++ {
					result, err := c.Call(fn, []Value{I32(4), I32(5)})
					if err != nil {
						return err
					}
					if bits(t, result.Values[0]) != 9 {
						t.Fatal(result)
					}
				}
				return nil
			})
			if fuel == 5 && !errors.Is(err, wasm.ErrLimit) || fuel == 6 && (err != nil || remaining != 0) {
				t.Fatal(fuel, remaining, err)
			}
		}
	})
}

func TestSessionMemoryCapabilitiesAndExpiration(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		_, foreign, foreignLinker := setupWithEngine(t, newEngine)
		memory, _ := s.NewMemory(MemoryType{Min: 1})
		other, _ := foreign.NewMemory(MemoryType{Min: 1})
		var saved *Caller
		_, err := s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			saved = c
			if err := c.WriteExternalMemory(memory, 42, []byte{7}); err != nil {
				return err
			}
			var bytes [1]byte
			if err := c.ReadExternalMemory(memory, 42, bytes[:]); err != nil || bytes[0] != 7 {
				t.Fatal(bytes, err)
			}
			if err := c.ReadExternalMemory(memory, 65536, bytes[:]); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			if err := c.WriteExternalMemory(memory, 65536, bytes[:]); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			for _, ex := range []Extern{other, {}} {
				if err := c.ReadExternalMemory(ex, 0, nil); !errors.Is(err, ErrReference) {
					t.Fatal(err)
				}
				if err := c.WriteExternalMemory(ex, 0, nil); !errors.Is(err, ErrReference) {
					t.Fatal(err)
				}
			}
			host, err := c.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, nil })
			if err != nil {
				return err
			}
			if _, err := c.Call(host, nil); err != nil {
				return err
			}
			if err := c.ReadExternalMemory(host, 0, nil); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
			if err := c.WriteExternalMemory(host, 0, nil); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
			if _, err := c.Instantiate(foreignLinker, nil); !errors.Is(err, ErrReference) {
				t.Fatal(err)
			}
			if _, err := c.Export(nil, "absent"); !errors.Is(err, ErrReference) {
				t.Fatal(err)
			}
			if err := memory.ReadMemory(0, nil); !errors.Is(err, ErrBusy) {
				t.Fatal(err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := saved.ReadExternalMemory(memory, 0, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("escaped session memory capability", err)
		}
		if _, err := saved.NewFunction(FunctionType{}, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if _, err := saved.Instantiate(nil, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if _, err := saved.Export(nil, ""); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	})
}

func TestSessionCleanupAfterPanicAndCancellation(t *testing.T) {
	e, s, _ := setup(t)
	defer e.Close()
	if _, err := s.WithCaller(nil, CallOptions{}, nil); !errors.Is(err, ErrType) {
		t.Fatal(err)
	}
	var saved *Caller
	sentinel := errors.New("session panic")
	func() {
		defer func() {
			if recover() != sentinel {
				t.Fatal("changed session panic")
			}
		}()
		s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			saved = c
			if _, err := c.NewExternRef("temporary"); err != nil {
				t.Fatal(err)
			}
			panic(sentinel)
		})
	}()
	if _, err := saved.NewExternRef(nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	usage, err := s.Usage()
	if err != nil || usage.HostRoots != 0 {
		t.Fatal(usage, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	_, err = s.WithCaller(ctx, CallOptions{}, func(*Caller) error { cancel(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	_, err = s.WithCaller(nil, CallOptions{}, func(*Caller) error { s.Close(); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
}
