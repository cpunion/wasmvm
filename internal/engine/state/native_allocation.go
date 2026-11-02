package state

import (
	"math"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// NativeAllocationStats reports committed batches/objects for internal probes.
func (s *Store) NativeAllocationStats() (uint64, uint64) {
	if s.nativeHeap == nil {
		return 0, 0
	}
	return s.nativeHeap.allocationBatches, s.nativeHeap.allocatedObjects
}

// PrepareNativeAllocations reserves handles and stable payload capacity. It
// neither traces roots nor allocates guest objects. Serialized native execution
// is the only borrower; FinishNativeAllocations must run on every exit/error.
func (s *Store) PrepareNativeAllocations() (uint64, error) {
	if s.closed {
		return 0, wasm.ErrClosed
	}
	h := s.nativeHeap
	if h == nil || h.pendingCount != 0 {
		return 0, wasm.ErrInvalid
	}
	if s.CollectEveryAllocation || s.usage.HeapObjects >= uint64(s.limits.MaxHeapObjects) || s.usage.HeapBytes >= s.limits.MaxHeapBytes || s.usage.HeapBytes >= s.heapThreshold || s.heapSerial == math.MaxUint64 {
		return 0, nil
	}
	if h.allocation == nil {
		size := uint64(nativeabi.HeapAllocationHeaderBytes + nativeabi.HeapAllocationBatch*8)
		var err error
		h.allocation, err = native.NewMemory(size, size)
		if err != nil {
			return 0, err
		}
		buffer, err := h.allocation.Buffer()
		if err != nil {
			return 0, err
		}
		h.allocationWords = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), int(size/8))
	}
	if err := s.reserveNativePayload(1); err != nil {
		return 0, err
	}
	words := h.allocationWords
	clear(words)
	refs := words[nativeabi.HeapAllocationHeaderBytes/8:]
	count := min(uint64(nativeabi.HeapAllocationBatch), uint64(s.limits.MaxHeapObjects)-s.usage.HeapObjects, math.MaxUint64-s.heapSerial)
	if len(s.heap) == 0 {
		s.heap = appendValue(s.heap, heapSlot{})
	}
	for h.pendingCount < count {
		index := uint32(len(s.heap))
		if len(s.heapFree) != 0 {
			n := len(s.heapFree) - 1
			index, s.heapFree = s.heapFree[n], s.heapFree[:n]
		} else {
			if uint64(len(s.heap)) > uint64(s.limits.MaxHeapObjects) {
				break
			}
			s.heap = appendValue(s.heap, heapSlot{Generation: 1})
		}
		refs[h.pendingCount] = heapTag | uint64(s.heap[index].Generation)<<32 | uint64(index)
		h.pendingCount++
	}
	if err := s.prepareHeapViews(uint64(len(s.heap))); err != nil {
		s.releaseNativeReservations(0)
		return 0, err
	}
	address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(words))))
	words[nativeabi.HeapAllocationCapacity/8] = h.pendingCount
	words[nativeabi.HeapAllocationHandles/8] = address + nativeabi.HeapAllocationHeaderBytes
	words[nativeabi.HeapAllocationBase/8] = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(h.data))))
	words[nativeabi.HeapAllocationUsed/8] = h.used
	words[nativeabi.HeapAllocationLimit/8] = uint64(len(h.data))
	words[nativeabi.HeapAllocationBudget/8] = min(s.limits.MaxHeapBytes-s.usage.HeapBytes, s.heapThreshold-s.usage.HeapBytes)
	words[nativeabi.HeapAllocationDescriptors/8] = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(h.views))))
	return address, nil
}

// FinishNativeAllocations publishes initialized native objects into the same
// typed heap used by the interpreter and collector. Local type IDs belong to
// the current entry's instance. Invalid batches discard their reservations.
func (s *Store) FinishNativeAllocations(types []uint32) error {
	h := s.nativeHeap
	if h == nil || h.pendingCount == 0 {
		return nil
	}
	return s.finishNativeAllocations(types)
}

