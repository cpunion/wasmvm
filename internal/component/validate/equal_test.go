package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestEqualityNominalityAndShapes(t *testing.T) {
	s := scope(t)
	for _, d := range []*model.TypeDef{
		{Kind: model.U8}, {Kind: model.String},
		{Kind: model.Tuple, Tuple: []model.TypeRef{*ref(model.U8), *ref(model.U8)}},
		{Kind: model.Record, Fields: []model.NamedType{{Name: "x", Type: *ref(model.U8)}}},
		{Kind: model.Variant, Cases: []model.TypeCase{{Name: "yes", Type: ref(model.Bool)}, {Name: "no"}}},
		{Kind: model.Result, Cases: []model.TypeCase{{}, {Type: ref(model.U8)}}},
		{Kind: model.Enum, Labels: []string{"a", "b"}},
		{Kind: model.Flags, Labels: []string{"a", "b"}},
		{Kind: model.List, Element: ref(model.U8)},
		{Kind: model.Map, Key: ref(model.U8), Element: ref(model.Bool)},
		{Kind: model.FixedList, Element: ref(model.U8), Length: 2},
		{Kind: model.Function, Parameters: []model.NamedType{{Name: "x", Type: *ref(model.U8)}}, Result: ref(model.Bool)},
		{Kind: model.AsyncFunction},
	} {
		a, b := define(t, s, d), define(t, s, d)
		if ok, err := Equal(a, b, 100); err != nil || !ok {
			t.Fatal(d, ok, err)
		}
		if ok, err := Equal(a, a, 100); err != nil || !ok {
			t.Fatal(ok, err)
		}
	}
	a, _ := s.Abstract()
	b, _ := s.Abstract()
	c, _ := s.Alias(a)
	for _, row := range []struct {
		x, y *Type
		want bool
	}{{nil, nil, true}, {a, nil, false}, {nil, a, false}, {a, b, false}, {a, c, true}, {a, a, true}} {
		if ok, err := Equal(row.x, row.y, 100); err != nil || ok != row.want {
			t.Fatal(row, ok, err)
		}
	}
	other := scope(t)
	foreign, _ := other.Abstract()
	if ok, _ := Equal(a, foreign, 100); ok {
		t.Fatal("resource ID collision between compilations")
	}
	localOwn := define(t, s, &model.TypeDef{Kind: model.Own, Resource: uint32(len(s.types) - 1)})
	foreignOwn := define(t, other, &model.TypeDef{Kind: model.Own, Resource: 0})
	if ok, _ := Equal(localOwn, foreignOwn, 100); ok {
		t.Fatal("cross-compilation handle identity")
	}
	for _, ds := range [][2]*model.TypeDef{
		{{Kind: model.U8}, {Kind: model.U16}},
		{{Kind: model.List, Element: ref(model.U8)}, {Kind: model.List, Element: ref(model.U16)}},
		{{Kind: model.Map, Element: ref(model.U8), Key: ref(model.Bool)}, {Kind: model.Map, Element: ref(model.U8), Key: ref(model.Char)}},
		{{Kind: model.Enum, Labels: []string{"a"}}, {Kind: model.Enum, Labels: []string{"b"}}},
		{{Kind: model.Record, Fields: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}}, {Kind: model.Record, Fields: []model.NamedType{{Name: "b", Type: *ref(model.U8)}}}},
		{{Kind: model.Record, Fields: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}}, {Kind: model.Record, Fields: []model.NamedType{{Name: "a", Type: *ref(model.U16)}}}},
		{{Kind: model.Option, Element: ref(model.U8)}, {Kind: model.Option, Element: ref(model.U16)}},
		{{Kind: model.Function}, {Kind: model.Function, Parameters: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}}},
		{{Kind: model.Function, Result: ref(model.U8)}, {Kind: model.Function}},
		{{Kind: model.Function, Parameters: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}}, {Kind: model.Function, Parameters: []model.NamedType{{Name: "b", Type: *ref(model.U8)}}}},
		{{Kind: model.Function, Parameters: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}}, {Kind: model.Function, Parameters: []model.NamedType{{Name: "a", Type: *ref(model.U16)}}}},
	} {
		if ok, err := Equal(define(t, s, ds[0]), define(t, s, ds[1]), 100); err != nil || ok {
			t.Fatal(ds, ok, err)
		}
	}
}

func TestEqualityWorkIsBoundedForSharedGraphs(t *testing.T) {
	a, b := scope(t), scope(t)
	build := func(s *TypeScope) *Type {
		v := define(t, s, &model.TypeDef{Kind: model.U8})
		for i := uint32(0); i < 24; i++ {
			v = define(t, s, &model.TypeDef{Kind: model.Tuple, Tuple: []model.TypeRef{*idx(i), *idx(i)}})
		}
		return v
	}
	x, y := build(a), build(b)
	if ok, err := Equal(x, y, 100); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if ok, err := Equal(x, y, 20); !errors.Is(err, wasm.ErrLimit) || ok {
		t.Fatal(ok, err)
	}
	for _, d := range []*model.TypeDef{
		{Kind: model.Enum, Labels: []string{"a", "b"}},
		{Kind: model.Tuple, Tuple: []model.TypeRef{*ref(model.U8)}},
		{Kind: model.Function, Parameters: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}},
	} {
		x, y := define(t, a, d), define(t, b, d)
		for _, budget := range []uint32{0, 1} {
			if _, err := Equal(x, y, budget); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(d, budget, err)
			}
		}
	}
}
