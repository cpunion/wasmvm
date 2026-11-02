package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarHostBitsReentryAndIndependentStacks(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		types := []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64}
		input := []Value{I32(^uint32(0)), I64(^uint64(0)), F32Bits(0x80000000), F64Bits(0x7ff8000000001234)}
		wantInput := []uint64{0xffffffff, ^uint64(0), 0x80000000, 0x7ff8000000001234}
		wantOutput := []uint64{0xffffffff, 42, 0x7fc01234, 0xfff800000000babe}
		var retained [][]uint64
		var expired *Caller
		host, err := s.NewScalarFunction(FunctionType{Params: types, Results: types}, func(c *Caller, stack []uint64) error {
			for i, want := range wantInput {
				if stack[i] != want {
					t.Fatalf("parameter %d: %x != %x", i, stack[i], want)
				}
			}
			expired = c
			retained = append(retained, stack)
			child, err := c.NewScalarFunction(FunctionType{Results: []wasm.ValueType{wasm.I64}}, func(_ *Caller, words []uint64) error {
				if len(words) != 1 || words[0] != 0 {
					t.Fatal("result-only stack", words)
				}
				words[0] = 42
				return nil
			})
			if err != nil {
				return err
			}
			result, err := c.Call(child, nil)
			if err != nil {
				return err
			}
			if bits(t, result.Values[0]) != 42 {
				t.Fatal(result)
			}
			if _, err := c.NewScalarFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
				t.Fatal("nil nested callback", err)
			}
			root, err := c.NewExternRef("temporary scalar callback root")
			if err != nil {
				return err
			}
			if err := c.Collect(); err != nil {
				return err
			}
			if got, err := s.ExternRef(root); err != nil || got != "temporary scalar callback root" {
				t.Fatal(got, err)
			}
			copy(stack, wantOutput)
			stack[0] |= 0xdeadbeef00000000
			stack[2] |= 0xdeadbeef00000000
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := l.Define("h", "scalar", host); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{{Params: types, Results: types}}, Imports: []module.Import{{Module: "h", Name: "scalar", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20, A: 2}, {Op: 0x20, A: 3}, {Op: 0x10}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Index: 1}}}
		fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
		for round := 0; round < 2; round++ {
			result, err := fn.Call(nil, input, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			clear(retained[round])
			for i, want := range wantOutput {
				if got := bits(t, result.Values[i]); got != want {
					t.Fatalf("result %d: %x != %x", i, got, want)
				}
			}
			if _, err := expired.NewScalarFunction(FunctionType{}, func(*Caller, []uint64) error { return nil }); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("retained caller", err)
			}
		}
		if _, err := fn.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrType) {
			t.Fatal("invalid arguments", err)
		}
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		usage, err := s.Usage()
		if err != nil || usage.HostRoots != 0 || usage.ExternRefs != 0 {
			t.Fatal("callback roots leaked", usage, err)
		}
	})
}

func TestScalarHostInvalidSignaturesFailuresAndContinuation(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		ok := ScalarHostFunction(func(*Caller, []uint64) error { return nil })
		if _, err := s.NewScalarFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		for _, typ := range []wasm.ValueType{wasm.V128, wasm.ExternRef, wasm.FuncRef, 0} {
			for _, signature := range []FunctionType{{Params: []wasm.ValueType{typ}}, {Results: []wasm.ValueType{typ}}} {
				if _, err := s.NewScalarFunction(signature, ok); !errors.Is(err, ErrType) {
					t.Fatal("invalid scalar signature", signature, err)
				}
			}
		}
		sentinel := errors.New("scalar host failure")
		for _, panicHost := range []bool{false, true} {
			var expired *Caller
			host, err := s.NewScalarFunction(FunctionType{}, func(c *Caller, stack []uint64) error {
				expired = c
				if len(stack) != 0 {
					t.Fatal(stack)
				}
				if _, err := c.NewExternRef("failed callback"); err != nil {
					return err
				}
				if panicHost {
					panic(sentinel)
				}
				return sentinel
			})
			if err != nil {
				t.Fatal(err)
			}
			func() {
				if panicHost {
					defer func() {
						if recover() != sentinel {
							t.Fatal("changed scalar panic")
						}
					}()
				}
				if _, err := host.Call(nil, nil, CallOptions{}); !errors.Is(err, sentinel) {
					t.Fatal(err)
				}
			}()
			if err := expired.Collect(); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
		}
		host, err := s.NewScalarFunction(FunctionType{}, ok)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := host.Call(ctx, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if _, err := s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			call, err := c.Start(host, nil)
			if err != nil {
				return err
			}
			_, done, err := c.Poll(call)
			if err != nil || !done {
				t.Fatal("scalar continuation", done, err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		usage, err := s.Usage()
		if err != nil || usage.HostRoots != 0 || usage.ExternRefs != 0 {
			t.Fatal(usage, err)
		}
		s.Close()
		if _, err := s.NewScalarFunction(FunctionType{}, ok); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	})
}

func TestScalarHostExceptionPreservesIdentity(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Features = wasm.ExceptionHandling
		e, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		s, err := e.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		tag, err := s.NewTag(FunctionType{Params: []wasm.ValueType{wasm.I32}})
		if err != nil {
			t.Fatal(err)
		}
		host, err := s.NewScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error {
			exception, err := c.NewException(tag, []Value{I32(99)})
			if err != nil {
				return err
			}
			return exception
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = host.Call(nil, nil, CallOptions{})
		var exception *Exception
		if !errors.As(err, &exception) {
			t.Fatal(err)
		}
		defer exception.Release()
		got, payload, err := exception.Payload()
		if err != nil || !tag.SameTag(got) || len(payload) != 1 || bits(t, payload[0]) != 99 {
			t.Fatal(got, payload, err)
		}
	})
}
