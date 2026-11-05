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

func TestOwnedSourceForwardingDropCancellationAndCleanup(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []uint64{0, 1, 3} {
			var typ ValueType
			var consumed *Resource
			var started atomic.Bool
			originals, wrappers := 0, 0
			imports := Imports{FunctionChecks: map[string]FunctionCheck{"make": func(info FunctionInfo) error { typ = info.Result(); return nil }}, Functions: map[string]HostFunction{
				"started": func(*Caller, []Value) (*Value, error) {
					v := &Value{}
					if started.Load() {
						v.Bits = 1
					}
					return v, nil
				},
				"make": func(c *Caller, _ []Value) (*Value, error) {
					original, err := c.NewSource(typ, AsyncSource{Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
						p := NewReadCompletion()
						started.Store(true)
						if mode == 3 {
							ctx := c.Context()
							go func() { <-ctx.Done(); p.Cancelled() }()
						} else {
							p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "*", Closed: true}, nil })
						}
						return p, nil
					}, Close: func() { originals++ }})
					if err != nil {
						return nil, err
					}
					consumed = original.Resource
					forwarded, err := c.NewSource(typ, AsyncSource{Owned: []*Resource{original.Resource}, ReadOwned: func(c *Caller, refs []*Resource, limit uint32) (*ReadCompletion, error) {
						r, err := c.ReadAsync(refs[0], limit)
						p := NewReadCompletion()
						if err != nil {
							if errors.Is(err, context.Canceled) {
								p.Cancelled()
								return p, nil
							}
							return nil, err
						}
						p.Resolve(func(*Caller) (ReadResult, error) { return r, nil })
						return p, nil
					}, Close: func() { wrappers++ }})
					return &forwarded, err
				},
			}}
			in := asyncInstance(t, backend, "host-source-stream-u8", imports)
			run, _ := in.Function("run")
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			v, err := run.Call(ctx, []Value{{Bits: mode}}, CallOptions{})
			cancel()
			if err != nil || mode == 0 && v.Value.Bits != 42 {
				t.Fatal(backend, mode, v, err)
			}
			if originals != 1 || wrappers != 1 || in.heldValues != 0 || in.heldBytes != 0 {
				t.Fatal("source ownership retained", backend, mode, originals, wrappers, in.heldValues, in.heldBytes)
			}
			if _, err := in.ReadAsync(nil, consumed, 1, CallOptions{}); err == nil {
				t.Fatal("source ownership stayed with sender")
			}
		}
	}
}
func TestOwnedSourcePublicationAndPanicCleanup(t *testing.T) {
	for _, panicClose := range []bool{false, true} {
		in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
		closed := 0
		var parent Value
		_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			child, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }, Close: func() {
				closed++
				if panicClose {
					panic("child")
				}
			}})
			if err != nil {
				return err
			}
			parent, err = c.NewSource(typ, AsyncSource{Owned: []*Resource{child.Resource}, Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }, Close: func() {
				if panicClose {
					panic("parent")
				}
			}})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		var panicked any
		func() {
			defer func() { panicked = recover() }()
			_, err = in.DropAsync(nil, parent.Resource, CallOptions{})
		}()
		if panicClose {
			if panicked != "parent" {
				t.Fatal("original panic replaced", panicked)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		if closed != 1 || in.heldValues != 0 || in.heldBytes != 0 {
			t.Fatal("drop retained owned capabilities", closed, in.heldValues, in.heldBytes)
		}
	}
}
func TestOwnedSourceValidationAndIdleClose(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	child, err := in.NewSource(nil, typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }})
	if err != nil {
		t.Fatal(err)
	}
	read := func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }
	for _, source := range []AsyncSource{
		{Read: read, ReadOwned: func(*Caller, []*Resource, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }},
		{Owned: []*Resource{child.Resource, child.Resource}, Read: read},
		{Owned: []*Resource{nil}, Read: read},
	} {
		if _, err := in.NewSource(nil, typ, source); err == nil {
			t.Fatal("invalid source ownership accepted")
		}
	}
	parent, err := in.NewSource(nil, typ, AsyncSource{Owned: []*Resource{child.Resource}, Read: read})
	if err != nil {
		t.Fatal(err)
	}
	if parent.Resource == nil || in.heldValues != 1 {
		t.Fatal("missing retained ownership")
	}
	in.Close()
	if err := in.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if in.heldValues != 0 || in.heldBytes != 0 || in.sourceCleanup != nil {
		t.Fatal("idle close retained owned source")
	}
	if _, err := new(Caller).ReadAsyncContext(context.Background(), nil, 1); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestSourceReadWithIndependentContext(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in, typ := sourceType(t, backend, "host-source-stream-u8")
		finished := make(chan struct{}, 1)
		_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			value, err := c.NewSource(typ, AsyncSource{Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
				p := NewReadCompletion()
				ctx := c.Context()
				go func() { <-ctx.Done(); p.Cancelled(); finished <- struct{}{} }()
				close(started)
				return p, nil
			}})
			if err != nil {
				return err
			}
			// Cancellation must follow source startup. A short wall deadline can
			// expire before startup on a busy host, leaving no worker to join.
			go func() {
				select {
				case <-started:
					cancel()
				case <-ctx.Done():
				}
			}()
			if _, err := c.ReadAsyncContext(ctx, value.Resource, 1); !errors.Is(err, context.Canceled) {
				t.Fatalf("%s independent read: %v", backend, err)
			}
			if c.Context().Err() != nil {
				t.Fatal("operation cancellation escaped to parent")
			}
			return c.DropAsync(value.Resource)
		})
		if err != nil {
			t.Fatal(backend, err)
		}
		select {
		case <-finished:
		case <-time.After(time.Second):
			t.Fatal("cancelled source worker retained")
		}
	}
}

