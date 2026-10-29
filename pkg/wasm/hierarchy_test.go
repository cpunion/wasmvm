package wasm_test

import (
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestDeepNominalHierarchyAndRollback(t *testing.T) {
	var h wasm.TypeHierarchy
	for i := uint32(0); i < 65536; i++ {
		h.Add(i)
	}
	for _, ancestor := range []uint32{1, 2, 123, 32769, 65535, 65536} {
		if !h.IsSubtype(65536, ancestor) {
			t.Fatal("missing ancestor", ancestor)
		}
		if ancestor != 65536 && h.IsSubtype(ancestor, 65536) {
			t.Fatal("reversed subtype")
		}
	}
	h.Add(0)
	if h.IsSubtype(65537, 1) || h.IsSubtype(65536, 65537) || h.IsSubtype(65536, 0) || h.IsSubtype(999999, 1) {
		t.Fatal("unrelated identities")
	}
	labels := []uint32{65536, 32768, 65537}
	if !h.ReferenceSubtype(wasm.FunctionReference(0, false), wasm.FunctionReference(1, true), labels) || h.ReferenceSubtype(wasm.FunctionReference(0, false), wasm.FunctionReference(2, true), labels) {
		t.Fatal("nominal reference types")
	}
	h.Rollback(2)
	h.Add(0)
	if h.IsSubtype(3, 2) || !h.IsSubtype(2, 1) {
		t.Fatal("rolled-back ancestor state")
	}
}
