package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCoreModuleTypesAndOuterAliases(t *testing.T) {
	s := scope(t)
	source := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, {}}, TypeDefs: []module.TypeDef{{Kind: module.FunctionKind}, {}}}
	if err := s.DefineCore(&model.CoreTypeDef{Start: 0, Count: 1}, source); err != nil {
		t.Fatal(err)
	}
	body := &module.Module{Types: []module.FuncType{{}, {Results: []wasm.ValueType{wasm.I64}}}, TypeDefs: []module.TypeDef{{}, {Kind: module.FunctionKind}}}
	d := &model.CoreTypeDef{Module: true, Start: 1, Count: 1, Types: body, Declarations: []model.CoreDeclaration{
		{Kind: 2, Alias: &model.Alias{Kind: 2, Sort: model.CoreTypeSort, Count: 1, Index: 0}},
		{Kind: 1, Start: 1, Count: 1},
		{Kind: 0, Module: "a", Name: "b", Type: module.Import{Kind: 0, Type: 0}},
		{Kind: 0, Module: "a", Name: "mem", Type: module.Import{Kind: 2}},
		{Kind: 3, Name: "f", Type: module.Import{Kind: 0, Type: 1}},
		{Kind: 3, Name: "g", Type: module.Import{Kind: 3, Global: module.GlobalType{Value: wasm.I32}}},
	}}
	if err := s.DefineCore(d, source); err != nil {
		t.Fatal(err)
	}
	m, err := s.Reference(model.ExternType{Sort: model.CoreModuleSort, Index: 1})
	if err != nil || len(m.CoreModule.Imports) != 2 || len(m.CoreModule.Exports) != 2 || m.CoreModule.Imports[0].Descriptor.Type != s.core.ids[0] {
		t.Fatal(m, err)
	}
	if _, err := s.Reference(model.ExternType{Sort: model.CoreModuleSort, Index: 0}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := s.aliasDeclaration(&model.Alias{Kind: 2, Sort: model.CoreTypeSort, Count: 0, Index: 1}); err != nil {
		t.Fatal(err)
	}
	if s.core.modules[2] != m.CoreModule {
		t.Fatal("module alias identity")
	}
	nested, _ := s.pool.Scope(s, false)
	if err := nested.core.alias(1, 0, true); err != nil {
		t.Fatal(err)
	}
	if nested.core.ids[0] != s.core.ids[0] {
		t.Fatal("Core type alias identity")
	}
	if err := nested.core.alias(1, 1, false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := nested.core.alias(3, 0, true); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := nested.core.alias(1, 100, true); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := Equal(m, m, 100); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestCoreGCAliasReferenceFamily(t *testing.T) {
	s := scope(t)
	outer := &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.StructKind, Open: true, Fields: []module.FieldType{{Value: wasm.I32}}}}}
	if err := s.DefineCore(&model.CoreTypeDef{Count: 1}, outer); err != nil {
		t.Fatal(err)
	}
	inner, _ := s.pool.Scope(s, false)
	if err := inner.core.alias(1, 0, true); err != nil {
		t.Fatal(err)
	}
	// The decoder cannot assign a family to an unresolved outer alias and
	// initially uses funcref. Semantic resolution repairs only alias targets.
	local := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{wasm.FunctionReference(0, true)}}}, TypeDefs: []module.TypeDef{{}, {Kind: module.FunctionKind}}}
	if err := inner.core.group(local, 1, 1); err != nil {
		t.Fatal(err)
	}
	functionID := inner.core.ids[1]
	f := s.pool.registry.Types[functionID].Function
	if f.Params[0].Storage() != wasm.StructRef.Storage() || !f.Params[0].Canonical() {
		t.Fatal(f)
	}
	if local.Types[1].Params[0].Storage() != wasm.FuncRef.Storage() {
		t.Fatal("mutated decoded graph")
	}
	desc, err := inner.core.external(module.Import{Kind: 3, Global: module.GlobalType{Value: wasm.FunctionReference(0, true)}})
	if err != nil || desc.Global.Value.Storage() != wasm.StructRef.Storage() || !desc.Global.Value.Canonical() {
		t.Fatal(desc, err)
	}
	if err := inner.core.group(&module.Module{Types: make([]module.FuncType, 2)}, 2, 0); err != nil {
		t.Fatal("empty rec group", err)
	}
}

func TestInvalidCoreModuleTypes(t *testing.T) {
	for _, declarations := range [][]model.CoreDeclaration{
		{{Kind: 3, Name: "a", Type: module.Import{Kind: 0}}},
		{{Kind: 3, Name: "a", Type: module.Import{Kind: 2}}, {Kind: 3, Name: "a", Type: module.Import{Kind: 2}}},
		{{Kind: 0, Module: "a", Name: "b", Type: module.Import{Kind: 2}}, {Kind: 0, Module: "a", Name: "b", Type: module.Import{Kind: 2}}},
		{{Kind: 0, Type: module.Import{Kind: 2, Memory: module.Limits{Min: 70000}}}},
		{{Kind: 1, Count: 1}}, {{Kind: 2}}, {{Kind: 2, Alias: &model.Alias{Kind: 0}}}, {{Kind: 9}},
	} {
		s := scope(t)
		err := s.DefineCore(&model.CoreTypeDef{Module: true, Count: 1, Types: new(module.Module), Declarations: declarations}, nil)
		if !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(declarations, err)
		}
	}
	s := scope(t)
	for _, d := range []*model.CoreTypeDef{nil, {Start: 1}, {Module: true, Count: 2}, {Count: 1}} {
		if err := s.DefineCore(d, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(d, err)
		}
	}
	for _, source := range []*module.Module{{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{}, {}}}, {Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}}}}} {
		if err := scope(t).core.group(source, 0, 2); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(source, err)
		}
	}
	s.core.ids = []uint32{0}
	s.core.types = &module.Module{Types: []module.FuncType{{}}}
	if err := s.core.alias(0, 0, true); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestCoreTypesWithinInterfaceDeclarations(t *testing.T) {
	s := scope(t)
	source := &module.Module{Types: []module.FuncType{{}}}
	define(t, s, &model.TypeDef{Kind: model.Instance, CoreTypes: source, Declarations: []model.TypeDeclaration{
		{Kind: 0, CoreType: &model.CoreTypeDef{Module: true, Count: 1}},
		{Kind: 4, Export: external("module", model.CoreModuleSort, 0)},
	}})
	for _, budget := range []uint32{1, 2, 3, 4} {
		s, _ := NewTypePool(Limits{MaxNodes: budget, MaxDepth: 10}).Scope(nil, true)
		if err := s.DefineCore(&model.CoreTypeDef{Module: true, Count: 1, Declarations: []model.CoreDeclaration{{Kind: 3, Type: module.Import{Kind: 2}}}}, nil); err != nil && !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(budget, err)
		}
	}
}
