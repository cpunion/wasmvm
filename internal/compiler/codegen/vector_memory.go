package codegen

import (
	"math/bits"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func nativeVectorMemoryOpcode(op uint32) bool {
	return op >= 0xfd0000 && op <= 0xfd000b || op >= 0xfd0054 && op <= 0xfd005d
}

func vectorDirectMemory(op uint32) bool {
	return op <= 0xfd0006 || op == 0xfd000b || op >= 0xfd005c
}

// The common memory emitter has already checked the entire access range. Lane
// instructions check only their lane width; no vector load may read past it.
func (e *functionEmitter) vectorMemoryARM64(in slots.Instruction, load bool, size uint32, index byte, scaledOffset uint32) {
	op, lane := in.Op&0xffff, uint32(e.f.Sources[in.Source].Wasm.Lane)
	d := e.vectorDestination(in.Dst)
	if !load {
		d = e.vectorOperand(in.B, 0)
	} else if op >= 84 && op <= 87 {
		e.vectorMove(d, e.vectorOperand(in.B, 0))
	}
	var ins uint32
	switch {
	case op == 0 || op == 11:
		ins = 0x3d800000
		if load {
			ins |= 1 << 22
		}
	case op >= 1 && op <= 6 || op == 93:
		ins = 0xfd400000 // LDR D zeroes the other 64 bits.
	case op == 92:
		ins = 0xbd400000 // LDR S zeroes the other 96 bits.
	case op >= 7 && op <= 10:
		ins = 0x4d40c000 | (op-7)<<10 // LD1R
	default:
		// LD1/ST1 lane: Q selects the high half; S:size indexes the low half.
		index := lane * size
		ins = 0x0d000000 | (index>>3)<<30 | (index&7)<<10
		ins |= [...]uint32{0, 0x4000, 0x8000, 0x8400}[bits.TrailingZeros32(size)]
		if load {
			ins |= 1 << 22
		}
	}
	if vectorDirectMemory(in.Op) {
		e.memoryAddressARM64(ins, d, index, scaledOffset)
	} else {
		e.a.word(ins | 9<<5 | uint32(d)) // LD1/ST1 require a complete address.
	}
	if op >= 1 && op <= 6 {
		ins = 0x0f00a400 | uint32(8<<((op-1)/2))<<16 // SSHLL from the low half.
		if op%2 == 0 {
			ins |= 1 << 29 // USHLL
		}
		e.a.word(ins | uint32(d)<<5 | uint32(d))
	}
	if load {
		e.vectorStore(in.Dst, d)
	}
}

// All x86 memory forms use [r10+rax], after widened address/bounds arithmetic.
func (e *functionEmitter) vectorMemoryAMD64(in slots.Instruction, load bool, size uint32) {
	op, lane := in.Op&0xffff, uint32(e.f.Sources[in.Source].Wasm.Lane)
	d := e.vectorDestination(in.Dst)
	if !load {
		d = e.vectorOperand(in.B, 0)
	}
	if op == 0 || op == 11 {
		code := byte(0x6f) // MOVDQU
		if !load {
			code = 0x7f
		}
		e.vectorAccessAMD64(0xf3, code, d)
	} else if op >= 1 && op <= 6 {
		e.vectorAccessAMD64(0xf3, 0x7e, d) // MOVQ loads exactly eight bytes.
		e.vectorZero(11)
		width := (op - 1) / 2
		if op%2 == 1 {
			e.a.xvector(0x64+byte(width), 11, d) // Sign mask from zero > input.
		}
		e.a.xvector(0x60+byte(width), d, 11) // Widen each low-half lane.
	} else if op == 92 || op == 93 {
		prefix, code := byte(0x66), byte(0x6e) // MOVD, with zeroed upper bits.
		if op == 93 {
			prefix, code = 0xf3, 0x7e
		}
		e.vectorAccessAMD64(prefix, code, d)
	} else if op >= 88 && op <= 91 {
		e.vectorMove(11, d)
		if lane != 0 {
			e.a.xvector(0x73, 3, 11)
			e.a.bytes(byte(lane * size))
		}
		e.a.xfloatMove(size == 8, false, 11, 1) // Preserve address RAX; value in RCX.
		if size == 2 {
			e.a.bytes(0x66)
		}
		rex, code := byte(0x41), byte(0x89)
		if size == 8 {
			rex |= 8
		} else if size == 1 {
			code = 0x88
		}
		e.a.bytes(rex, code, 0x0c, 0x02)
	} else {
		// Splat and lane loads read exactly their scalar width into RCX.
		if size <= 2 {
			e.a.bytes(0x41, 0x0f, 0xb6+byte(size-1), 0x0c, 0x02)
		} else {
			rex := byte(0x41)
			if size == 8 {
				rex |= 8
			}
			e.a.bytes(rex, 0x8b, 0x0c, 0x02)
		}
		e.registerMove(0, 1, size == 8) // Address no longer needed after the load.
		width := uint32(bits.TrailingZeros32(size))
		if op >= 7 && op <= 10 {
			e.vectorSplatBits(d, width)
		} else {
			e.vectorInsertAMD64(in.Dst, e.vectorOperand(in.B, 0), width, lane)
			return
		}
	}
	if load {
		e.vectorStore(in.Dst, d)
	}
}

func (e *functionEmitter) vectorAccessAMD64(prefix, opcode, reg byte) {
	e.a.bytes(prefix, 0x41|(reg>>3)<<2, 0x0f, opcode, (reg&7)<<3|4, 2)
}
