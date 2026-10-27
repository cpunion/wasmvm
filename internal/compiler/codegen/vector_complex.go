package codegen

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func nativeVectorComplexOpcode(op uint32) bool {
	if op>>16 != 0xfd {
		return false
	}
	op &= 0xffff
	if op == 13 || op == 14 || op == 98 || op == 101 || op == 102 || op == 123 || op >= 124 && op <= 127 || op == 130 || op == 133 || op == 134 || op == 155 || op == 186 {
		return true
	}
	if op < 128 || op > 223 {
		return false
	}
	kind := (op - 96) % 32
	return kind >= 7 && kind <= 10 || kind >= 28
}

func (e *functionEmitter) vectorComplexInstruction(in slots.Instruction) {
	l := e.vectorOperand(in.A, 0)
	r := byte(1)
	if in.B != slots.NoSlot {
		r = e.vectorOperand(in.B, 1)
	}
	if e.a.arch == "arm64" {
		d := e.vectorDestination(in.Dst)
		e.vectorComplexARM64(in, d, l, r)
		e.vectorStore(in.Dst, d)
		return
	}
	if in.Op == 0xfd0062 && e.x86Features&X86AVX512BITALG != 0 {
		d := e.vectorDestination(in.Dst)
		e.a.xevex(2, 1, 0x54, d, 0, l, false) // VPOPCNTB.
		e.vectorStore(in.Dst, d)
		return
	}
	if e.x86Features&X86AVX2 != 0 && e.vectorComplexAVX2(in, l, r) {
		return
	}
	// SSE2 legalizations use fixed scratch copies to preserve all cached slots.
	e.vectorMove(0, l)
	if in.B != slots.NoSlot {
		e.vectorMove(1, r)
	}
	e.vectorComplexAMD64(in)
	e.vectorStore(in.Dst, 2)
}

func (e *functionEmitter) vectorComplexARM64(in slots.Instruction, d, l, r byte) {
	op := in.Op & 0xffff
	ins, binaryOp := uint32(0), false
	switch op {
	case 13:
		e.vectorMove(3, l)
		e.vectorMove(4, r)
		b := e.f.Sources[in.Source].Wasm.Vector
		e.vectorConstant(2, binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:]))
		e.a.word(0x4e022060 | uint32(d)) // TBL {v3,v4}, v2.
		return
	case 14:
		ins, binaryOp = 0x4e000000, true // TBL; indices >=16 produce zero.
	case 98:
		ins = 0x4e205800 // CNT
	case 101, 102, 133, 134:
		size := (op - 101) / 32
		ins = 0x0e214800 | size<<22 // SQXTN/SQXTN2
		if op%2 == 0 {
			ins = 0x2e212800 | size<<22
		} // SQXTUN/SQXTUN2
		// Narrow through scratch when the final result aliases the second input.
		e.a.word(ins | uint32(l)<<5 | 2)
		e.a.word(ins | 1<<30 | uint32(r)<<5 | 2)
		e.vectorMove(d, 2)
		return
	case 123, 155:
		ins, binaryOp = 0x6e201400|((op-123)/32)<<22, true // URHADD
	case 124, 125, 126, 127:
		ins = 0x4e202800 | ((op-124)/2)<<22 | (op&1)<<29 // S/UADDLP
	case 130:
		ins, binaryOp = 0x6e60b400, true // SQRDMULH saturates MIN*MIN to MAX.
	case 186:
		e.a.word(0x0e60c000 | uint32(r)<<16 | uint32(l)<<5 | 3) // SMULL low
		e.a.word(0x4e60c000 | uint32(r)<<16 | uint32(l)<<5 | 4) // SMULL high
		e.a.word(0x4ea0bc00 | 4<<16 | 3<<5 | uint32(d))         // ADDP: modulo-i32 sums.
		return
	default:
		size, kind := (op-128)/32, (op-96)%32
		base := uint32(7)
		ins = 0x0f00a400 | (8<<size)<<16 // SSHLL/USHLL
		if kind >= 28 {
			base, ins, binaryOp = 28, 0x0e20c000|size<<22, true // SMULL/UMULL
		}
		ins |= ((kind-base)%2)<<30 | ((kind-base)/2)<<29
	}
	if binaryOp {
		ins |= uint32(r) << 16
	}
	e.a.word(ins | uint32(l)<<5 | uint32(d))
}

