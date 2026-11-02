package execute

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func prefixReads(n int) module.Expr {
	body := make(module.Expr, n)
	for i := range body {
		body[i] = module.Instr{Op: 0x20, Offset: uint64(100 + i)}
	}
	for i := 1; i < n; i++ {
		body = append(body, module.Instr{Op: 0x1a})
	}
	return body
}

func TestPrefixChainsPreserveWholeValuesAndMetering(t *testing.T) {
	const reads = 130
	for _, test := range []struct {
		typ wasm.ValueType
		v   value.Value
	}{
		{wasm.I32, value.Value{Lo: 0x80000000}},
		{wasm.I64, value.Value{Lo: 0xfedcba9876543210}},
		{wasm.F32, value.Value{Lo: 0xff800123}},
		{wasm.F64, value.Value{Lo: 0x7ff80000abcdef01}},
		{wasm.V128, value.Value{Lo: 0x123456789abcdef0, Hi: 0xfedcba9876543210}},
	} {
		s := state.New(state.DefaultLimits())
		body := append(prefixReads(reads), module.Instr{Op: 0xb})
		m := simple(body, test.typ)
		m.Types[0].Params = []wasm.ValueType{test.typ}
		in := instance(t, s, m, nil)
		fn, _ := s.Function(in.Functions[0])
		original := append([]slots.Instruction(nil), fn.Code.Code...)
		var hinted int
		for _, op := range original {
			if op.Op == 0x20 && op.Count == 1 {
				hinted++
			}
		}
		if hinted != reads {
			t.Fatal("fixture did not coalesce its local reads", test.typ, hinted)
		}
		for _, fuel := range []uint64{0, 2*reads - 2, 2*reads - 1} {
			for _, cached := range []bool{false, true} {
				copy(fn.Code.Code, original)
				if !cached {
					for i := range fn.Code.Code {
						if fn.Code.Code[i].Op == 0x20 {
							fn.Code.Code[i].Count = 0
						}
					}
				}
				e := New(s, nil, Config{Limits: DefaultLimits(), Metered: fuel != 0, Fuel: fuel})
				for range 2 {
					if err := e.Start(in.Functions[0], []value.Value{test.v}); err != nil {
						t.Fatal(err)
					}
					status := e.Run()
					if fuel == 2*reads-2 {
						if status != Failed || !errors.Is(e.Err(), wasm.ErrLimit) || e.Fuel() != 0 || e.steps != fuel {
							t.Fatal("prefix changed precise fuel", test.typ, cached, status, e.Err(), e.Fuel(), e.steps)
						}
					} else if status != Done || len(e.Results()) != 1 || e.Results()[0] != test.v || e.Fuel() != 0 {
						t.Fatal("prefix changed value bits or reuse", test.typ, cached, status, e.Err(), e.Results(), e.Fuel())
					}
				}
				e.Close()
			}
		}
		s.Close()
	}
}

func TestPrefixChainsStopAtPollingBoundariesAndBeforeTrap(t *testing.T) {
	body := prefixReads(70)
	body = append(body, module.Instr{Op: 0x24}, module.Instr{Op: 0x41, Bits: 65536}, module.Instr{Op: 0x28, Offset: 777}, module.Instr{Op: 0x1a}, module.Instr{Op: 0xb})
	m := simple(body)
	m.Types[0].Params = []wasm.ValueType{wasm.I32}
	m.Globals = []module.Global{{Type: module.GlobalType{Value: wasm.I32, Mutable: true}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}
	m.Memories = []module.Limits{{Min: 1}}
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, m, nil)
	for _, initial := range []uint64{0, 62, math.MaxUint64 - 1} {
		e := New(s, &cancelAfter{at: 2}, Config{Limits: DefaultLimits()})
		if err := e.Start(in.Functions[0], []value.Value{{Lo: 7}}); err != nil {
			t.Fatal(err)
		}
		e.steps = initial
		want := uint64(64)
		if initial == math.MaxUint64-1 {
			want = 0
		}
		if e.Run() != Failed || e.Err() != context.Canceled || e.steps != want || in.Globals[0].Value.Lo != 0 {
			t.Fatal("prefix crossed a polling boundary", initial, e.Err(), e.steps, in.Globals[0].Value)
		}
		e.Close()
	}
	e := New(s, nil, Config{Limits: DefaultLimits()})
	defer e.Close()
	if err := e.Start(in.Functions[0], []value.Value{{Lo: 7}}); err != nil {
		t.Fatal(err)
	}
	var trap *wasm.Trap
	if e.Run() != Failed || !errors.As(e.Err(), &trap) || trap.Code != wasm.TrapMemoryBounds || trap.Offset != 777 || in.Globals[0].Value.Lo != 7 {
		t.Fatal("prefix changed a following trap or visible effects", e.Err(), in.Globals[0].Value)
	}
}
