package codegen

import (
	"encoding/binary"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

type vectorAffine struct {
	first, second, a, b slots.Instruction
	prefix              *slots.Instruction
	input               slots.Slot
	end                 int
	size                uint32
	multiplier, addend  [2]uint64
}

// Recognize bounded integer affine expressions with two vector literals.
// Distributivity is exact modulo each lane width, including signed overflow.
// This does not apply to floating-point operations or cross a poll boundary.
func (e *functionEmitter) vectorAffinePlan(pc int) (p vectorAffine, ok bool) {
	start := pc
	if !e.unmetered || e.a.arch != "arm64" {
		return p, false
	}
	if e.f.Code[pc].Op == 0x20 {
		p.prefix = &e.f.Code[pc]
		pc++
	}
	if pc+3 >= e.codeLimit() || start/256 != (pc+3)/256 {
		return p, false
	}
	p.a, p.first, p.b, p.second = e.f.Code[pc], e.f.Code[pc+1], e.f.Code[pc+2], e.f.Code[pc+3]
	p.end = pc + 3
	if p.a.Op != 0xfd000c || p.b.Op != 0xfd000c || p.first.B != p.a.Dst || p.second.A != p.first.Dst || p.second.B != p.b.Dst || p.b.Dst == p.first.Dst {
		return p, false
	}
	if p.first.Op != 0xfd008e && p.first.Op != 0xfd0095 && p.first.Op != 0xfd00ae && p.first.Op != 0xfd00b5 {
		return p, false
	}
	addFirst := p.first.Op == 0xfd008e || p.first.Op == 0xfd00ae
	add, mul := uint32(0xfd008e), uint32(0xfd0095)
	p.size = 1
	if p.first.Op >= 0xfd00ae {
		add, mul, p.size = 0xfd00ae, 0xfd00b5, 2
	}
	if addFirst && p.second.Op != mul || !addFirst && p.second.Op != add {
		return p, false
	}
	p.input = p.first.A
	if p.input == p.a.Dst {
		return p, false
	}
	if p.prefix != nil {
		if p.prefix.Dst != p.input {
			return p, false
		}
		p.input = slots.Slot(e.f.Sources[p.prefix.Source].Wasm.A)
	}
	if p.input == p.a.Dst {
		return p, false
	}
	a, b := e.f.Sources[p.a.Source].Wasm.Vector, e.f.Sources[p.b.Source].Wasm.Vector
	multiplier, addend := a, b
	if addFirst {
		multiplier = b
		for i := 0; i < 16; i += 1 << p.size {
			if p.size == 1 {
				binary.LittleEndian.PutUint16(addend[i:], binary.LittleEndian.Uint16(a[i:])*binary.LittleEndian.Uint16(b[i:]))
			} else {
				binary.LittleEndian.PutUint32(addend[i:], binary.LittleEndian.Uint32(a[i:])*binary.LittleEndian.Uint32(b[i:]))
			}
		}
	}
	p.multiplier = [2]uint64{binary.LittleEndian.Uint64(multiplier[:8]), binary.LittleEndian.Uint64(multiplier[8:])}
	p.addend = [2]uint64{binary.LittleEndian.Uint64(addend[:8]), binary.LittleEndian.Uint64(addend[8:])}
	return p, true
}

func (e *functionEmitter) combineVectorAffine(pc int) (int, bool) {
	p, ok := e.vectorAffinePlan(pc)
	if !ok {
		return pc, false
	}
	input, hasInput := e.vectorRegisters[p.input]
	mult, hasMult := e.vectorConstants[p.multiplier]
	add, hasAdd := e.vectorConstants[p.addend]
	if !hasInput || !hasMult || !hasAdd {
		return pc, false
	}
	end, dst := p.end, p.second.Dst
	if end+1 < e.codeLimit() && pc/256 == (end+1)/256 {
		next := e.f.Code[end+1]
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == dst {
			dst = slots.Slot(e.f.Sources[next.Source].Wasm.A)
			end++
		}
	}
	d, ok := e.vectorRegisters[dst]
	if !ok {
		return pc, false
	}
	e.forgetVectorAliases(dst)
	final := map[slots.Slot]int{}
	if p.prefix != nil {
		final[p.prefix.Dst] = 0
	}
	final[p.a.Dst], final[p.first.Dst], final[p.b.Dst], final[p.second.Dst] = 1, 2, 3, 4
	final[dst] = 4
	if end != p.end && e.f.Code[end].Dst != slots.NoSlot {
		final[e.f.Code[end].Dst] = 4
	}
	live := func(s slots.Slot, kind int) bool {
		return final[s] == kind && (s == dst || !e.overwrittenBeforeExit(end+1, s))
	}
	if p.prefix != nil && live(p.prefix.Dst, 0) {
		e.vectorStore(p.prefix.Dst, input)
	}
	keepFirst := live(p.first.Dst, 2)
	if keepFirst {
		e.literal(p.a)
		r := e.vectorOperand(p.a.Dst, 0)
		kind := uint32(14)
		if p.first.Op == 0xfd0095 || p.first.Op == 0xfd00b5 {
			kind = 21
		}
		e.vectorBinary(1, input, r, p.size, kind)
	}
	if d == input {
		e.vectorMove(0, input)
		input = 0
	}
	e.vectorMove(d, add)
	e.a.word(0x4e209400 | p.size<<22 | uint32(mult)<<16 | uint32(input)<<5 | uint32(d))
	if keepFirst {
		e.vectorStore(p.first.Dst, 1)
	}
	if live(p.a.Dst, 1) {
		e.literal(p.a)
	}
	if live(p.b.Dst, 3) {
		e.literal(p.b)
	}
	if p.second.Dst != dst && live(p.second.Dst, 4) {
		e.vectorStore(p.second.Dst, d)
	}
	if end != p.end {
		if s := e.f.Code[end].Dst; s != slots.NoSlot && s != dst && live(s, 4) {
			e.vectorStore(s, d)
		}
	}
	return end, true
}
