package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func external(name string, sort model.Sort, index uint32) *model.Import {
	return &model.Import{Name: model.Name{Text: name}, Type: model.ExternType{Sort: sort, Index: index}}
}

func TestInterfaceDeclarationsAndResourceCapture(t *testing.T) {
	s := scope(t)
	for _, kind := range []model.Kind{model.Component, model.Instance} {
		d := &model.TypeDef{Kind: kind, Declarations: []model.TypeDeclaration{
			{Kind: 4, Export: &model.Import{Name: model.Name{Text: "file"}, Type: model.ExternType{Sort: model.TypeSort, Bound: 1}}},
			{Kind: 1, Type: &model.TypeDef{Kind: model.Borrow, Resource: 0}},
			{Kind: 1, Type: &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "self", Type: *idx(1)}}}},
			{Kind: 4, Export: external("[method]file.read", model.FunctionSort, 2)},
		}}
		v := define(t, s, d)
		if len(v.Exports) != 2 || v.Exports[1].Type.Parameters[0].Type.ResourceID != v.Exports[0].Type.Resource || v.resource {
			t.Fatal(v)
		}
		nested, err := s.pool.Scope(s, true)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := nested.Outer(1, uint32(len(s.types)-1)); err != nil {
			t.Fatal("bound resources treated as lexical captures", err)
		}
	}
	r, _ := s.Abstract()
	rIndex := uint32(len(s.types) - 1)
	d := &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{
		{Kind: 2, Alias: &model.Alias{Kind: 2, Sort: model.TypeSort, Count: 1, Index: rIndex}},
		{Kind: 4, Export: external("file", model.TypeSort, 0)},
	}}
	v := define(t, s, d)
	if !v.resource || v.minimumResource != r.Resource {
		t.Fatal("lost free resource capture", v)
	}
	nested, _ := s.pool.Scope(s, true)
	if _, err := nested.Outer(1, uint32(len(s.types)-1)); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	define(t, s, &model.TypeDef{Kind: model.Component, Declarations: []model.TypeDeclaration{
		{Kind: 1, Type: &model.TypeDef{Kind: model.Function}},
		{Kind: 3, Import: external("run", model.FunctionSort, 0)},
		{Kind: 4, Export: external("run", model.FunctionSort, 0)},
	}})
}

func TestExternSortsAndTypeNamespaceIsolation(t *testing.T) {
	s := scope(t)
	for _, kind := range []model.Kind{model.Function, model.AsyncFunction, model.Component, model.Instance, model.U8} {
		v := define(t, s, &model.TypeDef{Kind: kind})
		index := uint32(len(s.types) - 1)
		for _, sort := range []model.Sort{model.FunctionSort, model.ComponentSort, model.InstanceSort, model.TypeSort} {
			got, err := s.Reference(model.ExternType{Sort: sort, Index: index})
			want := sort == model.TypeSort || sort == model.FunctionSort && (kind == model.Function || kind == model.AsyncFunction) || sort == model.ComponentSort && kind == model.Component || sort == model.InstanceSort && kind == model.Instance
			if want && (err != nil || got != v) || !want && !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(kind, sort, got, err)
			}
		}
	}
	for _, ref := range []model.ExternType{{Sort: model.FunctionSort, Index: 100}, {Sort: model.TypeSort, Bound: 1}, {Sort: model.TypeSort, Bound: 2}} {
		if _, err := s.Reference(ref); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, sort := range []model.Sort{model.ValueSort} {
		if _, err := s.Reference(model.ExternType{Sort: sort}); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	for _, d := range []*model.TypeDef{
		{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 3, Import: external("run", model.FunctionSort, 0)}}},
		{Kind: model.Component, Declarations: []model.TypeDeclaration{{Kind: 4, Export: external("run", model.FunctionSort, 0)}}},
		{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 1, Type: &model.TypeDef{Kind: model.Resource, Representation: wasm.I32}}}},
		{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 99}}},
	} {
		if _, err := s.Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(d, err)
		}
	}
	for _, d := range []*model.TypeDef{
		{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 0}}},
		{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 2, Alias: &model.Alias{Kind: 2, Sort: model.CoreTypeSort}}}},
	} {
		if _, err := s.Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := s.Declare(nil, false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Declare(external("bad", model.FunctionSort, 100), false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Declare(external("bad_label", model.FunctionSort, 0), false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Declare(external("alias", model.TypeSort, 0), true); err != nil {
		t.Fatal(err)
	}
}

func TestInstanceExportTypeAliases(t *testing.T) {
	s := scope(t)
	v := define(t, s, &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{
		{Kind: 1, Type: &model.TypeDef{Kind: model.Instance}},
		{Kind: 4, Export: external("inner", model.InstanceSort, 0)},
		{Kind: 4, Export: &model.Import{Name: model.Name{Text: "file"}, Type: model.ExternType{Sort: model.TypeSort, Bound: 1}}},
	}})
	if _, err := s.Declare(external("filesystem", model.InstanceSort, 0), false); err != nil {
		t.Fatal(err)
	}
	if err := s.aliasDeclaration(&model.Alias{Kind: 0, Sort: model.TypeSort, Instance: 0, Name: "file"}); err != nil {
		t.Fatal(err)
	}
	if s.types[1] != s.items[model.InstanceSort][0].Exports[1].Type || s.types[1] == v.Exports[1].Type {
		t.Fatal("alias did not retain resource identity")
	}
	if err := s.aliasDeclaration(&model.Alias{Kind: 0, Sort: model.InstanceSort, Instance: 0, Name: "inner"}); err != nil {
		t.Fatal(err)
	}
	if len(s.items[model.InstanceSort]) != 2 {
		t.Fatal(s.items)
	}
	for _, a := range []*model.Alias{
		nil, {Kind: 0, Sort: model.FunctionSort}, {Kind: 1, Sort: model.TypeSort},
		{Kind: 2, Sort: model.ComponentSort}, {Kind: 0, Sort: model.TypeSort, Instance: 100},
		{Kind: 0, Sort: model.TypeSort, Instance: 0, Name: "absent"},
		{Kind: 0, Sort: model.InstanceSort, Instance: 0, Name: "file"},
	} {
		if err := s.aliasDeclaration(a); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(a, err)
		}
	}
}

func TestInterfaceMetadataAndDepthLimits(t *testing.T) {
	for _, budget := range []uint32{2, 3} {
		s, _ := NewTypePool(Limits{MaxNodes: budget, MaxDepth: 4}).Scope(nil, true)
		if _, err := s.Define(&model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 1, Type: &model.TypeDef{Kind: model.U8}}}}, nil); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(budget, err)
		}
	}
	p := NewTypePool(Limits{MaxNodes: 100, MaxDepth: 1})
	s, _ := p.Scope(nil, true)
	if _, err := s.Define(&model.TypeDef{Kind: model.Instance}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := Equal(define(t, scope(t), &model.TypeDef{Kind: model.Instance}), define(t, scope(t), &model.TypeDef{Kind: model.Instance}), 100); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
}
