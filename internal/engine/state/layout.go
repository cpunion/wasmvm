package state

import (
	"encoding/binary"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
	"unsafe"
)

type heapLayout struct {
	offsets    []uint32
	widths     []byte
	references []heapReference
	size       uint64
	overhead   uint64
	stride     int
}

// Reference positions and kinds are immutable properties of the admitted type.
// Arrays use one element entry; structs and exceptions retain byte offsets.
type heapReference struct {
	typ    wasm.ValueType
	offset uint32
}

func (s *Store) layout(typ uint32) (*heapLayout, error) {
	if uint64(typ) < uint64(len(s.heapLayouts)) {
		if l := s.heapLayouts[typ]; l != nil {
			return l, nil
		}
	}
	if typ == 0 || uint64(typ) >= uint64(len(s.registry.Types)) {
		return nil, wasm.ErrInvalid
	}
	d := s.registry.Types[typ].Definition

	if d.Kind == module.FunctionKind {
		d.Fields = make([]module.FieldType, len(s.registry.Types[typ].Function.Params))
		for i, t := range s.registry.Types[typ].Function.Params {
			d.Fields[i] = module.FieldType{Value: t}
		}
	}
	l := &heapLayout{overhead: 64}
	if d.Descriptor != 0 {
		l.overhead += 8
	}
	if d.Kind == module.ArrayKind {
		l.stride = d.Fields[0].HeapBytes()
		if t := d.Fields[0].Value; t.IsReference() {
			l.references = []heapReference{{typ: t}}
		}
	} else {
		l.offsets = make([]uint32, len(d.Fields))
		l.widths = make([]byte, len(d.Fields))
		for i, f := range d.Fields {
			if l.size > math.MaxUint32 {
				return nil, limit("heap layout")
			}
			l.offsets[i] = uint32(l.size)
			if f.Value.IsReference() {
				l.references = append(l.references, heapReference{typ: f.Value, offset: uint32(l.size)})
			}
			n := f.HeapBytes()
			l.widths[i] = byte(n)
			l.size += uint64(n)
		}
	}
	if uint64(typ) >= uint64(len(s.heapLayouts)) {
		n := min(uint64(len(s.registry.Types)), max(uint64(typ)+1, uint64(len(s.heapLayouts))*2))
		layouts := make([]*heapLayout, int(n))
		copy(layouts, s.heapLayouts)
		s.heapLayouts = layouts
	}
	s.heapLayouts[typ] = l
	return l, nil
}
func (s *Store) objectSize(o *heapObject) uint64 {
	layout := s.heapLayouts[o.Type]
	if layout.stride != 0 {
		return uint64(o.Length) * uint64(layout.stride)
	}
	return layout.size
}

func (s *Store) objectData(o *heapObject) []byte {
	return unsafe.Slice(o.Data, int(s.objectSize(o)))
}

func (s *Store) objectBytes(o *heapObject) uint64 {
	bytes := s.objectSize(o) + s.heapLayouts[o.Type].overhead
	if extra := s.existingObjectExtras(o); extra != nil {
		bytes += extra.Bytes
	}
	return bytes
}

func (s *Store) objectLocation(o *heapObject, i uint32) (unsafe.Pointer, int) {
	layout := s.heapLayouts[o.Type]
	var offset, width int
	if layout.stride != 0 {
		if i >= o.Length {
			panic("state: array field index out of bounds")
		}
		offset, width = int(i)*layout.stride, layout.stride
	} else {
		offset, width = int(layout.offsets[i]), int(layout.widths[i])
	}
	return unsafe.Add(unsafe.Pointer(o.Data), offset), width
}

// Location derives bounds from the initialized header and immutable layout.
// Exact-width byte arrays retain unaligned little-endian access without a
// variable-length slice header for every scalar field.
func (s *Store) getObject(o *heapObject, i uint32) value.Value {
	p, width := s.objectLocation(o, i)
	switch width {
	case 1:
		return value.Value{Lo: uint64(*(*byte)(p))}
	case 2:
		return value.Value{Lo: uint64(binary.LittleEndian.Uint16((*[2]byte)(p)[:]))}
	case 4:
		return value.Value{Lo: uint64(binary.LittleEndian.Uint32((*[4]byte)(p)[:]))}
	case 8:
		return value.Value{Lo: binary.LittleEndian.Uint64((*[8]byte)(p)[:])}
	case 16:
		data := (*[16]byte)(p)
		return value.Value{Lo: binary.LittleEndian.Uint64(data[:8]), Hi: binary.LittleEndian.Uint64(data[8:])}
	}
	panic("state: invalid validated data field")
}

func (s *Store) setObject(o *heapObject, i uint32, v value.Value) {
	p, width := s.objectLocation(o, i)
	switch width {
	case 1:
		*(*byte)(p) = byte(v.Lo)
	case 2:
		binary.LittleEndian.PutUint16((*[2]byte)(p)[:], uint16(v.Lo))
	case 4:
		binary.LittleEndian.PutUint32((*[4]byte)(p)[:], uint32(v.Lo))
	case 8:
		binary.LittleEndian.PutUint64((*[8]byte)(p)[:], v.Lo)
	case 16:
		data := (*[16]byte)(p)
		binary.LittleEndian.PutUint64(data[:8], v.Lo)
		binary.LittleEndian.PutUint64(data[8:], v.Hi)
	}
}
