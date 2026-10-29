package validate

import (
	"errors"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestNondefaultableLocalInitializationIsRequiredInUnreachableCode(t *testing.T) {
	m := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{{Locals: []module.Local{{Count: 1, Type: wasm.ExternRef.NonNull()}}, Body: module.Expr{{Op: 0}, {Op: 0x20}, {Op: 0x1a}, {Op: 0xb}}}}}
	if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.FunctionReferences); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("unreachable local.get bypassed initialization", err)
	}
	m.Functions[0].Body = module.Expr{{Op: 0}, {Op: 0x21}, {Op: 0x20}, {Op: 0x1a}, {Op: 0xb}}
	if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.FunctionReferences); err != nil {
		t.Fatal("unreachable local.set may initialize", err)
	}
}

func TestUnreachableGCConversionRefinesNullability(t *testing.T) {
	for _, tc := range []struct {
		op  uint32
		typ wasm.ValueType
	}{{0xfb001a, wasm.AnyRef.NonNull()}, {0xfb001b, wasm.ExternRef.NonNull()}} {
		m := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{tc.typ}}}, Functions: []module.Function{{Body: module.Expr{{Op: 0}, {Op: tc.op}, {Op: 0xb}}}}}
		if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.GC); err != nil {
			t.Fatal(err)
		}
	}
}
