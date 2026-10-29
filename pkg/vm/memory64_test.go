package vm

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	enginememory "github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMemory64Table64MixedWidthsAndReentry(t *testing.T) {
	input, err := os.ReadFile("testdata/memory64.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		if err := base.Validate(nil, input); err == nil {
			t.Fatal("64-bit indices enabled by default")
		}
		for _, metered := range []bool{false, true} {
			cfg := base.config
			cfg.Features = wasm.Memory64 | wasm.MultiMemory | wasm.TailCall
			e, err := NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			sc := DefaultStoreConfig()
			sc.Limits.MaxMemoryPages, sc.Limits.MaxTableElements = 4, 4
			sc.Execution.MaxFrames = 2
			s, err := e.NewStore(sc)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			m32, err := s.NewMemory(MemoryType{Min: 1, Max: 4, HasMax: true})
			if err != nil {
				t.Fatal(err)
			}
			m64, err := s.NewMemory(MemoryType{Min: 1, Max: 4, HasMax: true, Index64: true})
			if err != nil {
				t.Fatal(err)
			}
			t32, err := s.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Min: 1, Max: 4, HasMax: true}})
			if err != nil {
				t.Fatal(err)
			}
			t64, err := s.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Min: 1, Max: 4, HasMax: true, Index64: true}})
			if err != nil {
				t.Fatal(err)
			}
			var grow, growTable Extern
			host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
				for _, ex := range []Extern{grow, growTable} {
					r, err := c.Call(ex, []Value{I64(1)})
					if err != nil {
						return nil, err
					}
					if r.Values[0].Type() != wasm.I64 || bits(t, r.Values[0]) != 1 {
						t.Fatal("growth result width", r)
					}
				}
				return nil, c.WriteMemory(1, 65536, []byte{99, 0, 0, 0})
			})
			if err != nil {
				t.Fatal(err)
			}
			l := NewLinker(s)
			for name, ex := range map[string]Extern{"m32": m32, "m64": m64, "t32": t32, "t64": t64, "mutate": host} {
				if err := l.Define("h", name, ex); err != nil {
					t.Fatal(err)
				}
			}
			m, err := e.Compile(nil, input)
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			in := instantiate(t, l, m)
			grow, growTable = export(t, in, "grow"), export(t, in, "table-grow")
			options := CallOptions{Metered: metered, Fuel: 100000}
			call := func(name string, args ...Value) CallResult {
				r, err := export(t, in, name).Call(nil, args, options)
				if err != nil {
					t.Fatal(name, err)
				}
				return r
			}
			trap := func(name string, code wasm.TrapCode, args ...Value) {
				_, err := export(t, in, name).Call(nil, args, options)
				var got *wasm.Trap
				if !errors.As(err, &got) || got.Code != code {
					t.Fatal(name, err)
				}
			}
			for _, name := range []string{"call", "tail", "load"} {
				if got := bits(t, call(name, I64(0)).Values[0]); got != 42 {
					t.Fatal(name, got)
				}
			}
			if got := bits(t, call("touch").Values[0]); got != 99 {
				t.Fatal("memory view stale after reentry", got)
			}
			for _, name := range []string{"size", "table-size"} {
				r := call(name)
				if r.Values[0].Type() != wasm.I64 || bits(t, r.Values[0]) != 2 {
					t.Fatal(name, r)
				}
			}
			for _, high := range []uint64{1 << 32, 1 << 63, ^uint64(0)} {
				for _, name := range []string{"load", "offset-high", "offset-overflow"} {
					trap(name, wasm.TrapMemoryBounds, I64(high))
				}
				trap("store", wasm.TrapMemoryBounds, I64(high), I32(77))
				trap("vector", wasm.TrapMemoryBounds, I64(high), I64(0))
				for _, name := range []string{"call", "tail", "get"} {
					trap(name, wasm.TrapTableBounds, I64(high))
				}
				for _, name := range []string{"grow", "table-grow"} {
					if got := bits(t, call(name, I64(high)).Values[0]); got != ^uint64(0) {
						t.Fatal("truncated growth delta or failure", name, high, got)
					}
				}
			}
			trap("offset-overflow", wasm.TrapMemoryBounds, I64(1))
			trap("copy-64", wasm.TrapMemoryBounds, I64(0), I64(1), I64(^uint64(0)))
			trap("fill", wasm.TrapMemoryBounds, I64(1), I32(0), I64(1<<32))
			trap("table-copy-64", wasm.TrapTableBounds, I64(0), I64(1), I64(1<<32))
			call("init", I64(16), I32(0), I32(6))
			call("vector", I64(32), I64(16))
			call("copy-to-32", I32(48), I64(32), I32(16))
			call("copy-to-64", I64(64), I32(48), I32(16))
			call("copy-64", I64(66), I64(64), I64(6))
			call("fill", I64(72), I32(33), I64(2))
			var data [10]byte
			if err := m64.ReadMemory(64, data[:]); err != nil || !bytes.Equal(data[:], []byte("ababcdef!!")) {
				t.Fatal(data, err)
			}
			call("table-init", I64(1), I32(0), I32(1))
			ref := call("get", I64(1)).Values[0]
			call("set", I64(0), ref)
			call("table-fill", I64(0), ref, I64(2))
			call("table-copy-to-32", I32(0), I64(1), I32(1))
			call("table-copy-to-64", I64(0), I32(0), I32(1))
			call("table-copy-64", I64(1), I64(0), I64(1))
			ref.Release()
			if got := bits(t, call("tail", I64(1)).Values[0]); got != 42 {
				t.Fatal("table roots or descriptor stale", got)
			}
			for _, mismatch := range []string{"m64", "t64"} {
				wrong := NewLinker(s)
				for name, ex := range map[string]Extern{"m32": m32, "m64": m64, "t32": t32, "t64": t64, "mutate": host} {
					if name == mismatch {
						ex = m32
						if name == "t64" {
							ex = t32
						}
					}
					if err := wrong.Define("h", name, ex); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := wrong.Instantiate(nil, m, CallOptions{}); !errors.Is(err, ErrLink) {
					t.Fatal("import address width mismatch accepted", mismatch, err)
				}
			}
		}
	})
}

