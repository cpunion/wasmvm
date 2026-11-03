package state

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"math"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const (
	heapTag             uint64 = 1 << 63
	i31Tag              uint64 = 1 << 62
	externalTag         uint64 = 1 << 61
	MaxExternalIdentity uint64 = externalTag - 1
)

type ReferenceValue struct {
	Type  wasm.ValueType
	Value value.Value
}
type heapObject struct {
	Type          uint32
	Descriptor    uint64
	Continuation  ContinuationBody
	Tag           *Tag
	Data          []byte
	Layout        *heapLayout
	Length        uint32
	Bytes, Serial uint64
	Mark          bool
}
type heapSlot struct {
	Object     *heapObject
	Generation uint32
	// Free slots link through the existing alignment space. Zero terminates
	// the chain; live and retired slots never participate in it.
	Next uint32
}

// Handle-indexed pages keep object addresses stable without allocating a host
// object for every guest allocation. Pages are lazy and bounded by the handle
// quota; a recycled generation reuses its cleared metadata cell.
const heapObjectsPerPage = 64

func (s *Store) objectCell(index uint32) *heapObject {
	page := uint64(index) / heapObjectsPerPage
	for page >= uint64(len(s.heapObjects)) {
		remaining := uint64(s.limits.MaxHeapObjects) + 1 - uint64(len(s.heapObjects))*heapObjectsPerPage
		s.heapObjects = appendValue(s.heapObjects, make([]heapObject, int(min(heapObjectsPerPage, remaining))))
	}
	return &s.heapObjects[page][uint64(index)%heapObjectsPerPage]
}

func (s *Store) heapObject(ref uint64) (*heapObject, bool) {
	index := uint32(ref)
	if ref&heapTag == 0 || index == 0 || uint64(index) >= uint64(len(s.heap)) {
		return nil, false
	}
	slot := &s.heap[index]
	return slot.Object, slot.Object != nil && slot.Generation == uint32(ref>>32)&0x7fffffff
}
func (s *Store) allocateObject(typ uint32, length uint32, extra ...uint64) (uint64, error) {
	layout, err := s.layout(typ)
	if err != nil {
		return 0, err
	}
	payload := layout.size
	if layout.stride != 0 {
		payload = uint64(length) * uint64(layout.stride)
	}
	bytes := payload + 64
	if len(extra) != 0 {
		if extra[0] > math.MaxUint64-bytes {
			return 0, limit("guest continuation bytes")
		}
		bytes += extra[0]
	}
	if s.registry.Types[typ].Definition.Descriptor != 0 {
		bytes += 8
	}
	if s.HeapSafepoint != nil && (s.CollectEveryAllocation || s.usage.HeapObjects >= uint64(s.limits.MaxHeapObjects) || bytes > s.limits.MaxHeapBytes-s.usage.HeapBytes || s.usage.HeapBytes >= s.heapThreshold) {
		s.HeapSafepoint()
	}
	if s.usage.HeapObjects >= uint64(s.limits.MaxHeapObjects) || bytes > s.limits.MaxHeapBytes-s.usage.HeapBytes || payload > uint64(math.MaxInt) || s.heapSerial == math.MaxUint64 {
		return 0, limit("guest heap")
	}
	// Select a handle before reserving payload storage so limit failures cannot
	// consume arena space. Commit the free list only after allocations succeed.
	index := uint32(max(1, len(s.heap)))
	if s.heapFree != 0 {
		index = s.heapFree
	} else if uint64(len(s.heap)) > math.MaxUint32 || uint64(index) > uint64(s.limits.MaxHeapObjects) {
		return 0, limit("guest heap handles")
	}
	var data []byte
	var offset uint64
	if s.nativeHeap != nil {
		if err := s.prepareHeapViews(uint64(index) + 1); err != nil {
			return 0, err
		}
		data, offset, err = s.allocateNativePayload(payload)
		if err != nil {
			return 0, err
		}
	} else {
		data = make([]byte, int(payload))
	}
	if len(s.heap) == 0 {
		s.heap = appendValue(s.heap, heapSlot{})
	}
	if s.heapFree != 0 {
		s.heapFree = s.heap[index].Next
		s.heap[index].Next = 0
	} else {
		s.heap = appendValue(s.heap, heapSlot{Generation: 1})
	}
	s.heapSerial++
	ref := heapTag | uint64(s.heap[index].Generation)<<32 | uint64(index)
	o := s.objectCell(index)
	*o = heapObject{Type: typ, Data: data, Layout: layout, Length: length, Bytes: bytes, Serial: s.heapSerial}
	s.heap[index].Object = o
	if s.nativeHeap != nil {
		if payload != 0 {
			s.nativeHeap.objects = appendValue(s.nativeHeap.objects, nativeHeapAllocation{offset, ref})
		} else {
			s.trackNativeEmpty(index)
		}
		s.publishHeapObject(ref, o)
	}
	s.usage.HeapObjects++
	s.usage.HeapBytes += bytes
	if s.heapThreshold == 0 {
		s.heapThreshold = 1 << 20
	}
	return ref, nil
}
func (s *Store) releaseObject(index uint32) {
	slot := &s.heap[index]
	obj := slot.Object
	if obj == nil {
		return
	}
	if s.nativeHeap != nil && len(obj.Data) == 0 {
		s.nativeHeap.empty[index/64] &^= uint64(1) << (index % 64)
	}
	s.releaseHeapObject(index, obj)
}

