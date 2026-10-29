package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	corevalidate "github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func (p *TypePool) coreModule(m *module.Module) (*CoreModule, error) {
	if m == nil {
		return nil, invalid("missing embedded Core module")
	}
	nodes := uint64(len(m.Types)) + uint64(len(m.Imports)) + uint64(len(m.Exports)) + uint64(len(m.Functions)) + uint64(len(m.Tables)) + uint64(len(m.Memories)) + uint64(len(m.Globals)) + uint64(len(m.Tags)) + uint64(len(m.Elements)) + uint64(len(m.Data))
	for i, t := range m.Types {
		nodes += uint64(len(t.Params)) + uint64(len(t.Results))
		if len(m.TypeDefs) == len(m.Types) {
			nodes += uint64(len(m.TypeDefs[i].Fields)) + uint64(len(m.TypeDefs[i].Supers))
		}
	}
	for _, f := range m.Functions {
		nodes += uint64(f.LocalCount) + uint64(len(f.Locals)) + uint64(len(f.Body))
	}
	for _, g := range m.Globals {
		nodes += uint64(len(g.Init))
	}
	for _, e := range m.Elements {
		nodes += uint64(len(e.Offset)) + uint64(len(e.Functions))
		for _, x := range e.Expressions {
			nodes += uint64(len(x))
		}
	}
	for _, d := range m.Data {
		nodes += uint64(len(d.Offset))
	}
	if err := p.charge(nodes); err != nil {
		return nil, err
	}
	seen := map[[2]string]bool{}
	for _, im := range m.Imports {
		key := [2]string{im.Module, im.Name}
		if seen[key] {
			return nil, invalid("duplicate embedded Core import name")
		}
		seen[key] = true
	}
	checked, err := corevalidate.CheckWithFeatures(m, corevalidate.DefaultLimits(), wasm.SupportedFeatures)
	if err != nil {
		return nil, err
	}
	ids, ok := p.registry.InternModule(m, p.limits.MaxNodes)
	if !ok {
		return nil, limit("embedded canonical Core types")
	}
	typ := &CoreModuleType{}
	spaces := [5][]module.Import{}
	normalize := func(t module.Import) module.Import {
		if t.Kind == 0 || t.Kind == 4 {
			t.Type = ids[t.Type]
		}
		t.Table.Element = module.ResolveType(t.Table.Element, ids)
		t.Global.Value = module.ResolveType(t.Global.Value, ids)
		return t
	}
	for _, im := range m.Imports {
		desc := normalize(im)
		typ.Imports = append(typ.Imports, CoreExtern{Module: im.Module, Name: im.Name, Descriptor: desc})
		spaces[im.Kind] = append(spaces[im.Kind], desc)
	}
	for _, f := range m.Functions {
		spaces[0] = append(spaces[0], normalize(module.Import{Kind: 0, Type: f.Type}))
	}
	for _, t := range m.Tables {
		spaces[1] = append(spaces[1], normalize(module.Import{Kind: 1, Table: t}))
	}
	for _, t := range m.Memories {
		spaces[2] = append(spaces[2], module.Import{Kind: 2, Memory: t})
	}
	for _, g := range m.Globals {
		spaces[3] = append(spaces[3], normalize(module.Import{Kind: 3, Global: g.Type}))
	}
	for _, t := range m.Tags {
		spaces[4] = append(spaces[4], normalize(module.Import{Kind: 4, Type: t}))
	}
	for _, e := range m.Exports {
		typ.Exports = append(typ.Exports, CoreExtern{Name: e.Name, Descriptor: spaces[e.Kind][e.Index]})
	}
	return &CoreModule{Definition: checked, Type: typ}, nil
}
