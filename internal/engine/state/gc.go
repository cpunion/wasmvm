package state

import (
	"context"
	"encoding/binary"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// GC executes an individual checked heap operation for both backends. Guest
// objects contain typed handles; neither native code nor the collector stores
// Go pointers in integer-valued guest frames.
func (s *Store) GC(ctx context.Context, in *Instance, ins *module.Instr, args []value.Value, roots ...func(func(wasm.ValueType, value.Value))) (value.Value, error) {
	if s.HeapSafepoint == nil && (ins.Op&0xffff <= 1 || ins.Op&0xffff >= 6 && ins.Op&0xffff <= 10) {
		s.HeapSafepoint = func() {
			var xs []ReferenceValue
			if len(roots) != 0 {
				roots[0](func(t wasm.ValueType, v value.Value) { xs = append(xs, ReferenceValue{t, v}) })
			}
			active := append([]*Instance{in}, s.instances...)
			s.CollectReferences(xs, active, nil)
		}
		defer func() { s.HeapSafepoint = nil }()
	}
	op := ins.Op & 0xffff
	var result value.Value
	trap := func(code wasm.TrapCode) error { return &wasm.Trap{Code: code, Offset: ins.Offset} }
	typ := func(i uint32) uint32 { return in.Types[i] }
	object := func(ref uint64) (*heapObject, error) {
		if ref == 0 {
			return nil, trap(wasm.TrapNullReference)
		}
		o, ok := s.heapObject(ref)
		if !ok {
			return nil, ErrReference
		}
		return o, nil
	}
	field := func(i uint32) module.FieldType { return s.registry.Types[typ(i)].Definition.Fields[0] }
	pack := func(v value.Value, f module.FieldType) value.Value {
		if f.Packed != 0 {
			v.Lo &= (1 << f.Packed) - 1
			v.Hi = 0
		}
		return v
	}
	poll := func(i int) error {
		if i&255 == 0 && ctx != nil {
			return ctx.Err()
		}
		return nil
	}
	allocate := func(length uint32) (*heapObject, error) {
		// Operands stay rooted even when called from a constant-expression stack.
		mark := len(s.temporary)
		fs := s.registry.Types[typ(ins.A)].Definition.Fields
		for i, a := range args {
			ft := wasm.I32
			if op == 0 && i < len(fs) {
				ft = fs[i].Unpacked()
			} else if op == 6 && i == 0 || op == 8 {
				ft = fs[0].Unpacked()
			}
			if ft.IsReference() {
				s.temporary = append(s.temporary, ReferenceValue{ft, a})
			}
		}
		defer func() { clear(s.temporary[mark:]); s.temporary = s.temporary[:mark] }()
		ref, err := s.allocateObject(typ(ins.A), length)
		if err != nil {
			return nil, err
		}
		result.Lo = ref
		o, _ := s.heapObject(ref)
		return o, nil
	}
	switch op {
	case 0, 1:
		fs := s.registry.Types[typ(ins.A)].Definition.Fields
		o, err := allocate(uint32(len(fs)))
		if err != nil {
			return result, err
		}
		if op == 0 {
			for i, a := range args {
				o.set(uint32(i), pack(a, fs[i]))
			}
		}
	case 2, 3, 4, 5:
		o, err := object(args[0].Lo)
		if err != nil {
			return result, err
		}
		f := s.registry.Types[typ(ins.A)].Definition.Fields[ins.B]
		if op == 5 {
			o.set(ins.B, pack(args[1], f))
		} else {
			result = o.get(ins.B)
			if op == 3 {
				result.Lo = uint64(uint32(int32(uint32(result.Lo)<<(32-f.Packed)) >> (32 - f.Packed)))
			}
		}
	case 6, 7, 8:
		n := ins.B
		if op == 6 {
			n = uint32(args[1].Lo)
		} else if op == 7 {
			n = uint32(args[0].Lo)
		}
		o, err := allocate(n)
		if err != nil {
			return result, err
		}
		f := field(ins.A)
		if op != 7 {
			for i := range int(o.Length) {
				if err := poll(i); err != nil {
					return result, err
				}
				if op == 6 {
					o.set(uint32(i), pack(args[0], f))
				} else if op == 8 {
					o.set(uint32(i), pack(args[i], f))
				}
			}
		}
		if ctx != nil && ctx.Err() != nil {
			return result, ctx.Err()
		}
	case 9, 10:
		start, n := uint32(args[0].Lo), uint32(args[1].Lo)
		f := field(ins.A)
		if op == 9 {
			width := fieldBytes(f)
			data := in.Data[ins.B]
			if uint64(start) > uint64(len(data)) || uint64(n)*uint64(width) > uint64(len(data))-uint64(start) {
				return result, trap(wasm.TrapMemoryBounds)
			}
			o, err := allocate(n)
			if err != nil {
				return result, err
			}
			for i := range int(o.Length) {
				if err := poll(i); err != nil {
					return result, err
				}
				o.set(uint32(i), readField(data[int(start)+i*width:], width))
			}
		} else {
			data := in.Elements[ins.B]
			if !Range(len(data), uint64(start), uint64(n)) {
				return result, trap(wasm.TrapTableBounds)
			}
			o, err := allocate(n)
			if err != nil {
				return result, err
			}
			for i := range int(o.Length) {
				if err := poll(i); err != nil {
					return result, err
				}
				o.set(uint32(i), value.Value{Lo: data[int(start)+i]})
			}
		}
	case 11, 12, 13, 14:
		o, err := object(args[0].Lo)
		if err != nil {
			return result, err
		}
		i := uint32(args[1].Lo)
		if uint64(i) >= uint64(o.Length) {
			return result, trap(wasm.TrapArrayBounds)
		}
		f := field(ins.A)
		if op == 14 {
			o.set(i, pack(args[2], f))
		} else {
			result = o.get(i)
			if op == 12 {
				result.Lo = uint64(uint32(int32(uint32(result.Lo)<<(32-f.Packed)) >> (32 - f.Packed)))
			}
		}
	case 15:
		o, err := object(args[0].Lo)
		if err != nil {
			return result, err
		}
		result.Lo = uint64(o.Length)
	case 16:
		o, err := object(args[0].Lo)
		if err != nil {
			return result, err
		}
		start, n := uint32(args[1].Lo), uint32(args[3].Lo)
		if !Range(int(o.Length), uint64(start), uint64(n)) {
			return result, trap(wasm.TrapArrayBounds)
		}
		v := pack(args[2], field(ins.A))
		for i := uint32(0); i < n; i++ {
			if err := poll(int(i)); err != nil {
				return result, err
			}
			o.set(start+i, v)
		}
	case 17:
		dst, err := object(args[0].Lo)
		if err != nil {
			return result, err
		}
		src, err := object(args[2].Lo)
		if err != nil {
			return result, err
		}
		di, si, n := uint32(args[1].Lo), uint32(args[3].Lo), uint32(args[4].Lo)
		if !Range(int(dst.Length), uint64(di), uint64(n)) || !Range(int(src.Length), uint64(si), uint64(n)) {
			return result, trap(wasm.TrapArrayBounds)
		}
		// Poll without losing memmove overlap semantics across chunk boundaries.
		backwards := dst == src && di > si && uint64(di) < uint64(si)+uint64(n)
		for done := uint32(0); done < n; {
			if err := poll(0); err != nil {
				return result, err
			}
			k := min(n-done, 256)
			offset := done
			if backwards {
				offset = n - done - k
			}
			width := dst.Layout.stride
			copy(dst.Data[(int(di)+int(offset))*width:(int(di)+int(offset+k))*width], src.Data[(int(si)+int(offset))*width:(int(si)+int(offset+k))*width])
			done += k
		}
	case 18, 19:
		o, err := object(args[0].Lo)
		if err != nil {
			return result, err
		}
		di, si, n := uint32(args[1].Lo), uint32(args[2].Lo), uint32(args[3].Lo)
		if !Range(int(o.Length), uint64(di), uint64(n)) {
			return result, trap(wasm.TrapArrayBounds)
		}
		if op == 18 {
			width := fieldBytes(field(ins.A))
			data := in.Data[ins.B]
			if uint64(si) > uint64(len(data)) || uint64(n)*uint64(width) > uint64(len(data))-uint64(si) {
				return result, trap(wasm.TrapMemoryBounds)
			}
			for i := uint32(0); i < n; i++ {
				if err := poll(int(i)); err != nil {
					return result, err
				}
				o.set(di+i, readField(data[int(si)+int(i)*width:], width))
			}
		} else {
			data := in.Elements[ins.B]
			if !Range(len(data), uint64(si), uint64(n)) {
				return result, trap(wasm.TrapTableBounds)
			}
			for i := uint32(0); i < n; i++ {
				if err := poll(int(i)); err != nil {
					return result, err
				}
				o.set(di+i, value.Value{Lo: data[int(si)+int(i)]})
			}
		}
	case 20, 21, 22, 23:
		expected := module.ResolveType(ins.ReferenceType(), in.Types)
		ok := s.TestReference(args[0].Lo, expected)
		if op <= 21 {
			if ok {
				result.Lo = 1
			}
		} else {
			if !ok {
				return result, trap(wasm.TrapCastFailure)
			}
			result = args[0]
		}
	case 26:
		result.Lo = Internalize(args[0].Lo)
	case 27:
		result.Lo = Externalize(args[0].Lo)
	case 28:
		result.Lo = I31(args[0].Lo)
	case 29, 30:
		if args[0].Lo == 0 {
			return result, trap(wasm.TrapNullReference)
		}
		result.Lo = args[0].Lo & 0x7fffffff
		if op == 29 {
			result.Lo = uint64(uint32(int32(uint32(result.Lo)<<1) >> 1))
		}
	default:
		return result, wasm.ErrInvalid
	}
	return result, nil
}
func fieldBytes(f module.FieldType) int {
	if f.Packed != 0 {
		return int(f.Packed) / 8
	}
	switch f.Value {
	case wasm.I32, wasm.F32:
		return 4
	case wasm.I64, wasm.F64:
		return 8
	case wasm.V128:
		return 16
	}
	return 0
}
func readField(data []byte, width int) value.Value {
	switch width {
	case 1:
		return value.Value{Lo: uint64(data[0])}
	case 2:
		return value.Value{Lo: uint64(binary.LittleEndian.Uint16(data))}
	case 4:
		return value.Value{Lo: uint64(binary.LittleEndian.Uint32(data))}
	case 8:
		return value.Value{Lo: binary.LittleEndian.Uint64(data)}
	case 16:
		return value.Value{Lo: binary.LittleEndian.Uint64(data), Hi: binary.LittleEndian.Uint64(data[8:])}
	}
	panic("state: invalid validated data field")
}
