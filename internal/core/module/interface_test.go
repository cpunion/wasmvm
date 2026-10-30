package module

import (
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestInterfaceSnapshotCanonicalizesEveryDescriptor(t *testing.T) {
	m := &Module{Types: []FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, {Params: []wasm.ValueType{wasm.I32}}}}
	m.Imports = []Import{{Module: "env", Name: "f", Kind: 0, Type: 0}, {Module: "env", Name: "t", Kind: 1, Table: TableType{Element: wasm.FuncRef, Limits: Limits{Min: 1}}}, {Module: "env", Name: "m", Kind: 2, Memory: Limits{Min: 1, Index64: true}}, {Module: "env", Name: "g", Kind: 3, Global: GlobalType{Value: wasm.I32}}, {Module: "env", Name: "e", Kind: 4, Type: 1}}
	m.Functions = []Function{{Type: 0}}
	m.Tables = []TableType{{Element: m.Reference(0, true), Limits: Limits{Min: 2, Max: 3, HasMax: true}}}
	m.Memories = []Limits{{Min: 2, Max: 3, HasMax: true, Index64: true}}
	m.Globals = []Global{{Type: GlobalType{Value: m.Reference(0, true)}}}
	m.Tags = []uint32{1}
	for kind := byte(0); kind < 5; kind++ {
		m.Exports = append(m.Exports, Export{Name: string(rune('a' + kind)), Kind: kind, Index: 0}, Export{Name: string(rune('k' + kind)), Kind: kind, Index: 1})
	}
	snapshot := SnapshotInterface(m)
	if !snapshot.Valid() || snapshot.Nodes() == 0 || snapshot.definition.types.Functions != nil || snapshot.definition.types.Globals != nil || snapshot.definition.types.Data != nil {
		t.Fatal("snapshot retained executable definitions")
	}
	var registry TypeRegistry
	shape, ok := snapshot.Intern(&registry, 10)
	if !ok || len(shape.Imports) != 5 || len(shape.Exports) != 10 {
		t.Fatal(shape, ok)
	}
	for i, ex := range shape.Exports {
		if ex.Kind != byte(i/2) {
			t.Fatal(i, ex)
		}
	}
	if shape.Imports[0].Type == 0 || shape.Exports[1].Type != shape.Imports[0].Type || shape.Exports[8].Type != shape.Exports[9].Type {
		t.Fatal(shape)
	}
	if !shape.Exports[3].Table.Element.Canonical() || !shape.Exports[7].Global.Value.Canonical() {
		t.Fatal("indexed references were not canonicalized", shape)
	}
	shape.Imports[0].Type = 99
	shape.Exports[0].Name = "changed"
	if m.Imports[0].Type != 0 || m.Exports[0].Name == "changed" {
		t.Fatal("descriptor snapshot aliases source")
	}
	if _, ok := snapshot.Intern(&TypeRegistry{}, 0); ok {
		t.Fatal("type quota ignored")
	}
	if _, ok := snapshot.Intern(nil, 10); ok {
		t.Fatal("nil registry accepted")
	}
	zero := SnapshotInterface(nil)
	if zero.Valid() || zero.Nodes() != 0 {
		t.Fatal(zero)
	}
	if _, ok := zero.Intern(&registry, 10); ok {
		t.Fatal("zero interface accepted")
	}
}

func TestRegistryForkSharesImmutableTypesAndIsolatesIndices(t *testing.T) {
	m := &Module{Types: []FuncType{{Params: []wasm.ValueType{wasm.I32}}, {Params: []wasm.ValueType{wasm.I32}}}, TypeDefs: []TypeDef{{Kind: FunctionKind, Open: true}, {Kind: FunctionKind, Supers: []uint32{0}}}}
	var original TypeRegistry
	ids, ok := original.InternModule(m, 10)
	if !ok {
		t.Fatal("intern")
	}
	fork := original.Fork()
	if !fork.FunctionSubtype(ids[1], ids[0]) || fork.Count() != 2 {
		t.Fatal("hierarchy was not preserved")
	}
	other := &Module{Types: []FuncType{{Params: []wasm.ValueType{wasm.I64}}}}
	if _, ok := fork.InternModule(other, 10); !ok || fork.Count() != 3 || original.Count() != 2 || len(original.Keys) != 2 {
		t.Fatal("fork changed original registry")
	}
	for _, r := range []*TypeRegistry{nil, {}} {
		if fork := r.Fork(); fork.Count() != 0 || fork.Keys == nil {
			t.Fatal(fork)
		}
	}
}
