package state

import (
	"encoding/binary"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
)

type heapLayout struct {
	offsets []uint32
	widths  []byte
	size    uint64
	stride  int
}

func (s *Store) layout(typ uint32) (*heapLayout, error) {
	if l := s.heapLayouts[typ]; l != nil {
		return l, nil
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
	l := &heapLayout{}
	width := func(f module.FieldType) int {
		if f.Value.IsReference() {
			return 8
		}
		return fieldBytes(f)
	}
	if d.Kind == module.ArrayKind {
		l.stride = width(d.Fields[0])
	} else {
		l.offsets = make([]uint32, len(d.Fields))
		l.widths = make([]byte, len(d.Fields))
		for i, f := range d.Fields {
			if l.size > math.MaxUint32 {
				return nil, limit("heap layout")
			}
			l.offsets[i] = uint32(l.size)
			n := width(f)
			l.widths[i] = byte(n)
			l.size += uint64(n)
		}
	}
	if s.heapLayouts == nil {
		s.heapLayouts = make(map[uint32]*heapLayout)
	}
	s.heapLayouts[typ] = l
	return l, nil
}
func (o *heapObject) location(i uint32) (int, int) {
	if o.Layout.stride != 0 {
		return int(i) * o.Layout.stride, o.Layout.stride
	}
	return int(o.Layout.offsets[i]), int(o.Layout.widths[i])
}
func (o *heapObject) get(i uint32) value.Value {
	offset, width := o.location(i)
	return readField(o.Data[offset:], width)
}
func (o *heapObject) set(i uint32, v value.Value) {
	offset, width := o.location(i)
	data := o.Data[offset:]
	switch width {
	case 1:
		data[0] = byte(v.Lo)
	case 2:
		binary.LittleEndian.PutUint16(data, uint16(v.Lo))
	case 4:
		binary.LittleEndian.PutUint32(data, uint32(v.Lo))
	case 8:
		binary.LittleEndian.PutUint64(data, v.Lo)
	case 16:
		binary.LittleEndian.PutUint64(data, v.Lo)
		binary.LittleEndian.PutUint64(data[8:], v.Hi)
	}
}
