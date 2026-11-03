package state

import (
	"context"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func gcTrap(ins *module.Instr, code wasm.TrapCode) error {
	return &wasm.Trap{Code: code, Offset: ins.Offset}
}

func (s *Store) gcObject(ins *module.Instr, ref uint64) (*heapObject, error) {
	if ref == 0 {
		return nil, gcTrap(ins, wasm.TrapNullReference)
	}
	o, ok := s.heapObject(ref)
	if !ok {
		return nil, ErrReference
	}
	return o, nil
}

func (s *Store) gcElement(in *Instance, index uint32) module.FieldType {
	return s.registry.Types[in.Types[index]].Definition.Fields[0]
}

func packField(v value.Value, f module.FieldType) value.Value {
	if f.Packed != 0 {
		v.Lo &= (1 << f.Packed) - 1
		v.Hi = 0
	}
	return v
}

func gcPoll(ctx context.Context, index int) error {
	if index&255 == 0 && ctx != nil {
		return ctx.Err()
	}
	return nil
}

func (s *Store) popTemporary(mark int) {
	clear(s.temporary[mark:])
	s.temporary = s.temporary[:mark]
}

func (s *Store) gcAllocate(in *Instance, ins *module.Instr, args []value.Value, length uint32) (uint64, *heapObject, error) {
	// Default constructors have no operand roots to borrow. Keep their path
	// outside the rooted helper's deferred cleanup and exception frame.
	if len(args) == 0 {
		return s.gcAllocateObject(in.Types[ins.A], length)
	}
	return s.gcAllocateRooted(in, ins, args, length)
}

func (s *Store) gcAllocateRooted(in *Instance, ins *module.Instr, args []value.Value, length uint32) (uint64, *heapObject, error) {
	// Operands stay rooted even on a constant-expression stack. Restore the
	// borrowed root extent on ordinary errors and host-collector panics alike.
	mark := len(s.temporary)
	defer s.popTemporary(mark)
	fs := s.registry.Types[in.Types[ins.A]].Definition.Fields
	op := ins.Op & 0xffff
	for i, a := range args {
		ft := wasm.I32
		if (op == 0 || op == 32) && i < len(fs) {
			ft = fs[i].Unpacked()
		} else if (op == 32 || op == 33) && i == len(args)-1 {
			ft = wasm.AnyRef
		} else if op == 6 && i == 0 || op == 8 {
			ft = fs[0].Unpacked()
		}
		if ft.IsReference() {
			s.temporary = appendValue(s.temporary, ReferenceValue{ft, a})
		}
	}
	return s.gcAllocateObject(in.Types[ins.A], length)
}

func (s *Store) gcAllocateObject(typ, length uint32) (uint64, *heapObject, error) {
	ref, err := s.allocateObject(typ, length)
	if err != nil {
		return 0, nil, err
	}
	o, _ := s.heapObject(ref)
	return ref, o, nil
}
