package state

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

// Native entries borrow only owned native allocations. Collection and growth
// run at serialized safepoints, compact payloads in allocation order, and
// republish every moved address before reentry. Guest references remain stable
// generation-checked handles; neither payloads nor views contain Go pointers.
type nativeHeap struct {
	payload, slots *native.Memory
	data           []byte
	views          []uint64
	used           uint64
	objects        []*heapObject
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
	if size > uint64(len(h.data))-h.used {
		s.compactNativeHeap()
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
			return nil, 0, err
		}
		h.data, err = h.payload.Buffer()
		if err != nil {
			return nil, 0, err
		}
		for _, o := range h.objects {
			o.Data = h.data[o.nativeOffset : o.nativeOffset+uint64(len(o.Data))]
			s.publishHeapObject(o)
		}
	}
	offset := h.used
	h.used += size
	data := h.data[offset:h.used]
	clear(data)
	return data, offset, nil
}

func (s *Store) compactNativeHeap() {
	h := s.nativeHeap
	kept := h.objects[:0]
	var used uint64
	for _, o := range h.objects {
		if o.Data == nil {
			continue
		}
		size := uint64(len(o.Data))
		copy(h.data[used:used+size], o.Data)
		o.Data, o.nativeOffset = h.data[used:used+size], used
		used += size
		kept = append(kept, o)
		s.publishHeapObject(o)
	}
	clear(h.objects[len(kept):])
	h.objects, h.used = kept, used
}

func (s *Store) publishHeapObject(o *heapObject) {
	at := int(uint32(o.nativeHandle)) * 4
	var address uint64
	if len(o.Data) != 0 {
		address = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(o.Data))))
	}
	views := s.nativeHeap.views
	views[at], views[at+1], views[at+2] = o.nativeHandle, address, uint64(o.Length)
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
	if h.payload != nil {
		h.payload.Close()
	}
	if h.slots != nil {
		h.slots.Close()
	}
	*h = nativeHeap{}
}
