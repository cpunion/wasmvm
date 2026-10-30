package validate

import (
	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
)

type ValueDefinition struct {
	Definition *model.ValueDefinition
	Type       *Type
	Value      canonical.Value
}

type Start struct {
	Definition *model.Start
	Type       *Type
}

// A binding denotes a precise constant or one abstract value occurrence.
// Unknown imported values are equal only through an equality bound; constants
// compare by their typed contents. Symbols never grant runtime handle access.
type valueBinding struct {
	literal *canonical.Value
}

// ValueKey is an opaque compilation-local equality symbol. LiteralValue, when
// present, supplies its precise immutable contents. Consumers must copy those
// contents before handing mutable aggregates to host code.
func (t *Type) ValueKey() any { return t.valueBinding }
func (t *Type) LiteralValue() (canonical.Value, bool) {
	if t.valueBinding == nil || t.valueBinding.literal == nil {
		return canonical.Value{}, false
	}
	return *t.valueBinding.literal, true
}

func (s *TypeScope) valueType(ref model.TypeRef) (*Type, error) {
	v, err := s.Value(ref)
	if err != nil {
		return nil, err
	}
	info := s.pool.values[v]
	return &Type{Kind: v.Kind, Value: v, owner: s.pool, depth: info.depth, borrow: info.borrow, resource: info.resource, minimumResource: info.minimumResource, valueBinding: &valueBinding{}}, nil
}

