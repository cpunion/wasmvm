package vm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGuestLoopCancellationAndStartPanicRollback(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		d := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Body: module.Expr{{Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "loop", Kind: 0}}}
		loop := export(t, instantiate(t, l, compileDefinition(t, e, d)), "loop")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		if _, err := loop.Call(ctx, nil, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		sentinel := errors.New("start panic")
		host, _ := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { panic(sentinel) })
		l.Define("h", "panic", host)
		d = &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "panic", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 0xb}}}}, Memories: []module.Limits{{Min: 1}}, HasStart: true, Start: 1}
		compiled := compileDefinition(t, e, d)
		before, _ := s.Usage()
		func() {
			defer func() {
				if recover() != sentinel {
					t.Fatal("changed start panic")
				}
			}()
			l.Instantiate(nil, compiled, CallOptions{})
		}()
		after, err := s.Usage()
		if err != nil || before != after {
			t.Fatal(before, after, err)
		}
	})
}

func TestExecutionReuseAndClose(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, _, l := setupWithEngine(t, newEngine)
		m, _ := e.Compile(nil, addBinary)
		fn := export(t, instantiate(t, l, m), "add")
		exec, err := NewExecution(fn, CallOptions{Metered: true, Fuel: 3})
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			r, err := exec.Run(nil, []Value{I32(uint32(i)), I32(1)})
			if err != nil || bits(t, r.Values[0]) != uint64(i+1) {
				t.Fatal(r, err)
			}
		}
		exec.Close()
		exec.Close()
		if _, err := exec.Run(nil, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		if _, err := NewExecution(Extern{}, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
		_, s, _ := setupWithEngine(t, newEngine)
		entered, leave, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		host, _ := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { close(entered); <-leave; return nil, nil })
		running, _ := NewExecution(host, CallOptions{})
		go func() { _, err := running.Run(nil, nil); done <- err }()
		<-entered
		if _, err := running.Run(nil, nil); !errors.Is(err, ErrBusy) {
			t.Fatal(err)
		}
		running.Close()
		close(leave)
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if _, err := s.NewMemory(MemoryType{}); err != nil {
			t.Fatal("execution close closed Store", err)
		}
	})
}

func TestSharedReentryLimitsAndErrors(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, quota := range []string{"fuel", "frames", "slots", "reentry"} {
			t.Run(quota, func(t *testing.T) {
				e := newEngine()
				config := DefaultStoreConfig()
				switch quota {
				case "frames":
					config.Execution.MaxFrames = 1
				case "slots":
					config.Execution.MaxSlots = 3
				case "reentry":
					config.MaxReentry = 0
				}
				s, _ := e.NewStore(config)
				defer s.Close()
				l := NewLinker(s)
				m, _ := e.Compile(nil, addBinary)
				add := export(t, instantiate(t, l, m), "add")
				ref, _ := add.Reference()
				defer ref.Release()
				host, _ := s.NewFunction(FunctionType{Results: []wasm.ValueType{wasm.I32}}, func(c *Caller, _ []Value) ([]Value, error) {
					if _, err := c.CallRef(I32(1), nil); !errors.Is(err, ErrType) {
						t.Fatal(err)
					}
					r, err := c.CallRef(ref, []Value{I32(1), I32(2)})
					return r.Values, err
				})
				l.Define("h", "f", host)
				d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Imports: []module.Import{{Module: "h", Name: "f", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Kind: 0, Index: 1}}}
				fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
				if _, err := fn.Call(nil, nil, CallOptions{Metered: true, Fuel: 3}); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(quota, err)
				}
				if quota == "fuel" {
					r, err := fn.Call(nil, nil, CallOptions{Metered: true, Fuel: 4})
					if err != nil || r.FuelRemaining != 0 || bits(t, r.Values[0]) != 3 {
						t.Fatal(r, err)
					}
				}
			})
		}
		_, s, _ := setupWithEngine(t, newEngine)
		sentinel := errors.New("host error")
		bad, _ := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, sentinel })
		if _, err := bad.Call(nil, nil, CallOptions{}); err != sentinel {
			t.Fatal(err)
		}
		wrong, _ := s.NewFunction(FunctionType{Results: []wasm.ValueType{wasm.I32}}, func(*Caller, []Value) ([]Value, error) { return []Value{I64(1)}, nil })
		if _, err := wrong.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		_, foreign, _ := setupWithEngine(t, newEngine)
		foreignRef, _ := foreign.NewExternRef(1)
		wrongRef, _ := s.NewFunction(FunctionType{Results: []wasm.ValueType{wasm.ExternRef}}, func(*Caller, []Value) ([]Value, error) { return []Value{foreignRef}, nil })
		if _, err := wrongRef.Call(nil, nil, CallOptions{}); !errors.Is(err, ErrReference) {
			t.Fatal(err)
		}
	})
}

