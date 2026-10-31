package component

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHostCompletionExecution(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			type request struct {
				p      *HostCompletion
				ctx    context.Context
				n      uint64
				caller *Caller
			}
			requests := make(chan request, 4)
			ready := make(chan struct{}, 4)
			stop := make(chan struct{})
			done := make(chan struct{})
			var cancelWork bool // Accessed by the serialized starter, captured in each request.
			var direct *Function
			var workerErr error
			var mu sync.Mutex
			report := func(err error) {
				if err != nil {
					mu.Lock()
					workerErr = errors.Join(workerErr, err)
					mu.Unlock()
				}
			}
			go func() {
				defer close(done)
				for {
					select {
					case <-stop:
						return
					case <-ready:
						select {
						case <-stop:
							return
						case r := <-requests:
							report(r.p.Resolve(func(c *Caller) (*Value, error) {
								if c == r.caller {
									return nil, vm.ErrType
								}
								if _, err := r.caller.Call(direct, nil); !errors.Is(err, wasm.ErrClosed) {
									return nil, vm.ErrType
								}
								return &Value{Bits: r.n + 7}, nil
							}))
						}
					}
				}
			}()
			defer func() { close(stop); <-done }()
			imports := Imports{
				AsyncCompletions: map[string]AsyncHostFunction{"add": func(c *Caller, args []Value) (*HostCompletion, error) {
					p := NewHostCompletion()
					r := request{p: p, ctx: c.Context(), n: args[0].Bits, caller: c}
					if cancelWork {
						go func() { <-r.ctx.Done(); report(r.p.Cancelled()) }()
					} else {
						requests <- r
					}
					return p, nil
				}},
				Functions: map[string]HostFunction{"notify": func(*Caller, []Value) (*Value, error) {
					if !cancelWork {
						ready <- struct{}{}
					}
					return nil, nil
				}},
			}
			in := asyncInstance(t, backend, "async-host-completion", imports)
			direct, _ = in.Function("direct")
			churn, _ := in.Function("churn")
			for _, n := range []uint64{1, 17, 128} {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				v, err := churn.Call(ctx, []Value{{Bits: n}}, CallOptions{})
				cancel()
				if err != nil || v.Value == nil || v.Value.Bits != n {
					t.Fatal(n, v, err)
				}
				if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal("retained invocation state")
				}
			}
			cancelWork = true
			f, _ := in.Function("cancel")
			for _, async := range []uint64{0, 1} {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				v, err := f.Call(ctx, []Value{{Bits: async}}, CallOptions{})
				cancel()
				want := uint64(4)
				if async != 0 {
					want = uint64(^uint32(0))
				}
				if err != nil || v.Value == nil || (v.Value.Bits != want && !(async != 0 && v.Value.Bits == 4)) {
					t.Fatal(async, v, err)
				}
			}
			mu.Lock()
			err := workerErr
			mu.Unlock()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHostCompletionFailureAndRecovery(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			var mode string
			var last *HostCompletion
			var saved *Caller
			imports := Imports{Functions: map[string]HostFunction{"notify": func(*Caller, []Value) (*Value, error) { return nil, nil }}, AsyncCompletions: map[string]AsyncHostFunction{"add": func(c *Caller, args []Value) (*HostCompletion, error) {
				saved = c
				if mode == "nil" {
					return nil, nil
				}
				if mode == "reuse" {
					return last, nil
				}
				p := NewHostCompletion()
				last = p
				if mode == "cancelled" {
					p.Cancelled()
					return p, nil
				}
				if mode == "pending" {
					return p, nil
				}
				if mode == "start-error" {
					return nil, vm.ErrLink
				}
				if mode == "start-panic" {
					panic("start")
				}
				if err := p.Resolve(func(fresh *Caller) (*Value, error) {
					if fresh == saved {
						return nil, vm.ErrType
					}
					switch mode {
					case "result-error":
						return nil, vm.ErrLink
					case "result-panic":
						panic("resolve")
					case "shape":
						return nil, nil
					case "invalid":
						return &Value{Bits: 1 << 32}, nil
					}
					return &Value{Bits: args[0].Bits + 7}, nil
				}); err != nil {
					return nil, err
				}
				return p, nil
			}}}
			in := asyncInstance(t, backend, "async-host-completion", imports)
			direct, _ := in.Function("direct")
			invoke := func() (err error) {
				defer func() {
					if p := recover(); p != nil {
						err = wasm.ErrTrap
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				defer cancel()
				_, err = direct.Call(ctx, []Value{{Bits: 42}}, CallOptions{})
				return err
			}
			for _, m := range []string{"nil", "cancelled", "start-error", "start-panic", "result-error", "result-panic", "shape", "invalid", "pending", "reuse", "ok"} {
				mode = m
				err := invoke()
				if (m == "ok") != (err == nil) {
					t.Fatal(m, err)
				}
				if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal(m, "retained state")
				}
				if _, err := saved.Call(direct, nil); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal(m, "retained Caller", err)
				}
			}
			if err := last.Resolve(func(*Caller) (*Value, error) { return nil, nil }); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
			mode = "pending"
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			timer := time.AfterFunc(20*time.Millisecond, in.Close)
			_, err := direct.Call(ctx, []Value{{}}, CallOptions{})
			timer.Stop()
			if !errors.Is(err, context.Canceled) && !errors.Is(err, wasm.ErrClosed) {
				t.Fatal("close", err)
			}
			if in.tasks != nil || in.heldValues != 0 {
				t.Fatal("close retained state")
			}
		})
	}
}

func TestHostCompletionSharedPendingHandoff(t *testing.T) {
	p := NewHostCompletion()
	in := asyncInstance(t, "interpreter", "async-host-completion", Imports{
		Functions:        map[string]HostFunction{"notify": func(*Caller, []Value) (*Value, error) { return nil, nil }},
		AsyncCompletions: map[string]AsyncHostFunction{"add": func(*Caller, []Value) (*HostCompletion, error) { return p, nil }},
	})
	multi, _ := in.Function("multi")
	if _, err := multi.Call(nil, []Value{{Bits: 42}}, CallOptions{}); !errors.Is(err, vm.ErrType) {
		t.Fatal("shared pending completion accepted", err)
	}
	if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
		t.Fatal("shared completion retained state")
	}
	if err := p.Cancelled(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal("original invocation not aborted", err)
	}
}

func TestHostCompletionProtocol(t *testing.T) {
	var p *HostCompletion
	if !errors.Is(p.Resolve(nil), vm.ErrType) || !errors.Is(p.Cancelled(), vm.ErrType) {
		t.Fatal("nil completion")
	}
	p = NewHostCompletion()
	if err := p.Resolve(nil); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
	wake := make(chan struct{}, 1)
	if err := p.Cancelled(); err != nil {
		t.Fatal(err)
	}
	if err := p.Resolve(func(*Caller) (*Value, error) { return nil, nil }); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := p.Cancelled(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := p.bind(wake); err != nil {
		t.Fatal(err)
	}
	select {
	case <-wake:
	default:
		t.Fatal("lost pre-bind notification")
	}
	if !p.ready() {
		t.Fatal("not ready")
	}
	if err := p.bind(wake); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
	if _, cancelled, err := p.take(); err != nil || !cancelled {
		t.Fatal(cancelled, err)
	}
	if _, _, err := p.take(); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
	if p.ready() {
		t.Fatal("consumed result ready")
	}
	p.close()
	if err := p.bind(wake); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
}
