package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Nonresident operands remain in the validated, execution-owned VM frame.
// x86 can read them directly without moving both inputs through RAX/RCX.
// The ordinary fully resident path and this path share one opcode inventory.
func (e *functionEmitter) frameIntegerForm(op uint32, left, right slots.Slot) (byte, bool) {
	if !e.unmetered || e.a.arch != "amd64" || left == slots.NoSlot || right == slots.NoSlot {
		return 0, false
	}
	wide := integerWide(op)
	kind := op - 0x67
	if wide {
		kind = op - 0x79
	}
	opcode := x86IntegerBinaryOpcode(kind)
	if opcode == 0 {
		return 0, false
	}
	left, right = e.frameRead(left), e.frameRead(right)
	// A cached FP/vector word cannot be read through its stale frame backing.
	for _, s := range []slots.Slot{left, right} {
		if _, ok := e.floatRegisters[s]; ok {
			return 0, false
		}
		if _, ok := e.vectorRegisters[s]; ok {
			return 0, false
		}
	}
	return opcode, wide
}

func (e *functionEmitter) frameInteger(op uint32, dst, left, right slots.Slot) bool {
	opcode, wide := e.frameIntegerForm(op, left, right)
	if opcode == 0 {
		return false
	}
	left, right = e.frameRead(left), e.frameRead(right)
	l, lok := e.registers[left]
	r, rok := e.registers[right]
	kind := op - 0x67
	if wide {
		kind = op - 0x79
	}
	d, dok := e.registers[dst]
	accumulator := d
	if !dok {
		accumulator = 0
	}
	// Commutative forms put a resident input first. Subtraction must retain
	// the right operand when it aliases the destination, until it is consumed.
	if rok && kind != 4 && (!lok || accumulator == r && l != r) {
		left, right, l, r, lok, rok = right, left, r, l, rok, lok
	}
	if kind == 4 && rok && accumulator == r && (!lok || l != r) {
		accumulator = 0
	}
	if lok {
		if accumulator != l {
			e.registerMove(accumulator, l, wide)
		}

	} else {
		e.x86FrameRead(0x8b, accumulator, left, wide)
	}
	if rok {
		e.x86RegisterBinary(opcode, accumulator, r, wide)

	} else {
		if opcode != 0xaf {
			opcode += 2 // reg <- r/m direction, rather than r/m <- reg.
		}
		e.x86FrameRead(opcode, accumulator, right, wide)
	}
	if !dok || accumulator != d {
		e.slotStore(dst, 0, false)
	}
	e.mixedIntegers++
	return true
}

func (e *functionEmitter) x86FrameRead(opcode, reg byte, slot slots.Slot, wide bool) {
	rex := byte(0x40) | (reg>>3)<<2
	if wide {
		rex |= 8
	}
	if rex != 0x40 {
		e.a.bytes(rex)
	}
	if opcode == 0xaf {
		e.a.bytes(0x0f)
	}
	e.a.bytes(opcode, 0x86|(reg&7)<<3)
	e.a.word(uint32(slot) * 16)
}
