package module

import (
	"encoding/binary"
	"sort"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Import interns a validated immutable registry in this registry's identity
// space. The returned mapping includes the reserved zero identity. As with
// InternModule, a failed import may retain groups within the supplied limit.
func (r *TypeRegistry) Import(source *TypeRegistry, maximum uint32) ([]uint32, bool) {
	if r == nil || source == nil || uint64(source.Count()) > uint64(maximum) {
		return nil, false
	}
	n := source.Count()
	r.Init()
	if n == 0 {
		return []uint32{0}, true
	}
	groups := make([]TypeGroup, 0, len(source.Keys))
	for key, base := range source.Keys {
		if len(key) < 8 || base == 0 {
			return nil, false
		}
		count := binary.LittleEndian.Uint64([]byte(key[:8]))
		if count == 0 || uint64(base)+count > uint64(n)+1 {
			return nil, false
		}
		groups = append(groups, TypeGroup{Start: base - 1, Count: uint32(count)})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Start < groups[j].Start })
	cursor := uint32(0)
	for _, g := range groups {
		if g.Start != cursor {
			return nil, false
		}
		cursor += g.Count
	}
	if uint64(cursor) != uint64(n) {
		return nil, false
	}
	raw := func(t wasm.ValueType) wasm.ValueType {
		if id, ok := t.TypeIndex(); ok {
			return t.WithTypeIndex(id-1, false)
		}
		return t
	}
	m := &Module{Types: make([]FuncType, n), TypeDefs: make([]TypeDef, n), TypeGroups: groups}
	for i, t := range source.Types[1:] {
		f := FuncType{Params: make([]wasm.ValueType, len(t.Function.Params)), Results: make([]wasm.ValueType, len(t.Function.Results))}
		for j, v := range t.Function.Params {
			f.Params[j] = raw(v)
		}
		for j, v := range t.Function.Results {
			f.Results[j] = raw(v)
		}
		d := t.Definition
		d.Supers = append([]uint32(nil), d.Supers...)
		d.Fields = append([]FieldType(nil), d.Fields...)
		for j, id := range d.Supers {
			d.Supers[j] = id - 1
		}
		for j := range d.Fields {
			d.Fields[j].Value = raw(d.Fields[j].Value)
		}
		m.Types[i], m.TypeDefs[i] = f, d
	}
	ids, ok := r.InternModule(m, maximum)
	if !ok {
		return nil, false
	}
	return append([]uint32{0}, ids...), true
}
