package validate

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func constant(kind model.Kind, bytes ...byte) model.Definition {
	return model.Definition{Kind: 12, Value: &model.ValueDefinition{Type: *ref(kind), Bytes: bytes}}
}

func exportValue(name string, index uint32) model.Definition {
	return model.Definition{Kind: 11, Export: &model.Export{Name: model.Name{Text: name}, Item: model.SortIndex{Sort: model.ValueSort, Index: index}}}
}

func startComponent() *model.Binary {
	return &model.Binary{Definitions: []model.Definition{
		{Kind: 7, Type: &model.TypeDef{Kind: model.Function, Parameters: []model.NamedType{{Name: "input", Type: *ref(model.U32)}}, Result: ref(model.U32)}},
		{Kind: 10, Import: external("convert", model.FunctionSort, 0)},
		constant(model.U32, 42),
		{Kind: 9, Start: &model.Start{Arguments: []uint32{0}, Results: 1}},
		exportValue("answer", 1),
	}}
}

func TestStartAndValuePlansDoNotExecute(t *testing.T) {
	c, err := Check(startComponent(), DefaultLimits())
	if err != nil || len(c.Starts) != 1 || len(c.Values) != 1 || c.Values[0].Value.Bits != 42 || c.Type.Exports[0].Type.Value.Kind != model.U32 {
		t.Fatal(c, err)
	}
	// A missing host import would prevent execution, but not static checking.
	if c.Type.Imports[0].Name.Text != "convert" {
		t.Fatal(c.Type)
	}
	for _, mutate := range []func(*model.Binary){
		func(b *model.Binary) { b.Definitions[3].Start.Function = 1 },
		func(b *model.Binary) { b.Definitions[3].Start.Arguments = nil },
		func(b *model.Binary) { b.Definitions[3].Start.Arguments[0] = 2 },
		func(b *model.Binary) { b.Definitions[3].Start.Results = 0 },
		func(b *model.Binary) { b.Definitions[2] = constant(model.U8, 42) },
		func(b *model.Binary) { b.Definitions = b.Definitions[:4] },
		func(b *model.Binary) { b.Definitions[3].Start = nil },
		func(b *model.Binary) { b.Definitions[2].Value = nil },
		func(b *model.Binary) { b.Definitions[2].Value.Type = *idx(0) },
		func(b *model.Binary) { b.Definitions[2].Value.Bytes = []byte{128} },
	} {
		b := startComponent()
		mutate(b)
		if result, err := Check(b, DefaultLimits()); result != nil || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(result, err)
		}
	}
}

func TestTypeDirectedValueStartBinary(t *testing.T) {
	// Authored from the pinned Binary.md grammar: wasm-tools 1.258.0 cannot
	// generate value definitions. This fixture is not an independent oracle.
	input, err := os.ReadFile("testdata/value-start.wasm")
	if err != nil {
		t.Fatal(err)
	}
	b, err := decode.Decode(input, decode.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	c, err := Check(b, DefaultLimits())
	if err != nil || len(c.Values) != 1 || len(c.Starts) != 1 || c.Values[0].Value.Bits != 42 {
		t.Fatal(c, err)
	}
}

func TestValuesAreConsumedExactlyOnce(t *testing.T) {
	for _, definitions := range [][]model.Definition{
		{constant(model.U8, 42)},
		{constant(model.U8, 42), exportValue("first", 0), exportValue("second", 0)},
		{constant(model.U8, 42), exportValue("first", 0), exportValue("second", 1)},
		{{Kind: 10, Import: &model.Import{Name: model.Name{Text: "value"}, Type: model.ExternType{Sort: model.ValueSort, Bound: 1, Value: ref(model.U8)}}}},
	} {
		if result, err := Check(&model.Binary{Definitions: definitions}, DefaultLimits()); result != nil || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(result, err)
		}
	}
	// The same argument cannot satisfy two parameters even when its types match.
	b := startComponent()
	b.Definitions[0].Type.Parameters = append(b.Definitions[0].Type.Parameters, model.NamedType{Name: "other", Type: *ref(model.U32)})
	b.Definitions[3].Start.Arguments = []uint32{0, 0}
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestValuesMoveThroughInstanceAliasesAndInstantiation(t *testing.T) {
	bag := &model.InstanceExpr{Exports: []model.Export{{Name: model.Name{Text: "value"}, Item: model.SortIndex{Sort: model.ValueSort}}}}
	b := &model.Binary{Definitions: []model.Definition{
		constant(model.U8, 42),
		{Kind: 5, Instance: bag},
		{Kind: 6, Alias: &model.Alias{Sort: model.ValueSort, Name: "value"}},
		exportValue("answer", 1),
	}}
	if _, err := Check(b, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
	b.Definitions = append(b.Definitions, b.Definitions[2])
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("instance value aliased twice", err)
	}
	child := &model.Binary{Definitions: []model.Definition{
		{Kind: 10, Import: &model.Import{Name: model.Name{Text: "value"}, Type: model.ExternType{Sort: model.ValueSort, Bound: 1, Value: ref(model.U8)}}},
		exportValue("value", 0),
	}}
	b = &model.Binary{Definitions: []model.Definition{
		{Kind: 4, Component: child},
		constant(model.U8, 42),
		{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true, Arguments: []model.Argument{{Name: "value", Item: model.SortIndex{Sort: model.ValueSort}}}}},
		{Kind: 6, Alias: &model.Alias{Sort: model.ValueSort, Name: "value"}},
		exportValue("answer", 1),
	}}
	if _, err := Check(b, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
	b.Definitions = append(b.Definitions, exportValue("duplicate", 0))
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("consumed component argument reused", err)
	}
}

