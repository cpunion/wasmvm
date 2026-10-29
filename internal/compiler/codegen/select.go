package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Select bits without an unpredictable branch or floating-point arithmetic.
// Read every operand before writing the destination so aliases remain valid.
func (e *functionEmitter) scalarSelect(in slots.Instruction) {
	if e.unmetered && e.f.Slots[in.Dst] != wasm.F32 && e.f.Slots[in.Dst] != wasm.F64 {
		left := e.selectOperand(in.A, 0)
		right := e.selectOperand(in.B, 1)
		condition := e.selectOperand(in.C, 2)
		dst, cached := e.registers[in.Dst]
		if !cached {
			dst = e.operandRegister(0)
		}
		e.selectCondition(condition)
		e.selectBits(dst, left, right)
		if !cached {
			e.slotStore(in.Dst, 0, false)
		}
		return
	}
	e.slotLoad(in.A, 0, false, true)
	e.slotLoad(in.B, 1, false, true)
	e.slotLoad(in.C, 2, false, false)
	left, right, condition := e.operandRegister(0), e.operandRegister(1), e.operandRegister(2)
	e.selectCondition(condition)
	e.selectBits(left, left, right)
	if !e.unmetered {
		// The precise tier preserves both input words, including noncanonical
		// test/continuation state. These loads leave the condition flags intact.
		e.slotLoad(in.A, 1, true, true)
		e.slotLoad(in.B, 2, true, true)
		e.selectBits(right, right, condition)
	}
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.slotStore(in.Dst, 1, true)
	}
}

func (e *functionEmitter) selectOperand(slot slots.Slot, scratch byte) byte {
	if r, ok := e.registers[slot]; ok {
		return r
	}
	e.slotLoad(slot, scratch, false, true)
	return e.operandRegister(scratch)
}

func (e *functionEmitter) selectCondition(condition byte) {
	if e.a.arch == "arm64" {
		e.a.arm(false, 0x7100001f|uint32(condition)<<5) // cmp wCondition, #0.
	} else {
		if rex := byte(0x40) | condition>>3<<2 | condition>>3; rex != 0x40 {
			e.a.bytes(rex)
		}
		e.a.bytes(0x85, 0xc0|(condition&7)<<3|(condition&7))
	}
}

// Select left on NE and right on EQ, preserving flags for a second word.
func (e *functionEmitter) selectBits(dst, left, right byte) {
	if dst == left && dst == right {
		return
	}
	if e.a.arch == "arm64" {
		e.a.arm(true, 0x1a800000|uint32(right)<<16|1<<12|uint32(left)<<5|uint32(dst))
		return
	}
	cc, source := byte(4), right
	if dst == right && dst != left {
		cc, source = 5, left
	} else {
		e.registerMove(dst, left, true)
	}
	e.a.bytes(0x48|dst>>3<<2|source>>3, 0x0f, 0x40|cc, 0xc0|(dst&7)<<3|(source&7))
}
