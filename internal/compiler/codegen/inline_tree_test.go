package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func wrapLeaf(callee *slots.Function, target uint32) *slots.Function {
	return &slots.Function{
		Signature: callee.Signature, LocalCount: 2,
		Slots:   []wasm.ValueType{callee.Signature.Params[0], callee.Signature.Params[1], callee.Signature.Results[0]},
		Args:    []slots.Slot{0, 1, 2, 2},
		Sources: []ir.Source{{Wasm: module.Instr{Op: 0x10, A: target}, Inputs: callee.Signature.Params, Outputs: callee.Signature.Results}},
		Code: []slots.Instruction{
			{Op: 0x10, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
			{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Extra: 3, Count: 1},
		},
	}
}

func TestNestedLeafExpansionPreservesColdCallerFrames(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, op := range []uint32{0x6a, 0x6c, 0x7c, 0x7e} {
		inner := leafFixture(op, nil)
		middle := wrapLeaf(inner, 9)
		outer := wrapLeaf(middle, 8)
		for _, dst := range []slots.Slot{0, 2} {
			caller := leafCaller(outer, dst, 5)
			config := Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: outer, 8: middle, 9: inner}}
			compiled, lease, ctx := configuredContext(t, caller, config)
			if compiled.InlinedCalls != 1 || compiled.InlineFrames != 3 || compiled.InlineSlots != 11 {
				t.Fatal("missing original nested stack requirement", runtime.GOARCH, op, compiled.InlinedCalls, compiled.InlineFrames, compiled.InlineSlots)
			}
			for _, input := range []uint64{0, 1, 0x80000001, 0xfedcba9876543210, ^uint64(0)} {
				for pc := range caller.Code {
					initial := []value.Value{{Lo: input}, {Lo: input >> 1}, {Lo: 17}, {Lo: 19}, {Lo: 23}}
					for i := range initial {
						if caller.Slots[i] == wasm.I32 {
							initial[i].Lo = uint64(uint32(initial[i].Lo))
						}
						writeSlot(t, ctx, 16, slots.Slot(i), initial[i])
					}
					if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
						t.Fatal(err)
					}
					if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
						t.Fatal(op, pc, exit, err)
					}
					if pc < len(caller.Code)-1 {
						bits, trap := numeric.Eval(op, initial[0].Lo, initial[1].Lo)
						if trap != wasm.NoTrap {
							t.Fatal(trap)
						}
						initial[dst] = value.Value{Lo: bits}
					}
					for i, want := range initial {
						if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
							t.Fatal("nested expansion changed a cold frame", op, pc, i, got, want)
						}
					}
				}
			}
		}
	}
}

func TestLeafCallGraphBoundsAndCycles(t *testing.T) {
	leaf := leafFixture(0x6a, nil)
	catalog := map[uint32]*slots.Function{0: leaf}
	current := leaf
	for depth := uint32(1); depth <= 8; depth++ {
		current = wrapLeaf(current, depth-1)
		catalog[depth] = current
		if got := CanExpandLeaf(current, catalog); got != (depth < 8) {
			t.Fatal("unbounded nested expansion", depth, got)
		}
	}
	self := wrapLeaf(leaf, 0)
	if CanExpandLeaf(self, map[uint32]*slots.Function{0: self}) {
		t.Fatal("recursive expansion")
	}
	other := wrapLeaf(leaf, 1)
	if CanExpandLeaf(self, map[uint32]*slots.Function{0: other, 1: self}) {
		t.Fatal("mutually recursive expansion")
	}
	if CanExpandLeaf(self, nil) {
		t.Fatal("expanded an unknown/imported callee")
	}
	current, catalog = leaf, map[uint32]*slots.Function{0: leaf}
	for depth := uint32(1); depth <= 4; depth++ {
		parent := wrapLeaf(current, depth-1)
		call, ret := parent.Code[0], parent.Code[1]
		parent.Code = []slots.Instruction{call, call, call, call, ret}
		current, catalog[depth] = parent, parent
	}
	if CanExpandLeaf(current, catalog) {
		t.Fatal("shared definitions bypassed the total expansion budget")
	}
}
