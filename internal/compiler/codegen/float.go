package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

func isFloat(op uint32) bool {
	return op >= 0x5b && op <= 0x66 || op >= 0x8b && op <= 0xa6 || op >= 0xbc && op <= 0xbf
}

// Floating-point values share the existing scalar slot/register contract. FP
// scratch registers never survive an instruction or an exit. The C boundary
// establishes round-to-nearest, gradual underflow and masked FP exceptions.
func (e *functionEmitter) floating(in slots.Instruction) {
	if e.reinterpretRegister(in) || e.registerFloat(in) {
		return
	}
	op := in.Op
	wide := op >= 0x61 && op <= 0x66 || op >= 0x99 && op <= 0xa6 || op == 0xbd || op == 0xbf
	e.slotLoad(in.A, 0, false, wide)
	e.slotLoad(in.B, 1, false, wide)
	if op < 0xbc {
		if op >= 0x99 {
			op -= 14 // Share f32/f64 arithmetic instruction selection.
		} else if op >= 0x61 && op <= 0x66 {
			op -= 6
		}
		switch op {
		case 0x8b, 0x8c, 0x98:
			e.floatSign(op, wide)
		default:
			if e.a.arch == "arm64" {
				e.floatARM64(op, wide)
			} else {
				e.floatAMD64(op, wide)
			}
		}
	}
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(in.Dst, 0, true)
	}
}

// Sign instructions are bit operations: they preserve signaling NaN payloads.
func (e *functionEmitter) floatSign(op uint32, wide bool) {
	mask := uint64(1 << 31)
	and, xor := uint32(0x71), uint32(0x73)
	if wide {
		mask, and, xor = 1<<63, 0x83, 0x85
	}
	core := func(op uint32) {
		if e.a.arch == "amd64" {
			integerAMD64Core(e.a, op, wide, e.zero, e.overflow)
		} else {
			integerARM64Core(e.a, op, wide, e.zero, e.overflow)
		}
	}
	if op == 0x98 {
		// Save the magnitude while extracting the second operand's sign.
		e.registerMove(e.operandRegister(2), e.operandRegister(0), wide)
		core(xor)
		e.constant(1, mask)
		core(and)
		e.registerMove(e.operandRegister(1), e.operandRegister(2), wide)
		core(xor)
	} else if op == 0x8b {
		e.constant(1, ^mask)
		core(and)
	} else {
		e.constant(1, mask)
		core(xor)
	}
}

func (e *functionEmitter) floatARM64(op uint32, wide bool) {
	a := e.a
	width := uint32(0)
	if wide {
		width = 1 << 22
		a.word(0x9e670120) // fmov d0, x9
		a.word(0x9e670141) // fmov d1, x10
	} else {
		a.word(0x1e270120) // fmov s0, w9
		a.word(0x1e270141) // fmov s1, w10
	}
	if op <= 0x60 {
		a.word(0x1e212000 | width) // fcmp s0/d0, s1/d1
		// MI and LS reject unordered operands, unlike signed LT and LE.
		conditions := [...]uint32{0, 1, 4, 12, 9, 10}
		a.word(0x1a9f07e9 | (conditions[op-0x5b]^1)<<12)
		return
	}
	var ins uint32
	switch op {
	case 0x8d:
		ins = 0x1e24c000 // frintp
	case 0x8e:
		ins = 0x1e254000 // frintm
	case 0x8f:
		ins = 0x1e25c000 // frintz
	case 0x90:
		ins = 0x1e244000 // frintn
	case 0x91:
		ins = 0x1e21c000 // fsqrt
	case 0x92:
		ins = 0x1e212800 // fadd
	case 0x93:
		ins = 0x1e213800 // fsub
	case 0x94:
		ins = 0x1e210800 // fmul
	case 0x95:
		ins = 0x1e211800 // fdiv
	case 0x96:
		ins = 0x1e215800 // fmin (propagate NaN, preserve signed-zero ordering)
	case 0x97:
		ins = 0x1e214800 // fmax
	}
	a.word(ins | width)
	if wide {
		a.word(0x9e660009) // fmov x9, d0
	} else {
		a.word(0x1e260009) // fmov w9, s0
	}
}

func (a *assembler) xfloatMove(wide bool, toFloat bool, fp, integer byte) {
	a.bytes(0x66)
	rex := byte(0x40) | (fp>>3)<<2 | integer>>3
	if wide {
		rex |= 8
	}
	if rex != 0x40 {
		a.bytes(rex)
	}
	op := byte(0x6e)
	if !toFloat {
		op = 0x7e
	}
	a.bytes(0x0f, op, 0xc0|(fp&7)<<3|(integer&7))
}
func (a *assembler) xfloat(wide bool, op, dst, src byte) {
	prefix := byte(0xf3)
	if wide {
		prefix = 0xf2
	}
	a.bytes(prefix)
	if rex := byte(0x40) | (dst>>3)<<2 | src>>3; rex != 0x40 {
		a.bytes(rex)
	}
	a.bytes(0x0f, op, 0xc0|(dst&7)<<3|(src&7))
}
func (a *assembler) xfloatCompare(wide bool) {
	if wide {
		a.bytes(0x66)
	}
	a.bytes(0x0f, 0x2e, 0xc1) // ucomiss/ucomisd xmm0, xmm1
}

