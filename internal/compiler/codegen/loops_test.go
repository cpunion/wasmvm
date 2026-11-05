package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

func TestNativeLoopEligibilityKeepsObservableBoundaries(t *testing.T) {
	for _, kind := range []string{"accept", "precise", "trap", "memory", "memory-size", "call", "moves", "interior", "table", "poll-span"} {
		f := coalescedArithmeticLoopFixture(false, true, false)
		e := functionEmitter{f: f, unmetered: true}
		switch kind {
		case "precise":
			e.unmetered = false
		case "trap":
			f.Code[3].Op = 0x6d
		case "memory":
			f.Code[3].Op = 0x28
		case "memory-size":
			f.Code[3].Op = 0x3f
		case "call":
			f.Code[3].Op = 0x10
		case "moves":
			f.Edges[0].Count = 1
			f.Moves = []slots.Move{{Dst: 2, Src: 3}}
		case "interior":
			f.Edges = append(f.Edges, slots.Edge{PC: 5})
		case "table":
			f.Code[len(f.Code)-2].Op = slots.ControlBase + uint32(ir.Table)
		case "poll-span":
			branch, ret := f.Code[len(f.Code)-2], f.Code[len(f.Code)-1]
			f.Code = f.Code[:len(f.Code)-2]
			for len(f.Code) < 257 {
				f.Code = append(f.Code, slots.Instruction{Op: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot})
			}
			f.Code = append(f.Code, branch, ret)
			f.Edges[1].PC = uint32(len(f.Code) - 1)
		}
		regions := e.addNativeLoops(nil)
		if kind == "accept" {
			if regions[1].loop == nil || regions[1].loop.memory != nil || !regions[1].loop.flags {
				t.Fatal("ordinary loop not recognized")
			}
		} else if len(regions) != 0 {
			t.Fatal(kind, "observable or unproved region accepted")
		}
	}
}
