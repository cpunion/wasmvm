package vm

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarCallNumericBitsOverlappingResultsAndReuse(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, _, l := setupWithEngine(t, newEngine)
		params := []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64}
		results := []wasm.ValueType{wasm.F64, wasm.F32, wasm.I64, wasm.I32}
		d := &module.Module{Types: []module.FuncType{{Params: params, Results: results}}, Functions: []module.Function{{Body: module.Expr{
			{Op: 0x20, A: 3}, {Op: 0x20, A: 2}, {Op: 0x20, A: 1}, {Op: 0x20}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "reverse"}}}
		fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "reverse")
		call, err := fn.Scalar()
		if err != nil {
			t.Fatal(err)
		}
		for _, floating := range []uint64{0x8000000000000000, 0xfff800000000babe} {
			for round := range 3 {
				stack := []uint64{0xdeadbeefffffffff, uint64(round), 0xdeadbeef80000000, floating}
				want := []uint64{floating, 0x80000000, uint64(round), 0xffffffff}
				if _, err := call.Call(nil, stack, CallOptions{}); err != nil || !slices.Equal(stack, want) {
					t.Fatal(stack, want, err)
				}
			}
		}
	})
}

func TestScalarCallFuelFailuresAndUnmodifiedInput(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		m, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		call, err := export(t, instantiate(t, l, m), "add").Scalar()
		if err != nil {
			t.Fatal(err)
		}
		stack := []uint64{17, 25}
		if fuel, err := call.Call(nil, stack, CallOptions{Metered: true, Fuel: 3}); err != nil || fuel != 0 || !slices.Equal(stack, []uint64{42, 25}) {
			t.Fatal(fuel, stack, err)
		}
		copy(stack, []uint64{17, 25})
		if _, err := call.Call(nil, stack, CallOptions{Metered: true, Fuel: 2}); !errors.Is(err, wasm.ErrLimit) || !slices.Equal(stack, []uint64{17, 25}) {
			t.Fatal("fuel failure changed caller input", stack, err)
		}
		for _, bad := range [][]uint64{nil, {1}, {1, 2, 3}} {
			if _, err := call.Call(nil, bad, CallOptions{}); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := call.Call(ctx, stack, CallOptions{}); !errors.Is(err, context.Canceled) || !slices.Equal(stack, []uint64{17, 25}) {
			t.Fatal(stack, err)
		}
		if _, err := call.Call(ctx, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation must precede argument validation", err)
		}
		trap := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{{Body: module.Expr{{Op: 0}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "trap"}}}
		trapping, err := export(t, instantiate(t, l, compileDefinition(t, e, trap)), "trap").Scalar()
		if err != nil {
			t.Fatal(err)
		}
		output := []uint64{99}
		if _, err := trapping.Call(nil, output, CallOptions{}); !errors.Is(err, wasm.ErrTrap) || output[0] != 99 {
			t.Fatal(output, err)
		}
		s.Close()
		if _, err := call.Call(nil, stack, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if _, err := call.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("Store closure must precede argument validation", err)
		}
	})
}

func TestScalarCallInvalidBindingsAndHostPanicCleanup(t *testing.T) {
	var zero ScalarFunc
	for _, fn := range []*ScalarFunc{nil, &zero} {
		if _, err := fn.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
	}
	if _, err := (Extern{}).Scalar(); !errors.Is(err, ErrReference) {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		for _, typ := range []wasm.ValueType{wasm.V128, wasm.FuncRef, wasm.ExternRef} {
			for _, signature := range []FunctionType{{Params: []wasm.ValueType{typ}}, {Results: []wasm.ValueType{typ}}} {
				host, err := s.NewFunction(signature, func(*Caller, []Value) ([]Value, error) { return nil, nil })
				if err != nil {
					t.Fatal(err)
				}
				if _, err := host.Scalar(); !errors.Is(err, ErrType) {
					t.Fatal(err)
				}
			}
		}
		sentinel := errors.New("scalar invocation panic")
		first := true
		host, err := s.NewScalarFunction(FunctionType{}, func(*Caller, []uint64) error {
			if first {
				first = false
				panic(sentinel)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		call, err := host.Scalar()
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer func() {
				if recover() != sentinel {
					t.Fatal("host panic changed")
				}
			}()
			call.Call(nil, nil, CallOptions{})
		}()
		if _, err := call.Call(nil, nil, CallOptions{}); err != nil {
			t.Fatal("panic retained binding/Store admission", err)
		}
		if _, err := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, nil }); err != nil {
			t.Fatal(err)
		}
	})
}

func TestScalarCallBusyStoreReentryAndRetainedContext(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		child, err := s.NewScalarFunction(FunctionType{Results: []wasm.ValueType{wasm.I64}}, func(_ *Caller, stack []uint64) error { stack[0] = 42; return nil })
		if err != nil {
			t.Fatal(err)
		}
		entered, release := make(chan struct{}), make(chan struct{})
		defer func() {
			select {
			case <-release:
			default:
				close(release)
			}
		}()
		var retained context.Context
		host, err := s.NewScalarFunction(FunctionType{Results: []wasm.ValueType{wasm.I64}}, func(c *Caller, stack []uint64) error {
			retained = c.Context()
			close(entered)
			<-release
			r, err := c.Call(child, nil)
			if err == nil {
				stack[0] = bits(t, r.Values[0])
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		call, err := host.Scalar()
		if err != nil {
			t.Fatal(err)
		}
		other, err := child.Scalar()
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		stack := []uint64{0}
		go func() { _, err := call.Call(nil, stack, CallOptions{}); done <- err }()
		<-entered
		for _, busy := range []*ScalarFunc{call, other} {
			buffer := []uint64{99}
			if _, err := busy.Call(nil, buffer, CallOptions{}); !errors.Is(err, ErrBusy) || buffer[0] != 99 {
				t.Fatal(buffer, err)
			}
		}
		close(release)
		if err := <-done; err != nil || stack[0] != 42 || retained.Err() != context.Canceled {
			t.Fatal(stack, err, retained.Err())
		}
		if _, err := other.Call(nil, stack, CallOptions{}); err != nil || retained.Err() != context.Canceled {
			t.Fatal("old public context revived", err)
		}
	})
}