func (e *functionEmitter) vectorComplexAMD64(in slots.Instruction) {
	op := in.Op & 0xffff
	switch op {
	case 13:
		e.vectorShuffleAMD64(e.f.Sources[in.Source].Wasm.Vector)
	case 14:
		e.vectorSwizzleAMD64()
	case 98:
		e.vectorPopcountAMD64()
	case 101, 102, 133, 134:
		e.vectorNarrowAMD64(op)
	case 123, 155:
		e.vectorMove(2, 0)
		e.a.xvector(0xe0+byte((op-123)/32)*3, 2, 1) // PAVGB/PAVGW
	case 124, 125, 126, 127:
		e.vectorPairwiseAMD64(op)
	case 130:
		e.vectorQ15AMD64()
	case 186:
		e.vectorMove(2, 0)
		e.a.xvector(0xf5, 2, 1) // PMADDWD includes the modulo overflow case.
	default:
		size, kind := (op-96)/32, (op-96)%32
		if kind <= 10 {
			e.vectorExtendAMD64(2, 0, size, (kind-7)%2 != 0, kind >= 9)
		} else {
			e.vectorExtMultiplyAMD64(size, (kind-28)%2 != 0, kind >= 30)
		}
	}
}

func (e *functionEmitter) vectorShiftImmediateAMD64(reg byte, size, kind, count byte) {
	e.a.xvector(0x71+size, kind, reg)
	e.a.bytes(count)
}

// Unpack against a sign mask or zero. SSE2 has no packed widening instruction.
func (e *functionEmitter) vectorExtendAMD64(d, s byte, size uint32, high, unsigned bool) {
	if e.x86Features&X86SSE41 != 0 {
		if high {
			e.vectorMove(d, s)
			e.vectorShiftImmediateAMD64(d, 2, 3, 8)
			s = d
		}
		op := [...]byte{0x20, 0x23, 0x25}[size-1]
		if unsigned {
			op += 0x10
		}
		e.a.xvectorExtended(0x38, op, d, s)
		return
	}
	e.vectorZero(11)
	if !unsigned {
		e.a.xvector([]byte{0x64, 0x65, 0x66}[size-1], 11, s)
	}
	e.vectorMove(d, s)
	op := byte(0x60 + size - 1)
	if high {
		op += 8
	}
	e.a.xvector(op, d, 11)
}

func (e *functionEmitter) vectorExtMultiplyAMD64(size uint32, high, unsigned bool) {
	if size == 1 {
		e.vectorExtendAMD64(2, 0, size, high, unsigned)
		e.vectorExtendAMD64(12, 1, size, high, unsigned)
		e.a.xvector(0xd5, 2, 12)
		return
	}
	if size == 2 {
		e.vectorMove(11, 0)
		op := byte(0xe5) // PMULHW
		if unsigned {
			op = 0xe4
		}
		e.a.xvector(op, 11, 1)
		e.vectorMove(2, 0)
		e.a.xvector(0xd5, 2, 1)
		op = 0x61
		if high {
			op = 0x69
		}
		e.a.xvector(op, 2, 11)
		return
	}
	// PMULUDQ consumes even lanes. Duplicate the selected adjacent i32 pair.
	selectPair := byte(0x50)
	if high {
		selectPair = 0xfa
	}
	for _, reg := range []byte{0, 1} {
		e.a.xvector(0x70, reg, reg)
		e.a.bytes(selectPair)
	}
	if !unsigned {
		if e.x86Features&X86SSE41 != 0 {
			e.vectorMove(2, 0)
			e.a.xvectorExtended(0x38, 0x28, 2, 1) // PMULDQ
			return
		}
		// signed(a)*signed(b) = unsigned product minus each signed high half.
		e.vectorMove(11, 0)
		e.vectorMove(12, 1)
		e.vectorShiftImmediateAMD64(11, 1, 4, 31)
		e.vectorShiftImmediateAMD64(12, 1, 4, 31)
		e.a.xvector(0xdb, 11, 1)
		e.a.xvector(0xdb, 12, 0)
		e.vectorShiftImmediateAMD64(11, 2, 6, 32)
		e.vectorShiftImmediateAMD64(12, 2, 6, 32)
	}
	e.vectorMove(2, 0)
	e.a.xvector(0xf4, 2, 1)
	if !unsigned {
		e.a.xvector(0xfb, 2, 11)
		e.a.xvector(0xfb, 2, 12)
	}
}

