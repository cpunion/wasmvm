package vm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func tailSetup(t *testing.T, base *Engine, frames uint32, callback HostFunction) (*Engine, *Store, *Linker, *Instance) {
	t.Helper()
	config := base.config
	config.Features = wasm.TailCall
	e, err := NewEngineWithConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	sc := DefaultStoreConfig()
	sc.Execution.MaxFrames, sc.Execution.MaxSlots = frames, 64
	s, err := e.NewStore(sc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close(); e.Close() })
	if callback == nil {
		callback = func(_ *Caller, args []Value) ([]Value, error) { return args, nil }
	}
	host, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}, callback)
	if err != nil {
		t.Fatal(err)
	}
	l := NewLinker(s)
	if err := l.Define("h", "echo", host); err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile("testdata/tail-call.wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.Compile(nil, input)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	return e, s, l, instantiate(t, l, m)
}

func TestTailRecursionUsesOneFrameAndPreservesTuples(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, _, _, in := tailSetup(t, newEngine(), 1, nil)
		for _, metered := range []bool{false, true} {
			for _, name := range []string{"direct", "indirect"} {
				n, a, b := uint32(20000), uint32(0xfffffff0), uint32(17)
				x, y := a, b
				for range n {
					x, y = y, x+1
				}
				got, err := export(t, in, name).Call(nil, []Value{I32(n), I32(a), I32(b)}, CallOptions{Metered: metered, Fuel: 1_000_000})
				if err != nil || bits(t, got.Values[0]) != uint64(x*3+y*5) {
					t.Fatal(name, metered, got, err)
				}
			}
		}
		if _, err := export(t, in, "normal").Call(nil, []Value{I32(2), I32(1), I32(2)}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("ordinary recursion changed into a tail call", err)
		}
	})
}

func TestTailMixedResultsAndGrowingFrames(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _, in := tailSetup(t, newEngine(), 2, nil)
		ref, err := export(t, in, "fuel").Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer ref.Release()
		object, err := s.NewExternRef("tail payload")
		if err != nil {
			t.Fatal(err)
		}
		defer object.Release()
		for _, name := range []string{"values", "parent-values"} {
			for _, metered := range []bool{false, true} {
				got, err := export(t, in, name).Call(nil, []Value{I32(25), I64(0x123456789abcdef0), V128(7, 11), ref, object}, CallOptions{Metered: metered, Fuel: 100})
				if err != nil || len(got.Values) != 5 {
					t.Fatal(got, err)
				}
				lo, hi, err := got.Values[2].Bits()
				if bits(t, got.Values[0]) != 42 || bits(t, got.Values[1]) != 0x123456789abcdef0 || lo != 7 || hi != 11 || err != nil {
					t.Fatal(got, err)
				}
				if payload, err := s.ExternRef(got.Values[4]); err != nil || payload != "tail payload" {
					t.Fatal(payload, err)
				}
				answer, err := s.CallRef(nil, got.Values[3], nil, CallOptions{})
				if err != nil || bits(t, answer.Values[0]) != 42 {
					t.Fatal(answer, err)
				}
				for _, v := range got.Values {
					v.Release()
				}
			}
		}
	})
}

func TestTailHostCallerMemoryCollectionAndReentryAcrossModules(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		var answer Extern
		calls := 0
		callback := func(c *Caller, args []Value) ([]Value, error) {
			calls++
			var marker [1]byte
			if err := c.ReadMemory(0, 0, marker[:]); err != nil || marker[0] != 'T' {
				t.Fatal("host received the wrong module", marker, err)
			}
			if err := c.Collect(); err != nil {
				return nil, err
			}
			if payload, err := c.Store().ExternRef(args[0]); err != nil || payload != "retained" {
				t.Fatal(payload, err)
			}
			r, err := c.Call(answer, nil)
			if err != nil || bits(t, r.Values[0]) != 42 {
				t.Fatal("tail frame was not retired before reentry", r, err)
			}
			return args, nil
		}
		e, s, _, provider := tailSetup(t, newEngine(), 2, callback)
		answer = export(t, provider, "fuel")
		input, err := os.ReadFile("testdata/tail-call-import.wasm")
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		l := NewLinker(s)
		if err := l.Define("p", "tail-host", export(t, provider, "tail-host")); err != nil {
			t.Fatal(err)
		}
		consumer := instantiate(t, l, m)
		payload, err := s.NewExternRef("retained")
		if err != nil {
			t.Fatal(err)
		}
		defer payload.Release()
		for _, name := range []string{"tail-import", "ordinary-import"} {
			for _, metered := range []bool{false, true} {
				r, err := export(t, consumer, name).Call(nil, []Value{payload}, CallOptions{Metered: metered, Fuel: 100})
				if err != nil {
					t.Fatal(name, r, err)
				}
				if got, err := s.ExternRef(r.Values[0]); err != nil || got != "retained" {
					t.Fatal(got, err)
				}
				r.Values[0].Release()
			}
		}
		if calls != 4 {
			t.Fatal(calls)
		}
	})
}

func TestTailFuelIndirectTrapsAndCancellation(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, _, _, in := tailSetup(t, newEngine(), 1, nil)
		for fuel := uint64(0); fuel <= 2; fuel++ {
			r, err := export(t, in, "fuel").Call(nil, nil, CallOptions{Metered: true, Fuel: fuel})
			if fuel < 2 {
				if !errors.Is(err, wasm.ErrLimit) || r.FuelRemaining != 0 {
					t.Fatal(fuel, r, err)
				}
			} else if err != nil || bits(t, r.Values[0]) != 42 || r.FuelRemaining != 0 {
				t.Fatal(fuel, r, err)
			}
		}
		for _, tc := range []struct {
			index uint32
			trap  wasm.TrapCode
		}{
			{0, wasm.TrapIndirectCallType}, {3, wasm.TrapUninitializedElement}, {0xffffffff, wasm.TrapTableBounds},
		} {
			_, err := export(t, in, "indirect-index").Call(nil, []Value{I32(tc.index)}, CallOptions{})
			var trap *wasm.Trap
			if !errors.As(err, &trap) || trap.Code != tc.trap {
				t.Fatal(tc, err)
			}
		}
		r, err := export(t, in, "indirect-index").Call(nil, []Value{I32(2)}, CallOptions{})
		if err != nil || bits(t, r.Values[0]) != 42 {
			t.Fatal(r, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_, err = export(t, in, "forever").Call(ctx, nil, CallOptions{})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	})
}

func TestTailFeatureSelectionAndComposition(t *testing.T) {
	input, err := os.ReadFile("testdata/tail-call-combined.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, features := range []wasm.Features{0, wasm.TailCall, wasm.ExtendedConst, wasm.TailCall | wasm.ExtendedConst} {
			cfg := newEngine().config
			cfg.Features = features
			e, err := NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, input)
			if features != (wasm.TailCall | wasm.ExtendedConst) {
				if err == nil {
					t.Fatal("extension implicitly enabled", features)
				}
				continue
			}
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
			r, err := export(t, in, "run").Call(nil, nil, CallOptions{})
			if err != nil || bits(t, r.Values[0]) != 42 {
				t.Fatal(r, err)
			}
		}
	})
}
