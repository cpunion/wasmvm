package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func ref(kind model.Kind) *model.TypeRef { return &model.TypeRef{Primitive: kind} }
func idx(index uint32) *model.TypeRef    { return &model.TypeRef{Indexed: true, Index: index} }
func scope(t *testing.T) *TypeScope {
	t.Helper()
	s, err := NewTypePool(DefaultLimits()).Scope(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func define(t *testing.T, s *TypeScope, d *model.TypeDef) *Type {
	t.Helper()
	v, err := s.Define(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestValueTypesAndOwnedMetadata(t *testing.T) {
	s := scope(t)
	for kind := model.String; kind <= model.Bool; kind++ {
		a, err := s.Value(*ref(kind))
		b, again := s.Value(*ref(kind))
		if err != nil || again != nil || a != b || a.Kind != kind {
			t.Fatal(kind, a, b, err, again)
		}
		v := define(t, s, &model.TypeDef{Kind: kind})
		if v.Value.Kind != kind {
			t.Fatal(v)
		}
	}
	for _, d := range []*model.TypeDef{
		{Kind: model.ErrorContext},
		{Kind: model.Record, Fields: []model.NamedType{{Name: "is-XML", Type: *ref(model.Bool)}}},
		{Kind: model.Tuple, Tuple: []model.TypeRef{*ref(model.U64), *ref(model.String)}},
		{Kind: model.Variant, Cases: []model.TypeCase{{Name: "absent"}, {Name: "present", Type: ref(model.String)}}},
		{Kind: model.Result, Cases: []model.TypeCase{{Type: ref(model.U32)}, {Type: ref(model.String)}}},
		{Kind: model.Result, Cases: []model.TypeCase{{}, {}}},
		{Kind: model.List, Element: ref(model.String)},
		{Kind: model.FixedList, Element: ref(model.U8), Length: 16},
		{Kind: model.Map, Key: ref(model.String), Element: ref(model.Bool)},
		{Kind: model.Map, Key: ref(model.S64), Element: ref(model.Bool)},
		{Kind: model.Option, Element: ref(model.Bool)},
		{Kind: model.Stream}, {Kind: model.Future, Element: ref(model.Char)},
		{Kind: model.Function, Parameters: []model.NamedType{{Name: "input", Type: *ref(model.String)}}, Result: ref(model.String)},
		{Kind: model.AsyncFunction},
	} {
		v := define(t, s, d)
		if v.Kind != d.Kind || v.owner != s.pool {
			t.Fatal(v)
		}
	}
	labels := []string{"read", "write"}
	flags := define(t, s, &model.TypeDef{Kind: model.Flags, Labels: labels})
	labels[0] = "changed"
	if flags.Value.Labels[0] != "read" {
		t.Fatal("retained mutable input", flags)
	}
	enum := define(t, s, &model.TypeDef{Kind: model.Enum, Labels: []string{"no", "yes"}})
	list := define(t, s, &model.TypeDef{Kind: model.List, Element: idx(uint32(len(s.types) - 1))})
	if list.Value.Element != enum.Value {
		t.Fatal("lost DAG sharing")
	}
	if _, err := s.Alias(flags); err != nil {
		t.Fatal(err)
	}
	if got, err := s.At(uint32(len(s.types) - 1)); err != nil || got != flags {
		t.Fatal(got, err)
	}
}

func TestInvalidTypesDoNotBind(t *testing.T) {
	for _, d := range []*model.TypeDef{
		nil, {Kind: model.Record}, {Kind: model.Tuple}, {Kind: model.Variant}, {Kind: model.Flags}, {Kind: model.Enum},
		{Kind: model.Result}, {Kind: model.Result, Cases: []model.TypeCase{{}}},
		{Kind: model.Flags, Labels: make([]string, 33)},
		{Kind: model.List}, {Kind: model.Option}, {Kind: model.Map, Element: ref(model.U8)},
		{Kind: model.FixedList, Element: ref(model.U8)},
		{Kind: model.Map, Key: ref(model.F32), Element: ref(model.U8)},
		{Kind: model.Map, Key: idx(0), Element: ref(model.U8)},
		{Kind: model.Option, Element: idx(0)},
		{Kind: model.Tuple, Tuple: []model.TypeRef{*idx(0)}},
		{Kind: model.Variant, Cases: []model.TypeCase{{Name: "bad", Type: idx(0)}}},
		{Kind: model.Record, Fields: []model.NamedType{{Name: "valid", Type: *idx(0)}}},
		{Kind: model.Own}, {Kind: model.Borrow},
		{Kind: model.Stream, Element: ref(model.Char)},
		{Kind: model.Resource, Representation: wasm.F32},
		{Kind: model.Resource, Representation: wasm.I32, Destructor: new(uint32)},
	} {
		s := scope(t)
		v, err := s.Define(d, nil)
		if !errors.Is(err, wasm.ErrInvalid) || v != nil || len(s.types) != 0 {
			t.Fatal(d, v, err, s.types)
		}
	}
	for _, d := range []*model.TypeDef{{Kind: 0}} {
		if _, err := scope(t).Define(d, nil); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	s := scope(t)
	define(t, s, &model.TypeDef{Kind: model.Function})
	for _, d := range []*model.TypeDef{{Kind: model.List, Element: idx(0)}, {Kind: model.Own, Resource: 0}} {
		if _, err := s.Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := s.Value(model.TypeRef{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestLabels(t *testing.T) {
	for _, name := range []string{"a", "a-b-c", "a1-2-3", "A", "A-B-C", "A1-2-3", "a11-w0rds", "A11-4CR0NYMS", "m1x3d-4CR0NYMS"} {
		if err := label(name, map[string]bool{}); err != nil {
			t.Fatal(name, err)
		}
	}
	for _, name := range []string{"", "-a", "a-", "a--b", "a_b", "1-2-3", "NevEr", "GoNnA", "GIVe", "yOu", "a.1", "é", "foo/"} {
		for _, kind := range []model.Kind{model.Record, model.Variant, model.Flags, model.Enum, model.Function} {
			d := &model.TypeDef{Kind: kind, Fields: []model.NamedType{{Name: name, Type: *ref(model.U8)}}, Cases: []model.TypeCase{{Name: name}}, Labels: []string{name}, Parameters: []model.NamedType{{Name: name, Type: *ref(model.U8)}}}
			if _, err := scope(t).Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(name, kind, err)
			}
		}
	}
	for _, kind := range []model.Kind{model.Record, model.Variant, model.Flags, model.Enum, model.Function} {
		d := &model.TypeDef{Kind: kind, Fields: []model.NamedType{{Name: "a-B-c-D", Type: *ref(model.U8)}, {Name: "A-b-C-d", Type: *ref(model.U8)}}, Cases: []model.TypeCase{{Name: "a-B-c-D"}, {Name: "A-b-C-d"}}, Labels: []string{"a-B-c-D", "A-b-C-d"}, Parameters: []model.NamedType{{Name: "a-B-c-D", Type: *ref(model.U8)}, {Name: "A-b-C-d", Type: *ref(model.U8)}}}
		if _, err := scope(t).Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(kind, err)
		}
	}
}

func TestResourcesBorrowClosureAndScopes(t *testing.T) {
	s := scope(t)
	r1 := define(t, s, &model.TypeDef{Kind: model.Resource, Representation: wasm.I64})
	r2, err := s.Abstract()
	if err != nil || r1.Resource == r2.Resource {
		t.Fatal(r1, r2, err)
	}
	index := uint32(0)
	r3, err := s.Define(&model.TypeDef{Kind: model.Resource, Representation: wasm.I32, Destructor: &index}, []module.FuncType{{Params: []wasm.ValueType{wasm.I32}}})
	index = 999
	if err != nil || *r3.Destructor != 0 {
		t.Fatal(r3, err)
	}
	for _, signature := range []module.FuncType{{}, {Params: []wasm.ValueType{wasm.I64}}, {Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, {Params: []wasm.ValueType{wasm.I32, wasm.I32}}} {
		index = 0
		if _, err := s.Define(&model.TypeDef{Kind: model.Resource, Representation: wasm.I32, Destructor: &index}, []module.FuncType{signature}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(signature, err)
		}
	}
	borrowIndex := uint32(len(s.types))
	borrow := define(t, s, &model.TypeDef{Kind: model.Borrow, Resource: 0})
	define(t, s, &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "self", Type: *idx(borrowIndex)}}})
	for _, kind := range []model.Kind{model.List, model.Option, model.FixedList, model.Map, model.Tuple, model.Record, model.Variant, model.Result} {
		i := uint32(len(s.types))
		define(t, s, &model.TypeDef{Kind: kind, Element: idx(borrowIndex), Length: 2, Key: ref(model.U8), Tuple: []model.TypeRef{*idx(borrowIndex)}, Fields: []model.NamedType{{Name: "value", Type: *idx(borrowIndex)}}, Cases: []model.TypeCase{{Name: "one", Type: idx(borrowIndex)}, {Name: "two"}}})
		for _, d := range []*model.TypeDef{{Kind: model.Function, Result: idx(i)}, {Kind: model.Stream, Element: idx(i)}, {Kind: model.Future, Element: idx(i)}} {
			if _, err := s.Define(d, nil); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(kind, d, err)
			}
		}
	}
	if !borrow.borrow || !borrow.resource {
		t.Fatal(borrow)
	}
	decl, err := s.pool.Scope(s, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decl.Define(&model.TypeDef{Kind: model.Resource, Representation: wasm.I32}, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if got, err := decl.Outer(1, 0); err != nil || got != r1 {
		t.Fatal(got, err)
	}
	nested, err := s.pool.Scope(decl, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := nested.Outer(1, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	plain := define(t, s, &model.TypeDef{Kind: model.U8})
	if got, err := nested.Outer(2, uint32(len(s.types)-1)); err != nil || got != plain {
		t.Fatal(got, err)
	}
	if _, err := s.Outer(0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Outer(1, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := nested.Outer(0, 100); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Alias(nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := s.Alias(define(t, scope(t), &model.TypeDef{Kind: model.U8})); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := NewTypePool(DefaultLimits()).Scope(s, true); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestTypeLimitsAndDAGs(t *testing.T) {
	for _, limits := range []Limits{{}, {MaxNodes: 100}, {MaxNodes: 1, MaxDepth: 1}} {
		p := NewTypePool(limits)
		s, err := p.Scope(nil, true)
		if err == nil {
			_, err = s.Define(&model.TypeDef{Kind: model.U8}, nil)
		}
		if !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(limits, err)
		}
	}
	p := NewTypePool(Limits{MaxNodes: 100, MaxDepth: 2})
	s, _ := p.Scope(nil, true)
	child, _ := p.Scope(s, false)
	if _, err := p.Scope(child, false); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	define(t, s, &model.TypeDef{Kind: model.U8})
	define(t, s, &model.TypeDef{Kind: model.List, Element: idx(0)})
	if _, err := s.Define(&model.TypeDef{Kind: model.List, Element: idx(1)}, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	s = scope(t)
	define(t, s, &model.TypeDef{Kind: model.U8})
	for i := uint32(0); i < 27; i++ {
		define(t, s, &model.TypeDef{Kind: model.Tuple, Tuple: []model.TypeRef{*idx(i), *idx(i)}})
	}
	if _, err := s.Define(&model.TypeDef{Kind: model.Tuple, Tuple: []model.TypeRef{*idx(27), *idx(27)}}, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("accepted 2^28 byte value", err)
	}
	if _, err := scope(t).Define(&model.TypeDef{Kind: model.FixedList, Element: ref(model.U64), Length: 1 << 25}, nil); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("memory64 size", err)
	}
	for _, d := range []*model.TypeDef{
		{Kind: model.Record, Fields: []model.NamedType{{Name: "a", Type: *ref(model.U8)}}},
		{Kind: model.Tuple, Tuple: []model.TypeRef{*ref(model.U8)}},
		{Kind: model.Variant, Cases: []model.TypeCase{{Name: "a"}}},
		{Kind: model.Enum, Labels: []string{"a"}},
	} {
		s, _ := NewTypePool(Limits{MaxNodes: 2, MaxDepth: 10}).Scope(nil, true)
		if _, err := s.Define(d, nil); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(d, err)
		}
	}
}
