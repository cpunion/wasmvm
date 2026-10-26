package ir

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func diamond() *Function {
	constant := Source{Wasm: module.Instr{Op: 0x41}, Outputs: []wasm.ValueType{wasm.I32}}
	return &Function{Signature: module.FuncType{Results: []wasm.ValueType{wasm.I32}}, Values: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32}, Sources: []Source{constant, {Wasm: module.Instr{Op: 4}}, constant, constant, {Wasm: module.Instr{Op: 0xb}}}, Blocks: []Block{
		{Ops: []Operation{{Source: 0, Results: []Value{0}}}, Term: Terminator{Kind: If, Source: 1, Args: []Value{0}, Edges: []Edge{{Target: 1}, {Target: 2}}}},
		{Ops: []Operation{{Source: 2, Results: []Value{1}}}, Term: Terminator{Kind: Jump, Source: 4, Edges: []Edge{{Target: 3, Args: []Value{1}}}}},
		{Ops: []Operation{{Source: 3, Results: []Value{2}}}, Term: Terminator{Kind: Jump, Source: 4, Edges: []Edge{{Target: 3, Args: []Value{2}}}}},
		{Params: []Value{3}, Term: Terminator{Kind: Return, Source: 4, Args: []Value{3}}},
	}}
}

func TestVerifyDiamondAndLoop(t *testing.T) {
	f := diamond()
	if err := Verify(f, 10000); err != nil {
		t.Fatal(err)
	}
	// Replace one arm with a loop whose tuple is defined by its own parameter.
	f.Sources = append(f.Sources, Source{Wasm: module.Instr{Op: 0xd}})
	f.Blocks[0].Term.Edges[0].Args = []Value{0}
	f.Blocks[1].Ops = nil
	f.Blocks[1].Params = []Value{1}
	f.Blocks[1].Term = Terminator{Kind: If, Source: 5, Args: []Value{1}, Edges: []Edge{{Target: 1, Args: []Value{1}}, {Target: 3, Args: []Value{1}}}}
	if err := Verify(f, 10000); err != nil {
		t.Fatal(err)
	}
	f.Blocks[1].Term.Kind = Table
	f.Sources[5].Wasm.Op = 0xe
	if err := Verify(f, 10000); err != nil {
		t.Fatal(err)
	}
	if err := Verify(f, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func TestRejectInvalidIR(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Function)
	}{
		{"no entry", func(f *Function) { f.Blocks = nil }},
		{"unknown value type", func(f *Function) { f.Values[0] = 0 }},
		{"incorrect effect", func(f *Function) { f.Sources[0].Effects = MayCallHost }},
		{"unknown operation type", func(f *Function) { f.Sources[0].Outputs = []wasm.ValueType{0} }},
		{"value index", func(f *Function) { f.Blocks[0].Ops[0].Results[0] = 100 }},
		{"multiple definitions", func(f *Function) { f.Blocks[2].Ops[0].Results[0] = 1 }},
		{"source index", func(f *Function) { f.Blocks[0].Ops[0].Source = 100 }},
		{"control operation", func(f *Function) { f.Blocks[0].Ops[0].Source = 1 }},
		{"operation arity", func(f *Function) { f.Blocks[0].Ops[0].Results = nil }},
		{"input type", func(f *Function) {
			f.Sources[2].Inputs = []wasm.ValueType{wasm.F32}
			f.Blocks[1].Ops[0].Args = []Value{0}
		}},
		{"result type", func(f *Function) { f.Sources[0].Outputs = []wasm.ValueType{wasm.F32} }},
		{"jump arity", func(f *Function) { f.Blocks[1].Term.Args = []Value{1} }},
		{"jump source", func(f *Function) { f.Blocks[1].Term.Source = 0 }},
		{"branch condition", func(f *Function) { f.Blocks[0].Term.Args = nil }},
		{"branch targets", func(f *Function) { f.Blocks[0].Term.Edges = f.Blocks[0].Term.Edges[:1] }},
		{"return arity", func(f *Function) { f.Blocks[3].Term.Args = nil }},
		{"return type", func(f *Function) { f.Signature.Results = []wasm.ValueType{wasm.F32} }},
		{"trap source", func(f *Function) { f.Blocks[3].Term = Terminator{Kind: Trap, Source: 4} }},
		{"missing terminator", func(f *Function) { f.Blocks[3].Term = Terminator{} }},
		{"block index", func(f *Function) { f.Blocks[0].Term.Edges[0].Target = 100 }},
		{"edge arity", func(f *Function) { f.Blocks[1].Term.Edges[0].Args = nil }},
		{"edge type", func(f *Function) { f.Values[1] = wasm.F32; f.Sources[2].Outputs = []wasm.ValueType{wasm.F32} }},
		{"entry predecessor", func(f *Function) { f.Blocks[0].Term.Edges[0].Target = 0 }},
		{"undefined value", func(f *Function) { f.Values = append(f.Values, wasm.I32) }},
		{"unreachable block", func(f *Function) {
			f.Blocks = append(f.Blocks, Block{Term: Terminator{Kind: Return, Source: 4, Args: []Value{3}}})
		}},
		{"use before definition", func(f *Function) {
			f.Sources[0].Inputs = []wasm.ValueType{wasm.I32}
			f.Blocks[0].Ops[0].Args = []Value{0}
		}},
		{"non-dominating definition", func(f *Function) { f.Blocks[2].Term.Edges[0].Args = []Value{1} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := diamond()
			tc.change(f)
			if err := Verify(f, 10000); !errors.Is(err, ErrInvalid) {
				t.Fatalf("accepted corrupt IR or wrong error: %v", err)
			}
		})
	}
	if err := Verify(nil, 10000); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}

