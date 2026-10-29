package validate

import (
	"fmt"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	corevalidate "github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Component is published only after every definition has validated. Unsupported
// definitions fail the whole check; no start function or guest allocator runs
// during validation. The owned definition must remain immutable on success.
type Component struct {
	definition    *model.Binary
	scope         *TypeScope
	CoreModules   []*CoreModule
	CoreInstances []*CoreInstance
	Children      []*Component
	Canonical     []*CanonicalFunction
	Type          *Type
}

func (c *Component) Definition() *model.Binary { return c.definition }

type CoreModule struct {
	Definition *corevalidate.Module
	Type       *CoreModuleType
}

type CoreInstance struct {
	Exports []CoreExtern
	byName  map[string]CoreExtern
}

type componentContext struct {
	result    *Component
	parent    *componentContext
	core      map[model.Sort][]CoreExtern
	modules   []*CoreModuleType
	functions []module.FuncType
	pointers  map[*model.Type]bool
}

func Check(definition *model.Binary, limits Limits) (*Component, error) {
	p := NewTypePool(limits)
	return checkComponent(definition, p, nil)
}

func checkComponent(definition *model.Binary, pool *TypePool, parent *componentContext) (*Component, error) {
	if definition == nil {
		return nil, invalid("missing component definition")
	}
	var outer *TypeScope
	if parent != nil {
		outer = parent.result.scope
	}
	scope, err := pool.Scope(outer, true)
	if err != nil {
		return nil, err
	}
	c := &componentContext{result: &Component{definition: definition, scope: scope}, parent: parent, core: map[model.Sort][]CoreExtern{}, pointers: map[*model.Type]bool{}}
	if err := pool.charge(uint64(len(definition.Definitions))); err != nil {
		return nil, err
	}
	for _, d := range definition.Definitions {
		if err := c.definition(d); err != nil {
			return nil, fmt.Errorf("component definition at byte %#x: %w", d.Offset, err)
		}
	}
	c.result.Type = &Type{Kind: model.Component, Imports: scope.imports.ordered, Exports: scope.exports.ordered, owner: pool, depth: 1}
	for _, group := range [][]Extern{c.result.Type.Imports, c.result.Type.Exports} {
		for _, e := range group {
			c.result.Type.depth = max(c.result.Type.depth, e.Type.depth+1)
		}
	}
	if c.result.Type.depth > pool.limits.MaxDepth {
		return nil, limit("component interface nesting")
	}
	return c.result, nil
}

func (c *componentContext) definition(d model.Definition) error {
	s := c.result.scope
	switch d.Kind {
	case 1:
		m, err := s.pool.coreModule(d.CoreModule)
		if err != nil {
			return err
		}
		c.result.CoreModules = append(c.result.CoreModules, m)
		c.modules = append(c.modules, m.Type)
	case 2:
		return c.coreInstance(d.Instance)
	case 3:
		return s.DefineCore(d.CoreType, c.result.definition.CoreTypes)
	case 4:
		child, err := checkComponent(d.Component, s.pool, c)
		if err != nil {
			return err
		}
		c.result.Children = append(c.result.Children, child)
		s.items[model.ComponentSort] = append(s.items[model.ComponentSort], child.Type)
	case 6:
		return c.alias(d.Alias)
	case 7:
		_, err := s.Define(d.Type, c.functions)
		return err
	case 8:
		return c.canonical(d.Canonical)
	case 10:
		t, err := s.Declare(d.Import, false)
		if err != nil {
			return err
		}
		if d.Import.Type.Sort == model.CoreModuleSort {
			c.modules = append(c.modules, t.CoreModule)
		}
	case 11:
		return c.export(d.Export)
	default:
		return fmt.Errorf("%w: component definition kind %d", wasm.ErrUnsupported, d.Kind)
	}
	return nil
}

func (c *componentContext) addCore(e CoreExtern) error {
	if err := c.result.scope.pool.charge(1); err != nil {
		return err
	}
	sort := model.CoreFunctionSort + model.Sort(e.Descriptor.Kind)
	c.core[sort] = append(c.core[sort], e)
	if e.Descriptor.Kind == 0 {
		c.functions = append(c.functions, c.result.scope.pool.registry.Types[e.Descriptor.Type].Function)
	}
	return nil
}

func (c *componentContext) coreAt(item model.SortIndex) (CoreExtern, error) {
	if item.Sort < model.CoreFunctionSort || item.Sort > model.CoreTagSort || uint64(item.Index) >= uint64(len(c.core[item.Sort])) {
		return CoreExtern{}, invalid("Core external index out of bounds")
	}
	return c.core[item.Sort][item.Index], nil
}

func (c *componentContext) alias(a *model.Alias) error {
	if a == nil {
		return invalid("missing alias")
	}
	if a.Kind == 1 {
		if uint64(a.Instance) >= uint64(len(c.result.CoreInstances)) {
			return invalid("Core instance alias index out of bounds")
		}
		e, ok := c.result.CoreInstances[a.Instance].byName[a.Name]
		if !ok || model.CoreFunctionSort+model.Sort(e.Descriptor.Kind) != a.Sort {
			return invalid("Core instance export alias sort or name")
		}
		return c.addCore(e)
	}
	if a.Kind == 0 {
		s := c.result.scope
		instances := s.items[model.InstanceSort]
		if uint64(a.Instance) >= uint64(len(instances)) {
			return invalid("component instance alias index")
		}
		for _, e := range instances[a.Instance].Exports {
			if e.Name.Text != a.Name {
				continue
			}
			if e.Sort != a.Sort {
				return invalid("component instance export alias sort")
			}
			if a.Sort == model.TypeSort {
				_, err := s.Alias(e.Type)
				return err
			}
			if err := s.pool.charge(1); err != nil {
				return err
			}
			if a.Sort == model.CoreModuleSort {
				c.modules = append(c.modules, e.Type.CoreModule)
			} else {
				s.items[a.Sort] = append(s.items[a.Sort], e.Type)
			}
			return nil
		}
		return invalid("component instance export alias name")
	}
	if a.Kind == 2 && a.Sort == model.CoreModuleSort {
		at := c
		for i := uint32(0); i < a.Count; i++ {
			at = at.parent
			if at == nil {
				return invalid("outer component scope out of bounds")
			}
		}
		if uint64(a.Index) >= uint64(len(at.modules)) {
			return invalid("outer Core module index out of bounds")
		}
		if err := c.result.scope.pool.charge(1); err != nil {
			return err
		}
		c.modules = append(c.modules, at.modules[a.Index])
		return nil
	}
	return c.result.scope.aliasDeclaration(a)
}

func (c *componentContext) export(e *model.Export) error {
	if e == nil {
		return invalid("missing component export")
	}
	s := c.result.scope
	var t *Type
	switch e.Item.Sort {
	case model.TypeSort:
		var err error
		t, err = s.At(e.Item.Index)
		if err != nil {
			return err
		}
	case model.CoreModuleSort:
		if uint64(e.Item.Index) >= uint64(len(c.modules)) {
			return invalid("export Core module index")
		}
		t = &Type{Kind: 0x50, CoreModule: c.modules[e.Item.Index], owner: s.pool, depth: 1}
	case model.FunctionSort, model.InstanceSort, model.ComponentSort:
		items := s.items[e.Item.Sort]
		if uint64(e.Item.Index) >= uint64(len(items)) {
			return invalid("component export index")
		}
		t = items[e.Item.Index]
	default:
		return invalid("component export sort")
	}
	if e.Type != nil {
		expected, err := s.Reference(*e.Type)
		if err != nil {
			return err
		}
		if e.Type.Sort != e.Item.Sort {
			return invalid("component export type sort")
		}
		ok, err := s.pool.equal(t, expected)
		if err != nil {
			return err
		}
		if !ok {
			return invalid("component export type ascription")
		}
		t = expected
	}
	if err := s.checkVisible(t, true, e.Item.Sort == model.TypeSort); err != nil {
		return err
	}
	if e.Item.Sort == model.TypeSort {
		var err error
		t, err = s.publish(t, true)
		if err != nil {
			return err
		}
	} else if e.Item.Sort == model.InstanceSort {
		if err := s.expose(t, true); err != nil {
			return err
		}
	}
	if err := s.exports.add(e.Name, e.Item.Sort, t, s.pool); err != nil {
		return err
	}
	if e.Item.Sort == model.TypeSort {
		_, err := s.Alias(t)
		return err
	}
	if err := s.pool.charge(1); err != nil {
		return err
	}
	if e.Item.Sort == model.CoreModuleSort {
		c.modules = append(c.modules, t.CoreModule)
	} else {
		s.items[e.Item.Sort] = append(s.items[e.Item.Sort], t)
	}
	return nil
}

func (c *componentContext) coreInstance(x *model.InstanceExpr) error {
	if x == nil {
		return invalid("missing Core instance expression")
	}
	pool := c.result.scope.pool
	if err := pool.charge(1 + uint64(len(x.Arguments)) + uint64(len(x.Exports))); err != nil {
		return err
	}
	instance := &CoreInstance{byName: map[string]CoreExtern{}}
	if x.Instantiate {
		if uint64(x.Index) >= uint64(len(c.modules)) {
			return invalid("Core module index out of bounds")
		}
		m := c.modules[x.Index]
		arguments := map[string]*CoreInstance{}
		for _, a := range x.Arguments {
			if arguments[a.Name] != nil {
				return invalid("duplicate Core instantiation argument")
			}
			if a.Item.Sort != model.CoreInstanceSort || uint64(a.Item.Index) >= uint64(len(c.result.CoreInstances)) {
				return invalid("Core instantiation argument sort or index")
			}
			arguments[a.Name] = c.result.CoreInstances[a.Item.Index]
		}
		if err := pool.charge(uint64(len(m.Imports)) + uint64(len(m.Exports))); err != nil {
			return err
		}
		for _, im := range m.Imports {
			arg := arguments[im.Module]
			if arg == nil {
				return invalid("missing Core module instantiation argument")
			}
			actual, ok := arg.byName[im.Name]
			if !ok || !module.ExternSubtype(&pool.registry, actual.Descriptor, im.Descriptor) {
				return invalid("Core module import type or name mismatch")
			}
		}
		instance.Exports = append([]CoreExtern(nil), m.Exports...)
		for _, e := range instance.Exports {
			instance.byName[e.Name] = e
		}
	} else {
		for _, e := range x.Exports {
			if _, exists := instance.byName[e.Name.Text]; exists {
				return invalid("duplicate Core instance export")
			}
			actual, err := c.coreAt(e.Item)
			if err != nil {
				return err
			}
			actual.Module, actual.Name = "", e.Name.Text
			instance.Exports = append(instance.Exports, actual)
			instance.byName[e.Name.Text] = actual
		}
	}
	c.result.CoreInstances = append(c.result.CoreInstances, instance)
	return nil
}
