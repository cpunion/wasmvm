package codegen

// Select a literal operand without materializing a dead temporary. Every
// interior instruction retains an independent generic continuation.
func (e *functionEmitter) combineIntegerConstant(pc int) (int, bool) {
	if !e.unmetered || e.a.arch != "amd64" || pc+1 >= len(e.f.Code) || pc/256 != (pc+1)/256 {
		return 0, false
	}
	constant, in := e.f.Code[pc], e.f.Code[pc+1]
	if (constant.Op != 0x41 && constant.Op != 0x42) || constant.Dst != in.B || constant.Dst == in.A {
		return 0, false
	}
	bits := e.f.Sources[constant.Source].Wasm.Bits
	// Emit before selecting the arithmetic only when the constant remains live.
	// Selection may fail, so roll back this bounded speculative emission.
	before := len(e.a.code)
	if constant.Dst != in.Dst && !e.overwrittenBeforeExit(pc+2, constant.Dst) {
		e.literal(constant)
	}
	if !e.registerImmediate(in.Op, in.Dst, in.A, bits) {
		e.a.code = e.a.code[:before]
		return 0, false
	}
	return pc + 1, true
}

func (e *functionEmitter) x86ArithmeticImmediate(op uint32, dst, left byte, bits uint64) bool {
	wide, first := integerWide(op), uint32(0x6a)
	if wide {
		first = 0x7c
		if uint64(int64(int32(bits))) != bits {
			return false
		}
	}
	var ext byte
	switch op - first {
	case 0: // add
		ext = 0
	case 1: // sub
		ext = 5
	case 2: // mul: the immediate form has a separate source and destination.
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
	case 7: // and
		ext = 4
	case 8: // or
		ext = 1
	case 9: // xor
		ext = 6
	default:
		return false
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