// The collector already has a live cell. Empty-object sweeping updates its
// bitmap once per word; individual release/rollback removes the bit above.
func (s *Store) releaseHeapObject(index uint32, obj *heapObject) {
	slot := &s.heap[index]
	s.usage.HeapObjects--
	s.usage.HeapBytes -= obj.Bytes
	if s.nativeHeap != nil {
		clear(s.nativeHeap.views[int(index)*4 : int(index)*4+4])
	}
	*obj = heapObject{}
	slot.Object = nil
	if slot.Generation < 0x7fffffff {
		slot.Generation++
		slot.Next = s.heapFree
		s.heapFree = index
	}
}
func (s *Store) rollbackHeap(serial uint64) {
	for i := 1; i < len(s.heap); i++ {
		if o := s.heap[i].Object; o != nil && o.Serial > serial {
			s.releaseObject(uint32(i))
		}
	}
}
func I31(bits uint64) uint64 { return i31Tag | bits&0x7fffffff }
func Internalize(ref uint64) uint64 {
	if ref == 0 || ref&(heapTag|i31Tag) != 0 {
		return ref
	}
	return ref | externalTag
}
func Externalize(ref uint64) uint64 {
	if ref&externalTag != 0 && ref&(heapTag|i31Tag) == 0 {
		return ref &^ externalTag
	}
	return ref
}
func IsExternalizedGuest(ref uint64) bool { return ref&(heapTag|i31Tag) != 0 }

// TestReference accepts canonical expected types and validates generations
// before exposing any object. The caller already selected the static domain.
func (s *Store) TestReference(ref uint64, expected wasm.ValueType) bool {
	if ref == 0 {
		return expected.Nullable()
	}
	var actual wasm.ValueType
	switch {
	case expected.IsFunctionReference():
		f, ok := s.Function(ref)
		if !ok {
			return false
		}
		actual = wasm.CanonicalFunctionReference(f.TypeID, false)
	case expected.IsExternalReference():
		return byte(expected) != byte(wasm.NoExternRef)
	case ref&heapTag != 0:
		o, ok := s.heapObject(ref)
		if !ok {
			return false
		}
		if o.Tag != nil {
			return s.registry.Subtype(wasm.ExnRef.NonNull(), expected)
		}
		d := s.registry.Types[o.Type].Definition
		family := wasm.StructRef
		if d.Kind == 0x5e {
			family = wasm.ArrayRef
		} else if d.Kind == module.ContinuationKind {
			family = wasm.ContRef
		}
		actual = wasm.IndexedReference(family, o.Type, false).WithTypeIndex(o.Type, true)
	case expected.IsExceptionReference():
		return false
	case ref&i31Tag != 0:
		actual = wasm.I31Ref.NonNull()
	case ref&externalTag != 0:
		actual = wasm.AnyRef.NonNull()
	default:
		return false
	}
	if _, indexed := actual.TypeIndex(); indexed {
		actual = actual.WithExact(true)
	}
	return s.registry.Subtype(actual, expected)
}

// DescriptorEqual implements the common descriptor-cast predicate. A null
// descriptor traps even if the tested reference is null or has another kind.
func (s *Store) DescriptorEqual(ref, descriptor uint64, target wasm.ValueType) (bool, error) {
	if descriptor == 0 {
		return false, &wasm.Trap{Code: wasm.TrapNullReference}
	}
	if ref == 0 {
		return target.Nullable(), nil
	}
	obj, ok := s.heapObject(ref)
	return ok && obj.Descriptor == descriptor, nil
}
func (s *Store) IsLiveReference(ref uint64) bool {
	if ref&heapTag != 0 {
		_, ok := s.heapObject(ref)
		return ok
	}
	return true
}
