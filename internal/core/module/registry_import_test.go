package module

import (
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestRegistryImportPreservesRecursiveGroupsAndHierarchy(t *testing.T) {
	m := &Module{Types: make([]FuncType, 4), TypeDefs: make([]TypeDef, 4), TypeGroups: []TypeGroup{{Start: 0, Count: 2}, {Start: 2, Count: 1}, {Start: 3, Count: 1}}}
	m.TypeDefs[0] = TypeDef{Kind: StructKind, Open: true, Fields: []FieldType{{Value: wasm.I32}, {Value: wasm.IndexedReference(wasm.StructRef, 1, true)}}}
	m.TypeDefs[1] = TypeDef{Kind: StructKind, Fields: []FieldType{{Value: wasm.IndexedReference(wasm.StructRef, 0, true)}}}
	m.TypeDefs[2] = TypeDef{Kind: StructKind, Supers: []uint32{0}, Fields: append(append([]FieldType{}, m.TypeDefs[0].Fields...), FieldType{Value: wasm.I64})}
	m.TypeDefs[3] = TypeDef{Kind: FunctionKind}
	m.Types[3] = FuncType{Params: []wasm.ValueType{m.Reference(2, false)}, Results: []wasm.ValueType{m.Reference(0, true)}}
	var source, dest TypeRegistry
	src, ok := source.InternModule(m, 20)
	if !ok {
		t.Fatal("source intern")
	}
	if _, ok := dest.InternModule(&Module{Types: []FuncType{{Params: []wasm.ValueType{wasm.F64}}}}, 20); !ok {
		t.Fatal("seed")
	}
	ids, ok := dest.Import(&source, 20)
	if !ok || len(ids) != 5 || dest.Count() != 5 {
		t.Fatal(ids, ok, dest.Count())
	}
	parent, child := ids[src[0]], ids[src[2]]
	if !dest.Subtype(wasm.IndexedReference(wasm.StructRef, child, false), wasm.IndexedReference(wasm.StructRef, parent, true)) {
		t.Fatal("hierarchy lost")
	}
	fn := dest.Types[ids[src[3]]].Function
	p, _ := fn.Params[0].TypeIndex()
	r, _ := fn.Results[0].TypeIndex()
	edge, _ := dest.Types[parent].Definition.Fields[1].Value.TypeIndex()
	if p != child || r != parent || edge != ids[src[1]] || dest.Types[child].Definition.Supers[0] != parent {
		t.Fatal("reference mapping", p, r, edge, ids)
	}
	again, ok := dest.Import(&source, 20)
	if !ok || dest.Count() != 5 || again[src[3]] != ids[src[3]] {
		t.Fatal("not canonical", again, ok)
	}
	if source.Count() != 4 || source.Types[src[2]].Definition.Supers[0] != src[0] {
		t.Fatal("source mutated")
	}
	var limited TypeRegistry
	if _, ok := limited.InternModule(&Module{Types: []FuncType{{Params: []wasm.ValueType{wasm.F64}}}}, 20); !ok {
		t.Fatal("seed")
	}
	if _, ok := limited.Import(&source, 4); ok || limited.Count() > 4 {
		t.Fatal("quota ignored")
	}
	if _, ok := dest.Import(&source, 3); ok {
		t.Fatal("source quota ignored")
	}
	if _, ok := dest.Import(nil, 20); ok {
		t.Fatal("nil source accepted")
	}
	zero, ok := (&TypeRegistry{}).Import(&TypeRegistry{}, 0)
	if !ok || len(zero) != 1 || zero[0] != 0 {
		t.Fatal(zero, ok)
	}
}
