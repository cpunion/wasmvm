package execute

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPortableCounterPublishesAtHostYieldAndCompletion(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	host, _ := s.NewHost(module.FuncType{}, "pause")
	body := make(module.Expr, 193)
	for i := range body {
		body[i].Op = 1
	}
	body = append(body, module.Instr{Op: 0x10}, module.Instr{Op: 1}, module.Instr{Op: 0xb})
	m := simple(body)
	m.Types = append(m.Types, module.FuncType{})
	m.Imports = []module.Import{{Kind: 0, Type: 1}}
	in := instance(t, s, m, func(string, string) (state.Extern, bool) { return host, true })
	e := New(s, nil, Config{Limits: DefaultLimits(), Metered: true, Fuel: 512})
	defer e.Close()
	if err := e.Start(in.Functions[1], nil); err != nil {
		t.Fatal(err)
	}
	if e.Run() != HostRequest || e.steps <= 128 || e.steps != 512-e.Fuel() {
		t.Fatal("host yield lost charged steps", e.steps, e.Fuel(), e.Err())
	}
	before := e.steps
	if err := e.ResumeHost(nil, nil); err != nil {
		t.Fatal(err)
	}
	if e.Run() != Done || e.steps <= before || e.steps != 512-e.Fuel() {
		t.Fatal("completion lost resumed steps", before, e.steps, e.Fuel(), e.Err())
	}
}

func TestPortableCounterPublishesAtTrap(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	body := make(module.Expr, 193)
	for i := range body {
		body[i].Op = 1
	}
	body = append(body, module.Instr{Op: 0}, module.Instr{Op: 0xb})
	in := instance(t, s, simple(body), nil)
	e := New(s, nil, Config{Limits: DefaultLimits(), Metered: true, Fuel: 512})
	defer e.Close()
	if err := e.Start(in.Functions[0], nil); err != nil {
		t.Fatal(err)
	}
	var trap *wasm.Trap
	if e.Run() != Failed || !errors.As(e.Err(), &trap) || trap.Code != wasm.TrapUnreachable || e.steps <= 128 || e.steps != 512-e.Fuel() {
		t.Fatal("trap lost charged steps", e.steps, e.Fuel(), e.Err())
	}
}
