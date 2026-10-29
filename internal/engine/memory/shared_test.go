package memory

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSharedBackingReferencesAndSynchronizedOperations(t *testing.T) {
	limits := module.Limits{Min: 1, Max: 2, HasMax: true, Shared: true}
	d := NewDomain(4, false)
	m, err := NewShared(limits, 2, d)
	if err != nil {
		t.Fatal(err)
	}
	if m.Native() != nil || !m.Shared() || m.ReservationPages() != 2 || m.Domain() != d || !m.Retain() {
		t.Fatal("backing contract")
	}
	m.Release()
	if !m.Write(0, []byte{1, 2, 3, 4}) {
		t.Fatal("write")
	}
	var out [4]byte
	if !m.Read(0, out[:]) || out != [4]byte{1, 2, 3, 4} {
		t.Fatal(out)
	}
	if v, trap := m.Load(0x28, 0, 0, 0, value.Value{}); trap != wasm.NoTrap || v.Lo != 0x04030201 {
		t.Fatal(v, trap)
	}
	if trap := m.Store(0x36, 4, 0, 0, value.Value{Lo: 0x05060708}); trap != wasm.NoTrap {
		t.Fatal(trap)
	}
	if !m.Copy(8, 0, 8) || !m.Fill(16, 7, 4) || !m.Init(20, 1, 3, []byte{0, 8, 9, 10}) {
		t.Fatal("bulk")
	}
	p, err := New(module.Limits{Min: 1}, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Release()
	if !p.CopyFrom(m, 0, 8, 16) || !m.CopyFrom(p, 24, 0, 16) {
		t.Fatal("private/shared copy")
	}
	other, _ := NewShared(limits, 2, NewDomain(1, false))
	defer other.Release()
	if m.CopyFrom(other, 0, 0, 1) {
		t.Fatal("cross-domain copy")
	}
	if old, ok := m.Grow(1); !ok || old != 1 || m.Pages() != 2 {
		t.Fatal(old, ok)
	}
	if _, ok := m.Grow(1); ok {
		t.Fatal("growth exceeded reservation")
	}
	for _, ok := range []bool{m.Read(2*PageSize, out[:]), m.Write(2*PageSize, out[:]), m.Fill(2*PageSize, 1, 1), m.Init(0, 2, 1, []byte{1}), m.Copy(0, 2*PageSize, 1)} {
		if ok {
			t.Fatal("range accepted")
		}
	}
	if _, err := New(limits, 2); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := NewNative(limits, 2); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := NewShared(limits, 2, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	bad := limits
	bad.HasMax = false
	if _, err := NewShared(bad, 2, d); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := NewShared(limits, 0, d); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	m.refs = math.MaxUint64
	if m.Retain() {
		t.Fatal("reference overflow")
	}
	m.refs = 1
	m.Release()
	m.Release()
	if m.Retain() {
		t.Fatal("retired backing retained")
	}
	if _, ok := m.Grow(0); ok {
		t.Fatal("retired growth")
	}
}

func TestWaiterQuotaCancellationTimeoutAndNotification(t *testing.T) {
	d := NewDomain(1, false)
	m, err := NewShared(module.Limits{Min: 1, Max: 1, HasMax: true, Shared: true}, 1, d)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Release()
	type result struct {
		v    value.Value
		trap wasm.TrapCode
		err  error
	}
	waitFor := func(n uint32) {
		t.Helper()
		end := time.Now().Add(3 * time.Second)
		for d.Waiters() != n {
			if time.Now().After(end) {
				t.Fatal("waiters", d.Waiters(), n)
			}
			time.Sleep(100 * time.Microsecond)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan result, 1)
	go func() { v, trap, err := m.Atomic(ctx, d, 1, 0, 0, 0, ^uint64(0)); done <- result{v, trap, err} }()
	waitFor(1)
	if _, _, err := m.Atomic(nil, d, 2, 0, 0, 0, ^uint64(0)); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if v, _, err := m.Atomic(nil, d, 0, 4, 0, 1, 0); err != nil || v.Lo != 0 {
		t.Fatal(v, err)
	}
	cancel()
	got := <-done
	if !errors.Is(got.err, context.Canceled) || got.trap != wasm.NoTrap {
		t.Fatal(got)
	}
	waitFor(0)
	if len(d.queues) != 0 {
		t.Fatal("wait queue leaked")
	}
	if v, _, err := m.Atomic(nil, d, 1, 0, 0, 0, 1); err != nil || v.Lo != 2 {
		t.Fatal(v, err)
	}
	go func() { v, trap, err := m.Atomic(nil, d, 2, 0, 0, 0, ^uint64(0)); done <- result{v, trap, err} }()
	waitFor(1)
	if v, _, err := m.Atomic(nil, d, 0, 0, 0, math.MaxUint32, 0); err != nil || v.Lo != 1 {
		t.Fatal(v, err)
	}
	got = <-done
	if got.err != nil || got.v.Lo != 0 {
		t.Fatal(got)
	}
	waitFor(0)
	if _, _, err := m.Atomic(ctx, d, 1, 0, 0, 0, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, _, err := m.Atomic(nil, NewDomain(1, false), 0, 0, 0, 1, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, _, err := m.Atomic(nil, d, 4, 0, 0, 0, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	d.Fence()
}
