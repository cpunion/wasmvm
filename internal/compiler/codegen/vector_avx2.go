package codegen

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func (e *functionEmitter) vectorConstantOperand(bits [2]uint64, scratch byte) byte {
	if r, ok := e.vectorConstants[bits]; ok {
		return r
	}
	e.vectorConstant(scratch, bits[0], bits[1])
	return scratch
}

func (e *functionEmitter) vectorComplexAVX2(in slots.Instruction, l, r byte) bool {
	op := in.Op & 0xffff
	if op >= 124 && op <= 127 {
		return false // Retain the packed pairwise legalization.
	}
	d := e.vectorDestination(in.Dst)
	switch op {
	case 13:
		e.vectorShuffleAVX2(e.f.Sources[in.Source].Wasm.Vector, l, r)
		d = 2
	case 14:
		mask := e.vectorConstantOperand([2]uint64{0x7070707070707070, 0x7070707070707070}, 12)
		e.a.xvex(1, 1, 0xdc, 1, r, mask, false)
		e.a.xvex(2, 1, 0x00, d, l, 1, false)
	case 98:
		mask := e.vectorConstantOperand([2]uint64{0x0f0f0f0f0f0f0f0f, 0x0f0f0f0f0f0f0f0f}, 12)
		e.a.xvex(1, 1, 0xdb, 1, l, mask, false)
		e.a.xvexShift(0x71, 0, l, 2, 4)
		e.a.xvex(1, 1, 0xdb, 0, 0, mask, false)
		table := e.vectorConstantOperand([2]uint64{0x0302020102010100, 0x0403030203020201}, 12)
		e.a.xvex(2, 1, 0x00, 2, table, 0, false)
		e.a.xvex(2, 1, 0x00, 11, table, 1, false)
		e.a.xvex(1, 1, 0xfc, d, 2, 11, false)
	case 101, 102, 133, 134:
		mapByte, code := byte(1), map[uint32]byte{101: 0x63, 102: 0x67, 133: 0x6b}[op]
		if op == 134 {
			mapByte, code = 2, 0x2b
		}
		e.a.xvex(mapByte, 1, code, d, l, r, false)
	case 123, 155:
		e.a.xvex(1, 1, 0xe0+byte((op-123)/32)*3, d, l, r, false)
	case 130:
		e.a.xvex(2, 1, 0x0b, 2, l, r, false)
		mask := e.vectorConstantOperand([2]uint64{0x8000800080008000, 0x8000800080008000}, 12)
		e.a.xvex(1, 1, 0x75, 0, l, mask, false)
		e.a.xvex(1, 1, 0x75, 1, r, mask, false)
		e.a.xvex(1, 1, 0xdb, 0, 0, 1, false)
		e.a.xvex(1, 1, 0xef, d, 2, 0, false)
	case 186:
		e.a.xvex(1, 1, 0xf5, d, l, r, false)
	default:
		size, kind := (op-96)/32, (op-96)%32
		if kind <= 10 {
			e.vectorExtendAVX2(d, l, size, (kind-7)%2 != 0, kind >= 9)
		} else {
			e.vectorExtMultiplyAVX2(d, l, r, size, (kind-28)%2 != 0, kind >= 30)
		}
	}
	e.vectorStore(in.Dst, d)
	return true
}

func (e *functionEmitter) vectorExtendAVX2(d, s byte, size uint32, high, unsigned bool) {
	if high {
		e.a.xvexShift(0x73, d, s, 3, 8)
		s = d
	}
	code := [...]byte{0x20, 0x23, 0x25}[size-1]
	if unsigned {
		code += 0x10
	}
	e.a.xvex(2, 1, code, d, 0, s, false)
}

func (e *functionEmitter) vectorExtMultiplyAVX2(d, l, r byte, size uint32, high, unsigned bool) {
	if size == 1 {
		e.vectorExtendAVX2(2, l, size, high, unsigned)
		e.vectorExtendAVX2(12, r, size, high, unsigned)
		e.a.xvex(1, 1, 0xd5, d, 2, 12, false)
	} else if size == 2 {
		code := byte(0xe5)
		if unsigned {
			code = 0xe4
		}
		e.a.xvex(1, 1, 0xd5, 2, l, r, false)
		e.a.xvex(1, 1, code, 12, l, r, false)
		code = 0x61
		if high {
			code = 0x69
		}
		e.a.xvex(1, 1, code, d, 2, 12, false)
	} else {
		control := byte(0x50)
		if high {
			control = 0xfa
		}
		e.a.xvex(1, 1, 0x70, 2, 0, l, false)
		e.a.bytes(control)
		e.a.xvex(1, 1, 0x70, 12, 0, r, false)
		e.a.bytes(control)
		mapByte, code := byte(2), byte(0x28)
		if unsigned {
			mapByte, code = 1, 0xf4
		}
		e.a.xvex(mapByte, 1, code, d, 2, 12, false)
	}
}

func (e *functionEmitter) vectorShuffleAVX2(lanes [16]byte, l, r byte) {
	if source, control, ok := vectorDwordShuffle(lanes); ok {
		s := l
		if source == 1 {
			s = r
		}
		e.a.xvex(1, 1, 0x70, 2, 0, s, false)
		e.a.bytes(control)
		return
	}
	first := true
	for source, s := range []byte{l, r} {
		var mask [16]byte
		used := false
		for i, n := range lanes {
			mask[i] = 0x80
			if int(n/16) == source {
				mask[i], used = n%16, true
			}
		}
		if !used {
			continue
		}
		index := e.vectorConstantOperand([2]uint64{binary.LittleEndian.Uint64(mask[:8]), binary.LittleEndian.Uint64(mask[8:])}, 12)
		d := byte(2)
		if !first {
			d = 11
		}
		e.a.xvex(2, 1, 0x00, d, s, index, false)
		if !first {
			e.a.xvex(1, 1, 0xeb, 2, 2, 11, false)
		}
		first = false
	}
}

func vectorDwordShuffle(lanes [16]byte) (source, control byte, aligned bool) {
	source, aligned = lanes[0]/16, true
	for i := 0; i < 4; i++ {
		n := lanes[i*4]
		aligned = aligned && n/16 == source && n%4 == 0
		for j := 1; j < 4; j++ {
			aligned = aligned && lanes[i*4+j] == n+byte(j)
		}
		control |= (n % 16 / 4) << (2 * i)
	}
	return
}