func TestEffectsAndFuel(t *testing.T) {
	for _, tc := range []struct {
		op   uint32
		want Effects
	}{
		{0, MayTrap}, {0x20, ReadLocal}, {0x21, WriteLocal}, {0x23, ReadGlobal}, {0x24, WriteGlobal},
		{0x25, ReadTable | MayTrap}, {0x26, WriteTable | MayTrap}, {0x28, ReadMemory | MayTrap}, {0x36, WriteMemory | MayTrap},
		{0x3f, ReadMemory}, {0x40, ReadMemory | WriteMemory | MayAllocate}, {0x6d, MayTrap}, {0xa8, MayTrap},
		{0xfc0008, ReadSegment | WriteMemory | MayTrap}, {0xfc0009, WriteSegment}, {0xfc000a, ReadMemory | WriteMemory | MayTrap},
		{0xfc000b, WriteMemory | MayTrap}, {0xfc000c, ReadSegment | WriteTable | MayTrap}, {0xfc000d, WriteSegment},
		{0xfc000e, ReadTable | WriteTable | MayTrap}, {0xfc000f, ReadTable | WriteTable | MayAllocate}, {0xfc0010, ReadTable}, {0xfc0011, WriteTable | MayTrap},
		{0xfd000b, WriteMemory | MayTrap}, {0xfd0058, WriteMemory | MayTrap}, {0xfd0000, ReadMemory | MayTrap}, {0xfd0054, ReadMemory | MayTrap}, {0xfd005c, ReadMemory | MayTrap}, {0xfd000c, 0}, {0x41, 0},
	} {
		if got := EffectsOf(tc.op); got != tc.want {
			t.Errorf("%x: effects %x, want %x", tc.op, got, tc.want)
		}
	}
	if EffectsOf(0x10) != EffectsOf(0x11) || EffectsOf(0x10)&(MayCallHost|MaySuspend) != (MayCallHost|MaySuspend) {
		t.Fatal("call effects")
	}
	for _, op := range []uint32{0, 1, 2, 3, 4, 0xc, 0xd, 0xe, 0xf, 0x41, 0xfd007b} {
		if (Source{Wasm: module.Instr{Op: op}}).Fuel() != 1 {
			t.Fatal("instruction fuel")
		}
	}
	for _, op := range []uint32{5, 0xb} {
		if (Source{Wasm: module.Instr{Op: op}}).Fuel() != 0 {
			t.Fatal("delimiter fuel")
		}
	}
	f := &Function{Sources: []Source{{Wasm: module.Instr{Op: 0}, Effects: MayTrap}}, Blocks: []Block{{Term: Terminator{Kind: Trap}}}}
	if err := Verify(f, 100); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionTypesAndSharedBudget(t *testing.T) {
	f := diamond()
	f.Signature.Params = []wasm.ValueType{wasm.I64}
	if err := Verify(f, 10000); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	f.Locals = []wasm.ValueType{wasm.F64}
	if err := Verify(f, 10000); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	f.Locals[0] = wasm.I64
	budget := VerifyBudget{Remaining: 10000}
	if err := budget.Verify(f); err != nil {
		t.Fatal(err)
	}
	used := 10000 - budget.Remaining
	budget.Remaining = used
	if err := budget.Verify(f); err != nil || budget.Remaining != 0 {
		t.Fatal(budget, err)
	}
	if err := budget.Verify(f); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	f.Signature.Params[0] = 0
	if err := Verify(f, 10000); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
