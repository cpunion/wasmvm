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

func TestConstantPoolSpillsAndIndependentEntries(t *testing.T) {
	// More constants and slots than available pool/register entries exercise
	// both cached and ordinary literals, including the frame spill path.
	values := []uint64{0, 1, 0xffff, 0xffffffff, 0x12345678, 0x87654321,
		0xfedcba9876543210, 0x123456789abcdef0, 0xabcdef1234567890,
		0x1122334455667788, 0x2233445566778899, 0x33445566778899aa,
		0x445566778899aabb, 0x5566778899aabbcc, 0x66778899aabbccdd}
	f := &slots.Function{}
	for i, v := range values {
		f.Slots = append(f.Slots, wasm.I64)
		f.Args = append(f.Args, slots.Slot(i))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x42, Bits: v}})
		f.Code = append(f.Code, slots.Instruction{Op: 0x42, Source: uint32(i), Dst: slots.Slot(i), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: uint32(len(values))})
	for _, arch := range []string{"amd64", "arm64"} {
		if _, err := CompileUnmeteredFunction(arch, f, DefaultLimits()); err != nil {
			t.Fatal(arch, err)
		}
	}
	if !native.Available {
		return
	}
	_, lease, ctx := functionContext(t, f, true)
	for start := range f.Code {
		for i := range values {
			writeSlot(t, ctx, 16, slots.Slot(i), value.Value{Lo: uint64(i) + 99})
		}
		if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start)}); err != nil {
			t.Fatal(err)
		}
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
			t.Fatal(exit, err)
		}
		for i, v := range values {
			if i < start {
				v = uint64(i) + 99
			}
			if got := readSlot(t, ctx, 16, slots.Slot(i)); got != (value.Value{Lo: v}) {
				t.Fatal(start, i, got, v)
			}
		}
	}
}
