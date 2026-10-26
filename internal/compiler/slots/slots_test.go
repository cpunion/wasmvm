package slots

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompleteCore2Slots(t *testing.T) {
	scripts, modules, functions := 0, 0, 0
	var values, slots uint64
	err := filepath.WalkDir("../../../artifacts/core2/fixtures", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "script.json" {
			return nil
		}
		scripts++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var script struct {
			Commands []struct{ Type, Filename string }
		}
		if err := json.Unmarshal(data, &script); err != nil {
			return err
		}
		for _, cmd := range script.Commands {
			if cmd.Filename == "" || cmd.Type == "assert_malformed" || cmd.Type == "assert_invalid" {
				continue
			}
			name := filepath.Join(filepath.Dir(path), cmd.Filename)
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			m, err := decode.Decode(data, decode.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			checked, err := validate.Check(m, validate.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			lowered, err := lower.Compile(checked, lower.DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			compiled, err := Compile(lowered, DefaultLimits())
			if err != nil {
				t.Fatal(name, err)
			}
			if compiled.IR != lowered || len(compiled.Functions) != len(lowered.Functions) {
				t.Fatal("module identity")
			}
			for i, f := range compiled.Functions {
				checkSymbolic(t, name, lowered.Functions[i], f)
				functions++
				values += uint64(len(lowered.Functions[i].Values))
				slots += uint64(len(f.Slots)) - uint64(f.LocalCount)
			}
			modules++
		}
		return nil
	})
	if err != nil {
		t.Fatal("prepare pinned corpus:", err)
	}
	if scripts != 148 || modules != 1716 || functions != 6236 {
		t.Fatal(scripts, modules, functions)
	}
	t.Logf("Symbolically checked %d functions in %d modules: %d SSA values use %d temporary slots (including edge scratch); no guest execution assertions run.", functions, modules, values, slots)
}

// Interpret the allocator's output using symbolic SSA identities, independently
// of instruction semantics. Every operand must still contain its original
// value, and every selected edge must implement simultaneous assignment.
func checkSymbolic(t *testing.T, name string, f *ir.Function, out *Function) {
	t.Helper()
	for bi, b := range f.Blocks {
		state := make([]int, len(out.Slots))
		for i := range state {
			state[i] = -1
		}
		// Definitions from other blocks and current block parameters are live-in.
		for bj, other := range f.Blocks {
			for _, v := range other.Params {
				state[out.ValueSlots[v]] = int(v)
			}
			if bj != bi {
				for _, op := range other.Ops {
					for _, v := range op.Results {
						state[out.ValueSlots[v]] = int(v)
					}
				}
			}
		}
		// Block-local definitions may share storage with other blocks. They will be
		// overwritten before use; cross-block values have dedicated storage.
		expect := func(s Slot, v ir.Value) {
			if int(s) >= len(state) || state[s] != int(v) || out.Slots[s] != f.Values[v] {
				t.Fatalf("%s block %d: slot %d = %d, want value %d (%s)", name, bi, s, state[s], v, f.Values[v])
			}
		}
		for oi, op := range b.Ops {
			in := out.Code[int(out.BlockPC[bi])+oi]
			if in.Source != op.Source || in.Op != f.Sources[op.Source].Wasm.Op {
				t.Fatal("source mapping")
			}
			if in.Op == 0x10 || in.Op == 0x11 {
				if int(in.Count) != len(op.Args) {
					t.Fatal("call argument count")
				}
				for i, v := range op.Args {
					expect(out.Args[int(in.Extra)+i], v)
				}
				for i, v := range op.Results {
					s := out.Args[int(in.Extra)+len(op.Args)+i]
					state[s] = int(v)
				}
			} else {
				args := [3]Slot{in.A, in.B, in.C}
				for i, v := range op.Args {
					expect(args[i], v)
				}
				for _, v := range op.Results {
					state[in.Dst] = int(v)
				}
			}
		}
		term := out.Code[int(out.BlockPC[bi])+len(b.Ops)]
		if term.Op != ControlBase+uint32(b.Term.Kind) || term.Source != b.Term.Source {
			t.Fatal("terminator mapping")
		}
		if b.Term.Kind == ir.Return {
			for i, v := range b.Term.Args {
				expect(out.Args[int(term.Extra)+i], v)
			}
		} else if len(b.Term.Args) > 0 {
			expect(term.A, b.Term.Args[0])
		}
		if len(b.Term.Edges) > 0 && int(term.Count) != len(b.Term.Edges) {
			t.Fatal("edge count")
		}
		for ei, e := range b.Term.Edges {
			for _, v := range e.Args {
				expect(out.ValueSlots[v], v)
			}
			edge := out.Edges[int(term.Extra)+ei]
			if edge.PC != out.BlockPC[e.Target] {
				t.Fatal("branch target")
			}
			copied := append([]int(nil), state...)
			for _, move := range out.Moves[edge.Start : edge.Start+edge.Count] {
				if out.Slots[move.Dst] != out.Slots[move.Src] {
					t.Fatal("move type")
				}
				copied[move.Dst] = copied[move.Src]
			}
			for i, v := range f.Blocks[e.Target].Params {
				if copied[out.ValueSlots[v]] != int(e.Args[i]) {
					t.Fatalf("%s block %d edge %d: wrong parallel tuple", name, bi, ei)
				}
			}
		}
	}
}

func TestParallelMoves(t *testing.T) {
	rng := rand.New(rand.NewSource(20261026))
	for n := 0; n < 100; n++ {
		for round := 0; round < 100; round++ {
			input := make([]Move, n)
			before := make([]int, n+1)
			for i := range input {
				input[i] = Move{Slot(i), Slot(rng.Intn(n + 1))}
				before[i] = i + 100
			}
			before[n] = 999
			// Use a separate temporary, never the extra external input slot.
			state := append(append([]int(nil), before...), -1)
			ordered := parallelMoves(input, func(Slot) Slot { return Slot(n + 1) })
			for _, m := range ordered {
				state[m.Dst] = state[m.Src]
			}
			for _, m := range input {
				if state[m.Dst] != before[m.Src] {
					t.Fatal(input, ordered, state)
				}
			}
		}
	}
}

func makeIR(t *testing.T, m *module.Module) *lower.Module {
	t.Helper()
	checked, err := validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Compile(checked, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return lowered
}
func TestReuseAndLimits(t *testing.T) {
	body := module.Expr{{Op: 0x41}}
	for i := 0; i < 1000; i++ {
		body = append(body, module.Instr{Op: 0x41}, module.Instr{Op: 0x6a})
	}
	body = append(body, module.Instr{Op: 0xb})
	m := makeIR(t, &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{{Body: body}}})
	compiled, err := Compile(m, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	f := compiled.Functions[0]
	// Three local arithmetic temporaries plus the dedicated return block value.
	if len(f.Slots) != 4 {
		t.Fatalf("unbounded temporary growth: %d", len(f.Slots))
	}
	checkSymbolic(t, "reuse", m.Functions[0], f)
	for _, limits := range []Limits{{3, 100000, 100000}, {100, 0, 100000}, {100, 100000, 0}} {
		if got, err := Compile(m, limits); got != nil || !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(got, err)
		}
	}
	if got, err := Compile(nil, DefaultLimits()); got != nil || !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(got, err)
	}
	if got, err := Compile(&lower.Module{Functions: []*ir.Function{{}}}, DefaultLimits()); got != nil || !errors.Is(err, ir.ErrInvalid) {
		t.Fatal(got, err)
	}
}

func TestTypedCyclicEdges(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.V128, wasm.FuncRef, wasm.ExternRef} {
		constant := uint32(0x41)
		if typ.IsReference() {
			constant = 0xd0
		} else if typ == wasm.V128 {
			constant = 0xfd000c
		}
		f := &ir.Function{
			Signature: module.FuncType{Results: []wasm.ValueType{typ, typ}},
			Values:    []wasm.ValueType{typ, typ, wasm.I32, typ, typ, typ, typ},
			Sources: []ir.Source{
				{Wasm: module.Instr{Op: constant}, Outputs: []wasm.ValueType{typ}},
				{Wasm: module.Instr{Op: 0x41}, Outputs: []wasm.ValueType{wasm.I32}},
				{Wasm: module.Instr{Op: 3}}, {Wasm: module.Instr{Op: 0xe}}, {Wasm: module.Instr{Op: 0xb}},
			},
			Blocks: []ir.Block{
				{Ops: []ir.Operation{{Source: 0, Results: []ir.Value{0}}, {Source: 0, Results: []ir.Value{1}}, {Source: 1, Results: []ir.Value{2}}}, Term: ir.Terminator{Kind: ir.Jump, Source: 2, Edges: []ir.Edge{{Target: 1, Args: []ir.Value{0, 1}}}}},
				{Params: []ir.Value{3, 4}, Term: ir.Terminator{Kind: ir.Table, Source: 3, Args: []ir.Value{2}, Edges: []ir.Edge{{Target: 1, Args: []ir.Value{4, 3}}, {Target: 1, Args: []ir.Value{4, 3}}, {Target: 2, Args: []ir.Value{3, 4}}}}},
				{Params: []ir.Value{5, 6}, Term: ir.Terminator{Kind: ir.Return, Source: 4, Args: []ir.Value{5, 6}}},
			},
		}
		compiled, err := Compile(&lower.Module{Functions: []*ir.Function{f}}, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		out := compiled.Functions[0]
		if len(out.Slots) != 8 || out.Slots[7] != typ {
			t.Fatalf("scratch type/reuse: %v", out.Slots)
		}
		checkSymbolic(t, "cyclic edge", f, out)
	}
}

func FuzzSlots(f *testing.F) {
	f.Add([]byte("\x00asm\x01\x00\x00\x00"))
	f.Add([]byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 4, 1, 0x60, 0, 0, 3, 2, 1, 0, 10, 7, 1, 5, 0, 0x41, 1, 0x1a, 0xb})
	f.Fuzz(func(t *testing.T, data []byte) {
		limits := decode.DefaultLimits()
		limits.MaxBytes = 1 << 16
		limits.MaxNodes = 4096
		limits.MaxLocals = 4096
		limits.MaxNesting = 64
		m, err := decode.Decode(data, limits)
		if err != nil {
			return
		}
		checked, err := validate.Check(m, validate.Limits{MaxLocals: 4096, MaxStack: 4096, MaxControl: 65})
		if err != nil {
			return
		}
		lowered, err := lower.Compile(checked, lower.Limits{MaxNodes: 1 << 16, MaxVerifyWork: 1 << 20})
		if err != nil {
			return
		}
		compiled, err := Compile(lowered, Limits{4096, 1 << 16, 1 << 20})
		if err != nil {
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
			return
		}
		for i, fn := range compiled.Functions {
			checkSymbolic(t, "fuzz", lowered.Functions[i], fn)
		}
	})
}
