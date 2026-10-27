package codegen

import (
	"math/bits"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// Select UBFIZ for (x & lowMask) << shift, also recognizing multiplication
// by a power of two. The discarded mask result must be overwritten before any
// observation. Original PCs retain independent, unfused continuation entries.
func (e *functionEmitter) combineMaskShift(pc int) (int, bool) {
	if !e.unmetered || e.a.arch != "arm64" || pc+3 >= e.codeLimit() || pc/256 != (pc+3)/256 {
		return 0, false
	}
	mask, and, amount, shift := e.f.Code[pc], e.f.Code[pc+1], e.f.Code[pc+2], e.f.Code[pc+3]
	if (mask.Op != 0x41 && mask.Op != 0x42) || amount.Op != mask.Op {
		return 0, false
	}
	wide := mask.Op == 0x42
	andOp, shiftOp, mulOp, width := uint32(0x71), uint32(0x74), uint32(0x6c), uint32(32)
	if wide {
		andOp, shiftOp, mulOp, width = 0x83, 0x86, 0x7e, 64
	}
	if and.Op != andOp || shift.Op != shiftOp && shift.Op != mulOp ||
		and.B != mask.Dst || and.A == mask.Dst || shift.A != and.Dst || shift.B != amount.Dst || amount.Dst == and.Dst {
		return 0, false
	}
	v, n := e.f.Sources[mask.Source].Wasm.Bits, e.f.Sources[amount.Source].Wasm.Bits
	if !wide {
		v, n = uint64(uint32(v)), uint64(uint32(n))
	}
	if v == 0 || v&(v+1) != 0 {
		return 0, false
	}
	if shift.Op == mulOp {
		if n == 0 || n&(n-1) != 0 {
			return 0, false
		}
		n = uint64(bits.TrailingZeros64(n))
	}
	n &= uint64(width - 1)
	end, dst := pc+3, shift.Dst
	if end+1 < e.codeLimit() && (end+1)/256 == pc/256 {
		next := e.f.Code[end+1]
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == shift.Dst {
			local := slots.Slot(e.f.Sources[next.Source].Wasm.A)
			if _, ok := e.registers[local]; ok {
				end++
				dst = local
			}
		}
	}
	l, lok := e.registers[and.A]
	d, dok := e.registers[dst]
	if !lok || !dok {
		return 0, false
	}
	// A final assignment can overwrite a temporary as well as a following
	// instruction. Do not materialize any value superseded within the group.
	overwrites := func(start int, slot slots.Slot) bool {
		for _, in := range e.f.Code[start : end+1] {
			if in.Dst == slot || (in.Op == 0x21 || in.Op == 0x22) && slots.Slot(e.f.Sources[in.Source].Wasm.A) == slot {
				return true
			}
		}
		return e.overwrittenBeforeExit(end+1, slot)
	}
	if !overwrites(pc+3, and.Dst) {
		return 0, false
	}
	length := min(uint32(bits.Len64(v)), width-uint32(n))
	word := uint32(0x53000000) | ((width-uint32(n))&(width-1))<<16 | (length-1)<<10
	if wide {
		word |= 1<<31 | 1<<22
	}
	e.a.word(word | uint32(l)<<5 | uint32(d))
	if !overwrites(pc+1, mask.Dst) {
		e.literal(mask)
	}
	if !overwrites(pc+3, amount.Dst) {
		e.literal(amount)
	}
	if end != pc+3 {
		if !e.overwrittenBeforeExit(end+1, shift.Dst) {
			e.move(shift.Dst, dst)
		}
		if !e.overwrittenBeforeExit(end+1, e.f.Code[end].Dst) {
			e.move(e.f.Code[end].Dst, dst)
		}
	}
	return end, true
}
