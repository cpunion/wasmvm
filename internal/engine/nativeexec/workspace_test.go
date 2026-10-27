package nativeexec

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestInvocationWorkspaceLifecycle(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	e, fn, _ := prepare(t, fixture(t, false))
	if err := e.Begin(nil, false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("rebound active invocation", err)
	}
	for i := range 4 {
		if !e.Suspend() || !e.Suspend() || e.ctx != nil || e.runtimeInstance != nil {
			t.Fatal("failed to suspend idle native storage")
		}
		if _, _, err := e.Run(fn, 0, 1, 0, 1); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("entered a suspended workspace", err)
		}
		metered := i%2 == 0
		if err := e.Begin(nil, metered); err != nil {
			t.Fatal(err)
		}
		_, fuel, err := e.Run(fn, 0, 1, 0, 1)
		want := uint64(1)
		if metered {
			want = 0
		}
		if err != nil || fuel != want || e.values[0].Lo != 42 {
			t.Fatal(fuel, e.values, err)
		}
	}
	memory, err := native.NewMemory(16, 16)
	if err != nil {
		t.Fatal(err)
	}
	defer memory.Close()
	if err := e.native.BindMemory(memory); err != nil {
		t.Fatal(err)
	}
	memory.Close()
	if !e.Suspend() {
		t.Fatal("failed to drop memory binding")
	}
	// A retained same-memory binding would incorrectly bypass retain's closed check.
	if err := e.native.BindMemory(memory); !errors.Is(err, native.ErrClosed) {
		t.Fatal("suspended workspace retained retired memory", err)
	}
	e.Close()
	e.Close()
	if e.Suspend() || !errors.Is(e.Begin(nil, false), wasm.ErrInvalid) {
		t.Fatal("closed workspace was reused")
	}
	if _, _, err := e.Run(fn, 0, 1, 0, 0); !errors.Is(err, native.ErrClosed) {
		t.Fatal(err)
	}
	var zero Execution
	zero.Close()
	if zero.Suspend() || !errors.Is(zero.Begin(nil, false), wasm.ErrInvalid) {
		t.Fatal("zero workspace was reused")
	}
}

func TestCanceledWorkspaceCannotBeReused(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	e, fn, lease := prepare(t, fixture(t, false))
	if !e.Suspend() {
		t.Fatal("initial suspension")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := e.Begin(ctx, false); err != nil {
		t.Fatal(err)
	}
	cancel()
	// Observe the real asynchronous native cancellation writer before Suspend.
	// The control word remains sticky even when a new entry restores its state.
	deadline := time.Now().Add(time.Second)
	for {
		if err := e.native.SetState(native.State{Depth: 1, Budget: 64}); err != nil {
			t.Fatal(err)
		}
		exit, err := e.native.Enter(lease.code, lease.entries[fn.Code].ordinary)
		if err != nil {
			t.Fatal(err)
		}
		if exit == native.Poll {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cancellation writer did not run")
		}
		runtime.Gosched()
	}
	if e.Suspend() || e.Suspend() || !errors.Is(e.Begin(context.Background(), false), wasm.ErrInvalid) {
		t.Fatal("canceled native context was reused")
	}
}
