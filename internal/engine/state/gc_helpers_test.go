package state

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGCAllocationRestoresBorrowedRootsAfterCollectorPanic(t *testing.T) {
	s := New(DefaultLimits())
	defer s.Close()
	ids, ok := s.registry.InternModule(&module.Module{
		Types:    []module.FuncType{{}},
		TypeDefs: []module.TypeDef{{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.AnyRef}}}},
	}, s.limits.MaxTypes)
	if !ok {
		t.Fatal("type registry")
	}
	child, err := s.allocateObject(ids[0], 1)
	if err != nil {
		t.Fatal(err)
	}
	in := &Instance{Types: ids}
	ins := &module.Instr{Op: 0xfb0000}
	args := []value.Value{{Lo: child}}
	s.temporary = []ReferenceValue{{Type: wasm.AnyRef, Value: value.Value{Lo: I31(7)}}}
	s.CollectEveryAllocation = true
	s.HeapSafepoint = func() {
		if len(s.temporary) != 2 || s.temporary[1].Value.Lo != child {
			t.Fatal("constructor operand was not rooted")
		}
		s.CollectReferences(s.temporary, nil, nil)
		panic("collector interrupted")
	}
	func() {
		defer func() {
			if got := recover(); got != "collector interrupted" {
				t.Fatal("collector panic lost", got)
			}
		}()
		_, _ = s.GC(context.Background(), in, ins, args)
	}()
	if len(s.temporary) != 1 || s.temporary[0].Value.Lo != I31(7) || s.temporary[:cap(s.temporary)][1] != (ReferenceValue{}) {
		t.Fatal("borrowed roots survived collector panic", s.temporary)
	}
	s.HeapSafepoint = nil
	s.CollectEveryAllocation = false
	parent, err := s.GC(context.Background(), in, ins, args)
	if err != nil {
		t.Fatal(err)
	}
	o, ok := s.heapObject(parent.Lo)
	if !ok || s.getObject(o, 0).Lo != child || !s.IsLiveReference(child) || len(s.temporary) != 1 || s.existingObjectExtras(o) != nil {
		t.Fatal("constructor could not recover after collector panic")
	}
}

func TestDefaultGCAllocationPreservesOuterRootsAndCollectorFailure(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		s, typ := compactStore(t, offheap)
		old := compactObject(t, s, typ)
		s.temporary = []ReferenceValue{{Type: wasm.AnyRef, Value: value.Value{Lo: old}}}
		in := &Instance{Types: []uint32{typ}}
		ins := &module.Instr{Op: 0xfb0001}
		s.CollectEveryAllocation = true
		s.HeapSafepoint = func() {
			s.CollectReferences(s.temporary, nil, nil)
			panic("default collector interrupted")
		}
		before := s.Usage()
		func() {
			defer func() {
				if recover() != "default collector interrupted" {
					t.Fatal("default constructor changed collector panic")
				}
			}()
			s.GC(context.Background(), in, ins, nil)
		}()
		if s.Usage() != before || len(s.temporary) != 1 || s.temporary[0].Value.Lo != old || !s.IsLiveReference(old) {
			t.Fatal("default constructor changed outer roots or published a failed object")
		}
		s.CollectEveryAllocation = false
		s.HeapSafepoint = func() {}
		s.limits.MaxHeapObjects = 1
		if _, err := s.GC(context.Background(), in, ins, nil); !errors.Is(err, wasm.ErrLimit) || s.Usage() != before || len(s.temporary) != 1 {
			t.Fatal("default constructor lost quota or outer roots", err)
		}
		s.limits.MaxHeapObjects = 2
		v, err := s.GC(context.Background(), in, ins, nil)
		if err != nil || !s.IsLiveReference(v.Lo) || !s.IsLiveReference(old) || len(s.temporary) != 1 {
			t.Fatal("default constructor did not recover after collector failure", err)
		}
	}
}
