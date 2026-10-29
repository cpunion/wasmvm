package validate

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	corevalidate "github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// CoreExtern uses canonical Core type identities in its function/reference
// descriptors. It contains no runtime addresses or mutable instance state.
type CoreExtern struct {
	Module, Name string
	Descriptor   module.Import
}

type CoreModuleType struct {
	Imports, Exports []CoreExtern
}

type coreScope struct {
	pool    *TypePool
	parent  *coreScope
	types   *module.Module
	ids     []uint32
	aliases map[uint32]bool
	modules map[uint32]*CoreModuleType
}

func (c *coreScope) fragment() corevalidate.Fragment {
	return corevalidate.Fragment{Types: c.types, IDs: c.ids, Registry: &c.pool.registry, Features: wasm.SupportedFeatures, MaxTypes: c.pool.limits.MaxNodes}
}

// DefineCore binds the next lexical Core type expression, sharing the Core
// validator and canonical registry rather than revalidating every prefix.
func (s *TypeScope) DefineCore(def *model.CoreTypeDef, source *module.Module) error {
	if def == nil || def.Start != uint32(len(s.core.ids)) {
		return invalid("Core type definition order")
	}
	if !def.Module {
		return s.core.group(source, def.Start, def.Count)
	}
	if def.Count != 1 {
		return invalid("Core module type count")
	}
	if err := s.pool.charge(1 + uint64(len(def.Declarations))); err != nil {
		return err
	}
	body := &coreScope{pool: s.pool, parent: s.core, types: new(module.Module), modules: map[uint32]*CoreModuleType{}}
	m := &CoreModuleType{}
	imports, exports := map[[2]string]bool{}, map[string]bool{}
	for _, d := range def.Declarations {
		switch d.Kind {
		case 0, 3:
			desc, err := body.external(d.Type)
			if err != nil {
				return err
			}
			e := CoreExtern{Module: d.Module, Name: d.Name, Descriptor: desc}
			if d.Kind == 0 {
				key := [2]string{d.Module, d.Name}
				if imports[key] {
					return invalid("duplicate Core module type import")
				}
				imports[key] = true
				m.Imports = append(m.Imports, e)
			} else {
				if exports[d.Name] {
					return invalid("duplicate Core module type export")
				}
				exports[d.Name] = true
				m.Exports = append(m.Exports, e)
			}
		case 1:
			if err := body.group(def.Types, d.Start, d.Count); err != nil {
				return err
			}
		case 2:
			if d.Alias == nil || d.Alias.Kind != 2 || d.Alias.Sort != model.CoreTypeSort {
				return invalid("Core module type alias")
			}
			if err := body.alias(d.Alias.Count, d.Alias.Index, false); err != nil {
				return err
			}
		default:
			return invalid("unknown Core module type declaration")
		}
	}
	s.core.modules[def.Start] = m
	s.core.ids = append(s.core.ids, 0)
	s.core.types.Types = append(s.core.types.Types, module.FuncType{})
	s.core.types.TypeDefs = append(s.core.types.TypeDefs, module.TypeDef{})
	return nil
}

func (c *coreScope) group(source *module.Module, start, count uint32) error {
	end := uint64(start) + uint64(count)
	if source == nil || start != uint32(len(c.ids)) || end > uint64(len(source.Types)) || len(source.TypeDefs) != 0 && len(source.TypeDefs) != len(source.Types) {
		return invalid("Core lexical type group bounds")
	}
	metadata := 1 + uint64(count)
	for i := start; uint64(i) < end; i++ {
		metadata += uint64(len(source.Types[i].Params)) + uint64(len(source.Types[i].Results)) + uint64(len(source.TypeDef(i).Fields)) + uint64(len(source.TypeDef(i).Supers))
	}
	if err := c.pool.charge(metadata); err != nil {
		return err
	}
	for i := start; uint64(i) < end; i++ {
		f, d := source.Types[i], source.TypeDef(i)
		f.Params, f.Results = c.values(f.Params), c.values(f.Results)
		d.Supers = append([]uint32(nil), d.Supers...)
		d.Fields = append([]module.FieldType(nil), d.Fields...)
		for j := range d.Fields {
			d.Fields[j].Value = c.value(d.Fields[j].Value)
		}
		c.types.Types = append(c.types.Types, f)
		c.types.TypeDefs = append(c.types.TypeDefs, d)
		c.ids = append(c.ids, 0)
	}
	return c.fragment().TypeGroup(start, count)
}

func (c *coreScope) values(in []wasm.ValueType) []wasm.ValueType {
	out := make([]wasm.ValueType, len(in))
	for i, v := range in {
		out[i] = c.value(v)
	}
	return out
}

func (c *coreScope) value(v wasm.ValueType) wasm.ValueType {
	// Binary decoding cannot know the reference family of an unresolved outer
	// alias. Repair only these aliases after binding; ordinary decoded family
	// mismatches remain errors in the shared Core validator.
	if i, ok := v.TypeIndex(); ok && !v.Canonical() && c.aliases[i] {
		return c.types.Reference(i, v.Nullable())
	}
	return v
}

func (c *coreScope) alias(count, index uint32, moduleAllowed bool) error {
	at := c
	for i := uint32(0); i < count; i++ {
		at = at.parent
		if at == nil {
			return invalid("Core outer scope out of bounds")
		}
	}
	if uint64(index) >= uint64(len(at.ids)) {
		return invalid("Core outer type index out of bounds")
	}
	if m := at.modules[index]; m != nil {
		if !moduleAllowed {
			return invalid("Core module type aliases a module type")
		}
		if err := c.pool.charge(1); err != nil {
			return err
		}
		c.modules[uint32(len(c.ids))] = m
		c.ids = append(c.ids, 0)
		c.types.Types = append(c.types.Types, module.FuncType{})
		c.types.TypeDefs = append(c.types.TypeDefs, module.TypeDef{})
		return nil
	}
	id := at.ids[index]
	if id == 0 || uint64(id) >= uint64(len(c.pool.registry.Types)) {
		return invalid("unbound Core alias type")
	}
	if err := c.pool.charge(1); err != nil {
		return err
	}
	t := c.pool.registry.Types[id]
	if c.aliases == nil {
		c.aliases = map[uint32]bool{}
	}
	c.aliases[uint32(len(c.ids))] = true
	c.ids = append(c.ids, id)
	c.types.Types = append(c.types.Types, t.Function)
	c.types.TypeDefs = append(c.types.TypeDefs, t.Definition)
	return nil
}

func (c *coreScope) external(in module.Import) (module.Import, error) {
	in.Table.Element = c.value(in.Table.Element)
	in.Global.Value = c.value(in.Global.Value)
	if err := c.fragment().Extern(in); err != nil {
		return module.Import{}, err
	}
	if in.Kind == 0 || in.Kind == 4 {
		in.Type = c.ids[in.Type]
	}
	in.Table.Element = module.ResolveType(in.Table.Element, c.ids)
	in.Global.Value = module.ResolveType(in.Global.Value, c.ids)
	return in, nil
}