func (e *functionEmitter) vectorNarrowAMD64(op uint32) {
	if op == 134 && e.x86Features&X86SSE41 != 0 {
		e.vectorMove(2, 0)
		e.a.xvectorExtended(0x38, 0x2b, 2, 1) // PACKUSDW
		return
	}
	if op != 134 {
		e.vectorMove(2, 0)
		e.a.xvector(map[uint32]byte{101: 0x63, 102: 0x67, 133: 0x6b}[op], 2, 1)
		return
	}
	// Clamp negative i32 inputs before subtracting 32768, then use PACKSSDW
	// and flip each i16 sign bit. The subtraction cannot underflow after clamp.
	e.vectorZero(11)
	e.vectorZero(12)
	e.a.xvector(0x66, 11, 0)
	e.a.xvector(0x66, 12, 1)
	e.a.xvector(0xdf, 11, 0)
	e.a.xvector(0xdf, 12, 1)
	e.vectorMove(0, 11)
	e.vectorMove(1, 12)
	e.vectorConstant(11, 0x0000800000008000, 0x0000800000008000)
	e.a.xvector(0xfa, 0, 11)
	e.a.xvector(0xfa, 1, 11)
	e.vectorMove(2, 0)
	e.a.xvector(0x6b, 2, 1)
	e.vectorConstant(11, 0x8000800080008000, 0x8000800080008000)
	e.a.xvector(0xef, 2, 11)
}

func (e *functionEmitter) vectorPairwiseAMD64(op uint32) {
	size := byte((op - 124) / 2) // Input lane width: byte or word.
	e.vectorMove(2, 0)
	shift := byte(8 << size)
	e.vectorShiftImmediateAMD64(2, size, 6, shift)
	kind := byte(4) // Arithmetic right shift sign-extends each adjacent lane.
	if op&1 != 0 {
		kind = 2
	}
	e.vectorShiftImmediateAMD64(2, size, kind, shift)
	e.vectorShiftImmediateAMD64(0, size, kind, shift)
	e.a.xvector(0xfd+size, 2, 0)
}

func (e *functionEmitter) vectorPopcountAMD64() {
	if e.x86Features&X86SSSE3 != 0 {
		e.vectorFloatConstant(12, true, 0, 0x0f0f0f0f0f0f0f0f)
		e.vectorMove(2, 0)
		e.a.xvector(0xdb, 2, 12)
		e.vectorShiftImmediateAMD64(0, 0, 2, 4)
		e.a.xvector(0xdb, 0, 12)
		e.vectorConstant(12, 0x0302020102010100, 0x0403030203020201)
		e.vectorMove(11, 12)
		e.a.xvectorExtended(0x38, 0x00, 11, 2)
		e.vectorMove(2, 12)
		e.a.xvectorExtended(0x38, 0x00, 2, 0)
		e.a.xvector(0xfc, 2, 11)
		return
	}
	e.vectorMove(2, 0)
	for _, step := range []struct {
		shift byte
		mask  uint64
	}{{1, 0x5555555555555555}, {2, 0x3333333333333333}, {4, 0x0f0f0f0f0f0f0f0f}} {
		e.vectorConstant(12, step.mask, step.mask)
		e.vectorMove(11, 2)
		e.vectorShiftImmediateAMD64(11, 0, 2, step.shift)
		e.a.xvector(0xdb, 11, 12)
		if step.shift == 1 {
			e.a.xvector(0xf8, 2, 11)
		} else {
			e.a.xvector(0xdb, 2, 12)
			e.a.xvector(0xfc, 2, 11)
		}
	}
}