func TestNestedInstanceValuesAndPublishedAliases(t *testing.T) {
	inner := &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 4, Export: &model.Import{Name: model.Name{Text: "value"}, Type: model.ExternType{Sort: model.ValueSort, Bound: 1, Value: ref(model.U8)}}}}}
	outer := &model.TypeDef{Kind: model.Instance, Declarations: []model.TypeDeclaration{{Kind: 1, Type: inner}, {Kind: 4, Export: external("inner", model.InstanceSort, 0)}}}
	b := &model.Binary{Definitions: []model.Definition{
		{Kind: 7, Type: outer},
		{Kind: 10, Import: external("input", model.InstanceSort, 0)},
		{Kind: 6, Alias: &model.Alias{Sort: model.InstanceSort, Name: "inner"}},
		{Kind: 6, Alias: &model.Alias{Sort: model.ValueSort, Instance: 1, Name: "value"}},
		exportValue("answer", 0),
	}}
	if _, err := Check(b, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
	b.Definitions = append(b.Definitions, model.Definition{Kind: 11, Export: &model.Export{Name: model.Name{Text: "duplicate"}, Item: model.SortIndex{Sort: model.InstanceSort}}})
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("nested instance value reused", err)
	}
	b.Definitions = b.Definitions[:2]
	b.Definitions = append(b.Definitions, model.Definition{Kind: 11, Export: &model.Export{Name: model.Name{Text: "output"}, Item: model.SortIndex{Sort: model.InstanceSort}}})
	if _, err := Check(b, DefaultLimits()); err != nil {
		t.Fatal(err)
	}
	// Export adds an instance index but must not reset its consumed values.
	b.Definitions = append(b.Definitions, model.Definition{Kind: 6, Alias: &model.Alias{Sort: model.InstanceSort, Instance: 1, Name: "inner"}}, model.Definition{Kind: 6, Alias: &model.Alias{Sort: model.ValueSort, Instance: 2, Name: "value"}})
	if _, err := Check(b, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("published alias reset consumption", err)
	}
}

func TestValuesAndStartShareCompilationBudget(t *testing.T) {
	failed, passed := 0, 0
	for budget := uint32(1); budget < 200; budget++ {
		result, err := Check(startComponent(), Limits{MaxNodes: budget, MaxDepth: 128})
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

func TestValueEqualityBoundsAndAbstractExports(t *testing.T) {
	for _, second := range []byte{42, 43} {
		child := &model.Binary{Definitions: []model.Definition{
			{Kind: 10, Import: &model.Import{Name: model.Name{Text: "first"}, Type: model.ExternType{Sort: model.ValueSort, Bound: 1, Value: ref(model.U8)}}},
			{Kind: 10, Import: &model.Import{Name: model.Name{Text: "same"}, Type: model.ExternType{Sort: model.ValueSort}}},
			exportValue("first", 0), exportValue("same", 1),
		}}
		b := &model.Binary{Definitions: []model.Definition{
			{Kind: 4, Component: child},
			constant(model.U8, 42), constant(model.U8, second),
			{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true, Arguments: []model.Argument{
				{Name: "first", Item: model.SortIndex{Sort: model.ValueSort}},
				{Name: "same", Item: model.SortIndex{Sort: model.ValueSort, Index: 1}},
			}}},
			{Kind: 11, Export: &model.Export{Name: model.Name{Text: "output"}, Item: model.SortIndex{Sort: model.InstanceSort}}},
		}}
		_, err := Check(b, DefaultLimits())
		if second == 42 && err != nil || second != 42 && !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(second, err)
		}
	}
	// Type ascription deliberately hides the precise constant. Instantiating
	// this component twice must produce independent abstract value binders.
	child := &model.Binary{Definitions: []model.Definition{
		constant(model.U8, 42),
		exportValue("value", 0),
	}}
	child.Definitions[1].Export.Type = &model.ExternType{Sort: model.ValueSort, Bound: 1, Value: ref(model.U8)}
	b := &model.Binary{Definitions: []model.Definition{
		{Kind: 4, Component: child},
		{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true}},
		{Kind: 5, Instance: &model.InstanceExpr{Instantiate: true}},
		{Kind: 11, Export: &model.Export{Name: model.Name{Text: "first"}, Item: model.SortIndex{Sort: model.InstanceSort}}},
		{Kind: 11, Export: &model.Export{Name: model.Name{Text: "second"}, Item: model.SortIndex{Sort: model.InstanceSort, Index: 1}}},
	}}
	c, err := Check(b, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	a, z := c.Instances[0].Type.Exports[0].Type.valueBinding, c.Instances[1].Type.Exports[0].Type.valueBinding
	if a == z || a.literal != nil || z.literal != nil {
		t.Fatal("abstract value ascription leaked a constant or shared a binder")
	}
	child.Definitions[1].Export.Type = nil
	c, err = Check(b, DefaultLimits())
	if err != nil || c.Instances[0].Type.Exports[0].Type.valueBinding.literal.Bits != 42 {
		t.Fatal("implicit export lost its precise constant", c, err)
	}
}
