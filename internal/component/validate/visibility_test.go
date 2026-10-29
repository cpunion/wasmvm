package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPrivateAndPublishedValueNames(t *testing.T) {
	for _, d := range []*model.TypeDef{
		{Kind: model.Record, Fields: []model.NamedType{{Name: "x", Type: *ref(model.U32)}}},
		{Kind: model.Enum, Labels: []string{"a", "b"}},
		{Kind: model.Flags, Labels: []string{"read"}},
		{Kind: model.Variant, Cases: []model.TypeCase{{Name: "a"}, {Name: "b", Type: ref(model.U32)}}},
	} {
		s := scope(t)
		original := define(t, s, d)
		function := define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(0)})
		if _, err := s.Declare(external("private", model.FunctionSort, 1), true); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(d, err)
		}
		published, err := s.Declare(external("value", model.TypeSort, 0), true)
		if err != nil || published.Value == original.Value {
			t.Fatal(d, published, err)
		}
		if _, err := s.Declare(external("reimported-type", model.TypeSort, 2), false); err != nil {
			t.Fatal(err)
		}
		if ok, err := Equal(original, published, 100); err != nil || !ok {
			t.Fatal("publication changed structural equality", ok, err)
		}
		if err := s.checkVisible(function, true, false); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("publication named private binding", err)
		}
		index := uint32(len(s.types))
		define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(2)})
		if _, err := s.Declare(external("visible", model.FunctionSort, index), true); err != nil {
			t.Fatal(d, err)
		}
		if _, err := s.Declare(external("back-edge", model.FunctionSort, index), false); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("import depended on export", err)
		}
		importedIndex := uint32(len(s.types))
		if _, err := s.Declare(external("input-type", model.TypeSort, 0), false); err != nil {
			t.Fatal(err)
		}
		fnIndex := uint32(len(s.types))
		define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(importedIndex)})
		for _, exporting := range []bool{false, true} {
			if _, err := s.Declare(external("input-func", model.FunctionSort, fnIndex), exporting); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestResourceNamesAreSeparateFromNominalIdentity(t *testing.T) {
	s := scope(t)
	private := define(t, s, &model.TypeDef{Kind: model.Resource, Representation: wasm.I32})
	define(t, s, &model.TypeDef{Kind: model.Own, Resource: 0})
	privateFunc := define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(1)})
	public, err := s.Declare(external("resource", model.TypeSort, 0), true)
	if err != nil || private == public || private.Resource != public.Resource {
		t.Fatal(private, public, err)
	}
	if err := s.checkVisible(privateFunc, true, false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	publicIndex := uint32(len(s.types) - 1)
	if _, err := s.Declare(external("back-edge-resource", model.TypeSort, publicIndex), false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	ownedIndex := uint32(len(s.types))
	define(t, s, &model.TypeDef{Kind: model.Own, Resource: publicIndex})
	fnIndex := uint32(len(s.types))
	define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(ownedIndex)})
	if _, err := s.Declare(external("output", model.FunctionSort, fnIndex), true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Declare(external("input", model.FunctionSort, fnIndex), false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestDeferredInstanceVisibilityAndComponentBoundaries(t *testing.T) {
	s := scope(t)
	d := &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{
		{Kind: 1, Type: &model.TypeDef{Kind: model.Record, Fields: []model.NamedType{{Name: "x", Type: *ref(model.U32)}}}},
		{Kind: 1, Type: &model.TypeDef{Kind: model.Function, Result: idx(0)}},
		{Kind: 4, Export: external("f", model.FunctionSort, 1)},
	}}
	define(t, s, d)
	if _, err := s.Declare(external("instance", model.InstanceSort, 0), false); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Declare(external("instance-type", model.TypeSort, 0), true); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	d.Kind = model.Component
	if _, err := s.Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	define(t, s, &model.TypeDef{Kind: model.Record, Fields: []model.NamedType{{Name: "x", Type: *ref(model.U32)}}})
	publicIndex := uint32(len(s.types))
	if _, err := s.Declare(external("record", model.TypeSort, 1), false); err != nil {
		t.Fatal(err)
	}
	for _, component := range []bool{false, true} {
		child, err := s.pool.Scope(s, component)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := child.Outer(1, publicIndex); err != nil {
			t.Fatal(err)
		}
		f := define(t, child, &model.TypeDef{Kind: model.Function, Result: idx(0)})
		err = child.checkVisible(f, false, false)
		if component && !errors.Is(err, wasm.ErrInvalid) || !component && err != nil {
			t.Fatal(component, err)
		}
	}
}
