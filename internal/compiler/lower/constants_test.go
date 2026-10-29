package lower

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestInitializerFoldingPreservesDefinitionAndImports(t *testing.T) {
	expr := module.Expr{{Op: 0x41, Bits: 20}, {Op: 0x41, Bits: 22}, {Op: 0x6a}, {Op: 0xb}}
	dependent := module.Expr{{Op: 0x41, Bits: 20}, {Op: 0x23}, {Op: 0x6a}, {Op: 0xb}}
	d := &module.Module{
		Imports: []module.Import{{Kind: 3, Global: module.GlobalType{Value: wasm.I32}}},
		Globals: []module.Global{
			{Type: module.GlobalType{Value: wasm.I32}, Init: expr},
			{Type: module.GlobalType{Value: wasm.I32}, Init: dependent},
		},
		Tables:   []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}},
		Memories: []module.Limits{{Min: 1}},
		Elements: []module.Element{{Mode: module.Active, Offset: expr, Type: wasm.FuncRef, Expressions: []module.Expr{{{Op: 0xd0, A: uint32(wasm.FuncRef)}, {Op: 0xb}}}}},
		Data:     []module.Data{{Mode: module.Active, Offset: expr}},
	}
	v, err := validate.CheckWithFeatures(d, validate.DefaultLimits(), wasm.ExtendedConst)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Compile(v, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := m.FoldedConstant(expr); !ok || got != 42 {
		t.Fatal(got, ok)
	}
	for _, e := range []module.Expr{dependent, {{Op: 0x41}, {Op: 0xb}}, nil} {
		if got, ok := m.FoldedConstant(e); ok || got != 0 {
			t.Fatal("folded instance-dependent or simple expression", e, got, ok)
		}
	}
	if !reflect.DeepEqual(expr, module.Expr{{Op: 0x41, Bits: 20}, {Op: 0x41, Bits: 22}, {Op: 0x6a}, {Op: 0xb}}) || m.Validated.Definition() != d {
		t.Fatal("mutated validated input")
	}
	if _, err := Compile(v, Limits{MaxNodes: 3, MaxVerifyWork: 100}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("initializer work bypassed lowering limit", err)
	}
}

func TestInitializerFoldingLargeAndWrapped(t *testing.T) {
	var expr module.Expr
	for range 64 {
		expr = append(expr, module.Instr{Op: 0x42, Bits: ^uint64(0)})
	}
	for range 63 {
		expr = append(expr, module.Instr{Op: 0x7c})
	}
	expr = append(expr, module.Instr{Op: 0xb})
	d := &module.Module{Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I64}, Init: expr}}}
	v, err := validate.CheckWithFeatures(d, validate.DefaultLimits(), wasm.ExtendedConst)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Compile(v, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := m.FoldedConstant(expr); !ok || got != ^uint64(63) {
		t.Fatal(got, ok)
	}
}
