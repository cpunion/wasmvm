package state

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeHeapRelocationCollectionAndGenerations(t *testing.T) {
	s := NewNative(DefaultLimits())
	defer s.Close()
	m := &module.Module{Types: []module.FuncType{{}, {}}, TypeDefs: []module.TypeDef{
		{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}}},
		{Kind: module.StructKind},
	}}
	ids, _ := s.registry.InternModule(m, s.limits.MaxTypes)
	if !native.Available {
		if _, err := s.allocateObject(ids[0], 1); !errors.Is(err, native.ErrUnavailable) {
			t.Fatal(err)
		}
		return
	}
	var roots []ReferenceValue
	var objects []*heapObject
	var handles []uint64
	for i := range 150 {
		ref, err := s.allocateObject(ids[0], uint32(47+i))
		if err != nil {
			t.Fatal(err)
		}
		o, _ := s.heapObject(ref)
		s.setObject(o, 0, value.Value{Lo: uint64(i)})
		s.setObject(o, o.Length-1, value.Value{Lo: uint64(i + 1)})
		objects = append(objects, o)
		handles = append(handles, ref)
		if i%3 != 0 {
			roots = append(roots, ReferenceValue{Type: wasm.AnyRef, Value: value.Value{Lo: ref}})
		}
	}
	// Drop interleaved objects, then move every remaining payload. The same
	// handles and portable getters must observe the republished native bytes.
	s.CollectReferences(roots, nil, nil)
	var used uint64
	for i, o := range objects {
		ref := handles[i]
		at := int(uint32(ref)) * 4
		if i%3 == 0 {
			if o.Data != nil || s.nativeHeap.views[at] != 0 || s.IsLiveReference(ref) {
				t.Fatal("unreachable slot retained", i)
			}
			continue
		}
		if o.Data != &s.nativeHeap.data[used] || s.getObject(o, 0).Lo != uint64(i) || s.getObject(o, o.Length-1).Lo != uint64(i+1) {
			t.Fatal("relocation corrupted payload", i, used)
		}
		used += uint64(int(s.objectSize(o)))
		address := uint64(uintptr(unsafe.Pointer(o.Data)))
		if s.nativeHeap.views[at] != ref || s.nativeHeap.views[at+1] != address || s.nativeHeap.views[at+2] != uint64(o.Type)<<32|uint64(o.Length) {
			t.Fatal("native view not refreshed", i)
		}
	}
	if s.nativeHeap.used != used || len(s.nativeHeap.objects) != len(roots) {
		t.Fatal("dead payloads retained")
	}
	for range 20 {
		ref, err := s.allocateObject(ids[0], 8)
		if err != nil {
			t.Fatal(err)
		}
		o, _ := s.heapObject(ref)
		for _, b := range s.objectData(o) {
			if b != 0 {
				t.Fatal("reused bytes not zeroed")
			}
		}
	}
	// Empty structs publish live identities without exposing a data address.
	empty, err := s.allocateObject(ids[1], 0)
	if err != nil || s.nativeHeap.views[int(uint32(empty))*4+1] != 0 {
		t.Fatal(empty, err)
	}
	count, address := s.NativeHeapView()
	if count != uint64(len(s.heap)) || address == 0 {
		t.Fatal(count, address)
	}
	payload, slots := s.nativeHeap.payload, s.nativeHeap.slots
	s.Close()
	for _, storage := range []*native.Memory{payload, slots} {
		if _, err := storage.Buffer(); !errors.Is(err, native.ErrClosed) {
			t.Fatal("native allocation leaked", err)
		}
	}
	if count, address := s.NativeHeapView(); count != 0 || address != 0 || len(s.nativeHeap.objects) != 0 {
		t.Fatal("closed heap exposed")
	}
}

func TestNativeHeapExactQuotasAndHandleRetirement(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	limits := DefaultLimits()
	limits.MaxHeapObjects, limits.MaxHeapBytes = 1, 65
	s := NewNative(limits)
	defer s.Close()
	m := &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}}}}}
	ids, _ := s.registry.InternModule(m, limits.MaxTypes)
	if count, address := s.NativeHeapView(); count != 0 || address != 0 {
		t.Fatal("unallocated heap")
	}
	var previous uint64
	for range 150 {
		ref, err := s.allocateObject(ids[0], 1)
		if err != nil || ref == previous || previous != 0 && s.IsLiveReference(previous) {
			t.Fatal(ref, previous, err)
		}
		o, _ := s.heapObject(ref)
		if s.getObject(o, 0).Lo != 0 {
			t.Fatal("freed contents exposed")
		}
		s.setObject(o, 0, value.Value{Lo: 255})
		if _, err := s.allocateObject(ids[0], 0); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("object quota", err)
		}
		s.releaseObject(uint32(ref))
		previous = ref
	}
	if len(s.nativeHeap.objects) > 64 {
		t.Fatal("retired allocation metadata accumulated")
	}
	if _, err := s.allocateObject(ids[0], 2); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("payload quota", err)
	}
	ref, _ := s.allocateObject(ids[0], 0)
	s.heap[uint32(ref)].Generation = 0x7fffffff
	s.releaseObject(uint32(ref))
	before := s.nativeHeap.used
	if _, err := s.allocateObject(ids[0], 1); !errors.Is(err, wasm.ErrLimit) || s.nativeHeap.used != before {
		t.Fatal("retired handle consumed payload or wrapped", err)
	}
	plain := New(limits)
	defer plain.Close()
	if count, address := plain.NativeHeapView(); count != 0 || address != 0 {
		t.Fatal("portable heap exposed a host address")
	}
	// A large historical payload must not allow small failed instances to keep
	// an arena-capacity-sized list of retired host objects alive.
	limits.MaxHeapBytes = 1 << 20
	large := NewNative(limits)
	defer large.Close()
	ids, _ = large.registry.InternModule(m, limits.MaxTypes)
	if _, err := large.allocateObject(ids[0], 1<<18); err != nil {
		t.Fatal(err)
	}
	large.rollbackHeap(0)
	for range 150 {
		if _, err := large.allocateObject(ids[0], 1); err != nil {
			t.Fatal(err)
		}
		large.rollbackHeap(0)
	}
	if len(large.nativeHeap.objects) > 64 || large.Usage().HeapObjects != 0 {
		t.Fatal("large arena retained failed allocations", len(large.nativeHeap.objects))
	}
}
