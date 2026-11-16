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
	if !p.ready() {
		t.Fatal("publication is not ready")
	}
	if _, cancelled, ready, err := p.poll(wake, true); err != nil || !ready || !cancelled {
		t.Fatal("lost pre-bind cancellation", cancelled, ready, err)
	}
	select {
	case <-wake:
		t.Fatal("ready completion sent a redundant wake")
	default:
	}
	if _, _, _, err := p.poll(wake, true); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}
	if _, _, ready, err := p.poll(nil, false); err != nil || ready {
		t.Fatal(ready, err)
	}
	if p.ready() {
		t.Fatal("consumed result remains ready")
	}
	p.close()
	if _, _, _, err := p.poll(wake, true); !errors.Is(err, vm.ErrType) {
		t.Fatal(err)
	}

}

func TestHostCompletionConcurrentBindAndPublish(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		for range 32 {
			p := NewHostCompletion()
			wake := make(chan struct{}, 1)
			start := make(chan struct{})
			published := make(chan error, 1)
			go func() {
				<-start
				if cancelled {
					published <- p.Cancelled()
				} else {
					published <- p.Resolve(func(*Caller) (*Value, error) { return nil, nil })
				}
			}()
			close(start)
			resolve, gotCancelled, ready, err := p.poll(wake, true)
			if err != nil {
				t.Fatal(err)
			}
			if !ready {
				timer := time.NewTimer(3 * time.Second)
				select {
				case <-wake:
					timer.Stop()
				case <-timer.C:
					t.Fatal("publication was neither ready nor notified")
				}
			}
			if err = <-published; err != nil {
				t.Fatal(err)
			}
			if !ready {
				resolve, gotCancelled, ready, err = p.poll(nil, false)
			}
			if err != nil || gotCancelled != cancelled || (resolve == nil) != cancelled {
				t.Fatal("completion changed across subscription", gotCancelled, err)
			}
			p.close()
		}
	}
}

func TestHostCompletionPollReadinessAndCleanup(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		for _, immediate := range []bool{false, true} {
			for range 32 {
				cleaned := 0
				p := NewHostCompletionWithCleanup(func() { cleaned++ })
				wake := make(chan struct{}, 1)
				publish := func() error {
					if cancelled {
						return p.Cancelled()
					}
					return p.Resolve(func(*Caller) (*Value, error) { return &Value{Bits: 42}, nil })
				}
				if immediate {
					if err := publish(); err != nil || !p.ready() {
						t.Fatal("publication was not visible", err)
					}
				}
				resolve, gotCancelled, ready, err := p.poll(wake, true)
				if err != nil || ready != immediate {
					t.Fatal(ready, err)
				}
				if !immediate {
					done := make(chan error, 1)
					go func() { done <- publish() }()
					select {
					case <-wake:
					case <-time.After(time.Second):
						t.Fatal("pending publication lost its wake")
					}
					if err := <-done; err != nil || !p.ready() {
						t.Fatal("pending result was not visible", err)
					}
					resolve, gotCancelled, ready, err = p.poll(nil, false)
				} else {
					select {
					case <-wake:
						t.Fatal("immediate result sent an unnecessary wake")
					default:
					}
				}
				if err != nil || !ready || gotCancelled != cancelled || (resolve == nil) != cancelled || p.ready() {
					t.Fatal("poll lost or retained a handoff", gotCancelled, ready, err)
				}
				if resolve != nil {
					value, err := resolve(nil)
					if err != nil || value.Bits != 42 {
						t.Fatal(value, err)
					}
				}
				if _, _, ready, err := p.poll(nil, false); err != nil || ready {
					t.Fatal("consumed completion became ready", ready, err)
				}
				if _, _, _, err := p.poll(wake, true); !errors.Is(err, vm.ErrType) {
					t.Fatal("completion admitted twice", err)
				}
				p.close()
				p.close()
				if cleaned != 1 || p.ready() {
					t.Fatal("completion cleanup or readiness survived close", cleaned)
				}
			}
		}
	}
	p := NewHostCompletion()
	p.Resolve(func(*Caller) (*Value, error) { return nil, nil })
	if _, _, _, err := p.poll(nil, false); !errors.Is(err, vm.ErrType) {
		t.Fatal("unbound completion consumed", err)
	}
	p.close()
	if p.ready() {
		t.Fatal("closed unconsumed completion remains ready")
	}
}
