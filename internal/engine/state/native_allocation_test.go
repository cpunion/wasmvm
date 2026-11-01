package state

import (
	"encoding/binary"
	"errors"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func allocationStore(t *testing.T) (*Store, []uint32) {
	t.Helper()
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	limits := DefaultLimits()
	limits.MaxHeapObjects = 140
	s := NewNative(limits)
	t.Cleanup(s.Close)
	types, ok := s.registry.InternModule(&module.Module{Types: []module.FuncType{{}, {}}, TypeDefs: []module.TypeDef{
		{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32}, {Value: wasm.AnyRef}}},
		{Kind: module.StructKind},
	}}, limits.MaxTypes)
	if !ok {
		t.Fatal("type registry")
	}
	return s, types
}

func stagedAllocation(t *testing.T, s *Store, local uint32, number uint32, next uint64) uint64 {
	t.Helper()
	h := s.nativeHeap
	w := h.allocationWords
	n, at := w[nativeabi.HeapAllocationCursor/8], w[nativeabi.HeapAllocationUsed/8]
	ref := w[nativeabi.HeapAllocationHeaderBytes/8+int(n)]
	length, size := uint64(2), uint64(12)
	if local == 1 {
		length, size = 0, 0
	}
	address := uint64(0)
	if size != 0 {
		binary.LittleEndian.PutUint32(h.data[at:at+4], number)
		binary.LittleEndian.PutUint64(h.data[at+4:at+12], next)
		address = uint64(uintptr(unsafe.Pointer(&h.data[at])))
	}
	index := uint64(uint32(ref)) * 4
	h.views[index], h.views[index+1], h.views[index+2], h.views[index+3] = ref, address, length, uint64(local)
	w[nativeabi.HeapAllocationUsed/8] += size
	w[nativeabi.HeapAllocationBudget/8] -= size + 64
	w[nativeabi.HeapAllocationCursor/8]++
	return ref
}

func TestNativeAllocationBatchesPublishRootsQuotasAndReuse(t *testing.T) {
	s, types := allocationStore(t)
	var last, first uint64
	for batch := 0; batch < 2; batch++ {
		address, err := s.PrepareNativeAllocations()
		if err != nil || address == 0 {
			t.Fatal(address, err)
		}
		if _, err := s.PrepareNativeAllocations(); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("overlapping batches", err)
		}
		for i := 0; i < nativeabi.HeapAllocationBatch; i++ {
			last = stagedAllocation(t, s, 0, uint32(batch*64+i), last)
			if first == 0 {
				first = last
			}
			if s.IsLiveReference(last) {
				t.Fatal("unpublished object visible to host")
			}
		}
		if err := s.FinishNativeAllocations(types); err != nil {
			t.Fatal(err)
		}
		if err := s.FinishNativeAllocations(types); err != nil {
			t.Fatal("idle finish", err)
		}
	}
	if s.Usage().HeapObjects != 128 || s.Usage().HeapBytes != 128*76 || s.nativeHeap.allocatedObjects != 128 || s.nativeHeap.allocationBatches != 2 {
		t.Fatal("batch accounting", s.Usage())
	}
	s.CollectReferences([]ReferenceValue{{Type: wasm.AnyRef, Value: value.Value{Lo: last}}}, nil, nil)
	for i := 127; i >= 0; i-- {
		o, ok := s.heapObject(last)
		if !ok || o.get(0).Lo != uint64(i) {
			t.Fatal("batch root chain lost", i, last)
		}
		last = o.get(1).Lo
	}
	if last != 0 {
		t.Fatal("unexpected chain tail")
	}
	s.CollectReferences(nil, nil, nil)
	if _, err := s.PrepareNativeAllocations(); err != nil {
		t.Fatal(err)
	}
	empty := stagedAllocation(t, s, 1, 0, 0)
	if err := s.FinishNativeAllocations(types); err != nil {
		t.Fatal(err)
	}
	if s.IsLiveReference(first) || !s.IsLiveReference(empty) || s.Usage().HeapObjects != 1 || s.Usage().HeapBytes != 64 {
		t.Fatal("generation/empty object", first, empty, s.Usage())
	}
	// A skipped constructor leaves all reserved handles available, including
	// metadata pages that have not yet materialized their first object.
	for range 3 {
		if _, err := s.PrepareNativeAllocations(); err != nil {
			t.Fatal(err)
		}
		if err := s.FinishNativeAllocations(types); err != nil {
			t.Fatal(err)
		}
	}
	s.CollectEveryAllocation = true
	if address, err := s.PrepareNativeAllocations(); err != nil || address != 0 {
		t.Fatal("forced collection bypassed", address, err)
	}
	s.Close()
	if _, err := s.PrepareNativeAllocations(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal("closed heap", err)
	}
}

func TestNativeAllocationInvalidBatchesDiscardReservations(t *testing.T) {
	for _, kind := range []string{"count", "cursor", "type", "address", "length", "shape"} {
		t.Run(kind, func(t *testing.T) {
			s, types := allocationStore(t)
			if _, err := s.PrepareNativeAllocations(); err != nil {
				t.Fatal(err)
			}
			ref := stagedAllocation(t, s, 0, 42, 0)
			h := s.nativeHeap
			at := int(uint32(ref)) * 4
			switch kind {
			case "count":
				h.allocationWords[0] = nativeabi.HeapAllocationBatch + 1
			case "cursor":
				h.allocationWords[nativeabi.HeapAllocationUsed/8] = ^uint64(0)
			case "type":
				h.views[at+3] = uint64(len(types))
			case "address":
				h.views[at+1]++
			case "length":
				h.views[at+2]++
			case "shape":
				h.views[at+3] = 1
			}
			if err := s.FinishNativeAllocations(types); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal("invalid batch accepted", kind, err)
			}
			if s.Usage().HeapObjects != 0 || s.nativeHeap.pendingCount != 0 || s.IsLiveReference(ref) {
				t.Fatal("invalid batch leaked objects")
			}
			if _, err := s.PrepareNativeAllocations(); err != nil {
				t.Fatal(err)
			}
			stagedAllocation(t, s, 0, 17, 0)
			if err := s.FinishNativeAllocations(types); err != nil {
				t.Fatal("invalid batch poisoned next invocation", err)
			}
		})
	}
}
