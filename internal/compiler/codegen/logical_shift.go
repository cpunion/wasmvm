package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Distribute a constant shift/rotate over AND/OR/XOR. AArch64 can shift the
// second logical operand in the same instruction. Preparing the right operand
// first shortens the left operand's dependent path from two instructions to one.
// The original intermediate may disappear only when overwritten before an exit.
func (e *functionEmitter) logicalShift(pc int, in slots.Instruction) bool {
	if !e.unmetered || e.a.arch != "arm64" || pc+2 >= len(e.f.Code) || pc/256 != (pc+2)/256 {
		return false
	}
	wide, first, literal := integerWide(in.Op), uint32(0x71), uint32(0x41)
	if wide {
		first, literal = 0x83, 0x42
	}
	if in.Op < first || in.Op > first+2 {
		return false
	}
	constant, shift := e.f.Code[pc+1], e.f.Code[pc+2]
	if constant.Op != literal || shift.Op < first+3 || shift.Op > first+7 ||
		shift.A != e.f.Code[pc].Dst || shift.B != constant.Dst || constant.Dst == e.f.Code[pc].Dst {
		return false
	}
	if shift.Dst != e.f.Code[pc].Dst && !e.overwrittenBeforeExit(pc+3, e.f.Code[pc].Dst) {
		return false
	}
	l, lok := e.registers[in.A]
	r, rok := e.registers[in.B]
	d, dok := e.registers[shift.Dst]
	if !lok || !rok || !dok {
		return false
	}
	width := uint32(32)
	if wide {
		width = 64
	}
	count := uint32(e.f.Sources[constant.Source].Wasm.Bits) & (width - 1)
	kind := shift.Op - first - 3
	mode := [...]uint32{0, 2, 1, 3, 3}[kind] // LSL, ASR, LSR, ROR, ROR
	if kind == 3 {
		count = (width - count) & (width - 1)
	}
	// x13 is excluded from slot and constant allocation. Neither instruction
	// changes NZCV, and all operands are consumed before restoring constants.
	word := uint32(0x53000000)
	if mode == 3 {
		word = 0x13800000 | uint32(r)<<16 | count<<10
	} else {
		immr, imms := count, width-1
		if mode == 0 {
			immr, imms = (width-count)&(width-1), width-1-count
		} else if mode == 2 {
			word = 0x13000000
		}
		word |= immr<<16 | imms<<10
	}
	if wide {
		word |= 1<<31 | 1<<22
	}
	e.a.word(word | uint32(r)<<5 | 13)
	logical := [...]uint32{0x0a000000, 0x2a000000, 0x4a000000}[in.Op-first]
	e.a.arm(wide, logical|mode<<22|uint32(l)<<16|count<<10|13<<5|uint32(d))
	if constant.Dst != shift.Dst && !e.overwrittenBeforeExit(pc+3, constant.Dst) {
		e.literal(constant)
	}
	return true
}
