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
	values  map[literal]int // 1-based index into entries
	slots   map[Slot]int    // 1-based index into entries
	entries []literalSlot
}

func (p *literalPool) reset() {
	if len(p.values) != 0 {
		clear(p.values)
		clear(p.slots)
	}
	p.entries = p.entries[:0]
}

func (p *literalPool) lookup(key literal) (Slot, bool) {
	if idx := p.values[key]; idx != 0 {
		e := &p.entries[idx-1]
		e.refs++
		return e.slot, true
	}
	return NoSlot, false
}

func (p *literalPool) add(key literal, slot Slot, budget *budget) {
	budget.charge(3)
	if p.values == nil {
		p.values = make(map[literal]int)
		p.slots = make(map[Slot]int)
	}
	p.entries = append(p.entries, literalSlot{key, slot, 1})
	idx := len(p.entries)
	p.values[key], p.slots[slot] = idx, idx
}

func (p *literalPool) release(slot Slot) bool {
	if idx := p.slots[slot]; idx != 0 {
		e := &p.entries[idx-1]
		e.refs--
		if e.refs != 0 {
			return false
		}
		delete(p.values, e.value)
		delete(p.slots, slot)
	}
	return true
}
