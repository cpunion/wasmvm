package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarLeafPreservesLinkedContinuationAndCancellation(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, linker := setupWithEngine(t, newEngine)
		signature := FunctionType{Params: []wasm.ValueType{wasm.I32, wasm.I64}, Results: []wasm.ValueType{wasm.I32}}
		var leaf Extern
		var calls, checkpoints int
		var cancel context.CancelFunc
		_, err := s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			var err error
			leaf, err = c.NewScalarLeafFunction(signature, state.NewScalarLeaf(func(a, b uint64) (uint64, error) {
				calls++
				if b != 0xfeed00000005 {
					t.Fatal("full-width operand", b)
				}
				switch a {
				case 3:
					return 0, wasm.ErrTrap
				case 4:
					cancel()
				}
				return 1<<48 | (a + b&255), nil
			}))
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		checkType := FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
		check, err := s.NewFunction(checkType, func(_ *Caller, args []Value) ([]Value, error) {
			checkpoints++
			if got := bits(t, args[0]); got != 7 {
				t.Fatal("leaf result in ordinary host continuation", got)
			}
			return []Value{I32(11)}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		for name, fn := range map[string]Extern{"leaf": leaf, "check": check} {
			if err := linker.Define("test", name, fn); err != nil {
				t.Fatal(err)
			}
		}
		d := &module.Module{
			Types:   []module.FuncType{module.FuncType(signature), module.FuncType(checkType)},
			Imports: []module.Import{{Module: "test", Name: "leaf"}, {Module: "test", Name: "check", Type: 1}},
			Functions: []module.Function{
				{Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x10}, {Op: 0x10, A: 1}, {Op: 0xb}}},
				{Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x10, A: 2}, {Op: 0x41, Bits: 13}, {Op: 0x6a}, {Op: 0xb}}},
			},
			Exports: []module.Export{{Name: "run", Index: 3}},
		}
		run := export(t, instantiate(t, linker, compileDefinition(t, e, d)), "run")
		args := []Value{I32(2), I64(0xfeed00000005)}
		result, err := leaf.Call(nil, args, CallOptions{})
		if err != nil || bits(t, result.Values[0]) != 7 {
			t.Fatal("root leaf result", result, err)
		}
		for _, opts := range []CallOptions{{}, {Metered: true, Fuel: 100}} {
			result, err := run.Call(nil, args, opts)
			if err != nil || bits(t, result.Values[0]) != 24 {
				t.Fatal("linked return", result, err)
			}
		}
		args[0] = I32(3)
		if _, err := run.Call(nil, args, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("leaf trap", err)
		}
		ctx, stop := context.WithCancel(context.Background())
		cancel = stop
		defer stop()
		args[0] = I32(4)
		if _, err := run.Call(ctx, args, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal("canceled leaf resumed the guest", err)
		}
		if calls != 5 || checkpoints != 2 {
			t.Fatal("unexpected leaf or guest effects", calls, checkpoints)
		}
	})
}
