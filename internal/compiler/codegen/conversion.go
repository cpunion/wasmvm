package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func isConversion(op uint32) bool {
	return op >= 0xa8 && op <= 0xab || op >= 0xae && op <= 0xbb || op >= 0xfc0000 && op <= 0xfc0007
}

func (e *functionEmitter) loadFloat(slot slots.Slot, register byte, wide bool) {
	if source, ok := e.floatRegisters[slot]; ok {
		e.floatMove(register, source, wide)
	} else {
		e.slotLoad(slot, 0, false, wide)
		e.floatFromInteger(register, e.operandRegister(0), wide)
	}
}
func (e *functionEmitter) storeFloat(slot slots.Slot, register byte, wide bool) {
	if dst, ok := e.floatRegisters[slot]; ok {
		e.floatMove(dst, register, wide)
		return
	}
	e.floatToInteger(e.operandRegister(0), register, wide)
	e.slotStore(slot, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(slot, 0, true)
	}
}
func (e *functionEmitter) conversion(in slots.Instruction) {
	op := in.Op
	if op == 0xb6 || op == 0xbb {
		wide := op == 0xbb
		e.loadFloat(in.A, 0, !wide)
		if e.a.arch == "arm64" {
			ins := uint32(0x1e624000) // fcvt s0, d0
			if wide {
				ins = 0x1e22c000
			} // fcvt d0, s0
			e.a.word(ins)
		} else {
			e.a.xfloat(!wide, 0x5a, 0, 0)
		}
		e.storeFloat(in.Dst, 0, wide)
		return
	}
	if op >= 0xb2 && op <= 0xba {
		wide := op >= 0xb7
		variant := op - 0xb2
		if wide {
			variant = op - 0xb7
		}
		e.slotLoad(in.A, 0, false, variant >= 2)
		e.integerToFloat(variant >= 2, wide, variant&1 != 0)
		e.storeFloat(in.Dst, 0, wide)
		return
	}
	variant, saturating := op-0xa8, false
	if op >= 0xfc0000 {
		variant, saturating = op-0xfc0000, true
	} else if op >= 0xae {
		variant = op - 0xae + 4
	}
	floatWide, integerWide, unsigned := variant&2 != 0, variant&4 != 0, variant&1 != 0
	e.loadFloat(in.A, 0, floatWide)
	e.truncateFloat(floatWide, integerWide, unsigned, saturating, e.f.Sources[in.Source].Wasm.Offset)
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(in.Dst, 0, true)
	}
}

func (e *functionEmitter) integerToFloat(integerWide, floatWide, unsigned bool) {
	a := e.a
	if a.arch == "arm64" {
		ins := uint32(0x1e220120) // scvtf s0, w9
		if integerWide {
			ins |= 1 << 31
		}
		if floatWide {
			ins |= 1 << 22
		}
		if unsigned {
			ins |= 1 << 16
		}
		a.word(ins)
		return
	}
	prefix := byte(0xf3)
	if floatWide {
		prefix = 0xf2
	}
	convert := func(wide bool) {
		a.bytes(prefix)
		if wide {
			a.bytes(0x48)
		}
		a.bytes(0x0f, 0x2a, 0xc0) // cvtsi2ss/sd xmm0, eax/rax
	}
	if unsigned && integerWide {
		nonnegative, done := a.label(), a.label()
		a.xop(true, 0x85, 0xc0)
		a.branch(nonnegative, true, 9) // jns
		// Fold the discarded low bit into a sticky bit before conversion.
		// Doubling afterwards preserves nearest-even rounding over uint64.
		a.xop(true, 0x89, 0xc1)
		a.bytes(0x83, 0xe1, 1)
		a.xop(true, 0xd1, 0xe8)
		a.xop(true, 0x09, 0xc8)
		convert(true)
		a.xfloat(floatWide, 0x58, 0, 0)
		a.branch(done, false, 0)
		a.mark(nonnegative)
		convert(true)
		a.mark(done)
	} else {
		convert(integerWide || unsigned) // i32_u was zero-extended by slotLoad.
	}
}

