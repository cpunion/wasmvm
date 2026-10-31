package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestContinuationGuestFramesAndOutOfOrderResume(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		ready := false
		var expired []*Caller
		host, err := s.NewSuspendableFunction(FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, func(c *Caller, args []Value) (HostResult, error) {
			expired = append(expired, c)
			n := bits(t, args[0])
			return HostResult{Pending: &HostSuspension{Ready: func() bool { return ready }, Resume: func(current *Caller) (HostResult, error) {
				if current == c {
					t.Fatal("old caller reused")
				}
				if err := c.Collect(); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("caller did not expire", err)
				}
				if err := current.Collect(); err != nil {
					return HostResult{}, err
				}
				return HostResult{Values: []Value{I32(uint32(n + 3))}}, nil
			}}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := l.Define("h", "pause", host); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}}, Imports: []module.Import{{Module: "h", Name: "pause", Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x10}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Kind: 0, Index: 1}}}
		fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
		var saved *Continuation
		fuel, err := s.WithCaller(nil, CallOptions{Metered: true, Fuel: 1000}, func(c *Caller) error {
			a, err := c.Start(fn, []Value{I32(10)})
			if err != nil {
				return err
			}
			saved = a
			b, err := c.Start(fn, []Value{I32(20)})
			if err != nil {
				return err
			}
			for _, call := range []*Continuation{a, b} {
				if _, done, err := c.Poll(call); err != nil || done || call.Ready() {
					t.Fatal(done, err)
				}
			}
			if len(c.op.frames) != 2 {
				t.Fatal("parked frames lost")
			}
			if err := c.Collect(); err != nil {
				return err
			}
			ready = true
			for i, call := range []*Continuation{b, a} {
				for {
					r, done, err := c.Poll(call)
					if err != nil {
						return err
					}
					if done {
						want := uint64(27)
						if i == 1 {
							want = 17
						}
						if bits(t, r.Values[0]) != want {
							t.Fatal(r)
						}
						break
					}
				}
				if call.Ready() {
					t.Fatal("completed continuation ready")
				}
				if _, _, err := c.Poll(call); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal(err)
				}
			}
			if len(c.op.frames) != 0 || len(c.op.continuations) != 0 {
				t.Fatal("completed frames retained")
			}
			return nil
		})
		if err != nil || fuel >= 1000 || len(expired) != 4 {
			t.Fatal(fuel, err, len(expired))
		}
		_, err = s.WithCaller(nil, CallOptions{}, func(c *Caller) error { _, _, err := c.Poll(saved); return err })
		if !errors.Is(err, ErrReference) {
			t.Fatal("session capability escaped", err)
		}
	})
}

func TestContinuationRootsAndAbandonment(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		for _, complete := range []bool{false, true} {
			ready := false
			var captured Value
			var result Value
			host, err := s.NewSuspendableFunction(FunctionType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}, func(c *Caller, args []Value) (HostResult, error) {
				captured = args[0]
				temporary, err := c.NewExternRef("output")
				if err != nil {
					return HostResult{}, err
				}
				return HostResult{Pending: &HostSuspension{Ready: func() bool { return ready }, Resume: func(c *Caller) (HostResult, error) {
					if obj, err := s.ExternRef(captured); err != nil || obj != "input" {
						t.Fatal(obj, err)
					}
					return HostResult{Values: []Value{temporary}}, nil
				}}}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
				input, err := c.NewExternRef("input")
				if err != nil {
					return err
				}
				call, err := c.Start(host, []Value{input})
				if err != nil {
					return err
				}
				input.Release()
				if _, done, err := c.Poll(call); err != nil || done {
					t.Fatal(done, err)
				}
				if err := c.Collect(); err != nil {
					return err
				}
				if !complete {
					return nil
				}
				ready = true
				r, done, err := c.Poll(call)
				if err != nil || !done {
					t.Fatal(done, err)
				}
				result = r.Values[0]
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if complete {
				if obj, err := s.ExternRef(result); err != nil || obj != "output" {
					t.Fatal(obj, err)
				}
				result.Release()
			}
			if _, err := s.ExternRef(captured); !errors.Is(err, ErrReference) {
				t.Fatal("host argument remained rooted", err)
			}
			usage, err := s.Usage()
			if err != nil || usage.HostRoots != 0 {
				t.Fatal("continuation root leak", usage, err)
			}
		}
	})
}

func TestContinuationCancelAndHostFailures(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		for _, mode := range []string{"abort", "cancel", "error", "shape", "panic"} {
			ctx, cancel := context.WithCancel(context.Background())
			var call *Continuation
			var caught any
			func() {
				defer func() { caught = recover() }()
				_, err := s.WithCaller(ctx, CallOptions{}, func(c *Caller) error {
					fn, err := c.NewSuspendableFunction(FunctionType{}, func(*Caller, []Value) (HostResult, error) {
						return HostResult{Pending: &HostSuspension{Ready: func() bool { return true }, Resume: func(*Caller) (HostResult, error) {
							switch mode {
							case "error":
								return HostResult{}, ErrLink
							case "shape":
								return HostResult{Values: []Value{I32(1)}}, nil
							case "panic":
								panic("resumed host")
							}
							return HostResult{}, nil
						}}}, nil
					})
					if err != nil {
						return err
					}
					call, err = c.Start(fn, nil)
					if err != nil {
						return err
					}
					if _, done, err := c.Poll(call); err != nil || done {
						t.Fatal(done, err)
					}
					if mode == "abort" {
						return c.Abort(call)
					}
					if mode == "cancel" {
						cancel()
						return c.Abort(call)
					}
					_, _, err = c.Poll(call)
					return err
				})
				if mode == "error" && !errors.Is(err, ErrLink) || mode == "shape" && !errors.Is(err, ErrType) || mode == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatal(mode, err)
				}
			}()
			cancel()
			if mode == "panic" && caught != "resumed host" {
				t.Fatal(caught)
			}
			if !call.closed {
				t.Fatal("parked call survived session")
			}
		}
	})
}
