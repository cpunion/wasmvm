package codegen

import (
	"encoding/binary"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func selectFixture(typ wasm.ValueType, dst slots.Slot) *slots.Function {
	return &slots.Function{Slots: []wasm.ValueType{typ, typ, wasm.I32, typ},
		Signature: module.FuncType{Results: []wasm.ValueType{typ}}, Args: []slots.Slot{dst},
		Sources: []ir.Source{{Wasm: module.Instr{Op: 0x1c}}, {Wasm: module.Instr{Op: 0xb}}},
		Code: []slots.Instruction{
			{Op: 0x1c, A: 0, B: 1, C: 2, Dst: dst},
			{Op: slots.ControlBase + uint32(ir.Return), Source: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Count: 1},
		}}
}

func TestScalarSelectEncodings(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		compiled, err := CompileUnmeteredFunction(arch, selectFixture(wasm.I64, 3), DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for i := 0; i+4 <= len(compiled.Entry.Code); i++ {
			b := compiled.Entry.Code[i:]
			if arch == "arm64" && i%4 == 0 && binary.LittleEndian.Uint32(b)&0xffe00c00 == 0x9a800000 ||
				arch == "amd64" && b[0]&0xf8 == 0x48 && b[1] == 0x0f && (b[2] == 0x44 || b[2] == 0x45) {
				found = true
			}
		}
		if !found {
			t.Fatal("missing branchless scalar selection", arch)
		}
	}
}

func TestScalarSelectBitsAliasesAndConditions(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.FuncRef, wasm.ExternRef} {
		for _, fast := range []bool{false, true} {
			for _, dst := range []slots.Slot{0, 1, 2, 3} {
				if dst == 2 && typ != wasm.I32 {
					continue
				}
				f := selectFixture(typ, dst)
				_, lease, ctx := configuredContext(t, f, Config{Unmetered: fast})
				for _, condition := range []uint64{0, 1, 0x80000000, 1 << 32, (1 << 32) | 1} {
					left, right := value.Value{Lo: 0x7ff012347f812345, Hi: 17}, value.Value{Lo: 0x8000000080000000, Hi: 19}
					if typ == wasm.I32 || typ == wasm.F32 {
						left.Lo, right.Lo = uint64(uint32(left.Lo)), uint64(uint32(right.Lo))
					}
					initial := []value.Value{left, right, {Lo: condition}, {Lo: 99, Hi: 23}}
					if fast {
						for i := range initial {
							initial[i].Hi = 0
						}
						left, right = initial[0], initial[1]
					}
					for i, v := range initial {
						writeSlot(t, ctx, 16, slots.Slot(i), v)
					}
					if err := ctx.SetState(native.State{Used: 16, Depth: 1, Budget: 64}); err != nil {
						t.Fatal(err)
					}
					exit, err := ctx.Enter(lease, 0)
					want := right
					if uint32(condition) != 0 {
						want = left
					}
					if err != nil || exit != native.Return || readSlot(t, ctx, 16, dst) != want {
						t.Fatal(typ, fast, dst, condition, readSlot(t, ctx, 16, dst), want, err, exit)
					}
				}
			}
		}
	}
}

func TestScalarSelectSpilledOperands(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	f := selectFixture(wasm.I64, 3)
	var prefix []slots.Instruction
	for slot := slots.Slot(4); slot < 20; slot++ {
		f.Slots = append(f.Slots, wasm.I64)
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(slot)}})
		prefix = append(prefix, slots.Instruction{Op: 0x20, Source: uint32(len(f.Sources) - 1), Dst: slot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	// These independently used locals take the fixed registers, leaving every
	// select operand/result in its backing slot.
	f.LocalCount = uint32(len(f.Slots))
	f.Code = append(prefix, f.Code...)
	_, lease, ctx := configuredContext(t, f, Config{Unmetered: true})
	for _, condition := range []uint64{0, 1 << 32, 1, (1 << 32) | 1} {
		writeSlot(t, ctx, 16, 0, value.Value{Lo: 111})
		writeSlot(t, ctx, 16, 1, value.Value{Lo: 222})
		writeSlot(t, ctx, 16, 2, value.Value{Lo: condition})
		if err := ctx.SetState(native.State{Used: 16, Depth: 1}); err != nil {
			t.Fatal(err)
		}
		exit, err := ctx.Enter(lease, 0)
		want := uint64(222)
		if uint32(condition) != 0 {
			want = 111
		}
		if err != nil || exit != native.Return || readSlot(t, ctx, 16, 3) != (value.Value{Lo: want}) {
			t.Fatal(condition, exit, err, readSlot(t, ctx, 16, 3))
		}
	}
}
