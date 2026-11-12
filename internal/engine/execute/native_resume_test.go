package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestNativeResumeAdoptionChecksParentsAndRetiresFrames(t *testing.T) {
	e, batch := frameTransfer(t)
	if err := e.adoptNativeFrames(&batch); err != nil || len(e.frames) != 2 {
		t.Fatal(err, len(e.frames))
	}
	retained := e.frames[:cap(e.frames)]
	root := e.frames[0].function
	var pc uint32
	for i, in := range root.Code.Code {
		if in.Op == slots.ControlBase+uint32(ir.Return) {
			pc = uint32(i)
			break
		}
	}
	out := state.NativeFrames{Parents: 1, PC: pc, Frames: []state.NativeFrame{{Function: root, PC: pc}}, Values: e.values[:len(root.Code.Slots)]}
	for _, parents := range []uint32{2, ^uint32(0)} {
		invalid := out
		invalid.Parents = parents
		if err := e.adoptNativeFrames(&invalid); !errors.Is(err, wasm.ErrInvalid) || len(e.frames) != 2 {
			t.Fatal("invalid parent count mutated stack", parents, err)
		}
	}
	invalid := out
	invalid.Frames = nil
	if err := e.adoptNativeFrames(&invalid); !errors.Is(err, wasm.ErrInvalid) || len(e.frames) != 2 {
		t.Fatal("unvalidated parent retirement", err)
	}
	if err := e.adoptNativeFrames(&out); err != nil || len(e.frames) != 1 || e.frames[0].pc != pc || len(e.values) != len(root.Code.Slots) {
		t.Fatal("native parent return", err, e.frames)
	}
	if retained[1].function != nil || retained[1].dest != nil {
		t.Fatal("retired native frame retained roots")
	}
}
