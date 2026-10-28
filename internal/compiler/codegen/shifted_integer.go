package codegen

// Select a shifted second operand after a constant shift. The intermediate
// must be overwritten before any observation; independent PCs keep the original
// constant/shift/consumer sequence. All inputs are read before the result write.
func (e *functionEmitter) combineShiftedInteger(pc int) (int, bool) {
	if !e.unmetered || e.a.arch != "arm64" || pc+2 >= e.codeLimit() || pc/256 != (pc+2)/256 {
		return 0, false
	}
	constant, shift, in := e.f.Code[pc], e.f.Code[pc+1], e.f.Code[pc+2]
	wide := constant.Op == 0x42
	literal, first, add := uint32(0x41), uint32(0x74), uint32(0x6a)
	width := uint32(32)
	if wide {
		literal, first, add, width = 0x42, 0x86, 0x7c, 64
	}
	if constant.Op != literal || shift.Op < first || shift.Op > first+4 || shift.B != constant.Dst || shift.A == constant.Dst {
		return 0, false
	}
	word := uint32(0)
	switch in.Op - add {
	case 0:
		word = 0x0b000000
	case 1:
		word = 0x4b000000
	case 7:
		word = 0x0a000000
	case 8:
		word = 0x2a000000
	case 9:
		word = 0x4a000000
	default:
		return 0, false
	}
	left := in.A
	if in.B != shift.Dst {
		if in.A != shift.Dst || in.Op == add+1 {
			return 0, false
		}
		left = in.B
	}
	if left == shift.Dst || shift.Dst != in.Dst && !e.overwrittenBeforeExit(pc+3, shift.Dst) {
		return 0, false
	}
	l, lok := e.registers[left]
	r, rok := e.registers[shift.A]
	d, dok := e.registers[in.Dst]
	if !lok || !rok || !dok {
		return 0, false
	}
	kind := shift.Op - first
	mode := [...]uint32{0, 2, 1, 3, 3}[kind]
	if mode == 3 && (in.Op == add || in.Op == add+1) {
		return 0, false // ADD/SUB have no rotated-register form.
	}
	count := uint32(e.f.Sources[constant.Source].Wasm.Bits) & (width - 1)
	if kind == 3 {
		count = (width - count) & (width - 1)
	}
	if left == constant.Dst || constant.Dst != shift.Dst && constant.Dst != in.Dst && !e.overwrittenBeforeExit(pc+3, constant.Dst) {
		e.literal(constant)
	}
	e.a.arm(wide, word|mode<<22|uint32(r)<<16|count<<10|uint32(l)<<5|uint32(d))
	return pc + 2, true
}
