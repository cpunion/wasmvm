package state

import (
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
	Data          []byte
	Layout        *heapLayout
	Length        uint32
	Bytes, Serial uint64
	Mark          bool
}
type heapSlot struct {
	Object     *heapObject
	Generation uint32
}

func (s *Store) heapObject(ref uint64) (*heapObject, bool) {
	index := uint32(ref)
	if ref&heapTag == 0 || index == 0 || uint64(index) >= uint64(len(s.heap)) {
		return nil, false
	}
	slot := &s.heap[index]
	return slot.Object, slot.Object != nil && slot.Generation == uint32(ref>>32)&0x7fffffff
}
func (s *Store) allocateObject(typ uint32, length uint32) (uint64, error) {
	layout, err := s.layout(typ)
	if err != nil {
		return 0, err
	}
	payload := layout.size
	if layout.stride != 0 {
		payload = uint64(length) * uint64(layout.stride)
	}
	bytes := payload + 64
	if s.HeapSafepoint != nil && (s.CollectEveryAllocation || s.usage.HeapObjects >= uint64(s.limits.MaxHeapObjects) || bytes > s.limits.MaxHeapBytes-s.usage.HeapBytes || s.usage.HeapBytes >= s.heapThreshold) {
		s.HeapSafepoint()
	}
	if s.usage.HeapObjects >= uint64(s.limits.MaxHeapObjects) || bytes > s.limits.MaxHeapBytes-s.usage.HeapBytes || payload > uint64(math.MaxInt) || s.heapSerial == math.MaxUint64 {
		return 0, limit("guest heap")
	}
	if len(s.heap) == 0 {
		s.heap = append(s.heap, heapSlot{})
	}
	var index uint32
	if len(s.heapFree) != 0 {
		n := len(s.heapFree) - 1
		index = s.heapFree[n]
		s.heapFree = s.heapFree[:n]
	} else {
		if uint64(len(s.heap)) > math.MaxUint32 || uint64(len(s.heap)-1) >= uint64(s.limits.MaxHeapObjects) {
			return 0, limit("guest heap handles")
		}
		index = uint32(len(s.heap))
		s.heap = append(s.heap, heapSlot{Generation: 1})
	}
	s.heapSerial++
	s.heap[index].Object = &heapObject{Type: typ, Data: make([]byte, int(payload)), Layout: layout, Length: length, Bytes: bytes, Serial: s.heapSerial}
	s.usage.HeapObjects++
	s.usage.HeapBytes += bytes
	if s.heapThreshold == 0 {
		s.heapThreshold = 1 << 20
	}
	return heapTag | uint64(s.heap[index].Generation)<<32 | uint64(index), nil
}
func (s *Store) releaseObject(index uint32) {
	slot := &s.heap[index]
	obj := slot.Object
	if obj == nil {
		return
	}
	s.usage.HeapObjects--
	s.usage.HeapBytes -= obj.Bytes
	obj.Data = nil
	obj.Layout = nil
	slot.Object = nil
	if slot.Generation < 0x7fffffff {
		slot.Generation++
		s.heapFree = append(s.heapFree, index)
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
		d := s.registry.Types[o.Type].Definition
		family := wasm.StructRef
		if d.Kind == 0x5e {
			family = wasm.ArrayRef
		}
		actual = wasm.IndexedReference(family, o.Type, false).WithTypeIndex(o.Type, true)
	case ref&i31Tag != 0:
		actual = wasm.I31Ref.NonNull()
	case ref&externalTag != 0:
		actual = wasm.AnyRef.NonNull()
	default:
		return false
	}
	return s.registry.Subtype(actual, expected)
}
func (s *Store) IsLiveReference(ref uint64) bool {
	if ref&heapTag != 0 {
		_, ok := s.heapObject(ref)
		return ok
	}
	return true
}
