package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Select a literal operand without materializing a dead temporary. Every
// interior instruction retains an independent generic continuation.
func (e *functionEmitter) combineIntegerConstant(pc int) (int, bool) {
	if !e.unmetered || e.a.arch != "amd64" || pc+1 >= e.codeLimit() || pc/256 != (pc+1)/256 {
		return 0, false
	}
	constant, in := e.f.Code[pc], e.f.Code[pc+1]
	if (constant.Op != 0x41 && constant.Op != 0x42) || constant.Dst != in.B || constant.Dst == in.A {
		return 0, false
	}
	bits := e.f.Sources[constant.Source].Wasm.Bits
	end, original := pc+1, in.Dst
	if pc+2 < e.codeLimit() && pc/256 == (pc+2)/256 {
		next := e.f.Code[pc+2]
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == original {
			dst := slots.Slot(e.f.Sources[next.Source].Wasm.A)
			if e.canRegisterImmediate(in.Op, dst, in.A, bits) {
				in.Dst, end = dst, pc+2
			}
		}
	}
	if !e.canRegisterImmediate(in.Op, in.Dst, in.A, bits) {
		return 0, false
	}
	if constant.Dst != in.Dst && !e.overwrittenBeforeExit(pc+2, constant.Dst) {
		e.literal(constant)
	}
	e.registerImmediate(in.Op, in.Dst, in.A, bits)
	if end != pc+1 {
		if !e.overwrittenBeforeExit(end+1, original) {
			e.move(original, in.Dst)
		}
		e.move(e.f.Code[end].Dst, in.Dst)
	}
	return end, true
}

func x86ImmediateGroup(op uint32, bits uint64) (byte, bool) {
	wide, first := integerWide(op), uint32(0x6a)
	if wide {
		first = 0x7c
		if uint64(int64(int32(bits))) != bits {
			return 0, false
		}
	}
	switch op - first {
	case 0: // add
		return 0, true
	case 1: // sub
		return 5, true
	case 2: // mul
		return 8, true
	case 7: // and
		return 4, true
	case 8: // or
		return 1, true
	case 9: // xor
		return 6, true
	}
	return 0, false
}

func (e *functionEmitter) x86ArithmeticImmediate(op uint32, dst, left byte, bits uint64) bool {
	ext, ok := x86ImmediateGroup(op, bits)
	if !ok {
		return false
	}
	wide := integerWide(op)
	switch ext {
	case 8: // mul: the immediate form has a separate source and destination.
		rex := byte(0x40) | (dst>>3)<<2 | left>>3
		if wide {
			rex |= 8
		}
		if rex != 0x40 {
			e.a.bytes(rex)
		}
		e.a.bytes(0x69, 0xc0|(dst&7)<<3|left&7)
		e.a.word(uint32(bits))
		return true
	}
	if dst != left {
		e.registerMove(dst, left, wide)
	}
	rex := byte(0x40) | dst>>3
	if wide {
		rex |= 8
	}
	if rex != 0x40 {
		e.a.bytes(rex)
	}
	if int32(bits) >= -128 && int32(bits) <= 127 {
		e.a.bytes(0x83, 0xc0|ext<<3|dst&7, byte(bits))
	} else {
		e.a.bytes(0x81, 0xc0|ext<<3|dst&7)
		e.a.word(uint32(bits))
	}
	return true
}
