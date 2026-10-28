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

func leafFixture(op uint32, literal *uint64) *slots.Function {
	typ, result, constant := wasm.I32, wasm.I32, uint32(0x41)
	if integerWide(op) {
		typ, result, constant = wasm.I64, wasm.I64, 0x42
	}
	if op >= 0x45 && op <= 0x5a {
		result = wasm.I32
	}
	f := &slots.Function{Signature: module.FuncType{Params: []wasm.ValueType{typ, typ}, Results: []wasm.ValueType{result}},
		LocalCount: 2, Slots: []wasm.ValueType{typ, typ, typ, typ, result}, Args: []slots.Slot{4},
		Sources: []ir.Source{{Wasm: module.Instr{Op: 0x20}}, {Wasm: module.Instr{Op: 0x20, A: 1}}, {Wasm: module.Instr{Op: op}}},
		Code: []slots.Instruction{
			{Op: 0x20, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x20, Source: 1, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: op, Source: 2, Dst: 4, A: 2, B: 3, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
		}}
	if op == 0x45 || op == 0x50 {
		f.Code[2].B = slots.NoSlot
	}
	if literal != nil {
		f.Sources[1].Wasm = module.Instr{Op: constant, Bits: *literal}
		f.Code[1].Op = constant
	}
	return f
}

func leafCaller(callee *slots.Function, destination slots.Slot, count int) *slots.Function {
	f := &slots.Function{LocalCount: uint32(count)}
	for i := 0; i < count; i++ {
		f.Slots = append(f.Slots, callee.Signature.Params[0])
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(i)}})
		f.Code = append(f.Code, slots.Instruction{Op: 0x20, Source: uint32(i), Dst: slots.Slot(i), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	f.Slots[destination] = callee.Signature.Results[0]
	f.Args = []slots.Slot{0, 1, destination}
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x10, A: 7}, Inputs: callee.Signature.Params, Outputs: callee.Signature.Results})
	f.Code = append(f.Code, slots.Instruction{Op: 0x10, Source: uint32(count), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 2})
	for i := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
	}
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Extra: 3, Count: uint32(count)})
	return f
}

func TestLeafExpansionPreservesAllCallerSlotsAndEntries(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, op := range integerOps() {
		if !(op >= 0x45 && op <= 0x5a || op >= 0x6a && op <= 0x8a) || scalarCanTrap(op) || op >= 0x79 && op <= 0x7b {
			continue
		}
		small, large := uint64(7), uint64(0xfedcba9876543210)
		for _, literal := range []*uint64{nil, &small, &large} {
			callee := leafFixture(op, literal)
			for _, dst := range []slots.Slot{0, 2} {
				if dst == 0 && integerWide(op) && op <= 0x5a {
					continue // An i64 argument and i32 result have distinct typed slots.
				}
				f := leafCaller(callee, dst, 5)
				config := Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}}
				compiled, lease, ctx := configuredContext(t, f, config)
				if compiled.InlinedCalls != 1 || compiled.InlineSlots != 5 {
					t.Fatalf("not expanded: op=%x dst=%d literal=%v", op, dst, literal)
				}
				for _, input := range []uint64{0, 1, 0x80000001, 0x8000000000000001, ^uint64(0)} {
					for pc := range f.Code {
						initial := []value.Value{{Lo: input}, {Lo: input >> 1}, {Lo: 17}, {Lo: 19}, {Lo: 23}}
						for i := range initial {
							if f.Slots[i] == wasm.I32 {
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
						if pc <= len(f.Code)-2 {
							right := initial[1].Lo
							if literal != nil {
								right = *literal
							}
							expected, trap := numeric.Eval(op, initial[0].Lo, right)
							if trap != wasm.NoTrap {
								t.Fatal(trap)
							}
							initial[dst] = value.Value{Lo: expected}
						}
						for i, expected := range initial {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != expected {
								t.Fatalf("op=%x pc=%d slot=%d: got=%v want=%v", op, pc, i, got, expected)
							}
						}
					}
				}
			}
		}
	}
}