// Keep deferred rollback inside the active-batch path. In particular, an idle
// finish must not construct LLGo's heap-owned exception/defer records.
func (s *Store) finishNativeAllocations(types []uint32) error {
	h := s.nativeHeap
	words := h.allocationWords
	count, end := words[nativeabi.HeapAllocationCursor/8], words[nativeabi.HeapAllocationUsed/8]
	committed := uint64(0)
	defer func() { s.releaseNativeReservations(committed) }()
	if count > h.pendingCount || end < h.used || end > uint64(len(h.data)) {
		return wasm.ErrInvalid
	}
	refs := words[nativeabi.HeapAllocationHeaderBytes/8:]
	var layouts [nativeabi.HeapAllocationBatch]*heapLayout
	var identities [nativeabi.HeapAllocationBatch]uint32
	offset, bytes := h.used, uint64(0)
	previous := ^uint64(0)
	var typ uint32
	var layout *heapLayout
	for i, ref := range refs[:count] {
		at := uint64(uint32(ref)) * 4
		local := h.views[at+3]
		if h.views[at] != ref || local >= uint64(len(types)) || s.heap[uint32(ref)].Object != nil {
			return wasm.ErrInvalid
		}
		if local != previous {
			typ = types[local]
			var err error
			layout, err = s.layout(typ)
			if err != nil {
				return err
			}
			definition := s.registry.Types[typ].Definition
			if definition.Kind != module.StructKind || definition.Descriptor != 0 || definition.Describes != 0 || len(definition.Fields) > nativeabi.MaxNativeStructFields {
				return wasm.ErrInvalid
			}
			previous = local
		}
		if h.views[at+2] != uint64(len(layout.widths)) || layout.size > end-offset {
			return wasm.ErrInvalid
		}
		address := uint64(0)
		if layout.size != 0 {
			address = uint64(uintptr(unsafe.Pointer(&h.data[offset])))
		}
		if h.views[at+1] != address {
			return wasm.ErrInvalid
		}
		offset += layout.size
		bytes += layout.size + 64
		layouts[i], identities[i] = layout, typ
	}
	if offset != end || count > uint64(s.limits.MaxHeapObjects)-s.usage.HeapObjects || bytes > s.limits.MaxHeapBytes-s.usage.HeapBytes || count > math.MaxUint64-s.heapSerial {
		return wasm.ErrInvalid
	}
	offset = h.used
	for i, ref := range refs[:count] {
		layout := layouts[i]
		s.heapSerial++
		o := s.objectCell(uint32(ref))
		*o = heapObject{Type: identities[i], Layout: layout, Length: uint32(len(layout.widths)), Bytes: layout.size + 64, Serial: s.heapSerial}
		if layout.size != 0 {
			o.Data = h.data[offset : offset+layout.size]
			h.objects = appendValue(h.objects, nativeHeapAllocation{o, offset, ref})
		}
		s.heap[uint32(ref)].Object = o
		s.publishHeapObject(ref, o)
		offset += layout.size
	}
	h.used = end
	s.usage.HeapObjects += count
	s.usage.HeapBytes += bytes
	committed = count
	if count != 0 {
		h.allocationBatches++
		h.allocatedObjects += count
	}
	return nil
}

func (s *Store) releaseNativeReservations(committed uint64) {
	h := s.nativeHeap
	refs := h.allocationWords[nativeabi.HeapAllocationHeaderBytes/8:]
	for _, ref := range refs[committed:h.pendingCount] {
		index := uint32(ref)
		if int(index)*4+4 <= len(h.views) {
			clear(h.views[int(index)*4 : int(index)*4+4])
		}
		s.heapFree = appendValue(s.heapFree, index)
	}
	h.pendingCount = 0
	clear(h.allocationWords)
}
