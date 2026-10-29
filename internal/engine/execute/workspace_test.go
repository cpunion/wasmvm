package execute

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestQuiesceRebindPolicyAndBorrowedResults(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, simple(module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}, wasm.I32), nil)
	e := New(s, nil, Config{Limits: DefaultLimits()})
	defer e.Close()
	if err := e.Start(in.Functions[0], nil); err != nil {
		t.Fatal(err)
	}
	if err := e.Rebind(nil, DefaultLimits(), true, 1); err == nil {
		t.Fatal("rebound an active frame")
	}
	if e.Run() != Done {
		t.Fatal(e.Err())
	}
	results := e.Results()
	e.Quiesce()
	if results[0].Lo != 42 || len(e.frames) != 0 || e.pending != nil || e.context != context.Background() {
		t.Fatal("quiescence lost results or retained invocation roots")
	}
	for _, limits := range []Limits{{}, DefaultLimits()} {
		if err := e.Rebind(nil, limits, true, 1); err != nil {
			t.Fatal(err)
		}
		err := e.Start(in.Functions[0], nil)
		if limits.MaxFrames == 0 {
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
		} else if err != nil || e.Run() != Done || e.Results()[0].Lo != 42 || e.Fuel() != 0 {
			t.Fatal(err, e.Err(), e.Results())
		}
		e.Quiesce()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := e.Rebind(ctx, DefaultLimits(), false, 0); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(in.Functions[0], nil); err != nil || e.Run() != Failed || !errors.Is(e.Err(), context.Canceled) {
		t.Fatal(err, e.Err())
	}
	e.Quiesce()
	host, err := s.NewHost(module.FuncType{}, "host")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Start(host.Function, nil); err != nil || e.Run() != HostRequest {
		t.Fatal(err)
	}
	if err := e.Rebind(nil, DefaultLimits(), false, 0); err == nil {
		t.Fatal("rebound a pending host request")
	}
	e.Quiesce()
	if e.pending != nil || e.Footprint() != (Limits{}) {
		t.Fatal("pending request survived suspension")
	}
}
