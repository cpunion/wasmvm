package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopCountdown struct {
	counter, limit slots.Slot
	first, last    int
	register       byte
}

// A counter used only by its +1 update and unsigned comparison can instead
// retain limit-counter. The original backedge poll precedes the decrement;
// every actual exit reconstructs the original counter before frame publication.
func (e *functionEmitter) planLoopCountdown(start int, r memoryRegion) *loopCountdown {
	p := r.predicate
	if !e.unmetered || start%256 == 0 || r.loop == nil || !r.loop.yes || r.loop.always || r.loop.memory != nil || r.copies != nil || p == nil || p.literal != nil || p.in.Op != 0x49 || p.in.A == p.in.B || p.in.Dst == p.in.A || p.in.Dst == p.in.B {
		return nil
	}
	counter, limit := p.in.A, p.in.B
	if e.f.Slots[counter] != wasm.I32 || e.f.Slots[limit] != wasm.I32 {
		return nil
	}
	selfCopy := func(pc int) bool {
		in := e.f.Code[pc]
		return in.Op == 0x20 && in.Dst == slots.Slot(e.f.Sources[in.Source].Wasm.A)
	}
	last := r.end - 1
	for last >= start && selfCopy(last) {
		last--
	}
	if last-start < 2 {
		return nil
	}
	write, add, literal := e.f.Code[last], e.f.Code[last-1], e.f.Code[last-2]
	if (write.Op != 0x21 && write.Op != 0x22) || slots.Slot(e.f.Sources[write.Source].Wasm.A) != counter || write.A != p.in.Dst || write.Dst != slots.NoSlot && write.Dst != counter || add.Op != 0x6a || add.A != counter || add.B != p.in.Dst || add.Dst != p.in.Dst || literal.Op != 0x41 || literal.Dst != p.in.Dst || uint32(e.f.Sources[literal.Source].Wasm.Bits) != 1 {
		return nil
	}
	first := last - 2
	for pc := start; pc < first; pc++ {
		in := e.f.Code[pc]
		if selfCopy(pc) {
			continue
		}
		if in.A == counter || in.B == counter || in.C == counter || in.Dst == counter || in.Dst == limit {
			return nil
		}
		if in.Op == 0x20 && slots.Slot(e.f.Sources[in.Source].Wasm.A) == counter {
			return nil
		}
		if (in.Op == 0x21 || in.Op == 0x22) && (slots.Slot(e.f.Sources[in.Source].Wasm.A) == counter || slots.Slot(e.f.Sources[in.Source].Wasm.A) == limit) {
			return nil
		}
	}
	plan := e.currentRegisterPlan()
	if plan == nil {
		return nil
	}
	reg, ok := plan.homes[counter]
	bound, bok := plan.homes[limit]
	if !ok || !bok || reg == bound {
		return nil
	}
	valid := func(p *registerPlan) bool {
		if p == nil {
			return false
		}
		if r, ok := p.homes[counter]; !ok || r != reg {
			return false
		}
		if r, ok := p.homes[limit]; !ok || r != bound {
			return false
		}
		for s, r := range p.homes {
			if r == reg && s != counter || r == bound && s != limit || e.a.arch == "amd64" && e.usesMemory && r == 11 {
				return false
			}
		}
		return true
	}
	if !valid(plan) {
		return nil
	}
	if r.values != nil {
		for _, p := range r.values.plans {
			if !valid(p) {
				return nil
			}
		}
	}
	return &loopCountdown{counter, limit, first, last, reg}
}

func (e *functionEmitter) initializeLoopCountdown(c *loopCountdown, start int) {
	cold := e.a.label()
	e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, start, e.currentRegisterPlan()})
	e.slotLoad(c.counter, 0, false, false)
	e.slotLoad(c.limit, 1, false, false)
	e.compare(false)
	e.branch(cold, "hs")
	e.integer(slots.Instruction{Op: 0x6b, Dst: c.counter, A: c.limit, B: c.counter})
}

func (e *functionEmitter) countdownBackedge(c *loopCountdown, target int) {
	if e.a.arch == "arm64" {
		e.a.word(0x71000400 | uint32(c.register)<<5 | uint32(c.register)) // subs wN,wN,#1
	} else {
		if c.register >= 8 {
			e.a.bytes(0x41)
		}
		e.a.bytes(0x83, 0xe8|c.register&7, 1) // sub r32,1
	}
	e.branch(target, "ne")
}

func (e *functionEmitter) restoreLoopCountdown(c *loopCountdown, pending bool) {
	e.integer(slots.Instruction{Op: 0x6b, Dst: c.counter, A: c.limit, B: c.counter})
	if pending {
		e.registerImmediate(0x6a, c.counter, c.counter, 1)
	}
}
