package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Variable counts are masked once, before broadcasting to every packed lane.
func (e *functionEmitter) vectorShift(in slots.Instruction, l byte, size, kind uint32) {
	e.slotLoad(in.B, 0, false, false)
	if e.a.arch == "arm64" {
		e.a.word(0x12000129 | (size+2)<<10) // AND w9, w9, laneBits-1.
		if kind != 11 {
			e.a.word(0x4b0903e9) // NEG w9; SSHL/USHL use signed low-byte counts.
		}
		e.vectorSplatBits(1, size)
		ins := uint32(0x6e214400) // USHL also implements signed/unsigned left shift.
		if kind == 12 {
			ins = 0x4e214400 // SSHL
		}
		e.a.word(ins | size<<22 | uint32(l)<<5 | 2)
	} else {
		e.a.bytes(0x83, 0xe0, byte((8<<size)-1))
		e.floatFromInteger(1, 0, true)
		if size == 0 {
			e.vectorByteShiftAMD64(l, kind)
		} else if size == 3 && kind == 12 {
			// Arithmetic qword shift: ((x XOR sign) >> n) XOR sign. SSE2's
			// dword arithmetic shift duplicates each qword's high sign bit.
			e.a.xvector(0x70, 11, l)
			e.a.bytes(0xf5)
			e.vectorShiftImmediateAMD64(11, 1, 4, 31)
			e.vectorMove(2, l)
			e.a.xvector(0xef, 2, 11)
			e.a.xvector(0xd3, 2, 1)
			e.a.xvector(0xef, 2, 11)
		} else {
			op := byte(0xd0 + size) // PSRLW/D/Q
			if kind == 11 {
				op = 0xf0 + byte(size)
			} else if kind == 12 {
				op = 0xe0 + byte(size)
			}
			e.vectorMove(2, l)
			e.a.xvector(op, 2, 1)
		}
	}
	e.vectorStore(in.Dst, 2)
}

func (e *functionEmitter) vectorByteShiftAMD64(l byte, kind uint32) {
	if kind == 12 {
		e.vectorZero(11)
		e.a.xvector(0x64, 11, l)
		e.vectorMove(2, l)
		e.vectorMove(12, l)
		e.a.xvector(0x60, 2, 11)
		e.a.xvector(0x68, 12, 11)
		e.a.xvector(0xe1, 2, 1)
		e.a.xvector(0xe1, 12, 1)
		e.a.xvector(0x63, 2, 12)
		return
	}
	// Word shifts can leak bits across bytes. A repeated shifted byte mask
	// removes exactly that contamination without splitting individual lanes.
	e.a.bytes(0x89, 0xc1, 0xb8)
	e.a.word(255)
	op, shift := byte(0xd1), byte(0xe8)
	if kind == 11 {
		op, shift = 0xf1, 0xe0
	}
	e.a.bytes(0xd3, shift, 0x25)
	e.a.word(255)
	e.vectorSplatBits(12, 0)
	e.vectorMove(2, l)
	e.a.xvector(op, 2, 1)
	e.a.xvector(0xdb, 2, 12)
}

func (e *functionEmitter) vectorReduce(in slots.Instruction, l byte, size, kind uint32) {
	if e.a.arch == "arm64" {
		if kind == 4 {
			e.vectorBitmaskARM64(l, size)
		} else {
			if kind == 3 {
				e.a.word(0x4e209800 | size<<22 | uint32(l)<<5 | 2) // CMEQ lanes,#0.
				l = 2
			}
			e.a.word(0x6e30a800 | uint32(l)<<5 | 2) // UMAXV bytes detects any set bit.
			e.a.word(0x0e013c49)                    // UMOV w9,v2.b[0].
			e.test(false)
			condition := uint32(0x1a9f07e9) // CSET w9,NE (any_true).
			if kind == 3 {
				condition = 0x1a9f17e9 // CSET w9,EQ (no zero lane).
			}
			e.a.word(condition)
		}
	} else if kind == 4 {
		if size >= 2 {
			e.a.xpacked(size == 3, 0x50, 0, l) // MOVMSKPS/PD
		} else {
			if size == 1 {
				e.vectorMove(2, l)
				e.a.xvector(0x63, 2, 2) // PACKSSWB preserves every signed i16 sign.
				l = 2
			}
			e.a.xvector(0xd7, 0, l) // PMOVMSKB
			if size == 1 {
				e.a.bytes(0x25)
				e.a.word(255)
			}
		}
	} else {
		e.vectorZero(11)
		e.vectorMove(2, l)
		op := byte(0x74)
		if kind == 3 {
			op += byte(min(size, 2))
		}
		e.a.xvector(op, 2, 11)
		if kind == 3 && size == 3 {
			e.a.xvector(0x70, 11, 2)
			e.a.bytes(0xb1)
			e.a.xvector(0xdb, 2, 11)
		}
		e.a.xvector(0xd7, 0, 2)
		if kind == 3 {
			e.a.bytes(0x85, 0xc0, 0x0f, 0x94, 0xc0) // No zero lane.
		} else {
			e.a.bytes(0x3d)
			e.a.word(65535)
			e.a.bytes(0x0f, 0x95, 0xc0) // At least one nonzero byte.
		}
		e.a.bytes(0x0f, 0xb6, 0xc0)
	}
	e.vectorScalarResult(in.Dst)
}

func (e *functionEmitter) vectorBitmaskARM64(l byte, size uint32) {
	if size == 3 {
		e.vectorExtract(l, 3, 0, false)
		e.a.word(0xd37ffd2a) // LSR x10,x9,#63.
		e.vectorExtract(l, 3, 1, false)
		e.a.word(0xd37ffd29) // LSR x9,x9,#63.
		e.a.word(0x2a090549) // ORR w9,w10,w9,LSL #1.
		return
	}
	bits := vectorBitWeights(size)
	weight, ok := e.vectorConstants[bits]
	if !ok {
		e.vectorConstant(1, bits[0], bits[1])
		weight = 1
	}
	e.a.word(0x6f000400 | ((8<<size)+1)<<16 | uint32(l)<<5 | 2) // USHR,#laneBits-1.
	e.a.word(0x4e209c42 | size<<22 | uint32(weight)<<16)        // MUL by bit positions.
	if size == 0 {
		e.a.word(0x0e31b843) // ADDV b3,low eight bytes.
		e.a.word(0x6e024042) // EXT high eight bytes into low half.
	}
	ins := uint32(0x0e31b842) // ADDV low eight bytes.
	if size > 0 {
		ins |= 1<<30 | size<<22 // ADDV all eight halfwords/four words.
	}
	e.a.word(ins)
	e.a.word(0x0e003c49 | (1<<size)<<16) // UMOV w9,scalar lane.
	if size == 0 {
		e.a.word(0x2a0903ea) // MOV w10,w9 (high byte).
		e.a.word(0x0e013c69) // UMOV w9,v3.b[0] (low byte).
		e.a.word(0x2a0a2129) // ORR w9,w9,w10,LSL #8.
	}
}

func vectorBitWeights(size uint32) [2]uint64 {
	return [...][2]uint64{
		{0x8040201008040201, 0x8040201008040201},
		{0x0008000400020001, 0x0080004000200010},
		{0x0000000200000001, 0x0000000800000004},
	}[size]
}
