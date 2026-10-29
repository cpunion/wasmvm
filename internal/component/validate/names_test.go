package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestInterfaceNamesVersionsAndAttributes(t *testing.T) {
	for _, row := range []struct{ name, implements, suffix string }{
		{"wasi:http/types", "", ""}, {"wasi:http/types@1.0.0", "", ""},
		{"a-b:c-d/e-f@123456.7890.488", "", ""}, {"a:b/c@0.0.0-abcd.1.2+efg.4.ee.5", "", ""},
		{"a:b/c@1", "", ".2.3"}, {"a:b/c@0.2", "", ".6+sha.5114f85"},
		{"a:b/c@0.0.1", "", "+b"}, {"a:b/c@0.0.0", "", ""},
		{"a:b/c@1.2.3-alpha.1", "", "+sha.5114"},
		{"a", "a:b/c", ""}, {"a", "a:b/c@1", ".2.3"}, {"a", "a:b/c@0.2", ".1"},
		{"a", "a:b/c@0.0.1", ""}, {"a", "a:b/c@1.2.3", ""},
	} {
		n := model.Name{Text: row.name, Attributes: []model.Attribute{{Kind: 2, Value: "☃︎"}}}
		if row.implements != "" {
			n.Attributes = append(n.Attributes, model.Attribute{Kind: 0, Value: row.implements})
		}
		if row.suffix != "" {
			n.Attributes = append(n.Attributes, model.Attribute{Kind: 1, Value: row.suffix})
		}
		s := scope(t)
		tpe := define(t, s, &model.TypeDef{Kind: model.Instance})
		if err := s.imports.add(n, model.InstanceSort, tpe, s.pool); err != nil {
			t.Fatal(row, err)
		}
		n.Attributes[0].Value = "changed"
		if s.imports.ordered[0].Name.Attributes[0].Value != "☃︎" {
			t.Fatal("retained mutable attributes")
		}
	}
	for _, name := range []string{"a:b/c@", "a:b/c@.", "a:b/c@1.", "a:b/c@a.2", "a:b/c@2.b", "a:b/c@2.0x0", "a:b/c@2.0.0+", "a:b/c@2.0.0-", "a:b/c@01.0.0", "a:b/c@1.0.0-01", "a:b/c@1.0.0+a..b", "a:b/c@1.0.0-a_", "a:b/c@1.0.0+a+", "a:b/c@18446744073709551616.0.0", "foo:bar:baz/qux", "foo:bar/baz/qux", "a:b", "A:b/c", "a:B/c", "a:/c", "a:b/", "a", "a:b/c@1", "a:b/c@0.2"} {
		s := scope(t)
		tpe := define(t, s, &model.TypeDef{Kind: model.Instance})
		attrs := []model.Attribute{}
		if name == "a" {
			attrs = append(attrs, model.Attribute{Kind: 0, Value: "not-valid"})
		}
		if err := s.imports.add(model.Name{Text: name, Attributes: attrs}, model.InstanceSort, tpe, s.pool); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(name, err)
		}
	}
	for _, row := range []struct {
		name  string
		sort  model.Sort
		attrs []model.Attribute
	}{
		{"a", model.InstanceSort, []model.Attribute{{Kind: 0, Value: ""}}},
		{"a", model.FunctionSort, []model.Attribute{{Kind: 0, Value: "a:b/c"}}},
		{"a:b/c", model.InstanceSort, []model.Attribute{{Kind: 0, Value: "a:b/c"}}},
		{"a", model.InstanceSort, []model.Attribute{{Kind: 3}}},
		{"a", model.InstanceSort, []model.Attribute{{Kind: 2}, {Kind: 2}}},
		{"a", model.InstanceSort, []model.Attribute{{Kind: 1}}},
		{"a", model.FunctionSort, []model.Attribute{{Kind: 1, Value: ".2.3"}}},
		{"a", model.InstanceSort, []model.Attribute{{Kind: 1, Value: ".2.3"}}},
		{"a:b/c", model.InstanceSort, []model.Attribute{{Kind: 1, Value: ".2.3"}}},
		{"a:b/c@1", model.InstanceSort, []model.Attribute{{Kind: 1, Value: ".2.3-alpha"}}},
		{"a:b/c@0.0.1", model.InstanceSort, []model.Attribute{{Kind: 1, Value: "-alpha"}}},
		{"a:b/c@1.2.3", model.InstanceSort, []model.Attribute{{Kind: 1, Value: "+sha"}}},
		{"a:b/c@0", model.InstanceSort, []model.Attribute{{Kind: 1, Value: ".1.0"}}},
	} {
		s := scope(t)
		if err := s.imports.add(model.Name{Text: row.name, Attributes: row.attrs}, row.sort, define(t, s, &model.TypeDef{Kind: model.Function}), s.pool); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(row, err)
		}
	}
}

