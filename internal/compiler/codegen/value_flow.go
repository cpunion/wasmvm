package codegen

import (
	"cmp"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
)

type valueFlowNode struct {
	slot               slots.Slot
	first, last, block int
	inputs, count      int
	weight             uint64
	phi, opaque        bool
}

type valueFlowAccess struct {
	slot  slots.Slot
	write bool
	guard bool
}
type valueFlowOperation struct{ first, count int }
type valueFlowBlock struct {
	first, end int
	current    map[slots.Slot]int
	entry      map[slots.Slot]int
	pred       []int
	opaque     bool
}
type valueFlowEdge struct {
	from, to int
	values   map[slots.Slot]int
}
type valueFlow struct {
	nodes     []valueFlowNode
	inputs    []int
	accesses  []valueFlowAccess
	ops       []valueFlowOperation
	blocks    []valueFlowBlock
	edges     []valueFlowEdge
	pending   []int
	remaining int
	tuples    uint64
}
type valueFlowLimit struct{}

func (g *valueFlow) charge(n int) {
	if n > g.remaining {
		panic(valueFlowLimit{})
	}
	g.remaining -= n
}

func (g *valueFlow) tuple(n uint64) {
	if n > g.tuples {
		panic(valueFlowLimit{})
	}
	g.tuples -= n
}

func (g *valueFlow) node(s slots.Slot, block, pc, inputs, count int) int {
	g.charge(1)
	id := len(g.nodes)
	g.nodes = append(g.nodes, valueFlowNode{slot: s, block: block, first: pc, last: pc, inputs: inputs, count: count})
	return id
}

// Allocate the entry identity before following predecessors. A flat pending
// list resolves even irreducible cycles without recursion or a fixed-point
// sweep over the Cartesian product of all blocks and all frame slots.
func (g *valueFlow) entry(block int, s slots.Slot) int {
	if id, ok := g.blocks[block].entry[s]; ok {
		return id
	}
	g.charge(2)
	if g.blocks[block].entry == nil {
		g.blocks[block].entry = map[slots.Slot]int{}
	}
	id := g.node(s, block, g.blocks[block].first, 0, 0)
	g.nodes[id].phi = true
	g.blocks[block].entry[s] = id
	g.pending = append(g.pending, id)
	return id
}

func (g *valueFlow) endValue(block int, s slots.Slot) int {
	if id, ok := g.blocks[block].current[s]; ok {
		return id
	}
	return g.entry(block, s)
}

func (g *valueFlow) touch(id, pc int, weight uint64) {
	g.nodes[id].last = max(g.nodes[id].last, pc)
	g.nodes[id].weight += weight
}

func (g *valueFlow) read(block int, s slots.Slot, pc int, weight uint64) {
	if s == slots.NoSlot {
		return
	}
	g.charge(2)
	id := g.endValue(block, s)
	g.touch(id, pc, weight)
	g.inputs = append(g.inputs, id)
	g.accesses = append(g.accesses, valueFlowAccess{slot: s})
}

func (g *valueFlow) write(block int, s slots.Slot, pc, inputs, count int, weight uint64) {
	if s == slots.NoSlot {
		return
	}
	g.charge(1)
	if g.blocks[block].current == nil {
		g.blocks[block].current = map[slots.Slot]int{}
	}
	if _, ok := g.blocks[block].current[s]; !ok {
		g.charge(1)
	}
	id := g.node(s, block, pc, inputs, count)
	g.nodes[id].weight = weight
	g.blocks[block].current[s] = id
	g.accesses = append(g.accesses, valueFlowAccess{slot: s, write: true})
}

func (g *valueFlow) edge(from, to int) int {
	g.charge(2)
	id := len(g.edges)
	g.edges = append(g.edges, valueFlowEdge{from: from, to: to})
	g.blocks[to].pred = append(g.blocks[to].pred, id)
	return id
}

