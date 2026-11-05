package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// Compare the complete ordinary CFG, not just shorter interval lengths. Both
// candidates use the same physical maps and selector boundaries. Transfers
// include spill/reload shuffles as well as frame traffic. This is a bounded
// selection heuristic; actual elapsed acceptance remains a separate gate.
func (e *functionEmitter) valueAllocationCost(plans []*registerPlan, regions map[int]memoryRegion, frequencies []uint64) uint64 {
	defaultPlan := e.currentRegisterPlan()
	at := func(pc int) *registerPlan {
		if plans != nil {
			return plans[pc]
		}
		return defaultPlan
	}
	cost := 2 * valuePlanTransitionCost(defaultPlan, at(0))
	region := memoryRegion{end: -1}
	for pc, in := range e.f.Code {
		if next, ok := regions[pc]; ok {
			region = next
		} else if pc > region.end {
			region = memoryRegion{end: -1}
		}
		weight := frequency(frequencies, pc)
		plan := at(pc)
		cost += weight * e.valuePlanAccessCost(pc, plan, region)
		if pc != 0 {
			op := e.f.Code[pc-1].Op
			if op < slots.ControlBase && op != 8 && op != 10 && op != 0x12 && op != 0x13 && op != 0x15 {
				cost += weight * 2 * valuePlanTransitionCost(at(pc-1), plan)
			}
		}
		if in.Op == slots.ControlBase+uint32(ir.Return) || in.Op < slots.ControlBase && !plainOpcode(in.Op) {
			cost += weight * 2 * valuePlanTransitionCost(plan, nil)
		}
	}
	for _, edge := range e.valueFlow.edges {
		pc := e.valueFlow.blocks[edge.from].end - 1
		to := e.valueFlow.blocks[edge.to].first
		// Fallthrough transfers were counted in the instruction walk.
		if to == pc+1 && !registerBranch(e.f.Code[pc].Op) {
			continue
		}
		cost += frequency(frequencies, pc) * 2 * valuePlanTransitionCost(at(pc), at(to))
	}
	return cost
}
