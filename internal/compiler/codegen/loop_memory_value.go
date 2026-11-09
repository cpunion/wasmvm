package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopMemoryValue struct {
	load, store int
	address     uint32
	width       uint32
	register    byte
}

// Only one checked private word is observed by this nontrapping loop. Keep its
// last stored bits in an otherwise unused register; all original stores and
// backedge polls remain. No cache survives a native entry or a cold continuation.
func (e *functionEmitter) planLoopMemoryValue(start int, r memoryRegion) *loopMemoryValue {
	if !e.unmetered || e.sharedDefaultMemory || e.memoryViews || r.loop == nil || r.loop.memory != nil || r.copies != nil || start%256 == 0 || len(r.addresses) != 0 {
		return nil
	}
	cache := loopMemoryValue{load: -1, store: -1}
	constants := map[slots.Slot]uint32{}
	for pc := start; pc <= r.end; pc++ {
		in := e.f.Code[pc]
		src := e.f.Sources[in.Source].Wasm
		a, aok := constants[in.A]
		if isMemoryAccess(in.Op) {
			if !r.proven[pc] || !aok || e.f.Slots[in.A] != wasm.I32 || src.Memory != 0 || src.MemoryOffsetHigh != 0 {
				return nil
			}
			address := uint64(a) + uint64(src.B)
			if address > 0x7fffffff {
				return nil
			}
			if cache.load < 0 && (in.Op == 0x28 || in.Op == 0x29) {
				cache.load = pc
				cache.address = uint32(address)
				cache.width = 4
				if in.Op == 0x29 {
					cache.width = 8
				}
			} else if cache.load >= 0 && cache.store < 0 && (in.Op == 0x36 && cache.width == 4 || in.Op == 0x37 && cache.width == 8) && address == uint64(cache.address) {
				cache.store = pc
			} else {
				return nil
			}
		}
		var v uint32
		known := false
		switch in.Op {
		case 0x41:
			v, known = uint32(src.Bits), true
		case 0x20:
			v, known = constants[slots.Slot(src.A)]
		case 0x21, 0x22:
			v, known = a, aok
			delete(constants, slots.Slot(src.A))
			if known {
				constants[slots.Slot(src.A)] = v
			}
		}
		if in.Dst != slots.NoSlot {
			delete(constants, in.Dst)
			if known {
				constants[in.Dst] = v
			}
		}
	}
	if cache.load < 0 || cache.store < 0 {
		return nil
	}
	if e.a.arch == "arm64" && cache.address > 255 && (cache.address%cache.width != 0 || cache.address/cache.width >= 4096) {
		return nil
	}
	var used [32]bool
	mark := func(plan *registerPlan) {
		if plan != nil {
			for _, reg := range plan.homes {
				used[reg] = true
			}
		}
	}
	mark(e.currentRegisterPlan())
	if r.values != nil {
		for _, plan := range r.values.plans {
			mark(plan)
		}
	}
	for _, reg := range e.constantRegisters {
		used[reg] = true
	}
	for _, reg := range e.runtimeRegisters {
		used[reg] = true
	}
	for _, entry := range e.heapCaches {
		used[entry.key], used[entry.address] = true, true
	}
	pool := append(e.integerRegisterPool(), e.loopAuxiliaryRegisters(start, r)...)
	for _, reg := range pool {
		if !used[reg] {
			cache.register = reg
			return &cache
		}
	}
	return nil
}

func (e *functionEmitter) initializeLoopMemoryValue(cache *loopMemoryValue) {
	base := byte(10)
	if e.a.arch == "arm64" {
		base = 16
	}
	e.heapLoad(cache.width, false, base, cache.address)
	e.registerMove(cache.register, e.operandRegister(0), cache.width == 8)
}

func (e *functionEmitter) loadLoopMemoryValue(in slots.Instruction, cache *loopMemoryValue) {
	if reg, ok := e.registers[in.Dst]; ok {
		e.registerMove(reg, cache.register, cache.width == 8)
		return
	}
	e.registerMove(e.operandRegister(0), cache.register, cache.width == 8)
	e.slotStore(in.Dst, 0, false)
}

func (e *functionEmitter) storeLoopMemoryValue(in slots.Instruction, cache *loopMemoryValue) {
	if reg, ok := e.registers[in.B]; ok {
		e.registerMove(cache.register, reg, cache.width == 8)
		return
	}
	e.slotLoad(in.B, 0, false, cache.width == 8)
	e.registerMove(cache.register, e.operandRegister(0), cache.width == 8)
}
