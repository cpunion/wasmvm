package canonical

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNominalOwnershipAndReentrantDestructor(t *testing.T) {
	r := NewResources(4)
	var destroyed []uint64
	typeID := &ResourceType{}
	typeID.Drop = func(ctx context.Context, rep uint64) error {
		destroyed = append(destroyed, rep)
		if _, err := r.Rep(typeID, 1); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("destructor resurrected handle", err)
		}
		if rep == 42 {
			id, err := r.New(typeID, 43)
			if err != nil {
				t.Fatal(err)
			}
			return r.Drop(ctx, typeID, id)
		}
		return nil
	}
	id, err := r.New(typeID, 42)
	if err != nil || id != 1 {
		t.Fatal(id, err)
	}
	if rep, err := r.Rep(typeID, id); err != nil || rep != 42 {
		t.Fatal(rep, err)
	}
	if _, err := r.Rep(&ResourceType{}, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("structural resource identity", err)
	}
	if err := r.Drop(nil, typeID, id); err != nil {
		t.Fatal(err)
	}
	if len(destroyed) != 2 || destroyed[0] != 42 || destroyed[1] != 43 {
		t.Fatal(destroyed)
	}
	if err := r.Drop(nil, typeID, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("double destructor", err)
	}
	id, err = r.New(typeID, 44)
	if err != nil || id <= 2 {
		t.Fatal(id, err)
	}
	if rep, err := r.Take(typeID, id); err != nil || rep != 44 || len(destroyed) != 2 {
		t.Fatal(rep, err, destroyed)
	}
	if _, err := r.Take(typeID, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := r.New(nil, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := r.Rep(nil, 0); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
}

func TestBorrowedHandlesAndLenderLifetime(t *testing.T) {
	source, target := NewResources(8), NewResources(8)
	typeID := &ResourceType{Drop: func(context.Context, uint64) error { t.Fatal("borrow destructor"); return nil }}
	id, _ := source.New(typeID, 0x1234)
	scope := NewBorrowScope(8)
	rep, err := scope.Lift(source, typeID, id)
	if err != nil || rep != 0x1234 {
		t.Fatal(rep, err)
	}
	borrow, err := scope.Lower(target, typeID, rep)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := target.Rep(typeID, borrow); err != nil || got != rep {
		t.Fatal(got, err)
	}
	if _, err := target.Take(typeID, borrow); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("borrow became owned", err)
	}
	if _, err := source.Take(typeID, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("moved lent resource", err)
	}
	if err := source.Drop(nil, typeID, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("dropped lent resource", err)
	}
	if err := source.Drain(nil); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := target.Drain(nil); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := target.Drop(nil, typeID, borrow); err != nil {
		t.Fatal(err)
	}
	if err := scope.End(); err != nil {
		t.Fatal(err)
	}
	if err := scope.End(); err != nil {
		t.Fatal(err)
	}
	if _, err := source.Take(typeID, id); err != nil {
		t.Fatal(err)
	}
	if _, err := scope.Lift(source, typeID, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := scope.Lower(target, typeID, rep); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	scope = NewBorrowScope(2)
	borrow, err = scope.Lower(target, typeID, rep)
	if err != nil {
		t.Fatal(err)
	}
	if err := scope.End(); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("escaped borrow accepted", err)
	}
	if _, err := target.Rep(typeID, borrow); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("escaped borrow survived cleanup", err)
	}
}

func TestResourceQuotasCancellationAndCleanupErrors(t *testing.T) {
	r := NewResources(1)
	typeID := &ResourceType{}
	id, _ := r.New(typeID, 1)
	if _, err := r.New(typeID, 2); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.Drop(ctx, typeID, id); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := r.Rep(typeID, id); err != nil {
		t.Fatal("canceled drop consumed handle", err)
	}
	scope := NewBorrowScope(0)
	if _, err := scope.Lift(r, typeID, id); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := scope.Lower(r, typeID, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	scope = NewBorrowScope(4)
	if _, err := scope.Lower(r, typeID, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := scope.Lift(r, &ResourceType{}, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := scope.Lower(r, nil, 1); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	r.entries[id].lends = ^uint32(0)
	if _, err := scope.Lift(r, typeID, id); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	r.entries[id].lends = 0
	if err := r.Drop(nil, typeID, id); err != nil {
		t.Fatal(err)
	}
	r.next = 1 << 28
	if _, err := r.New(typeID, 3); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	r = NewResources(4)
	called := 0
	typeID = &ResourceType{Drop: func(context.Context, uint64) error { called++; return context.Canceled }}
	id, _ = r.New(typeID, 1)
	if err := r.Drop(nil, typeID, id); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	r.New(typeID, 2)
	r.New(typeID, 3)
	if err := r.Drain(nil); !errors.Is(err, context.Canceled) || called != 3 || len(r.entries) != 0 {
		t.Fatal(err, called)
	}
	if err := r.Drain(nil); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentBorrowScopesPinOwner(t *testing.T) {
	r := NewResources(1024)
	var destroyed atomic.Uint32
	typeID := &ResourceType{Drop: func(context.Context, uint64) error { destroyed.Add(1); return nil }}
	id, _ := r.New(typeID, 7)
	var workers sync.WaitGroup
	for i := 0; i < 16; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for n := 0; n < 50; n++ {
				scope := NewBorrowScope(4)
				rep, err := scope.Lift(r, typeID, id)
				if err != nil || rep != 7 {
					t.Error(rep, err)
					return
				}
				borrow, err := scope.Lower(r, typeID, rep)
				if err != nil {
					t.Error(err)
					return
				}
				if err := r.Drop(nil, typeID, borrow); err != nil {
					t.Error(err)
					return
				}
				if err := scope.End(); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	workers.Wait()
	if destroyed.Load() != 0 {
		t.Fatal("borrow destroyed resource")
	}
	if err := r.Drain(nil); err != nil || destroyed.Load() != 1 {
		t.Fatal(err, destroyed.Load())
	}
}

func TestSharedResourceQuotaAndClosedDomains(t *testing.T) {
	budget := NewResourceBudget(1)
	a, b := NewResourceTable(budget), NewResourceTable(budget)
	typeID := &ResourceType{}
	rep := uint64(0xfedcba9876543210)
	id, err := a.New(typeID, rep)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.New(typeID, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("domains did not share quota", err)
	}
	if got, err := a.Take(typeID, id); err != nil || got != rep {
		t.Fatal("64-bit representation truncated", got, err)
	}
	id, err = b.New(typeID, rep)
	if err != nil {
		t.Fatal("move retained quota", err)
	}
	b.Close()
	b.Close()
	if _, err := b.Rep(typeID, id); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal("closed domain retained authority", err)
	}
	if _, err := b.New(typeID, rep); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal("closed domain published a handle", err)
	}
	if _, err := a.New(typeID, rep); err != nil {
		t.Fatal("close retained quota", err)
	}
	if err := a.Drain(nil); err != nil || budget.live.Load() != 0 {
		t.Fatal("drain retained quota", err, budget.live.Load())
	}
	scope := NewBorrowScope(1)
	if _, err := scope.Lower(a, typeID, rep); err != nil {
		t.Fatal(err)
	}
	if err := scope.End(); !errors.Is(err, wasm.ErrTrap) || budget.live.Load() != 0 {
		t.Fatal("escaped borrow retained quota", err, budget.live.Load())
	}
}
