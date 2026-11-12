package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Physical frame transfers preserve the same complete low word as the scratch
// path. F32 spills still zero its unused upper 32 bits through that path;
// F64 and integer homes can transfer directly without an intermediate move.
func (e *functionEmitter) frameRegister(slot slots.Slot, reg byte, floating, store bool) {
	if !store {
		slot = e.frameRead(slot)
	}
	if floating && store && e.f.Slots[slot] == wasm.F32 {
		e.floatToInteger(e.operandRegister(0), reg, false)
		e.slotStore(slot, 0, false)
		return
	}
	offset := uint32(slot) * 16
	a := e.a
	if a.arch == "arm64" {
		base := uint32(12)
		wide := !floating || e.f.Slots[slot] == wasm.F64
		scale := uint32(8)
		ins := uint32(0xf9400000)
		if floating {
			ins = 0xfd400000
			if !wide {
				ins, scale = 0xbd400000, 4
			}
		}
		if store {
			ins -= 0x00400000
		}
		if offset/scale >= 4096 {
			a.aconst(15, uint64(offset), false)
			a.word(0x8b0f018f)
			base, offset = 15, 0
		}
		a.word(ins | uint32(reg) | base<<5 | (offset/scale)<<10)
		return
	}
	if !floating {
		a.bytes(0x48 | (reg>>3)<<2)
		op := byte(0x8b)
		if store {
			op = 0x89
		}
		a.bytes(op, 0x86|(reg&7)<<3)
	} else {
		prefix, op := byte(0xf3), byte(0x7e)
		if store {
			prefix, op = 0x66, 0xd6
		} else if e.f.Slots[slot] == wasm.F32 {
			prefix, op = 0x66, 0x6e
		}
		a.bytes(prefix)
		if reg >= 8 {
			a.bytes(0x44)
		}
		a.bytes(0x0f, op, 0x86|(reg&7)<<3)
	}
	a.word(offset)
}