func TestMemory64EverySIMDMemoryInstruction(t *testing.T) {
	input, err := os.ReadFile("testdata/memory64-simd.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Features = wasm.Memory64
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
		s, err := e.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		in := instantiate(t, NewLinker(s), m)
		mem := export(t, in, "memory")
		ref, err := enginememory.New(MemoryType{Min: 1, Index64: true}, 1)
		if err != nil {
			t.Fatal(err)
		}
		defer ref.Release()
		v := V128(0x8877665544332211, 0xffeeddccbbaa0099)
		for _, metered := range []bool{false, true} {
			for op := uint32(0); op <= 93; op++ {
				if op > 11 && op < 84 {
					continue
				}
				lane := byte(0)
				if op >= 84 && op <= 91 {
					lane = byte((16 >> ((op - 84) % 4)) - 1)
				}
				for _, address := range []uint64{0, 1, 48, 65536, 1 << 32, ^uint64(0)} {
					var before, want, got [96]byte
					for i := range before {
						before[i] = byte(i*13 + 129)
					}
					ref.Write(0, before[:])
					if err := mem.WriteMemory(0, before[:]); err != nil {
						t.Fatal(err)
					}
					result, trap := v.bits, wasm.NoTrap
					if op == 11 || op >= 88 && op <= 91 {
						trap = ref.Store(0xfd0000|op, address, 3, lane, v.bits)
					} else {
						result, trap = ref.Load(0xfd0000|op, address, 3, lane, v.bits)
					}
					r, err := export(t, in, fmt.Sprint("op-", op)).Call(nil, []Value{I64(address), v}, CallOptions{Metered: metered, Fuel: 100})
					if trap != wasm.NoTrap {
						var actual *wasm.Trap
						if !errors.As(err, &actual) || actual.Code != trap {
							t.Fatal(op, address, err, trap)
						}
					} else if err != nil || r.Values[0].bits != result {
						t.Fatal(op, address, r, err, result)
					}
					ref.Read(0, want[:])
					if err := mem.ReadMemory(0, got[:]); err != nil || got != want {
						t.Fatal(op, address, "memory effects", err)
					}
				}
			}
		}
	})
}

func TestMemory64EmptyObjectsFeatureGatesAndQuotas(t *testing.T) {
	input, err := os.ReadFile("testdata/memory64-empty.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		base := newEngine()
		blocked, err := base.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer blocked.Close()
		if _, err := blocked.NewMemory(MemoryType{Index64: true}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal("host memory bypassed feature gate", err)
		}
		if _, err := blocked.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Index64: true}}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal("host table bypassed feature gate", err)
		}
		cfg := base.config
		cfg.Features = wasm.Memory64
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
		s, err := e.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		in := instantiate(t, NewLinker(s), m)
		for _, metered := range []bool{false, true} {
			options := CallOptions{Metered: metered, Fuel: 100}
			for _, name := range []string{"run", "table-size"} {
				r, err := export(t, in, name).Call(nil, nil, options)
				if err != nil || r.Values[0].Type() != wasm.I64 || bits(t, r.Values[0]) != 0 {
					t.Fatal(name, r, err)
				}
			}
			for _, name := range []string{"grow", "table-grow"} {
				for _, delta := range []uint64{0, 1, 1 << 32, ^uint64(0)} {
					r, err := export(t, in, name).Call(nil, []Value{I64(delta)}, options)
					want := ^uint64(0)
					if delta == 0 {
						want = 0
					}
					if err != nil || r.Values[0].Type() != wasm.I64 || bits(t, r.Values[0]) != want {
						t.Fatal(name, delta, r, err)
					}
				}
			}
			_, err := export(t, in, "load").Call(nil, []Value{I64(0)}, options)
			var trap *wasm.Trap
			if !errors.As(err, &trap) || trap.Code != wasm.TrapMemoryBounds {
				t.Fatal("empty memory load", err)
			}
		}
		before, err := s.Usage()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.NewMemory(MemoryType{Min: 1 << 48, Index64: true}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("valid large memory declaration must meet the host quota", err)
		}
		if _, err := s.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Min: 1 << 63, Index64: true}}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("table quota truncated", err)
		}
		after, err := s.Usage()
		if err != nil || before != after {
			t.Fatal("failed allocation changed accounting", before, after, err)
		}
		if _, err := s.NewMemory(MemoryType{Max: ^uint64(0), HasMax: true}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("invalid memory32 limit accepted", err)
		}
		if _, err := s.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Max: 1 << 32, HasMax: true}}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("invalid table32 limit accepted", err)
		}
	})
}
