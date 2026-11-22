package execute

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestAtomicConstantsRespectPollFuelAndMemoryEffects(t *testing.T) {
	m := simple(module.Expr{
		{Op: 0x41, Bits: 4}, {Op: 0x41, Bits: 17},
		{Op: 0xfe001e, A: 2, Offset: 700}, {Op: 0x1a},
		{Op: 0x41, Bits: 4}, {Op: 0xfe0010, A: 2}, {Op: 0xb},
	}, wasm.I32)
	m.Memories = []module.Limits{{Min: 1, Max: 1, HasMax: true, Shared: true}}
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instanceWithFeatures(t, s, m, nil, wasm.Core3Features|wasm.Threads)
	fn, _ := s.Function(in.Functions[0])
	code := fn.Code.Code
	if len(code) < 3 || code[0].Op != 0x41 || code[1].Op != 0x41 ||
		code[0].Count != 1 || code[1].Count != 1 || code[2].Op != 0xfe001e {
		t.Fatal("atomic constant prefix is absent", code)
	}
	tests := []struct {
		ctx     context.Context
		initial uint64
		fuel    uint64
		want    error
		steps   uint64
	}{
		{ctx: &cancelAfter{at: 2}, initial: 61, want: context.Canceled, steps: 64},
		{ctx: &cancelAfter{at: 2}, initial: 62, want: context.Canceled, steps: 64},
		{fuel: 2, want: wasm.ErrLimit, steps: 2},
		{},
	}
	for _, test := range tests {
		e := New(s, test.ctx, Config{Limits: DefaultLimits(), Metered: test.fuel != 0, Fuel: test.fuel})
		if err := e.Start(in.Functions[0], nil); err != nil {
			t.Fatal(err)
		}
		e.steps = test.initial
		status := e.Run()
		if test.want != nil {
			if status != Failed || !errors.Is(e.Err(), test.want) || e.steps != test.steps || e.Fuel() != 0 {
				t.Fatal(status, e.Err(), e.steps, e.Fuel())
			}
			got, trap := in.Memories[0].Load(0x28, 4, 0, 0, value.Value{})
			if got.Lo != 0 || trap != wasm.NoTrap {
				t.Fatal("atomic executed before an admission boundary", got, trap)
			}
		} else if status != Done || e.Results()[0].Lo != 17 {
			t.Fatal(status, e.Results(), e.Err())
		}
		e.Close()
	}
}