func (e *functionEmitter) compareFloatConstant(number float64, wide bool) {
	bits := uint64(math.Float32bits(float32(number)))
	if wide {
		bits = math.Float64bits(number)
	}
	e.constant(1, bits)
	e.floatFromInteger(1, e.operandRegister(1), wide)
	if e.a.arch == "arm64" {
		ins := uint32(0x1e212000)
		if wide {
			ins |= 1 << 22
		}
		e.a.word(ins)
	} else {
		e.a.xfloatCompare(wide)
	}
}
func (e *functionEmitter) floatBranch(label int, condition string) {
	var cc byte
	switch condition {
	case "unordered":
		cc = 10
	case "lt":
		cc = 2
	case "le":
		cc = 6
	case "ge":
		cc = 3
	}
	if e.a.arch == "arm64" {
		switch condition {
		case "unordered":
			cc = 6
		case "lt":
			cc = 4
		case "le":
			cc = 9
		case "ge":
			cc = 10
		}
	}
	e.a.branch(label, true, cc)
}

func (e *functionEmitter) truncateFloat(floatWide, integerWide, unsigned, saturating bool, source uint64) {
	a := e.a
	width := 32
	if integerWide {
		width = 64
	}
	upper := math.Ldexp(1, width)
	lower := float64(-1)
	condition := "le" // Negative fractions greater than -1 truncate to unsigned zero.
	maximum := uint64(math.MaxUint32)
	if integerWide {
		maximum = math.MaxUint64
	}
	minimum := uint64(0)
	if !unsigned {
		upper /= 2
		lower, condition = -upper, "lt"
		maximum >>= 1
		minimum = maximum + 1
		// f64 -> i32 admits (-2^31-1, -2^31) after truncation. Other
		// signed boundaries have no representable fraction in this interval.
		if floatWide && !integerWide && !saturating {
			lower--
			condition = "le"
		}
	} else if saturating {
		lower, condition = 0, "lt"
	}
	var nan, low, high, done int
	if saturating {
		nan, low, high, done = a.label(), a.label(), a.label(), a.label()
	} else {
		nan = e.trapExit(wasm.TrapInvalidConversion, source)
		low = e.trapExit(wasm.TrapIntegerOverflow, source)
		high = low
	}
	e.compareFloatConstant(lower, floatWide)
	e.floatBranch(nan, "unordered")
	e.floatBranch(low, condition)
	e.compareFloatConstant(upper, floatWide)
	e.floatBranch(high, "ge")
	if a.arch == "arm64" {
		ins := uint32(0x1e380009) // fcvtzs w9, s0
		if floatWide {
			ins |= 1 << 22
		}
		if integerWide {
			ins |= 1 << 31
		}
		if unsigned {
			ins |= 1 << 16
		}
		a.word(ins)
	} else {
		prefix := byte(0xf3)
		if floatWide {
			prefix = 0xf2
		}
		convert := func() { a.bytes(prefix, 0x48, 0x0f, 0x2c, 0xc0) } // cvttss/sd2si rax, xmm0
		if unsigned && integerWide {
			highHalf, converted := a.label(), a.label()
			e.compareFloatConstant(math.Ldexp(1, 63), floatWide)
			e.floatBranch(highHalf, "ge")
			convert()
			a.branch(converted, false, 0)
			a.mark(highHalf)
			a.xfloat(floatWide, 0x5c, 0, 1)
			convert()
			e.constant(1, 1<<63)
			a.xop(true, 0x31, 0xc8)
			a.mark(converted)
		} else {
			convert()
		}
		if !integerWide {
			a.bytes(0x89, 0xc0)
		} // Canonical zero-extended i32 bits.
	}
	if saturating {
		a.branch(done, false, 0)
		a.mark(nan)
		e.constant(0, 0)
		a.branch(done, false, 0)
		a.mark(low)
		e.constant(0, minimum)
		a.branch(done, false, 0)
		a.mark(high)
		e.constant(0, maximum)
		a.mark(done)
	}
}
