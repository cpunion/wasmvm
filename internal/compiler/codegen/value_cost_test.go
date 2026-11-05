package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestValueAllocationChargesShufflesAndCompleteEdges(t *testing.T) {
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}, Code: make([]slots.Instruction, 12)}
	g := &valueFlow{ops: make([]valueFlowOperation, len(f.Code))}
	for pc := range f.Code {
		f.Code[pc].Op = 0x6a
		g.ops[pc] = valueFlowOperation{first: pc, count: 1}
		g.accesses = append(g.accesses, valueFlowAccess{slot: slots.Slot(pc % 2)})
	}
	f.Code[len(f.Code)-1].Op = slots.ControlBase + uint32(ir.Return)
	e := &functionEmitter{f: f, valueFlow: g}
	base := &registerPlan{homes: map[slots.Slot]byte{0: 3}, slots: []slots.Slot{0}}
	other := &registerPlan{homes: map[slots.Slot]byte{1: 3}, slots: []slots.Slot{1}}
	both := &registerPlan{homes: map[slots.Slot]byte{0: 3, 1: 4}, slots: []slots.Slot{0, 1}}
	stable, shuffled, useful := make([]*registerPlan, len(f.Code)), make([]*registerPlan, len(f.Code)), make([]*registerPlan, len(f.Code))
	for pc := range f.Code {
		stable[pc], shuffled[pc], useful[pc] = base, base, both
		if pc%2 != 0 {
			shuffled[pc] = other
		}
	}
	oldCost := e.valueAllocationCost(stable, nil, nil)
	if cost := e.valueAllocationCost(shuffled, nil, nil); cost <= oldCost {
		t.Fatal("zero misses hid repeated raw-word spill/reload shuffles", cost, oldCost)
	}
	if cost := e.valueAllocationCost(useful, nil, nil); cost >= oldCost {
		t.Fatal("profitable additional residency was rejected", cost, oldCost)
	}
	// An actual backedge pays its source/destination mapping even without a
	// linear fallthrough boundary. Increasing its frequency must increase cost.
	f.Code[len(f.Code)-1].Op = slots.ControlBase + uint32(ir.Jump)
	g.blocks = []valueFlowBlock{{first: 0, end: len(f.Code)}}
	g.edges = []valueFlowEdge{{from: 0, to: 0}}
	low := e.valueAllocationCost(shuffled, nil, nil)
	weights := make([]uint64, len(f.Code))
	for i := range weights {
		weights[i] = 1
	}
	weights[len(weights)-1] = 100
	if high := e.valueAllocationCost(shuffled, nil, weights); high <= low {
		t.Fatal("taken backedge frequency did not charge its transfer", low, high)
	}
}
