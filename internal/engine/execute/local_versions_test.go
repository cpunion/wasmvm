package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLocalVersionsAcrossSuspensionAndFrameReuse(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.FuncRef, wasm.ExternRef} {
		t.Run(typ.String(), func(t *testing.T) {
			s := state.New(state.DefaultLimits())
			defer s.Close()
			host, err := s.NewHost(module.FuncType{}, "pause")
			if err != nil {
				t.Fatal(err)
			}
			m := &module.Module{
				Types:   []module.FuncType{{}, {Params: []wasm.ValueType{typ, typ}, Results: []wasm.ValueType{typ, typ, typ}}},
				Imports: []module.Import{{Kind: 0}},
				Functions: []module.Function{{Type: 1, LocalCount: 1, Locals: []module.Local{{Type: typ, Count: 1}}, Body: module.Expr{
					{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x21}, {Op: 0x10},
					{Op: 0x20}, {Op: 0x20, A: 2}, {Op: 0x20, A: 1}, {Op: 0x21, A: 2}, {Op: 0xb},
				}}},
			}
			in := instance(t, s, m, func(string, string) (state.Extern, bool) { return host, true })
			for _, metered := range []bool{false, true} {
				e := New(s, nil, Config{Limits: DefaultLimits(), Metered: metered, Fuel: 100})
				for round := uint64(1); round <= 2; round++ {
					args := []value.Value{{Lo: 17 * round}, {Lo: 31 * round}}
					if typ == wasm.V128 {
						args[0].Hi, args[1].Hi = ^round, ^(round + 1)
					}
					if err := e.Start(in.Functions[1], args); err != nil || e.Run() != HostRequest {
						t.Fatal(err, e.Err())
					}
					if typ.IsReference() {
						retained := map[uint64]bool{}
						e.RetainedValues(func(got wasm.ValueType, v value.Value) {
							if got != typ {
								t.Fatal("incorrect root type", got)
							}
							retained[v.Lo] = true
						})
						if !retained[args[0].Lo] || !retained[args[1].Lo] {
							t.Fatal("overwritten local lost live old root", retained)
						}
					}
					if err := e.ResumeHost(nil, nil); err != nil || e.Run() != Done {
						t.Fatal(err, e.Err())
					}
					want := []value.Value{args[0], args[1], {}}
					if len(e.Results()) != len(want) {
						t.Fatal("result arity changed", e.Results())
					}
					for i, got := range e.Results() {
						if got != want[i] {
							t.Fatal("local version or zero initialization changed", round, i, got, want[i])
						}
					}
					if metered && e.Fuel() != 92 {
						t.Fatal("copy fuel charge changed", e.Fuel())
					}
				}
				e.Close()
			}
			quota := DefaultLimits()
			quota.MaxSlots = 2 // The two parameters and additional local alone need three.
			e := New(s, nil, Config{Limits: quota})
			if err := e.Start(in.Functions[1], []value.Value{{}, {}}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("coalescing bypassed frame quota", err)
			}
			e.Close()
		})
	}
}
