package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Baseline SSE2 and AArch64 NEON share this selection contract. Every other
// SIMD opcode retains its explicit typed helper; no CPU extension is assumed.
func nativeVectorOpcode(op uint32) bool {
	if nativeVectorFloatOpcode(op) || nativeVectorComplexOpcode(op) {
		return true
	}
	if op>>16 != 0xfd {
		return false
	}
	op &= 0xffff
	if op >= 15 && op <= 64 || op >= 77 && op <= 83 || op >= 214 && op <= 219 {
		return true
	}
	if op < 96 || op > 213 {
		return false
	}
	size, kind := (op-96)/32, (op-96)%32
	switch kind {
	case 0, 1, 3, 4, 11, 12, 13, 14, 17:
		return true
	case 15, 16, 18, 19:
		return size < 2
	case 21:
		return size != 0
	case 22, 23, 24, 25:
		return size < 3
	}
	return false
}

func (e *functionEmitter) vectorInstruction(in slots.Instruction) {
	if nativeVectorComplexOpcode(in.Op) {
		e.vectorComplexInstruction(in)
		return
	}
	if nativeVectorFloatOpcode(in.Op) {
		e.vectorFloatInstruction(in)
		return
	}
	op := in.Op & 0xffff
	if op <= 20 {
		e.vectorSplat(in)
		return
	}
	if op <= 34 {
		e.vectorLaneInstruction(in)
		return
	}
	l := e.vectorOperand(in.A, 0)
	d := e.vectorDestination(in.Dst)
	if op == 83 {
		e.vectorReduce(in, l, 3, 0)
		return
	}
	if op >= 96 && op < 214 {
		size, kind := (op-96)/32, (op-96)%32
		if kind == 3 || kind == 4 {
			e.vectorReduce(in, l, size, kind)
			return
		}
		if kind >= 11 && kind <= 13 {
			e.vectorShift(in, l, size, kind)
			return
		}
		if kind == 0 || kind == 1 {
			e.vectorUnary(d, l, size, kind == 0)
			e.vectorStore(in.Dst, d)
			return
		}
	}
	if op == 77 {
		e.vectorNot(d, l)
		e.vectorStore(in.Dst, d)
		return
	}
	r := e.vectorOperand(in.B, 1)
	if op >= 35 && op <= 64 || op >= 214 && op <= 219 {
		size, kind := (op-35)/10, (op-35)%10
		if op >= 214 {
			size, kind = 3, []uint32{0, 1, 2, 4, 6, 8}[op-214]
		}
		e.vectorCompare(d, l, r, size, kind)
	} else if op == 82 {
		c := e.vectorOperand(in.C, 2)
		e.vectorSelectTo(d, l, r, c)
	} else if op >= 78 && op <= 81 {
		e.vectorBinary(d, l, r, 0, op)
	} else {
		size, kind := (op-96)/32, (op-96)%32
		if kind >= 22 && kind <= 25 {
			// Select the right operand only when it is strictly better.
			left, right := l, r
			if kind >= 24 {
				left, right = r, l
			}
			e.vectorCompare(2, left, right, size, 4+kind%2)
			e.vectorSelectTo(d, r, l, 2)
		} else if size == 3 && kind == 21 {
			e.vectorMultiply64(l, r)
			d = 2
		} else {
			e.vectorBinary(d, l, r, size, kind)
		}
	}
	e.vectorStore(in.Dst, d)
}

func (e *functionEmitter) vectorNot(d, l byte) {
	if e.a.arch == "arm64" {
		e.a.word(0x6e205800 | uint32(l)<<5 | uint32(d))
		return
	}
	e.vectorMove(d, l)
	e.a.xvector(0x76, 11, 11)
	e.a.xvector(0xef, d, 11)
}

