package canonical

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTaskReturnBorrowBoundary(t *testing.T) {
	owner, guest := NewResources(4), NewResources(4)
	defer owner.Close()
	defer guest.Close()
	typ := &ResourceType{}
	id, err := owner.New(typ, 42)
	if err != nil {
		t.Fatal(err)
	}
	scope := NewBorrowScope(4)
	rep, err := scope.Lift(owner, typ, id)
	if err != nil || rep != 42 {
		t.Fatal(rep, err)
	}
	if err := scope.CheckReturned(); err != nil {
		t.Fatal("lenders are not borrowed guest handles", err)
	}
	borrow, err := scope.Lower(guest, typ, rep)
	if err != nil {
		t.Fatal(err)
	}
	if err := scope.CheckReturned(); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("escaped guest handle", err)
	}
	if err := guest.Drop(nil, typ, borrow); err != nil {
		t.Fatal(err)
	}
	if err := scope.CheckReturned(); err != nil {
		t.Fatal(err)
	}
	if err := scope.End(); err != nil {
		t.Fatal(err)
	}
	if err := scope.CheckReturned(); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("ended scope reused", err)
	}
	if _, err := owner.Take(typ, id); err != nil {
		t.Fatal("lender remained pinned", err)
	}
}