func TestLeafCallsPreserveLogicalQuotasAndExactFuel(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		d := &module.Module{
			Types: []module.FuncType{
				{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}},
				{Results: []wasm.ValueType{wasm.I32}},
			},
			Functions: []module.Function{
				{Body: module.Expr{{Op: 0x20}, {Op: 4, Block: -1}, {Op: 0x10, A: 1}, {Op: 5}, {Op: 0x41, Bits: 7}, {Op: 0xb}, {Op: 0xb}}},
				{Type: 1, Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}},
			},
			Exports: []module.Export{{Name: "run", Kind: 0}},
		}
		for _, quota := range []string{"enough", "frames", "slots"} {
			t.Run(quota, func(t *testing.T) {
				e := newEngine()
				defer e.Close()
				m := compileDefinition(t, e, d)
				config := DefaultStoreConfig()
				if quota == "frames" {
					config.Execution.MaxFrames = 1
				} else if quota == "slots" {
					config.Execution.MaxSlots = uint32(len(m.code.Functions[0].Slots))
				}
				s, err := e.NewStore(config)
				if err != nil {
					t.Fatal(err)
				}
				defer s.Close()
				f := export(t, instantiate(t, NewLinker(s), m), "run")
				// An untaken call must not fail an eagerly checked frame quota.
				r, err := f.Call(nil, []Value{I32(0)}, CallOptions{})
				if err != nil || bits(t, r.Values[0]) != 7 {
					t.Fatal(r, err)
				}
				r, err = f.Call(nil, []Value{I32(1)}, CallOptions{})
				if quota != "enough" {
					if !errors.Is(err, wasm.ErrLimit) {
						t.Fatal("logical frame quota bypassed", r, err)
					}
					return
				}
				if err != nil || bits(t, r.Values[0]) != 42 {
					t.Fatal(r, err)
				}
				for fuel := uint64(0); fuel <= 5; fuel++ {
					r, err = f.Call(nil, []Value{I32(1)}, CallOptions{Metered: true, Fuel: fuel})
					if fuel < 4 {
						if !errors.Is(err, wasm.ErrLimit) {
							t.Fatal("fuel boundary", fuel, r, err)
						}
					} else if err != nil || bits(t, r.Values[0]) != 42 || r.FuelRemaining != fuel-4 {
						t.Fatal("fuel accounting", fuel, r, err)
					}
				}
			})
		}
	})
}

func TestConcurrentCallerUseAndCanceledCapability(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		entered, leave, childDone := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		child, _ := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { close(entered); <-leave; return nil, nil })
		outer, _ := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			go func() { _, err := c.Call(child, nil); childDone <- err }()
			<-entered
			if _, err := c.Call(child, nil); !errors.Is(err, ErrBusy) {
				t.Fatal(err)
			}
			if err := c.Collect(); !errors.Is(err, ErrBusy) {
				t.Fatal(err)
			}
			close(leave)
			if err := <-childDone; err != nil {
				t.Fatal(err)
			}
			s.Close()
			if _, err := c.Call(child, nil); !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			return nil, nil
		})
		if _, err := outer.Call(nil, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})
}

func TestHostArgumentQuotaAndFunctionRefRoot(t *testing.T) {
	e := NewEngine()
	config := DefaultStoreConfig()
	config.MaxHostRoots = 1
	s, _ := e.NewStore(config)
	defer s.Close()
	fn, _ := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef}}, func(*Caller, []Value) ([]Value, error) { t.Fatal("callback received unrooted input"); return nil, nil })
	ref, _ := s.NewExternRef(1)
	if _, err := fn.Call(nil, []Value{ref}, CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	ref.Release()
	s.Collect()
	host, _ := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { return nil, nil })
	root, _ := host.Reference()
	s.Collect()
	if _, err := s.CallRef(nil, root, nil, CallOptions{}); err != nil {
		t.Fatal(err)
	}
	root.Release()
}
