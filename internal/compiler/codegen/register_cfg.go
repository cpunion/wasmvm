package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

type registerCFGEdge struct{ from, to int }

func registerBranch(op uint32) bool {
	return op >= slots.ControlBase+uint32(ir.Jump) && (op <= slots.ControlBase+uint32(ir.Table) || op == slots.ControlBase+uint32(ir.IfCast) || op == slots.ControlBase+uint32(ir.IfNull))
}

// Iterative Tarjan traversal identifies actual cycles, independently of block
// layout. A backward exit to an earlier return block is not a loop. Edges are
// bounded by the caller's tuple and map-cell budgets before graph allocation.
func registerLoopComponents(blocks int, edges []registerCFGEdge) []registerLoop {
	head := make([]int, blocks)
	for i := range head {
		head[i] = -1
	}
	next := make([]int, len(edges))
	for i, edge := range edges {
		next[i], head[edge.from] = head[edge.from], i
	}
	indices, low := make([]int, blocks), make([]int, blocks)
	active := make([]bool, blocks)
	var pending []int
	type frame struct{ block, edge int }
	var dfs []frame
	clock := 0
	visit := func(block int) {
		clock++
		indices[block], low[block], active[block] = clock, clock, true
		pending = append(pending, block)
		dfs = append(dfs, frame{block, head[block]})
	}
	var loops []registerLoop
	for start := 0; start < blocks; start++ {
		if indices[start] != 0 {
			continue
		}
		visit(start)
		for len(dfs) != 0 {
			current := &dfs[len(dfs)-1]
			block := current.block
			if current.edge >= 0 {
				edge := current.edge
				current.edge = next[edge]
				target := edges[edge].to
				if indices[target] == 0 {
					visit(target)
				} else if active[target] {
					low[block] = min(low[block], indices[target])
				}
				continue
			}
			dfs = dfs[:len(dfs)-1]
			if len(dfs) != 0 {
				parent := dfs[len(dfs)-1].block
				low[parent] = min(low[parent], low[block])
			}
			if low[block] != indices[block] {
				continue
			}
			span := registerLoop{block, block}
			count := 0
			for {
				member := pending[len(pending)-1]
				pending = pending[:len(pending)-1]
				active[member] = false
				span.first, span.last = min(span.first, member), max(span.last, member)
				count++
				if member == block {
					break
				}
			}
			if count > 1 {
				loops = append(loops, span)
			} else {
				for edge := head[block]; edge >= 0; edge = next[edge] {
					if edges[edge].to == block {
						loops = append(loops, span)
						break
					}
				}
			}
		}
	}
	return loops
}