func (e *functionEmitter) vectorBinary(d, l, r byte, size, kind uint32) {
	if e.a.arch == "arm64" {
		ins := map[uint32]uint32{14: 0x4e208400, 17: 0x6e208400, 21: 0x4e209c00, 15: 0x4e200c00, 16: 0x6e200c00, 18: 0x4e202c00, 19: 0x6e202c00,
			78: 0x4e201c00, 79: 0x4e601c00, 80: 0x4ea01c00, 81: 0x6e201c00}[kind]
		e.a.word(ins | size<<22 | uint32(r)<<16 | uint32(l)<<5 | uint32(d))
		return
	}
	if d == r && d != l {
		e.vectorMove(1, r)
		r = 1
	}
	if kind == 79 {
		e.vectorMove(11, r)
		e.a.xvector(0xdf, 11, l)
		e.vectorMove(d, 11)
		return
	}
	if kind == 21 && size == 2 {
		if e.x86Features&X86SSE41 != 0 {
			e.vectorMove(d, l)
			e.a.xvectorExtended(0x38, 0x40, d, r) // PMULLD
			return
		}
		// SSE2 multiplies only even u32 lanes into u64 products. Select each
		// low half and interleave, preserving modulo-2^32 signed/unsigned bits.
		e.vectorMove(11, l)
		e.vectorMove(12, r)
		e.a.xvector(0x73, 3, 11)
		e.a.bytes(4)
		e.a.xvector(0x73, 3, 12)
		e.a.bytes(4)
		e.a.xvector(0xf4, 11, 12)
		e.vectorMove(d, l)
		e.a.xvector(0xf4, d, r)
		e.a.xvector(0x70, 11, 11)
		e.a.bytes(0x88)
		e.a.xvector(0x70, d, d)
		e.a.bytes(0x88)
		e.a.xvector(0x62, d, 11)
		return
	}
	e.vectorMove(d, l)
	var opcode byte
	switch kind {
	case 14:
		opcode = []byte{0xfc, 0xfd, 0xfe, 0xd4}[size]
	case 17:
		opcode = []byte{0xf8, 0xf9, 0xfa, 0xfb}[size]
	case 21:
		opcode = 0xd5
	case 15:
		opcode = 0xec + byte(size)
	case 16:
		opcode = 0xdc + byte(size)
	case 18:
		opcode = 0xe8 + byte(size)
	case 19:
		opcode = 0xd8 + byte(size)
	case 78:
		opcode = 0xdb
	case 80:
		opcode = 0xeb
	case 81:
		opcode = 0xef
	}
	e.a.xvector(opcode, d, r)
}

func (e *functionEmitter) vectorUnary(d, l byte, size uint32, absolute bool) {
	if e.a.arch == "arm64" {
		ins := uint32(0x6e20b800)
		if absolute {
			ins = 0x4e20b800
		}
		e.a.word(ins | size<<22 | uint32(l)<<5 | uint32(d))
		return
	}
	if !absolute {
		e.a.xvector(0xef, 11, 11)
		e.a.xvector([]byte{0xf8, 0xf9, 0xfa, 0xfb}[size], 11, l)
		e.vectorMove(d, 11)
		return
	}
	if size == 0 {
		e.a.xvector(0xef, 11, 11)
		e.a.xvector(0x64, 11, l)
	} else {
		e.vectorMove(11, l)
		if size == 3 {
			e.a.xvector(0x70, 11, 11)
			e.a.bytes(0xf5)
		}
		shift, count := byte(0x72), byte(31)
		if size == 1 {
			shift, count = 0x71, 15
		}
		e.a.xvector(shift, 4, 11)
		e.a.bytes(count)
	}
	e.vectorMove(d, l)
	e.a.xvector(0xef, d, 11)
	e.a.xvector([]byte{0xf8, 0xf9, 0xfa, 0xfb}[size], d, 11)
}

// Select into scratch 2 without overwriting an aliased operand or mask.
func (e *functionEmitter) vectorSelect(l, r, mask byte) {
	e.vectorSelectTo(2, l, r, mask)
}

func (e *functionEmitter) vectorSelectTo(d, l, r, mask byte) {
	if e.a.arch == "arm64" {
		// BSL, BIT and BIF let any of the three operands occupy the result.
		// Read all inputs in place, without inserting copies into the chain.
		ins := uint32(0x6e601c00)
		if mask != d && l == d {
			ins, l, r = 0x6ee01c00, r, mask // BIF: keep l where mask is set.
		} else if mask != d && r == d {
			ins, r = 0x6ea01c00, mask // BIT: replace r where mask is set.
		} else {
			e.vectorMove(d, mask)
		}
		e.a.word(ins | uint32(r)<<16 | uint32(l)<<5 | uint32(d))
		return
	}
	e.vectorMove(11, l)
	e.a.xvector(0xef, 11, r)
	e.a.xvector(0xdb, 11, mask)
	e.vectorMove(2, r)
	e.a.xvector(0xef, 2, 11)
	e.vectorMove(d, 2)
}

func (e *functionEmitter) vectorCompare(d, l, r byte, size, kind uint32) {
	if e.a.arch == "amd64" && size == 3 && kind > 1 {
		e.vectorCompare64(d, l, r, kind)
		return
	}
	invert := kind == 1 || kind >= 6
	if kind == 2 || kind == 3 || kind >= 8 {
		l, r = r, l
	}
	if e.a.arch == "arm64" {
		ins := uint32(0x6e208c00)
		if kind > 1 {
			ins = 0x4e203400
			if kind%2 == 1 {
				ins = 0x6e203400
			}
		}
		e.a.word(ins | size<<22 | uint32(r)<<16 | uint32(l)<<5 | uint32(d))
	} else if kind > 1 && kind%2 == 1 {
		// Flip each sign bit to map unsigned ordering to signed ordering.
		mask := uint64(0)
		for i := uint32(0); i < 64; i += 8 << size {
			mask |= 1 << ((8 << size) - 1 + i)
		}
		e.vectorConstant(11, mask, mask)
		e.vectorMove(12, r)
		e.a.xvector(0xef, 12, 11)
		e.vectorMove(d, l)
		e.a.xvector(0xef, d, 11)
		e.a.xvector([]byte{0x64, 0x65, 0x66}[size], d, 12)
	} else {
		if d == r && d != l {
			e.vectorMove(1, r)
			r = 1
		}
		e.vectorMove(d, l)
		if kind <= 1 {
			e.a.xvector([]byte{0x74, 0x75, 0x76, 0x76}[size], d, r)
			if size == 3 {
				e.vectorMove(11, d)
				e.a.xvector(0x70, 11, 11)
				e.a.bytes(0xb1)
				e.a.xvector(0xdb, d, 11)
			}
		} else {
			e.a.xvector([]byte{0x64, 0x65, 0x66}[size], d, r)
		}
	}
	if invert {
		e.vectorNot(d, d)
	}
}

