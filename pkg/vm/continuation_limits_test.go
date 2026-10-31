package vm

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestContinuationSharedFrameQuotaCanShrinkAndGrow(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e := newEngine()
		cfg := DefaultStoreConfig()
		cfg.Execution.MaxFrames = 2
		s, err := e.NewStore(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		l := NewLinker(s)
		ready := false
		host, err := s.NewSuspendableFunction(FunctionType{Results: []wasm.ValueType{wasm.I32}}, func(*Caller, []Value) (HostResult, error) {
			if ready {
				return HostResult{Values: []Value{I32(7)}}, nil
			}
			return HostResult{Pending: &HostSuspension{Ready: func() bool { return ready }, Resume: func(*Caller) (HostResult, error) { return HostResult{Values: []Value{I32(0)}}, nil }}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		l.Define("h", "pause", host)
		d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Imports: []module.Import{{Module: "h", Name: "pause", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 0x1a}, {Op: 0x10, A: 2}, {Op: 0xb}}}, {Body: module.Expr{{Op: 0x10}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Kind: 0, Index: 1}}}
		fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
		_, err = s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			a, err := c.Start(fn, nil)
			if err != nil {
				return err
			}
			b, err := c.Start(fn, nil)
			if err != nil {
				return err
			}
			if _, err := c.Start(fn, nil); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("parked frames escaped quota", err)
			}
			for _, call := range []*Continuation{a, b} {
				if _, done, err := c.Poll(call); err != nil || done {
					t.Fatal(done, err)
				}
			}
			ready = true
			if _, _, err := c.Poll(a); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("peer retained frame not charged", err)
			}
			result, done, err := c.Poll(b)
			if err != nil || !done || bits(t, result.Values[0]) != 7 {
				t.Fatal("released peer budget unavailable", result, done, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestContinuationLimitsAndInvalidCapabilities(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e := newEngine()
		cfg := DefaultStoreConfig()
		cfg.MaxContinuations = 1
		s, err := e.NewStore(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		memory, _ := s.NewMemory(MemoryType{Min: 1})
		fn, err := s.NewSuspendableFunction(FunctionType{}, func(*Caller, []Value) (HostResult, error) {
			return HostResult{Pending: &HostSuspension{Ready: func() bool { return false }, Resume: func(*Caller) (HostResult, error) { return HostResult{}, nil }}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.NewSuspendableFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		var expired *Caller
		_, err = s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			expired = c
			if _, err := c.Start(Extern{}, nil); !errors.Is(err, ErrReference) {
				t.Fatal(err)
			}
			if _, err := c.Start(memory, nil); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
			if _, err := c.Start(fn, []Value{I32(0)}); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
			if _, err := c.NewSuspendableFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
			if _, _, err := c.Poll(nil); !errors.Is(err, ErrReference) {
				t.Fatal(err)
			}
			if err := c.Abort(nil); !errors.Is(err, ErrReference) {
				t.Fatal(err)
			}
			call, err := c.Start(fn, nil)
			if err != nil {
				return err
			}
			if _, err := c.Start(fn, nil); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("continuation quota", err)
			}
			if err := c.Abort(call); err != nil {
				return err
			}
			if err := c.Abort(call); err != nil {
				return err
			}
			_, err = c.Start(fn, nil)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := expired.Start(fn, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if err := expired.Abort(nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if _, _, err := expired.Poll(nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	})
}
