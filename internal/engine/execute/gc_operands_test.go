package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// One reusable execution alternates operand counts across collections and
// errors. Scratch values are borrowed only by the synchronous checked helper.
func TestGCOperandWorkspaceCollectionFailureAndReuse(t *testing.T) {
	limits := state.DefaultLimits()
	limits.MaxHeapObjects = 2
	s := state.New(limits)
	defer s.Close()
	s.CollectEveryAllocation = true
	m := &module.Module{
		Types: []module.FuncType{{}, {Results: []wasm.ValueType{wasm.I32}}},
		TypeDefs: []module.TypeDef{
			{Kind: module.ArrayKind, Fields: []module.FieldType{{Value: wasm.I32}}},
			{Kind: module.FunctionKind},
		},
	}
	for _, count := range []uint32{1, 9} {
		body := module.Expr{}
		for i := uint32(0); i < count; i++ {
			body = append(body, module.Instr{Op: 0x41, Bits: uint64(i + 100)})
		}
		body = append(body, module.Instr{Op: 0xfb0008, A: 0, B: count}, module.Instr{Op: 0xfb000f}, module.Instr{Op: 0xb})
		m.Functions = append(m.Functions, module.Function{Type: 1, Body: body})
	}
	in := instanceWithFeatures(t, s, m, nil, wasm.Core3Features)
	e := New(s, nil, Config{Limits: DefaultLimits()})
	defer e.Close()
	checkScratch := func() {
		t.Helper()
		if len(e.gcArgs) != 0 {
			t.Fatal("retained operand length", len(e.gcArgs))
		}
		for _, v := range e.gcArgs[:cap(e.gcArgs)] {
			if v != (value.Value{}) {
				t.Fatal("retained operand value", v)
			}
		}
	}
	for _, index := range []int{0, 1, 0, 1} {
		if err := e.Start(in.Functions[index], nil); err != nil || e.Run() != Done {
			t.Fatal(err, e.Err())
		}
		want := uint64(1)
		if index == 1 {
			want = 9
		}
		if e.Results()[0].Lo != want {
			t.Fatal(e.Results(), want)
		}
		checkScratch()
	}
	if cap(e.gcArgs) != 9 || s.Usage().HeapObjects != 1 {
		t.Fatal("scratch growth or forced collection", cap(e.gcArgs), s.Usage())
	}
	// Exhaust the heap without a collector, then restore the same workspace.
	s.HeapSafepoint = func() {}
	for i := 0; i < 2; i++ {
		if err := e.Start(in.Functions[1], nil); err != nil {
			t.Fatal(err)
		}
		status := e.Run()
		if i == 0 && status != Done || i == 1 && (status != Failed || !errors.Is(e.Err(), wasm.ErrLimit)) {
			t.Fatal(status, e.Err())
		}
		checkScratch()
	}
	s.HeapSafepoint = nil
	s.CollectReferences(nil, []*state.Instance{in}, nil)
	if err := e.Start(in.Functions[0], nil); err != nil || e.Run() != Done || e.Results()[0].Lo != 1 {
		t.Fatal(err, e.Err(), e.Results())
	}
	checkScratch()
	e.Close()
	if e.gcArgs != nil {
		t.Fatal("closed workspace retained GC scratch")
	}
}
