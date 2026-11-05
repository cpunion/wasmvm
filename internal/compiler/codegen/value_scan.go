package codegen

import (
	"cmp"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func (e *functionEmitter) prepareLinearScan(regions map[int]memoryRegion, maxCells int) {
	if !e.unmetered {
		return
	}
	starts := make([]bool, len(e.f.Code))
	starts[0] = true
	remaining := maxCells
	if len(e.f.Handlers)+len(e.f.Prompts) > remaining {
		e.prepareSlotLinearScan(regions, maxCells)
		return
	}
	for _, edge := range e.f.Edges {
		starts[edge.PC] = true
	}
	for _, handler := range e.f.Handlers {
		remaining -= len(handler.Catches)
		if remaining < 0 {
			e.prepareSlotLinearScan(regions, maxCells)
			return
		}
		for _, c := range handler.Catches {
			if uint64(c.PC) >= uint64(len(starts)) {
				e.prepareSlotLinearScan(regions, maxCells)
				return
			}
			starts[c.PC] = true
		}
	}
	for _, catches := range e.f.Prompts {
		remaining -= len(catches)
		if remaining < 0 {
			e.prepareSlotLinearScan(regions, maxCells)
			return
		}
		for _, c := range catches {
			if uint64(c.PC) >= uint64(len(starts)) {
				e.prepareSlotLinearScan(regions, maxCells)
				return
			}
			starts[c.PC] = true
		}
	}
	for pc, in := range e.f.Code {
		if pc%256 == 0 {
			starts[pc] = true
		}
		if !plainOpcode(in.Op) && pc+1 < len(starts) {
			starts[pc+1] = true
		}
	}
	for pc, region := range regions {
		starts[pc] = true
		if region.end+1 < len(starts) {
			starts[region.end+1] = true
		}
	}
	var blocks []int
	for pc, first := range starts {
		if first {
			blocks = append(blocks, pc)
		}
	}
	flow := buildValueFlow(e.f, blocks, regions, instructionFrequencies(e.f), maxCells)
	if flow == nil {
		e.prepareSlotLinearScan(regions, maxCells)
		return
	}
	intervals := flow.intervals()
	// Ordinary cold mappings stay stable around cycles. The private emitter
	// may split the raw value ranges with explicit backedge transfers; arbitrary
	// checked continuations keep the existing stable SCC contract.
	cfg := make([]registerCFGEdge, len(flow.edges))
	for i, edge := range flow.edges {
		cfg[i] = registerCFGEdge{edge.from, edge.to}
	}
	loops := registerLoopComponents(len(flow.blocks), cfg)
	for i := range loops {
		loops[i].first, loops[i].last = flow.blocks[loops[i].first].first, flow.blocks[loops[i].last].end-1
	}
	order := append([]registerInterval(nil), intervals...)
	extendRegisterLoops(order, loops)
	order = mergeValueIntervals(order)
	for _, interval := range order {
		starts[interval.first] = true
		if interval.last+1 < len(starts) {
			starts[interval.last+1] = true
		}
	}
	points := blocks[:0]
	for pc, first := range starts {
		if first {
			points = append(points, pc)
		}
	}
	sortValueIntervals(order)
	e.valueFlow, e.valueIntervals = flow, intervals
	e.installRegisterPlans(order, points, maxCells)
}

func sortValueIntervals(intervals []registerInterval) {
	slices.SortFunc(intervals, func(a, b registerInterval) int {
		if n := cmp.Compare(a.first, b.first); n != 0 {
			return n
		}
		if intervalBefore(a, b) {
			return -1
		}
		if intervalBefore(b, a) {
			return 1
		}
		return 0
	})
}

func valuePlanTransitionCost(from, to *registerPlan) uint64 {
	var cost uint64
	if from != nil {
		for _, s := range from.slots {
			a, b := registerLocation(from, s)
			c, d := registerLocation(to, s)
			if a != c || b != d {
				cost++
			}
		}
	}
	if to != nil {
		for _, s := range to.slots {
			a, b := registerLocation(from, s)
			c, d := registerLocation(to, s)
			if a != c || b != d {
				cost++
			}
		}
	}
	return cost
}

func (e *functionEmitter) valuePlanAccessCost(pc int, plan *registerPlan, r memoryRegion) uint64 {
	if r.elided[pc] {
		return 0
	}
	op := e.valueFlow.ops[pc]
	var cost uint64
	for _, access := range e.valueFlow.accesses[op.first : op.first+op.count] {
		if access.guard {
			if _, ok := r.guards[pc]; !ok {
				continue
			}
		}
		kind, _ := registerLocation(plan, access.slot)
		if kind == 0 && access.slot != slots.NoSlot {
			cost++
		}
	}
	return cost
}
