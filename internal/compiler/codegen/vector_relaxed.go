package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

func nativeRelaxedVectorOpcode(op uint32) bool {
	return op >= 0xfd0100 && op <= 0xfd0113
}

// Preserve the interpreter's fixed permitted projection. Strict operations
// reuse existing legalization, allocation and cold-continuation contracts.
func (e *functionEmitter) relaxedVectorInstruction(in slots.Instruction) {
	op := in.Op & 0xffff
	var fixed uint32
	switch {
	case op == 0x100:
		fixed = 14
	case op >= 0x101 && op <= 0x104:
		fixed = [...]uint32{248, 249, 252, 253}[op-0x101]
	case op >= 0x109 && op <= 0x10c:
		fixed = 82
	case op >= 0x10d && op <= 0x110:
		fixed = [...]uint32{232, 233, 244, 245}[op-0x10d]
	case op == 0x111:
		fixed = 130
	}
	if fixed != 0 {
		in.Op = 0xfd0000 | fixed
		e.vectorInstruction(in)
		return
	}
	l, r := e.vectorOperand(in.A, 0), e.vectorOperand(in.B, 1)
	if op >= 0x105 && op <= 0x108 {
		wide := op >= 0x107
		if e.a.arch == "arm64" {
			e.vectorFloatARM64(0x94, wide, 2, l, r)
			if op&1 == 0 {
				e.vectorFloatARM64(0x8c, wide, 2, 2, 0)
			}
			c := e.vectorOperand(in.C, 1)
			e.vectorFloatARM64(0x92, wide, 2, 2, c)
		} else {
			e.vectorMove(2, l)
			e.a.xpacked(wide, 0x59, 2, r) // MULPS/PD rounds before ADDPS/PD.
			if op&1 == 0 {
				e.vectorFloatMaskAMD64(11, wide, true)
				e.a.xvector(0xef, 2, 11)
			}
			c := e.vectorOperand(in.C, 1)
			e.a.xpacked(wide, 0x58, 2, c)
		}
	} else {
		if e.a.arch == "arm64" {
			e.a.word(0x0e20c000 | uint32(r)<<16 | uint32(l)<<5 | 3) // SMULL bytes.
			e.a.word(0x4e20c000 | uint32(r)<<16 | uint32(l)<<5 | 4) // SMULL2 bytes.
			if op == 0x112 {
				e.a.word(0x4e60bc00 | 4<<16 | 3<<5 | 2) // ADDP halfwords, modulo 16 bits.
			} else {
				e.a.word(0x4e602800 | 3<<5 | 3) // SADDLP widens before pair sums.
				e.a.word(0x4e602800 | 4<<5 | 4)
				e.a.word(0x4ea0bc00 | 4<<16 | 3<<5 | 2)
			}
		} else {
			e.relaxedDotAMD64(l, r, op == 0x113)
		}
		if op == 0x113 {
			c := e.vectorOperand(in.C, 1)
			e.vectorBinary(2, 2, c, 2, 14)
		}
	}
	e.vectorStore(in.Dst, 2)
}

// SSE2 needs two signed widening multiplies; no SSSE3/VNNI is assumed.
// Intermediate products stay exact, including -128 * -128. Only the final
// halfword result wraps; the four-byte dot widens before accumulation.
func (e *functionEmitter) relaxedDotAMD64(l, r byte, add bool) {
	e.vectorMove(0, l)
	e.vectorMove(1, r)
	e.vectorMove(12, 0)
	e.vectorExtendAMD64(0, 0, 1, true, false)
	e.vectorExtendAMD64(2, 1, 1, true, false)
	e.a.xvector(0xd5, 0, 2) // High PMULLW.
	e.vectorExtendAMD64(2, 12, 1, false, false)
	e.vectorExtendAMD64(1, 1, 1, false, false)
	e.a.xvector(0xd5, 2, 1) // Low PMULLW.
	if add {
		e.vectorConstant(11, 0x0001000100010001, 0x0001000100010001)
		for _, reg := range []byte{2, 0} {
			e.a.xvector(0xf5, reg, 11) // PMADDWD: exact adjacent products into i32.
			e.vectorMove(1, reg)
			e.vectorShiftImmediateAMD64(1, 2, 3, 4)
			e.a.xvector(0xfe, reg, 1)
			e.a.xvector(0x70, reg, reg)
			e.a.bytes(0x88)
		}
		e.a.xvector(0x6c, 2, 0)
	} else {
		for _, reg := range []byte{2, 0} {
			e.vectorMove(1, reg)
			e.vectorShiftImmediateAMD64(1, 1, 2, 16)
			e.a.xvector(0xfd, reg, 1)
			e.vectorShiftImmediateAMD64(reg, 1, 6, 16)
			e.vectorShiftImmediateAMD64(reg, 1, 4, 16)
		}
		e.a.xvector(0x6b, 2, 0) // PACKSSDW of already sign-extended pair sums.
	}
}
