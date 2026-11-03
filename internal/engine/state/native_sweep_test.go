package state

import (
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Check the complete inventory after collection or capacity maintenance, when
// no retired records remain. In particular, empty objects must participate.
func checkNativeInventory(t *testing.T, s *Store) {
	t.Helper()
	seen := make(map[uint64]bool)
	var used uint64
	for _, a := range s.nativeHeap.objects {
		o, ok := s.heapObject(a.handle)
		if !ok || seen[a.handle] || a.offset != used || len(o.Data) == 0 {
			t.Fatal("invalid allocation inventory", a)
		}
		seen[a.handle] = true
		if unsafe.SliceData(o.Data) != &s.nativeHeap.data[used] {
			t.Fatal("payload not rebound", a.handle)
		}
		used += uint64(len(o.Data))
	}
	wantEmpty := make([]uint64, len(s.nativeHeap.empty))
	for i, slot := range s.heap {
		o := slot.Object
		if o == nil {
			continue
		}
		ref := heapTag | uint64(slot.Generation)<<32 | uint64(i)
		address := uint64(0)
		if len(o.Data) == 0 {
			if o.Data != nil || i/64 >= len(wantEmpty) || seen[ref] {
				t.Fatal("empty object missing or duplicated", i)
			}
			wantEmpty[i/64] |= uint64(1) << (i % 64)
			seen[ref] = true
		} else {
			address = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(o.Data))))
		}
		if !seen[ref] {
			t.Fatal("live object missing from sweep inventory", i)
		}
		v := s.nativeHeap.views[i*4 : i*4+4]
		if v[0] != ref || v[1] != address || v[2] != uint64(o.Length) || v[3] != 0 {
			t.Fatal("incoherent native view", ref, v)
		}
	}
	for page, word := range wantEmpty {
		if s.nativeHeap.empty[page] != word {
			t.Fatal("empty bitmap exposes a dead or missing handle", page)
		}
	}
	if uint64(len(seen)) != s.Usage().HeapObjects || used != s.nativeHeap.used || uint64(len(wantEmpty)) > (uint64(s.limits.MaxHeapObjects)+64)/64 {
		t.Fatal("inventory accounting", len(seen), used, s.Usage())
	}
}

func TestNativeSweepIncludesEmptySlowAndBatchObjects(t *testing.T) {
	s, types := allocationStore(t)
	var refs []uint64
	var roots []ReferenceValue
	var expectedBytes uint64
	add := func(ref uint64, keep bool) {
		refs = append(refs, ref)
		if keep {
			roots = append(roots, ReferenceValue{wasm.AnyRef, value.Value{Lo: ref}})
			o, _ := s.heapObject(ref)
			expectedBytes += o.Bytes
		}
	}
	// A dead prefix forces payload movement past both empty and nonempty
	// survivors; batch publication also leaves unused reservations.
	for i, local := range []int{0, 1, 0} {
		ref, err := s.allocateObject(types[local], uint32(2*(1-local)))
		if err != nil {
			t.Fatal(err)
		}
		add(ref, i != 0)
	}
	if _, err := s.PrepareNativeAllocations(65); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 61; i++ {
		stagedAllocation(t, s, uint32(i%2), uint32(i), 0)
	}
	if len(s.nativeHeap.objects) != 2 || s.nativeHeap.empty[0] != 1<<2 {
		t.Fatal("unpublished reservations entered sweep inventory")
	}
	words := s.nativeHeap.allocationWords
	first := nativeabi.HeapAllocationHeaderBytes / 8
	batch := append([]uint64(nil), words[first:first+61]...)
	if err := s.FinishNativeAllocations(types); err != nil {
		t.Fatal(err)
	}
	for i, ref := range batch {
		add(ref, i%3 == 0)
	}
	// Tagged zero-argument exceptions also have no payload bytes.
	tag, err := s.NewTag(module.FuncType{})
	if err != nil {
		t.Fatal(err)
	}
	ex, err := s.NewException(tag.Tag, nil)
	if err != nil {
		t.Fatal(err)
	}
	add(ex.Ref, true)
	roots[len(roots)-1].Type = wasm.ExnRef
	checkNativeInventory(t, s)
	s.CollectReferences(roots, nil, nil)
	checkNativeInventory(t, s)
	if s.Usage().HeapObjects != uint64(len(roots)) || s.Usage().HeapBytes != expectedBytes || s.heapThreshold != max(1<<20, expectedBytes*2) {
		t.Fatal("sweep did not restore quota/threshold", s.Usage())
	}
	for _, ref := range refs {
		keep := false
		for _, root := range roots {
			keep = keep || root.Value.Lo == ref
		}
		if s.IsLiveReference(ref) != keep {
			t.Fatal("wrong generation swept", ref, keep)
		}
		if keep {
			o, _ := s.heapObject(ref)
			if o.Mark {
				t.Fatal("mark survived collection")
			}
		}
	}
	// Repeated collection must trace survivors again, including empty objects.
	s.CollectReferences(roots, nil, nil)
	checkNativeInventory(t, s)
	s.CollectReferences(nil, nil, nil)
	checkNativeInventory(t, s)
	if s.Usage().HeapBytes != 0 || freeHeapSlots(t, s) != len(s.heap)-1 {
		t.Fatal("collection retained objects or reservations", s.Usage())
	}
}

