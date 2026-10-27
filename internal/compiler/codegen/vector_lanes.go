package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

func vectorLane(op uint32) (size uint32, replace, signed bool) {
	switch op {
	case 21, 22, 23:
		return 0, op == 23, op == 21
	case 24, 25, 26:
		return 1, op == 26, op == 24
	case 27, 28, 31, 32:
		return 2, op == 28 || op == 32, false
	default:
		return 3, op == 30 || op == 34, false
	}
}

func (e *functionEmitter) vectorExtract(src byte, size, lane uint32, signed bool) {
	if e.a.arch == "arm64" {
		ins := uint32(0x0e003c09)
		if signed {
			ins = 0x0e002c09
		}
		if size == 3 {
			ins |= 1 << 30
		}
		e.a.word(ins | ((1<<size)|(lane<<(size+1)))<<16 | uint32(src)<<5)
		return
	}
	e.vectorMove(11, src)
	if lane != 0 {
		e.a.xvector(0x73, 3, 11)
		e.a.bytes(byte(lane << size))
	}
	e.floatToInteger(0, 11, size == 3)
	if size < 2 {
		if signed {
			e.a.bytes(0x0f, []byte{0xbe, 0xbf}[size], 0xc0)
		} else {
			e.a.bytes(0x25)
			e.a.word((1 << (8 << size)) - 1)
		}
	}
}

func (e *functionEmitter) vectorScalarResult(dst slots.Slot) {
	e.slotStore(dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(dst, 0, true)
	}
}

func (e *functionEmitter) vectorSplat(in slots.Instruction) {
	size := []uint32{0, 1, 2, 3, 2, 3}[in.Op-0xfd000f]
	e.slotLoad(in.A, 0, false, size == 3)
	d := e.vectorDestination(in.Dst)
	if e.a.arch == "arm64" {
		e.a.word(0x4e000d20 | (1<<size)<<16 | uint32(d))
	} else {
		e.floatFromInteger(d, 0, size == 3)
		if size == 3 {
			e.a.xvector(0x6c, d, d)
		} else {
			if size == 0 {
				e.a.xvector(0x60, d, d)
			}
			if size <= 1 {
				e.a.xvector(0x61, d, d)
			}
			e.a.xvector(0x70, d, d)
			e.a.bytes(0)
		}
	}
	e.vectorStore(in.Dst, d)
}

func (e *functionEmitter) vectorLaneInstruction(in slots.Instruction) {
	size, replace, signed := vectorLane(in.Op & 0xffff)
	lane := uint32(e.f.Sources[in.Source].Wasm.Lane)
	l := e.vectorOperand(in.A, 0)
	if !replace {
		e.vectorExtract(l, size, lane, signed)
		e.vectorScalarResult(in.Dst)
		return
	}
	if e.a.arch == "arm64" {
		d := e.vectorDestination(in.Dst)
		e.slotLoad(in.B, 0, false, size == 3)
		e.vectorMove(d, l)
		e.a.word(0x4e001d20 | ((1<<size)|(lane<<(size+1)))<<16 | uint32(d))
		e.vectorStore(in.Dst, d)
		return
	}
	// SSE2 lacks a general lane insert. Mask and merge without touching any
	// allocated input register; MOVD/Q zeroes all unused replacement bits.
	e.vectorMove(0, l)
	e.slotLoad(in.B, 0, false, size == 3)
	if size < 2 {
		e.a.bytes(0x25)
		e.a.word((1 << (8 << size)) - 1)
	}
	e.floatFromInteger(2, 0, size == 3)
	if lane != 0 {
		e.a.xvector(0x73, 7, 2)
		e.a.bytes(byte(lane << size))
	}
	mask := ^uint64(0) >> (64 - (8 << size))
	shift := lane * (8 << size)
	lo, hi := ^uint64(0), ^uint64(0)
	if shift < 64 {
		lo &^= mask << shift
	} else {
		hi &^= mask << (shift - 64)
	}
	e.vectorConstant(12, lo, hi)
	e.a.xvector(0xdb, 0, 12)
	e.a.xvector(0xeb, 0, 2)
	e.vectorStore(in.Dst, 0)
}
