package preview3

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func TestHTTPReplyAbandonPreservesPublishedFailure(t *testing.T) {
	reply := &httpReply{done: make(chan struct{})}
	reply.set(nil, vm.ErrBusy)
	if err := reply.abandon(); !errors.Is(err, vm.ErrBusy) {
		t.Fatal(err)
	}
	// A later cancellation cannot replace the already committed failure.
	reply.set(nil, context.Canceled)
	if err := reply.abandon(); !errors.Is(err, vm.ErrBusy) {
		t.Fatal(err)
	}

	unfinished := &httpReply{done: make(chan struct{})}
	if err := unfinished.abandon(); err != nil {
		t.Fatal(err)
	}
	unfinished.set(nil, vm.ErrBusy)
	if err := unfinished.abandon(); err != nil {
		t.Fatal("late failure published after abandonment", err)
	}
}

func TestHTTPIncomingBusyStoreDoesNotRetainRequest(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{})
		p := w.HTTP()
		in := instantiate(t, compileFixture(t, backend, "world-rust-http"), w.Imports())
		p.mu.Lock()
		initial := len(p.entries)
		p.mu.Unlock()

		// Pause one accepted idle worker before its next job starts. This uses
		// the real worker accounting and handoff protocol, with explicit gates
		// instead of a timing race between request publication and Store entry.
		jobs := make(chan func(), 1)
		admitted, resume := make(chan struct{}), make(chan struct{})
		var releaseWorker sync.Once
		defer releaseWorker.Do(func() { close(resume) })
		owner := p.network.io
		owner.mu.Lock()
		owner.workers.Add(1)
		owner.idleWorkers = append(owner.idleWorkers, jobs)
		owner.mu.Unlock()
		go func() {
			fn, ok := <-jobs
			if !ok {
				owner.workers.Done()
				return
			}
			close(admitted)
			<-resume
			owner.runWorker(jobs, fn)
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		called := make(chan error, 1)
		go func() {
			response, err := p.CallIncoming(ctx, in, httptest.NewRequest("POST", "http://localhost/", strings.NewReader("body")), component.CallOptions{})
			if response != nil {
				_ = response.Body.Close()
			}
			called <- err
		}()
		select {
		case <-admitted:
		case <-ctx.Done():
			t.Fatal("handler worker did not accept the request")
		}
		entered, release := make(chan struct{}), make(chan struct{})
		var releaseLease sync.Once
		defer releaseLease.Do(func() { close(release) })
		leaseDone := make(chan error, 1)
		go func() {
			_, err := in.WithCaller(ctx, component.CallOptions{}, func(*component.Caller) error {
				close(entered)
				<-release
				return nil
			})
			leaseDone <- err
		}()
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal("competing Store lease did not enter")
		}
		releaseWorker.Do(func() { close(resume) })
		select {
		case err := <-called:
			if !errors.Is(err, vm.ErrBusy) {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal("rejected Store entry did not return")
		}
		releaseLease.Do(func() { close(release) })
		if err := <-leaseDone; err != nil {
			t.Fatal(err)
		}
		p.mu.Lock()
		remaining := len(p.entries)
		p.mu.Unlock()
		if remaining != initial {
			t.Fatal("request resource escaped a rejected Store", remaining, initial)
		}
	})
}
