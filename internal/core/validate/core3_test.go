package validate

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestCore3ConstantGlobalVisibilityAndLimits(t *testing.T) {
	literal := module.Expr{{Op: 0x41, Bits: 1}, {Op: 0xb}}
	reference := func(index uint32) module.Expr { return module.Expr{{Op: 0x23, A: index}, {Op: 0xb}} }
	makeModule := func() *module.Module {
		return &module.Module{Globals: []module.Global{
			{Type: module.GlobalType{Value: wasm.I32}, Init: literal},
			{Type: module.GlobalType{Value: wasm.I32}, Init: reference(0)},
		}}
	}
	if _, err := CheckWithFeatures(makeModule(), DefaultLimits(), wasm.Core3Features); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*module.Module){
		func(m *module.Module) { m.Globals[0].Init = reference(0) },
		func(m *module.Module) { m.Globals[0].Init = reference(1) },
		func(m *module.Module) { m.Globals[1].Init = reference(2) },
		func(m *module.Module) { m.Globals[0].Type.Mutable = true },
		func(m *module.Module) {
			m.Globals[0].Type.Value = wasm.I64
			m.Globals[0].Init = module.Expr{{Op: 0x42}, {Op: 0xb}}
		},
	} {
		m := makeModule()
		change(m)
		if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.Core3Features); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	m := makeModule()
	limits := DefaultLimits()
	limits.MaxStack = 0
	if _, err := CheckWithFeatures(m, limits, wasm.Core3Features); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	// Table definitions are checked before global definitions, even when a
	// same-type immutable global appears earlier in the binary section order.
	m = &module.Module{Tables: []module.TableType{{Element: wasm.FuncRef}}, TableInit: []module.Expr{reference(0)}, Globals: []module.Global{{Type: module.GlobalType{Value: wasm.FuncRef}, Init: module.Expr{{Op: 0xd0, A: uint32(wasm.FuncRef)}, {Op: 0xb}}}}}
	if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.Core3Features); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("table saw a defined global", err)
	}
}
