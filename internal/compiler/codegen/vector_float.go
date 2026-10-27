package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

func nativeVectorFloatOpcode(op uint32) bool {
	if op>>16 != 0xfd {
		return false
	}
	op &= 0xffff
	return op >= 65 && op <= 76 || op == 94 || op == 95 ||
		op >= 103 && op <= 106 || op == 116 || op == 117 || op == 122 || op == 148 ||
		op >= 224 && op <= 255 && op != 226 && op != 238
}

// Scalar operation numbers describe the shared IEEE behavior; the packed
// encoders below preserve all lanes and never execute a scalar guest helper.
func vectorFloatOperation(op uint32) (scalar uint32, wide, unary bool) {
	switch {
	case op >= 65 && op <= 76:
		return 0x5b + (op-65)%6, op >= 71, false
	case op >= 103 && op <= 106:
		return 0x8d + op - 103, false, true
	case op == 116 || op == 117:
		return 0x8d + op - 116, true, true
	case op == 122:
		return 0x8f, true, true
	case op == 148:
		return 0x90, true, true
	}
	wide = op >= 236
	if wide {
		op -= 12
	}
	switch {
	case op <= 225:
		return 0x8b + op - 224, wide, true
	case op == 227:
		return 0x91, wide, true
	case op <= 233:
		return 0x92 + op - 228, wide, false
	default:
		return 0x98 + op - 234, wide, false // Ordered pseudo-min/max selections.
	}
}

func (e *functionEmitter) vectorFloatInstruction(in slots.Instruction) {
	op := in.Op & 0xffff
	if op == 94 || op == 95 || op >= 248 {
		e.vectorConversion(in)
		return
	}
	scalar, wide, unary := vectorFloatOperation(op)
	l := e.vectorOperand(in.A, 0)
	r := byte(1)
	if !unary {
		r = e.vectorOperand(in.B, 1)
	}
	d := e.vectorDestination(in.Dst)
	if e.a.arch == "arm64" {
		e.vectorFloatARM64(scalar, wide, d, l, r)
	} else {
		switch {
		case scalar <= 0x60:
			e.vectorFloatCompareAMD64(scalar, wide, d, l, r)
		case scalar == 0x8b || scalar == 0x8c:
			e.vectorFloatMaskAMD64(11, wide, scalar == 0x8c)
			e.vectorMove(d, l)
			ins := byte(0xdb) // PAND preserves even signaling NaN payloads.
			if scalar == 0x8c {
				ins = 0xef
			}
			e.a.xvector(ins, d, 11)
		case scalar >= 0x8d && scalar <= 0x90:
			e.vectorRoundAMD64(scalar, wide, l)
			d = 2
		case scalar == 0x91:
			e.a.xpacked(wide, 0x51, d, l)
		case scalar >= 0x96:
			e.vectorMinMaxAMD64(scalar, wide, l, r)
			d = 2
		default:
			if d == r && d != l {
				e.vectorMove(12, r)
				r = 12
			}
			e.vectorMove(d, l)
			e.a.xpacked(wide, [...]byte{0x58, 0x5c, 0x59, 0x5e}[scalar-0x92], d, r)
		}
	}
	e.vectorStore(in.Dst, d)
}

func (a *assembler) xpacked(wide bool, opcode, dst, src byte) {
	if wide {
		a.bytes(0x66)
	}
	if rex := byte(0x40) | (dst>>3)<<2 | src>>3; rex != 0x40 {
		a.bytes(rex)
	}
	a.bytes(0x0f, opcode, 0xc0|(dst&7)<<3|(src&7))
}

func (e *functionEmitter) vectorFloatARM64(op uint32, wide bool, d, l, r byte) {
	width := uint32(0)
	if wide {
		width = 1 << 22
	}
	if op >= 0x98 {
		// Strict ordered comparison preserves the selected operand's raw bits,
		// including NaN payloads and the first operand when zeros compare equal.
		a, b := l, r
		if op == 0x99 {
			a, b = b, a
		}
		mask := d
		if d == l || d == r {
			mask = 2
		}
		e.a.word(0x6ea0e400 | width | uint32(b)<<16 | uint32(a)<<5 | uint32(mask))
		e.vectorSelectTo(d, r, l, mask)
		return
	}
	if op <= 0x60 {
		kind := op - 0x5b
		ins := uint32(0x4e20e400)
		if kind >= 2 {
			ins = 0x6ea0e400
			if kind >= 4 {
				ins = 0x6e20e400
			}
			if kind == 2 || kind == 4 {
				l, r = r, l
			}
		}
		e.a.word(ins | width | uint32(r)<<16 | uint32(l)<<5 | uint32(d))
		if kind == 1 {
			e.vectorNot(d, d)
		}
		return
	}
	ins := [...]uint32{0x4ea0f800, 0x6ea0f800, 0x4ea18800, 0x4e219800,
		0x4ea19800, 0x4e218800, 0x6ea1f800, 0x4e20d400, 0x4ea0d400,
		0x6e20dc00, 0x6e20fc00, 0x4ea0f400, 0x4e20f400}[op-0x8b]
	if op >= 0x92 {
		ins |= uint32(r) << 16
	}
	e.a.word(ins | width | uint32(l)<<5 | uint32(d))
}