func (e *functionEmitter) vectorQ15AMD64() {
	if e.x86Features&X86SSSE3 != 0 {
		e.vectorMove(2, 0)
		e.a.xvectorExtended(0x38, 0x0b, 2, 1) // PMULHRSW, then fix MIN*MIN.
		e.vectorFloatConstant(12, true, 0, 0x8000800080008000)
		e.a.xvector(0x75, 0, 12)
		e.a.xvector(0x75, 1, 12)
		e.a.xvector(0xdb, 0, 1)
		e.a.xvector(0xef, 2, 0)
		return
	}
	e.vectorMove(11, 0)
	e.vectorMove(2, 0)
	e.a.xvector(0xe5, 11, 1)
	e.a.xvector(0xd5, 2, 1)
	e.vectorShiftImmediateAMD64(2, 0, 2, 14)
	e.vectorFloatConstant(12, true, 0, 0x0001000100010001)
	e.a.xvector(0xfd, 2, 12)
	e.vectorShiftImmediateAMD64(2, 0, 2, 1)
	e.vectorShiftImmediateAMD64(11, 0, 6, 1)
	e.a.xvector(0xfd, 2, 11)
	e.vectorConstant(12, 0x8000800080008000, 0x8000800080008000)
	e.a.xvector(0x75, 0, 12)
	e.a.xvector(0x75, 1, 12)
	e.a.xvector(0xdb, 0, 1)
	e.a.xvector(0xef, 2, 0)
}

func (e *functionEmitter) vectorShuffleAMD64(lanes [16]byte) {
	// A single PSHUFD covers any aligned four-lane permutation of one input.
	source, control, aligned := vectorDwordShuffle(lanes)
	if aligned {
		e.a.xvector(0x70, 2, source)
		e.a.bytes(control)
		return
	}
	if e.x86Features&X86SSSE3 != 0 {
		e.vectorZero(2)
		for source := byte(0); source < 2; source++ {
			var mask [16]byte
			used := false
			for i, n := range lanes {
				mask[i] = 0x80
				if n/16 == source {
					mask[i], used = n%16, true
				}
			}
			if !used {
				continue
			}
			e.vectorConstant(12, binary.LittleEndian.Uint64(mask[:8]), binary.LittleEndian.Uint64(mask[8:]))
			e.vectorMove(11, source)
			e.a.xvectorExtended(0x38, 0x00, 11, 12)
			e.a.xvector(0xeb, 2, 11)
		}
		return
	}
	// Group equal byte shifts. Each mask selects only bytes from that source;
	// the ordered finite scan keeps code bytes deterministic without a map.
	e.vectorZero(2)
	for source := byte(0); source < 2; source++ {
		for shift := -15; shift <= 15; shift++ {
			var mask [16]byte
			used := false
			for i, n := range lanes {
				if n/16 == source && int(n%16)-i == shift {
					mask[i], used = 255, true
				}
			}
			if !used {
				continue
			}
			e.vectorConstant(12, binary.LittleEndian.Uint64(mask[:8]), binary.LittleEndian.Uint64(mask[8:]))
			e.vectorMove(11, source)
			if shift != 0 {
				kind, count := byte(3), shift
				if shift < 0 {
					kind, count = 7, -shift
				}
				e.vectorShiftImmediateAMD64(11, 2, kind, byte(count))
			}
			e.a.xvector(0xdb, 11, 12)
			e.a.xvector(0xeb, 2, 11)
		}
	}
}

func (e *functionEmitter) vectorSwizzleAMD64() {
	if e.x86Features&X86SSSE3 != 0 {
		// Saturating +0x70 sets bit 7 for every invalid Wasm index, including
		// 16..127 which raw PSHUFB would incorrectly wrap into its table.
		e.vectorFloatConstant(12, true, 0, 0x7070707070707070)
		e.vectorMove(11, 1)
		e.a.xvector(0xdc, 11, 12)
		e.vectorMove(2, 0)
		e.a.xvectorExtended(0x38, 0x00, 2, 11)
		return
	}
	// SSE2 has no byte table lookup. Compare every table index in parallel;
	// all 16 output lanes participate and out-of-range indices never match.
	e.vectorZero(2)
	for i := uint32(0); i < 16; i++ {
		bits := uint64(i) * 0x0101010101010101
		e.vectorConstant(12, bits, bits)
		e.vectorExtract(0, 0, i, false)
		e.vectorSplatBits(11, 0)
		e.a.xvector(0x74, 12, 1)
		e.a.xvector(0xdb, 11, 12)
		e.a.xvector(0xeb, 2, 11)
	}
}
