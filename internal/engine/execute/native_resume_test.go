package execute

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
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

func TestCapturedContinuationsDiscardBackendResumeState(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}, {}}, TypeDefs: []module.TypeDef{{Kind: module.FunctionKind}, {Kind: module.ContinuationKind}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}}}}
	in := instanceWithFeatures(t, s, d, nil, wasm.Core3Features|wasm.StackSwitching)
	e := New(s, nil, Config{Limits: DefaultLimits()})
	defer e.Close()
	if err := e.Start(in.Functions[0], nil); err != nil {
		t.Fatal(err)
	}
	e.frames[0].nativeResume = 123
	e.prompts = []prompt{{instance: in}}
	ref, err := e.capture(0, in.Types[1], nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := s.Continuation(ref, false)
	if err != nil {
		t.Fatal(err)
	}
	body := raw.(*continuationBody)
	if body.frames[0].nativeResume != 0 || e.frames[0].nativeResume != 123 {
		t.Fatal("captured backend state or mutated live source")
	}
	if unsafe.Sizeof(frame{}) > 64 || body.HeapBytes() < 96+uint64(len(body.frames))*uint64(unsafe.Sizeof(frame{}))+uint64(len(body.values))*uint64(unsafe.Sizeof(value.Value{})) {
		t.Fatal("captured frame storage exceeds its byte charge")
	}
	// Restoration must also discard optional metadata from an older carrier.
	body.frames[0].nativeResume = 456
	if err := e.restore(body, nil, []slots.Slot{0}, nil); err != nil || e.frames[len(e.frames)-1].nativeResume != 0 {
		t.Fatal("backend state crossed execution restoration", err)
	}
}
