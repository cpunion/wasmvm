package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFragmentGroupsAliasesAndRollback(t *testing.T) {
	m := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{wasm.I32}}}, TypeDefs: []module.TypeDef{
		{Kind: module.StructKind, Open: true, Fields: []module.FieldType{{Value: wasm.I32}}}, {Kind: module.FunctionKind},
	}}
	registry := new(module.TypeRegistry)
	f := Fragment{Types: m, IDs: make([]uint32, 2), Registry: registry, Features: wasm.SupportedFeatures, MaxTypes: 100}
	if err := f.TypeGroup(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := f.TypeGroup(1, 1); err != nil {
		t.Fatal(err)
	}
	if f.IDs[0] == 0 || f.IDs[1] == 0 {
		t.Fatal(f.IDs)
	}
	if err := f.TypeGroup(1, 1); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	alias := registry.Types[f.IDs[0]]
	child := &module.Module{Types: []module.FuncType{alias.Function, {}}, TypeDefs: []module.TypeDef{alias.Definition,
		{Kind: module.StructKind, Supers: []uint32{0}, Fields: []module.FieldType{{Value: wasm.I32}}},
	}}
	g := Fragment{Types: child, IDs: []uint32{f.IDs[0], 0}, Registry: registry, Features: f.Features, MaxTypes: 100}
	if err := g.TypeGroup(1, 1); err != nil {
		t.Fatal(err)
	}
	mark := registry.Count()
	child.TypeDefs[1].Fields[0].Value = wasm.F64
	g.IDs[1] = 0
	if err := g.TypeGroup(1, 1); !errors.Is(err, wasm.ErrInvalid) || g.IDs[1] != 0 || registry.Count() != mark {
		t.Fatal(err, g.IDs, registry.Count(), mark)
	}
	child.TypeDefs[1].Fields[0].Value = wasm.I32
	if err := g.TypeGroup(1, 1); err != nil {
		t.Fatal("rollback corrupted registry", err)
	}
}

func TestFragmentExternalDescriptors(t *testing.T) {
	m := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}}}}
	f := Fragment{Types: m, IDs: make([]uint32, 1), Registry: new(module.TypeRegistry), Features: wasm.SupportedFeatures, MaxTypes: 10}
	if err := f.TypeGroup(0, 1); err != nil {
		t.Fatal(err)
	}
	for _, d := range []module.Import{
		{Kind: 0}, {Kind: 4}, {Kind: 1, Table: module.TableType{Element: wasm.FuncRef}},
		{Kind: 2, Memory: module.Limits{Index64: true, Min: 1 << 20}},
		{Kind: 2, Memory: module.Limits{Shared: true, HasMax: true, Max: 1}},
		{Kind: 3, Global: module.GlobalType{Value: wasm.FunctionReference(0, true)}},
	} {
		if err := f.Extern(d); err != nil {
			t.Fatal(d, err)
		}
	}
	for _, d := range []module.Import{
		{Kind: 5}, {Kind: 0, Type: 1}, {Kind: 2, Memory: module.Limits{Min: 65537}},
		{Kind: 2, Memory: module.Limits{Shared: true}},
		{Kind: 1, Table: module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Shared: true}}},
		{Kind: 1, Table: module.TableType{Element: wasm.I32}},
		{Kind: 3, Global: module.GlobalType{Value: wasm.FunctionReference(1, true)}},
	} {
		if err := f.Extern(d); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(d, err)
		}
	}
	f.IDs[0] = 0
	if err := f.Extern(module.Import{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	f.Features = ^wasm.SupportedFeatures
	if err := f.Extern(module.Import{}); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	if err := f.TypeGroup(0, 1); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestFragmentInvalidInventoriesAndLimits(t *testing.T) {
	for _, f := range []Fragment{
		{}, {Registry: new(module.TypeRegistry)},
		{Registry: new(module.TypeRegistry), Types: &module.Module{Types: []module.FuncType{{}}}},
		{Registry: new(module.TypeRegistry), Types: &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.FunctionKind}, {}}}, IDs: make([]uint32, 1)},
		{Registry: new(module.TypeRegistry), Types: &module.Module{Types: []module.FuncType{{}}, TypeDefs: []module.TypeDef{{Kind: module.FunctionKind}}}, IDs: make([]uint32, 1)},
	} {
		if err := f.TypeGroup(0, 1); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(f, err)
		}
	}
	if err := (Fragment{}).Extern(module.Import{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	f := Fragment{Registry: new(module.TypeRegistry), Types: &module.Module{Types: []module.FuncType{{}}}, IDs: make([]uint32, 1)}
	if err := f.TypeGroup(0, 0); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	if err := f.TypeGroup(0, 1); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	for _, parent := range []module.TypeDef{{}, {Kind: module.StructKind, Open: true}} {
		m := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{wasm.IndexedReference(wasm.StructRef, 0, true)}}}, TypeDefs: []module.TypeDef{parent, {Kind: module.FunctionKind}}}
		g := Fragment{Registry: new(module.TypeRegistry), Types: m, IDs: make([]uint32, 2), Features: wasm.SupportedFeatures, MaxTypes: 10}
		if err := g.TypeGroup(1, 1); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(parent, err)
		}
		m.Types[1] = module.FuncType{}
		m.TypeDefs[1] = module.TypeDef{Kind: module.StructKind, Supers: []uint32{0}}
		if err := g.TypeGroup(1, 1); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(parent, err)
		}
	}
}
