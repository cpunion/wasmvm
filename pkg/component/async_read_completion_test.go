package component

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAwaitReadSiblingProgressAndIndependentCancellation(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			in, typ := sourceType(t, backend, "host-source-stream-u8")
			operation, cancel := context.WithCancel(context.Background())
			defer cancel()
			var completed, closed atomic.Uint32
			ctx, stop := context.WithTimeout(context.Background(), 3*time.Second)
			defer stop()
			_, err := in.WithCaller(ctx, CallOptions{}, func(c *Caller) error {
				for _, wait := range []bool{true, false} {
					source, err := c.NewSource(typ, AsyncSource{Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
						p := NewReadCompletion()
						if wait {
							ctx := c.Context()
							go func() { <-ctx.Done(); p.Cancelled() }()
						} else {
							p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "*", Closed: true}, nil })
						}
						return p, nil
					}, Close: func() { closed.Add(1) }})
					if err != nil {
						return err
					}
					if err = c.Schedule(AsyncWork{Owned: []*Resource{source.Resource}, Start: func(c *Caller, refs []*Resource) (*HostCompletion, error) {
						readContext := context.Background()
						if wait {
							readContext = operation
						}
						return c.AwaitRead(readContext, refs[0], 1, func(fresh *Caller, r ReadResult, err error) (*Value, error) {
							if wait {
								if !errors.Is(err, context.Canceled) || fresh.Context().Err() != nil {
									t.Error(r, err)
								}
							} else {
								if err != nil || r.Bytes != "*" || !r.Closed {
									t.Error(r, err)
								}
								cancel()
							}
							completed.Add(1)
							return nil, nil
						})
					}}); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil || completed.Load() != 2 || closed.Load() != 2 || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal(err, completed.Load(), closed.Load(), in.heldValues, in.heldBytes)
			}
		})
	}
}
func TestAwaitReadValidationAndStarterRollback(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	closed := 0
	_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
		v, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }, Close: func() { closed++ }})
		if err != nil {
			return err
		}
		consume := func(*Caller, ReadResult, error) (*Value, error) { t.Error("abandoned read resolved"); return nil, nil }
		for _, limit := range []uint32{1 << 28} {
			if p, err := c.AwaitRead(nil, v.Resource, limit, consume); p != nil || !errors.Is(err, wasm.ErrTrap) {
				t.Error(p, err)
			}
		}
		if p, err := c.AwaitRead(nil, v.Resource, 1, nil); p != nil || !errors.Is(err, vm.ErrType) {
			t.Error(p, err)
		}
		if p, err := c.AwaitRead(nil, nil, 1, consume); p != nil || err == nil {
			t.Error(p, err)
		}
		p, err := c.AwaitRead(nil, v.Resource, 1, consume)
		if p == nil || err != nil {
			return err
		}
		if p, err := c.AwaitRead(nil, v.Resource, 1, consume); p != nil || !errors.Is(err, wasm.ErrTrap) {
			t.Error(p, err)
		}
		return vm.ErrLink
	})
	if !errors.Is(err, vm.ErrLink) || closed != 1 {
		t.Fatal(err, closed)
	}
	if _, err := new(Caller).AwaitRead(nil, nil, 1, nil); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestAwaitReadWorkCancellationWakesOperation(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	signal := NewCancellation()
	started := make(chan struct{})
	finished := make(chan struct{})
	go func() { <-started; signal.Cancel() }()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := in.WithCaller(ctx, CallOptions{}, func(c *Caller) error {
		v, err := c.NewSource(typ, AsyncSource{Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
			p := NewReadCompletion()
			ctx := c.Context()
			close(started)
			go func() { <-ctx.Done(); p.Cancelled(); close(finished) }()
			return p, nil
		}})
		if err != nil {
			return err
		}
		return c.Schedule(AsyncWork{Owned: []*Resource{v.Resource}, Cancel: signal, Start: func(c *Caller, refs []*Resource) (*HostCompletion, error) {
			p, err := c.AwaitRead(context.Background(), refs[0], 1, func(_ *Caller, _ ReadResult, err error) (*Value, error) {
				if !errors.Is(err, context.Canceled) {
					t.Error(err)
				}
				return nil, nil
			})
			return p, err
		}})
	})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	case <-ctx.Done():
		t.Fatal("source worker stranded")
	}
}

func TestAwaitReadReadyEventChecksCancellationBeforeHandoff(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, cancellation := range []string{"none", "operation", "parent"} {
			t.Run(backend+"/"+cancellation, func(t *testing.T) {
				in, typ := sourceType(t, backend, "host-source-stream-u8")
				closed, consumed := 0, 0
				_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
					original := c.ctx
					parent, cancelParent := context.WithCancel(c.Context())
					defer cancelParent()
					operation, cancelOperation := context.WithCancel(context.Background())
					defer cancelOperation()
					c.ctx = parent
					source, err := c.NewSource(typ, AsyncSource{
						ReadBytes: func(uint32) (ReadResult, bool, error) {
							return ReadResult{Bytes: "*", Closed: true}, true, nil
						},
						Read: func(*Caller, uint32) (*ReadCompletion, error) {
							return nil, vm.ErrLink
						},
						Close: func() { closed++ },
					})
					if err != nil {
						return err
					}
					p, err := c.AwaitRead(operation, source.Resource, 1, func(_ *Caller, result ReadResult, failed error) (*Value, error) {
						consumed++
						if cancellation == "none" {
							if failed != nil || result.Bytes != "*" || !result.Closed {
								t.Error(result, failed)
							}
						} else if !errors.Is(failed, context.Canceled) || result.Bytes != "" {
							t.Error("canceled event published its data", result, failed)
						}
						return nil, nil
					})
					c.ctx = original
					if err != nil || !p.ready() {
						t.Fatal("read did not complete synchronously", err)
					}
					if err = p.bind(in.scheduler().wake); err != nil {
						return err
					}
					if cancellation == "operation" {
						cancelOperation()
					} else if cancellation == "parent" {
						cancelParent()
					}
					resolve, _, err := p.take()
					if err != nil {
						return err
					}
					_, err = resolve(c)
					p.close()
					if err != nil {
						return err
					}
					return c.DropAsync(source.Resource)
				})
				if err != nil || closed != 1 || consumed != 1 || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal(err, closed, consumed, in.heldValues, in.heldBytes)
				}
			})
		}
	}
}
