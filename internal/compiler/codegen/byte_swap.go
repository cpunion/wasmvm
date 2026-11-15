package codegen

import (
	"math/bits"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// A byte expression names each original byte with 1..4, or zero for a cleared
// byte. Only whole-byte masks/shifts and disjoint logical merges are admitted.
type byteExpression struct {
	bytes, literal uint32
	constant       bool
}

func byteMask(v uint32) bool {
	for range 4 {
		if b := byte(v); b != 0 && b != 255 {
			return false
		}
		v >>= 8
	}
	return true
}

func mergeBytes(a, b uint32, xor bool) (uint32, bool) {
	var out uint32
	for i := uint32(0); i < 32; i += 8 {
		x, y := byte(a>>i), byte(b>>i)
		if x != 0 && y != 0 && x != y {
			return 0, false
		}
		v := x | y
		if xor && x == y {
			v = 0
		}
		out |= uint32(v) << i
	}
	return out, true
}

// Recognize a bounded pure byte permutation rather than a producer-specific
// instruction sequence. Dead intermediates may disappear; original PCs retain
// their ordinary cold entries, and precise-fuel code retains every operation.
func (e *functionEmitter) combineByteSwap(start int) (int, bool) {
	if !e.unmetered || start+2 >= e.codeLimit() {
		return 0, false
	}
	first, next := e.f.Code[start], e.f.Code[start+1]
	root := slots.NoSlot
	if first.Op == 0x20 && next.Op == 0x41 {
		root = slots.Slot(e.f.Sources[first.Source].Wasm.A)
	} else if first.Op == 0x41 && next.B == first.Dst && next.A != first.Dst {
		root = next.A
	}
	if root == slots.NoSlot {
		return 0, false
	}
	if e.f.Slots[root] != wasm.I32 {
		return 0, false
	}
	l, ok := e.registers[e.frameRead(root)]
	if !ok {
		return 0, false
	}
	values := map[slots.Slot]byteExpression{root: {bytes: 0x04030201}}
	written := map[slots.Slot]bool{}
	var order []slots.Slot
	for pc := start; pc < e.codeLimit() && pc-start < 24 && pc/256 == start/256; pc++ {
		in := e.f.Code[pc]
		if in.Dst == slots.NoSlot || e.f.Slots[in.Dst] != wasm.I32 {
			return 0, false
		}
		a, aok := values[in.A]
		b, bok := values[in.B]
		var v byteExpression
		switch in.Op {
		case 0x20:
			v, ok = values[slots.Slot(e.f.Sources[in.Source].Wasm.A)]
		case 0x41:
			v, ok = byteExpression{literal: uint32(e.f.Sources[in.Source].Wasm.Bits), constant: true}, true
		case 0x71:
			if a.constant {
				a, b = b, a
			}
			ok = aok && bok && !a.constant && b.constant && byteMask(b.literal)
			v.bytes = a.bytes & b.literal
		case 0x72, 0x73:
			ok = aok && bok && !a.constant && !b.constant
			if ok {
				v.bytes, ok = mergeBytes(a.bytes, b.bytes, in.Op == 0x73)
			}
		case 0x74, 0x76, 0x77, 0x78:
			n := b.literal & 31
			ok = aok && bok && !a.constant && b.constant && n%8 == 0
			switch in.Op {
			case 0x74:
				v.bytes = a.bytes << n
			case 0x76:
				v.bytes = a.bytes >> n
			case 0x77:
				v.bytes = bits.RotateLeft32(a.bytes, int(n))
			case 0x78:
				v.bytes = bits.RotateLeft32(a.bytes, -int(n))
			}
		default:
			return 0, false
		}
		if !ok {
			return 0, false
		}
		if !written[in.Dst] {
			order = append(order, in.Dst)
		}
		values[in.Dst], written[in.Dst] = v, true
		if v.constant || v.bytes != 0x01020304 {
			continue
		}
		d, cached := e.registers[in.Dst]
		if !cached {
			return 0, false
		}
		masks := map[slots.Slot]uint32{}
		for _, slot := range order {
			value := values[slot]
			if slot == in.Dst || slot == root && !value.constant && value.bytes == 0x04030201 || e.overwrittenBeforeExit(pc+1, slot) {
				continue
			}
			if reg, cached := e.registers[slot]; cached && reg == d || d == e.operandRegister(0) {
				return 0, false
			}
			if value.constant {
				continue
			}
			var mask uint32
			for n := uint32(0); n < 32; n += 8 {
				b := byte(value.bytes >> n)
				if b != 0 && b != byte(uint32(0x01020304)>>n) {
					return 0, false
				}
				if b != 0 {
					mask |= 255 << n
				}
			}
			if !e.canRegisterImmediate(0x71, slot, in.Dst, uint64(mask)) || e.registers[slot] == e.operandRegister(0) {
				return 0, false
			}
			masks[slot] = mask
		}
		if e.a.arch == "arm64" {
			e.a.word(0x5ac00800 | uint32(l)<<5 | uint32(d)) // REV Wd, Wn.
		} else {
			e.registerMove(d, l, false)
			if d >= 8 {
				e.a.bytes(0x41)
			}
			e.a.bytes(0x0f, 0xc8|d&7) // BSWAP r32.
		}
		// Preserve observable intermediate words without replaying the shuffle.
		// Masked projections of the final permutation need one logical operation.
		for _, slot := range order {
			if mask, ok := masks[slot]; ok {
				e.registerImmediate(0x71, slot, in.Dst, uint64(mask))
			}
		}
		for _, slot := range order {
			if value := values[slot]; value.constant && slot != in.Dst && !e.overwrittenBeforeExit(pc+1, slot) {
				e.constant(0, uint64(value.literal))
				e.slotStore(slot, 0, false)
			}
		}
		return pc, true
	}
	return 0, false
}
