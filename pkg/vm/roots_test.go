package vm

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSuspendedLocalsKeepExternRefsAlive(t *testing.T) {
	e, s, l := setup(t)
	input, _ := s.NewExternRef("only guest local")
	collector, _ := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) { return nil, c.Collect() })
	host, _ := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
		input.Release()
		_, err := c.Call(collector, nil)
		return nil, err
	})
	l.Define("h", "collect", host)
	d := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}}, Imports: []module.Import{{Module: "h", Name: "collect", Kind: 0}}, Functions: []module.Function{{Type: 1, Body: module.Expr{{Op: 0x10}, {Op: 0x20}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Kind: 0, Index: 1}}}
	fn := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
	r, err := fn.Call(nil, []Value{input}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.ExternRef(r.Values[0]); err != nil || got != "only guest local" {
		t.Fatal(got, err)
	}
	r.Values[0].Release()
	s.Collect()
	u, _ := s.Usage()
	if u.ExternRefs != 0 || u.HostRoots != 0 {
		t.Fatal(u)
	}
}

func TestFailedStartFunctionEscapesThroughHostRoot(t *testing.T) {
	e, s, l := setup(t)
	var escaped Value
	host, _ := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.FuncRef}}, func(c *Caller, args []Value) ([]Value, error) {
		var err error
		escaped, err = s.Root(args[0])
		return nil, err
	})
	l.Define("h", "save", host)
	d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.FuncRef}}, {Results: []wasm.ValueType{wasm.I32}}, {}}, Imports: []module.Import{{Module: "h", Name: "save", Kind: 0}}, Functions: []module.Function{{Type: 1, Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}}, {Type: 2, Body: module.Expr{{Op: 0xd2, A: 1}, {Op: 0x10}, {Op: 0}, {Op: 0xb}}}}, Elements: []module.Element{{Type: wasm.FuncRef, Mode: module.Declarative, Functions: []uint32{1}}}, HasStart: true, Start: 2}
	compiled := compileDefinition(t, e, d)
	if _, err := l.Instantiate(nil, compiled, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	s.Collect()
	r, err := s.CallRef(nil, escaped, nil, CallOptions{})
	if err != nil || bits(t, r.Values[0]) != 42 {
		t.Fatal(r, err)
	}
	escaped.Release()
	s.Collect()
	u, _ := s.Usage()
	if u.Core.Instances != 0 || u.Core.Functions != 1 || u.HostRoots != 0 {
		t.Fatal(u)
	}
}
