package codegen

import (
	"bytes"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestConstantPoolPressureAndContinuations(t *testing.T) {
	for _, wide := range []bool{false, true} {
		op, typ := uint32(0x41), wasm.I32
		if wide {
			op, typ = 0x42, wasm.I64
		}
		f := &slots.Function{}
		for i := 0; i < 24; i++ {
			v := uint64(i+1) * 0x123456789ab
			if !wide {
				v = uint64(uint32(v))
			}
			f.Slots = append(f.Slots, typ)
			f.Args = append(f.Args, slots.Slot(i))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: v}})
			f.Code = append(f.Code, slots.Instruction{Op: op, Source: uint32(i), Dst: slots.Slot(i), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
		f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: 24, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		for _, arch := range []string{"arm64", "amd64"} {
			first, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 3; i++ {
				next, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
				if err != nil || !bytes.Equal(first.Entry.Code, next.Entry.Code) {
					t.Fatal("nondeterministic constant allocation", arch, err)
				}
			}
		}
		if native.Available {
			checkLinearIntegerEntries(t, f)
		}
	}
}