func TestLeafExpansionKeepsOrdinaryCallsWhenIneligible(t *testing.T) {
	for _, kind := range []string{"metered", "import", "indirect", "trap", "control", "large", "references", "pressure", "signature"} {
		t.Run(kind, func(t *testing.T) {
			callee := leafFixture(0x6a, nil)
			f := leafCaller(callee, 2, 5)
			config := Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}}
			switch kind {
			case "metered":
				config.Unmetered = false
			case "import":
				config.DirectCallees = nil
			case "indirect":
				f.Code[5].Op = 0x11
			case "trap":
				callee.Code[2].Op = 0x6d
			case "control":
				callee.Code[2].Op = slots.ControlBase + uint32(ir.Trap)
			case "large":
				callee.Code = append(make([]slots.Instruction, 33), callee.Code...)
			case "references":
				callee.Slots[2] = wasm.ExternRef
			case "pressure":
				f = leafCaller(callee, 31, 32)
				// Give the other live locals real uses. Eliminated self-copies
				// no longer raise their priority above the call's output tuple.
				source := uint32(len(f.Sources))
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x72}, Inputs: []wasm.ValueType{wasm.I32, wasm.I32}, Outputs: []wasm.ValueType{wasm.I32}})
				tail := append([]slots.Instruction(nil), f.Code[32:]...)
				f.Code = f.Code[:32]
				for s := slots.Slot(0); s < 32; s++ {
					if s != 31 {
						f.Code = append(f.Code, slots.Instruction{Op: 0x72, Source: source, Dst: s, A: s, B: s, C: slots.NoSlot})
					}
				}
				f.Code = append(f.Code, tail...)
			case "signature":
				callee.Signature.Params = nil
			}
			for _, arch := range []string{"amd64", "arm64"} {
				compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
				if err != nil || compiled.InlinedCalls != 0 || compiled.Calls != 1 {
					t.Fatal(arch, compiled, err)
				}
			}
		})
	}
}

func TestLeafExpansionAfterForwardedReturn(t *testing.T) {
	callee := leafFixture(0x6a, nil)
	// The former return block remains a valid continuation, but entry returns
	// before reaching it. Do not lose direct-call expansion after forwarding.
	callee.Code = append(callee.Code, callee.Code[len(callee.Code)-1])
	if !CanExpandLeaf(callee) {
		t.Fatal("forwarded leaf no longer recognized")
	}
	f := leafCaller(callee, 2, 5)
	for _, arch := range []string{"amd64", "arm64"} {
		compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}})
		if err != nil {
			t.Fatal(arch, err)
		}
		if compiled.InlinedCalls != 1 || compiled.Calls != 1 {
			t.Fatal(arch, compiled.InlinedCalls, compiled.Calls)
		}
	}
}

func TestExpandedResultForwardingAliasesAndColdEntries(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for local := uint32(0); local < 5; local++ {
		for _, dead := range []bool{false, true} {
			callee := leafFixture(0x6a, nil)
			f := leafCaller(callee, 2, 5)
			last := f.Code[len(f.Code)-1]
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x22, A: local}})
			f.Code[len(f.Code)-1] = slots.Instruction{Op: 0x22, Source: uint32(len(f.Sources) - 1), Dst: 4, A: 2, B: slots.NoSlot, C: slots.NoSlot}
			if dead {
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x41, Bits: 99}})
				f.Code = append(f.Code, slots.Instruction{Op: 0x41, Source: uint32(len(f.Sources) - 1), Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
			}
			f.Code = append(f.Code, last)
			compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}})
			if compiled.InlinedCalls != 1 {
				t.Fatal("missing expansion", local, dead)
			}
			for pc := range f.Code {
				expected := []uint64{17, 25, 31, 37, 43}
				for i, v := range expected {
					writeSlot(t, ctx, 16, slots.Slot(i), value.Value{Lo: v})
				}
				if pc <= 5 {
					expected[2] = 42
				}
				if pc <= 6 {
					expected[local], expected[4] = expected[2], expected[2]
				}
				if dead && pc <= 7 {
					expected[2] = 99
				}
				if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
					t.Fatal(err)
				}
				if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
					t.Fatal(exit, err)
				}
				for i, v := range expected {
					if got := readSlot(t, ctx, 16, slots.Slot(i)); got != (value.Value{Lo: v}) {
						t.Fatal(local, dead, pc, i, got, v)
					}
				}
			}
		}
	}
}

