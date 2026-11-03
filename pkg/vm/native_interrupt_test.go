package vm

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestOperationInterruptsNestedNativeFrames(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, closeStore := range []bool{false, true} {
		e, s, linker := setupWithEngine(t, func() *Engine {
			config := DefaultEngineConfig()
			config.Backend = "native"
			engine, err := NewEngineWithConfig(config)
			if err != nil {
				t.Fatal(err)
			}
			return engine
		})
		entered := make(chan struct{})
		var retained context.Context
		notify, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			if len(c.op.frames) != 2 {
				t.Fatal("nested native frame was not registered", len(c.op.frames))
			}
			retained = c.Context()
			close(entered)
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		var loop Extern
		reenter, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			_, err := c.Call(loop, nil)
			return nil, err
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := linker.DefineModule("h", map[string]Extern{"notify": notify, "reenter": reenter}); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{{}},
			Imports: []module.Import{{Module: "h", Name: "notify"}, {Module: "h", Name: "reenter"}},
			Functions: []module.Function{
				{Body: module.Expr{{Op: 0x10}, {Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}},
				{Body: module.Expr{{Op: 0x10, A: 1}, {Op: 0xb}}},
			}, Exports: []module.Export{{Name: "loop", Index: 2}, {Name: "outer", Index: 3}}}
		in := instantiate(t, linker, compileDefinition(t, e, d))
		loop, outer := export(t, in, "loop"), export(t, in, "outer")
		parent := context.Background()
		cancel := func() {}
		if !closeStore {
			parent, cancel = context.WithCancel(parent)
		}
		defer cancel()
		done := make(chan error, 1)
		go func() { _, err := outer.Call(parent, nil, CallOptions{}); done <- err }()
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			t.Fatal("nested native call did not reach the host boundary")
		}
		runtime.GC()
		if closeStore {
			s.Close()
		} else {
			cancel()
		}
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatal("nested native call missed cancellation", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("native cancellation failed to drain both frames")
		}
		if !errors.Is(retained.Err(), context.Canceled) || context.Cause(retained) != context.Canceled {
			t.Fatal("retained caller context lost cancellation")
		}
		if s.active != nil {
			t.Fatal("interrupted operation remained active")
		}
		if !closeStore {
			compiled, err := e.Compile(nil, addBinary)
			if err != nil {
				t.Fatal(err)
			}
			defer compiled.Close()
			add := export(t, instantiate(t, linker, compiled), "add")
			r, err := add.Call(nil, []Value{I32(20), I32(22)}, CallOptions{})
			if err != nil || bits(t, r.Values[0]) != 42 {
				t.Fatal("later operation reused interrupted native state", err)
			}
		}
		s.Close()
		if err := s.WaitClosed(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRetiredOperationWatcherCannotInterruptReusedWorkspace(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, linker := setupWithEngine(t, newEngine)
		compiled, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer compiled.Close()
		add := export(t, instantiate(t, linker, compiled), "add")
		var old *operation
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err = s.WithCaller(parent, CallOptions{}, func(c *Caller) error {
			old = c.op
			_, err := c.Call(add, []Value{I32(20), I32(22)})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		workspace := s.workspace
		probe, err := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) {
			// Simulate an old watcher that was already scheduled when stopped.
			// Its original workspace now belongs to this different operation.
			old.interrupt()
			cancel()
			runtime.GC()
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := linker.Define("h", "probe", probe); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "probe"}},
			Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 0xb}}}},
			Exports:   []module.Export{{Name: "run", Index: 1}}}
		fn := export(t, instantiate(t, linker, compileDefinition(t, e, d)), "run")
		ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
		defer stop()
		if _, err := fn.Call(ctx, nil, CallOptions{}); err != nil || s.workspace != workspace {
			t.Fatal("retired watcher reached reused storage", err)
		}
	})
}

func TestHostPanicDrainsConcurrentCapability(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, linker := setupWithEngine(t, newEngine)
		entered := make(chan struct{})
		done := make(chan error, 1)
		notify, err := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) {
			close(entered)
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := linker.Define("h", "notify", notify); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "notify"}},
			Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}}},
			Exports:   []module.Export{{Name: "loop", Index: 1}}}
		loop := export(t, instantiate(t, linker, compileDefinition(t, e, d)), "loop")
		sentinel := errors.New("outer callback panicked")
		host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			go func() { _, err := c.Call(loop, nil); done <- err }()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Error("concurrent capability did not enter")
			}
			panic(sentinel)
		})
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer func() {
				if got := recover(); got != sentinel {
					t.Fatal("panic was not preserved", got)
				}
			}()
			host.Call(nil, nil, CallOptions{})
		}()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatal("concurrent capability was not canceled", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("panic did not drain concurrent capability")
		}
		compiled, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer compiled.Close()
		add := export(t, instantiate(t, linker, compiled), "add")
		if r, err := add.Call(nil, []Value{I32(20), I32(22)}, CallOptions{}); err != nil || bits(t, r.Values[0]) != 42 {
			t.Fatal("panic left the Store busy or interrupted", err)
		}
	})
}
