package codegen

import (
	"cmp"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopValuePlans struct {
	first int
	plans []*registerPlan
	next  []int
}

func (p *loopValuePlans) at(pc int) *registerPlan { return p.plans[pc-p.first] }

func compareIntervalSlot(interval registerInterval, slot slots.Slot) int {
	return cmp.Compare(interval.slot, slot)
}

func (e *functionEmitter) loopValuePlans(start int, r memoryRegion, exit *registerPlan) *loopValuePlans {
	if e.valueFlow == nil || r.loop == nil {
		return nil
	}
	end := r.end + 1 // Include the branch's actual condition use.
	old := e.currentRegisterPlan()
	if old == nil {
		return nil
	}
	for pc := start; pc <= r.end; pc++ {
		in := e.f.Code[pc]
		if !plainOpcode(in.Op) || scalarCanTrap(in.Op) || isMemoryAccess(in.Op) && !r.proven[pc] {
			return nil
		}
	}
	// Existing equal-copy bundles keep their shared home. Reentering a bundle
	// midway would otherwise need a different live-in member at each split.
	// Ordinary slots still split across definitions in every physical bank.
	shared := e.coalescedRegisterSlots()
	pinned := &registerPlan{homes: map[slots.Slot]byte{}, floats: map[slots.Slot]byte{}, vectors: map[slots.Slot]byte{}, shared: old.shared}
	var integerUsed, floatUsed [32]bool
	for _, s := range old.slots {
		if !shared[s] {
			continue
		}
		kind, reg := registerLocation(old, s)
		setRegisterHome(pinned, s, kind, reg)
		pinned.slots = append(pinned.slots, s)
		if kind == 1 {
			integerUsed[reg] = true
		} else {
			floatUsed[reg] = true
		}
	}
	accesses := map[slots.Slot][]int{}
	for pc := start; pc <= end; pc++ {
		if r.elided[pc] {
			continue
		}
		op := e.valueFlow.ops[pc]
		for _, access := range e.valueFlow.accesses[op.first : op.first+op.count] {
			if access.guard {
				if _, ok := r.guards[pc]; !ok {
					continue
				}
			}
			if !shared[access.slot] {
				accesses[access.slot] = append(accesses[access.slot], pc)
			}
		}
	}
	var order []registerInterval
	for slot, uses := range accesses {
		first, _ := slices.BinarySearchFunc(e.valueIntervals, slot, compareIntervalSlot)
		after, _ := slices.BinarySearchFunc(e.valueIntervals, slot+1, compareIntervalSlot)
		intervals := e.valueIntervals[first:after]
		position, _ := slices.BinarySearchFunc(intervals, start, func(v registerInterval, pc int) int { return cmp.Compare(v.last, pc) })
		for _, interval := range intervals[position:] {
			if interval.first > end {
				break
			}
			interval.first, interval.last = max(start, interval.first), min(end, interval.last)
			lo, _ := slices.BinarySearch(uses, interval.first)
			hi, _ := slices.BinarySearch(uses, interval.last+1)
			interval.priority = uint64(hi - lo)
			if interval.priority != 0 {
				order = append(order, interval)
			}
		}
	}
	sortValueIntervals(order)
	integers := scanBank{}
	for _, reg := range append(e.integerRegisterPool(), e.loopAuxiliaryRegisters(start, r)...) {
		if !integerUsed[reg] {
			integers.pool = append(integers.pool, reg)
		}
	}
	floats := scanBank{}
	for _, reg := range e.floatRegisterPool() {
		if !floatUsed[reg] {
			floats.pool = append(floats.pool, reg)
		}
	}
	result := &loopValuePlans{first: start, plans: make([]*registerPlan, end-start+1)}
	position := 0
	remaining := len(result.plans) * 128
	var previous *registerPlan
	for pc := start; pc <= end; pc++ {
		var inInteger, inFloat []registerInterval
		for position < len(order) && order[position].first == pc {
			interval := order[position]
			kind := e.f.Slots[interval.slot]
			if kind == wasm.F32 || kind == wasm.F64 || kind == wasm.V128 {
				inFloat = append(inFloat, interval)
			} else {
				inInteger = append(inInteger, interval)
			}
			position++
		}
		if pc == start {
			inInteger = integers.seed(old, inInteger, false)
			inFloat = floats.seed(old, inFloat, true)
		}
		integers.advance(pc, inInteger)
		floats.advance(pc, inFloat)
		plan := registerBankPlan(e.f, &integers, &floats, pinned)
		if equalRegisterPlans(plan, previous) {
			plan = previous
		} else {
			remaining -= len(plan.slots) + 1
			if remaining < 0 {
				return nil
			}
		}
		result.plans[pc-start], previous = plan, plan
	}
	// Include all mapping transfers, the taken backedge and the ordinary exit.
	// A cheaper analytical frame-access count is a selection heuristic only;
	// elapsed competitive acceptance still requires the actual benchmark.
	baseline := valuePlanTransitionCost(old, exit)
	cost := valuePlanTransitionCost(old, result.at(start)) + valuePlanTransitionCost(result.at(end), exit) + valuePlanTransitionCost(result.at(end), result.at(start))
	previous = result.at(start)
	for pc := start; pc <= end; pc++ {
		plan := result.at(pc)
		cost += valuePlanTransitionCost(previous, plan) + e.valuePlanAccessCost(pc, plan, r)
		baseline += e.valuePlanAccessCost(pc, old, r)
		previous = plan
	}
	if cost >= baseline {
		return nil
	}
	result.next = make([]int, len(result.plans))
	boundary := end + 1
	for pc := end; pc >= start; pc-- {
		if pc < end && result.at(pc) != result.at(pc+1) {
			boundary = pc + 1
		}
		result.next[pc-start] = boundary
	}
	return result
}
