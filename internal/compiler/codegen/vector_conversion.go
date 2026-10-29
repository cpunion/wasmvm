package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

func (e *functionEmitter) vectorConversion(in slots.Instruction) {
	op := in.Op & 0xffff
	l := e.vectorOperand(in.A, 0)
	d := e.vectorDestination(in.Dst)
	if e.a.arch == "arm64" {
		var ins uint32
		switch op {
		case 94:
			ins = 0x0e616800 // FCVTN also clears the upper two f32 lanes.
		case 95:
			ins = 0x0e617800 // FCVTL consumes only the low two f32 lanes.
		case 248, 249:
			ins = 0x4ea1b800 | (op&1)<<29
		case 250, 251:
			ins = 0x4e21d800 | (op&1)<<29
		case 252, 253:
			e.a.word(0x4ee1b800 | (op&1)<<29 | uint32(l)<<5 | uint32(d))
			l, ins = d, 0x0ea14800|(op&1)<<29 // Saturate widened integer lanes to i32.
		case 254, 255:
			e.a.word(0x0f20a400 | (op&1)<<29 | uint32(l)<<5 | uint32(d))
			l, ins = d, 0x4e61d800|(op&1)<<29 // Widen i32 before converting to f64.
		}
		e.a.word(ins | uint32(l)<<5 | uint32(d))
	} else {
		switch op {
		case 94, 95:
			e.a.xpacked(op == 94, 0x5a, d, l)
		case 248, 249, 252, 253:
			e.vectorSaturateAMD64(op >= 252, op&1 != 0, l)
			d = 2
		case 250:
			e.a.xpacked(false, 0x5b, d, l) // CVTDQ2PS
		case 254:
			e.a.xfloat(false, 0xe6, d, l) // CVTDQ2PD
		case 251:
			// Both 16-bit halves convert exactly. Their scaled sum rounds only
			// once, unlike adding 2^31 after an already-rounded signed cast.
			e.vectorMove(2, l)
			e.vectorMove(0, l)
			e.a.xvector(0x72, 2, 2)
			e.a.bytes(16)
			e.a.xvector(0x72, 6, 0)
			e.a.bytes(16)
			e.a.xvector(0x72, 2, 0)
			e.a.bytes(16)
			e.a.xpacked(false, 0x5b, 2, 2)
			e.a.xpacked(false, 0x5b, 0, 0)
			e.vectorFloatConstant(1, false, 0x47800000, 0)
			e.a.xpacked(false, 0x59, 2, 1)
			e.a.xpacked(false, 0x58, 2, 0)
			d = 2
		case 255:
			// IEEE binary64 represents every u32 exactly. Build 2^52+x from
			// integer lane bits, then subtract the exactly represented bias.
			e.vectorMove(2, l)
			e.vectorZero(0)
			e.a.xvector(0x62, 2, 0)
			e.vectorFloatConstant(1, true, 0, 0x4330000000000000)
			e.a.xvector(0xeb, 2, 1)
			e.a.xpacked(true, 0x5c, 2, 1)
			d = 2
		}
	}
	e.vectorStore(in.Dst, d)
}

func (e *functionEmitter) vectorTruncateAMD64(wide bool, d, s byte) {
	if wide {
		e.a.xpacked(true, 0xe6, d, s) // CVTTPD2DQ clears the upper two lanes.
	} else {
		e.a.xfloat(false, 0x5b, d, s) // CVTTPS2DQ
	}
}

// Convert a two-lane f64 comparison mask to two low i32 masks and zero padding.
func (e *functionEmitter) vectorMask32AMD64(r byte, wide bool) {
	if wide {
		e.a.xvector(0x70, r, r)
		e.a.bytes(0x88)
		e.a.xfloat(false, 0x7e, r, r) // MOVQ zeroes the upper half.
	}
}

func (e *functionEmitter) vectorSaturateAMD64(wide, unsigned bool, l byte) {
	e.vectorMove(11, l)
	if !unsigned {
		e.vectorMove(0, 11)
		e.a.xpacked(wide, 0xc2, 0, 11)
		e.a.bytes(7) // Replace unordered lanes with zero before conversion.
		e.a.xvector(0xdb, 11, 0)
		e.vectorFloatConstant(1, wide, 0x4f000000, 0x41e0000000000000)
		e.a.xpacked(wide, 0xc2, 1, 11)
		e.a.bytes(2) // Positive overflow produces INT_MIN, corrected to INT_MAX.
		e.vectorTruncateAMD64(wide, 2, 11)
		e.vectorMask32AMD64(1, wide)
		e.a.xvector(0xef, 2, 1)
		return
	}
	e.vectorZero(0)
	e.a.xpacked(wide, 0x5f, 11, 0) // MAX(input, +0) maps negatives and NaNs to zero.
	e.vectorFloatConstant(1, wide, 0x4f000000, 0x41e0000000000000)
	e.vectorMove(0, 1)
	e.a.xpacked(wide, 0xc2, 0, 11)
	e.a.bytes(2)
	e.a.xvector(0xdb, 1, 0)
	e.vectorMove(2, 11)
	e.a.xpacked(wide, 0x5c, 2, 1) // Subtract 2^31 only from the high unsigned half.
	e.vectorTruncateAMD64(wide, 2, 2)
	e.vectorMask32AMD64(0, wide)
	e.a.xvector(0x72, 6, 0)
	e.a.bytes(31)
	e.a.xvector(0xef, 2, 0)
	e.vectorFloatConstant(1, wide, 0x4f800000, 0x41f0000000000000)
	e.a.xpacked(wide, 0xc2, 1, 11)
	e.a.bytes(2)
	e.vectorMask32AMD64(1, wide)
	e.a.xvector(0xeb, 2, 1) // Overflow saturates to UINT_MAX, without touching padding.
}
