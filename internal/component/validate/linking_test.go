package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func coreFunction(typ uint32) module.Function {
	return module.Function{Type: typ, Body: module.Expr{{Op: 0}, {Op: 0x0b}}}
}

func composedCore() *model.Binary {
	provider := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, {Params: []wasm.ValueType{wasm.I32}}},
		Functions: []module.Function{coreFunction(0)}, Memories: []module.Limits{{Min: 1}}, Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}},
		Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: module.Expr{{Op: 0x41}, {Op: 0x0b}}}}, Tags: []uint32{1}}
	consumer := &module.Module{Types: provider.Types}
	names := []string{"f", "table", "memory", "global", "tag"}
	for i, name := range names {
		provider.Exports = append(provider.Exports, module.Export{Name: name, Kind: byte(i)})
		im := module.Import{Module: "env", Name: name, Kind: byte(i)}
		switch i {
		case 1:
			im.Table = provider.Tables[0]
		case 2:
			im.Memory = provider.Memories[0]
		case 3:
			im.Global = provider.Globals[0].Type
		case 4:
			im.Type = 1
		}
		consumer.Imports = append(consumer.Imports, im)
		consumer.Exports = append(consumer.Exports, module.Export{Name: name, Kind: byte(i)})
	}
	b := &model.Binary{Definitions: []model.Definition{
		{Kind: 1, CoreModule: provider},
		{Kind: 2, Instance: &model.InstanceExpr{Instantiate: true}},
		{Kind: 1, CoreModule: consumer},
		{Kind: 2, Instance: &model.InstanceExpr{Instantiate: true, Index: 1, Arguments: []model.Argument{{Name: "env", Item: model.SortIndex{Sort: model.CoreInstanceSort}}, {Name: "NotKebab-Case", Item: model.SortIndex{Sort: model.CoreInstanceSort}}}}},
	}}
	bag := &model.InstanceExpr{}
	for i, name := range names {
		sort := model.CoreFunctionSort + model.Sort(i)
		b.Definitions = append(b.Definitions, model.Definition{Kind: 6, Alias: &model.Alias{Kind: 1, Sort: sort, Instance: 1, Name: name}})
		bag.Exports = append(bag.Exports, model.Export{Name: model.Name{Text: name}, Item: model.SortIndex{Sort: sort}})
	}
	b.Definitions = append(b.Definitions, model.Definition{Kind: 2, Instance: bag}, model.Definition{Kind: 11, Export: &model.Export{Name: model.Name{Text: "library"}, Item: model.SortIndex{Sort: model.CoreModuleSort, Index: 1}}})
	return b
}

func TestCoreCompositionAllSortsAndNoExecution(t *testing.T) {
	b := composedCore()
	c, err := Check(b, DefaultLimits())
	if err != nil || c.Definition() != b || len(c.CoreInstances) != 3 || len(c.CoreModules) != 2 || len(c.Type.Exports) != 1 {
		t.Fatal(c, err)
	}
	for _, name := range []string{"f", "table", "memory", "global", "tag"} {
		if _, ok := c.CoreInstances[2].byName[name]; !ok {
			t.Fatal(name)
		}
	}
	// This start would trap if instantiated. Static validation must not run it.
	start := &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{coreFunction(0)}, HasStart: true}
	if _, err := Check(&model.Binary{Definitions: []model.Definition{{Kind: 1, CoreModule: start}}}, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
}

func TestCoreCompositionRejectsBadBindings(t *testing.T) {
	for _, mutate := range []func(*model.Binary){
		func(b *model.Binary) { b.Definitions[3].Instance.Index = 100 },
		func(b *model.Binary) { b.Definitions[3].Instance.Arguments = nil },
		func(b *model.Binary) { b.Definitions[3].Instance.Arguments[1].Name = "env" },
		func(b *model.Binary) { b.Definitions[3].Instance.Arguments[0].Item.Index = 100 },
		func(b *model.Binary) { b.Definitions[3].Instance.Arguments[0].Item.Sort = model.InstanceSort },
		func(b *model.Binary) { b.Definitions[2].CoreModule.Imports[0].Name = "missing" },
		func(b *model.Binary) { b.Definitions[2].CoreModule.Imports[2].Memory.Min = 2 },
		func(b *model.Binary) { b.Definitions[4].Alias.Instance = 100 },
		func(b *model.Binary) { b.Definitions[4].Alias.Name = "missing" },
		func(b *model.Binary) { b.Definitions[4].Alias.Sort = model.CoreGlobalSort },
		func(b *model.Binary) { b.Definitions[9].Instance.Exports[1].Name.Text = "f" },
		func(b *model.Binary) { b.Definitions[9].Instance.Exports[0].Item.Index = 100 },
		func(b *model.Binary) { b.Definitions[9].Instance.Exports[0].Item.Sort = model.FunctionSort },
	} {
		b := composedCore()
		mutate(b)
		c, err := Check(b, DefaultLimits())
		if c != nil || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(c, err)
		}
	}
}

