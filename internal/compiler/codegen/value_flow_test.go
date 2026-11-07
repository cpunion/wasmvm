package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestValueFlowPhisAndOrderedEdgeCopyVersions(t *testing.T) {
	f := &slots.Function{LocalCount: 6, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32}, Args: []slots.Slot{4}, Sources: []ir.Source{{}}}
	for _, in := range []slots.Instruction{
		{Op: 0x41, Dst: 0}, {Op: slots.ControlBase + uint32(ir.If), A: 3, Extra: 0, Count: 2},
		{Op: 0x41, Dst: 0}, {Op: slots.ControlBase + uint32(ir.Jump), Extra: 2, Count: 1},
		{Op: 0x41, Dst: 0}, {Op: slots.ControlBase + uint32(ir.Jump), Extra: 3, Count: 1},
		{Op: 0x6a, Dst: 4, A: 0, B: 1}, {Op: slots.ControlBase + uint32(ir.If), A: 3, Extra: 4, Count: 2},
		{Op: slots.ControlBase + uint32(ir.Return), Extra: 0, Count: 1},
	} {
		if in.Op != 0x41 && in.Op != 0x6a {
			in.Dst = slots.NoSlot
		}
		if in.Op != 0x6a {
			in.B = slots.NoSlot
		}
		if in.Op != 0x6a && in.Op != slots.ControlBase+uint32(ir.If) {
			in.A = slots.NoSlot
		}
		in.C = slots.NoSlot
		f.Code = append(f.Code, in)
	}
	f.Edges = []slots.Edge{{PC: 2}, {PC: 4}, {PC: 6, Count: 3}, {PC: 6, Start: 3, Count: 1}, {PC: 2}, {PC: 8}}
	f.Moves = []slots.Move{{Dst: 5, Src: 0}, {Dst: 0, Src: 1}, {Dst: 1, Src: 5}, {Dst: 1, Src: 0}}
	g := buildValueFlow(f, []int{0, 2, 4, 6, 8}, nil, nil, 1<<18)
	if g == nil {
		t.Fatal("missing CFG versions")
	}
	edge := g.edges[2]
	zero, one, scratch := edge.values[0], edge.values[1], edge.values[5]
	if g.nodes[one].count != 1 || g.inputs[g.nodes[one].inputs] != scratch || g.inputs[g.nodes[zero].inputs] == scratch {
		t.Fatal("scheduled copy cycle lost its intermediate version")
	}
	for _, slot := range []slots.Slot{0, 1} {
		phi := g.nodes[g.blocks[3].entry[slot]]
		if !phi.phi || phi.opaque || phi.count != 2 {
			t.Fatal("join did not merge both selected edge versions", slot, phi)
		}
	}
	if !g.nodes[g.blocks[0].entry[3]].opaque {
		t.Fatal("entry word became a predecessor equality fact")
	}
	intervals := g.intervals()
	for i := 1; i < len(intervals); i++ {
		if intervals[i-1].slot == intervals[i].slot && intervals[i].first <= intervals[i-1].last+1 {
			t.Fatal("one canonical slot has overlapping register candidates")
		}
	}
	if buildValueFlow(f, []int{0, 2, 4, 6, 8}, nil, nil, 1) != nil {
		t.Fatal("flow cell budget ignored")
	}
}

func TestValueFlowOpaqueResumeAndTupleOutputs(t *testing.T) {
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32}, Args: []slots.Slot{0, 1, 2, 3},
		Sources: []ir.Source{{Outputs: []wasm.ValueType{wasm.I32}}, {Wasm: module.Instr{Op: 0x6a}}, {}},
		Code: []slots.Instruction{
			{Op: 0x10, Source: 0, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
			{Op: 0x6a, Source: 1, Dst: 3, A: 2, B: 0, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.Return), Source: 2, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Extra: 3, Count: 1}},
		Prompts: map[uint32][]slots.Catch{0: {{PC: 1}}}}
	g := buildValueFlow(f, []int{0, 1}, nil, nil, 1000)
	if g == nil {
		t.Fatal("missing tuple flow")
	}
	output := g.nodes[g.blocks[0].current[2]]
	if output.slot != 2 || output.count != 2 {
		t.Fatal("call outputs were treated as additional inputs", output)
	}
	if !g.nodes[g.blocks[1].entry[0]].opaque {
		t.Fatal("driver resume discarded its opaque initial word")
	}
	if len(g.accesses) != 7 {
		t.Fatal("tuple aliases duplicated instruction accesses", len(g.accesses))
	}
}

func TestValueFlowResolvesDeepPredecessorsWithoutRecursion(t *testing.T) {
	const depth = 10000
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}, Sources: []ir.Source{{}}, Args: []slots.Slot{1}}
	starts := make([]int, depth)
	for i := 0; i < depth-1; i++ {
		starts[i] = i
		f.Edges = append(f.Edges, slots.Edge{PC: uint32(i + 1)})
		f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Jump), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Extra: uint32(i), Count: 1})
	}
	starts[depth-1] = depth - 1
	f.Code = append(f.Code, slots.Instruction{Op: 0x6a, Dst: 1, A: 0, B: 0, C: slots.NoSlot}, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1})
	g := buildValueFlow(f, starts, nil, nil, 1<<18)
	if g == nil || len(g.nodes) != depth+1 || len(g.pending) != depth {
		t.Fatal("predecessor resolution was incomplete")
	}
}
