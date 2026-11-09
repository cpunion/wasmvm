package preview3

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func awaitIdle(t *testing.T, p *IO, count int) []chan func() {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		p.mu.Lock()
		idle, active := append([]chan func(){}, p.idleWorkers...), p.activeWorkers
		p.mu.Unlock()
		if len(idle) == count && active == 0 {
			return idle
		}
		if time.Now().After(deadline) {
			t.Fatal("workers did not quiesce", len(idle), active, count)
		}
		runtime.Gosched()
	}
}

func TestIOWorkerReuseQuotaAndIdleRetirement(t *testing.T) {
	if _, err := NewIO(IOConfig{MaxStreams: 1, MaxIdleWorkers: 3}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	p, err := NewIO(IOConfig{MaxStreams: 2, MaxIdleWorkers: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer closeIO(t, p)
	var completed atomic.Int32
	if err := p.worker(func() { completed.Add(1) }); err != nil {
		t.Fatal(err)
	}
	idle := awaitIdle(t, p, 1)
	for range 20 {
		if err := p.worker(func() { completed.Add(1) }); err != nil {
			t.Fatal(err)
		}
		if next := awaitIdle(t, p, 1); next[0] != idle[0] {
			t.Fatal("sequential work created another worker")
		}
	}
	if completed.Load() != 21 {
		t.Fatal(completed.Load())
	}
	// Busy workers must not serialize accepted operations: all four can
	// begin even with a one-worker idle retention limit.
	started := make(chan struct{}, 4)
	resume := make(chan struct{})
	for range 4 {
		if err := p.worker(func() { started <- struct{}{}; <-resume }); err != nil {
			t.Fatal(err)
		}
	}
	for range 4 {
		select {
		case <-started:
		case <-time.After(3 * time.Second):
			t.Fatal("an accepted blocking operation was queued behind a busy worker")
		}
	}
	if err := p.worker(func() { t.Error("over-limit work executed") }); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	close(resume)
	awaitIdle(t, p, 1)
	p.Close()
	if err := p.worker(func() { t.Error("closed work executed") }); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestIOWorkerCloseJoinsRunningAndIdleWorkers(t *testing.T) {
	p, _ := NewIO(IOConfig{})
	if err := p.worker(func() {}); err != nil {
		t.Fatal(err)
	}
	awaitIdle(t, p, 1)
	started := make(chan struct{})
	resume := make(chan struct{})
	if err := p.worker(func() { close(started); <-resume }); err != nil {
		t.Fatal(err)
	}
	<-started
	p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("close did not join the running worker", err)
	}
	close(resume)
	closeIO(t, p)
	p.mu.Lock()
	active, idle := p.activeWorkers, len(p.idleWorkers)
	p.mu.Unlock()
	if active != 0 || idle != 0 {
		t.Fatal(active, idle)
	}
}

func TestIOWorkerConcurrentSubmitAndClose(t *testing.T) {
	for range 20 {
		p, _ := NewIO(IOConfig{MaxStreams: 4, MaxIdleWorkers: 2})
		var accepted, completed atomic.Int32
		var submit sync.WaitGroup
		for range 8 {
			submit.Go(func() {
				for range 20 {
					err := p.worker(func() { completed.Add(1) })
					if err == nil {
						accepted.Add(1)
					} else if !errors.Is(err, wasm.ErrClosed) && !errors.Is(err, wasm.ErrLimit) {
						t.Error(err)
					}
				}
			})
		}
		p.Close()
		submit.Wait()
		closeIO(t, p)
		if accepted.Load() != completed.Load() {
			t.Fatal("accepted work was lost or duplicated", accepted.Load(), completed.Load())
		}
	}
}
