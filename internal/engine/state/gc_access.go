package state

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// GCAccessArity identifies the nonallocating checked scalar access contract.
func GCAccessArity(op uint32) int {
	switch op & 0xffff {
	case 2, 3, 4, 15, 28, 29, 30, 34:
		return 1
	case 5, 11, 12, 13:
		return 2
	case 14:
		return 3
	}
	return 0
}

// GCAccess shares the typed heap and complete handle checks.
// Separate them from allocating callbacks so both host compilers can keep
// ordinary interpreter access free of closure construction and safepoints.
func (s *Store) GCAccess(in *Instance, ins *module.Instr, a, b, c value.Value) (value.Value, error) {
	op := ins.Op & 0xffff
	var result value.Value
	if op == 28 {
		return value.Value{Lo: I31(a.Lo)}, nil
	}
	ref := a.Lo
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
	index := ins.B
	field := fields[0]
	if op <= 5 {
		field = fields[index]
	} else {
		index = uint32(b.Lo)
		if index >= o.Length {
			return result, &wasm.Trap{Code: wasm.TrapArrayBounds, Offset: ins.Offset}
		}
	}
	if op == 5 || op == 14 {
		v := b
		if op == 14 {
			v = c
		}
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