func TestSourceControllerAndOwnershipAnchor(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		in, typ := sourceType(t, backend, "host-source-stream-u8")
		signal := NewCancellation()
		closed := 0
		var anchor, consumer Value
		_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
			original, err := c.NewSource(typ, AsyncSource{Read: func(*Caller, uint32) (*ReadCompletion, error) {
				p := NewReadCompletion()
				p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "*", Closed: true}, nil })
				return p, nil
			}, Close: func() { closed++ }})
			if err != nil {
				return err
			}
			var refs []*Resource
			anchor, refs, err = c.NewOwnedSource(typ, AsyncSource{Cancel: signal, Owned: []*Resource{original.Resource}, Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }})
			if err != nil {
				return err
			}
			consumer, err = c.NewSource(typ, AsyncSource{Owned: refs, ReadOwned: func(c *Caller, refs []*Resource, n uint32) (*ReadCompletion, error) {
				r, err := c.ReadAsync(refs[0], n)
				if err != nil {
					return nil, err
				}
				p := NewReadCompletion()
				p.Resolve(func(*Caller) (ReadResult, error) { return r, nil })
				return p, nil
			}})
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		// Cancellation must still wake the next Store session after the scheduler
		// that created this anchor has gone away, without reclaiming moved ownership.
		signal.Cancel()
		signal.Cancel()
		r, err := in.ReadAsync(nil, consumer.Resource, 1, CallOptions{})
		if err != nil || r.Bytes != "*" || closed != 1 {
			t.Fatal(r, closed, err)
		}
		in.DropAsync(nil, consumer.Resource, CallOptions{})
		in.DropAsync(nil, anchor.Resource, CallOptions{})
		if in.heldValues != 0 || in.heldBytes != 0 {
			t.Fatal("anchor retained ownership")
		}
	}
}
func TestSourceControllerActivePrecancelAndReusedSignal(t *testing.T) {
	for _, preCancelled := range []bool{false, true} {
		in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
		signal := NewCancellation()
		started := make(chan struct{}, 1)
		finished := make(chan struct{}, 1)
		if preCancelled {
			signal.Cancel()
		}
		value, err := in.NewSource(nil, typ, AsyncSource{Cancel: signal, Read: func(c *Caller, _ uint32) (*ReadCompletion, error) {
			p := NewReadCompletion()
			ctx := c.Context()
			started <- struct{}{}
			go func() { <-ctx.Done(); p.Cancelled(); finished <- struct{}{} }()
			return p, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		result := make(chan ReadResult, 1)
		failed := make(chan error, 1)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		go func() { r, err := in.ReadAsync(ctx, value.Resource, 1, CallOptions{}); result <- r; failed <- err }()
		if !preCancelled {
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			signal.Cancel()
		}
		select {
		case r := <-result:
			if err := <-failed; err != nil || !r.Closed {
				t.Fatal(r, err)
			}
		case <-ctx.Done():
			t.Fatal("controller stranded read")
		}
		if !preCancelled {
			select {
			case <-finished:
			case <-ctx.Done():
				t.Fatal("controller retained source work")
			}
		}
		in.DropAsync(nil, value.Resource, CallOptions{})
	}
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	signal := NewCancellation()
	source := AsyncSource{Cancel: signal, Read: func(*Caller, uint32) (*ReadCompletion, error) {
		p := NewReadCompletion()
		p.Resolve(func(*Caller) (ReadResult, error) { return ReadResult{Bytes: "*", Closed: true}, nil })
		return p, nil
	}}
	value, err := in.NewSource(nil, typ, source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.NewSource(nil, typ, source); !errors.Is(err, vm.ErrType) {
		t.Fatal("reused source signal", err)
	}
	r, err := in.ReadAsync(nil, value.Resource, 1, CallOptions{})
	if err != nil || r.Bytes != "*" {
		t.Fatal("failed construction damaged first signal", r, err)
	}
	in.DropAsync(nil, value.Resource, CallOptions{})
}

func TestScheduledCleanupIncludesAbandonedStart(t *testing.T) {
	for _, mode := range []string{"success", "before-start", "start-error", "panic"} {
		in, _ := sourceType(t, "interpreter", "host-source-stream-u8")
		cleanups := 0
		var panicked any
		func() {
			defer func() { panicked = recover() }()
			_, err := in.WithCaller(nil, CallOptions{}, func(c *Caller) error {
				if err := c.Schedule(AsyncWork{Start: func(*Caller, []*Resource) (*HostCompletion, error) {
					if mode == "start-error" {
						return nil, vm.ErrLink
					}
					if mode == "panic" {
						panic("starter")
					}
					return voidCompletion(), nil
				}, Cleanup: func() {
					cleanups++
					if mode == "panic" {
						panic("cleanup")
					}
				}}); err != nil {
					return err
				}
				if mode == "before-start" {
					return vm.ErrLink
				}
				return nil
			})
			if mode == "success" && err != nil || mode != "success" && mode != "panic" && !errors.Is(err, vm.ErrLink) {
				t.Fatal(mode, err)
			}
		}()
		if cleanups != 1 || mode == "panic" && panicked != "starter" {
			t.Fatal(mode, cleanups, panicked)
		}
	}
}

func TestSourceLateCancellationAfterJoinedClose(t *testing.T) {
	in, typ := sourceType(t, "interpreter", "host-source-stream-u8")
	signal := NewCancellation()
	_, err := in.NewSource(nil, typ, AsyncSource{Cancel: signal, Read: func(*Caller, uint32) (*ReadCompletion, error) { return nil, vm.ErrLink }})
	if err != nil {
		t.Fatal(err)
	}
	signal.mu.Lock()
	late := signal.cancel
	signal.mu.Unlock()
	if late == nil {
		t.Fatal("missing source cancellation binding")
	}
	in.Close()
	if err := in.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	late()
	in.mu.Lock()
	queued := in.sourceCleanup != nil
	in.mu.Unlock()
	if queued {
		t.Fatal("late cancellation queued work after joined shutdown")
	}
}
