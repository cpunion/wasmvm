package slots

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCallVeneerLoweringRejectsComputedWork(t *testing.T) {
	for _, name := range []string{"direct", "indirect", "return", "duplicate", "before", "after", "local", "dynamic"} {
		t.Run(name, func(t *testing.T) {
			body := module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x10}, {Op: 0xb}}
			want := true
			switch name {
			case "indirect":
				body = module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x41}, {Op: 0x11}, {Op: 0xb}}
			case "return":
				body = module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x10}, {Op: 0xf}, {Op: 0xb}}
			case "duplicate":
				body[1].A = 0
			case "before":
				body = module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x7c}, {Op: 0x20, A: 1}, {Op: 0x10}, {Op: 0xb}}
				want = false
			case "after":
				body = module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x10}, {Op: 0x42, Bits: 1}, {Op: 0x7c}, {Op: 0xb}}
				want = false
			case "local":
				body = append(module.Expr{{Op: 0x42, Bits: 1}, {Op: 0x21}}, body...)
				want = false
			case "dynamic":
				body = module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20}, {Op: 0xa7}, {Op: 0x11}, {Op: 0xb}}
				want = false
			}
			d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I64, wasm.I64}, Results: []wasm.ValueType{wasm.I64}}},
				Imports: []module.Import{{Kind: 0}}, Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}},
				Functions: []module.Function{{Body: body}}}
			m, err := Compile(makeIR(t, d), DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			plan := m.Functions[0].Veneer
			if (plan != nil) != want {
				t.Fatal("incorrect forwarding classification", plan)
			}
			if plan != nil {
				callFuel, returnFuel := uint64(3), uint64(0)
				if name == "indirect" {
					callFuel++
				}
				if name == "return" {
					returnFuel++
				}
				if plan.CallFuel != callFuel || plan.ReturnFuel != returnFuel {
					t.Fatal("changed instruction fuel", plan)
				}
			}
		})
	}
}
