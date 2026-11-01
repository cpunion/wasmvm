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
	if n := GCAccessArity(ins.Op); n != 0 {
		var b, c value.Value
		if n > 1 {
			b = args[1]
		}
		if n > 2 {
			c = args[2]
		}
		return s.GCAccess(in, ins, args[0], b, c)
	}
	if s.HeapSafepoint == nil && (ins.Op&0xffff <= 1 || ins.Op&0xffff >= 6 && ins.Op&0xffff <= 10 || ins.Op&0xffff == 32 || ins.Op&0xffff == 33) {
		return s.gcRooted(ctx, in, ins, args, roots...)
	}
	return s.gcSlow(ctx, in, ins, args)
}

// gcRooted supplies a temporary collector only to internal callers that do not
// already have the embedding Store's complete execution/root visitor.
func (s *Store) gcRooted(ctx context.Context, in *Instance, ins *module.Instr, args []value.Value, roots ...func(func(wasm.ValueType, value.Value))) (value.Value, error) {
	s.HeapSafepoint = func() {
		var xs []ReferenceValue
		if len(roots) != 0 {
			roots[0](func(t wasm.ValueType, v value.Value) { xs = appendValue(xs, ReferenceValue{t, v}) })
		}
		active := append([]*Instance{in}, s.instances...)
		s.CollectReferences(xs, active, nil)
	}
	defer func() { s.HeapSafepoint = nil }()
	return s.gcSlow(ctx, in, ins, args)
}

// Allocation and bulk/cast operations preserve their checked safepoints without
// constructing unrelated closures for each operation.
func (s *Store) gcSlow(ctx context.Context, in *Instance, ins *module.Instr, args []value.Value) (value.Value, error) {
	op := ins.Op & 0xffff
	var result value.Value
	switch op {
	case 0, 1, 32, 33:
		fs := s.registry.Types[in.Types[ins.A]].Definition.Fields
		var descriptor uint64
		if op >= 32 {
			descriptor = args[len(args)-1].Lo
			if descriptor == 0 {
				return result, gcTrap(ins, wasm.TrapNullReference)
			}
		}
		ref, o, err := s.gcAllocate(in, ins, args, uint32(len(fs)))
		result.Lo = ref
		if err != nil {
			return result, err
		}
		o.Descriptor = descriptor
		if op == 0 || op == 32 {
			for i, a := range args[:len(fs)] {
				o.set(uint32(i), packField(a, fs[i]))
			}
		}
	case 6, 7, 8:
		n := ins.B
		if op == 6 {
			n = uint32(args[1].Lo)
		} else if op == 7 {
			n = uint32(args[0].Lo)
		}
		ref, o, err := s.gcAllocate(in, ins, args, n)
		result.Lo = ref
		if err != nil {
			return result, err
		}
		f := s.gcElement(in, ins.A)
		if op != 7 {
			for i := range int(o.Length) {
				if err := gcPoll(ctx, i); err != nil {
					return result, err
				}
				if op == 6 {
					o.set(uint32(i), packField(args[0], f))
				} else if op == 8 {
					o.set(uint32(i), packField(args[i], f))
				}
			}
		}
		if ctx != nil && ctx.Err() != nil {
			return result, ctx.Err()
		}
	case 9, 10:
		start, n := uint32(args[0].Lo), uint32(args[1].Lo)
		f := s.gcElement(in, ins.A)
		if op == 9 {
			width := fieldBytes(f)
			data := in.Data[ins.B]
			if uint64(start) > uint64(len(data)) || uint64(n)*uint64(width) > uint64(len(data))-uint64(start) {
				return result, gcTrap(ins, wasm.TrapMemoryBounds)
			}
			ref, o, err := s.gcAllocate(in, ins, args, n)
			result.Lo = ref
			if err != nil {
				return result, err
			}
			for i := range int(o.Length) {
				if err := gcPoll(ctx, i); err != nil {
					return result, err
				}
				o.set(uint32(i), readField(data[int(start)+i*width:], width))
			}
		} else {
			data := in.Elements[ins.B]
			if !Range(len(data), uint64(start), uint64(n)) {
				return result, gcTrap(ins, wasm.TrapTableBounds)
			}
			ref, o, err := s.gcAllocate(in, ins, args, n)
			result.Lo = ref
			if err != nil {
				return result, err
			}
			for i := range int(o.Length) {
				if err := gcPoll(ctx, i); err != nil {
					return result, err
				}
				o.set(uint32(i), value.Value{Lo: data[int(start)+i]})
			}
		}
	case 16:
		o, err := s.gcObject(ins, args[0].Lo)
		if err != nil {
			return result, err
		}
		start, n := uint32(args[1].Lo), uint32(args[3].Lo)
		if !Range(int(o.Length), uint64(start), uint64(n)) {
			return result, gcTrap(ins, wasm.TrapArrayBounds)
		}
		v := packField(args[2], s.gcElement(in, ins.A))
		for i := uint32(0); i < n; i++ {
			if err := gcPoll(ctx, int(i)); err != nil {
				return result, err
			}
			o.set(start+i, v)
		}
	case 17:
		dst, err := s.gcObject(ins, args[0].Lo)
		if err != nil {
			return result, err
		}
		src, err := s.gcObject(ins, args[2].Lo)
		if err != nil {
			return result, err
		}
		di, si, n := uint32(args[1].Lo), uint32(args[3].Lo), uint32(args[4].Lo)
		if !Range(int(dst.Length), uint64(di), uint64(n)) || !Range(int(src.Length), uint64(si), uint64(n)) {
			return result, gcTrap(ins, wasm.TrapArrayBounds)
		}
		// Poll without losing memmove overlap semantics across chunk boundaries.
		backwards := dst == src && di > si && uint64(di) < uint64(si)+uint64(n)
		for done := uint32(0); done < n; {
			if err := gcPoll(ctx, 0); err != nil {
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
		o, err := s.gcObject(ins, args[0].Lo)
		if err != nil {
			return result, err
		}
		di, si, n := uint32(args[1].Lo), uint32(args[2].Lo), uint32(args[3].Lo)
		if !Range(int(o.Length), uint64(di), uint64(n)) {
			return result, gcTrap(ins, wasm.TrapArrayBounds)
		}
		if op == 18 {
			width := fieldBytes(s.gcElement(in, ins.A))
			data := in.Data[ins.B]
			if uint64(si) > uint64(len(data)) || uint64(n)*uint64(width) > uint64(len(data))-uint64(si) {
				return result, gcTrap(ins, wasm.TrapMemoryBounds)
			}
			for i := uint32(0); i < n; i++ {
				if err := gcPoll(ctx, int(i)); err != nil {
					return result, err
				}
				o.set(di+i, readField(data[int(si)+int(i)*width:], width))
			}
		} else {
			data := in.Elements[ins.B]
			if !Range(len(data), uint64(si), uint64(n)) {
				return result, gcTrap(ins, wasm.TrapTableBounds)
			}
			for i := uint32(0); i < n; i++ {
				if err := gcPoll(ctx, int(i)); err != nil {
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
				return result, gcTrap(ins, wasm.TrapCastFailure)
			}
			result = args[0]
		}
	case 26:
		result.Lo = Internalize(args[0].Lo)
	case 27:
		result.Lo = Externalize(args[0].Lo)
	case 35, 36:
		matched, err := s.DescriptorEqual(args[0].Lo, args[1].Lo, ins.ReferenceType())
		if err != nil {
			return result, gcTrap(ins, wasm.TrapNullReference)
		}
		if !matched {
			return result, gcTrap(ins, wasm.TrapCastFailure)
		}
		result = args[0]
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
