package codegen

import (
	"encoding/binary"
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// Vector and scalar FP allocations share a physical register pool. Scratch
// vectors 0-2 and the target's auxiliary registers never hold live guest slots.
func (e *functionEmitter) vectorMove(dst, src byte) {
	if dst == src {
		return
	}
	if e.a.arch == "arm64" {
		e.a.word(0x4ea01c00 | uint32(src)<<16 | uint32(src)<<5 | uint32(dst))
	} else {
		e.a.xvector(0x6f, dst, src)
	}
}

func (a *assembler) xvector(op, dst, src byte) {
	a.bytes(0x66)
	if rex := byte(0x40) | (dst>>3)<<2 | src>>3; rex != 0x40 {
		a.bytes(rex)
	}
	a.bytes(0x0f, op, 0xc0|(dst&7)<<3|(src&7))
}

func (e *functionEmitter) vectorMemory(slot slots.Slot, reg byte, store bool) {
	offset := uint32(slot) * 16
	if e.a.arch == "arm64" {
		base := uint32(12)
		if offset/16 >= 4096 {
			e.a.aconst(15, uint64(offset), false)
			e.a.word(0x8b0f018f)
			base, offset = 15, 0
		}
		ins := uint32(0x3dc00000)
		if store {
			ins = 0x3d800000
		}
		e.a.word(ins | (offset/16)<<10 | base<<5 | uint32(reg))
	} else {
		e.a.bytes(0xf3)
		if reg >= 8 {
			e.a.bytes(0x44)
		}
		op := byte(0x6f)
		if store {
			op = 0x7f
		}
		e.a.bytes(0x0f, op, 0x86|(reg&7)<<3)
		e.a.word(offset)
	}
}

func (e *functionEmitter) vectorOperand(slot slots.Slot, scratch byte) byte {
	if r, ok := e.vectorRegisters[slot]; ok {
		return r
	}
	e.vectorMemory(slot, scratch, false)
	return scratch
}
func (e *functionEmitter) vectorStore(slot slots.Slot, reg byte) {
	if r, ok := e.vectorRegisters[slot]; ok {
		e.vectorMove(r, reg)
	} else {
		e.vectorMemory(slot, reg, true)
	}
}
func (e *functionEmitter) vectorDestination(slot slots.Slot) byte {
	if r, ok := e.vectorRegisters[slot]; ok {
		return r
	}
	return 2
}
func (e *functionEmitter) vectorConstant(dst byte, lo, hi uint64) {
	e.constant(0, lo)
	e.floatFromInteger(dst, e.operandRegister(0), true)
	e.constant(0, hi)
	if e.a.arch == "arm64" {
		e.a.word(0x4e181d20 | uint32(dst))
	} else {
		tmp := byte(12)
		if dst == 12 {
			tmp = 11
		}
		e.floatFromInteger(tmp, 0, true)
		e.a.xvector(0x6c, dst, tmp)
	}
}

// Expensive vector literals are immutable for an entry. Stable ordering keeps
// code bytes independent of map iteration and preserves all cold continuations.
func (e *functionEmitter) prepareVectorConstants() {
	weights := map[[2]uint64]int{}
	for _, in := range e.f.Code {
		if in.Op == 0xfd000c {
			b := e.f.Sources[in.Source].Wasm.Vector
			weights[[2]uint64{binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:])}]++
		}
	}
	var values [][2]uint64
	for v := range weights {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		if weights[values[i]] != weights[values[j]] {
			return weights[values[i]] > weights[values[j]]
		}
		if values[i][0] != values[j][0] {
			return values[i][0] < values[j][0]
		}
		return values[i][1] < values[j][1]
	})
	pool := []byte{13, 14, 15}
	if e.a.arch == "arm64" {
		pool = []byte{18, 19, 20, 21, 22, 23}
	}
	e.vectorConstants = map[[2]uint64]byte{}
	for i, v := range values[:min(len(values), len(pool))] {
		e.vectorConstant(pool[i], v[0], v[1])
		e.vectorConstants[v] = pool[i]
	}
}

// Substitute an immutable literal only for an input; it must never become the
// writable result register. Restore the slot assignment before publishing any
// still-observable copy of the literal.
func (e *functionEmitter) pooledVector(in, constant slots.Instruction) bool {
	if !nativeVectorOpcode(in.Op) || constant.Op != 0xfd000c || constant.Dst != in.B || in.A == in.B || in.C == in.B || in.Dst == in.B {
		return false
	}
	b := e.f.Sources[constant.Source].Wasm.Vector
	r, ok := e.vectorConstants[[2]uint64{binary.LittleEndian.Uint64(b[:8]), binary.LittleEndian.Uint64(b[8:])}]
	if !ok {
		return false
	}
	previous, had := e.vectorRegisters[in.B]
	e.vectorRegisters[in.B] = r
	e.vectorInstruction(in)
	if had {
		e.vectorRegisters[in.B] = previous
	} else {
		delete(e.vectorRegisters, in.B)
	}
	return true
}
