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

func TestMixedScalarForwardingEveryColdFrame(t *testing.T) {
	for _, op := range []uint32{0x6a, 0x6b, 0x6c, 0x71, 0x72, 0x73, 0x7c, 0x7d, 0x7e, 0x83, 0x84, 0x85} {
		typ, add := wasm.I32, uint32(0x6a)
		if integerWide(op) {
			typ, add = wasm.I64, 0x7c
		}
		f := &slots.Function{LocalCount: 30}
		for s := range 30 {
			f.Slots = append(f.Slots, typ)
			f.Args = append(f.Args, slots.Slot(s))
			f.Signature.Results = append(f.Signature.Results, typ)
		}
		emit := func(op uint32, dst, a, b slots.Slot, local uint32) {
			source := uint32(len(f.Sources))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Offset: uint64(source) + 100}})
			f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
		}
		emit(0x20, 3, slots.NoSlot, slots.NoSlot, 0)
		emit(0x20, 4, slots.NoSlot, slots.NoSlot, 1)
		emit(op, 2, 3, 4, 0)
		emit(0x22, 6, 2, slots.NoSlot, 5)
		// All words remain observable at return. Hot independent locals put
		// pressure on the GPR bank before the copied arithmetic is selected.
		for s := slots.Slot(8); s < 28; s++ {
			for range 5 {
				emit(add, s, s, 28, 0)
			}
		}
		emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
		f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
			if err != nil {
				t.Fatal(arch, op, err)
			}
			if arch == "amd64" && (c.ForwardedScalars == 0 || c.MixedIntegers == 0) {
				t.Fatal(arch, op, c.ForwardedScalars, c.MixedIntegers)
			}
		}
		if !native.Available {
			continue
		}
		_, lease, ctx := functionContext(t, f, true)
		for start := range f.Code {
			for _, bits := range []uint64{0, 0xffffffff, 1 << 63, 0xfedcba9876543210} {
				initial := make([]value.Value, len(f.Slots))
				for s := range initial {
					initial[s] = value.Value{Lo: bits ^ uint64(s), Hi: 0x8765432100000000 | uint64(s)}
					writeSlot(t, ctx, 16, slots.Slot(s), initial[s])
				}
				want, exit, pc, payload := modelFrameCopies(f, initial, start, -1, false, nil)
				if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start), Fuel: 42, Budget: 37}); err != nil {
					t.Fatal(err)
				}
				got, err := ctx.Enter(lease, 0)
				state, stateErr := ctx.State()
				if err != nil || stateErr != nil || got != exit || state.Continuation != pc || state.Payload[0] != payload || state.Fuel != 42 || state.Budget != 37 {
					t.Fatal(op, start, bits, got, state, err, stateErr)
				}
				for s, word := range want {
					if got := readSlot(t, ctx, 16, slots.Slot(s)); got != word {
						t.Fatal(op, start, bits, "word", s, got, word)
					}
				}
			}
		}
	}
}
