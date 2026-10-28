package codegen

import (
	"bytes"
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
		pattern := []byte{0x48, 0x0f, 0x44, 0xc1} // cmove rax, rcx.
		if arch == "arm64" {
			pattern = []byte{0x7f, 0x01, 0x00, 0x71, 0x29, 0x11, 0x8a, 0x9a} // cmp w11, #0; csel x9, x9, x10, ne.
		}
		if !bytes.Contains(compiled.Entry.Code, pattern) {
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
