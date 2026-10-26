package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Branch on the i32 condition in its allocated register. Keep long-range
// relocation support on arm64 by inverting CBZ over an unconditional branch.
func (e *functionEmitter) branchZero(slot slots.Slot, label int) {
	r, ok := e.registers[slot]
	if !ok {
		e.testSlot(slot)
		e.branch(label, "eq")
		return
	}
	if e.a.arch == "arm64" {
		e.a.word(0x35000040 | uint32(r)) // cbnz wR, +8
		e.a.branch(label, false, 0)
	} else {
		if rex := byte(0x40) | (r>>3)<<2 | r>>3; rex != 0x40 {
			e.a.bytes(rex)
		}
		e.a.bytes(0x85, 0xc0|(r&7)<<3|(r&7)) // test r32, r32
		e.branch(label, "eq")
	}
}
