package state

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativePayloadRunsAcrossDeadGapsAndReusedHandles(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	s := NewNative(DefaultLimits())
	defer s.Close()
	types, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{
		Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}},
	}}}, s.limits.MaxTypes)
	if !ok {
		t.Fatal("type registry")
	}
	var roots []ReferenceValue
	var expected [][]byte
	var dead []uint64
	for round := range 3 {
		for i := range 27 {
			// Keep long adjacent runs, isolated survivors, a final dead suffix,
			// and zero-byte objects. Later rounds recycle nonsequential handles.
			length := uint32((i*7 + round*11) % 29)
			ref, err := s.allocateObject(types[0], length)
			if err != nil {
				t.Fatal(err)
			}
			o, _ := s.heapObject(ref)
			for j := range o.Length {
				s.setObject(o, j, value.Value{Lo: uint64(i*31 + int(j) + round)})
			}
			if i == 0 || i >= 24 || i%7 == 0 || round == 1 && i%3 == 0 {
				dead = append(dead, ref)
			} else {
				roots = append(roots, ReferenceValue{wasm.AnyRef, value.Value{Lo: ref}})
				expected = append(expected, append([]byte(nil), s.objectData(o)...))
			}
		}
		// Capacity maintenance must preserve marks and all live payloads even
		// when rollback leaves holes before the next full collection.
		if round == 1 {
			for _, ref := range dead {
				if s.IsLiveReference(ref) {
					s.releaseObject(uint32(ref))
				}
			}
			s.compactNativeHeap(false)
		}
		s.CollectReferences(roots, nil, nil)
		checkNativeInventory(t, s)
		for i, r := range roots {
			o, ok := s.heapObject(r.Value.Lo)
			if !ok || string(s.objectData(o)) != string(expected[i]) || s.heap[uint32(o.Ref)].Next&heapMarked != 0 {
				t.Fatal("run relocation changed a survivor", round, i)
			}
		}
		for _, ref := range dead {
			if s.IsLiveReference(ref) {
				t.Fatal("dead generation retained", round, ref)
			}
		}
	}
	s.CollectReferences(nil, nil, nil)
	checkNativeInventory(t, s)
}
