package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (a *assembler) xop(wide bool, b ...byte) {
	if wide {
		a.bytes(0x48)
	}
	a.bytes(b...)
}
func (a *assembler) xconst(reg byte, v uint64) { a.bytes(0x48, 0xb8+reg); a.imm64(v) }

func integerAMD64(a *assembler, op uint32, wide bool) {
	a.xop(wide, 0x8b, 0x47, byte(nativeabi.ResultLoOffset)) // rax/eax = a
	a.xop(wide, 0x8b, 0x4f, byte(nativeabi.ResultHiOffset)) // rcx/ecx = b
	zero, overflow, exit := a.label(), a.label(), a.label()
	divisorCheck, overflowCheck := false, false
	if op <= 0x5a {
		index := op - 0x45
		if wide {
			index = op - 0x50
		}
		if index == 0 {
			a.xop(wide, 0x85, 0xc0)
		} else {
			a.xop(wide, 0x39, 0xc8)
		}
		cc := [...]byte{4, 4, 5, 12, 2, 15, 7, 14, 6, 13, 3}[index]
		a.bytes(0x0f, 0x90+cc, 0xc0, 0x0f, 0xb6, 0xc0)
	} else if op >= 0x67 && op <= 0x8a {
		kind := op - 0x67
		if wide {
			kind = op - 0x79
		}
		switch kind {
		case 0, 1: // bsr/bsf are baseline instructions; handle zero explicitly.
			isZero, done := a.label(), a.label()
			a.xop(wide, 0x85, 0xc0)
			a.branch(isZero, true, 4)
			ins := byte(0xbd)
			if kind == 1 {
				ins = 0xbc
			}
			a.xop(wide, 0x0f, ins, 0xc0)
			bits := byte(32)
			if wide {
				bits = 64
			}
			if kind == 0 {
				a.bytes(0x83, 0xf0, bits-1)
			}
			a.branch(done, false, 0)
			a.mark(isZero)
			a.bytes(0xb8)
			a.word(uint32(bits))
			a.mark(done)
		case 2:
			xpopcount(a)
		case 3:
			a.xop(wide, 0x01, 0xc8)
		case 4:
			a.xop(wide, 0x29, 0xc8)
		case 5:
			a.xop(wide, 0x0f, 0xaf, 0xc1)
		case 6, 7, 8, 9:
			divisorCheck = true
			a.xop(wide, 0x85, 0xc9)
			a.branch(zero, true, 4)
			signed, rem := kind == 6 || kind == 8, kind >= 8
			done := a.label()
			if signed {
				ordinary := a.label()
				a.xop(wide, 0x83, 0xf9, 0xff)
				a.branch(ordinary, true, 5)
				min := uint64(1) << 31
				if wide {
					min = uint64(1) << 63
				}
				a.xconst(2, min)
				a.xop(wide, 0x39, 0xd0)
				a.branch(ordinary, true, 5)
				if rem {
					a.bytes(0x31, 0xc0)
					a.branch(done, false, 0)
				} else {
					overflowCheck = true
					a.branch(overflow, false, 0)
				}
				a.mark(ordinary)
				a.xop(wide, 0x99)
				a.xop(wide, 0xf7, 0xf9)
			} else {
				a.bytes(0x31, 0xd2)
				a.xop(wide, 0xf7, 0xf1)
			}
			if rem {
				a.xop(wide, 0x89, 0xd0)
			}
			a.mark(done)
		case 10:
			a.xop(wide, 0x21, 0xc8)
		case 11:
			a.xop(wide, 0x09, 0xc8)
		case 12:
			a.xop(wide, 0x31, 0xc8)
		case 13:
			a.xop(wide, 0xd3, 0xe0)
		case 14:
			a.xop(wide, 0xd3, 0xf8)
		case 15:
			a.xop(wide, 0xd3, 0xe8)
		case 16:
			a.xop(wide, 0xd3, 0xc0)
		case 17:
			a.xop(wide, 0xd3, 0xc8)
		}
	} else {
		switch op {
		case 0xa7, 0xad:
			a.bytes(0x89, 0xc0)
		case 0xac, 0xc4:
			a.bytes(0x48, 0x63, 0xc0)
		case 0xc0, 0xc2:
			a.xop(wide, 0x0f, 0xbe, 0xc0)
		case 0xc1, 0xc3:
			a.xop(wide, 0x0f, 0xbf, 0xc0)
		}
	}
	a.bytes(0x48, 0x89, 0x47, byte(nativeabi.ResultLoOffset), 0xb8)
	a.word(uint32(nativeabi.Return))
	if divisorCheck {
		a.branch(exit, false, 0)
		a.mark(zero)
		xtrap(a, wasm.TrapIntegerDivideByZero)
		if overflowCheck {
			a.branch(exit, false, 0)
			a.mark(overflow)
			xtrap(a, wasm.TrapIntegerOverflow)
		}
	}
	a.mark(exit)
}

func xtrap(a *assembler, trap wasm.TrapCode) {
	a.bytes(0x48, 0xc7, 0x47, byte(nativeabi.PayloadOffset))
	a.word(uint32(trap))
	a.bytes(0xb8)
	a.word(uint32(nativeabi.Trap))
}

// SWAR popcount works on the baseline x86-64 CPU without optional POPCNT/BMI.
func xpopcount(a *assembler) {
	mask := func(v uint64) { a.bytes(0x49, 0xb8); a.imm64(v) }
	a.bytes(0x48, 0x89, 0xc2, 0x48, 0xc1, 0xea, 1)
	mask(0x5555555555555555)
	a.bytes(0x4c, 0x21, 0xc2, 0x48, 0x29, 0xd0, 0x48, 0x89, 0xc2, 0x48, 0xc1, 0xea, 2)
	mask(0x3333333333333333)
	a.bytes(0x4c, 0x21, 0xc0, 0x4c, 0x21, 0xc2, 0x48, 0x01, 0xd0, 0x48, 0x89, 0xc2, 0x48, 0xc1, 0xea, 4, 0x48, 0x01, 0xd0)
	mask(0x0f0f0f0f0f0f0f0f)
	a.bytes(0x4c, 0x21, 0xc0)
	mask(0x0101010101010101)
	a.bytes(0x49, 0x0f, 0xaf, 0xc0, 0x48, 0xc1, 0xe8, 56)
}
