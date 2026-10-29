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
