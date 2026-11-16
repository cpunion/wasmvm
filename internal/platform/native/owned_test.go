package native

import (
	"runtime"
	"testing"
	"time"
)

func TestOwnedEntryKeepsSchedulerProgressAndDrainsClosedLease(t *testing.T) {
	if !Available {
		return
	}
	code, err := Publish(cancelProbe(), []uint32{0}, 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(code.Close)
	lease, err := code.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(lease.Close)
	code.Close()
	type result struct {
		status Exit
		err    error
	}
	done := make(chan result, 2)
	var contexts []*Context
	pending := 0
	defer func() {
		for _, ctx := range contexts {
			ctx.Cancel()
		}
		for pending != 0 {
			select {
			case <-done:
				pending--
			case <-time.After(5 * time.Second):
				// A failing entry still owns its record; do not free it early.
				t.Error("owned entry cleanup did not drain")
				return
			}
		}
		for _, ctx := range contexts {
			ctx.Close()
		}
	}()
	for range 2 {
		ctx, err := NewContext(0, 0, 16)
		if err != nil {
			t.Fatal(err)
		}
		contexts = append(contexts, ctx)
		pending++
		go func() {
			state := State{Depth: 1}
			status, err := ctx.RunOwnedState(lease, 0, &state)
			done <- result{status, err}
		}()
	}
	deadline := time.Now().Add(5 * time.Second)
	for lease.state.Load() != 2 {
		if time.Now().After(deadline) {
			t.Fatal("both owned entries did not acquire their code pins")
		}
		runtime.Gosched()
	}
	// Native execution remains inside the supported C transition, allowing a
	// collection to finish while both entries wait for an explicit interrupt.
	runtime.GC()
	lease.Close()
	for _, ctx := range contexts {
		ctx.Cancel()
	}
	for range 2 {
		select {
		case got := <-done:
			pending--
			if got.status != Poll || got.err != nil {
				t.Fatal("owned cancellation lost its entry pin", got)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("owned entry did not drain")
		}
	}
	if code.address != nil {
		t.Fatal("retired code survived its final owned entry")
	}
}
