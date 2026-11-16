package canonical

import (
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestReferenceAllocatorKeepsFreshAffineStateAcrossBlocksAndClear(t *testing.T) {
	var allocator ReferenceAllocator
	table, typ, owner := NewResources(128), &ResourceType{}, new(int)
	var originals []*Reference
	var headers []Reference
	var moved []*Reference
	keys := map[any]bool{}
	for i := range 65 {
		r := allocator.New(table, typ, uint32(i+1), owner, nil)
		key := ReferenceKey(r)
		if keys[key] {
			t.Fatal("allocator reused a published affine state")
		}
		keys[key] = true
		originals, headers = append(originals, r), append(headers, *r)
		if i%3 == 0 {
			next, err := allocator.Move(&headers[i])
			if err != nil || ReferenceKey(next) == key {
				t.Fatal("move reused the previous binding", err)
			}
			moved = append(moved, next)
		} else {
			moved = append(moved, nil)
		}
		if i%7 == 0 {
			allocator.Clear()
			runtime.GC()
		}
	}
	allocator.Clear()
	for i, r := range originals {
		if moved[i] != nil {
			for _, old := range []*Reference{r, &headers[i]} {
				if _, _, _, _, _, err := ReferenceInfo(old); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("old alias was revived", i, err)
				}
				if _, err := allocator.Move(old); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("old alias moved twice", i, err)
				}
			}
			r = moved[i]
		}
		gotTable, gotType, id, gotOwner, borrow, err := ReferenceInfo(r)
		if err != nil || gotTable != table || gotType != typ || id != uint32(i+1) || gotOwner != owner || borrow {
			t.Fatal("neighboring publication changed a retained binding", i, id, gotOwner, err)
		}
	}
	if _, err := allocator.Move(nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := allocator.Move(&Reference{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestReferenceAllocatorDoesNotReviveBorrowsOrInvalidatedNeighbors(t *testing.T) {
	var allocator ReferenceAllocator
	scope := NewBorrowScope(4)
	borrowed := allocator.New(nil, nil, 9, nil, scope)
	alias := *borrowed
	if _, err := allocator.Move(borrowed); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("borrowed capability moved", err)
	}
	neighbor := allocator.New(nil, nil, 10, nil, nil)
	InvalidateReference(neighbor)
	if err := scope.End(); err != nil {
		t.Fatal(err)
	}
	allocator.Clear()
	for i := range 17 {
		fresh := allocator.New(nil, nil, uint32(i+20), nil, nil)
		if _, _, _, _, _, err := ReferenceInfo(fresh); err != nil {
			t.Fatal(err)
		}
	}
	runtime.GC()
	for _, stale := range []*Reference{borrowed, &alias, neighbor} {
		if _, _, _, _, _, err := ReferenceInfo(stale); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("expired capability revived after allocator clear", err)
		}
	}
}
