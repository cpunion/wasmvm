package validate

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"math"
)

func (c *context) checkTypes() {
	m := c.m
	if len(m.TypeDefs) != 0 && (c.features&wasm.GC == 0 || len(m.TypeDefs) != len(m.Types)) {
		c.fail("invalid composite type inventory")
	}
	groups := m.TypeGroups
	if groups == nil {
		groups = make([]module.TypeGroup, len(m.Types))
		for i := range groups {
			groups[i] = module.TypeGroup{Start: uint32(i), Count: 1}
		}
	}
	c.typeLabels = make([]uint32, len(m.Types))
	next := uint64(0)
	for _, g := range groups {
		end := uint64(g.Start) + uint64(g.Count)
		if uint64(g.Start) != next || end > uint64(len(m.Types)) {
			c.fail("invalid recursive type group")
		}
		c.checkTypeGroup(g, math.MaxUint32)
		next = end
	}
	if next != uint64(len(m.Types)) {
		c.fail("incomplete type groups")
	}
}
func (c *context) checkTypeGroup(g module.TypeGroup, maximum uint32) {
	m := c.m
	end := uint64(g.Start) + uint64(g.Count)
	for i := g.Start; uint64(i) < end; i++ {
		bound := int(i)
		if c.features&wasm.GC != 0 {
			bound = int(end)
		}
		for _, ts := range [][]wasm.ValueType{m.Types[i].Params, m.Types[i].Results} {
			for _, t := range ts {
				c.valueType(t, bound)
			}
		}
		d := m.TypeDef(i)
		if len(d.Supers) > 1 {
			c.fail("multiple supertypes")
		}
		for _, parent := range d.Supers {
			c.index(parent, int(i), "supertype")
			if c.lexicalTypes && (c.m.TypeDef(parent).Kind == 0 || parent < c.lexicalBound && (c.typeLabels[parent] == 0 || uint64(c.typeLabels[parent]) >= uint64(len(c.registry.Types)))) {
				c.fail("unbound lexical supertype")
			}
			if !m.TypeDef(parent).Open {
				c.fail("final supertype")
			}
		}
		switch d.Kind {
		case module.FunctionKind:
			if len(d.Fields) != 0 {
				c.fail("function fields")
			}
		case module.StructKind, module.ArrayKind:
			if d.Kind == module.ArrayKind && len(d.Fields) != 1 {
				c.fail("array field count")
			}
			if len(m.Types[i].Params) != 0 || len(m.Types[i].Results) != 0 {
				c.fail("nonfunction signature")
			}
		default:
			c.fail("unknown composite type")
		}
		for _, f := range d.Fields {
			c.valueType(f.Value, bound)
			if f.Packed != 0 && (f.Packed != 8 && f.Packed != 16 || f.Value != wasm.I32) {
				c.fail("invalid packed field")
			}
		}
	}
	if !c.registry.InternGroup(m, g.Start, g.Count, c.typeLabels, maximum) {
		c.limit("canonical types")
	}
	for i := g.Start; uint64(i) < end; i++ {
		for _, p := range m.TypeDef(i).Supers {
			c.structuralSubtype(i, p)
		}
	}
}

func (c *context) storageMatches(a, b module.FieldType) bool {
	if a.Packed != b.Packed || a.Mutable != b.Mutable {
		return false
	}
	if a.Packed != 0 {
		return true
	}
	return c.matches(a.Value, b.Value) && (!a.Mutable || c.matches(b.Value, a.Value))
}
func (c *context) structuralSubtype(child, parent uint32) {
	a, b := c.m.TypeDef(child), c.m.TypeDef(parent)
	if a.Kind != b.Kind {
		c.fail("supertype kind mismatch")
	}
	if a.Kind == module.FunctionKind {
		f, g := c.m.Types[child], c.m.Types[parent]
		if len(f.Params) != len(g.Params) || len(f.Results) != len(g.Results) {
			c.fail("function subtype arity")
		}
		for i, t := range f.Params {
			if !c.matches(g.Params[i], t) {
				c.fail("function parameter variance")
			}
		}
		for i, t := range f.Results {
			if !c.matches(t, g.Results[i]) {
				c.fail("function result variance")
			}
		}
	} else {
		if len(a.Fields) < len(b.Fields) {
			c.fail("struct subtype width")
		}
		for i, f := range b.Fields {
			if !c.storageMatches(a.Fields[i], f) {
				c.fail("field subtype variance")
			}
		}
	}
}