func TestCanonicalMemory64AndResourceBuiltins(t *testing.T) {
	m := &module.Module{Types: []module.FuncType{
		{Params: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64}, Results: []wasm.ValueType{wasm.I64}},
		{Params: []wasm.ValueType{wasm.I64, wasm.I64}}, {},
	}, Functions: []module.Function{coreFunction(0), coreFunction(1), coreFunction(2)}, Memories: []module.Limits{{Min: 1, Index64: true}}, Exports: []module.Export{{Name: "realloc", Kind: 0}, {Name: "f", Kind: 0, Index: 1}, {Name: "post", Kind: 0, Index: 2}, {Name: "mem", Kind: 2}}}
	b := &model.Binary{Definitions: []model.Definition{{Kind: 1, CoreModule: m}, {Kind: 2, Instance: &model.InstanceExpr{Instantiate: true}}}}
	for _, name := range []string{"realloc", "f", "post"} {
		b.Definitions = append(b.Definitions, model.Definition{Kind: 6, Alias: &model.Alias{Kind: 1, Sort: model.CoreFunctionSort, Name: name}})
	}
	b.Definitions = append(b.Definitions,
		model.Definition{Kind: 6, Alias: &model.Alias{Kind: 1, Sort: model.CoreMemorySort, Name: "mem"}},
		model.Definition{Kind: 7, Type: &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "msg", Type: *ref(model.String)}}}},
		model.Definition{Kind: 8, Canonical: &model.Canonical{Op: 0, Index: 1, Options: []model.CanonicalOption{{Kind: 2}, {Kind: 3}, {Kind: 4}, {Kind: 5, Index: 2}}}},
	)
	c, err := Check(b, DefaultLimits())
	if err != nil || len(c.Canonical) != 1 || !c.Canonical[0].Memory64 || c.Canonical[0].Encoding != 2 || len(c.Canonical[0].ABI.Params) != 2 || c.Canonical[0].ABI.Params[0] != wasm.I64 {
		t.Fatal(c, err)
	}
	for _, rep := range []wasm.ValueType{wasm.I32, wasm.I64} {
		r := &model.Binary{Definitions: []model.Definition{{Kind: 7, Type: &model.TypeDef{Kind: model.Resource, Representation: rep}}, {Kind: 8, Canonical: &model.Canonical{Op: 2}}, {Kind: 8, Canonical: &model.Canonical{Op: 4}}, {Kind: 8, Canonical: &model.Canonical{Op: 3}}}}
		c, err := Check(r, DefaultLimits())
		if err != nil || len(c.Canonical) != 3 || c.Canonical[0].Signature.Params[0] != rep || c.Canonical[1].Signature.Results[0] != rep {
			t.Fatal(c, err)
		}
	}
}

func TestComponentCheckLimitsAndUnsupported(t *testing.T) {
	if _, err := Check(nil, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := Check(&model.Binary{}, Limits{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	for _, kind := range []byte{9, 12, 255} {
		c, err := Check(&model.Binary{Definitions: []model.Definition{{Kind: kind}}}, DefaultLimits())
		if c != nil || !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(kind, c, err)
		}
	}
	for _, d := range []model.Definition{{Kind: 1}, {Kind: 2}, {Kind: 6}, {Kind: 8}, {Kind: 11}, {Kind: 8, Canonical: &model.Canonical{Op: 3}}, {Kind: 8, Canonical: &model.Canonical{Op: 1}}} {
		if _, err := Check(&model.Binary{Definitions: []model.Definition{d}}, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(d, err)
		}
	}
	for budget := uint32(1); budget < 100; budget++ {
		_, err := Check(composedCore(), Limits{MaxNodes: budget, MaxDepth: 128})
		if err != nil && !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(budget, err)
		}
	}
}

func TestNestedCoreAliasesAndImportedFunctionAliases(t *testing.T) {
	child := &model.Binary{Definitions: []model.Definition{
		{Kind: 6, Alias: &model.Alias{Kind: 2, Sort: model.CoreModuleSort, Count: 1}},
		{Kind: 2, Instance: &model.InstanceExpr{Instantiate: true}},
	}}
	root := &model.Binary{Definitions: []model.Definition{{Kind: 1, CoreModule: new(module.Module)}, {Kind: 4, Component: child}}}
	c, err := Check(root, DefaultLimits())
	if err != nil || len(c.Children) != 1 || len(c.Children[0].CoreInstances) != 1 {
		t.Fatal(c, err)
	}
	for _, a := range []*model.Alias{{Kind: 2, Sort: model.CoreModuleSort, Count: 3}, {Kind: 2, Sort: model.CoreModuleSort, Count: 1, Index: 100}} {
		child.Definitions[0].Alias = a
		if _, err := Check(root, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	instance := &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{
		{Kind: 1, Type: &model.TypeDef{Kind: model.Function}},
		{Kind: 4, Export: external("f", model.FunctionSort, 0)},
	}}
	b := &model.Binary{Definitions: []model.Definition{
		{Kind: 7, Type: instance},
		{Kind: 10, Import: external("interface", model.InstanceSort, 0)},
		{Kind: 6, Alias: &model.Alias{Kind: 0, Sort: model.FunctionSort, Name: "f"}},
		{Kind: 8, Canonical: &model.Canonical{Op: 1}},
		{Kind: 7, Type: &model.TypeDef{Kind: model.Function}},
		{Kind: 11, Export: &model.Export{Name: model.Name{Text: "f"}, Item: model.SortIndex{Sort: model.FunctionSort}, Type: &model.ExternType{Sort: model.FunctionSort, Index: 1}}},
	}}
	if _, err := Check(b, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
	for _, a := range []*model.Alias{{Kind: 0, Sort: model.FunctionSort, Instance: 100, Name: "f"}, {Kind: 0, Sort: model.TypeSort, Name: "f"}, {Kind: 0, Sort: model.FunctionSort, Name: "missing"}} {
		b.Definitions[2].Alias = a
		if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
}
