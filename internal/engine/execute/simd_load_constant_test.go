package execute

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSIMDLoadConstantBinaryPreservesObservations(t *testing.T) {
	m := simple(module.Expr{
		{Op: 0x20},
		{Op: 0xfd0000, Offset: 701},
		{Op: 0xfd000c, Vector: [16]byte{0: 1, 4: 3, 8: 5, 12: 7}},
		{Op: 0xfd00b1}, // i32x4.sub
		{Op: 0xb},
	}, wasm.V128)
	m.Types[0].Params = []wasm.ValueType{wasm.I32}
	m.Memories = []module.Limits{{Min: 1, Max: 1, HasMax: true}}
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instanceWithFeatures(t, s, m, nil, wasm.Core3Features)
	fn, _ := s.Function(in.Functions[0])
	load, constant := fn.Code.Code[1], fn.Code.Code[2]
	if load.Op != 0xfd0000 || load.Count != slots.SIMDLoadConstantBinary || constant.Op != 0xfd000c {
		t.Fatal("load group was not selected", fn.Code.Code)
	}
	loaded := value.Value{Lo: 0x0000000200000000, Hi: 0x0000000900000004}
	literal := value.Value{Lo: 0x0000000300000001, Hi: 0x0000000700000005}
	if trap := in.Memories[0].Store(0xfd000b, 1, 0, 0, loaded); trap != wasm.NoTrap {
		t.Fatal(trap)
	}
	for _, tc := range []struct {
		name                   string
		address, steps, fuel   uint64
		stop                   error
		wantSteps              uint64
		wantLoad, wantConstant value.Value
	}{
		{name: "wrapping subtraction", address: 1},
		{name: "load trap", address: 65521, wantSteps: 2},
		{name: "poll on constant", address: 1, steps: 61, stop: context.Canceled, wantSteps: 64, wantLoad: loaded},
		{name: "poll on binary", address: 1, steps: 60, stop: context.Canceled, wantSteps: 64, wantLoad: loaded, wantConstant: literal},
		{name: "fuel on binary", address: 1, fuel: 3, stop: wasm.ErrLimit, wantSteps: 3, wantLoad: loaded, wantConstant: literal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ctx context.Context
			if tc.stop == context.Canceled {
				ctx = &cancelAfter{at: 2}
			}
			e := New(s, ctx, Config{Limits: DefaultLimits(), Metered: tc.fuel != 0, Fuel: tc.fuel})
			defer e.Close()
			if err := e.Start(in.Functions[0], []value.Value{{Lo: tc.address}}); err != nil {
				t.Fatal(err)
			}
			frame := e.values[e.frames[0].base:]
			e.steps = tc.steps
			status := e.Run()
			if tc.wantSteps == 0 {
				want := value.Value{Lo: 0xffffffffffffffff, Hi: 0x00000002ffffffff}
				if status != Done || e.Results()[0] != want {
					t.Fatal(status, e.Err(), e.Results(), want)
				}
				return
			}
			if status != Failed || e.steps != tc.wantSteps || frame[load.Dst] != tc.wantLoad ||
				frame[constant.Dst] != tc.wantConstant {
				t.Fatal(status, e.Err(), e.steps, frame[load.Dst], frame[constant.Dst])
			}
			if tc.stop != nil {
				if !errors.Is(e.Err(), tc.stop) {
					t.Fatal(e.Err(), tc.stop)
				}
			} else {
				var trap *wasm.Trap
				if !errors.As(e.Err(), &trap) || trap.Code != wasm.TrapMemoryBounds || trap.Offset != 701 {
					t.Fatal(e.Err())
				}
			}
		})
	}
}
