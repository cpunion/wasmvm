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
		if !e.Suspend() || !e.Suspend() || e.ctx != nil || e.runtimeInstance != nil || e.entryState != (native.State{}) {
			t.Fatal("failed to suspend idle native storage")
		}
		if _, _, err := e.Run(fn, 0, 1, 0, 1); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("entered a suspended workspace", err)
		}
		metered := i%2 == 0
		if err := e.Begin(nil, metered); err != nil {
			t.Fatal(err)
		}
		pc, fuel, err := e.Run(fn, 0, 1, 0, 1)
		want := uint64(1)
		if metered {
			want = 0
		}
		if err != nil || fuel != want {
			t.Fatal(fuel, e.values, err)
		}
		in := fn.Code.Code[pc]
		if in.Count != 1 || e.values[fn.Code.Args[in.Extra]].Lo != 42 {
			t.Fatal("workspace result", in, e.values)
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
	if e.entryState != (native.State{}) {
		t.Fatal("closed workspace retained state")
	}
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
		exit, err := e.native.Enter(lease.code, lease.entries[fn.Index-lease.first].ordinary)
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

func TestOwnedInvocationCancellationAndRebinding(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	e, fn, lease := prepare(t, fixture(t, false))
	if !e.Suspend() {
		t.Fatal("initial suspension")
	}
	for range 4 {
		ctx, cancel := context.WithCancel(context.Background())
		if err := e.BeginOwned(ctx, false); err != nil {
			t.Fatal(err)
		}
		if err := e.BeginOwned(ctx, false); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("rebound active owned invocation", err)
		}
		if _, _, err := e.Run(fn, 0, 1, 0, 1); err != nil || !e.Suspend() {
			t.Fatal("owned invocation could not quiesce", err)
		}
		// The former owner has detached this workspace before canceling its
		// context. The retired context must not write the next native binding.
		cancel()
		runtime.GC()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := e.BeginOwned(ctx, false); err != nil {
		t.Fatal(err)
	}
	cancel()
	e.Interrupt()
	if exit, err := e.native.Enter(lease.code, lease.entries[fn.Index-lease.first].ordinary); err != nil || exit != native.Poll {
		t.Fatal("owned interrupt did not reach native control", exit, err)
	}
	if _, _, err := e.Run(fn, 0, 1, 0, 0); !errors.Is(err, context.Canceled) {
		t.Fatal("owned invocation lost cancellation", err)
	}
	if e.Suspend() || e.Suspend() || !errors.Is(e.BeginOwned(nil, false), wasm.ErrInvalid) {
		t.Fatal("interrupted owned context was recycled")
	}
}
