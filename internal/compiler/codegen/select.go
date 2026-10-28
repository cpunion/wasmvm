package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Select bits without an unpredictable branch or floating-point arithmetic.
// Read every operand before writing the destination so aliases remain valid.
func (e *functionEmitter) scalarSelect(in slots.Instruction) {
	e.slotLoad(in.A, 0, false, true)
	e.slotLoad(in.B, 1, false, true)
	e.slotLoad(in.C, 2, false, false)
	left, right, condition := e.operandRegister(0), e.operandRegister(1), e.operandRegister(2)
	if e.a.arch == "amd64" {
		e.a.bytes(0x85, 0xc0|condition<<3|condition)    // test the low i32 condition.
		e.a.bytes(0x48, 0x0f, 0x44, 0xc0|left<<3|right) // cmove right into left.
	} else {
		e.a.arm(false, 0x7100001f|uint32(condition)<<5)                                // cmp wCondition, #0.
		e.a.arm(true, 0x1a800000|uint32(right)<<16|1<<12|uint32(left)<<5|uint32(left)) // csel left, left, right, ne.
	}
	if !e.unmetered {
		// The precise tier preserves both input words, including noncanonical
		// test/continuation state. These loads leave the condition flags intact.
		e.slotLoad(in.A, 1, true, true)
		e.slotLoad(in.B, 2, true, true)
		if e.a.arch == "amd64" {
			e.a.bytes(0x48, 0x0f, 0x44, 0xc0|right<<3|condition)
		} else {
			e.a.arm(true, 0x1a800000|uint32(condition)<<16|1<<12|uint32(right)<<5|uint32(right))
		}
	}
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.slotStore(in.Dst, 1, true)
	}
}