func (e *functionEmitter) vectorFloatCompareAMD64(op uint32, wide bool, d, l, r byte) {
	kind := op - 0x5b
	if kind == 3 || kind == 5 {
		l, r = r, l
	}
	if d == r && d != l {
		e.vectorMove(12, r)
		r = 12
	}
	e.vectorMove(d, l)
	e.a.xpacked(wide, 0xc2, d, r)
	e.a.bytes([...]byte{0, 4, 1, 1, 2, 2}[kind])
}

// Construct lane-wise magnitude/sign masks with baseline SSE2 instructions.
func (e *functionEmitter) vectorFloatMaskAMD64(d byte, wide, sign bool) {
	e.a.xvector(0x76, d, d)
	op, count, kind := byte(0x72), byte(1), byte(2)
	if wide {
		op = 0x73
	}
	if sign {
		kind, count = 6, 31
		if wide {
			count = 63
		}
	}
	e.a.xvector(op, kind, d)
	e.a.bytes(count)
}

func (e *functionEmitter) vectorFloatConstant(d byte, wide bool, narrow, broad uint64) {
	bits, size := narrow, uint32(2)
	if wide {
		bits, size = broad, 3
	}
	e.constant(0, bits)
	e.vectorSplatBits(d, size)
}

func (e *functionEmitter) vectorMinMaxAMD64(op uint32, wide bool, l, r byte) {
	e.vectorMove(0, l)
	e.vectorMove(1, r)
	if op >= 0x98 {
		left, right := byte(0), byte(1)
		if op == 0x99 {
			left, right = right, left
		}
		e.vectorMove(2, right)
		e.a.xpacked(wide, 0xc2, 2, left) // pmin: right < left; pmax: left < right.
		e.a.bytes(1)
		e.vectorSelect(1, 0, 2)
		return
	}
	e.vectorMove(2, 0)
	e.vectorMove(12, 1)
	code, combine := byte(0x5d), byte(0xeb) // OR makes min(+0,-0) negative.
	if op == 0x97 {
		code, combine = 0x5f, 0xdb // AND makes max(+0,-0) positive.
	}
	e.a.xpacked(wide, code, 2, 1)
	e.a.xpacked(wide, code, 12, 0)
	e.a.xvector(combine, 2, 12)
	e.vectorMove(12, 0)
	e.a.xpacked(wide, 0xc2, 12, 1)
	e.a.bytes(3) // Unordered mask; either NaN must produce an arithmetic NaN.
	e.a.xpacked(wide, 0x58, 0, 1)
	e.vectorSelect(0, 2, 12)
}

// For |x| < 2^mantissa, adding/subtracting a signed 2^mantissa rounds to
// nearest-even. Large finite lanes are already integral; adding zero quiets
// NaNs. Ordered comparisons adjust the small result for ceil/floor/trunc.
func (e *functionEmitter) vectorRoundAMD64(op uint32, wide bool, l byte) {
	e.vectorMove(11, l)
	e.vectorFloatConstant(1, wide, 0x4b000000, 0x4330000000000000)
	e.vectorFloatMaskAMD64(0, wide, false)
	e.a.xvector(0xdb, 0, 11)
	e.a.xpacked(wide, 0xc2, 0, 1)
	e.a.bytes(1) // Ordered |x| < 2^mantissa.
	e.vectorFloatMaskAMD64(12, wide, true)
	e.a.xvector(0xdb, 12, 11)
	e.a.xvector(0xeb, 1, 12)
	e.vectorMove(2, 11)
	e.a.xpacked(wide, 0x58, 2, 1)
	e.a.xpacked(wide, 0x5c, 2, 1)
	e.vectorMove(1, 11)
	e.vectorZero(12)
	e.a.xpacked(wide, 0x58, 1, 12)
	e.a.xvector(0xdb, 2, 0)
	e.a.xvector(0xdf, 0, 1)
	e.a.xvector(0xeb, 2, 0)
	if op == 0x8f {
		e.vectorFloatMaskAMD64(12, wide, false)
		e.vectorMove(0, 2)
		e.vectorMove(1, 11)
		e.a.xvector(0xdb, 0, 12)
		e.a.xvector(0xdb, 1, 12)
		e.a.xpacked(wide, 0xc2, 1, 0)
		e.a.bytes(1) // |x| < |nearest(x)| needs adjustment toward zero.
		e.vectorFloatConstant(0, wide, 0x3f800000, 0x3ff0000000000000)
		e.vectorFloatMaskAMD64(12, wide, true)
		e.a.xvector(0xdb, 12, 11)
		e.a.xvector(0xeb, 0, 12)
		e.a.xvector(0xdb, 0, 1)
		e.a.xpacked(wide, 0x5c, 2, 0)
	} else if op != 0x90 {
		left, right, adjust := byte(11), byte(2), byte(0x5c)
		if op == 0x8d {
			left, right, adjust = 2, 11, 0x58
		}
		e.vectorMove(0, left)
		e.a.xpacked(wide, 0xc2, 0, right)
		e.a.bytes(1)
		e.vectorFloatConstant(1, wide, 0x3f800000, 0x3ff0000000000000)
		e.a.xvector(0xdb, 0, 1)
		e.a.xpacked(wide, adjust, 2, 0)
	}
	e.vectorFloatMaskAMD64(0, wide, true)
	e.a.xvector(0xdb, 0, 11)
	e.a.xvector(0xeb, 2, 0)
}
