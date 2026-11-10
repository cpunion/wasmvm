package canonical

import (
	"errors"
	"sync"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestOpaqueReferencesShareMoveState(t *testing.T) {
	table := NewResources(2)
	typ := &ResourceType{}
	id, _ := table.New(typ, 42)
	ref := NewReference(table, typ, id, "owner", nil)
	alias := *ref
	if ReferenceKey(ref) != ReferenceKey(&alias) {
		t.Fatal("copy did not share state")
	}
	moved, err := MoveReference(ref)
	if err != nil {
		t.Fatal(err)
	}
	for _, stale := range []*Reference{ref, &alias} {
		if _, _, _, _, _, err := ReferenceInfo(stale); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if _, err := MoveReference(stale); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
	}
	gotTable, gotType, gotID, owner, borrow, err := ReferenceInfo(moved)
	if err != nil || gotTable != table || gotType != typ || gotID != id || owner != "owner" || borrow {
		t.Fatal(err)
	}
	scope := NewBorrowScope(1)
	if _, err := scope.Lift(table, typ, id); err != nil {
		t.Fatal(err)
	}
	if err := table.CheckOwn(typ, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	borrowed := NewReference(table, typ, id, "owner", scope)
	if _, err := MoveReference(borrowed); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := scope.End(); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, _, err := ReferenceInfo(borrowed); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	if err := table.CheckOwn(typ, id); err != nil {
		t.Fatal(err)
	}
	if err := table.CheckOwn(&ResourceType{}, id); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	InvalidateReference(moved)
	if _, _, _, _, _, err := ReferenceInfo(moved); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	for _, zero := range []*Reference{nil, new(Reference)} {
		if _, _, _, _, _, err := ReferenceInfo(zero); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		if _, err := MoveReference(zero); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		InvalidateReference(zero)
	}
	if ReferenceKey(nil) != nil {
		t.Fatal("nil key")
	}
}

func TestConcurrentReferenceMoveHasOneOwner(t *testing.T) {
	table := NewResources(1)
	typ := &ResourceType{}
	id, _ := table.New(typ, 1)
	ref := NewReference(table, typ, id, nil, nil)
	var wg sync.WaitGroup
	winners := make(chan *Reference, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := MoveReference(ref)
			if err == nil {
				winners <- v
			} else if !errors.Is(err, wasm.ErrTrap) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	close(winners)
	if len(winners) != 1 {
		t.Fatal("ownership duplicated", len(winners))
	}
	if err := table.CheckOwn(typ, id); err != nil {
		t.Fatal(err)
	}
	table.Close()
	if err := table.CheckOwn(typ, id); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

// Keep live and nil controls next to invalid-reference inspection: normal
// affine cleanup frequently examines aliases invalidated by a successful move.
func BenchmarkReferenceInspection(b *testing.B) {
	for _, name := range []string{"live", "expired", "ended-borrow", "nil"} {
		b.Run(name, func(b *testing.B) {
			ref := NewReference(nil, nil, 1, nil, nil)
			switch name {
			case "expired":
				InvalidateReference(ref)
			case "ended-borrow":
				scope := NewBorrowScope(1)
				ref = NewReference(nil, nil, 1, nil, scope)
				if err := scope.End(); err != nil {
					b.Fatal(err)
				}
			case "nil":
				ref = nil
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, _, _, _, _, err := ReferenceInfo(ref)
				switch name {
				case "live":
					if err != nil {
						b.Fatal(err)
					}
				case "nil":
					if !errors.Is(err, wasm.ErrInvalid) {
						b.Fatal(err)
					}
				default:
					if !errors.Is(err, wasm.ErrTrap) {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