func TestNativeEmptyPayloadMaintenanceBoundsRollbackBitmap(t *testing.T) {
	s, types := allocationStore(t)
	for _, size := range []uint32{0, 1 << 18} {
		// Exercise both an arena-free heap and one with a large old capacity.
		if size != 0 {
			if err := s.reserveNativePayload(uint64(size)); err != nil {
				t.Fatal(err)
			}
		}
		for range 150 {
			ref, err := s.allocateObject(types[1], 0)
			if err != nil {
				t.Fatal(err)
			}
			s.rollbackHeap(0)
			if s.IsLiveReference(ref) || len(s.nativeHeap.objects) != 0 || len(s.nativeHeap.empty) != 1 || s.nativeHeap.empty[0] != 0 {
				t.Fatal("empty rollback retained a bit or relocation record")
			}
		}
	}
	checkNative := func() {
		t.Helper()
		// Grow forces tombstone maintenance without tracing roots. Live objects
		// retain marks and bytes; empty objects keep a zero native address.
		if err := s.reserveNativePayload(uint64(len(s.nativeHeap.data)) + 1); err != nil {
			t.Fatal(err)
		}
		checkNativeInventory(t, s)
	}
	for _, local := range []int{1, 0, 1, 0} {
		ref, err := s.allocateObject(types[local], uint32(2*(1-local)))
		if err != nil {
			t.Fatal(err)
		}
		o, _ := s.heapObject(ref)
		o.Mark = true
		if local == 0 {
			o.set(0, value.Value{Lo: 42})
		}
	}
	checkNative()
	for _, slot := range s.heap {
		o := slot.Object
		if o == nil {
			continue
		}
		if !o.Mark || len(o.Data) != 0 && o.get(0).Lo != 42 {
			t.Fatal("capacity maintenance altered live object")
		}
		o.Mark = false
	}
	s.CollectReferences(nil, nil, nil)
	checkNativeInventory(t, s)
}

func TestNativeSweepThresholdUsesSurvivingBytes(t *testing.T) {
	s, _ := allocationStore(t)
	types, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{
		Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}},
	}}}, s.limits.MaxTypes)
	if !ok {
		t.Fatal("type registry")
	}
	alive, err := s.allocateObject(types[0], 600000)
	if err != nil {
		t.Fatal(err)
	}
	dead, err := s.allocateObject(types[0], 700000)
	if err != nil {
		t.Fatal(err)
	}
	s.CollectReferences([]ReferenceValue{{wasm.AnyRef, value.Value{Lo: alive}}}, nil, nil)
	checkNativeInventory(t, s)
	if s.IsLiveReference(dead) || s.Usage().HeapBytes != 600064 || s.heapThreshold != 1200128 {
		t.Fatal("threshold includes swept bytes", s.heapThreshold, s.Usage())
	}
	s.CollectReferences(nil, nil, nil)
	if s.heapThreshold != 1<<20 || s.Usage().HeapBytes != 0 {
		t.Fatal("empty heap threshold", s.heapThreshold, s.Usage())
	}
}
