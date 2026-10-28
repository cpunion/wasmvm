package slots

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type literal struct {
	typ    wasm.ValueType
	lo, hi uint64
}

func literalOf(in module.Instr, typ wasm.ValueType) (literal, bool) {
	if in.Op >= 0x41 && in.Op <= 0x44 {
		return literal{typ: typ, lo: in.Bits}, true
	}
	if in.Op == 0xfd000c {
		return literal{typ, binary.LittleEndian.Uint64(in.Vector[:8]), binary.LittleEndian.Uint64(in.Vector[8:])}, true
	}
	return literal{}, false
}

type literalSlot struct {
	value literal
	slot  Slot
	refs  uint32
}

// Only simultaneously live, block-local definitions share literal storage.
// Last-alias release removes both indexes before returning the slot for reuse.
type literalPool struct {
	values map[literal]*literalSlot
	slots  map[Slot]*literalSlot
}

func (p *literalPool) add(key literal, slot Slot, budget *budget) {
	budget.charge(3)
	if p.values == nil {
		p.values = make(map[literal]*literalSlot)
		p.slots = make(map[Slot]*literalSlot)
	}
	v := &literalSlot{key, slot, 1}
	p.values[key], p.slots[slot] = v, v
}

func (p *literalPool) release(slot Slot) bool {
	if v := p.slots[slot]; v != nil {
		v.refs--
		if v.refs != 0 {
			return false
		}
		delete(p.values, v.value)
		delete(p.slots, slot)
	}
	return true
}
