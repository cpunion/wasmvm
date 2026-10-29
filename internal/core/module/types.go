package module

import "github.com/cpunion/wasmvm/pkg/wasm"

const (
	FunctionKind byte = 0x60
	StructKind   byte = 0x5f
	ArrayKind    byte = 0x5e
)

type FieldType struct {
	Value   wasm.ValueType
	Packed  byte
	Mutable bool
}

func (f FieldType) Unpacked() wasm.ValueType {
	if f.Packed != 0 {
		return wasm.I32
	}
	return f.Value
}

type TypeDef struct {
	Kind   byte
	Open   bool
	Supers []uint32
	Fields []FieldType
}
type TypeGroup struct{ Start, Count uint32 }

func (m *Module) TypeDef(index uint32) TypeDef {
	if m.TypeDefs == nil {
		return TypeDef{Kind: FunctionKind}
	}
	return m.TypeDefs[index]
}
func (m *Module) Reference(index uint32, nullable bool) wasm.ValueType {
	family := wasm.FuncRef
	if uint64(index) < uint64(len(m.Types)) {
		switch m.TypeDef(index).Kind {
		case StructKind:
			family = wasm.StructRef
		case ArrayKind:
			family = wasm.ArrayRef
		}
	}
	return wasm.IndexedReference(family, index, nullable)
}

type CanonicalType struct {
	Definition TypeDef
	Function   FuncType
}

// TypeRegistry interns iso-recursive groups without recursively expanding
// their shapes. Keys distinguish group-relative back edges from external IDs.
type TypeRegistry struct {
	Types     []CanonicalType
	Hierarchy wasm.TypeHierarchy
	Keys      map[string]uint32
}

func (r *TypeRegistry) Init() {
	if r.Keys == nil {
		r.Keys = make(map[string]uint32)
		r.Types = []CanonicalType{{}}
		r.Hierarchy.Init()
	}
}
func (r *TypeRegistry) Count() int {
	if len(r.Types) == 0 {
		return 0
	}
	return len(r.Types) - 1
}
func (r *TypeRegistry) Rollback(n int) {
	clear(r.Types[n+1:])
	r.Types = r.Types[:n+1]
	r.Hierarchy.Rollback(n)
	for key, id := range r.Keys {
		if int(id) > n {
			delete(r.Keys, key)
		}
	}
}
func ResolveType(t wasm.ValueType, ids []uint32) wasm.ValueType {
	if i, ok := t.TypeIndex(); ok && !t.Canonical() {
		return t.WithTypeIndex(ids[i], true)
	}
	return t
}
func ResolveSignature(t FuncType, ids []uint32) FuncType {
	out := FuncType{Params: make([]wasm.ValueType, len(t.Params)), Results: make([]wasm.ValueType, len(t.Results))}
	for i, v := range t.Params {
		out.Params[i] = ResolveType(v, ids)
	}
	for i, v := range t.Results {
		out.Results[i] = ResolveType(v, ids)
	}
	return out
}
func (r *TypeRegistry) InternGroup(m *Module, start, count uint32, ids []uint32, maximum uint32) bool {
	r.Init()
	var key []byte
	word := func(n uint64) {
		for i := uint(0); i < 64; i += 8 {
			key = append(key, byte(n>>i))
		}
	}
	index := func(i uint32) uint64 {
		if i >= start && uint64(i) < uint64(start)+uint64(count) {
			return uint64(i - start)
		}
		return uint64(ids[i]) | 1<<32
	}
	typ := func(t wasm.ValueType) {
		if i, ok := t.TypeIndex(); ok && !t.Canonical() {
			if i >= start && uint64(i) < uint64(start)+uint64(count) {
				t = t.WithTypeIndex(i-start, false)
			} else {
				t = t.WithTypeIndex(ids[i], true)
			}
		}
		word(uint64(t))
	}
	word(uint64(count))
	for i := start; i < start+count; i++ {
		d := m.TypeDef(i)
		word(uint64(d.Kind))
		if d.Open {
			word(1)
		} else {
			word(0)
		}
		word(uint64(len(d.Supers)))
		for _, s := range d.Supers {
			word(index(s))
		}
		f := m.Types[i]
		word(uint64(len(f.Params)))
		for _, v := range f.Params {
			typ(v)
		}
		word(uint64(len(f.Results)))
		for _, v := range f.Results {
			typ(v)
		}
		word(uint64(len(d.Fields)))
		for _, f := range d.Fields {
			typ(f.Value)
			word(uint64(f.Packed))
			if f.Mutable {
				word(1)
			} else {
				word(0)
			}
		}
	}
	if base, ok := r.Keys[string(key)]; ok {
		for i := uint32(0); i < count; i++ {
			ids[start+i] = base + i
		}
		return true
	}
	if uint64(r.Count())+uint64(count) > uint64(maximum) {
		return false
	}
	base := uint32(len(r.Types))
	r.Keys[string(key)] = base
	for i := uint32(0); i < count; i++ {
		ids[start+i] = base + i
	}
	for i := start; i < start+count; i++ {
		d := m.TypeDef(i)
		d.Supers = append([]uint32(nil), d.Supers...)
		for j, s := range d.Supers {
			d.Supers[j] = ids[s]
		}
		d.Fields = append([]FieldType(nil), d.Fields...)
		for j := range d.Fields {
			d.Fields[j].Value = ResolveType(d.Fields[j].Value, ids)
		}
		r.Types = append(r.Types, CanonicalType{d, ResolveSignature(m.Types[i], ids)})
		p := uint32(0)
		if len(d.Supers) != 0 {
			p = d.Supers[0]
		}
		r.Hierarchy.Add(p)
	}
	return true
}
func (r *TypeRegistry) InternModule(m *Module, maximum uint32) ([]uint32, bool) {
	ids := make([]uint32, len(m.Types))
	if m.TypeGroups == nil {
		for i := range m.Types {
			if !r.InternGroup(m, uint32(i), 1, ids, maximum) {
				return nil, false
			}
		}
	} else {
		for _, g := range m.TypeGroups {
			if !r.InternGroup(m, g.Start, g.Count, ids, maximum) {
				return nil, false
			}
		}
	}
	return ids, true
}
func (r *TypeRegistry) Subtype(actual, expected wasm.ValueType) bool {
	return r.Hierarchy.ReferenceSubtype(actual, expected, nil)
}
func (r *TypeRegistry) FunctionSubtype(actual, expected uint32) bool {
	return r.Subtype(wasm.CanonicalFunctionReference(actual, false), wasm.CanonicalFunctionReference(expected, false))
}
