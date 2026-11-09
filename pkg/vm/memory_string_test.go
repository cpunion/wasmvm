package vm

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSessionImmutableMemoryCapabilities(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		_, other, _ := setupWithEngine(t, newEngine)
		memory, err := s.NewMemory(MemoryType{Min: 1, Max: 2, HasMax: true})
		if err != nil {
			t.Fatal(err)
		}
		foreign, _ := other.NewMemory(MemoryType{Min: 1})
		host, _ := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, nil })
		const data = "\x00\xff\x80hello"
		var saved *Caller
		var snapshot string
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err = s.WithCaller(ctx, CallOptions{}, func(c *Caller) error {
			saved = c
			if err := c.WriteExternalMemoryString(memory, 42, data); err != nil {
				return err
			}
			var err error
			snapshot, err = c.ReadExternalMemoryString(memory, 42, uint64(len(data)))
			if err != nil || snapshot != data {
				t.Fatal(snapshot, err)
			}
			if err := c.WriteExternalMemoryString(memory, 42, "changed!"); err != nil {
				t.Fatal(err)
			}
			for _, row := range []struct {
				memory Extern
				at, n  uint64
				want   error
			}{{memory, 65536, 1, wasm.ErrTrap}, {memory, math.MaxUint64, 1, wasm.ErrTrap}, {foreign, 0, 0, ErrReference}, {Extern{}, 0, 0, ErrReference}, {host, 0, 0, ErrType}} {
				if _, err := c.ReadExternalMemoryString(row.memory, row.at, row.n); !errors.Is(err, row.want) {
					t.Fatal(row, err)
				}
				if err := c.WriteExternalMemoryString(row.memory, row.at, "x"); !errors.Is(err, row.want) {
					t.Fatal(row, err)
				}
			}
			if got, err := c.ReadExternalMemoryString(memory, 65536, 0); err != nil || got != "" {
				t.Fatal(got, err)
			}
			if err := c.WriteExternalMemoryString(memory, 65536, ""); err != nil {
				t.Fatal(err)
			}
			c.mu.Lock()
			_, readBusy := c.ReadExternalMemoryString(memory, 0, 0)
			writeBusy := c.WriteExternalMemoryString(memory, 0, "")
			c.mu.Unlock()
			if !errors.Is(readBusy, ErrBusy) || !errors.Is(writeBusy, ErrBusy) {
				t.Fatal(readBusy, writeBusy)
			}
			cancel()
			if _, err := c.ReadExternalMemoryString(memory, 0, 0); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			if err := c.WriteExternalMemoryString(memory, 0, ""); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			return nil
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if _, err := saved.ReadExternalMemoryString(memory, 0, 0); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if err := saved.WriteExternalMemoryString(memory, 0, ""); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if _, ok, err := memory.GrowMemory(1); err != nil || !ok {
			t.Fatal(ok, err)
		}
		s.Close()
		if snapshot != data {
			t.Fatal("snapshot outlived its ownership", snapshot)
		}
	})
}
