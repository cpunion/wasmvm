package vm

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPrivateOperationReuseResetsFuelAndFailureState(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		m, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		fn := export(t, instantiate(t, l, m), "add")
		call, err := fn.Scalar()
		if err != nil {
			t.Fatal(err)
		}
		var previous *operation
		for range 4 {
			for _, options := range []CallOptions{{Metered: true, Fuel: 2}, {Metered: true, Fuel: 9}, {}} {
				stack := []uint64{17, 25}
				fuel, err := call.Call(nil, stack, options)
				if options.Metered && options.Fuel == 2 {
					if !errors.Is(err, wasm.ErrLimit) || stack[0] != 17 {
						t.Fatal("failed invocation changed input", stack, err)
					}
				} else if err != nil || stack[0] != 42 || (options.Metered && fuel != 6) {
					t.Fatal("previous invocation affected fuel/results", stack, fuel, err)
				}
				op := s.idleOperation
				if op == nil || (previous != nil && op != previous) || s.active != nil || len(op.frames) != 0 || op.depth != 0 || len(op.pending) != 0 {
					t.Fatal("private execution did not quiesce for reuse")
				}
				previous = op
			}
		}
		s.Close()
		if s.idleOperation != nil {
			t.Fatal("closed Store retained private execution")
		}
	})
}

func TestPrivateOperationObservedStateIsNeverRebound(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		for _, observe := range []func(*operation){
			func(op *operation) { op.publicContext() },
			func(op *operation) { op.ctx.Done() },
			func(op *operation) { op.ctx.Value(struct{}{}) },
			func(op *operation) { op.newCaller(nil) },
		} {
			op, err := s.begin(nil, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			observe(op)
			op.finish()
			next, err := s.begin(nil, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if next == op || op.ctx.Err() != context.Canceled || op.publicContext().Err() != context.Canceled {
				t.Fatal("observed cancellation state revived")
			}
			next.finish()
		}
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		for _, ctx := range []context.Context{parent, context.WithValue(context.Background(), struct{}{}, 42)} {
			op, err := s.begin(ctx, CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			op.finish()
			if s.idleOperation != nil || op.ctx.Err() != context.Canceled || ctx.Err() != nil {
				t.Fatal("custom parent entered cache or was canceled")
			}
		}
	})
}

func TestPrivateOperationRetainedCallerObservesOriginalLifetime(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		var retained *Caller
		host, err := s.NewScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error {
			retained = c // Deliberately do not request Context while active.
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := host.Call(nil, nil, CallOptions{}); err != nil {
			t.Fatal(err)
		}
		old := retained.op
		m, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		add := export(t, instantiate(t, l, m), "add")
		for range 4 {
			if _, err := add.Call(nil, []Value{I32(20), I32(22)}, CallOptions{}); err != nil {
				t.Fatal(err)
			}
			if s.idleOperation == old || retained.Context().Err() != context.Canceled || retained.Store() != s {
				t.Fatal("retained callback acquired a later operation")
			}
			if _, err := retained.Call(add, nil); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("retained callback revived", err)
			}
		}
	})
}

func TestPrivateOperationCloseInterruptsReusedExecution(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		d := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{
			{Body: module.Expr{{Op: 0xb}}},
			{Body: module.Expr{{Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}},
		}, Exports: []module.Export{{Name: "warm"}, {Name: "loop", Index: 1}}}
		in := instantiate(t, l, compileDefinition(t, e, d))
		warm, loop := export(t, in, "warm"), export(t, in, "loop")
		if _, err := warm.Call(nil, nil, CallOptions{}); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { _, err := loop.Call(nil, nil, CallOptions{}); done <- err }()
		deadline := time.Now().Add(5 * time.Second)
		for {
			s.mu.Lock()
			started := s.active != nil
			s.mu.Unlock()
			if started {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("execution did not start")
			}
			runtime.Gosched()
		}
		s.Close()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) || s.idleOperation != nil {
				t.Fatal("close missed private execution", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("private execution did not stop")
		}
		if err := s.WaitClosed(context.Background()); err != nil {
			t.Fatal(err)
		}
	})
}
