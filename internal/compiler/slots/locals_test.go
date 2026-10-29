package slots

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLocalValueVersions(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.FuncRef, wasm.ExternRef} {
		for _, test := range []struct {
			name    string
			body    module.Expr
			results int
			aliases []bool // local.get and local.tee results, in source order
		}{
			{"repeated reads", module.Expr{{Op: 0x20}, {Op: 0x20}, {Op: 0xb}}, 2, []bool{true, true}},
			{"old read across write", module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x21}, {Op: 0xb}}, 1, []bool{false, true}},
			{"tee result", module.Expr{{Op: 0x20, A: 1}, {Op: 0x22}, {Op: 0xb}}, 1, []bool{true, true}},
			{"old tee across write", module.Expr{{Op: 0x20, A: 1}, {Op: 0x22}, {Op: 0x20, A: 1}, {Op: 0x21}, {Op: 0xb}}, 1, []bool{true, false, true}},
			{"self write boundary", module.Expr{{Op: 0x20}, {Op: 0x21}, {Op: 0x20}, {Op: 0xb}}, 1, []bool{false, true}},
			{"cross block read", module.Expr{{Op: 0x20}, {Op: 2, Block: -64}, {Op: 0x20, A: 1}, {Op: 0x21}, {Op: 0xb}, {Op: 0xb}}, 1, []bool{false, true}},
			{"block write epochs", module.Expr{{Op: 2, Block: -64}, {Op: 0x20, A: 1}, {Op: 0x21}, {Op: 0xb}, {Op: 0x20}, {Op: 0xb}}, 1, []bool{true, true}},
		} {
			t.Run(typ.String()+"/"+test.name, func(t *testing.T) {
				sig := module.FuncType{Params: []wasm.ValueType{typ, typ}}
				for i := 0; i < test.results; i++ {
					sig.Results = append(sig.Results, typ)
				}
				m := makeIR(t, &module.Module{Types: []module.FuncType{sig}, Functions: []module.Function{{Body: test.body}}})
				compiled, err := Compile(m, DefaultLimits())
				if err != nil {
					t.Fatal(err)
				}
				f, out := m.Functions[0], compiled.Functions[0]
				checkSymbolic(t, test.name, f, out)
				reads := 0
				for bi, b := range f.Blocks {
					for oi, op := range b.Ops {
						src := f.Sources[op.Source].Wasm
						in := out.Code[int(out.BlockPC[bi])+oi]
						if in.Op != src.Op || in.Source != op.Source {
							t.Fatal("original instruction/fuel boundary changed", in, src)
						}
						if src.Op == 0x20 || src.Op == 0x22 {
							alias := out.ValueSlots[op.Results[0]] == Slot(src.A)
							if reads >= len(test.aliases) || alias != test.aliases[reads] {
								t.Fatal("incorrect local lifetime", reads, alias, out.ValueSlots)
							}
							reads++
						}
					}
				}
				if reads != len(test.aliases) {
					t.Fatal("missing local operations", reads)
				}
				limits := DefaultLimits()
				limits.MaxSlots = uint32(len(out.Slots))
				if _, err := Compile(m, limits); err != nil {
					t.Fatal("exact frame quota", err)
				}
				limits.MaxSlots--
				if got, err := Compile(m, limits); got != nil || !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("frame quota returned partial artifact", got, err)
				}
			})
		}
	}
}
