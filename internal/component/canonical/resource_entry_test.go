package canonical

import (
	"runtime"
	"testing"
)

func TestFreshResourceEntriesSurviveReuseAndSourceClose(t *testing.T) {
	for _, guest := range []bool{false, true} {
		budget := NewResourceBudget(2)
		r := NewResourceTable(budget)
		if guest {
			r.Close()
			r = NewGuestResourceTable(budget)
		}
		target := NewGuestResourceTable(budget)
		typ := &ResourceType{}
		seen := map[*resourceEntry]bool{}
		var previousID uint32
		for i := uint64(0); i < 65; i++ {
			var id uint32
			var err error
			var scope *BorrowScope
			switch i % 4 {
			case 0:
				id, err = r.New(typ, i)
			case 1:
				id, err = r.AddOpaque(i)
			case 2:
				value, _ := NewErrorContext("diagnostic")
				id, err = r.AddErrorContext(value)
			case 3:
				scope = NewBorrowScope(1)
				id, err = scope.Lower(r, typ, i)
			}
			if err != nil {
				t.Fatal(guest, i, err)
			}
			if i != 0 && (guest && id != previousID || !guest && id <= previousID) {
				t.Fatal("changed guest LIFO or host nonreuse policy", guest, id, previousID)
			}
			previousID = id
			entry := r.entryLocked(id)
			if seen[entry] {
				t.Fatal("retired entry identity was rebound", guest, i)
			}
			seen[entry] = true
			switch i % 4 {
			case 0:
				if rep, err := r.Take(typ, id); err != nil || rep != i {
					t.Fatal(rep, err)
				}
			case 1:
				err = r.RemoveOpaque(id)
			case 2:
				err = r.DropErrorContext(id)
			case 3:
				err = r.Drop(nil, typ, id)
				if end := scope.End(); end != nil {
					t.Fatal(end)
				}
			}
			if err != nil || budget.live.Load() != 0 {
				t.Fatal("publication retained quota", err, budget.live.Load())
			}
			if i%8 == 0 {
				runtime.GC()
			}
		}
		id, err := r.AddOpaque("moved")
		if err != nil {
			t.Fatal(err)
		}
		entry := r.entryLocked(id)
		moved, err := r.MoveOpaque(id, target)
		if err != nil || target.entryLocked(moved) != entry {
			t.Fatal("transfer changed the publication object", err)
		}
		r.Close()
		runtime.GC()
		if value, err := target.Opaque(moved); err != nil || value != "moved" {
			t.Fatal("source close invalidated a transferred entry", value, err)
		}
		if r.entryBlock != nil {
			t.Fatal("closed source retains its allocation cursor")
		}
		target.Close()
		if budget.live.Load() != 0 {
			t.Fatal("moved publication retained quota")
		}
	}
}
