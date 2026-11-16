package nativeexec

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestResultPrefixesBeyondImmediateRangePreserveCompleteTuples(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	const count = 1025
	types := make([]wasm.ValueType, count)
	args := make([]value.Value, count)
	var caller, callee module.Expr
	for i := range count {
		types[i] = wasm.I64
		args[i] = value.Value{Lo: uint64(i*17 + 1), Hi: ^uint64(i*31 + 7)}
		caller = append(caller, module.Instr{Op: 0x20, A: uint32(i)})
		callee = append(callee, module.Instr{Op: 0x20, A: uint32(count - 1 - i)})
	}
	caller = append(caller, module.Instr{Op: 0x10, A: 1}, module.Instr{Op: 0xb})
	callee = append(callee, module.Instr{Op: 0xb})
	m := callSlots(t, &module.Module{Types: []module.FuncType{{Params: types, Results: types}}, Functions: []module.Function{{Body: caller}, {Body: callee}}})
	for _, metered := range []bool{false, true} {
		s := state.NewNative(state.DefaultLimits())
		t.Cleanup(s.Close)
		in, err := s.Allocate(m, nil)
		if err != nil {
			t.Fatal(err)
		}
		e, n := callExecution(t, s, m, in, metered, execute.Limits{MaxFrames: 128, MaxSlots: 16384}, 16384)
		for range 2 {
			if err := e.Start(in.Functions[0], args); err != nil {
				t.Fatal(err)
			}
			e.SetFuel(10000)
			if status := e.Run(); status != execute.Done {
				t.Fatal("large prefix return", metered, status, e.Err())
			}
			got := e.Results()
			if len(got) != count {
				t.Fatal("incomplete result tuple", len(got))
			}
			for i, result := range got {
				if result != args[count-1-i] {
					t.Fatal("result prefix lost value words", metered, i, result, args[count-1-i])
				}
			}
			if len(n.callWords) == 0 || n.callWords[nativeabi.CallHeaderBytes/8+nativeabi.CallResumeOffset/8] == 0 {
				t.Fatal("large tuple bypassed private native calls")
			}
		}
	}
}
