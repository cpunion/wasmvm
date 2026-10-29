package ir

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTailTerminatorVerifiesCalleeSchema(t *testing.T) {
	makeFunction := func() *Function {
		return &Function{
			Signature: module.FuncType{Results: []wasm.ValueType{wasm.I32}},
			Values:    []wasm.ValueType{wasm.I32},
			Sources: []Source{
				{Wasm: module.Instr{Op: 0x41}, Outputs: []wasm.ValueType{wasm.I32}},
				{Wasm: module.Instr{Op: 0x12, A: 1}, Inputs: []wasm.ValueType{wasm.I32}, Outputs: []wasm.ValueType{wasm.I32}, Effects: EffectsOf(0x12)},
			},
			Blocks: []Block{{Ops: []Operation{{Source: 0, Results: []Value{0}}}, Term: Terminator{Kind: TailCall, Source: 1, Args: []Value{0}}}},
		}
	}
	if err := (&VerifyBudget{Remaining: 1000}).Verify(makeFunction()); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Function){
		func(f *Function) { f.Signature.Results[0] = wasm.I64 },
		func(f *Function) { f.Sources[1].Outputs = nil },
		func(f *Function) { f.Sources[1].Inputs[0] = wasm.I64 },
		func(f *Function) { f.Sources[1].Wasm.Op = 0x10 },
		func(f *Function) { f.Blocks[0].Term.Edges = []Edge{{Target: 0}} },
	} {
		f := makeFunction()
		change(f)
		if err := (&VerifyBudget{Remaining: 1000}).Verify(f); err == nil {
			t.Fatal("invalid tail schema accepted")
		}
	}
}
