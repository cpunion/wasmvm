package wasm_test

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestValueTypes(t *testing.T) {
	names := map[wasm.ValueType]string{wasm.ExnRef: "exnref", wasm.NoExnRef: "nullexnref", wasm.I32: "i32", wasm.I64: "i64", wasm.F32: "f32", wasm.F64: "f64", wasm.V128: "v128", wasm.FuncRef: "funcref", wasm.ExternRef: "externref", wasm.AnyRef: "anyref", wasm.EqRef: "eqref", wasm.I31Ref: "i31ref", wasm.StructRef: "structref", wasm.ArrayRef: "arrayref", wasm.NoneRef: "nullref", wasm.NoFuncRef: "nullfuncref", wasm.NoExternRef: "nullexternref"}
	for i := 0; i < 256; i++ {
		typ := wasm.ValueType(i)
		name, valid := names[typ]
		if typ.Valid() != valid {
			t.Fatalf("valid(%d)", i)
		}
		if valid && typ.String() != name {
			t.Fatalf("name(%d) = %q", i, typ)
		}
		if typ.IsReference() != (typ >= wasm.ExnRef && typ <= wasm.NoExnRef) {
			t.Fatalf("reference(%d)", i)
		}
	}
	if wasm.ValueType(0).String() != "type(0x0)" {
		t.Fatal("unknown type formatting")
	}
}

func TestErrorCategories(t *testing.T) {
	for _, category := range []error{wasm.ErrMalformed, wasm.ErrInvalid, wasm.ErrLimit, wasm.ErrUnsupported} {
		err := &wasm.Error{Kind: category, Offset: 17, Detail: "test"}
		if !errors.Is(err, category) || err.Error() != "wasm byte 0x11: "+category.Error()+": test" {
			t.Fatal(err)
		}
	}
}
