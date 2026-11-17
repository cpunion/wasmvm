package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func countedLeafFixture(wide bool, count uint64) *slots.Function {
	typ, literal, add, less := wasm.I32, uint32(0x41), uint32(0x6a), uint32(0x49)
	if wide {
		typ, literal, add, less = wasm.I64, 0x42, 0x7c, 0x54
	}
	f := &slots.Function{
		Signature:  module.FuncType{Params: []wasm.ValueType{typ, typ}, Results: []wasm.ValueType{typ}},
		LocalCount: 3, Slots: []wasm.ValueType{typ, typ, typ, typ, wasm.I32}, Args: []slots.Slot{0},
		Edges: []slots.Edge{{PC: 0}, {PC: 9}},
	}
	f.Code = []slots.Instruction{
		{Op: literal, Dst: 3},
		{Op: add + 2, Dst: 0, A: 0, B: 3},
		{Op: literal, Dst: 3},
		{Op: add, Dst: 0, A: 0, B: 3},
		{Op: literal, Dst: 3},
		{Op: add, Dst: 2, A: 2, B: 3},
		{Op: literal, Dst: 3},
		{Op: less, Dst: 4, A: 2, B: 3},
		{Op: slots.ControlBase + uint32(ir.If), A: 4, Count: 2},
		{Op: slots.ControlBase + uint32(ir.Return), Count: 1},
	}
	for pc := range f.Code {
		in := &f.Code[pc]
		in.Source, in.C = uint32(pc), slots.NoSlot
		if in.Op == literal || in.Op >= slots.ControlBase {
			in.B = slots.NoSlot
			if in.Op != slots.ControlBase+uint32(ir.If) {
				in.A = slots.NoSlot
			}
		}
		if in.Op >= slots.ControlBase {
			in.Dst = slots.NoSlot
		}
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: in.Op}})
	}
	f.Sources[0].Wasm.Bits = 1664525
	f.Sources[2].Wasm.Bits = 1013904223
	f.Sources[4].Wasm.Bits = 1
	f.Sources[6].Wasm.Bits = count
	return f
}

func TestBoundedConstantControlPreservesColdCallerValues(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, count := range []uint64{0, 1, 2, 3, 7, 16} {
			callee := countedLeafFixture(wide, count)
			for _, destination := range []slots.Slot{0, 2} {
				caller := leafCaller(callee, destination, 5)
				config := Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}}
				for _, arch := range []string{"arm64", "amd64"} {
					c, err := CompileFunctionWithConfig(arch, caller, DefaultLimits(), config)
					if err != nil || c.InlinedCalls != 1 || c.InlineFrames != 1 || c.InlineSlots != uint32(len(callee.Slots)) {
						t.Fatal(arch, wide, count, c, err)
					}
				}
				if !native.Available {
					continue
				}
				_, lease, ctx := configuredContext(t, caller, config)
				for _, seed := range []uint64{0, 1, 0x80000001, 0xfedcba9876543210, ^uint64(0)} {
					for pc := range caller.Code {
						initial := []value.Value{{Lo: seed}, {Lo: seed >> 1}, {Lo: 17}, {Lo: 19}, {Lo: 23}}
						for i := range initial {
							if !wide {
								initial[i].Lo = uint64(uint32(initial[i].Lo))
							}
							writeSlot(t, ctx, 16, slots.Slot(i), initial[i])
						}
						if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
							t.Fatal(err)
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(wide, count, pc, exit, err)
						}
						if pc < len(caller.Code)-1 {
							x := initial[0].Lo
							for i := uint64(0); i < max(count, 1); i++ {
								x = x*1664525 + 1013904223
								if !wide {
									x = uint64(uint32(x))
								}
							}
							initial[destination] = value.Value{Lo: x}
						}
						for i, want := range initial {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
								t.Fatal(wide, count, seed, pc, i, got, want)
							}
						}
					}
				}
			}
		}
	}
}

func TestConstantControlRejectsUnknownOrUnboundedPaths(t *testing.T) {
	for _, kind := range []string{"unknown", "infinite", "budget", "trap", "memory"} {
		f := countedLeafFixture(false, 2)
		switch kind {
		case "unknown":
			f.Code[8].A = 0
		case "infinite":
			f.Edges[1].PC = 0
		case "budget":
			f.Sources[6].Wasm.Bits = 1000
		case "trap":
			f.Code[1].Op = 0x6d
		case "memory":
			f.Code[1].Op = 0x28
		}
		if CanExpandLeaf(f, nil) {
			t.Fatal("expanded ineligible control flow", kind)
		}
	}
}
