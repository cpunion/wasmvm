package codegen

import (
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Spare FP/SIMD registers retain complete integer Lo words as bits, with no
// numeric conversion. Fixed private mappings need only entry/exit transfers;
// the existing variable plans keep their independently allocated typed banks.
func (e *functionEmitter) loopRawRegisterPlan(v *loopValues, r memoryRegion, proof *memoryLoop) *registerPlan {
	old := e.currentRegisterPlan()
	if v == nil || old == nil || r.values != nil {
		return nil
	}
	var used [32]bool
	for _, bank := range []map[slots.Slot]byte{old.floats, old.vectors, old.raw} {
		for _, reg := range bank {
			used[reg] = true
		}
	}
	for _, reg := range e.vectorConstants {
		used[reg] = true
	}
	var free []byte
	for _, reg := range e.floatRegisterPool() {
		if !used[reg] {
			free = append(free, reg)
		}
	}
	if len(free) == 0 {
		return nil
	}
	eligible := func(s slots.Slot) bool {
		if s == slots.NoSlot || e.f.Slots[s] != wasm.I32 && e.f.Slots[s] != wasm.I64 || len(v.bySlot[s]) == 0 || r.copies != nil && r.copies.deferred[s] {
			return false
		}
		kind, _ := registerLocation(old, s)
		return kind == 0
	}
	needed := false
	for s := range v.bySlot {
		needed = needed || eligible(s)
	}
	if !needed {
		return nil
	}
	weights := map[slots.Slot]uint64{}
	collectSlotPriorities(e.f, map[int]memoryRegion{v.first: r}, nil, v.first, r.end+2, false, func(s slots.Slot, w uint64) {
		if eligible(s) {
			weights[s] += w
		}
	})
	iterations := uint64(1)
	if proof != nil {
		iterations = uint64(proof.limit) / uint64(proof.stride)
	}
	var candidates []registerCandidate
	for s, weight := range weights {
		// A bounded ranking heuristic, not an elapsed-cost claim. Include
		// entry/exit transfers rather than caching a single cold reference.
		if weight*iterations > 4 {
			candidates = append(candidates, registerCandidate{key: uint64(s), weight: weight})
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	slices.SortFunc(candidates, compareRegisterCandidates)
	candidates = candidates[:min(len(candidates), len(free))]
	plan := cloneRegisterPlan(old)
	for i, c := range candidates {
		s := slots.Slot(c.key)
		plan.raw[s] = free[i]
		plan.slots = append(plan.slots, s)
	}
	return plan
}
