package validate

import "github.com/cpunion/wasmvm/internal/component/model"

type ComponentInstance struct {
	Definition *model.InstanceExpr
	Type       *Type
}

func (c *componentContext) item(item model.SortIndex) (Extern, error) {
	s := c.result.scope
	e := Extern{Sort: item.Sort}
	var err error
	switch item.Sort {
	case model.TypeSort:
		e.Type, err = s.At(item.Index)
	case model.CoreModuleSort:
		if uint64(item.Index) >= uint64(len(c.modules)) {
			return e, invalid("Core module argument index")
		}
		e.Type = &Type{Kind: 0x50, CoreModule: c.modules[item.Index], owner: s.pool, depth: 1}
	case model.FunctionSort, model.ComponentSort, model.InstanceSort, model.ValueSort:
		if uint64(item.Index) >= uint64(len(s.items[item.Sort])) {
			return e, invalid("component argument index")
		}
		e.Type = s.items[item.Sort][item.Index]
	default:
		return e, invalid("component argument sort")
	}
	return e, err
}

func (c *componentContext) componentInstance(expr *model.InstanceExpr) error {
	if expr == nil {
		return invalid("missing component instance")
	}
	p := c.result.scope.pool
	if err := p.charge(1 + uint64(len(expr.Arguments)) + uint64(len(expr.Exports))); err != nil {
		return err
	}
	var instance *Type
	if expr.Instantiate {
		components := c.result.scope.items[model.ComponentSort]
		if uint64(expr.Index) >= uint64(len(components)) {
			return invalid("component instantiation index")
		}
		typ := components[expr.Index]
		args := map[string]Extern{}
		for _, arg := range expr.Arguments {
			if _, ok := args[arg.Name]; ok {
				return invalid("duplicate component instantiation argument")
			}
			e, err := c.item(arg.Item)
			if err != nil {
				return err
			}
			if err := c.use(arg.Item); err != nil {
				return err
			}
			args[arg.Name] = e
		}
		r := p.relation()
		if err := p.charge(uint64(len(typ.Imports))); err != nil {
			return err
		}
		for _, im := range typ.Imports {
			actual, ok := args[im.Name.Text]
			if !ok {
				return invalid("missing component instantiation argument")
			}
			ok, err := r.external(actual, im)
			if err != nil {
				return err
			}
			if !ok {
				return invalid("component instantiation argument type")
			}
		}
		if err := c.freshExports(typ.Exports, r); err != nil {
			return err
		}
		var err error
		instance, err = r.rewrite().typ(&Type{Kind: model.Instance, Exports: typ.Exports, owner: p, depth: typ.depth})
		if err != nil {
			return err
		}
	} else {
		var names nameSet
		depth := uint32(1)
		for _, e := range expr.Exports {
			actual, err := c.item(e.Item)
			if err != nil {
				return err
			}
			if err := c.use(e.Item); err != nil {
				return err
			}
			parsed, err := parseName(e.Name.Text)
			if err != nil {
				return err
			}
			if parsed.annotation != "" {
				if err := c.result.scope.checkVisible(actual.Type, true, false); err != nil {
					return err
				}
				if parsed.resource != "" {
					resource := names.raw[parsed.resource]
					if resource.Type == nil || !c.result.scope.visibility.exports[nameable{resource: resource.Type}] {
						return invalid("annotated resource is not externally nameable")
					}
				}
			}
			if err := names.add(e.Name, e.Item.Sort, actual.Type, p); err != nil {
				return err
			}
			depth = max(depth, actual.Type.depth+1)
		}
		if depth > p.limits.MaxDepth {
			return limit("synthesized instance nesting")
		}
		instance = &Type{Kind: model.Instance, Exports: names.ordered, owner: p, depth: depth}
	}
	c.result.scope.items[model.InstanceSort] = append(c.result.scope.items[model.InstanceSort], instance)
	c.result.Instances = append(c.result.Instances, &ComponentInstance{Definition: expr, Type: instance})
	return c.addLinear(model.InstanceSort, instance, nil)
}

func (c *componentContext) freshExports(exports []Extern, r *relation) error {
	visited := map[*Type]bool{}
	var collect func([]Extern) error
	collect = func(exports []Extern) error {
		if err := r.pool.charge(uint64(len(exports))); err != nil {
			return err
		}
		for _, e := range exports {
			if e.Sort == model.ValueSort && e.Fresh && r.symbols[e.Type.valueBinding] == nil {
				if c.localValues == nil {
					c.localValues = map[*valueBinding]bool{}
				}
				v := &valueBinding{}
				r.symbols[e.Type.valueBinding] = v
				c.localValues[v] = true
			} else if e.Sort == model.TypeSort && e.Fresh && r.bindings[e.Type.Resource] == nil {
				t, err := r.pool.abstract()
				if err != nil {
					return err
				}
				r.bindings[e.Type.Resource] = t
				c.localResources[t.Resource] = true
			} else if e.Sort == model.InstanceSort && !visited[e.Type] {
				visited[e.Type] = true
				if err := collect(e.Type.Exports); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return collect(exports)
}

func (c *componentContext) exportBinders(e *Extern) error {
	seen := map[*Type]bool{}
	var visit func(*Extern) error
	visit = func(e *Extern) error {
		if err := c.result.scope.pool.charge(1); err != nil {
			return err
		}
		if e.Sort == model.ValueSort {
			e.Fresh = c.localValues[e.Type.valueBinding]
		} else if e.Sort == model.TypeSort && e.Type.Kind == model.Resource {
			id := e.Type.Resource
			e.Fresh = c.localResources[id] && !c.exportedResources[id]
			c.exportedResources[id] = true
		} else if e.Sort == model.InstanceSort && !seen[e.Type] {
			seen[e.Type] = true
			for i := range e.Type.Exports {
				if err := visit(&e.Type.Exports[i]); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return visit(e)
}
