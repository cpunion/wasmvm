package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type scalarExitBackend struct{ movingFrameBackend }

func (*scalarExitBackend) RunFrames(*state.NativeFrames) error { return nil }

func TestNativeRuntimeRejectsScalarFallback(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, &module.Module{
		Types:     []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}},
		Functions: []module.Function{{Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}}},
	}, nil)
	e := New(s, nil, Config{Limits: DefaultLimits(), Backend: &scalarExitBackend{}})
	defer e.Close()
	if err := e.Start(in.Functions[0], nil); err != nil {
		t.Fatal(err)
	}
	if e.Run() != Failed || !errors.Is(e.Err(), wasm.ErrInvalid) || len(e.Results()) != 0 || len(e.frames) != 0 || len(e.values) != 0 {
		t.Fatal("interpreted ordinary work after an invalid native exit", e.Err(), e.Results())
	}
}
