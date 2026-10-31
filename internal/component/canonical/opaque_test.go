package canonical

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestOpaqueHandleDomainAndQuota(t *testing.T) {
	budget := NewResourceBudget(2)
	a, b := NewResourceTable(budget), NewResourceTable(budget)
	if _, err := a.AddOpaque(nil); err == nil {
		t.Fatal("nil entry")
	}
	value := new(int)
	id, err := a.AddOpaque(value)
	if err != nil {
		t.Fatal(err)
	}
	typ := &ResourceType{}
	resource, err := b.New(typ, 9)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.AddOpaque(value); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := a.Rep(typ, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if _, err := b.Opaque(resource); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := b.RemoveOpaque(resource); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if got, err := a.Opaque(id); err != nil || got != value {
		t.Fatal(got, err)
	}
	if err := a.RemoveOpaque(id); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Opaque(id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := a.RemoveOpaque(id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	next, err := a.AddOpaque(value)
	if err != nil || next <= id {
		t.Fatal(next, err)
	}
	if err := a.Drain(nil); err != nil {
		t.Fatal(err)
	}
	a.Close()
	if _, err := a.AddOpaque(value); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := a.Opaque(next); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := a.RemoveOpaque(next); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	b.Close()
	if budget.live.Load() != 0 {
		t.Fatal("shared quota leaked")
	}
}
