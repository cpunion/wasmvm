package module

// Interface retains immutable type declarations and object descriptors from a
// validated frozen module. It retains no function bodies, data or initializers.
// Its fields stay private so callers cannot edit a compiled module's type graph.
type Interface struct{ definition *interfaceDefinition }

type interfaceDefinition struct {
	types     *Module
	imports   []Import
	exports   []Export
	functions []uint32
	tables    []TableType
	memories  []Limits
	globals   []GlobalType
	tags      []uint32
	nodes     uint64
}

// SnapshotInterface requires an immutable, validated module. Returned views
// can outlive compilation handles without retaining executable publications.
func SnapshotInterface(m *Module) Interface {
	if m == nil {
		return Interface{}
	}
	d := &interfaceDefinition{types: &Module{Types: m.Types, TypeDefs: m.TypeDefs, TypeGroups: m.TypeGroups}, imports: m.Imports, exports: m.Exports, tables: m.Tables, memories: m.Memories, tags: m.Tags, functions: make([]uint32, len(m.Functions)), globals: make([]GlobalType, len(m.Globals))}
	d.nodes = uint64(len(m.Types)) + uint64(len(m.TypeGroups)) + uint64(len(m.Imports)) + uint64(len(m.Exports)) + uint64(len(m.Functions)) + uint64(len(m.Tables)) + uint64(len(m.Memories)) + uint64(len(m.Globals)) + uint64(len(m.Tags))
	for i, t := range m.Types {
		definition := m.TypeDef(uint32(i))
		d.nodes += uint64(len(t.Params)) + uint64(len(t.Results)) + uint64(len(definition.Supers)) + uint64(len(definition.Fields))
	}
	for i, f := range m.Functions {
		d.functions[i] = f.Type
	}
	for i, g := range m.Globals {
		d.globals[i] = g.Type
	}
	return Interface{d}
}

func (i Interface) Valid() bool { return i.definition != nil }
func (i Interface) Nodes() uint64 {
	if !i.Valid() {
		return 0
	}
	return i.definition.nodes
}

// InterfaceShape owns canonical import/export descriptors in the supplied
// registry's identity space. It exposes no original declaration slices.
type InterfaceShape struct{ Imports, Exports []Import }

func (i Interface) Intern(registry *TypeRegistry, maximum uint32) (InterfaceShape, bool) {
	if !i.Valid() || registry == nil {
		return InterfaceShape{}, false
	}
	d := i.definition
	ids, ok := registry.InternModule(d.types, maximum)
	if !ok {
		return InterfaceShape{}, false
	}
	normalize := func(e Import) Import {
		if e.Kind == 0 || e.Kind == 4 {
			e.Type = ids[e.Type]
		}
		e.Table.Element = ResolveType(e.Table.Element, ids)
		e.Global.Value = ResolveType(e.Global.Value, ids)
		return e
	}
	out := InterfaceShape{}
	spaces := [5][]Import{}
	for _, im := range d.imports {
		e := normalize(im)
		out.Imports = append(out.Imports, e)
		spaces[im.Kind] = append(spaces[im.Kind], e)
	}
	for _, ex := range d.exports {
		var e Import
		if ex.Index < uint32(len(spaces[ex.Kind])) {
			e = spaces[ex.Kind][ex.Index]
		} else {
			x := ex.Index - uint32(len(spaces[ex.Kind]))
			e.Kind = ex.Kind
			switch ex.Kind {
			case 0:
				e.Type = d.functions[x]
			case 1:
				e.Table = d.tables[x]
			case 2:
				e.Memory = d.memories[x]
			case 3:
				e.Global = d.globals[x]
			case 4:
				e.Type = d.tags[x]
			}
			e = normalize(e)
		}
		e.Module, e.Name = "", ex.Name
		out.Exports = append(out.Exports, e)
	}
	return out, true
}

// Fork copies the mutable registry index once; existing canonical definitions
// are immutable and remain shared. New groups cannot mutate the source registry.
func (r *TypeRegistry) Fork() TypeRegistry {
	var out TypeRegistry
	out.Init()
	if r == nil || len(r.Types) == 0 {
		return out
	}
	out.Types = append([]CanonicalType(nil), r.Types...)
	out.Keys = make(map[string]uint32, len(r.Keys))
	for key, id := range r.Keys {
		out.Keys[key] = id
	}
	for _, t := range r.Types[1:] {
		parent := uint32(0)
		if len(t.Definition.Supers) != 0 {
			parent = t.Definition.Supers[0]
		}
		out.Hierarchy.Add(parent)
	}
	return out
}