func (e *functionEmitter) floatAMD64(op uint32, wide bool) {
	a := e.a
	a.xfloatMove(wide, true, 0, 0)
	a.xfloatMove(wide, true, 1, 1)
	if op <= 0x60 {
		predicate := [...]byte{0, 4, 1, 1, 2, 2}
		dst, src := byte(0), byte(1)
		if op == 0x5e || op == 0x60 {
			dst, src = src, dst
		}
		a.xfloat(wide, 0xc2, dst, src)
		a.bytes(predicate[op-0x5b])
		a.xfloatMove(false, false, dst, 0)
		a.bytes(0x83, 0xe0, 1) // comparison mask -> canonical i32 boolean
		return
	}
	switch op {
	case 0x8d, 0x8e, 0x8f, 0x90:
		e.floatRoundAMD64(op, wide)
		return
	case 0x91:
		a.xfloat(wide, 0x51, 0, 0)
	case 0x92, 0x93, 0x94, 0x95:
		instructions := [...]byte{0x58, 0x5c, 0x59, 0x5e}
		a.xfloat(wide, instructions[op-0x92], 0, 1)
	case 0x96, 0x97:
		nan, ordered, done := a.label(), a.label(), a.label()
		a.xfloatCompare(wide)
		a.branch(nan, true, 10) // unordered
		a.branch(ordered, true, 5)
		// Equal non-NaN values differ only in a possible zero sign bit.
		bitop := byte(0x56) // orps: min(+0,-0) = -0
		if op == 0x97 {
			bitop = 0x54 // andps: max(+0,-0) = +0
		}
		a.bytes(0x0f, bitop, 0xc1)
		a.branch(done, false, 0)
		a.mark(nan)
		a.xfloat(wide, 0x58, 0, 1) // arithmetic NaN, including signaling inputs
		a.branch(done, false, 0)
		a.mark(ordered)
		instruction := byte(0x5d)
		if op == 0x97 {
			instruction = 0x5f
		}
		a.xfloat(wide, instruction, 0, 1)
		a.mark(done)
	}
	a.xfloatMove(wide, false, 0, 0)
}

// SSE2 is the baseline: avoid an undeclared SSE4.1 dependency for rounding.
// Below 2^mantissa, truncation to i64 and exact conversion back are safe.
// Larger finite inputs are already integral; adding zero quiets NaNs there.
func (e *functionEmitter) floatRoundAMD64(op uint32, wide bool) {
	a := e.a
	large, finish := a.label(), a.label()
	sign, integral, one := uint64(1<<31), uint64(0x4b000000), uint64(0x3f800000)
	prefix := byte(0xf3)
	if wide {
		sign, integral, one = 1<<63, 0x4330000000000000, 0x3ff0000000000000
		prefix = 0xf2
	}
	e.constant(1, sign)
	a.xop(wide, 0x21, 0xc1) // rcx = original sign
	a.xop(wide, 0x89, 0xc2) // rdx = original bits
	a.xop(wide, 0x31, 0xca) // rdx = absolute bits
	e.constant(0, integral)
	a.xop(wide, 0x39, 0xc2)
	a.branch(large, true, 3) // unsigned abs >= 2^mantissa (includes NaN)
	if op == 0x90 {
		a.xop(wide, 0x09, 0xc8) // signed 2^mantissa
		a.xfloatMove(wide, true, 2, 0)
		a.xfloat(wide, 0x58, 0, 2)
		a.xfloat(wide, 0x5c, 0, 2)
	} else {
		a.bytes(0x0f, 0x28, 0xc8)               // xmm1 = original input
		a.bytes(prefix, 0x48, 0x0f, 0x2c, 0xc0) // cvttss/sd2si rax, xmm0
		a.bytes(prefix, 0x48, 0x0f, 0x2a, 0xc0) // cvtsi2ss/sd xmm0, rax
		if op != 0x8f {
			a.xfloatCompare(wide)
			cc, adjust := byte(6), byte(0x5c) // floor: skip when rounded <= original
			if op == 0x8d {
				cc, adjust = 3, 0x58 // ceil: skip when rounded >= original
			}
			a.branch(finish, true, cc)
			e.constant(2, one)
			a.xfloatMove(wide, true, 2, 2)
			a.xfloat(wide, adjust, 0, 2)
		}
	}
	a.branch(finish, false, 0)
	a.mark(large)
	a.bytes(0x0f, 0x57, 0xd2) // xorps xmm2, xmm2
	a.xfloat(wide, 0x58, 0, 2)
	a.mark(finish)
	a.xfloatMove(wide, false, 0, 0)
	a.xop(wide, 0x09, 0xc8) // Preserve negative zero after rounding.
}

func (e *functionEmitter) reinterpretRegister(in slots.Instruction) bool {
	switch in.Op {
	case 0xbc, 0xbd:
		s, sok := e.floatRegisters[in.A]
		d, dok := e.registers[in.Dst]
		if sok && dok {
			e.floatToInteger(d, s, in.Op == 0xbd)
			return true
		}
	case 0xbe, 0xbf:
		s, sok := e.registers[in.A]
		d, dok := e.floatRegisters[in.Dst]
		if sok && dok {
			e.floatFromInteger(d, s, in.Op == 0xbf)
			return true
		}
	}
	return false
}
