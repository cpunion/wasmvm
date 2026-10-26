package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (a *assembler) arm(wide bool, op uint32) {
	if wide {
		op |= 1 << 31
	}
	a.word(op)
}
func (a *assembler) aconst(reg uint32, v uint64, wide bool) {
	count := 2
	if wide {
		count = 4
	}
	for i := 0; i < count; i++ {
		op := uint32(0x52800000)
		if i > 0 {
			op = 0x72800000
		}
		a.arm(wide, op|uint32(i)<<21|uint32(uint16(v>>(16*i)))<<5|reg)
	}
}

func integerARM64(a *assembler, op uint32, wide bool) {
	load := uint32(0xb9400000)
	shift := 2
	if wide {
		load = 0xf9400000
		shift = 3
	}
	a.word(load | uint32(nativeabi.ResultLoOffset>>shift)<<10 | 9)
	a.word(load | uint32(nativeabi.ResultHiOffset>>shift)<<10 | 10)
	zero, overflow, exit := a.label(), a.label(), a.label()
	divisorCheck, overflowCheck := false, false
	if op <= 0x5a {
		index := op - 0x45
		if wide {
			index = op - 0x50
		}
		if index == 0 {
			a.arm(wide, 0x7100013f)
		} else {
			a.arm(wide, 0x6b0a013f)
		}
		cc := [...]uint32{0, 0, 1, 11, 3, 12, 8, 13, 9, 10, 2}[index]
		a.word(0x1a9f07e9 | ((cc ^ 1) << 12))
	} else if op >= 0x67 && op <= 0x8a {
		kind := op - 0x67
		if wide {
			kind = op - 0x79
		}
		switch kind {
		case 0:
			a.arm(wide, 0x5ac01129)
		case 1:
			a.arm(wide, 0x5ac00129)
			a.arm(wide, 0x5ac01129)
		case 2:
			// AArch64 baseline SIMD counts all eight bytes; i32 loads zero the high bits.
			a.word(0x9e670120)
			a.word(0x0e205800)
			a.word(0x0e31b800)
			a.word(0x0e013c09)
		case 3:
			a.arm(wide, 0x0b0a0129)
		case 4:
			a.arm(wide, 0x4b0a0129)
		case 5:
			a.arm(wide, 0x1b0a7d29)
		case 6, 7, 8, 9:
			divisorCheck = true
			a.arm(wide, 0x7100015f)
			a.branch(zero, true, 0) // cmp b,0
			signed, rem := kind == 6 || kind == 8, kind >= 8
			if signed && !rem {
				overflowCheck = true
				ordinary := a.label()
				min := uint64(1) << 31
				if wide {
					min = uint64(1) << 63
				}
				a.aconst(11, min, wide)
				a.arm(wide, 0x6b0b013f)
				a.branch(ordinary, true, 1)
				a.arm(wide, 0x3100055f)
				a.branch(overflow, true, 0) // cmn b,1
				a.mark(ordinary)
			}
			div := uint32(0x1aca0920)
			if signed {
				div = 0x1aca0d20
			}
			if rem {
				a.arm(wide, div|11)
				a.arm(wide, 0x1b0aa569)
			} else {
				a.arm(wide, div|9)
			}
		case 10:
			a.arm(wide, 0x0a0a0129)
		case 11:
			a.arm(wide, 0x2a0a0129)
		case 12:
			a.arm(wide, 0x4a0a0129)
		case 13:
			a.arm(wide, 0x1aca2129)
		case 14:
			a.arm(wide, 0x1aca2929)
		case 15:
			a.arm(wide, 0x1aca2529)
		case 16:
			a.arm(wide, 0x4b0a03ea)
			a.arm(wide, 0x1aca2d29)
		case 17:
			a.arm(wide, 0x1aca2d29)
		}
	} else {
		switch op {
		case 0xa7, 0xad:
			a.word(0x2a0903e9)
		case 0xac, 0xc4:
			a.word(0x93407d29)
		case 0xc0:
			a.word(0x13001d29)
		case 0xc1:
			a.word(0x13003d29)
		case 0xc2:
			a.word(0x93401d29)
		case 0xc3:
			a.word(0x93403d29)
		}
	}
	a.word(0xf9000009 | uint32(nativeabi.ResultLoOffset/8)<<10)
	a.word(0x52800000 | uint32(nativeabi.Return)<<5)
	if divisorCheck {
		a.branch(exit, false, 0)
		a.mark(zero)
		atrap(a, wasm.TrapIntegerDivideByZero)
		if overflowCheck {
			a.branch(exit, false, 0)
			a.mark(overflow)
			atrap(a, wasm.TrapIntegerOverflow)
		}
	}
	a.mark(exit)
}

func atrap(a *assembler, trap wasm.TrapCode) {
	a.aconst(9, uint64(trap), false)
	a.word(0xf9000009 | uint32(nativeabi.PayloadOffset/8)<<10)
	a.word(0x52800000 | uint32(nativeabi.Trap)<<5)
}
