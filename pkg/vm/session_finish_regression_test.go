package vm

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/core/module"
	"testing"
	"time"
)

func TestSessionReturnDrainsConcurrentCapability(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, mode := range []string{"return", "panic"} {
			t.Run(mode, func(t *testing.T) {
				e, s, linker := setupWithEngine(t, newEngine)
				entered := make(chan struct{})
				worker := make(chan error, 1)
				notify, err := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { close(entered); return nil, nil })
				if err != nil {
					t.Fatal(err)
				}
				if err := linker.Define("h", "notify", notify); err != nil {
					t.Fatal(err)
				}
				d := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Module: "h", Name: "notify"}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x10}, {Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "loop", Index: 1}}}
				loop := export(t, instantiate(t, linker, compileDefinition(t, e, d)), "loop")
				sentinel := errors.New("session callback panic")
				type result struct {
					caught any
					err    error
				}
				done := make(chan result, 1)
				go func() {
					r := result{}
					defer func() { r.caught = recover(); done <- r }()
					_, r.err = s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
						go func() { _, err := c.Call(loop, nil); worker <- err }()
						select {
						case <-entered:
						case <-time.After(2 * time.Second):
							return errors.New("capability did not enter")
						}
						if mode == "panic" {
							panic(sentinel)
						}
						return nil
					})
				}()
				var r result
				select {
				case r = <-done:
				case <-time.After(2 * time.Second):
					// Rescue the blocked baseline so the regression probe does not leak a
					// guest loop or test goroutine after recording the bounded-exit failure.
					s.Close()
					select {
					case <-done:
					case <-time.After(2 * time.Second):
						t.Fatal("Store close did not rescue session")
					}
					t.Fatal("session waited for the capability before canceling it")
				}
				if r.err != nil || mode == "panic" && r.caught != sentinel || mode == "return" && r.caught != nil {
					t.Fatal("session result", r.err, r.caught)
				}
				select {
				case err := <-worker:
					if !errors.Is(err, context.Canceled) {
						t.Fatal("capability result", err)
					}
				case <-time.After(2 * time.Second):
					t.Fatal("capability did not drain")
				}
				compiled, err := e.Compile(nil, addBinary)
				if err != nil {
					t.Fatal(err)
				}
				defer compiled.Close()
				add := export(t, instantiate(t, linker, compiled), "add")
				if r, err := add.Call(nil, []Value{I32(20), I32(22)}, CallOptions{}); err != nil || bits(t, r.Values[0]) != 42 {
					t.Fatal("session left Store busy or interrupted", err)
				}
			})
		}
	})
}
