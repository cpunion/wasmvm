package state

import (
	"testing"
	"unsafe"
)

func TestMetadataSlabsPreserveCellsOptionalStorageAndQuota(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		s, typ := compactStore(t, offheap)
		s.limits.MaxHeapObjects = 1100
		var refs []uint64
		var cells []*heapObject
		for range 1025 {
			ref := compactObject(t, s, typ)
			o, _ := s.heapObject(ref)
			refs, cells = append(refs, ref), append(cells, o)
		}
		allocated := 0
		for page := range s.heapObjects {
			p := &s.heapObjects[page]
			allocated += len(p.objects)
			if len(p.objects) > heapObjectsPerPage || p.blockPages == 0 || p.blockPages > 8 || p.blockOffset >= p.blockPages || p.extras != nil {
				t.Fatal("invalid or eager metadata page", page)
			}
			first := page - int(p.blockOffset)
			if unsafe.Pointer(&p.objects[0]) != unsafe.Add(unsafe.Pointer(&s.heapObjects[first].objects[0]), uintptr(p.blockOffset)*heapObjectsPerPage*unsafe.Sizeof(heapObject{})) {
				t.Fatal("logical pages do not share stable slab storage")
			}
		}
		if allocated != int(s.limits.MaxHeapObjects)+1 || len(s.heapObjects[len(s.heapObjects)-1].objects) != 13 {
			t.Fatal("slab exceeded metadata quota", allocated)
		}
		// Pages one and two share the second slab. Touching an optional cell
		// materializes only that slab and keeps unrelated pages absent.
		left, right := refs[63], refs[127]
		a, b := s.objectExtras(left), s.objectExtras(right)
		tag := &Tag{Owner: s}
		a.Tag, b.Tag = &Tag{}, tag
		if s.heapObjects[0].extras != nil || s.heapObjects[3].extras != nil || len(s.heapObjects[1].extras) != 64 || len(s.heapObjects[2].extras) != 64 {
			t.Fatal("optional metadata crossed its slab")
		}
		if unsafe.Pointer(&s.heapObjects[2].extras[0]) != unsafe.Add(unsafe.Pointer(&s.heapObjects[1].extras[0]), heapObjectsPerPage*unsafe.Sizeof(heapObjectExtra{})) {
			t.Fatal("optional cells were allocated individually")
		}
		s.releaseObject(uint32(left))
		if *a != (heapObjectExtra{}) || b.Tag != tag || !s.IsLiveReference(right) {
			t.Fatal("release retained references or cleared a neighboring cell")
		}
		for i, ref := range refs {
			if ref == left {
				continue
			}
			if o, ok := s.heapObject(ref); !ok || o != cells[i] {
				t.Fatal("slab growth moved a live object")
			}
		}
	}
}
