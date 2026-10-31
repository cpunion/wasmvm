package component

import (
	"context"
	"errors"
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
	"time"
)

func voidCompletion() *HostCompletion {
	p := NewHostCompletion()
	p.Resolve(func(*Caller) (*Value, error) { return nil, nil })
	return p
}
func TestScheduledWorkConsumesGuestStreams(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, cancelled := range []bool{false, true} {
			for _, name := range []string{"host-result-future", "host-result-stream"} {
				in := asyncInstance(t, backend, name, Imports{})
				run, _ := in.Function("run")
				signal := NewCancellation()
				if cancelled {
					signal.Cancel()
				}
				starts, dones := 0, 0
				var old *Caller
				_, err := run.CallWithResult(nil, nil, CallOptions{}, func(c *Caller, v *Value) error {
					old = c
					err := c.Schedule(AsyncWork{Owned: []*Resource{v.Resource}, Cancel: signal, Start: func(fresh *Caller, refs []*Resource) (*HostCompletion, error) {
						starts++
						if fresh == old {
							return nil, vm.ErrType
						}
						if _, err := old.ReadAsync(nil, 1); !errors.Is(err, wasm.ErrClosed) {
							return nil, vm.ErrType
						}
						r, err := fresh.ReadAsync(refs[0], 1)
						if err != nil || r.Count != 1 || r.Values[0].Bits != 42 {
							t.Fatal(name, r, err)
						}
						return voidCompletion(), nil
					}, Done: func(_ *Caller, boolCancelled bool) error {
						dones++
						if boolCancelled != cancelled {
							t.Fatal(boolCancelled)
						}
						return nil
					}})
					if err != nil {
						return err
					}
					if _, err := c.ReadAsync(v.Resource, 1); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("scheduled ownership stayed with sender", err)
					}
					return nil
				})
				want := 1
				if cancelled {
					want = 0
				}
				if err != nil || starts != want || dones != 1 {
					t.Fatal(name, cancelled, starts, dones, err)
				}
				if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
					t.Fatal("work retained state")
				}
			}
		}
	}
}
func TestScheduledCancellationInterruptsNestedHostRead(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in, typ := sourceType(t, backend, "host-source-stream-u8")
		signal := NewCancellation()
		started := make(chan struct{}, 1)
		worker := make(chan struct{}, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		go func() { <-started; signal.Cancel() }()
		done := false
		_, err := in.WithCaller(ctx, CallOptions{}, func(c *Caller) error {
			v, err := c.NewSource(typ, AsyncSource{Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
				p := NewReadCompletion()
				workCtx := c.Context()
				started <- struct{}{}
				go func() { <-workCtx.Done(); p.Cancelled(); worker <- struct{}{} }()
				return p, nil
			}})
			if err != nil {
				return err
			}
			return c.Schedule(AsyncWork{Cancel: signal, Owned: []*Resource{v.Resource}, Start: func(c *Caller, refs []*Resource) (*HostCompletion, error) {
				_, err := c.ReadAsync(refs[0], 1)
				if !errors.Is(err, context.Canceled) {
					t.Fatal("nested read ignored work context", err)
				}
				p := NewHostCompletion()
				p.Cancelled()
				return p, nil
			}, Done: func(_ *Caller, cancelled bool) error { done = cancelled; return nil }})
		})
		if err != nil || !done {
			t.Fatal(done, err)
		}
		select {
		case <-worker:
		case <-time.After(time.Second):
			t.Fatal("source worker survived cancellation")
		}
		if in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
			t.Fatal("cancel retained state")
		}
		probe := canonical.NewResourceTable(in.resourceBudget)
		for range in.config.MaxResources {
			if _, err := probe.AddOpaque(1); err != nil {
				t.Fatal("cancelled work retained endpoint", err)
			}
		}
		probe.Close()
	}
}
func TestScheduleValidationAndFailureCleanup(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	used := NewCancellation()
	_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
		if err := c.Schedule(AsyncWork{}); !errors.Is(err, vm.ErrType) {
			t.Fatal(err)
		}
		value, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }})
		if err != nil {
			return err
		}
		if err := c.Schedule(AsyncWork{Owned: []*Resource{value.Resource, value.Resource}, Start: func(*Caller, []*Resource) (*HostCompletion, error) { return voidCompletion(), nil }}); !errors.Is(err, vm.ErrType) {
			t.Fatal("duplicate work ownership", err)
		}
		if err := c.DropAsync(value.Resource); err != nil {
			return err
		}
		if err := c.Schedule(AsyncWork{Cancel: used, Start: func(*Caller, []*Resource) (*HostCompletion, error) { return voidCompletion(), nil }}); err != nil {
			return err
		}
		if err := c.Schedule(AsyncWork{Cancel: used, Start: func(*Caller, []*Resource) (*HostCompletion, error) { return voidCompletion(), nil }}); !errors.Is(err, vm.ErrType) {
			t.Fatal("reused cancellation", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"error", "panic", "shape", "nil"} {
		closed := 0
		invoke := func() (err error) {
			defer func() {
				if recover() != nil {
					err = wasm.ErrTrap
				}
			}()
			_, err = in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
				value, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }, Close: func() { closed++ }})
				if err != nil {
					return err
				}
				return c.Schedule(AsyncWork{Owned: []*Resource{value.Resource}, Start: func(*Caller, []*Resource) (*HostCompletion, error) {
					switch mode {
					case "error":
						return nil, vm.ErrLink
					case "panic":
						panic("work")
					case "nil":
						return nil, nil
					}
					p := NewHostCompletion()
					p.Resolve(func(*Caller) (*Value, error) { return &Value{Bits: 42}, nil })
					return p, nil
				}})
			})
			return err
		}
		if err := invoke(); err == nil {
			t.Fatal(mode)
		}
		if closed != 1 || in.tasks != nil || in.heldValues != 0 || in.heldBytes != 0 {
			t.Fatal(mode, "failed work retained ownership", closed)
		}
	}
}
