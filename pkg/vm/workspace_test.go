package vm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCallWorkspaceReuseAndPolicy(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		m, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		add := export(t, instantiate(t, l, m), "add")
		var workspace *callWorkspace
		for i := range 8 {
			options := CallOptions{Metered: i%2 == 0, Fuel: 3}
			r, err := add.Call(nil, []Value{I32(uint32(i)), I32(1)}, options)
			fuel := uint64(3)
			if options.Metered {
				fuel = 0
			}
			if err != nil || bits(t, r.Values[0]) != uint64(i+1) || r.FuelRemaining != fuel {
				t.Fatal(r, err)
			}
			if s.workspace == nil || workspace != nil && workspace != s.workspace {
				t.Fatal("outer-call workspace was not reused")
			}
			workspace = s.workspace
		}
		if _, err := add.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrType) || s.workspace != workspace {
			t.Fatal("invalid arguments changed the idle workspace", err)
		}
		if _, err := add.Call(nil, []Value{I32(1), I32(2)}, CallOptions{Metered: true, Fuel: 1}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		// Switch functions and slot shapes: every reused local must start at zero.
		d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}},
			Functions: []module.Function{{Locals: []module.Local{{Count: 1, Type: wasm.I32}}, Body: module.Expr{
				{Op: 0x20}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x22}, {Op: 0xb}}},
				{Body: module.Expr{{Op: 0}, {Op: 0xb}}}},
			Exports: []module.Export{{Name: "local"}, {Name: "trap", Index: 1}}}
		in := instantiate(t, l, compileDefinition(t, e, d))
		local, trap := export(t, in, "local"), export(t, in, "trap")
		for range 3 {
			if _, err := trap.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			r, err := local.Call(nil, nil, CallOptions{})
			if err != nil || bits(t, r.Values[0]) != 1 || s.workspace != workspace {
				t.Fatal(r, err)
			}
		}
		s.Close()
		if s.workspace != nil || s.WaitClosed(context.Background()) != nil {
			t.Fatal("closed Store retained its workspace")
		}
	})
}

func TestCallWorkspaceCancellationPanicAndReentry(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		m, err := e.Compile(nil, addBinary)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		add := export(t, instantiate(t, l, m), "add")
		check := func() {
			t.Helper()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			r, err := add.Call(ctx, []Value{I32(17), I32(25)}, CallOptions{Metered: true, Fuel: 3})
			if err != nil || bits(t, r.Values[0]) != 42 {
				t.Fatal(r, err)
			}
		}
		check()
		d := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Body: module.Expr{
			{Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "loop"}}}
		loop := export(t, instantiate(t, l, compileDefinition(t, e, d)), "loop")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		if _, err := loop.Call(ctx, nil, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		check()
		sentinel := errors.New("host panic")
		host, err := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { panic(sentinel) })
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer func() {
				if recover() != sentinel {
					t.Fatal("host panic was not preserved")
				}
			}()
			host.Call(nil, nil, CallOptions{})
		}()
		check()
		workspace := s.workspace
		host, err = s.NewFunction(FunctionType{Results: []wasm.ValueType{wasm.I32}}, func(c *Caller, _ []Value) ([]Value, error) {
			if s.workspace != nil {
				t.Fatal("active outer workspace was cached")
			}
			r, err := c.Call(add, []Value{I32(2), I32(3)})
			if s.workspace != nil {
				t.Fatal("nested workspace escaped to the outer-call cache")
			}
			return r.Values, err
		})
		if err != nil {
			t.Fatal(err)
		}
		r, err := host.Call(nil, nil, CallOptions{})
		if err != nil || bits(t, r.Values[0]) != 5 || s.workspace != workspace {
			t.Fatal(r, err)
		}
		// An idle driver's borrowed raw results must not become Store GC roots.
		ref, err := s.NewExternRef("temporary")
		if err != nil {
			t.Fatal(err)
		}
		echo, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}},
			func(_ *Caller, args []Value) ([]Value, error) { return args, nil })
		if err != nil {
			t.Fatal(err)
		}
		r, err = echo.Call(nil, []Value{ref}, CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		ref.Release()
		r.Values[0].Release()
		if err := s.Collect(); err != nil {
			t.Fatal(err)
		}
		usage, err := s.Usage()
		if err != nil || usage.ExternRefs != 0 || usage.HostRoots != 0 || s.workspace != workspace {
			t.Fatal(usage, err)
		}
	})
}
