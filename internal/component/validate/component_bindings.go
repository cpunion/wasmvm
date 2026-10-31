package validate

import (
	"sort"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// ComponentBindings matches independent immutable compilation schemas in one
// bounded registry. Existing subtype, binder and Core relations remain shared.
type ComponentBindings struct {
	owner   *TypePool
	pool    *TypePool
	sources map[*TypePool]*componentSource
	types   map[*Type]*Type
	values  map[*model.Type]*model.Type
	symbols map[*valueBinding]*valueBinding
	checked map[typeComparison]bool
}

type componentSource struct {
	core      []uint32
	resources map[uint64]uint64
}

func (t *Type) ComponentBindings() (*ComponentBindings, error) {
	if t == nil || t.owner == nil || t.Kind != model.Component {
		return nil, invalid("missing component import type")
	}
	return &ComponentBindings{owner: t.owner, pool: NewTypePool(t.owner.limits), sources: map[*TypePool]*componentSource{}, types: map[*Type]*Type{}, values: map[*model.Type]*model.Type{}, symbols: map[*valueBinding]*valueBinding{}, checked: map[typeComparison]bool{}}, nil
}

func (b *ComponentBindings) Check(expected, actual *Type) (bool, error) {
	if b == nil || expected == nil || actual == nil || expected.owner != b.owner || actual.owner == nil || expected.Kind != model.Component || actual.Kind != model.Component {
		return false, invalid("foreign component binding")
	}
	key := typeComparison{a: actual, b: expected}
	if b.checked[key] {
		return true, nil
	}
	e, err := b.typ(expected)
	if err != nil {
		return false, err
	}
	a, err := b.typ(actual)
	if err != nil {
		return false, err
	}
	ok, err := b.pool.relation().typ(a, e)
	if ok && err == nil {
		b.checked[key] = true
	}
	return ok, err
}

func (b *ComponentBindings) source(p *TypePool) (*componentSource, error) {
	if s := b.sources[p]; s != nil {
		return s, nil
	}
	if err := b.pool.charge(1); err != nil {
		return nil, err
	}
	s := &componentSource{resources: map[uint64]uint64{}}
	b.sources[p] = s
	return s, nil
}

func (b *ComponentBindings) resource(p *TypePool, id uint64) (uint64, error) {
	if id == 0 {
		return 0, nil
	}
	s, err := b.source(p)
	if err != nil {
		return 0, err
	}
	if got := s.resources[id]; got != 0 {
		return got, nil
	}
	t, err := b.pool.abstract()
	if err != nil {
		return 0, err
	}
	s.resources[id] = t.Resource
	return t.Resource, nil
}

func (b *ComponentBindings) value(p *TypePool, in *model.Type) (*model.Type, error) {
	if in == nil {
		return nil, nil
	}
	if out := b.values[in]; out != nil {
		return out, nil
	}
	if err := b.pool.charge(1 + uint64(len(in.Fields)+len(in.Cases)+len(in.Labels))); err != nil {
		return nil, err
	}
	out := *in
	b.values[in] = &out
	var err error
	out.ResourceID, err = b.resource(p, in.ResourceID)
	if err != nil {
		return nil, err
	}
	out.Fields = append([]model.Field(nil), in.Fields...)
	out.Cases = append([]model.Field(nil), in.Cases...)
	for _, fields := range [][]model.Field{out.Fields, out.Cases} {
		for i := range fields {
			fields[i].Type, err = b.value(p, fields[i].Type)
			if err != nil {
				return nil, err
			}
		}
	}
	out.Key, err = b.value(p, in.Key)
	if err != nil {
		return nil, err
	}
	out.Element, err = b.value(p, in.Element)
	if err != nil {
		return nil, err
	}
	info := p.values[in]
	if info.resourceName != nil {
		info.resourceName, err = b.typ(info.resourceName)
		if err != nil {
			return nil, err
		}
	}
	info.minimumResource = out.ResourceID
	for _, fields := range [][]model.Field{out.Fields, out.Cases} {
		for _, f := range fields {
			if f.Type != nil {
				info.minimumResource = minimum(info.minimumResource, b.pool.values[f.Type].minimumResource)
			}
		}
	}
	for _, v := range []*model.Type{out.Key, out.Element} {
		if v != nil {
			info.minimumResource = minimum(info.minimumResource, b.pool.values[v].minimumResource)
		}
	}
	b.pool.values[&out] = info
	return &out, nil
}

func (b *ComponentBindings) typ(in *Type) (*Type, error) {
	if in == nil {
		return nil, nil
	}
	if out := b.types[in]; out != nil {
		return out, nil
	}
	if err := b.pool.charge(1 + uint64(len(in.Parameters)+len(in.Imports)+len(in.Exports)+len(in.Captures))); err != nil {
		return nil, err
	}
	out := *in
	out.owner = b.pool
	out.definingScope = nil
	b.types[in] = &out
	var err error
	out.Resource, err = b.resource(in.owner, in.Resource)
	if err != nil {
		return nil, err
	}
	out.minimumResource, err = b.resource(in.owner, in.minimumResource)
	if err != nil {
		return nil, err
	}
	out.Value, err = b.value(in.owner, in.Value)
	if err != nil {
		return nil, err
	}
	out.Result, err = b.value(in.owner, in.Result)
	if err != nil {
		return nil, err
	}
	out.Parameters = append([]model.Field(nil), in.Parameters...)
	for i := range out.Parameters {
		out.Parameters[i].Type, err = b.value(in.owner, in.Parameters[i].Type)
		if err != nil {
			return nil, err
		}
	}
	for _, pair := range []struct {
		in  []Extern
		out *[]Extern
	}{{in.Imports, &out.Imports}, {in.Exports, &out.Exports}} {
		*pair.out = append([]Extern(nil), pair.in...)
		for i := range *pair.out {
			(*pair.out)[i].Type, err = b.typ(pair.in[i].Type)
			if err != nil {
				return nil, err
			}
		}
	}
	out.Captures = append([]uint64(nil), in.Captures...)
	for i, id := range out.Captures {
		out.Captures[i], err = b.resource(in.owner, id)
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out.Captures, func(i, j int) bool { return out.Captures[i] < out.Captures[j] })
	out.minimumResource = 0
	if out.Kind == model.Resource {
		out.minimumResource = out.Resource
	} else if out.Kind == model.Component || out.Kind == model.Instance {
		if len(out.Captures) != 0 {
			out.minimumResource = out.Captures[0]
		}
	} else {
		for _, v := range []*model.Type{out.Value, out.Result} {
			if v != nil {
				out.minimumResource = minimum(out.minimumResource, b.pool.values[v].minimumResource)
			}
		}
		for _, f := range out.Parameters {
			out.minimumResource = minimum(out.minimumResource, b.pool.values[f.Type].minimumResource)
		}
	}
	if in.valueBinding != nil {
		symbol := b.symbols[in.valueBinding]
		if symbol == nil {
			if err := b.pool.charge(1); err != nil {
				return nil, err
			}
			symbol = &valueBinding{literal: in.valueBinding.literal}
			b.symbols[in.valueBinding] = symbol
		}
		out.valueBinding = symbol
	}
	if in.CoreModule != nil {
		out.CoreModule, err = b.core(in.owner, in.CoreModule)
		if err != nil {
			return nil, err
		}
	}
	return &out, nil
}

func (b *ComponentBindings) core(p *TypePool, in *CoreModuleType) (*CoreModuleType, error) {
	s, err := b.source(p)
	if err != nil {
		return nil, err
	}
	if s.core == nil {
		nodes := uint64(p.registry.Count())
		for i := 1; i < len(p.registry.Types); i++ {
			t := p.registry.Types[i]
			nodes += uint64(len(t.Function.Params) + len(t.Function.Results) + len(t.Definition.Fields) + len(t.Definition.Supers))
		}
		if err := b.pool.charge(nodes); err != nil {
			return nil, err
		}
		ids, ok := b.pool.registry.Import(&p.registry, b.pool.limits.MaxNodes)
		if !ok {
			return nil, limit("component binding Core registry")
		}
		s.core = ids
	}

	if err := b.pool.charge(1 + uint64(len(in.Imports)+len(in.Exports))); err != nil {
		return nil, err
	}
	remap := func(t wasm.ValueType) wasm.ValueType {
		if id, ok := t.TypeIndex(); ok {
			return t.WithTypeIndex(s.core[id], true)
		}
		return t
	}
	out := &CoreModuleType{Imports: append([]CoreExtern(nil), in.Imports...), Exports: append([]CoreExtern(nil), in.Exports...)}
	for _, members := range [][]CoreExtern{out.Imports, out.Exports} {
		for i := range members {
			d := &members[i].Descriptor
			if d.Kind == 0 || d.Kind == 4 {
				d.Type = s.core[d.Type]
			}
			d.Table.Element = remap(d.Table.Element)
			d.Global.Value = remap(d.Global.Value)
		}
	}
	return out, nil
}
