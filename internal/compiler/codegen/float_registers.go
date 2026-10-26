package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (e *functionEmitter) floatFromInteger(dst, src byte, wide bool) {
	if e.a.arch == "amd64" {
		e.a.xfloatMove(wide, true, dst, src)
	} else {
		ins := uint32(0x1e270000)
		if wide {
			ins = 0x9e670000
		}
		e.a.word(ins | uint32(src)<<5 | uint32(dst))
	}
}
func (e *functionEmitter) floatToInteger(dst, src byte, wide bool) {
	if e.a.arch == "amd64" {
		e.a.xfloatMove(wide, false, src, dst)
	} else {
		ins := uint32(0x1e260000)
		if wide {
			ins = 0x9e660000
		}
		e.a.word(ins | uint32(src)<<5 | uint32(dst))
	}
}
func (e *functionEmitter) floatMove(dst, src byte, wide bool) {
	if dst == src {
		return
	}
	if e.a.arch == "arm64" {
		ins := uint32(0x1e204000)
		if wide {
			ins |= 1 << 22
		}
		e.a.word(ins | uint32(src)<<5 | uint32(dst))
	} else {
		if rex := byte(0x40) | (dst>>3)<<2 | src>>3; rex != 0x40 {
			e.a.bytes(rex)
		}
		e.a.bytes(0x0f, 0x28, 0xc0|(dst&7)<<3|(src&7))
	}
}

// Select arithmetic directly in fixed FP registers. Scratch registers 0-2 are
// excluded from allocation; constrained operations retain the common lowering.
func (e *functionEmitter) registerFloat(in slots.Instruction) bool {
	d, dok := e.floatRegisters[in.Dst]
	l, lok := e.floatRegisters[in.A]
	r, rok := e.floatRegisters[in.B]
	if !lok {
		return false
	}
	wide := e.f.Slots[in.A] == wasm.F64
	op := in.Op
	if op >= 0x99 && op <= 0xa6 {
		op -= 14
	} else if op >= 0x61 && op <= 0x66 {
		op -= 6
	}
	if op >= 0x5b && op <= 0x60 && rok && e.a.arch == "arm64" {
		d, dok = e.registers[in.Dst]
		if !dok {
			return false
		}
		ins := uint32(0x1e202000)
		if wide {
			ins |= 1 << 22
		}
		e.a.word(ins | uint32(r)<<16 | uint32(l)<<5)
		conditions := [...]uint32{0, 1, 4, 12, 9, 10}
		e.a.word(0x1a9f07e0 | (conditions[op-0x5b]^1)<<12 | uint32(d))
		return true
	}
	if !dok || op < 0x8b || op > 0x97 || op >= 0x92 && !rok {
		return false
	}
	if e.a.arch == "arm64" {
		ins := [...]uint32{0x1e20c000, 0x1e214000, 0x1e24c000, 0x1e254000,
			0x1e25c000, 0x1e244000, 0x1e21c000, 0x1e202800, 0x1e203800,
			0x1e200800, 0x1e201800, 0x1e205800, 0x1e204800}[op-0x8b]
		if wide {
			ins |= 1 << 22
		}
		if op >= 0x92 {
			ins |= uint32(r) << 16
		}
		e.a.word(ins | uint32(l)<<5 | uint32(d))
		return true
	}
	if op == 0x91 {
		e.a.xfloat(wide, 0x51, d, l)
		return true
	}
	if op >= 0x92 && op <= 0x95 {
		if d == r && d != l {
			e.floatMove(0, r, wide)
			r = 0
		}
		e.floatMove(d, l, wide)
		instructions := [...]byte{0x58, 0x5c, 0x59, 0x5e}
		e.a.xfloat(wide, instructions[op-0x92], d, r)
		return true
	}
	return false
}

// ARM scalar FP immediates encode a sign, an eight-value exponent range and
// four mantissa bits. Reject every nonrepresentable bit pattern exactly.
func floatImmediate(bits uint64, wide bool) (byte, bool) {
	mantissa, exponent := uint(23), uint(8)
	if wide {
		mantissa, exponent = 52, 11
	}
	if bits&((uint64(1)<<(mantissa-4))-1) != 0 {
		return 0, false
	}
	exp := (bits >> mantissa) & ((1 << exponent) - 1)
	high := exp &^ 3
	if high != 1<<(exponent-1) && high != (1<<(exponent-1))-4 {
		return 0, false
	}
	return byte((bits>>(mantissa+exponent))<<7 | ((exp>>(exponent-1))^1)<<6 | (exp&3)<<4 | (bits>>(mantissa-4))&15), true
}