func TestLeafExpansionLeavesCancellationAtTheCall(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	callee := leafFixture(0x6a, nil)
	f := leafCaller(callee, 2, 5)
	compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}})
	if compiled.InlinedCalls != 1 {
		t.Fatal(runtime.GOARCH, "missing expansion")
	}
	writeSlot(t, ctx, 16, 2, value.Value{Lo: 99})
	if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: 5}); err != nil {
		t.Fatal(err)
	}
	ctx.Cancel()
	if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Poll || readSlot(t, ctx, 16, 2).Lo != 99 {
		t.Fatal(exit, err)
	}
}

func TestLeafMultiplyAddPreservesAliasesAndSharedProducts(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64} {
		for _, commute := range []bool{false, true} {
			for _, shared := range []bool{false, true} {
				add, mul := uint32(0x6a), uint32(0x6c)
				if typ == wasm.I64 {
					add, mul = 0x7c, 0x7e
				}
				callee := &slots.Function{Signature: module.FuncType{Params: []wasm.ValueType{typ, typ, typ}, Results: []wasm.ValueType{typ}}, LocalCount: 3,
					Slots: []wasm.ValueType{typ, typ, typ, typ, typ}, Args: []slots.Slot{4},
					Sources: []ir.Source{{Wasm: module.Instr{Op: mul}}, {Wasm: module.Instr{Op: add}}},
					Code: []slots.Instruction{
						{Op: mul, Dst: 3, A: 0, B: 1, C: slots.NoSlot},
						{Op: add, Source: 1, Dst: 4, A: 3, B: 2, C: slots.NoSlot},
						{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
					}}
				if commute {
					callee.Code[1].A, callee.Code[1].B = 2, 3
				}
				if shared {
					callee.Signature.Results = append(callee.Signature.Results, typ)
					callee.Args = []slots.Slot{3, 4}
					callee.Code[2].Count = 2
				}
				f := leafCaller(callee, 2, 5)
				f.Code[5].Count = 3
				f.Args = []slots.Slot{0, 1, 3, 2}
				if shared {
					f.Args = append(f.Args, 4)
				}
				f.Code[6].Extra = uint32(len(f.Args))
				f.Args = append(f.Args, 0, 1, 2, 3, 4)
				compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}})
				if compiled.InlinedCalls != 1 {
					t.Fatal("multiply-add leaf not expanded", typ, commute, shared)
				}
				for _, arg := range []uint64{0, 1, 0x80000001, ^uint64(0)} {
					for pc := range f.Code {
						initial := []value.Value{{Lo: arg}, {Lo: arg ^ 0x12345678}, {Lo: 3}, {Lo: arg >> 1}, {Lo: 7}}
						for i := range initial {
							if typ == wasm.I32 {
								initial[i].Lo = uint64(uint32(initial[i].Lo))
							}
							writeSlot(t, ctx, 16, slots.Slot(i), initial[i])
						}
						if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
							t.Fatal(err)
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(exit, err)
						}
						if pc <= 5 {
							product := initial[0].Lo * initial[1].Lo
							sum := product + initial[3].Lo
							if typ == wasm.I32 {
								product, sum = uint64(uint32(product)), uint64(uint32(sum))
							}
							initial[2].Lo = sum
							if shared {
								initial[2].Lo, initial[4].Lo = product, sum
							}
						}
						for i, want := range initial {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
								t.Fatal(typ, commute, shared, arg, pc, i, got, want)
							}
						}
					}
				}
			}
		}
	}
}
