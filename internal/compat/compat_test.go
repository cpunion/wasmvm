package compat

import (
	"context"
	"encoding/binary"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Guest numeric slots must preserve bits rather than rely on host byte order.
func TestNumericBits(t *testing.T) {
	bits := []uint64{0, 1, 0x8000000000000000, 0x7ff0000000000000, 0x7ff8000000000042, 0xffffffffffffffff}
	var buf [8]byte
	for _, want := range bits {
		binary.LittleEndian.PutUint64(buf[:], want)
		if got := binary.LittleEndian.Uint64(buf[:]); got != want {
			t.Fatalf("little-endian round trip: got %#x, want %#x", got, want)
		}
		if got := math.Float64bits(math.Float64frombits(want)); got != want {
			t.Fatalf("float bit round trip: got %#x, want %#x", got, want)
		}
	}
	binary.LittleEndian.PutUint64(buf[:], 0x0807060504030201)
	for i, b := range buf {
		if b != byte(i+1) {
			t.Fatalf("byte %d: got %d, want %d", i, b, i+1)
		}
	}
}

// Typed host references must remain reachable under either host collector.
// This intentionally makes no assertion about collection or finalizer timing.
func TestReachableReferenceCycle(t *testing.T) {
	type node struct {
		next    *node
		payload []byte
	}
	a := &node{payload: []byte{11, 22, 33}}
	b := &node{next: a, payload: []byte{44, 55, 66}}
	a.next = b
	roots := []any{a}
	a, b = nil, nil
	runtime.GC()
	root := roots[0].(*node)
	if root.next.next != root || root.payload[2] != 33 || root.next.payload[0] != 44 {
		t.Fatal("typed references or object identity were lost")
	}
	runtime.KeepAlive(roots)
}

func TestConcurrentAtomicsAndLocks(t *testing.T) {
	const workers, iterations = 4, 1000
	var atomicCount atomic.Uint64
	var mu sync.Mutex
	var lockedCount uint64
	var wg sync.WaitGroup
	wg.Add(workers)
	for n := 0; n < workers; n++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				atomicCount.Add(1)
				mu.Lock()
				lockedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	const want = workers * iterations
	if got := atomicCount.Load(); got != want || lockedCount != want {
		t.Fatalf("atomic=%d locked=%d, want %d", got, lockedCount, want)
	}
	if !atomicCount.CompareAndSwap(want, 7) || atomicCount.Swap(9) != 7 || atomicCount.Load() != 9 {
		t.Fatal("atomic compare/swap contract failed")
	}
}

// A synchronized waiter handshake checks the host primitives needed by the
// planned wait registry. It does not implement or certify Wasm atomic.wait.
func TestConditionWakeup(t *testing.T) {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := make(chan struct{})
	done := make(chan struct{})
	released := false
	go func() {
		mu.Lock()
		close(ready)
		for !released {
			cond.Wait()
		}
		mu.Unlock()
		close(done)
	}()
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("waiter did not start")
	}
	mu.Lock()
	released = true
	cond.Signal()
	mu.Unlock()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("waiter did not wake")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		<-ctx.Done()
		done <- ctx.Err()
	}()
	cancel()
	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("got %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation did not propagate")
	}
}
