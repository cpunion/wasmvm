package state

import (
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func compactStore(t *testing.T, offheap bool) (*Store, uint32) {
	t.Helper()
	if offheap && !native.Available {
		t.Skip("native platform unavailable")
	}
	s := New(DefaultLimits())
	if offheap {
		s = NewNative(DefaultLimits())
	}
	t.Cleanup(s.Close)
	ids, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.StructKind}}}, s.limits.MaxTypes)
	if !ok {
		t.Fatal("type registry")
	}
	return s, ids[0]
}

func compactObject(t *testing.T, s *Store, typ uint32) uint64 {
	t.Helper()
	ref, err := s.allocateObject(typ, 0)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

func TestHeapEpochNestedRollbackSurvivesCollectionAndReuse(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		s, typ := compactStore(t, offheap)
		old := compactObject(t, s, typ)
		before := s.Usage()
		outer := s.heapSerial
		parent := compactObject(t, s, typ)
		inner := s.heapSerial
		retired := compactObject(t, s, typ)
		s.releaseObject(uint32(retired))
		reused := compactObject(t, s, typ)
		if uint32(retired) != uint32(reused) || retired == reused {
			t.Fatal("fixture did not reuse a generation")
		}
		roots := []ReferenceValue{{Type: wasm.AnyRef, Value: value.Value{Lo: old}}, {Type: wasm.AnyRef, Value: value.Value{Lo: parent}}, {Type: wasm.AnyRef, Value: value.Value{Lo: reused}}}
		s.CollectReferences(roots, nil, nil)
		s.rollbackHeap(inner)
		if s.IsLiveReference(reused) || !s.IsLiveReference(parent) || !s.IsLiveReference(old) {
			t.Fatal("nested rollback changed its parent")
		}
		committed := compactObject(t, s, typ)
		if !s.IsLiveReference(committed) {
			t.Fatal("nested commit was not retained")
		}
		s.rollbackHeap(outer)
		if !s.IsLiveReference(old) || s.IsLiveReference(parent) || s.IsLiveReference(committed) || s.Usage() != before {
			t.Fatal("parent rollback lost preexisting state", s.Usage(), before)
		}
	}
}

func TestHeapEpochNativeBatchPreservesSerialExhaustion(t *testing.T) {
	s, types := allocationStore(t)
	old := compactObject(t, s, types[1])
	s.heapSerial = math.MaxUint64 - 2
	mark := s.heapSerial
	if err := s.reserveNativePayload(12); err != nil {
		t.Fatal(err)
	}
	if address, err := s.PrepareNativeAllocations(2); err != nil || address == 0 {
		t.Fatal(address, err)
	}
	first := stagedAllocation(t, s, 0, 42, old)
	second := stagedAllocation(t, s, 1, 0, 0)
	if err := s.FinishNativeAllocations(types); err != nil || s.heapSerial != math.MaxUint64 {
		t.Fatal("batch lost allocation epochs", err)
	}
	if address, err := s.PrepareNativeAllocations(nativeabi.HeapAllocationBatch); err != nil || address != 0 {
		t.Fatal("serial limit admitted reservations", address, err)
	}
	s.rollbackHeap(mark)
	if !s.IsLiveReference(old) || s.IsLiveReference(first) || s.IsLiveReference(second) || s.Usage().HeapObjects != 1 || s.heapSerial != math.MaxUint64 {
		t.Fatal("batch rollback changed preexisting object or allocation epoch")
	}
	if _, err := s.allocateObject(types[1], 0); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("rollback rewound serial exhaustion", err)
	}
}

func TestHeapEpochFollowsResolverErrorsAndPanicCleanup(t *testing.T) {
	for _, offheap := range []bool{false, true} {
		s, _ := compactStore(t, offheap)
		host, err := s.NewHost(module.FuncType{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		tag, err := s.NewTag(module.FuncType{})
		if err != nil {
			t.Fatal(err)
		}
		old, err := s.NewException(tag.Tag, nil)
		if err != nil {
			t.Fatal("exception allocation failed", err)
		}
		def := &module.Module{Types: []module.FuncType{{}}, Imports: []module.Import{{Kind: 0, Name: "host"}}}
		compiled := build(t, def)
		before := s.Usage()
		var parent, child uint64
		_, err = s.Allocate(compiled, func(string, string) (Extern, bool) {
			ex, e := s.NewException(tag.Tag, nil)
			if e != nil {
				t.Fatal(e)
			}
			parent = ex.Ref
			_, e = s.Allocate(compiled, func(string, string) (Extern, bool) {
				ex, e := s.NewException(tag.Tag, nil)
				if e != nil {
					t.Fatal(e)
				}
				child = ex.Ref
				return Extern{}, false
			})
			if !errors.Is(e, ErrLink) || s.IsLiveReference(child) || !s.IsLiveReference(parent) {
				t.Fatal("nested resolver rollback", e)
			}
			return Extern{}, false
		})
		if !errors.Is(err, ErrLink) || s.IsLiveReference(parent) || !s.IsLiveReference(old.Ref) || s.Usage() != before {
			t.Fatal("resolver error leaked allocations", err, s.Usage(), before)
		}
		func() {
			defer func() {
				if recover() != "resolver panic" {
					t.Fatal("resolver panic changed")
				}
			}()
			s.Allocate(compiled, func(string, string) (Extern, bool) {
				s.NewException(tag.Tag, nil)
				panic("resolver panic")
			})
		}()
		if len(s.allocating) != 0 {
			t.Fatal("host panic retained initializer roots")
		}
		if _, err := s.Allocate(compiled, func(string, string) (Extern, bool) { return host, true }); err != nil {
			t.Fatal("Store could not allocate after resolver panic", err)
		}
	}
}