func (e *functionEmitter) vectorScalarCore(op uint32) {
	if e.a.arch == "arm64" {
		integerARM64Core(e.a, op, integerWide(op), e.zero, e.overflow)
	} else {
		integerAMD64Core(e.a, op, integerWide(op), e.zero, e.overflow)
	}
}

func (e *functionEmitter) vectorInsertResult(size, lane uint32) {
	if e.a.arch == "arm64" {
		e.a.word(0x4e001d22 | ((1<<size)|(lane<<(size+1)))<<16)
		return
	}
	if size < 2 {
		e.a.bytes(0x25)
		e.a.word((1 << (8 << size)) - 1)
	}
	e.floatFromInteger(12, 0, size == 3)
	if lane != 0 {
		e.a.xvector(0x73, 7, 12)
		e.a.bytes(byte(lane << size))
	}
	e.a.xvector(0xeb, 2, 12)
}
func (e *functionEmitter) vectorZero(reg byte) {
	if e.a.arch == "arm64" {
		e.a.word(0x6e201c00 | uint32(reg)<<16 | uint32(reg)<<5 | uint32(reg))
	} else {
		e.a.xvector(0xef, reg, reg)
	}
}

func (e *functionEmitter) vectorMultiply64(l, r byte) {
	e.vectorZero(2)
	for lane := uint32(0); lane < 2; lane++ {
		e.vectorExtract(l, 3, lane, false)
		e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
		e.vectorExtract(r, 3, lane, false)
		e.registerMove(e.operandRegister(1), e.operandRegister(2), true)
		e.vectorScalarCore(0x7e)
		e.vectorInsertResult(3, lane)
	}
}
func (e *functionEmitter) vectorCompare64(d, l, r byte, kind uint32) {
	e.vectorZero(2)
	for lane := uint32(0); lane < 2; lane++ {
		e.vectorExtract(r, 3, lane, false)
		e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
		e.vectorExtract(l, 3, lane, false)
		e.registerMove(e.operandRegister(1), e.operandRegister(2), true)
		e.vectorScalarCore(0x51 + kind)
		e.a.bytes(0x48, 0xf7, 0xd8) // neg rax: expand the predicate to all 64 bits.
		e.vectorInsertResult(3, lane)
	}
	e.vectorMove(d, 2)
}

func (e *functionEmitter) vectorShift(in slots.Instruction, l byte, size, kind uint32) {
	e.vectorZero(2)
	for lane := uint32(0); lane < (16 >> size); lane++ {
		e.vectorExtract(l, size, lane, kind == 12 && size < 2)
		e.slotLoad(in.B, 1, false, false)
		if e.a.arch == "arm64" {
			e.a.word(0x1200014a | (size+2)<<10)
		} else {
			e.a.bytes(0x83, 0xe1, byte((8<<size)-1))
		}
		op := uint32(0x74) + kind - 11
		if size == 3 {
			op += 0x12
		}
		e.vectorScalarCore(op)
		e.vectorInsertResult(size, lane)
	}
	e.vectorStore(in.Dst, 2)
}

func (e *functionEmitter) vectorReduce(in slots.Instruction, l byte, size, kind uint32) {
	initial := uint64(0)
	if kind == 3 {
		initial = 1
	}
	e.constant(2, initial)
	for lane := uint32(0); lane < (16 >> size); lane++ {
		e.vectorExtract(l, size, lane, false)
		if kind == 4 {
			e.constant(1, uint64((8<<size)-1))
			op := uint32(0x76)
			if size == 3 {
				op = 0x88
			}
			e.vectorScalarCore(op)
			e.constant(1, uint64(lane))
			e.vectorScalarCore(0x74)
		} else {
			e.constant(1, 0)
			op := uint32(0x47)
			if size == 3 {
				op = 0x52
			}
			e.vectorScalarCore(op)
		}
		e.registerMove(e.operandRegister(1), e.operandRegister(2), false)
		op := uint32(0x72)
		if kind == 3 {
			op = 0x71
		}
		e.vectorScalarCore(op)
		e.registerMove(e.operandRegister(2), e.operandRegister(0), false)
	}
	e.vectorScalarResult(in.Dst)
}
