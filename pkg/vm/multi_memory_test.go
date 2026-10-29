package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMultipleMemoryViewsAliasesGrowthAndBulk(t *testing.T) {
	input, err := os.ReadFile("testdata/multi-memory.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		if err := base.Validate(nil, input); err == nil {
			t.Fatal("multi-memory enabled by default")
		}
		for _, alias := range []bool{false, true} {
			for _, metered := range []bool{false, true} {
				cfg := base.config
				cfg.Features = wasm.MultiMemory
				e, err := NewEngineWithConfig(cfg)
				if err != nil {
					t.Fatal(err)
				}
				s, err := e.NewStore(DefaultStoreConfig())
				if err != nil {
					t.Fatal(err)
				}
				l := NewLinker(s)
				left, err := s.NewMemory(MemoryType{Min: 1, Max: 4, HasMax: true})
				if err != nil {
					t.Fatal(err)
				}
				right := left
				if !alias {
					right, err = s.NewMemory(MemoryType{Min: 1, Max: 4, HasMax: true})
					if err != nil {
						t.Fatal(err)
					}
				}
				global, err := s.NewGlobal(GlobalType{Value: wasm.I32, Mutable: true}, I32(7))
				if err != nil {
					t.Fatal(err)
				}
				var grow Extern
				host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
					v, err := c.Call(grow, []Value{I32(1)})
					if err != nil {
						return nil, err
					}
					if bits(t, v.Values[0]) != 1 {
						t.Fatal("wrong memory grew", v)
					}
					var word [4]byte
					binary.LittleEndian.PutUint32(word[:], 99)
					return nil, c.WriteMemory(1, 65536, word[:])
				})
				if err != nil {
					t.Fatal(err)
				}
				for name, ex := range map[string]Extern{"left": left, "right": right, "g0": global, "g1": global, "mutate": host} {
					if err := l.Define("h", name, ex); err != nil {
						t.Fatal(err)
					}
				}
				m, err := e.Compile(nil, input)
				if err != nil {
					t.Fatal(err)
				}
				in := instantiate(t, l, m)
				grow = export(t, in, "grow-right")
				options := CallOptions{Metered: metered, Fuel: 10000}
				r, err := export(t, in, "mixed").Call(nil, nil, options)
				want := uint64(171)
				if alias {
					want++
				}
				if err != nil || bits(t, r.Values[0]) != want {
					t.Fatal(alias, metered, r, err)
				}
				g, err := global.GlobalGet()
				if err != nil || bits(t, g) != 19 {
					t.Fatal(g, err)
				}
				r, err = export(t, in, "sizes").Call(nil, nil, options)
				first := uint64(1)
				if alias {
					first = 2
				}
				if err != nil || bits(t, r.Values[0]) != first || bits(t, r.Values[1]) != 2 || bits(t, r.Values[2]) != 1 {
					t.Fatal(r, err)
				}
				pattern := make([]byte, 10000)
				for i := range pattern {
					pattern[i] = byte(i % 251)
				}
				if err := left.WriteMemory(0, pattern); err != nil {
					t.Fatal(err)
				}
				if _, err := export(t, in, "copy").Call(nil, []Value{I32(2501), I32(0), I32(7000)}, options); err != nil {
					t.Fatal(err)
				}
				got := make([]byte, 7000)
				if err := right.ReadMemory(2501, got); err != nil || !bytes.Equal(got, pattern[:7000]) {
					t.Fatal("overlapping/cross-memory chunks", alias, err)
				}
				before := make([]byte, 32)
				if err := right.ReadMemory(0, before); err != nil {
					t.Fatal(err)
				}
				_, err = export(t, in, "copy").Call(nil, []Value{I32(0), I32(0xfffffff0), I32(32)}, options)
				var trap *wasm.Trap
				if !errors.As(err, &trap) || trap.Code != wasm.TrapMemoryBounds {
					t.Fatal(err)
				}
				after := make([]byte, 32)
				if err := right.ReadMemory(0, after); err != nil || !bytes.Equal(before, after) {
					t.Fatal("bounds failure wrote destination", err)
				}
				vector := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
				if err := right.WriteMemory(0, vector); err != nil {
					t.Fatal(err)
				}
				r, err = export(t, in, "vector").Call(nil, nil, options)
				if err != nil {
					t.Fatal(err)
				}
				lo, hi, err := r.Values[0].Bits()
				if err != nil || lo != binary.LittleEndian.Uint64(vector[:8]) || hi != binary.LittleEndian.Uint64(vector[8:]) {
					t.Fatal(r, err)
				}
				private := export(t, in, "private")
				if _, err := export(t, in, "fill").Call(nil, []Value{I32(123), I32(0xab), I32(5000)}, options); err != nil {
					t.Fatal(err)
				}
				filled := make([]byte, 5000)
				if err := private.ReadMemory(123, filled); err != nil || !bytes.Equal(filled, bytes.Repeat([]byte{0xab}, 5000)) {
					t.Fatal(err)
				}
				s.Close()
				m.Close()
				e.Close()
			}
		}
	})
}

func TestMultipleMemoryEmptyZeroAndQuotaRollback(t *testing.T) {
	input, err := os.ReadFile("testdata/multi-memory-basic.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Features = wasm.MultiMemory
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		for _, limit := range []uint32{1, 2} {
			config := DefaultStoreConfig()
			config.Limits.MaxMemories = limit
			s, err := e.NewStore(config)
			if err != nil {
				t.Fatal(err)
			}
			in, err := NewLinker(s).Instantiate(nil, m, CallOptions{})
			if limit == 1 {
				if !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(err)
				}
				u, err := s.Usage()
				if err != nil || u.Core.Instances != 0 || u.Core.Functions != 0 || u.Core.Memories != 0 || u.Core.MemoryPages != 0 {
					t.Fatal(u, err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				r, err := export(t, in, "run").Call(nil, nil, CallOptions{Metered: true, Fuel: 2})
				if err != nil || bits(t, r.Values[0]) != 42 || r.FuelRemaining != 0 {
					t.Fatal(r, err)
				}
			}
			s.Close()
		}
	})
}
