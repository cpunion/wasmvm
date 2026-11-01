package codegen

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func wideSlotFunction(op uint32, large bool) *slots.Function {
	types := []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64}
	args := []slots.Slot{0, 2, 1, 3, 0, 1}
	count := uint32(4)
	if op >= 0xfc0015 {
		args, count = []slots.Slot{0, 1, 1, 0}, 2
	} else if large {
		types = make([]wasm.ValueType, 2052)
		for i := range types {
			types[i] = wasm.I64
		}
		args[1], args[3] = 2050, 2051
	}
	outputs := []wasm.ValueType{wasm.I64, wasm.I64}
	result := append([]slots.Slot(nil), args[count:]...)
	args = append(args, result...)
	return &slots.Function{
		Signature: module.FuncType{Params: types, Results: outputs}, LocalCount: uint32(len(types)), Slots: types,
		Sources: []ir.Source{{Wasm: module.Instr{Op: op}, Outputs: outputs}, {}}, Args: args,
		Code: []slots.Instruction{
			{Op: op, Dst: result[0], A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: count},
			{Op: slots.ControlBase + uint32(ir.Return), Source: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Extra: count + 2, Count: 2},
		},
	}
}

func TestWideNativeAliasedResultsLargeFrameAndFuel(t *testing.T) {
	for op := uint32(0xfc0013); op <= 0xfc0016; op++ {
		for _, large := range []bool{false, true} {
			f := wideSlotFunction(op, large)
			for _, arch := range []string{"amd64", "arm64"} {
				for _, fast := range []bool{false, true} {
					compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: fast})
					if err != nil || compiled.Helpers != 0 {
						t.Fatal(arch, op, fast, err)
					}
				}
			}
			if !native.Available {
				continue
			}
			for _, fast := range []bool{false, true} {
				compiled, lease, ctx := functionContext(t, f, fast)
				args := make([]value.Value, len(f.Slots))
				args[0].Lo, args[1].Lo = ^uint64(0), 1
				if op <= 0xfc0014 {
					args[f.Args[1]].Lo, args[f.Args[3]].Lo = 7, 3
				}
				want := []value.Value{{Lo: 0}, {Lo: 11}}
				switch op {
				case 0xfc0014:
					want = []value.Value{{Lo: ^uint64(1)}, {Lo: 4}}
				case 0xfc0015:
					want = []value.Value{{Lo: ^uint64(0)}, {Lo: ^uint64(0)}}
				case 0xfc0016:
					want = []value.Value{{Lo: ^uint64(0)}, {Lo: 0}}
				}
				got, exit, _ := runFunction(t, compiled, lease, ctx, args, 100, !fast, 1)
				if exit != native.Return || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
					t.Fatal(op, large, fast, got, exit, want)
				}
				if !fast {
					if _, exit, _ := runFunction(t, compiled, lease, ctx, args, 0, true, 1); exit != native.FuelExhausted {
						t.Fatal("wide instruction bypassed fuel", exit)
					}
				}
			}
		}
	}
}

func TestWideNativeRejectsInvalidTuples(t *testing.T) {
	for _, change := range []func(*slots.Function){
		func(f *slots.Function) { f.Code[0].Count = 3 },
		func(f *slots.Function) { f.Sources[0].Outputs = nil },
		func(f *slots.Function) { f.Code[0].Extra = ^uint32(0) },
		func(f *slots.Function) { f.Args[5] = f.Args[4] },
		func(f *slots.Function) { f.Slots[3] = wasm.I32 },
	} {
		f := wideSlotFunction(0xfc0013, false)
		change(f)
		if _, err := CompileFunction("arm64", f, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
}
