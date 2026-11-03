package state

import (
	"math/bits"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

// Native entries borrow only owned native allocations. Collection and growth
// run at serialized safepoints, compact payloads in allocation order, and
// republish every moved address before reentry. Guest references remain stable
// generation-checked handles; neither payloads nor views contain Go pointers.
type nativeHeap struct {
	payload, slots                      *native.Memory
	data                                []byte
	views                               []uint64
	used                                uint64
	objects                             []nativeHeapAllocation
	empty                               []uint64
	allocation                          *native.Memory
	allocationWords                     []uint64
	pendingCount                        uint64
	allocationBatches, allocatedObjects uint64
}

// Nonempty payloads retain allocation order for in-place relocation. Numeric
// handles avoid duplicate host pointers and validate reused generations. Empty
// objects use one bit per handle instead of appending a relocation record.
type nativeHeapAllocation struct {
	offset, handle uint64
}

func (s *Store) prepareHeapViews(count uint64) error {
	h := s.nativeHeap
	if count <= uint64(len(h.views))/4 {
		return nil
	}
	maximum := (uint64(s.limits.MaxHeapObjects) + 1) * nativeabi.HeapSlotBytes
	size := min(maximum, max(count*nativeabi.HeapSlotBytes, max(64*nativeabi.HeapSlotBytes, uint64(len(h.views))*16)))
	var err error
	if h.slots == nil {
		h.slots, err = native.NewMemory(size, maximum)
	} else {
		err = h.slots.Grow(size)
	}
	if err != nil {
		return err
	}
	buffer, err := h.slots.Buffer()
	if err != nil {
		return err
	}
	h.views = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
	return nil
}

func (s *Store) allocateNativePayload(size uint64) ([]byte, uint64, error) {
	h := s.nativeHeap
	if size == 0 {
		return nil, 0, nil
	}
	if err := s.reserveNativePayload(size); err != nil {
		return nil, 0, err
	}
	offset := h.used
	h.used += size
	data := h.data[offset:h.used]
	clear(data)
	return data, offset, nil
}

func (s *Store) reserveNativePayload(size uint64) error {
	h := s.nativeHeap
	// Rollbacks can repeatedly reuse handles without triggering tracing. Bound
	// tombstone retention independently of a previously large arena capacity;
	// the live/dead ratio keeps compaction amortized for nearly full heaps.
	if size > uint64(len(h.data))-h.used || uint64(len(h.objects)) >= 2*s.usage.HeapObjects+64 {
		s.compactNativeHeap(false)
	}
	if size > uint64(len(h.data))-h.used {
		capacity := min(s.limits.MaxHeapBytes, max(h.used+size, max(4096, uint64(len(h.data))*2)))
		var err error
		if h.payload == nil {
			h.payload, err = native.NewMemory(capacity, s.limits.MaxHeapBytes)
		} else {
			err = h.payload.Grow(capacity)
		}
		if err != nil {
			return err
		}
		h.data, err = h.payload.Buffer()
		if err != nil {
			return err
		}
		for _, a := range h.objects {
			o, _ := s.heapObject(a.handle)
			o.Data = h.data[a.offset : a.offset+uint64(len(o.Data))]
			s.publishHeapObject(a.handle, o)
		}
	}
	return nil
}

// Collection sweeps nonempty objects while compacting their payloads. Capacity
// maintenance removes only tombstones and leaves live mark bits alone.
func (s *Store) compactNativeHeap(sweep bool) {
	h := s.nativeHeap
	kept := 0
	var used uint64
	for _, a := range h.objects {
		o, ok := s.heapObject(a.handle)
		if !ok {
			continue
		}
		if sweep {
			if !o.Mark {
				s.releaseObject(uint32(a.handle))
				continue
			}
			o.Mark = false
		}
		size := uint64(len(o.Data))
		if size != 0 && a.offset != used {
			copy(h.data[used:used+size], o.Data)
			o.Data = h.data[used : used+size]
			s.publishHeapObject(a.handle, o)
		}
		a.offset = used
		used += size
		h.objects[kept] = a
		kept++
	}
	h.objects, h.used = h.objects[:kept], used
}

func (s *Store) trackNativeEmpty(index uint32) {
	h := s.nativeHeap
	for int(index/64) >= len(h.empty) {
		h.empty = appendValue(h.empty, uint64(0))
	}
	h.empty[index/64] |= uint64(1) << (index % 64)
}

// Bits name live cells directly, so rollback and release clear them before a
// handle can be reserved for another generation. No tombstones accumulate.
func (s *Store) sweepNativeEmpty() {
	for page, word := range s.nativeHeap.empty {
		for word != 0 {
			index := uint32(page*64 + bits.TrailingZeros64(word))
			o := s.heap[index].Object
			if o.Mark {
				o.Mark = false
			} else {
				s.releaseObject(index)
			}
			word &= word - 1
		}
	}
}

func (s *Store) publishHeapObject(ref uint64, o *heapObject) {
	at := int(uint32(ref)) * 4
	var address uint64
	if len(o.Data) != 0 {
		address = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(o.Data))))
	}
	views := s.nativeHeap.views
	views[at], views[at+1], views[at+2], views[at+3] = ref, address, uint64(o.Length), 0
}

// NativeHeapView borrows the handle view until the next serialized safepoint.
// The native driver refreshes it on every entry, including after host calls.
func (s *Store) NativeHeapView() (count, address uint64) {
	if s.closed || s.nativeHeap == nil || len(s.nativeHeap.views) == 0 {
		return 0, 0
	}
	return uint64(len(s.heap)), uint64(uintptr(unsafe.Pointer(unsafe.SliceData(s.nativeHeap.views))))
}

func (h *nativeHeap) close() {
	if h.allocation != nil {
		h.allocation.Close()
	}
	if h.payload != nil {
		h.payload.Close()
	}
	if h.slots != nil {
		h.slots.Close()
	}
	*h = nativeHeap{}
}
