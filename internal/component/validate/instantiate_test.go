package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func resourceInterface() *model.TypeDef {
	return &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{
		{Kind: 4, Export: &model.Import{Name: model.Name{Text: "file"}, Type: model.ExternType{Sort: model.TypeSort, Bound: 1}}},
		{Kind: 4, Export: external("same-file", model.TypeSort, 0)},
		{Kind: 1, Type: &model.TypeDef{Kind: model.Borrow, Resource: 0}},
		{Kind: 1, Type: &model.TypeDef{Kind: model.Own, Resource: 0}},
		{Kind: 1, Type: &model.TypeDef{Kind: model.Record, Fields: []model.NamedType{{Name: "input", Type: *idx(2)}, {Name: "output", Type: *idx(3)}}}},
		{Kind: 4, Export: external("handles", model.TypeSort, 4)},
		{Kind: 1, Type: &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "handles", Type: *idx(5)}}, Result: idx(3)}},
		{Kind: 4, Export: external("use", model.FunctionSort, 6)},
	}}
}

func TestInstanceImportsFreshenResourcesAndPreserveEquality(t *testing.T) {
	s := scope(t)
	template := define(t, s, resourceInterface())
	first, err := s.Declare(external("first", model.InstanceSort, 0), false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Declare(external("second", model.InstanceSort, 0), false)
	if err != nil {
		t.Fatal(err)
	}
	if first.Exports[0].Type.Resource == second.Exports[0].Type.Resource || first.Exports[0].Type.Resource == template.Exports[0].Type.Resource {
		t.Fatal("distinct imports shared an abstract resource")
	}
	for _, v := range []*Type{first, second} {
		r := v.Exports[0].Type.Resource
		if v.Exports[0].Type.Resource != v.Exports[1].Type.Resource {
			t.Fatal("lost equality alias")
		}
		f := v.Exports[3].Type
		if f.Result.ResourceID != r || f.Parameters[0].Type.Fields[0].Type.ResourceID != r || f.Parameters[0].Type.Fields[1].Type.ResourceID != r {
			t.Fatal("incomplete dependent substitution", f)
		}
		if f.Result != f.Parameters[0].Type.Fields[1].Type {
			t.Fatal("lost value sharing")
		}
		if len(v.Captures) != 0 || v.resource {
			t.Fatal("local binders became captures", v)
		}
		if _, err := s.pool.layouts.Build(f.Parameters[0].Type); err != nil {
			t.Fatal(err)
		}
	}
	if equal, err := Equal(first.Exports[3].Type, second.Exports[3].Type, 100); err != nil || equal {
		t.Fatal(equal, err)
	}
}

func TestNestedResourcesAndCapturedComponentTypes(t *testing.T) {
	s := scope(t)
	template := define(t, s, &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{
		{Kind: 4, Export: &model.Import{Name: model.Name{Text: "file"}, Type: model.ExternType{Sort: model.TypeSort, Bound: 1}}},
		{Kind: 1, Type: &model.TypeDef{Kind: model.Component, Declarations: []model.TypeDeclaration{
			{Kind: 2, Alias: &model.Alias{Kind: 2, Sort: model.TypeSort, Count: 1, Index: 0}},
			{Kind: 4, Export: external("file", model.TypeSort, 0)},
			{Kind: 1, Type: &model.TypeDef{Kind: model.Own, Resource: 1}},
			{Kind: 1, Type: &model.TypeDef{Kind: model.Function, Result: idx(2)}},
			{Kind: 4, Export: external("make", model.FunctionSort, 3)},
		}}},
		{Kind: 4, Export: external("maker", model.ComponentSort, 1)},
		{Kind: 1, Type: resourceInterface()},
		{Kind: 4, Export: external("inner", model.InstanceSort, 2)},
	}})
	for i := 0; i < 2; i++ {
		bound, err := s.pool.freshInstance(template)
		if err != nil {
			t.Fatal(err)
		}
		id := bound.Exports[0].Type.Resource
		maker := bound.Exports[1].Type
		if len(maker.Captures) != 1 || maker.Captures[0] != id || maker.Exports[1].Type.Result.ResourceID != id {
			t.Fatal("lost component capture", maker)
		}
		if bound.Exports[2].Type.Exports[0].Type.Resource == template.Exports[2].Type.Exports[0].Type.Resource {
			t.Fatal("nested binder was not freshened")
		}
		if len(bound.Captures) != 0 {
			t.Fatal("closed template gained captures")
		}
	}
}

func TestResourceSubstitutionBudgets(t *testing.T) {
	failed, passed := 0, 0
	for available := uint32(0); available < 100; available++ {
		s := scope(t)
		template := define(t, s, resourceInterface())
		s.pool.limits.MaxNodes = s.pool.nodes + available
		result, err := s.pool.freshInstance(template)
		if err != nil {
			if !errors.Is(err, wasm.ErrLimit) || result != nil {
				t.Fatal(available, result, err)
			}
			failed++
		} else {
			passed++
		}
	}
	if failed == 0 || passed == 0 {
		t.Fatal(failed, passed)
	}
}

func TestCaptureTraversalChargesRepeatedEdges(t *testing.T) {
	s, err := NewTypePool(Limits{MaxNodes: 2048, MaxDepth: 128}).Scope(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	define(t, s, &model.TypeDef{Kind: model.Resource, Representation: wasm.I32})
	define(t, s, &model.TypeDef{Kind: model.Own, Resource: 0})
	fields := make([]model.TypeRef, 512)
	for i := range fields {
		fields[i] = *idx(1)
	}
	define(t, s, &model.TypeDef{Kind: model.Tuple, Tuple: fields})
	function := define(t, s, &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "input", Type: *idx(2)}}})
	completed := 0
	for i := 0; i < 10; i++ {
		_, err = s.pool.captures(function, ^uint64(0))
		if err != nil {
			break
		}
		completed++
	}
	if completed == 0 || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("repeated shared edges escaped traversal quota", completed, err)
	}
}
