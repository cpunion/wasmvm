package codegen

import (
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// A residency interval includes every read and write of a slot. Unlike semantic
// liveness it never permits discarding an old frame word: leaving a location
// always stores that word, including values observable through a cold entry.
type registerInterval struct {
	slot        slots.Slot
	first, last int
	priority    uint64
}

type intervalQueue []registerInterval

func intervalBefore(a, b registerInterval) bool {
	if a.priority != b.priority {
		return a.priority > b.priority
	}
	if a.last != b.last {
		return a.last < b.last
	}
	return a.slot < b.slot
}

// Typed heap operations avoid interface boxes for compiler-owned intervals.
func (q *intervalQueue) push(v registerInterval) {
	*q = append(*q, v)
	for child := len(*q) - 1; child > 0; {
		parent := (child - 1) / 2
		if !intervalBefore((*q)[child], (*q)[parent]) {
			break
		}
		(*q)[parent], (*q)[child] = (*q)[child], (*q)[parent]
		child = parent
	}
}
func (q *intervalQueue) pop() registerInterval {
	a := *q
	v := a[0]
	a[0] = a[len(a)-1]
	a = a[:len(a)-1]
	for parent := 0; parent*2+1 < len(a); {
		child := parent*2 + 1
		if child+1 < len(a) && intervalBefore(a[child+1], a[child]) {
			child++
		}
		if !intervalBefore(a[child], a[parent]) {
			break
		}
		a[parent], a[child] = a[child], a[parent]
		parent = child
	}
	*q = a
	return v
}

type registerResident struct {
	interval registerInterval
	valid    bool
}

// Each bank shares the target inventory with the existing constrained-register
// lowering. Expiration and cost-based eviction split residency at block
// boundaries. Spilled intervals can reenter when a register becomes available.
type scanBank struct {
	pool     []byte
	resident []registerResident
	pending  intervalQueue
}

func (b *scanBank) advance(block int, incoming []registerInterval) {
	if b.resident == nil {
		b.resident = make([]registerResident, len(b.pool))
	}
	for i, r := range b.resident {
		if r.valid && r.interval.last < block {
			b.resident[i].valid = false
		}
	}
	for _, interval := range incoming {
		b.pending.push(interval)
	}
	for len(b.pending) != 0 {
		candidate := b.pending[0]
		if candidate.last < block {
			b.pending.pop()
			continue
		}
		free, worst := -1, -1
		for i, r := range b.resident {
			if !r.valid {
				free = i
				break
			}
			if worst < 0 || intervalBefore(b.resident[worst].interval, r.interval) {
				worst = i
			}
		}
		if free < 0 {
			if worst < 0 || !intervalBefore(candidate, b.resident[worst].interval) {
				break
			}
			free = worst
		}
		b.pending.pop()
		if b.resident[free].valid {
			b.pending.push(b.resident[free].interval)
		}
		b.resident[free] = registerResident{candidate, true}
	}
}

func (e *functionEmitter) prepareLinearScan(regions map[int]memoryRegion, maxCells int) {
	if !e.unmetered {
		return
	}
	starts := make([]bool, len(e.f.Code))
	starts[0] = true
	for _, edge := range e.f.Edges {
		starts[edge.PC] = true
	}
	for pc, in := range e.f.Code {
		if pc%256 == 0 {
			starts[pc] = true
		}
		if !plainOpcode(in.Op) && pc+1 < len(starts) {
			starts[pc+1] = true
		}
	}
	for start, r := range regions {
		starts[start] = true
		if r.end+1 < len(starts) {
			starts[r.end+1] = true
		}
	}
	var blocks []int
	for pc, start := range starts {
		if start {
			blocks = append(blocks, pc)
		}
	}
	if len(blocks) < 2 || len(blocks) > maxCells {
		return
	}
	var cfg []registerCFGEdge
	for block := 0; block+1 < len(blocks); block++ {
		op := e.f.Code[blocks[block+1]-1].Op
		if op < slots.ControlBase && op != 8 && op != 10 && op != 0x12 && op != 0x13 && op != 0x15 {
			cfg = append(cfg, registerCFGEdge{block, block + 1})
		}
	}
	if len(cfg)+len(blocks) > maxCells {
		return
	}
	intervals := make([]registerInterval, len(e.f.Slots))
	for s := range intervals {
		intervals[s] = registerInterval{slot: slots.Slot(s), first: -1, last: -1}
		if len(e.registerPriorities) != 0 {
			intervals[s].priority = e.registerPriorities[s]
		}
	}
	// Repeated tuple/edge lists cannot multiply compilation work without bound.
	// Exhaustion retains the existing conservative function-wide allocator.
	work := uint64(len(e.f.Code)+len(e.f.Args)+len(e.f.Moves)+len(e.f.Edges)) * 8
	block := 0
	for pc, in := range e.f.Code {
		if block+1 < len(blocks) && pc == blocks[block+1] {
			block++
		}
		use := func(s slots.Slot) {
			if s == slots.NoSlot {
				return
			}
			v := &intervals[s]
			if v.first < 0 {
				v.first = block
			}
			v.last = block
		}
		use(in.A)
		use(in.B)
		use(in.C)
		use(in.Dst)
		if in.Op >= 0x20 && in.Op <= 0x22 {
			use(slots.Slot(e.f.Sources[in.Source].Wasm.A))
		}
		if in.Op == 8 || in.Op == 10 || in.Op >= 0x10 && in.Op <= 0x15 || in.Op>>16 == 0xfb || in.Op >= 0xe0 && in.Op <= 0xe6 || in.Op == slots.ControlBase+uint32(ir.Return) || module.IsWideArithmetic(in.Op) {
			count := uint64(in.Count)
			if in.Op == 0x10 || in.Op == 0x11 || in.Op == 0x14 || in.Op>>16 == 0xfb || in.Op >= 0xe0 && in.Op <= 0xe6 || module.IsWideArithmetic(in.Op) {
				count += uint64(len(e.f.Sources[in.Source].Outputs))
			}
			if count > work {
				return
			}
			work -= count
			for _, s := range e.f.Args[uint64(in.Extra) : uint64(in.Extra)+count] {
				use(s)
			}
		} else if registerBranch(in.Op) {
			if uint64(in.Count) > work {
				return
			}
			work -= uint64(in.Count)
			if uint64(len(cfg))+uint64(len(blocks))+uint64(in.Count) > uint64(maxCells) {
				return
			}
			for _, edge := range e.f.Edges[in.Extra : in.Extra+in.Count] {
				cfg = append(cfg, registerCFGEdge{block, sort.SearchInts(blocks, int(edge.PC))})
				if uint64(edge.Count) > work {
					return
				}
				work -= uint64(edge.Count)
				for _, m := range e.f.Moves[edge.Start : edge.Start+edge.Count] {
					use(m.Src)
					use(m.Dst)
				}
			}
		}
	}
	extendRegisterLoops(intervals, registerLoopComponents(len(blocks), cfg))
	var order []registerInterval
	for _, interval := range intervals {
		if interval.first >= 0 {
			order = append(order, interval)
		}
	}
	sort.Slice(order, func(i, j int) bool {
		if order[i].first != order[j].first {
			return order[i].first < order[j].first
		}
		return intervalBefore(order[i], order[j])
	})
	integers := scanBank{pool: e.integerRegisterPool()}
	floats := scanBank{pool: []byte{3, 4, 5, 6, 7, 8, 9, 10}}
	if e.a.arch == "arm64" {
		floats.pool = []byte{8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	}
	baseline := e.currentRegisterPlan()
	plans := make([]*registerPlan, len(e.f.Code))
	var previous *registerPlan
	position := 0
	useful := false
	// Bound materialized register-map cells independently of instruction count.
	// Large hostile CFGs retain conservative allocation instead of amplifying
	// the input into one map per register bank per instruction.
	remainingCells := maxCells
	for block, start := range blocks {
		var inInteger, inFloat []registerInterval
		for position < len(order) && order[position].first == block {
			interval := order[position]
			t := e.f.Slots[interval.slot]
			if t == wasm.F32 || t == wasm.F64 || t == wasm.V128 {
				inFloat = append(inFloat, interval)
			} else {
				inInteger = append(inInteger, interval)
			}
			position++
		}
		integers.advance(block, inInteger)
		floats.advance(block, inFloat)
		plan := &registerPlan{homes: map[slots.Slot]byte{}, floats: map[slots.Slot]byte{}, vectors: map[slots.Slot]byte{}}
		for i, r := range integers.resident {
			if r.valid {
				plan.homes[r.interval.slot] = integers.pool[i]
				plan.slots = append(plan.slots, r.interval.slot)
				if baseline == nil {
					useful = true
				} else if _, ok := baseline.homes[r.interval.slot]; !ok {
					useful = true
				}
			}
		}
		for i, r := range floats.resident {
			if r.valid {
				s := r.interval.slot
				if e.f.Slots[s] == wasm.V128 {
					plan.vectors[s] = floats.pool[i]
				} else {
					plan.floats[s] = floats.pool[i]
				}
				plan.slots = append(plan.slots, s)
				if baseline == nil {
					useful = true
				} else if _, f := baseline.floats[s]; !f {
					if _, v := baseline.vectors[s]; !v {
						useful = true
					}
				}
			}
		}
		sort.Slice(plan.slots, func(i, j int) bool { return plan.slots[i] < plan.slots[j] })
		if equalRegisterPlans(previous, plan) {
			plan = previous
		} else {
			remainingCells -= len(plan.slots) + 1
			if remainingCells < 0 {
				return
			}
		}
		end := len(plans)
		if block+1 < len(blocks) {
			end = blocks[block+1]
		}
		for pc := start; pc < end; pc++ {
			plans[pc] = plan
		}
		previous = plan
	}
	if useful {
		e.allocationPlans, e.allocationStarts = plans, blocks
	}
}

func equalRegisterPlans(a, b *registerPlan) bool {
	if a == nil || b == nil {
		return a == b
	}
	equal := func(a, b map[slots.Slot]byte) bool {
		if len(a) != len(b) {
			return false
		}
		for s, r := range a {
			if other, ok := b[s]; !ok || r != other {
				return false
			}
		}
		return true
	}
	return equal(a.homes, b.homes) && equal(a.floats, b.floats) && equal(a.vectors, b.vectors)
}

type allocationEdge struct {
	label, target int
	from, to      *registerPlan
}

// All branch paths carry their actual source mapping. The destination label is
// after its fallthrough transition, so both paths arrive with the same words.
func (e *functionEmitter) targetPC(pcs []int, pc int) int {
	if e.allocationPlans == nil || e.currentRegisterPlan() == e.allocationPlans[pc] {
		return pcs[pc]
	}
	label := e.a.label()
	e.allocationEdges = append(e.allocationEdges, allocationEdge{label, pcs[pc], e.currentRegisterPlan(), e.allocationPlans[pc]})
	return label
}

func (e *functionEmitter) emitAllocationEdges() {
	previous := e.currentRegisterPlan()
	for _, edge := range e.allocationEdges {
		e.a.mark(edge.label)
		e.useRegisterPlan(edge.from)
		e.transitionRegisters(edge.to)
		e.a.branch(edge.target, false, 0)
	}
	e.useRegisterPlan(previous)
}

// One entry transition per distinct plan serves every original PC in its
// block. The second shared dispatch retains arbitrary continuation identity
// without copying a complete spill/reload sequence for every instruction.
func (e *functionEmitter) emitAllocationEntries(pcs []int) {
	if len(e.allocationEntries) == 0 {
		return
	}
	previous := e.currentRegisterPlan()
	dispatch := e.a.label()
	for _, entry := range e.allocationEntries {
		e.a.mark(entry.label)
		e.useRegisterPlan(entry.from)
		e.transitionRegisters(entry.to)
		e.contextLoad(0, nativeabi.ContinuationOffset)
		e.a.branch(dispatch, false, 0)
	}
	e.a.mark(dispatch)
	e.a.dispatch(pcs)
	e.useRegisterPlan(previous)
}
