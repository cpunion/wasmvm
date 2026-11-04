package state

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"math"
	"unsafe"

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
	Ref          uint64
	Data         *byte
	Length, Type uint32
	Serial       uint64
}

// Exceptional payloads and extension-specific roots are allocated only when
// present. Ordinary structs and arrays retain a compact shared object record.
type heapObjectExtra struct {
	Descriptor   uint64
	Tag          *Tag
	Continuation ContinuationBody
	Bytes        uint64
}

// Optional cells share a lazily allocated page; exceptional objects do not
// allocate a separate host object for each tag, descriptor or continuation.
func (s *Store) objectExtras(ref uint64) *heapObjectExtra {
	index := uint32(ref)
	slot := &s.heap[index]
	if slot.Next&heapHasExtra == 0 {
		page := uint64(index) / heapObjectsPerPage
		if s.heapObjects[page].extras == nil {
			s.materializeExtraSlab(page)
		}
		slot.Next |= heapHasExtra
	}
	return s.existingObjectExtras(slot.Object)
}

func (s *Store) materializeExtraSlab(page uint64) {
	p := &s.heapObjects[page]
	first := page - uint64(p.blockOffset)
	pages := uint64(p.blockPages)
	count := (pages-1)*heapObjectsPerPage + uint64(len(s.heapObjects[first+pages-1].objects))
	cells := make([]heapObjectExtra, int(count))
	at := 0
	for i := first; i < first+pages; i++ {
		end := at + len(s.heapObjects[i].objects)
		s.heapObjects[i].extras = cells[at:end:end]
		at = end
	}
}

// Only published objects use live-slot flags. A cleared header cannot confuse
// a free-list index with a mark or optional-cell flag.
func (s *Store) existingObjectExtras(o *heapObject) *heapObjectExtra {
	index := uint64(uint32(o.Ref))
	if index == 0 || s.heap[index].Next&heapHasExtra == 0 {
		return nil
	}
	return &s.heapObjects[index/heapObjectsPerPage].extras[index%heapObjectsPerPage]
}

func (s *Store) objectDescriptor(o *heapObject) uint64 {
	if extra := s.existingObjectExtras(o); extra != nil {
		return extra.Descriptor
	}
	return 0
}

type heapSlot struct {
	Object     *heapObject
	Generation uint32
	// Free slots link through the existing alignment space; zero terminates
	// the chain. Live slots use mark/optional flags; retired slots use zero.
	Next uint32
}

const (
	heapMarked uint32 = 1 << iota
	heapHasExtra
)

func (s *Store) markObject(o *heapObject) bool {
	slot := &s.heap[uint32(o.Ref)]
	if slot.Next&heapMarked != 0 {
		return false
	}
	slot.Next |= heapMarked
	return true
}

func (s *Store) objectMarked(o *heapObject) bool {
	return o.Ref != 0 && s.heap[uint32(o.Ref)].Next&heapMarked != 0
}

func (s *Store) clearObjectMark(o *heapObject) {
	s.heap[uint32(o.Ref)].Next &^= heapMarked
}

// Handle-indexed pages keep object addresses stable without allocating a host
// object for every guest allocation. Pages are lazy and bounded by the handle
// quota; a recycled generation reuses its cleared metadata cell.
const heapObjectsPerPage = 64

type heapObjectPage struct {
	objects                 []heapObject
	extras                  []heapObjectExtra
	hasEmpty                bool
	blockOffset, blockPages uint8
}

// Small Stores start with one logical page. Larger portable heaps share typed
// allocations capped at eight pages; native pages alias stable owned headers.
// The final slab and page stop at the handle quota, including the null slot.
func (s *Store) materializeObjectSlabs(page uint64) error {
	for page >= uint64(len(s.heapObjects)) {
		first := uint64(len(s.heapObjects))
		remaining := uint64(s.limits.MaxHeapObjects) + 1 - first*heapObjectsPerPage
		count := min(remaining, min(uint64(8), first+1)*heapObjectsPerPage)
		if count == 0 {
			panic("state: metadata handle quota")
		}
		var cells []heapObject
		if s.nativeHeap == nil {
			cells = make([]heapObject, int(count))
		} else {
			end := first*heapObjectsPerPage + count
			if err := s.prepareHeapViews(end); err != nil {
				return err
			}
			headers := unsafe.Slice((*heapObject)(unsafe.Pointer(unsafe.SliceData(s.nativeHeap.views))), len(s.nativeHeap.views)/4)
			cells = headers[first*heapObjectsPerPage : end : end]
		}
		pages := (count + heapObjectsPerPage - 1) / heapObjectsPerPage
		for i, at := uint64(0), 0; i < pages; i++ {
			end := min(len(cells), at+heapObjectsPerPage)
			s.heapObjects = appendValue(s.heapObjects, heapObjectPage{
				objects: cells[at:end:end], blockOffset: uint8(i), blockPages: uint8(pages),
			})
			at = end
		}
	}
	return nil
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
	bytes := payload + layout.overhead
	if len(extra) != 0 {
		if extra[0] > math.MaxUint64-bytes {
			return 0, limit("guest continuation bytes")
		}
		bytes += extra[0]
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
	page := uint64(index) / heapObjectsPerPage
	if page >= uint64(len(s.heapObjects)) {
		if err := s.materializeObjectSlabs(page); err != nil {
			return 0, err
		}
	}
	o := &s.heapObjects[page].objects[uint64(index)%heapObjectsPerPage]
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
	*o = heapObject{Ref: ref, Type: typ, Data: unsafe.SliceData(data), Length: length, Serial: s.heapSerial}
	s.heap[index].Object = o
	if len(extra) != 0 {
		s.objectExtras(ref).Bytes = extra[0]
	}
	if s.nativeHeap != nil {
		if payload != 0 {
			s.nativeHeap.objects = appendValue(s.nativeHeap.objects, nativeHeapAllocation{offset, ref})
		} else {
			s.trackNativeEmpty(index)
		}
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
	s.releaseHeapObject(index, obj)
}

// The collector already has a live cell. Clearing its identity also removes it
// from the empty-object inventory, including individual release and rollback.
func (s *Store) releaseHeapObject(index uint32, obj *heapObject) {
	slot := &s.heap[index]
	s.usage.HeapObjects--
	layout := s.heapLayouts[obj.Type]
	bytes := layout.payloadSize(obj.Length) + layout.overhead
	if slot.Next&heapHasExtra != 0 {
		extra := &s.heapObjects[uint64(index)/heapObjectsPerPage].extras[uint64(index)%heapObjectsPerPage]
		bytes += extra.Bytes
		*extra = heapObjectExtra{}
	}
	s.usage.HeapBytes -= bytes
	*obj = heapObject{}
	slot.Object = nil
	slot.Next = 0
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
		if extra := s.existingObjectExtras(o); extra != nil && extra.Tag != nil {
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
	return ok && s.objectDescriptor(obj) == descriptor, nil
}
func (s *Store) IsLiveReference(ref uint64) bool {
	if ref&heapTag != 0 {
		_, ok := s.heapObject(ref)
		return ok
	}
	return true
}
