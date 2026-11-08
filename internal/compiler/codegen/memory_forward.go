package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// A successful private store proves the contained load's bounds. Forward only
// across copies: no other memory effect, trap, helper, poll or CFG entry can
// intervene. The selector retains every copied frame word and cold load body.
func (e *functionEmitter) forwardMemoryStoreLoad(start int, store slots.Instruction) (int, bool) {
	if !e.unmetered || e.memoryViews || e.sharedDefaultMemory ||
		store.Op < 0x36 || store.Op > 0x3e || store.Op == 0x38 || store.Op == 0x39 {
		return 0, false
	}
	// Forwarding extends the stored value's live range. A frame alias can
	// resolve to a different home; neither home may be reused in the span.
	shared := e.coalescedRegisterSlots()
	if shared[store.B] || shared[e.frameRead(store.B)] {
		return 0, false
	}
	_, width := memoryAccess(store.Op)
	offset := e.f.Sources[store.Source].Wasm.MemoryOffset()
	origins := map[slots.Slot]slots.Slot{}
	origin := func(s slots.Slot) slots.Slot {
		if v, ok := origins[s]; ok {
			return v
		}
		return s
	}
	for end := start + 1; end < e.codeLimit() && end-start <= 8 && end/256 == start/256; end++ {
		load := e.f.Code[end]
		if load.Op == 1 || load.Op == 0x1a {
			continue
		}
		if load.Op == 0x20 {
			local := slots.Slot(e.f.Sources[load.Source].Wasm.A)
			if load.Dst != local && (load.Dst == store.A || load.Dst == store.B ||
				load.Dst == e.frameRead(store.A) || load.Dst == e.frameRead(store.B)) {
				return 0, false
			}
			origins[load.Dst] = origin(local)
			continue
		}
		if load.Op < 0x28 || load.Op > 0x35 || load.Op == 0x2a || load.Op == 0x2b || origin(load.A) != store.A {
			return 0, false
		}
		_, size := memoryAccess(load.Op)
		at := e.f.Sources[load.Source].Wasm.MemoryOffset()
		if at < offset || at-offset > uint64(width) || uint64(size) > uint64(width)-(at-offset) {
			return 0, false
		}
		e.memoryAt(store, e.provenPCs[start], offset, e.memoryOperands[start])
		for _, copy := range e.f.Code[start+1 : end] {
			e.plain(copy, false)
		}
		e.forwardMemoryValue(store.B, load.Dst, uint32(at-offset)*8, size,
			load.Op == 0x2c || load.Op == 0x2e || load.Op == 0x30 || load.Op == 0x32 || load.Op == 0x34)
		e.boundsElided++
		e.forwardedMemory++
		return end, true
	}
	return 0, false
}

func (e *functionEmitter) forwardMemoryValue(source, dst slots.Slot, shift, width uint32, signed bool) {
	wide := e.f.Slots[dst] == wasm.I64
	if e.a.arch == "arm64" {
		input, inputRegister := e.registers[e.frameRead(source)]
		if !inputRegister {
			e.slotLoad(source, 0, false, e.f.Slots[source] == wasm.I64)
			input = 9
		}
		output, outputRegister := e.registers[dst]
		if !outputRegister {
			output = 9
		}
		if width == 8 {
			e.registerMove(output, input, true)
		} else {
			// Extract directly between allocated homes. A separate shift,
			// extension and two scratch moves lengthen the dependency chain.
			high := shift + width*8 - 1
			wideForm := wide || high >= 32
			op := uint32(0x53000000) // UBFM Wd,Wn,#lsb,#msb.
			if wideForm {
				op = 0xd3400000
			}
			if signed && (!wideForm || wide) {
				op &^= 1 << 30
			}
			e.a.word(op | shift<<16 | high<<10 | uint32(input)<<5 | uint32(output))
			if signed && wideForm && !wide {
				// A high i64 byte/halfword still produces a canonical i32.
				e.a.word(0x13000000 | (width*8-1)<<10 | uint32(output)<<5 | uint32(output))
			}
		}
		if !outputRegister {
			e.slotStore(dst, 0, false)
		}
		return
	} else {
		e.slotLoad(source, 0, false, e.f.Slots[source] == wasm.I64)
		if shift != 0 {
			e.a.bytes(0x48, 0xc1, 0xe8, byte(shift))
		}
		if width < 4 {
			op := byte(0xb6)
			if width == 2 {
				op++
			}
			if signed {
				op += 8
				if wide {
					e.a.bytes(0x48)
				}
			}
			e.a.bytes(0x0f, op, 0xc0)
		} else if width == 4 {
			if signed {
				e.a.bytes(0x48, 0x63, 0xc0)
			} else {
				e.a.bytes(0x89, 0xc0)
			}
		}
	}
	e.slotStore(dst, 0, false)
}
