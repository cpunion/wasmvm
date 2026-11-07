package execute

import (
	"context"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type interruptFrameBackend struct {
	batchFrameBackend
	interrupts int
}

func (b *interruptFrameBackend) Interrupt() { b.interrupts++ }

func TestBackendFrameContractSurvivesWorkspacePolicyChanges(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{
		{Body: module.Expr{{Op: 0x20}, {Op: 0x10, A: 1}, {Op: 0x20}, {Op: 0x6a}, {Op: 0xb}}},
		{Body: module.Expr{{Op: 0x20}, {Op: 0xb}}},
	}}, nil)
	b := &interruptFrameBackend{batchFrameBackend: batchFrameBackend{store: s}}
	e := New(s, nil, Config{Limits: DefaultLimits(), Backend: b})
	defer e.Close()
	for i := 0; i < 4; i++ {
		limits := Limits{MaxFrames: uint32(3 + i), MaxSlots: uint32(100 + i)}
		b.metered = i%2 != 0
		if err := e.Rebind(context.Background(), limits, b.metered, 100); err != nil {
			t.Fatal(err)
		}
		e.Interrupt()
		arg := uint64(7 + i)
		if err := e.Start(in.Functions[0], []value.Value{{Lo: arg}}); err != nil || e.Run() != Done || e.Err() != nil {
			t.Fatal(i, err, e.Err())
		}
		if b.batches != i+1 || b.interrupts != i+1 || b.limits != limits || len(e.Results()) != 1 || e.Results()[0].Lo != 2*arg {
			t.Fatal("backend identity or invocation policy changed", i, b.batches, b.interrupts, b.limits, e.Results())
		}
		e.Quiesce()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := e.Rebind(ctx, DefaultLimits(), false, 0); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(in.Functions[0], []value.Value{{Lo: 9}}); err != nil || e.Run() != Failed || e.Err() != context.Canceled || b.batches != 4 {
		t.Fatal("canceled invocation entered backend", err, e.Err(), b.batches)
	}
}

func TestBackendFrameModelAndPortableDispatchRemainReusable(t *testing.T) {
	for _, backend := range []Backend{nil, &movingFrameBackend{}} {
		s := state.New(state.DefaultLimits())
		in := instance(t, s, &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x41, Bits: 19}, {Op: 0xb}}}}}, nil)
		e := New(s, nil, Config{Limits: DefaultLimits(), Backend: backend})
		for i := 0; i < 2; i++ {
			e.Interrupt()
			if err := e.Rebind(nil, DefaultLimits(), false, 0); err != nil {
				t.Fatal(err)
			}
			if err := e.Start(in.Functions[0], nil); err != nil || e.Run() != Done || len(e.Results()) != 1 || e.Results()[0].Lo != 19 {
				t.Fatal(err, e.Err(), e.Results())
			}
			e.Quiesce()
		}
		e.Close()
		s.Close()
	}
}
