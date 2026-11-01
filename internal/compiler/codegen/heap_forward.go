package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Forward a just-stored field only on verified fallthrough. Interior PCs retain
// ordinary checked cold bodies. No call, allocation, mutation or poll boundary
// can intervene; the successful store already proved liveness and array bounds.
func (e *functionEmitter) forwardHeapStoreLoad(start int) (int, bool) {
	if !e.unmetered {
		return 0, false
	}
	store := e.f.Code[start]
	a, ok := e.heapAccesses[store.Source]
	if !ok || !a.store {
		return 0, false
	}
	end := start + 1
	for ; end < e.codeLimit() && end-start <= 4 && end/256 == start/256; end++ {
		in := e.f.Code[end]
		if in.Op == 1 || in.Op == 0x1a || in.Op == 0x20 && in.Dst == slots.Slot(e.f.Sources[in.Source].Wasm.A) {
			continue
		}
		b, ok := e.heapAccesses[in.Source]
		if !ok || b.store || b.length || in.Op>>16 != 0xfb || b.cacheClass != a.cacheClass || b.array != a.array || b.offset != a.offset || b.width != a.width || in.A != store.A || a.array && in.B != store.B {
			return 0, false
		}
		value := store.B
		if a.array {
			value = e.f.Args[store.Extra+2]
		}
		e.heapInstruction(store)
		e.forwardHeapValue(value, in.Dst, b.width, b.signed)
		return end, true
	}
	return 0, false
}

func (e *functionEmitter) forwardHeapValue(source, dst slots.Slot, width uint32, signed bool) {
	if source == dst && width >= 8 {
		return
	}
	if width == 16 {
		e.vectorStore(dst, e.vectorOperand(source, 0))
		return
	}
	if source == dst && width == 4 {
		if reg, ok := e.registers[dst]; ok {
			e.registerMove(reg, reg, false)
			return
		}
		if _, ok := e.floatRegisters[dst]; ok {
			return // An allocated f32 home already contains exactly 32 bits.
		}
	}
	e.slotLoad(source, 0, false, width == 8)
	if width < 4 {
		if e.a.arch == "arm64" {
			op := uint32(0x53000129) // ubfm w9,w9,#0,#bits-1
			if signed {
				op = 0x13000129
			}
			e.a.word(op | (width*8-1)<<10)
		} else {
			op := byte(0xb6)
			if width == 2 {
				op++
			}
			if signed {
				op += 8
			}
			e.a.bytes(0x0f, op, 0xc0)
		}
	}
	e.heapResult(dst)
}
