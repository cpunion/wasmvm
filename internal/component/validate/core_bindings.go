package validate

import "github.com/cpunion/wasmvm/internal/core/module"

// CoreBindings owns one instantiation's bounded module matching registry.
// Compiled type pools remain immutable, including under concurrent binding.
type CoreBindings struct {
	owner   *TypePool
	pool    *TypePool
	actual  map[module.Interface]*CoreModuleType
	checked map[coreBinding]bool
}
type coreBinding struct {
	expected *Type
	actual   module.Interface
}

func (t *Type) CoreBindings() (*CoreBindings, error) {
	if t == nil || t.owner == nil || t.CoreModule == nil {
		return nil, invalid("missing Core module import type")
	}
	p := t.owner
	count := p.registry.Count()
	if uint64(count) > uint64(p.limits.MaxNodes) {
		return nil, limit("Core binding registry")
	}
	return &CoreBindings{owner: p, pool: &TypePool{limits: p.limits, nodes: uint32(count), registry: p.registry.Fork()}, actual: map[module.Interface]*CoreModuleType{}, checked: map[coreBinding]bool{}}, nil
}

func (b *CoreBindings) Check(t *Type, actual module.Interface) (bool, error) {
	if b == nil || t == nil || t.owner != b.owner || t.CoreModule == nil || !actual.Valid() {
		return false, invalid("foreign Core module binding")
	}
	key := coreBinding{t, actual}
	if b.checked[key] {
		return true, nil
	}
	if err := b.pool.charge(1); err != nil {
		return false, err
	}
	typ := b.actual[actual]
	if typ == nil {
		if err := b.pool.charge(actual.Nodes()); err != nil {
			return false, err
		}
		shape, ok := actual.Intern(&b.pool.registry, b.pool.limits.MaxNodes)
		if !ok {
			return false, limit("host canonical Core types")
		}
		typ = coreShape(shape)
		b.actual[actual] = typ
	}
	ok, err := (&relation{pool: b.pool}).coreModule(typ, t.CoreModule)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	b.checked[key] = true
	return true, nil
}

func coreShape(shape module.InterfaceShape) *CoreModuleType {
	typ := &CoreModuleType{}
	for _, im := range shape.Imports {
		typ.Imports = append(typ.Imports, CoreExtern{Module: im.Module, Name: im.Name, Descriptor: im})
	}
	for _, ex := range shape.Exports {
		typ.Exports = append(typ.Exports, CoreExtern{Name: ex.Name, Descriptor: ex})
	}
	return typ
}
