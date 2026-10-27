package codegen

import (
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFallthroughPollingPreservesBoundAndColdValues(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, kind := range []string{"fallthrough", "forward", "backward", "copies", "conditional", "table", "periodic", "chain"} {
		t.Run(kind, func(t *testing.T) {
			f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64}, Args: []slots.Slot{0},
				Sources: []ir.Source{{Wasm: module.Instr{Op: 1}}, {Wasm: module.Instr{Op: 0x42, Bits: 42}}},
				Code: []slots.Instruction{
					{Op: 1}, {Op: slots.ControlBase + uint32(ir.Jump), Count: 1}, {Op: 0x42, Source: 1, Dst: 0},
					{Op: slots.ControlBase + uint32(ir.Return), Count: 1}},
				Edges: []slots.Edge{{PC: 2}}}
			start, wantPC, wantValue := uint64(1), uint64(1), uint64(7)
			switch kind {
			case "fallthrough":
				wantPC, wantValue = 3, 42
			case "forward":
				f.Edges[0].PC = 3
			case "backward":
				f.Edges[0].PC = 1
			case "copies":
				f.Edges[0].Count = 1
				f.Moves = []slots.Move{{Dst: 0, Src: 1}}
			case "conditional", "table":
				f.Code[1].Op = slots.ControlBase + uint32(ir.Table)
				f.Code[1].Count = 1
				if kind == "conditional" {
					f.Code[1].Op = slots.ControlBase + uint32(ir.If)
					f.Code[1].Count = 2
					f.Edges = append(f.Edges, f.Edges[0])
				}
			case "periodic", "chain":
				f.Code, f.Edges = nil, nil
				for i := range 300 {
					f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Jump), Extra: uint32(i), Count: 1})
					f.Edges = append(f.Edges, slots.Edge{PC: uint32(i + 1)})
				}
				f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: 1})
				wantPC = 256
				if kind == "periodic" {
					start = 256
				}
			}
			for _, fast := range []bool{false, true} {
				compile := CompileFunction
				if fast {
					compile = CompileUnmeteredFunction
				}
				compiled, err := compile(runtime.GOARCH, f, DefaultLimits())
				if err != nil {
					t.Fatal(err)
				}
				// Cancel after the adapter's early check, before a cold PC entry.
				// The generated poll must publish live registers and the exact PC.
				var prefix []byte
				if runtime.GOARCH == "arm64" {
					for _, word := range []uint32{0xd2800029, 0x9100000a | uint32(nativeabi.ControlOffset)<<10, 0xc89ffd49} {
						prefix = binary.LittleEndian.AppendUint32(prefix, word)
					}
				} else {
					prefix = []byte{0x48, 0xc7, 0x87}
					prefix = binary.LittleEndian.AppendUint32(prefix, uint32(nativeabi.ControlOffset))
					prefix = binary.LittleEndian.AppendUint32(prefix, 1)
				}
				code, err := native.Publish(append(prefix, compiled.Entry.Code...), []uint32{0}, DefaultLimits().MaxCodeBytes)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(code.Close)
				lease, err := code.Acquire()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(lease.Close)
				c, err := native.NewContext(32, 32, 1)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(c.Close)
				writeSlot(t, c, 0, 0, value.Value{Lo: 7})
				writeSlot(t, c, 0, 1, value.Value{Lo: 42})
				if err := c.SetState(native.State{Depth: 1, Continuation: start, Budget: 600}); err != nil {
					t.Fatal(err)
				}
				exit, err := c.Enter(lease, 0)
				state, stateErr := c.State()
				pc, v := wantPC, wantValue
				if !fast {
					pc, v = start, 7
				}
				if err != nil || stateErr != nil || exit != native.Poll || state.Continuation != pc || readSlot(t, c, 0, 0) != (value.Value{Lo: v}) {
					t.Fatalf("fast=%v: exit %v state %+v errors %v/%v, want PC %d value %d", fast, exit, state, err, stateErr, pc, v)
				}
			}
		})
	}
}
