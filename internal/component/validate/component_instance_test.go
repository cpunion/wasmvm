package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func resourceComponent() *model.Binary {
	return &model.Binary{Definitions: []model.Definition{
		{Kind: 7, Type: &model.TypeDef{Kind: model.Resource, Representation: wasm.I32}},
		{Kind: 11, Export: &model.Export{Name: model.Name{Text: "file"}, Item: model.SortIndex{Sort: model.TypeSort}}},
		{Kind: 11, Export: &model.Export{Name: model.Name{Text: "same-file"}, Item: model.SortIndex{Sort: model.TypeSort, Index: 1}}},
	}}
}

func TestComponentInstantiationGeneratesResourceIdentities(t *testing.T) {
	b := &model.Binary{Definitions: []model.Definition{
		{Kind: 4, Component: resourceComponent()},
		{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true}},
		{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true}},
	}}
	c, err := Check(b, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	first, second := c.Instances[0].Type.Exports, c.Instances[1].Type.Exports
	if first[0].Type.Resource == second[0].Type.Resource || first[0].Type.Resource == c.Children[0].Type.Exports[0].Type.Resource {
		t.Fatal("component instantiations shared a generated resource")
	}
	if first[0].Type.Resource != first[1].Type.Resource || second[0].Type.Resource != second[1].Type.Resource {
		t.Fatal("instantiation broke equality aliases")
	}
	// A child-generated resource is opaque to its parent.
	b.Definitions = append(b.Definitions,
		model.Definition{Kind: 6, Alias: &model.Alias{Kind: 0, Sort: model.TypeSort, Name: "file"}},
		model.Definition{Kind: 8, Canonical: &model.Canonical{Op: 2}},
	)
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("parent acquired child resource representation", err)
	}
}

func TestImportedResourceRoundTripRetainsAuthority(t *testing.T) {
	child := &model.Binary{Definitions: []model.Definition{
		{Kind: 10, Import: &model.Import{Name: model.Name{Text: "file"}, Type: model.ExternType{Sort: model.TypeSort, Bound: 1}}},
		{Kind: 11, Export: &model.Export{Name: model.Name{Text: "file"}, Item: model.SortIndex{Sort: model.TypeSort}}},
	}}
	b := &model.Binary{Definitions: []model.Definition{
		{Kind: 7, Type: &model.TypeDef{Kind: model.Resource, Representation: wasm.I64}},
		{Kind: 4, Component: child},
		{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true, Arguments: []model.Argument{{Name: "file", Item: model.SortIndex{Sort: model.TypeSort}}}}},
		{Kind: 6, Alias: &model.Alias{Kind: 0, Sort: model.TypeSort, Name: "file"}},
		{Kind: 8, Canonical: &model.Canonical{Op: 2, Index: 1}},
	}}
	c, err := Check(b, DefaultLimits())
	if err != nil || c.Canonical[0].Signature.Params[0] != wasm.I64 {
		t.Fatal(c, err)
	}
	child.Definitions = append(child.Definitions, model.Definition{Kind: 8, Canonical: &model.Canonical{Op: 2}})
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("child acquired imported resource representation", err)
	}
}

func TestComponentInstantiationRejectsInvalidArguments(t *testing.T) {
	for _, expr := range []*model.InstanceExpr{
		nil,
		{Instantiate: true, Index: 1},
		{Instantiate: true, Arguments: []model.Argument{{Name: "extra", Item: model.SortIndex{Sort: model.TypeSort, Index: 100}}}},
		{Instantiate: true, Arguments: []model.Argument{{Name: "extra", Item: model.SortIndex{Sort: model.ValueSort}}}},
		{Instantiate: true, Arguments: []model.Argument{{Name: "extra", Item: model.SortIndex{Sort: model.ComponentSort}}, {Name: "extra", Item: model.SortIndex{Sort: model.ComponentSort}}}},
	} {
		b := &model.Binary{Definitions: []model.Definition{{Kind: 4, Component: resourceComponent()}, {Kind: 5, Instance: expr}}}
		if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(expr, err)
		}
	}
}

func TestComponentInstantiationWorkBudget(t *testing.T) {
	failed, passed := 0, 0
	for budget := uint32(1); budget <= 500; budget++ {
		b := &model.Binary{Definitions: []model.Definition{
			{Kind: 4, Component: resourceComponent()},
			{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true}},
		}}
		result, err := Check(b, Limits{MaxNodes: budget, MaxDepth: 128})
		if err != nil {
			if !errors.Is(err, wasm.ErrLimit) || result != nil {
				t.Fatal(budget, result, err)
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

func TestResourceSubstitutionChainsAreBounded(t *testing.T) {
	s := scope(t)
	a, _ := s.pool.abstract()
	b, _ := s.pool.abstract()
	c, _ := s.pool.abstract()
	r := s.pool.relation()
	for _, pair := range [][2]*Type{{b, a}, {c, b}} {
		if ok, err := r.bind(pair[0], pair[1]); !ok || err != nil {
			t.Fatal(ok, err)
		}
	}
	rewrite := r.rewrite()
	if id, err := rewrite.id(a.Resource); err != nil || id != c.Resource {
		t.Fatal(id, err)
	}
	s.pool.limits.MaxNodes = s.pool.nodes
	if _, err := rewrite.id(a.Resource); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("substitution chain escaped work quota", err)
	}
}

func TestAsyncBuiltinStaticPlans(t *testing.T) {
	for _, op := range []byte{0x0e, 0x15} {
		kind := model.Stream
		if op == 0x15 {
			kind = model.Future
		}
		b := &model.Binary{Definitions: []model.Definition{
			{Kind: 7, Type: &model.TypeDef{Kind: kind}},
			{Kind: 8, Canonical: &model.Canonical{Op: op}},
		}}
		c, err := Check(b, DefaultLimits())
		if err != nil || !c.Canonical[0].Async || c.Canonical[0].Signature.Results[0] != wasm.I64 {
			t.Fatal(c, err)
		}
		b.Definitions[0].Type.Kind = model.List
		b.Definitions[0].Type.Element = ref(model.U8)
		if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("wrong async constructor type accepted", err)
		}
		b.Definitions = b.Definitions[1:]
		if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("out-of-bounds constructor type accepted", err)
		}
	}
	for _, op := range []byte{0x20, 0x21} {
		m := &module.Module{Memories: []module.Limits{{Index64: true}}, Exports: []module.Export{{Name: "memory", Kind: 2}}}
		b := &model.Binary{Definitions: []model.Definition{
			{Kind: 1, CoreModule: m},
			{Kind: 2, Instance: &model.InstanceExpr{Instantiate: true}},
			{Kind: 6, Alias: &model.Alias{Kind: 1, Sort: model.CoreMemorySort, Name: "memory"}},
			{Kind: 8, Canonical: &model.Canonical{Op: op}},
		}}
		c, err := Check(b, DefaultLimits())
		if err != nil || !c.Canonical[0].Async || !c.Canonical[0].Memory64 || c.Canonical[0].Signature.Params[1] != wasm.I64 {
			t.Fatal(c, err)
		}
		m.Memories[0] = module.Limits{Shared: true, HasMax: true, Max: 1}
		if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("shared async canonical memory accepted", err)
		}
		b.Definitions[3].Canonical.Index = 1
		if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("out-of-bounds async memory accepted", err)
		}
	}
}
