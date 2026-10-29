package state

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
	"testing"
)

func TestHeapGenerationsRetirementAndQuotas(t *testing.T) {
	limits := DefaultLimits()
	limits.MaxHeapObjects = 1
	limits.MaxHeapBytes = 64
	s := New(limits)
	defer s.Close()
	m := &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.StructKind}}}
	ids, ok := s.registry.InternModule(m, limits.MaxTypes)
	if !ok {
		t.Fatal("type")
	}
	first, err := s.allocateObject(ids[0], 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.allocateObject(ids[0], 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("object quota", err)
	}
	s.releaseObject(uint32(first))
	next, err := s.allocateObject(ids[0], 0)
	if err != nil {
		t.Fatal(err)
	}
	if uint32(first) != uint32(next) || first == next {
		t.Fatal("slot was not safely reused")
	}
	if _, ok := s.heapObject(first); ok {
		t.Fatal("stale generation accepted")
	}
	if s.IsLiveReference(first) {
		t.Fatal("stale liveness")
	}
	if _, ok := s.heapObject(next); !ok {
		t.Fatal("live object missing")
	}
	slot := &s.heap[uint32(next)]
	slot.Generation = 0x7fffffff
	s.releaseObject(uint32(next))
	if _, err := s.allocateObject(ids[0], 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("retired generation wrapped", err)
	}
	s2 := New(DefaultLimits())
	defer s2.Close()
	m2 := &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}}}}}
	ids2, _ := s2.registry.InternModule(m2, s2.limits.MaxTypes)
	if _, err := s2.allocateObject(ids2[0], math.MaxUint32); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("byte quota", err)
	}
	s2.heapSerial = math.MaxUint64
	if _, err := s2.allocateObject(ids2[0], 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("serial wrapped", err)
	}
}

func TestInternalExternalReferenceTracing(t *testing.T) {
	s := New(DefaultLimits())
	defer s.Close()
	refs := []ReferenceValue{{Type: wasm.AnyRef, Value: value.Value{Lo: Internalize(123)}}, {Type: wasm.ExternRef, Value: value.Value{Lo: 456}}, {Type: wasm.I31Ref, Value: value.Value{Lo: I31(0)}}, {Type: wasm.AnyRef, Value: value.Value{Lo: 0}}}
	marked := map[uint64]bool{}
	s.CollectReferences(refs, nil, func(id uint64) { marked[id] = true })
	if len(marked) != 2 || !marked[123] || !marked[456] {
		t.Fatal(marked)
	}
	if Externalize(Internalize(123)) != 123 || Internalize(0) != 0 || Externalize(I31(0)) != I31(0) || !IsExternalizedGuest(I31(0)) {
		t.Fatal("reference conversion")
	}
}

func TestCompactHeapFieldAndArrayStorage(t *testing.T) {
	s := New(DefaultLimits())
	defer s.Close()
	m := &module.Module{Types: []module.FuncType{{}, {}}, TypeDefs: []module.TypeDef{
		{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I64}, {Value: wasm.V128}, {Value: wasm.ExternRef}}},
		{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}}},
	}}
	ids, ok := s.registry.InternModule(m, s.limits.MaxTypes)
	if !ok {
		t.Fatal("type")
	}
	ref, err := s.allocateObject(ids[0], 4)
	if err != nil {
		t.Fatal(err)
	}
	o, _ := s.heapObject(ref)
	if len(o.Data) != 33 {
		t.Fatal("struct fields padded to frame width", len(o.Data))
	}
	values := []value.Value{{Lo: 255}, {Lo: math.MaxUint64}, {Lo: 3, Hi: 7}, {Lo: 123}}
	for i, v := range values {
		o.set(uint32(i), v)
		if got := o.get(uint32(i)); got != v {
			t.Fatal(i, got, v)
		}
	}
	ref, err = s.allocateObject(ids[1], 1024)
	if err != nil {
		t.Fatal(err)
	}
	o, _ = s.heapObject(ref)
	if len(o.Data) != 1024 || o.Bytes != 1088 {
		t.Fatal("packed array storage", len(o.Data), o.Bytes)
	}
	o.set(1023, value.Value{Lo: 0x123})
	if o.get(1023).Lo != 0x23 {
		t.Fatal("packed truncation")
	}
}
