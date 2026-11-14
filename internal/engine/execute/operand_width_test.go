package execute

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPortableScalarRoutingPreservesWholeValueCopies(t *testing.T) {
	for _, op := range []uint32{0x1b, 0x1c} {
		selectValue := module.Instr{Op: op}
		if op == 0x1c {
			selectValue.Types = []wasm.ValueType{wasm.V128}
		}
		s := state.New(state.DefaultLimits())
		defer s.Close()
		m := &module.Module{
			Types:   []module.FuncType{{Params: []wasm.ValueType{wasm.V128, wasm.V128, wasm.I32}, Results: []wasm.ValueType{wasm.V128, wasm.V128}}},
			Globals: []module.Global{{Type: module.GlobalType{Value: wasm.V128, Mutable: true}, Init: module.Expr{{Op: 0xfd000c}, {Op: 0xb}}}},
			Functions: []module.Function{{LocalCount: 1, Locals: []module.Local{{Count: 1, Type: wasm.V128}}, Body: module.Expr{
				{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20, A: 2}, selectValue,
				{Op: 0x22, A: 3}, {Op: 0x24}, {Op: 0x20, A: 3}, {Op: 0x23}, {Op: 0xb},
			}}},
		}
		in := instance(t, s, m, nil)
		for _, metered := range []bool{false, true} {
			e := New(s, nil, Config{Limits: DefaultLimits(), Metered: metered, Fuel: 100})
			defer e.Close()
			for _, condition := range []uint64{0, 1, 0} {
				args := []value.Value{{Lo: 0x123456789abcdef0, Hi: 0xfedcba9876543210}, {Lo: 0x1122334455667788, Hi: 0x8877665544332211}, {Lo: condition}}
				want := args[1]
				if condition != 0 {
					want = args[0]
				}
				if err := e.Start(in.Functions[0], args); err != nil || e.Run() != Done {
					t.Fatal(op, metered, err, e.Err())
				}
				if got := e.Results(); len(got) != 2 || got[0] != want || got[1] != want || in.Globals[0].Value != want {
					t.Fatal("select/local/global copy truncated a vector", op, metered, got, in.Globals[0].Value, want)
				}
			}
		}
	}
}
