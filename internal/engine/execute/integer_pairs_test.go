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

func TestIntegerPairsPreserveValuesFuelAndReuse(t *testing.T) {
	for _, test := range []struct{ op, input, constant, want uint32 }{
		{0x6a, 0xffffffff, 7, 6}, {0x6b, 0, 1, 0xffffffff},
		{0x6c, 0x80000001, 3, 0x80000003}, {0x71, 0xdeadbeef, 0x7fffffff, 0x5eadbeef},
		{0x73, 1, 0xffffffff, 0xfffffffe}, {0x74, 0x12345678, 36, 0x23456780},
		{0x77, 0x12345678, 4, 0x23456781}, {0x48, 0xffffffff, 0, 1}, {0x4f, 0xffffffff, 0, 1},
	} {
		s := state.New(state.DefaultLimits())
		m := simple(module.Expr{{Op: 0x20}, {Op: 0x41, Bits: uint64(int64(int32(test.constant)))}, {Op: test.op}, {Op: 0xb}}, wasm.I32)
		m.Types[0].Params = []wasm.ValueType{wasm.I32}
		in := instance(t, s, m, nil)
		fn, _ := s.Function(in.Functions[0])
		hinted := false
		for _, op := range fn.Code.Code {
			hinted = hinted || op.Op == 0x41 && op.Count != 0
		}
		if !hinted {
			t.Fatal("fixture does not exercise an integer pair", test.op)
		}
		for _, metered := range []bool{false, true} {
			e := New(s, nil, Config{Limits: DefaultLimits(), Metered: metered, Fuel: 100})
			for range 2 {
				if err := e.Start(in.Functions[0], []value.Value{{Lo: uint64(test.input)}}); err != nil || e.Run() != Done || uint32(e.Results()[0].Lo) != test.want || e.Results()[0].Hi != 0 {
					t.Fatal(test, metered, err, e.Err(), e.Results())
				}
				if metered && e.Fuel() != 97 {
					t.Fatal("changed source fuel", e.Fuel())
				}
			}
			e.Close()
		}
		s.Close()
	}
}

func TestIntegerPairsDoNotCrossPollingOrFuelBoundaries(t *testing.T) {
	body := make(module.Expr, 61)
	for i := range body {
		body[i].Op = 1
	}
	body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x41, Bits: 7}, module.Instr{Op: 0x6a}, module.Instr{Op: 0x24}, module.Instr{Op: 0xb})
	m := simple(body)
	m.Types[0].Params = []wasm.ValueType{wasm.I32}
	m.Globals = []module.Global{{Type: module.GlobalType{Value: wasm.I32, Mutable: true}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, m, nil)
	e := New(s, &cancelAfter{at: 2}, Config{Limits: DefaultLimits()})
	defer e.Close()
	if err := e.Start(in.Functions[0], []value.Value{{}}); err != nil || e.Run() != Failed || e.Err() != context.Canceled || e.steps != 64 || in.Globals[0].Value.Lo != 0 {
		t.Fatal("pair bypassed cancellation", err, e.Err(), e.steps, in.Globals[0].Value)
	}
	for _, fuel := range []uint64{63, 65} {
		e.Quiesce()
		if err := e.Rebind(context.Background(), DefaultLimits(), true, fuel); err != nil {
			t.Fatal(err)
		}
		if err := e.Start(in.Functions[0], []value.Value{{}}); err != nil {
			t.Fatal(err)
		}
		status := e.Run()
		if e.Fuel() != 0 || fuel == 63 && (status != Failed || !errors.Is(e.Err(), wasm.ErrLimit) || e.steps != 63 || in.Globals[0].Value.Lo != 0) || fuel == 65 && (status != Done || in.Globals[0].Value.Lo != 7) {
			t.Fatal("pair changed precise fuel", fuel, status, e.Err(), e.steps, in.Globals[0].Value)
		}
	}
}
