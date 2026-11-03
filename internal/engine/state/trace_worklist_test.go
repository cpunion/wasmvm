package state

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTraceWorklistBoundsDeepSharedCyclesAndReleasesPointers(t *testing.T) {
	const count = 4096
	for _, offheap := range []bool{false, true} {
		if offheap && !native.Available {
			continue
		}
		s := New(DefaultLimits())
		if offheap {
			s = NewNative(DefaultLimits())
		}
		t.Cleanup(s.Close)
		m := &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{
			Kind: module.StructKind,
			Fields: []module.FieldType{{Value: wasm.I64}, {Value: wasm.AnyRef},
				{Value: wasm.AnyRef}, {Value: wasm.ExternRef}},
		}}}
		types, ok := s.registry.InternModule(m, s.limits.MaxTypes)
		if !ok {
			t.Fatal("type registry")
		}
		allocate := func(external uint64) uint64 {
			ref, err := s.allocateObject(types[0], 4)
			if err != nil {
				t.Fatal(err)
			}
			o, _ := s.heapObject(ref)
			o.set(3, value.Value{Lo: external})
			return ref
		}
		orphan, shared := allocate(99), allocate(77)
		var head uint64
		for i := 0; i < count; i++ {
			ref := allocate(uint64(1000 + i))
			o, _ := s.heapObject(ref)
			// Numeric bits must not keep an otherwise unreachable object alive.
			o.set(0, value.Value{Lo: orphan})
			o.set(1, value.Value{Lo: head})
			o.set(2, value.Value{Lo: shared})
			head = ref
		}
		leaf, _ := s.heapObject(shared)
		leaf.set(1, value.Value{Lo: head})
		roots := []ReferenceValue{{wasm.AnyRef, value.Value{Lo: head}},
			{wasm.AnyRef, value.Value{Lo: head}}}
		externals := make(map[uint64]int)
		s.CollectReferences(roots, nil, func(ref uint64) { externals[ref]++ })
		if s.Usage().HeapObjects != count+1 || s.IsLiveReference(orphan) {
			t.Fatal("reachable cycle or numeric-only orphan", offheap, s.Usage())
		}
		if len(externals) != count+1 || externals[77] != 1 || externals[99] != 0 {
			t.Fatal("shared child visited repeatedly or orphan traced", offheap, len(externals))
		}
		for i := 0; i < count; i++ {
			if externals[uint64(1000+i)] != 1 {
				t.Fatal("lost external reference", offheap, i)
			}
		}
		if len(s.markObjects) != 0 || cap(s.markObjects) > 32 {
			t.Fatal("trace scratch space retained the whole chain", offheap, len(s.markObjects), cap(s.markObjects))
		}
		for _, object := range s.markObjects[:cap(s.markObjects)] {
			if object != nil {
				t.Fatal("trace scratch space retained an object")
			}
		}
		s.CollectReferences(nil, nil, nil)
		if s.Usage().HeapObjects != 0 || s.IsLiveReference(head) || s.IsLiveReference(shared) {
			t.Fatal("unrooted cycle survived", offheap, s.Usage())
		}
	}
}