// Slot lowering has already scheduled parallel copies, including scratch
// words for cycles. Consume that ordered program, rather than interpreting the
// moves again as an unordered parallel assignment.
func (g *valueFlow) moves(edge int, moves []slots.Move, pc int, weight uint64) {
	if len(moves) == 0 {
		return
	}
	from := g.edges[edge].from
	g.edges[edge].values = map[slots.Slot]int{}
	for _, move := range moves {
		g.charge(4)
		src, found := g.edges[edge].values[move.Src]
		if !found {
			src = g.endValue(from, move.Src)
		}
		g.touch(src, pc, weight)
		start := len(g.inputs)
		g.inputs = append(g.inputs, src)
		dst := g.node(move.Dst, from, pc, start, 1)
		g.nodes[dst].weight = weight
		g.edges[edge].values[move.Dst] = dst
		g.accesses = append(g.accesses, valueFlowAccess{slot: move.Src}, valueFlowAccess{slot: move.Dst, write: true})
	}
}

// These facts guide residency, not semantic deletion or cold-entry aliases.
// Every physical mapping remains one-to-one and all displaced frame words are
// stored. Opaque driver entries therefore need no speculative equality guard.
func buildValueFlow(f *slots.Function, starts []int, regions map[int]memoryRegion, frequencies []uint64, cells int) (result *valueFlow) {
	defer func() {
		if p := recover(); p != nil {
			if _, ok := p.(valueFlowLimit); !ok {
				panic(p)
			}
			result = nil
		}
	}()
	g := &valueFlow{remaining: cells, tuples: uint64(len(f.Code)+len(f.Args)+len(f.Moves)+len(f.Edges)) * 8}
	g.charge(len(starts) + len(f.Code))
	g.ops = make([]valueFlowOperation, len(f.Code))
	g.blocks = make([]valueFlowBlock, len(starts))
	for block, first := range starts {
		end := len(f.Code)
		if block+1 < len(starts) {
			end = starts[block+1]
		}
		g.blocks[block] = valueFlowBlock{first: first, end: end, opaque: block == 0}
	}
	blockAt := func(pc int) int {
		block, found := slices.BinarySearch(starts, pc)
		if !found {
			panic(valueFlowLimit{})
		}
		return block
	}
	for _, handler := range f.Handlers {
		g.charge(len(handler.Catches))
		for _, c := range handler.Catches {
			g.blocks[blockAt(int(c.PC))].opaque = true
		}
	}
	for _, catches := range f.Prompts {
		g.charge(len(catches))
		for _, c := range catches {
			g.blocks[blockAt(int(c.PC))].opaque = true
		}
	}
	region := memoryRegion{end: -1}
	for block, b := range g.blocks {
		for pc := b.first; pc < b.end; pc++ {
			if next, ok := regions[pc]; ok {
				region = next
			} else if pc > region.end {
				region = memoryRegion{end: -1}
			}
			g.ops[pc].first = len(g.accesses)
			if region.elided[pc] {
				continue
			}
			in := f.Code[pc]
			if address, ok := region.addresses[pc]; ok {
				in.A = address.local
			}
			weight := frequency(frequencies, pc)
			input := len(g.inputs)
			if guard, ok := region.guards[pc]; ok {
				g.read(block, guard.address, pc, weight)
				if guard.address != slots.NoSlot {
					g.accesses[len(g.accesses)-1].guard = true
				}
			}
			tuple := in.Op == 8 || in.Op == 10 || in.Op >= 0x10 && in.Op <= 0x15 || in.Op>>16 == 0xfb ||
				in.Op >= 0xe0 && in.Op <= 0xe6 || in.Op == slots.ControlBase+uint32(ir.Return) || module.IsWideArithmetic(in.Op)
			outputs := tuple && (in.Op == 0x10 || in.Op == 0x11 || in.Op == 0x14 || in.Op>>16 == 0xfb || in.Op >= 0xe0 && in.Op <= 0xe6 || module.IsWideArithmetic(in.Op))
			if tuple {
				g.tuple(uint64(in.Count))
				g.charge(int(in.Count))
				for _, s := range f.Args[in.Extra : in.Extra+in.Count] {
					g.read(block, s, pc, weight)
				}
				// Indirect calls keep their dynamic table/reference operand inline.
				if in.Op >= 0x10 && in.Op <= 0x15 {
					g.read(block, in.A, pc, weight)
					g.read(block, in.B, pc, weight)
					g.read(block, in.C, pc, weight)
				}
			} else {
				g.read(block, in.A, pc, weight)
				g.read(block, in.B, pc, weight)
				g.read(block, in.C, pc, weight)
			}
			if in.Op == 0x20 {
				g.read(block, slots.Slot(f.Sources[in.Source].Wasm.A), pc, weight)
			}
			count := len(g.inputs) - input
			if outputs {
				n := uint32(len(f.Sources[in.Source].Outputs))
				g.tuple(uint64(n))
				g.charge(int(n))
				for _, s := range f.Args[in.Extra+in.Count : in.Extra+in.Count+n] {
					g.write(block, s, pc, input, count, weight)
				}
			} else {
				g.write(block, in.Dst, pc, input, count, weight)
			}
			if in.Op == 0x21 || in.Op == 0x22 {
				g.write(block, slots.Slot(f.Sources[in.Source].Wasm.A), pc, input, count, weight)
			}
			if registerBranch(in.Op) {
				g.tuple(uint64(in.Count))
				g.charge(int(in.Count))
				for _, edge := range f.Edges[in.Extra : in.Extra+in.Count] {
					id := g.edge(block, blockAt(int(edge.PC)))
					g.tuple(uint64(edge.Count))
					g.charge(int(edge.Count))
					g.moves(id, f.Moves[edge.Start:edge.Start+edge.Count], pc, weight)
				}
			}
			g.ops[pc].count = len(g.accesses) - g.ops[pc].first
		}
		op := f.Code[b.end-1].Op
		if block+1 < len(g.blocks) && op < slots.ControlBase && op != 8 && op != 10 && op != 0x12 && op != 0x13 && op != 0x15 {
			g.edge(block, block+1)
		}
	}
	for next := 0; next < len(g.pending); next++ {
		id := g.pending[next]
		node := g.nodes[id]
		block := g.blocks[node.block]
		g.nodes[id].opaque = block.opaque || len(block.pred) == 0
		first := len(g.inputs)
		for _, edge := range block.pred {
			g.charge(1)
			e := g.edges[edge]
			value, ok := e.values[node.slot]
			if !ok {
				value = g.endValue(e.from, node.slot)
			}
			g.touch(value, g.blocks[e.from].end-1, 0)
			g.inputs = append(g.inputs, value)
		}
		g.nodes[id].inputs, g.nodes[id].count = first, len(g.inputs)-first
	}
	return g
}

// A value may stop occupying a register after its last access, but the mapping
// transition still stores its word. Merge overlapping versions of one physical
// slot so the register banks never install two owners for that same slot.
func (g *valueFlow) intervals() []registerInterval {
	intervals := make([]registerInterval, 0, len(g.nodes))
	for _, node := range g.nodes {
		intervals = append(intervals, registerInterval{slot: node.slot, first: node.first, last: node.last, priority: node.weight})
	}
	return mergeValueIntervals(intervals)
}

func mergeValueIntervals(intervals []registerInterval) []registerInterval {
	slices.SortFunc(intervals, func(a, b registerInterval) int {
		if n := cmp.Compare(a.slot, b.slot); n != 0 {
			return n
		}
		return cmp.Compare(a.first, b.first)
	})
	merged := intervals[:0]
	for _, interval := range intervals {
		if len(merged) != 0 {
			previous := &merged[len(merged)-1]
			if previous.slot == interval.slot && interval.first <= previous.last+1 {
				previous.last = max(previous.last, interval.last)
				previous.priority += interval.priority
				continue
			}
		}
		merged = append(merged, interval)
	}
	return merged
}
