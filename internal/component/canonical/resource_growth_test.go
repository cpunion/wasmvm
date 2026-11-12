package canonical

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGuestGrowthKeepsBorrowAndReentrantDrainOwnership(t *testing.T) {
	budget := NewResourceBudget(512)
	host, guest := NewResourceTable(budget), NewGuestResourceTable(budget)
	defer host.Close()
	defer guest.Close()
	typ := &ResourceType{}
	owner, err := host.New(typ, 71)
	if err != nil {
		t.Fatal(err)
	}
	scope := NewBorrowScope(512)
	rep, err := scope.Lift(host, typ, owner)
	if err != nil {
		t.Fatal(err)
	}
	borrow, err := scope.Lower(guest, typ, rep)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 257; i++ {
		if _, err := guest.New(typ, uint64(i)); err != nil {
			t.Fatal(i, err)
		}
	}
	if got, err := guest.Rep(typ, borrow); err != nil || got != 71 {
		t.Fatal("growth changed borrow", got, err)
	}
	if _, err := host.Take(typ, owner); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("growth lost lender", err)
	}
	// An escaped borrow must still be removed using its original entry identity.
	if err := scope.End(); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("escaped borrow survived growth", err)
	}
	if _, err := guest.Rep(typ, borrow); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("borrow not reclaimed", err)
	}
	if got, err := host.Take(typ, owner); err != nil || got != 71 {
		t.Fatal("lender not released", got, err)
	}
	if err := guest.Drain(nil); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("destructor failure")
	var destroyed int
	var reentrant uint32
	drop := &ResourceType{Drop: func(context.Context, uint64) error {
		destroyed++
		if destroyed == 1 {
			var err error
			reentrant, err = guest.New(typ, 99)
			if err != nil {
				t.Fatal(err)
			}
			return sentinel
		}
		return nil
	}}
	for i := 0; i < 257; i++ {
		if _, err := guest.New(drop, uint64(i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := guest.Drain(nil); !errors.Is(err, sentinel) || destroyed != 257 {
		t.Fatal("drain stopped early", destroyed, err)
	}
	if got, err := guest.Rep(typ, reentrant); err != nil || got != 99 {
		t.Fatal("drain consumed reentrant publication", got, err)
	}
	if budget.live.Load() != 1 {
		t.Fatal("shared quota not released", budget.live.Load())
	}
	guest.Close()
	host.Close()
	if budget.live.Load() != 0 {
		t.Fatal("close leaked quota", budget.live.Load())
	}
}
