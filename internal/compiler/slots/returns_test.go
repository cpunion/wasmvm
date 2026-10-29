package slots

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestForwardReturnsRespectControlAndTuple(t *testing.T) {
	for _, kind := range []string{"forward", "backward", "reordered", "effects", "subset"} {
		t.Run(kind, func(t *testing.T) {
			f := &ir.Function{
				Signature: module.FuncType{Results: []wasm.ValueType{wasm.I32, wasm.I32}},
				Values:    []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32},
				Sources: []ir.Source{
					{Wasm: module.Instr{Op: 0x41}, Outputs: []wasm.ValueType{wasm.I32}},
					{Wasm: module.Instr{Op: 1}}, {Wasm: module.Instr{Op: 0xc}}, {Wasm: module.Instr{Op: 0xb}},
				},
				Blocks: []ir.Block{
					{Ops: []ir.Operation{{Source: 0, Results: []ir.Value{0}}, {Source: 0, Results: []ir.Value{1}}}, Term: ir.Terminator{Kind: ir.Jump, Source: 2, Edges: []ir.Edge{{Target: 1, Args: []ir.Value{1, 0}}}}},
					{Params: []ir.Value{2, 3}, Term: ir.Terminator{Kind: ir.Return, Source: 3, Args: []ir.Value{2, 3}}},
				},
			}
			from := 0
			switch kind {
			case "backward":
				body := f.Blocks[0]
				f.Blocks[0] = ir.Block{Term: ir.Terminator{Kind: ir.Jump, Source: 2, Edges: []ir.Edge{{Target: 2}}}}
				f.Blocks = append(f.Blocks, body)
				from = 2
			case "reordered":
				f.Blocks[1].Term.Args = []ir.Value{3, 2}
			case "effects":
				f.Blocks[1].Ops = []ir.Operation{{Source: 1}}
			case "subset":
				f.Signature.Results = f.Signature.Results[:1]
				f.Blocks[1].Term.Args = []ir.Value{2}
			}
			compiled, err := Compile(&lower.Module{Functions: []*ir.Function{f}}, DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			out := compiled.Functions[0]
			term := out.Code[int(out.BlockPC[from])+len(f.Blocks[from].Ops)]
			forwarded := term.Op == ControlBase+uint32(ir.Return)
			if forwarded != (kind == "forward") || term.Source != 2 {
				t.Fatal("incorrect return forwarding", term)
			}
			checkSymbolic(t, kind, f, out)
		})
	}
}