func TestAnnotatedFunctionsAndStrongUniqueness(t *testing.T) {
	s := scope(t)
	r, err := s.Declare(&model.Import{Name: model.Name{Text: "file"}, Type: model.ExternType{Sort: model.TypeSort, Bound: 1}}, false)
	if err != nil {
		t.Fatal(err)
	}
	own := define(t, s, &model.TypeDef{Kind: model.Own, Resource: 0})
	borrow := define(t, s, &model.TypeDef{Kind: model.Borrow, Resource: 0})
	ctor := define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(1)})
	method := define(t, s, &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "self", Type: *idx(2)}}})
	plain := define(t, s, &model.TypeDef{Kind: model.Function})
	if own.Value.ResourceID != r.Resource || borrow.Value.ResourceID != r.Resource {
		t.Fatal("handle identity")
	}
	for _, row := range []struct {
		name string
		tpe  *Type
	}{
		{"[constructor]file", ctor}, {"[method]file.read", method}, {"[static]file.create", plain},
	} {
		if err := s.imports.add(model.Name{Text: row.name}, model.FunctionSort, row.tpe, s.pool); err != nil {
			t.Fatal(row.name, err)
		}
	}
	for _, name := range []string{"[constructor]", "[method]file", "[method]file.", "[method].read", "[static]file.a.b", "[method][get][set]file.a", "[unknown]file"} {
		if _, err := parseName(name); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(name, err)
		}
	}
	for _, row := range []struct {
		name string
		sort model.Sort
		tpe  *Type
	}{
		{"[constructor]file", model.InstanceSort, plain},
		{"[constructor]missing", model.FunctionSort, ctor},
		{"[constructor]file", model.FunctionSort, plain},
		{"[method]file.close", model.FunctionSort, plain},
		{"[method]file.close", model.FunctionSort, ctor},
		{"[static]missing.close", model.FunctionSort, plain},
	} {
		n := nameSet{}
		if err := n.add(model.Name{Text: "file"}, model.TypeSort, r, s.pool); err != nil {
			t.Fatal(err)
		}
		if err := n.add(model.Name{Text: row.name}, row.sort, row.tpe, s.pool); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(row, err)
		}
	}
	for _, names := range [][2]string{{"a-b", "AB"}, {"file", "[static]file.file"}, {"[method]file.read", "[static]file.read"}} {
		a, err := parseName(names[0])
		if err != nil {
			t.Fatal(err)
		}
		b, err := parseName(names[1])
		if err != nil {
			t.Fatal(err)
		}
		if a.key != b.key {
			t.Fatal(names, a, b)
		}
	}
	if err := s.imports.add(model.Name{Text: "FILE"}, model.FunctionSort, plain, s.pool); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestPropertySignatures(t *testing.T) {
	s := scope(t)
	getter := define(t, s, &model.TypeDef{Kind: model.Function, Result: ref(model.String)})
	setter := define(t, s, &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "value", Type: *ref(model.String)}}})
	if err := s.imports.add(model.Name{Text: "[get]name"}, model.FunctionSort, getter, s.pool); err != nil {
		t.Fatal(err)
	}
	if err := s.imports.add(model.Name{Text: "[set]name"}, model.FunctionSort, setter, s.pool); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name string
		d    *model.TypeDef
	}{
		{"[get]absent", &model.TypeDef{Kind: model.Function}},
		{"[get]extra", &model.TypeDef{Kind: model.Function, Result: ref(model.U8), Parameters: []model.NamedType{{Name: "x", Type: *ref(model.U8)}}}},
		{"[get]async", &model.TypeDef{Kind: model.AsyncFunction, Result: ref(model.U8)}},
		{"[set]absent", &model.TypeDef{Kind: model.Function}},
		{"[set]name", &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "value", Type: *ref(model.U8)}}}},
		{"[set]name", &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "value", Type: *ref(model.String)}}, Result: ref(model.Bool)}},
	} {
		n := nameSet{}
		if err := n.add(model.Name{Text: "[get]name"}, model.FunctionSort, getter, s.pool); err != nil {
			t.Fatal(err)
		}
		if err := n.add(model.Name{Text: row.name}, model.FunctionSort, define(t, s, row.d), s.pool); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(row, err)
		}
	}
	resultIndex := uint32(len(s.types))
	define(t, s, &model.TypeDef{Kind: model.Result, Cases: []model.TypeCase{{Type: ref(model.String)}, {Type: ref(model.U8)}}})
	wrapped := define(t, s, &model.TypeDef{Kind: model.Function, Result: idx(resultIndex)})
	n := nameSet{}
	if err := n.add(model.Name{Text: "[get]value"}, model.FunctionSort, wrapped, s.pool); err != nil {
		t.Fatal(err)
	}
	if err := n.add(model.Name{Text: "[set]value"}, model.FunctionSort, setter, s.pool); err != nil {
		t.Fatal(err)
	}
}
