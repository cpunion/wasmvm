package ir

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestExceptionalEdgesAndSchemas(t *testing.T) {
	makeFunction := func() *Function {
		return &Function{
			Signature: module.FuncType{Results: []wasm.ValueType{wasm.ExnRef}},
			Values:    []wasm.ValueType{wasm.ExnRef},
			Sources:   []Source{{Wasm: module.Instr{Op: 8}, Effects: EffectsOf(8)}, {Wasm: module.Instr{Op: 0xb}}},
			Handlers:  []Handler{{Catches: []Catch{{Kind: 3, Target: 1, Types: []wasm.ValueType{wasm.ExnRef.NonNull()}}}}},
			Blocks:    []Block{{Handler: 1, Term: Terminator{Kind: Throw, Source: 0}}, {Params: []Value{0}, Term: Terminator{Kind: Return, Source: 1, Args: []Value{0}}}},
		}
	}
	if err := Verify(makeFunction(), 1000); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Function){
		func(f *Function) { f.Handlers[0].Parent = 1 },
		func(f *Function) { f.Handlers[0].Catches = nil },
		func(f *Function) { f.Handlers[0].Catches[0].Kind = 4 },
		func(f *Function) { f.Handlers[0].Catches[0].Target = 2 },
		func(f *Function) { f.Handlers[0].Catches[0].Types = nil },
		func(f *Function) { f.Handlers[0].Catches[0].Types[0] = wasm.I32 },
		func(f *Function) { f.Handlers[0].Catches[0].Types[0] = wasm.ExnRef },
		func(f *Function) { f.Handlers[0].Catches[0].Kind = 2 },
		func(f *Function) { f.Blocks[0].Handler = 2 },
		func(f *Function) { f.Blocks[0].Term.Edges = []Edge{{Target: 1}} },
		func(f *Function) { f.Sources[0].Inputs = []wasm.ValueType{wasm.I32} },
	} {
		f := makeFunction()
		change(f)
		if err := Verify(f, 1000); err == nil {
			t.Fatal("invalid handler/throw accepted")
		}
	}
}