func (c *componentContext) value(def *model.ValueDefinition) error {
	if def == nil {
		return invalid("missing value definition")
	}
	s := c.result.scope
	t, err := s.valueType(def.Type)
	if err != nil {
		return err
	}
	v, err := canonical.DecodeLiteral(t.Value, def.Bytes, canonical.Limits{MaxNodes: s.pool.limits.MaxNodes - s.pool.nodes, MaxDepth: s.pool.limits.MaxDepth, MaxBytes: uint64(s.pool.limits.MaxNodes) * 64})
	if err != nil {
		return err
	}
	var charge func(canonical.Value) error
	charge = func(v canonical.Value) error {
		if err := s.pool.charge(1 + (uint64(len(v.Text))+63)/64); err != nil {
			return err
		}
		for _, child := range v.Elements {
			if err := charge(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := charge(v); err != nil {
		return err
	}
	t.valueBinding.literal = &v
	c.result.Values = append(c.result.Values, &ValueDefinition{Definition: def, Type: t, Value: v})
	s.items[model.ValueSort] = append(s.items[model.ValueSort], t)
	return c.addLinear(model.ValueSort, t, nil)
}

func (c *componentContext) start(def *model.Start) error {
	if def == nil {
		return invalid("missing start definition")
	}
	e, err := c.item(model.SortIndex{Sort: model.FunctionSort, Index: def.Function})
	if err != nil {
		return err
	}
	f := e.Type
	results := uint32(0)
	if f.Result != nil {
		results = 1
	}
	if len(f.Parameters) != len(def.Arguments) || def.Results != results {
		return invalid("start arity")
	}
	r := c.result.scope.pool.relation()
	for i, index := range def.Arguments {
		item := model.SortIndex{Sort: model.ValueSort, Index: index}
		arg, err := c.item(item)
		if err != nil {
			return err
		}
		ok, err := r.value(arg.Type.Value, f.Parameters[i].Type)
		if err != nil {
			return err
		}
		if !ok {
			return invalid("start argument type")
		}
		if err := c.use(item); err != nil {
			return err
		}
	}
	if f.Result != nil {
		s := c.result.scope
		info := s.pool.values[f.Result]
		t := &Type{Kind: f.Result.Kind, Value: f.Result, owner: s.pool, depth: info.depth, resource: info.resource, minimumResource: info.minimumResource, valueBinding: &valueBinding{}}
		if c.localValues == nil {
			c.localValues = map[*valueBinding]bool{}
		}
		c.localValues[t.valueBinding] = true
		s.items[model.ValueSort] = append(s.items[model.ValueSort], t)
		if err := c.addLinear(model.ValueSort, t, nil); err != nil {
			return err
		}
	}
	c.result.Starts = append(c.result.Starts, &Start{Definition: def, Type: f})
	return nil
}

func (p *TypePool) sameValue(a, b *valueBinding) (bool, error) {
	if a == b {
		return true, nil
	}
	if a == nil || b == nil || a.literal == nil || b.literal == nil {
		return false, nil
	}
	var equal func(canonical.Value, canonical.Value) (bool, error)
	equal = func(a, b canonical.Value) (bool, error) {
		if err := p.charge(1 + uint64(len(a.Text)+len(b.Text))/64); err != nil {
			return false, err
		}
		if a.Bits != b.Bits || a.Case != b.Case || a.Text != b.Text || len(a.Elements) != len(b.Elements) {
			return false, nil
		}
		for i := range a.Elements {
			if ok, err := equal(a.Elements[i], b.Elements[i]); err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	}
	return equal(*a.literal, *b.literal)
}

// Linear obligations belong to runtime value occurrences, never shared type
// metadata. Instance aliases share their obligations; moving a value creates a
// new obligation at the destination. Failure aborts the compilation.
type linear struct {
	used     bool
	children map[string]*linear
}

func (c *componentContext) newLinear(sort model.Sort, t *Type) (*linear, error) {
	if sort != model.ValueSort && sort != model.InstanceSort {
		return nil, nil
	}
	if err := c.result.scope.pool.charge(1); err != nil {
		return nil, err
	}
	if sort == model.ValueSort {
		l := &linear{}
		c.obligations = append(c.obligations, l)
		return l, nil
	}
	has, err := c.hasValues(t)
	if err != nil || !has {
		return nil, err
	}
	l := &linear{children: map[string]*linear{}}
	for _, e := range t.Exports {
		child, err := c.newLinear(e.Sort, e.Type)
		if err != nil {
			return nil, err
		}
		if child != nil {
			l.children[e.Name.Text] = child
		}
	}
	return l, nil
}

func (c *componentContext) hasValues(t *Type) (bool, error) {
	if v, ok := c.containsValues[t]; ok {
		return v, nil
	}
	if err := c.result.scope.pool.charge(1 + uint64(len(t.Exports))); err != nil {
		return false, err
	}
	has := false
	for _, e := range t.Exports {
		if e.Sort == model.ValueSort {
			has = true
		} else if e.Sort == model.InstanceSort {
			child, err := c.hasValues(e.Type)
			if err != nil {
				return false, err
			}
			has = has || child
		}
	}
	if c.containsValues == nil {
		c.containsValues = map[*Type]bool{}
	}
	c.containsValues[t] = has
	return has, nil
}

func (c *componentContext) addLinear(sort model.Sort, t *Type, existing *linear) error {
	if sort != model.ValueSort && sort != model.InstanceSort {
		return nil
	}
	var err error
	if existing == nil {
		existing, err = c.newLinear(sort, t)
		if err != nil {
			return err
		}
	}
	if c.linear == nil {
		c.linear = map[model.Sort][]*linear{}
	}
	c.linear[sort] = append(c.linear[sort], existing)
	return nil
}

func (c *componentContext) useLinear(l *linear) error {
	if l == nil {
		return nil
	}
	if err := c.result.scope.pool.charge(1); err != nil {
		return err
	}
	if l.children == nil {
		if l.used {
			return invalid("value consumed more than once")
		}
		l.used = true
	} else {
		for _, child := range l.children {
			if err := c.useLinear(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *componentContext) use(item model.SortIndex) error {
	if item.Sort != model.ValueSort && item.Sort != model.InstanceSort {
		return nil
	}
	return c.useLinear(c.linear[item.Sort][item.Index])
}
