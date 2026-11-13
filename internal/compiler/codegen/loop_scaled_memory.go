package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopScaledMemory struct {
	input slots.Slot
	mask  uint32
	scale byte
	load  slots.Instruction
}

// An entry guard establishes an unsigned low-bit range. A complete bounded
// transfer through the body must preserve it at the backedge. This is a guarded
// inductive invariant, never an assumption inherited by an arbitrary cold PC.
func (e *functionEmitter) planLoopScaledMemory(start int, r memoryRegion) *loopScaledMemory {
	if !e.unmetered || e.a.arch != "amd64" || r.values != nil || r.copies != nil || r.cache != nil || e.memoryViews || len(r.addresses) != 0 || start+5 > r.end {
		return nil
	}
	get, literal, mask := e.f.Code[start], e.f.Code[start+1], e.f.Code[start+2]
	factor, product, load := e.f.Code[start+3], e.f.Code[start+4], e.f.Code[start+5]
	input := slots.Slot(e.f.Sources[get.Source].Wasm.A)
	if get.Op != 0x20 || get.Dst != input || input >= slots.Slot(e.f.LocalCount) || e.f.Slots[input] != wasm.I32 ||
		literal.Op != 0x41 || mask.Op != 0x71 || mask.A != input || mask.B != literal.Dst || mask.Dst != literal.Dst ||
		factor.Op != 0x41 || product.A != mask.Dst || product.B != factor.Dst || product.Dst != mask.Dst ||
		load.Op < 0x28 || load.Op > 0x35 || load.A != product.Dst || load.Dst != factor.Dst || !r.proven[start+5] {
		return nil
	}
	n := uint32(e.f.Sources[literal.Source].Wasm.Bits)
	scale := r.scales[start+4]
	if n == 0 || n == 0xffffffff || n&(n+1) != 0 || scale == 0 || e.f.Sources[load.Source].Wasm.MemoryOffset() > 0x7fffffff {
		return nil
	}
	var used [16]bool
	for _, reg := range e.registers {
		if reg >= 16 || used[reg] {
			return nil
		}
		used[reg] = true
	}
	if _, resident := e.registers[input]; !resident {
		return nil
	}
	for pc := start; pc <= start+5; pc++ {
		if r.elided[pc] {
			return nil
		}
		if _, remapped := r.addresses[pc]; remapped {
			return nil
		}
	}
	probe := *e
	probe.regionEnd, probe.regionLimit = r.end+1, r.end+1
	probe.provenPCs, probe.guardPCs, probe.elidedPCs, probe.addresses = r.proven, r.guards, r.elided, r.addresses
	probe.regionPredicate = r.predicate
	if !probe.overwrittenBeforeExit(start+6, product.Dst) {
		return nil
	}
	values := map[slots.Slot]boundedValue{input: {max: uint64(n)}}
	for pc := start; pc <= r.end; pc++ {
		updateI32Bounds(e.f, &e.f.Code[pc], values)
	}
	if bound, known := values[input]; !known || bound.max > uint64(n) {
		return nil
	}
	return &loopScaledMemory{input, n, scale, load}
}

func (e *functionEmitter) guardLoopScaledMemory(plan *loopScaledMemory, start int) {
	cold := e.a.label()
	e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, start, e.currentRegisterPlan()})
	// Compare the whole incoming Lo word: a canonical low i32 alone cannot
	// justify using a raw 64-bit register in an x86 effective address.
	e.slotLoad(plan.input, 0, false, true)
	e.constant(1, uint64(plan.mask))
	e.compare(true)
	e.a.branch(cold, true, 7) // Unsigned whole-word index > mask.
}
