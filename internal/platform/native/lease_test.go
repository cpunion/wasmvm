package native

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestLeaseAdmissionRollbackAndOverflow(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	code, lease, ctx := makeProbe(t, addProbe())
	if _, err := lease.begin(1); err != ErrInvalid || lease.state.Load() != 0 {
		t.Fatal("invalid entry retained an admission", err, lease.state.Load())
	}
	lease.state.Store(leaseClosed - 1)
	if _, err := lease.begin(0); err != ErrLimit || lease.state.Load() != leaseClosed-1 {
		t.Fatal("entry count wrapped", err, lease.state.Load())
	}
	lease.state.Store(0)
	state := State{ResultLo: 20, ResultHi: 22, Payload: [4]uint64{uint64(Return)}}
	if exit, err := ctx.RunState(lease, 0, &state); err != nil || exit != Return || state.ResultLo != 42 || lease.state.Load() != 0 {
		t.Fatal("failed admission damaged subsequent entry", exit, err, state, lease.state.Load())
	}
	code.Close()
	lease.Close()
	lease.Close()
	if _, err := lease.begin(0); err != ErrClosed || code.address != nil {
		t.Fatal("closed lease admitted or retained code", err)
	}
	var zero Lease
	if _, err := zero.begin(0); err != ErrClosed {
		t.Fatal("zero lease admitted", err)
	}
	zero.Close()
	zero.Close()
}

func TestClosedLeaseDrainsBothNativeEntries(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	code, lease, first := makeProbe(t, cancelProbe())
	second, err := NewContext(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	contexts := []*Context{first, second}
	done := []chan error{make(chan error, 1), make(chan error, 1)}
	for i, ctx := range contexts {
		go func() {
			exit, err := ctx.Enter(lease, 0)
			if err == nil && exit != Poll {
				err = ErrInvalid
			}
			done[i] <- err
		}()
	}
	deadline := time.Now().Add(5 * time.Second)
	for lease.state.Load() != 2 {
		if time.Now().After(deadline) {
			t.Fatal("both entries did not start", lease.state.Load())
		}
		runtime.Gosched()
	}
	code.Close()
	lease.Close()
	if lease.state.Load() != leaseClosed|2 {
		t.Fatal("closing lost admitted entries", lease.state.Load())
	}
	if _, err := lease.begin(0); err != ErrClosed {
		t.Fatal("closed lease admitted a third entry", err)
	}
	for i, ctx := range contexts {
		ctx.Cancel()
		select {
		case err := <-done[i]:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("entry did not finish", i)
		}
		code.mu.Lock()
		retained := code.address != nil
		code.mu.Unlock()
		if retained != (i == 0) {
			t.Fatal("code lifetime did not follow admitted entries", i, retained)
		}
		if i == 0 {
			runtime.GC()
		}
	}
	if lease.code != nil || lease.state.Load() != leaseClosed {
		t.Fatal("drained lease retained code", lease.state.Load())
	}
}

func TestLeaseAdmissionRacesWithClose(t *testing.T) {
	if !Available {
		t.Skip("native platform unavailable")
	}
	code, lease, _ := makeProbe(t, addProbe())
	entered := make(chan struct{}, 4)
	errors := make(chan error, 4)
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < 1000; i++ {
				entry, err := lease.begin(0)
				if err == ErrClosed {
					return
				}
				if err != nil || entry == nil {
					errors <- ErrInvalid
					return
				}
				if i == 0 {
					entered <- struct{}{}
				}
				runtime.Gosched()
				lease.end()
			}
		}()
	}
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("no admission started")
	}
	code.Close()
	lease.Close()
	workers.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
	if code.address != nil || lease.code != nil || lease.state.Load() != leaseClosed {
		t.Fatal("concurrent close leaked a code pin", lease.state.Load())
	}
}
