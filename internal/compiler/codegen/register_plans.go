package codegen

import (
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Immutable mappings describe each native program point, including late exits.
// All value classes reach their canonical frame slots at driver boundaries.
// Immutable constants and runtime metadata retain function-wide registers.
type registerPlan struct {
	homes   map[slots.Slot]byte
	floats  map[slots.Slot]byte
	vectors map[slots.Slot]byte
	raw     map[slots.Slot]byte
	slots   []slots.Slot
	shared  bool
}

func (e *functionEmitter) currentRegisterPlan() *registerPlan {
	if e.registerPlan == nil && (len(e.registers) != 0 || len(e.registerSlots) != 0) {
		e.registerPlan = &registerPlan{homes: e.registers, floats: e.floatRegisters, vectors: e.vectorRegisters, raw: e.rawRegisters, slots: e.registerSlots}
	}
	return e.registerPlan
}
func (e *functionEmitter) useRegisterPlan(plan *registerPlan) {
	e.registerPlan = plan
	if plan == nil {
		e.registers = nil
		e.floatRegisters, e.vectorRegisters = nil, nil
		e.rawRegisters = nil
		e.registerSlots = nil
		return
	}
	e.registers, e.registerSlots = plan.homes, plan.slots
	e.floatRegisters, e.vectorRegisters = plan.floats, plan.vectors
	e.rawRegisters = plan.raw
}

// Spill all displaced words before loading any new home. Equal homes keep
// their complete current value; no transition clobbers a still-needed source.
func registerLocation(p *registerPlan, s slots.Slot) (byte, byte) {
	if p != nil {
		if r, ok := p.homes[s]; ok {
			return 1, r
		}
		if r, ok := p.floats[s]; ok {
			return 2, r
		}
		if r, ok := p.vectors[s]; ok {
			return 3, r
		}
		if r, ok := p.raw[s]; ok {
			return 4, r
		}
	}
	return 0, 0
}

func (e *functionEmitter) transitionRegisters(next *registerPlan) {
	previous := e.currentRegisterPlan()
	if previous == next {
		return
	}
	clear(e.vectorAliases)
	e.registers, e.floatRegisters, e.vectorRegisters = nil, nil, nil
	e.rawRegisters = nil
	if previous != nil {
		for _, s := range previous.slots {
			kind, r := registerLocation(previous, s)
			nextKind, nextReg := registerLocation(next, s)
			if kind == nextKind && r == nextReg {
				continue
			}
			switch kind {
			case 1:
				e.registerMove(e.operandRegister(0), r, true)
				e.slotStore(s, 0, false)
			case 2:
				e.floatToInteger(e.operandRegister(0), r, e.f.Slots[s] == wasm.F64)
				e.slotStore(s, 0, false)
			case 3:
				e.vectorMemory(s, r, true)
			case 4:
				e.floatToInteger(e.operandRegister(0), r, true)
				e.slotStore(s, 0, false)
			}
		}
	}
	if next != nil {
		for _, s := range next.slots {
			kind, r := registerLocation(next, s)
			oldKind, oldReg := registerLocation(previous, s)
			if kind == oldKind && r == oldReg {
				continue
			}
			switch kind {
			case 1:
				e.slotLoad(s, 0, false, true)
				e.registerMove(r, e.operandRegister(0), true)
			case 2:
				e.slotLoad(s, 0, false, e.f.Slots[s] == wasm.F64)
				e.floatFromInteger(r, e.operandRegister(0), e.f.Slots[s] == wasm.F64)
			case 3:
				e.vectorMemory(s, r, false)
			case 4:
				e.slotLoad(s, 0, false, true)
				e.floatFromInteger(r, e.operandRegister(0), true)
			}
		}
	}
	e.useRegisterPlan(next)
}

// Bounded loop-local priorities reuse the ordinary slot-use model. Keeping
// common homes avoids register permutations; only genuinely displaced homes
// cross the canonical frame. Compilation work scales with the region, not
// with the number of unused slots or the whole function for each loop.
func (e *functionEmitter) loopRegisterPlan(start int, r memoryRegion) *registerPlan {
	old := e.currentRegisterPlan()
	if old == nil || r.loop == nil && r.unroll == nil {
		return nil
	}
	extra := e.loopAuxiliaryRegisters(start, r)
	var proof *memoryLoop
	if r.loop != nil {
		proof = r.loop.memory
	}
	if proof == nil && len(extra) == 0 {
		return nil
	}
	weights := map[slots.Slot]uint64{}
	collectSlotPriorities(e.f, map[int]memoryRegion{start: r}, nil, start, r.end+2, false, func(s slots.Slot, w uint64) {
		if s != slots.NoSlot && e.f.Slots[s] != wasm.F32 && e.f.Slots[s] != wasm.F64 && e.f.Slots[s] != wasm.V128 {
			weights[s] += w
		}
	})
	candidates := make([]registerCandidate, 0, len(weights))
	for s, w := range weights {
		if w != 0 {
			candidates = append(candidates, registerCandidate{key: uint64(s), weight: w})
		}
	}
	slices.SortFunc(candidates, compareRegisterCandidates)
	pool := append(e.integerRegisterPool(), extra...)
	if len(candidates) > len(pool) {
		candidates = candidates[:len(pool)]
	}
	selected := map[slots.Slot]bool{}
	var before, after uint64
	for s, w := range weights {
		if _, ok := old.homes[s]; ok {
			before += w
		}
	}
	for _, c := range candidates {
		selected[slots.Slot(c.key)] = true
		after += c.weight
	}
	if after <= before {
		return nil
	}
	homes := map[slots.Slot]byte{}
	var used [32]bool
	for _, c := range candidates {
		s := slots.Slot(c.key)
		if reg, ok := old.homes[s]; ok {
			homes[s] = reg
			used[reg] = true
		}
	}
	for _, c := range candidates {
		s := slots.Slot(c.key)
		if _, ok := homes[s]; ok {
			continue
		}
		for _, reg := range pool {
			if !used[reg] {
				homes[s] = reg
				used[reg] = true
				break
			}
		}
	}
	// Spare homes retain other function values for free when possible.
	for _, s := range old.slots {
		if reg, ok := old.homes[s]; ok && !used[reg] {
			homes[s] = reg
			used[reg] = true
		}
	}
	var transfers uint64
	for s, reg := range old.homes {
		other, ok := homes[s]
		if !ok || other != reg {
			transfers++
		}
	}
	for s, reg := range homes {
		other, ok := old.homes[s]
		if !ok || other != reg {
			transfers++
		}
	}
	iterations := uint64(1)
	if r.unroll != nil {
		iterations = uint64(r.unroll.copies)
	} else if proof != nil {
		iterations = uint64(proof.limit) / uint64(proof.stride)
	}
	if iterations*(after-before) <= 2*transfers {
		return nil
	}
	plan := &registerPlan{homes: homes, floats: old.floats, vectors: old.vectors}
	for _, s := range old.slots {
		if _, ok := old.homes[s]; !ok {
			plan.slots = append(plan.slots, s)
		}
	}
	for _, c := range candidates {
		plan.slots = append(plan.slots, slots.Slot(c.key))
	}
	for _, s := range old.slots {
		if _, ok := homes[s]; ok && !selected[s] {
			plan.slots = append(plan.slots, s)
		}
	}
	return plan
}
