package state

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Nonallocating operations share the typed heap and complete handle checks.
// Separate them from allocating callbacks so both host compilers can keep
// ordinary interpreter access free of closure construction and safepoints.
func (s *Store) gcAccess(in *Instance, ins *module.Instr, args []value.Value) (value.Value, error) {
	op := ins.Op & 0xffff
	var result value.Value
	if op == 28 {
		return value.Value{Lo: I31(args[0].Lo)}, nil
	}
	ref := args[0].Lo
	if ref == 0 {
		return result, &wasm.Trap{Code: wasm.TrapNullReference, Offset: ins.Offset}
	}
	if op == 29 || op == 30 {
		result.Lo = ref & 0x7fffffff
		if op == 29 {
			result.Lo = uint64(uint32(int32(uint32(result.Lo)<<1) >> 1))
		}
		return result, nil
	}
	o, ok := s.heapObject(ref)
	if !ok {
		return result, ErrReference
	}
	if op == 15 {
		return value.Value{Lo: uint64(o.Length)}, nil
	}
	if op == 34 {
		return value.Value{Lo: o.Descriptor}, nil
	}
	fields := s.registry.Types[in.Types[ins.A]].Definition.Fields
	index, operand := ins.B, 1
	field := fields[0]
	if op <= 5 {
		field = fields[index]
	} else {
		index, operand = uint32(args[1].Lo), 2
		if index >= o.Length {
			return result, &wasm.Trap{Code: wasm.TrapArrayBounds, Offset: ins.Offset}
		}
	}
	if op == 5 || op == 14 {
		v := args[operand]
		if field.Packed != 0 {
			v.Lo &= (1 << field.Packed) - 1
			v.Hi = 0
		}
		o.set(index, v)
		return result, nil
	}
	result = o.get(index)
	if op == 3 || op == 12 {
		result.Lo = uint64(uint32(int32(uint32(result.Lo)<<(32-field.Packed)) >> (32 - field.Packed)))
	}
	return result, nil
}
