package state

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestObjectMetadataReusesGenerationsAcrossPagesAndRelocation(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		if offheap && !native.Available {
			continue
		}
		limits := DefaultLimits()
		limits.MaxHeapObjects = 133
		s := New(limits)
		if offheap {
			s = NewNative(limits)
		}
		t.Cleanup(s.Close)
		m := &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32}}}}}
		types, ok := s.registry.InternModule(m, limits.MaxTypes)
		if !ok {
			t.Fatal("type registry quota")
		}
		var handles []uint64
		var objects []*heapObject
		for i := 0; i < 133; i++ {
			ref, err := s.allocateObject(types[0], 17)
			if err != nil {
				t.Fatal(offheap, i, err)
			}
			o, _ := s.heapObject(ref)
			o.set(0, value.Value{Lo: uint64(i)})
			handles, objects = append(handles, ref), append(objects, o)
		}
		if len(s.heapObjects) != 3 || len(s.heapObjects[2].objects) != 6 {
			t.Fatal("metadata exceeds handle quota", len(s.heapObjects))
		}
		for round := 0; round < 3; round++ {
			for i := 0; i < len(handles); i += 2 {
				old := handles[i]
				objects[i].Tag = &Tag{}
				objects[i].Descriptor = handles[0]
				objects[i].Mark = true
				s.releaseObject(uint32(old))
				if s.IsLiveReference(old) || objects[i].Tag != nil || objects[i].Descriptor != 0 || objects[i].Mark || objects[i].Layout != nil || objects[i].Continuation != nil || objects[i].Data != nil {
					t.Fatal("dead metadata retained references", old)
				}
				ref, err := s.allocateObject(types[0], uint32(1+(i+round)%23))
				if err != nil || ref == old || uint32(ref) != uint32(old) {
					t.Fatal("handle generation reuse", old, ref, err)
				}
				o, _ := s.heapObject(ref)
				if o != objects[i] {
					t.Fatal("metadata address moved")
				}
				for j := uint32(0); j < o.Length; j++ {
					if o.get(j) != (value.Value{}) {
						t.Fatal("new generation exposed old bytes", i, j)
					}
				}
				o.set(0, value.Value{Lo: uint64(i)})
				handles[i] = ref
			}
			roots := make([]ReferenceValue, len(handles))
			for i, ref := range handles {
				roots[i] = ReferenceValue{Type: wasm.AnyRef, Value: value.Value{Lo: ref}}
			}
			s.CollectReferences(roots, nil, nil)
			for i, ref := range handles {
				o, ok := s.heapObject(ref)
				if !ok || o != objects[i] || o.get(0).Lo != uint64(i) {
					t.Fatal("relocation followed a retired generation", offheap, i, ref)
				}
				if offheap && s.nativeHeap.views[int(uint32(ref))*4] != ref {
					t.Fatal("retired handle republished", i)
				}
			}
			if len(s.heapObjects) != 3 || offheap && len(s.nativeHeap.objects) != len(handles) {
				t.Fatal("metadata accumulated across generations")
			}
		}
		s.CollectReferences(nil, nil, nil)
		if s.Usage().HeapObjects != 0 {
			t.Fatal("unrooted objects survived")
		}
		s.Close()
		if s.heapObjects != nil {
			t.Fatal("closed Store retained metadata pages")
		}
	}
}
