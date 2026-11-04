package state

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func freeHeapSlots(t *testing.T, s *Store) int {
	t.Helper()
	seen := make([]bool, len(s.heap))
	count := 0
	for index := s.heapFree; index != 0; index = s.heap[index].Next {
		if uint64(index) >= uint64(len(s.heap)) || seen[index] {
			t.Fatal("free chain contains an invalid index or cycle", index)
		}
		slot := s.heap[index]
		if slot.Object != nil || slot.Generation == 0 || slot.Generation > 0x7fffffff {
			t.Fatal("free chain exposes a live object or invalid generation", index)
		}
		seen[index] = true
		count++
	}
	for i, slot := range s.heap {
		if slot.Object != nil && slot.Next != 0 {
			t.Fatal("live slot retains a free-chain link", i)
		}
	}
	return count
}

func TestHeapFreeChainFailureReuseAndRetirement(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		if offheap && !native.Available {
			continue
		}
		limits := DefaultLimits()
		limits.MaxHeapObjects, limits.MaxHeapBytes = 3, 256
		s := New(limits)
		if offheap {
			s = NewNative(limits)
		}
		t.Cleanup(s.Close)
		ids, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{
			Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32}},
		}}}, limits.MaxTypes)
		if !ok {
			t.Fatal("type registry")
		}
		var refs [3]uint64
		for i := range refs {
			ref, err := s.allocateObject(ids[0], 1)
			if err != nil {
				t.Fatal(err)
			}
			refs[i] = ref
			o, _ := s.heapObject(ref)
			s.setObject(o, 0, value.Value{Lo: uint64(42 + i)})
		}
		s.releaseObject(uint32(refs[1]))
		s.releaseObject(uint32(refs[0]))
		if freeHeapSlots(t, s) != 2 {
			t.Fatal("released slots lost")
		}
		if _, err := s.allocateObject(ids[0], 1000); !errors.Is(err, wasm.ErrLimit) || s.heapFree != uint32(refs[0]) || freeHeapSlots(t, s) != 2 {
			t.Fatal("failed allocation consumed or reordered free slots", offheap, err)
		}
		for _, old := range refs[:2] {
			ref, err := s.allocateObject(ids[0], 1)
			if err != nil || uint32(ref) != uint32(old) || ref == old || s.IsLiveReference(old) {
				t.Fatal("LIFO reuse or stale-generation check", offheap, old, ref, err)
			}
			o, _ := s.heapObject(ref)
			if s.getObject(o, 0) != (value.Value{}) {
				t.Fatal("old payload exposed")
			}
		}
		for _, ref := range []uint64{refs[0], refs[2]} {
			s.heap[uint32(ref)].Generation = 0x7fffffff
			s.releaseObject(uint32(ref))
		}
		if freeHeapSlots(t, s) != 0 {
			t.Fatal("retired identity reentered free chain")
		}
		if _, err := s.allocateObject(ids[0], 0); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("retired identity wrapped", err)
		}
		s.releaseObject(uint32(refs[1]))
		ref, err := s.allocateObject(ids[0], 0)
		if err != nil || uint32(ref) != uint32(refs[1]) || freeHeapSlots(t, s) != 0 {
			t.Fatal("available identity lost beside retired slots", ref, err)
		}
		s.Close()
		if s.heapFree != 0 || freeHeapSlots(t, s) != 0 {
			t.Fatal("closed Store retained free slots")
		}
	}
}

func TestNativeFreeChainReturnsInvalidAndUnusedReservations(t *testing.T) {
	s, types := allocationStore(t)
	old, err := s.allocateObject(types[0], 2)
	if err != nil {
		t.Fatal(err)
	}
	before, serial := s.Usage(), s.heapSerial
	if _, err := s.PrepareNativeAllocations(65); err != nil {
		t.Fatal(err)
	}
	first := nativeabi.HeapAllocationHeaderBytes / 8
	reserved := append([]uint64(nil), s.nativeHeap.allocationWords[first:first+65]...)
	for i := 0; i < 10; i++ {
		stagedAllocation(t, s, uint32(i%2), uint32(i), 0)
	}
	s.nativeHeap.views[uint64(uint32(reserved[9]))*4+1]++
	if err := s.FinishNativeAllocations(types); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("invalid batch accepted", err)
	}
	if s.Usage() != before || s.heapSerial != serial || !s.IsLiveReference(old) || freeHeapSlots(t, s) != 65 {
		t.Fatal("invalid prefix publication changed the live heap", s.Usage())
	}
	for _, ref := range reserved {
		if s.IsLiveReference(ref) {
			t.Fatal("invalid batch published an object", ref)
		}
	}
	if _, err := s.PrepareNativeAllocations(65); err != nil {
		t.Fatal(err)
	}
	for i, ref := range reserved {
		if s.nativeHeap.allocationWords[first+64-i] != ref {
			t.Fatal("unpublished reservation changed its generation or was lost")
		}
	}
	roots := []ReferenceValue{{wasm.AnyRef, value.Value{Lo: old}}}
	for i := 0; i < 13; i++ {
		ref := stagedAllocation(t, s, uint32(i%2), uint32(i), 0)
		roots = append(roots, ReferenceValue{wasm.AnyRef, value.Value{Lo: ref}})
	}
	if err := s.FinishNativeAllocations(types); err != nil || freeHeapSlots(t, s) != 52 {
		t.Fatal("partial batch lost unused reservations", err)
	}
	s.CollectReferences(roots, nil, nil)
	if s.Usage().HeapObjects != 14 || freeHeapSlots(t, s) != 52 {
		t.Fatal("mixed-size batch lost rooted objects", s.Usage())
	}
	s.CollectReferences(nil, nil, nil)
	if s.Usage().HeapObjects != 0 || freeHeapSlots(t, s) != 66 {
		t.Fatal("collection lost free slots", s.Usage())
	}
}
