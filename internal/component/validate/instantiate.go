package validate

import (
	"sort"

	"github.com/cpunion/wasmvm/internal/component/model"
)

// captures computes free resource identities, excluding resources introduced
// since this interface declaration began. Cached child-interface captures
// prevent an inner interface's private binders from escaping into the result.
func (p *TypePool) captures(root *Type, cutoff uint64) ([]uint64, error) {
	resources := map[uint64]bool{}
	types, values := map[*Type]bool{}, map[*model.Type]bool{}
	add := func(id uint64) {
		if id != 0 && id < cutoff {
			resources[id] = true
		}
	}
	var value func(*model.Type) error
	value = func(v *model.Type) error {
		if v == nil || values[v] {
			return nil
		}
		if err := p.charge(1 + uint64(len(v.Fields)) + uint64(len(v.Cases))); err != nil {
			return err
		}
		values[v] = true
		add(v.ResourceID)
		for _, fs := range [][]model.Field{v.Fields, v.Cases} {
			for _, f := range fs {
				if err := value(f.Type); err != nil {
					return err
				}
			}
		}
		if err := value(v.Element); err != nil {
			return err
		}
		return value(v.Key)
	}
	var typ func(*Type, bool) error
	typ = func(t *Type, top bool) error {
		if types[t] {
			return nil
		}
		if err := p.charge(1 + uint64(len(t.Imports)) + uint64(len(t.Exports)) + uint64(len(t.Parameters))); err != nil {
			return err
		}
		types[t] = true
		add(t.Resource)
		if !top && (t.Kind == model.Component || t.Kind == model.Instance) {
			if err := p.charge(uint64(len(t.Captures))); err != nil {
				return err
			}
			for _, id := range t.Captures {
				add(id)
			}
			return nil
		}
		for _, es := range [][]Extern{t.Imports, t.Exports} {
			for _, e := range es {
				if err := typ(e.Type, false); err != nil {
					return err
				}
			}
		}
		if err := value(t.Value); err != nil {
			return err
		}
		for _, f := range t.Parameters {
			if err := value(f.Type); err != nil {
				return err
			}
		}
		return value(t.Result)
	}
	if err := typ(root, true); err != nil {
		return nil, err
	}
	if err := p.charge(uint64(len(resources))); err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(resources))
	for id := range resources {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// freshInstance eliminates an instance type's abstract export binders at an
// import/export declaration. Repeated uses of the same instance type therefore
// receive distinct resources; equality aliases within each use stay equal.
// Component instantiation's import substitution is a separate linking step.
func (p *TypePool) freshInstance(t *Type) (*Type, error) {
	bindings := map[uint64]*Type{}
	symbols := map[*valueBinding]*valueBinding{}
	visited := map[*Type]bool{}
	var collect func(*Type) error
	collect = func(t *Type) error {
		if visited[t] {
			return nil
		}
		if err := p.charge(1 + uint64(len(t.Exports))); err != nil {
			return err
		}
		visited[t] = true
		for _, e := range t.Exports {
			if e.Sort == model.ValueSort && e.Fresh {
				if symbols[e.Type.valueBinding] == nil {
					symbols[e.Type.valueBinding] = &valueBinding{}
				}
			} else if e.Sort == model.TypeSort && e.Fresh {
				if bindings[e.Type.Resource] == nil {
					r, err := p.abstract()
					if err != nil {
						return err
					}
					bindings[e.Type.Resource] = r
				}
			} else if e.Sort == model.InstanceSort {
				if err := collect(e.Type); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := collect(t); err != nil {
		return nil, err
	}
	c := substitution{pool: p, bindings: bindings, symbols: symbols, types: map[*Type]*Type{}, values: map[*model.Type]*model.Type{}, cloneAll: true}
	return c.typ(t)
}

type substitution struct {
	pool          *TypePool
	bindings      map[uint64]*Type
	types         map[*Type]*Type
	values        map[*model.Type]*model.Type
	typeBindings  map[*Type]*Type
	valueBindings map[*model.Type]*model.Type
	cloneAll      bool
	publishing    *TypeScope
	symbols       map[*valueBinding]*valueBinding
}

func (c *substitution) symbol(in *valueBinding) (*valueBinding, error) {
	for n := 0; n <= len(c.symbols); n++ {
		next := c.symbols[in]
		if next == nil || next == in {
			return in, nil
		}
		if err := c.pool.charge(1); err != nil {
			return nil, err
		}
		in = next
	}
	return nil, invalid("cyclic value substitution")
}

func (c *substitution) id(id uint64) (uint64, error) {
	for n := 0; n <= len(c.bindings); n++ {
		r := c.bindings[id]
		if r == nil || r.Resource == id {
			return id, nil
		}
		if err := c.pool.charge(1); err != nil {
			return 0, err
		}
		id = r.Resource
	}
	return 0, invalid("cyclic resource substitution")
}

func (c *substitution) value(in *model.Type) (*model.Type, error) {
	if in == nil {
		return nil, nil
	}
	if replacement, ok := c.valueBindings[in]; ok {
		return replacement, nil
	}
	if primitive(in.Kind) || !c.cloneAll && !c.pool.values[in].resource {
		return in, nil
	}
	if out := c.values[in]; out != nil {
		return out, nil
	}
	if err := c.pool.charge(1 + uint64(len(in.Fields)) + uint64(len(in.Cases))); err != nil {
		return nil, err
	}
	out := *in
	c.values[in] = &out
	var err error
	out.ResourceID, err = c.id(in.ResourceID)
	if err != nil {
		return nil, err
	}
	for _, fs := range []struct {
		in  []model.Field
		out *[]model.Field
	}{{in.Fields, &out.Fields}, {in.Cases, &out.Cases}} {
		*fs.out = append([]model.Field(nil), fs.in...)
		for i := range *fs.out {
			(*fs.out)[i].Type, err = c.value(fs.in[i].Type)
			if err != nil {
				return nil, err
			}
		}
	}
	out.Element, err = c.value(in.Element)
	if err != nil {
		return nil, err
	}
	out.Key, err = c.value(in.Key)
	if err != nil {
		return nil, err
	}
	info := c.pool.values[in]
	if info.resourceName != nil {
		info.resourceName, err = c.typ(info.resourceName)
		if err != nil {
			return nil, err
		}
	}
	info.minimumResource = out.ResourceID
	for _, fs := range [][]model.Field{out.Fields, out.Cases} {
		for _, f := range fs {
			if f.Type != nil {
				info.minimumResource = minimum(info.minimumResource, c.pool.values[f.Type].minimumResource)
			}
		}
	}
	for _, v := range []*model.Type{out.Element, out.Key} {
		if v != nil {
			info.minimumResource = minimum(info.minimumResource, c.pool.values[v].minimumResource)
		}
	}
	c.pool.values[&out] = info
	return &out, nil
}

func (c *substitution) typ(in *Type) (*Type, error) {
	if replacement, ok := c.typeBindings[in]; ok {
		return replacement, nil
	}
	if out := c.types[in]; out != nil {
		return out, nil
	}
	if in.Kind == model.Resource {
		if !c.cloneAll && c.bindings[in.Resource] == nil {
			return in, nil
		}
		if err := c.pool.charge(1); err != nil {
			return nil, err
		}
		out := *in
		var err error
		out.Resource, err = c.id(in.Resource)
		if err != nil {
			return nil, err
		}
		out.minimumResource = out.Resource
		if c.bindings[in.Resource] != nil {
			out.Representation = 0
			out.Destructor = nil
			out.definingScope = nil
		}
		c.types[in] = &out
		return &out, nil
	}
	if err := c.pool.charge(1 + uint64(len(in.Parameters)) + uint64(len(in.Captures)) + uint64(len(in.Imports)) + uint64(len(in.Exports))); err != nil {
		return nil, err
	}
	out := *in
	c.types[in] = &out
	var err error
	out.valueBinding, err = c.symbol(in.valueBinding)
	if err != nil {
		return nil, err
	}
	out.Value, err = c.value(in.Value)
	if err != nil {
		return nil, err
	}
	out.Result, err = c.value(in.Result)
	if err != nil {
		return nil, err
	}
	out.Parameters = append([]model.Field(nil), in.Parameters...)
	for i := range out.Parameters {
		out.Parameters[i].Type, err = c.value(in.Parameters[i].Type)
		if err != nil {
			return nil, err
		}
	}
	for _, es := range []struct {
		in  []Extern
		out *[]Extern
	}{{in.Imports, &out.Imports}, {in.Exports, &out.Exports}} {
		*es.out = append([]Extern(nil), es.in...)
		for i := range *es.out {
			if c.publishing != nil && es.in[i].Sort == model.TypeSort {
				(*es.out)[i].Type, err = c.publicType(es.in[i].Type)
			} else {
				(*es.out)[i].Type, err = c.typ(es.in[i].Type)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	out.Captures = append([]uint64(nil), in.Captures...)
	for i, id := range out.Captures {
		out.Captures[i], err = c.id(id)
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out.Captures, func(i, j int) bool { return out.Captures[i] < out.Captures[j] })
	out.minimumResource = 0
	if out.Kind == model.Component || out.Kind == model.Instance {
		if len(out.Captures) != 0 {
			out.minimumResource = out.Captures[0]
		}
	} else {
		for _, v := range []*model.Type{out.Value, out.Result} {
			if v != nil {
				out.minimumResource = minimum(out.minimumResource, c.pool.values[v].minimumResource)
			}
		}
		for _, f := range out.Parameters {
			out.minimumResource = minimum(out.minimumResource, c.pool.values[f.Type].minimumResource)
		}
	}
	return &out, nil
}

func (c *substitution) publicType(in *Type) (*Type, error) {
	known := c.publishing.visibility.exports
	if in.Kind == model.Resource && known[nameable{resource: in}] || in.Value != nil && named(in.Value.Kind) && known[nameable{value: in.Value}] {
		// Existing external names remain usable by dependent functions. Only
		// this exported type binding receives a new name, preserving order.
		if err := c.pool.charge(1); err != nil {
			return nil, err
		}
		out := *in
		if in.Value != nil {
			value := *in.Value
			out.Value = &value
			c.pool.values[&value] = c.pool.values[in.Value]
		}
		return &out, nil
	}
	return c.typ(in)
}
