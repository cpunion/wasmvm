package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCallerCloseInstanceCapability(t *testing.T) {
	for _, c := range []*Caller{nil, {}} {
		if err := c.CloseInstance(); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	}
	e, err := NewEngine(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "host-reexport"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	var retained *Caller
	var copied Caller
	var action func(*Caller) error
	in, err := m.Instantiate(nil, Imports{Functions: map[string]HostFunction{
		"run": func(c *Caller, _ []Value) (*Value, error) {
			retained, copied = c, *c
			if action != nil {
				return nil, action(c)
			}
			return nil, nil
		},
	}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	run, err := in.Function("run")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := run.Call(nil, nil, CallOptions{}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []*Caller{retained, &copied} {
		if err := c.CloseInstance(); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("expired caller retained shutdown authority", err)
		}
	}
	action = func(c *Caller) error {
		c.mu.Lock()
		err := copied.CloseInstance()
		c.mu.Unlock()
		if !errors.Is(err, vm.ErrBusy) {
			t.Fatal("copied caller bypassed serialization", err)
		}
		return nil
	}
	if _, err := run.Call(nil, nil, CallOptions{}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	action = func(c *Caller) error {
		cancel()
		return c.CloseInstance()
	}
	if _, err := run.Call(ctx, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	action = nil
	if _, err := run.Call(nil, nil, CallOptions{}); err != nil {
		t.Fatal("a rejected close stopped the Store", err)
	}
	sentinel := errors.New("host requested shutdown")
	action = func(c *Caller) error {
		if err := c.CloseInstance(); err != nil {
			t.Fatal(err)
		}
		return sentinel
	}
	if _, err := run.Call(nil, nil, CallOptions{}); !errors.Is(err, sentinel) {
		t.Fatal("cancellation hid the explicit callback outcome", err)
	}
	if _, err := run.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := in.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
}
