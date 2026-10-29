package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestForwardedReturnFuelAndValues(t *testing.T) {
	types := []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.FuncRef, wasm.ExternRef}
	args := []value.Value{{Lo: 0xffffffff}, {Lo: 0xfedcba9876543210}, {Lo: 0x7f812345}, {Lo: 0x8000000000000000}, {Lo: 7, Hi: 19}, {}, {Lo: 99}}
	for _, explicit := range []bool{false, true} {
		body := module.Expr{}
		results := []wasm.ValueType{}
		for i := len(types) - 1; i >= 0; i-- {
			body = append(body, module.Instr{Op: 0x20, A: uint32(i), Offset: uint64(10 + len(body))})
			results = append(results, types[i])
		}
		cost := uint64(len(types))
		if explicit {
			body = append(body, module.Instr{Op: 0xf, Offset: 17})
			cost++
		}
		body = append(body, module.Instr{Op: 0xb, Offset: 18})
		s := state.New(state.DefaultLimits())
		in := instance(t, s, &module.Module{Types: []module.FuncType{{Params: types, Results: results}}, Functions: []module.Function{{Body: body}}}, nil)
		for fuel := uint64(0); fuel <= cost; fuel++ {
			e := New(s, nil, Config{Limits: DefaultLimits(), Metered: true, Fuel: fuel})
			if err := e.Start(in.Functions[0], args); err != nil {
				t.Fatal(err)
			}
			status := e.Run()
			if e.Fuel() != 0 {
				t.Fatal("fuel mismatch", explicit, fuel, e.Fuel())
			}
			if fuel < cost {
				var limit *wasm.Error
				if status != Failed || !errors.As(e.Err(), &limit) || !errors.Is(limit, wasm.ErrLimit) || limit.Offset != 10+fuel {
					t.Fatal("fuel boundary", explicit, fuel, status, e.Err())
				}
				continue
			}
			if status != Done || len(e.Results()) != len(args) {
				t.Fatal(status, e.Err(), e.Results())
			}
			for i, got := range e.Results() {
				if got != args[len(args)-1-i] {
					t.Fatal("return tuple", explicit, i, got)
				}
			}
		}
		s.Close()
	}
}
