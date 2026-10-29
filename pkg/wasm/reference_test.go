package wasm_test

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
	"testing"
)

func TestTypedReferenceDomainsAndNullability(t *testing.T) {
	for _, index := range []uint32{0, 1, 1 << 16, math.MaxUint32} {
		for _, canonical := range []bool{false, true} {
			for _, nullable := range []bool{false, true} {
				typ := wasm.FunctionReference(index, nullable)
				if canonical {
					typ = wasm.CanonicalFunctionReference(index, nullable)
				}
				got, ok := typ.TypeIndex()
				if !typ.Valid() || !typ.IsFunctionReference() || typ.IsExternalReference() || !ok || got != index || typ.Canonical() != canonical || typ.Nullable() != nullable || typ.Defaultable() != nullable || typ.Storage() != wasm.FuncRef {
					t.Fatal(typ)
				}
				if typ.NonNull().Nullable() || !typ.WithNull().Nullable() || typ.NonNull().WithNull() != typ.WithNull() {
					t.Fatal(typ)
				}
				if !wasm.ReferenceSubtype(typ.NonNull(), typ.WithNull(), nil) || wasm.ReferenceSubtype(typ.WithNull(), typ.NonNull(), nil) {
					t.Fatal(typ)
				}
				if typ.String() == "" {
					t.Fatal("type name")
				}
			}
		}
	}
	for _, typ := range []wasm.ValueType{wasm.FuncRef, wasm.ExternRef} {
		if !typ.Nullable() || !typ.NonNull().Valid() || typ.NonNull().Defaultable() || typ.NonNull().WithNull() != typ || typ.NonNull().String() == "" {
			t.Fatal(typ)
		}
	}
	for _, typ := range []wasm.ValueType{0, wasm.I32, wasm.V128} {
		if typ.NonNull() != typ || typ.WithNull() != typ || typ.Storage() != typ {
			t.Fatal(typ)
		}
	}
	for _, typ := range []wasm.ValueType{1<<48 | wasm.FuncRef, 1<<9 | wasm.FuncRef, 1<<12 | wasm.FuncRef, 1<<8 | 1<<16 | wasm.FuncRef, 1<<8 | 1<<11 | wasm.ExternRef, 1<<8 | 1<<9 | wasm.ExternRef} {
		if typ.Valid() {
			t.Fatal("malformed packed type", typ)
		}
	}
	labels := []uint32{0, 0, 2}
	a, b, c := wasm.FunctionReference(0, false), wasm.FunctionReference(1, true), wasm.FunctionReference(2, true)
	if !wasm.ReferenceSubtype(a, b, labels) || wasm.ReferenceSubtype(a, c, labels) || wasm.ReferenceSubtype(a, b, nil) || wasm.ReferenceSubtype(a, wasm.ExternRef, nil) || wasm.ReferenceSubtype(wasm.FuncRef, b, labels) || wasm.ReferenceSubtype(a, wasm.CanonicalFunctionReference(0, true), nil) || wasm.ReferenceSubtype(c, b, labels[:2]) {
		t.Fatal("reference type equivalence")
	}
	if !wasm.ReferenceSubtype(a, wasm.FuncRef, nil) || !wasm.ReferenceSubtype(a, a, nil) {
		t.Fatal("abstract supertype")
	}
}
