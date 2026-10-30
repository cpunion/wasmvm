package validate

import (
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
)

// relation compares one linking/ascription operation. Fresh abstract binders
// may be substituted, while equality-bound aliases impose identity constraints.
// Contravariant component imports can bind variables in either direction.
type relation struct {
	pool       *TypePool
	bindings   map[uint64]*Type
	types      map[*Type]*Type
	values     map[*model.Type]*model.Type
	seenTypes  map[typeComparison]bool
	seenValues map[valueComparison]bool
	generation uint64
	symbols    map[*valueBinding]*valueBinding
}
type typeComparison struct {
	a, b       *Type
	generation uint64
}
type valueComparison struct {
	a, b       *model.Type
	generation uint64
}

func (p *TypePool) relation() *relation {
	return &relation{pool: p, bindings: map[uint64]*Type{}, types: map[*Type]*Type{}, values: map[*model.Type]*model.Type{}, symbols: map[*valueBinding]*valueBinding{}, seenTypes: map[typeComparison]bool{}, seenValues: map[valueComparison]bool{}}
}

func (r *relation) id(id uint64) (uint64, error) {
	for n := 0; n <= len(r.bindings); n++ {
		other := r.bindings[id]
		if other == nil || other.Resource == id {
			return id, nil
		}
		if err := r.pool.charge(1); err != nil {
			return 0, err
		}
		id = other.Resource
	}
	return 0, invalid("cyclic resource substitution")
}

func (r *relation) bind(actual, expected *Type) (bool, error) {
	if actual.Kind != model.Resource || expected.Kind != model.Resource {
		return false, nil
	}
	a, err := r.id(actual.Resource)
	if err != nil {
		return false, err
	}
	b, err := r.id(expected.Resource)
	if err != nil {
		return false, err
	}
	if a == b {
		return true, nil
	}
	if r.bindings[expected.Resource] != nil {
		return false, nil
	}
	if err := r.pool.charge(1); err != nil {
		return false, err
	}
	r.bindings[expected.Resource] = actual
	r.generation++
	return true, nil
}

func (r *relation) external(actual, expected Extern) (bool, error) {
	if actual.Sort != expected.Sort {
		return false, nil
	}
	var ok bool
	var err error
	if expected.Sort == model.ValueSort {
		ok, err = r.value(actual.Type.Value, expected.Type.Value)
		if err != nil || !ok {
			return ok, err
		}
		rewrite := r.rewrite()
		a, err := rewrite.symbol(actual.Type.valueBinding)
		if err != nil {
			return false, err
		}
		b, err := rewrite.symbol(expected.Type.valueBinding)
		if err != nil {
			return false, err
		}
		if expected.Fresh && b == expected.Type.valueBinding {
			if a != b {
				r.symbols[b] = a
				r.generation++
			}
			return true, nil
		}
		return r.pool.sameValue(a, b)
	}
	if expected.Sort == model.TypeSort && expected.Fresh {
		ok, err = r.bind(actual.Type, expected.Type)
	} else {
		ok, err = r.typ(actual.Type, expected.Type)
	}
	if err != nil || !ok {
		return ok, err
	}
	if expected.Sort == model.TypeSort {
		r.types[expected.Type] = actual.Type
		if expected.Type.Value != nil {
			r.values[expected.Type.Value] = actual.Type.Value
		}
	}
	return true, nil
}

func (r *relation) typ(actual, expected *Type) (bool, error) {
	if actual == nil || expected == nil {
		return actual == expected, nil
	}
	if actual.Kind != expected.Kind {
		return false, nil
	}
	if actual.owner != r.pool || expected.owner != r.pool {
		return false, invalid("foreign linking type")
	}
	if actual.Kind == model.Resource {
		a, err := r.id(actual.Resource)
		if err != nil {
			return false, err
		}
		b, err := r.id(expected.Resource)
		return a == b, err
	}
	key := typeComparison{actual, expected, r.generation}
	if r.seenTypes[key] {
		return true, nil
	}
	if err := r.pool.charge(1 + uint64(len(actual.Parameters)) + uint64(len(expected.Parameters))); err != nil {
		return false, err
	}
	var ok bool
	var err error
	switch actual.Kind {
	case model.Instance, model.Component:
		if actual.Kind == model.Component {
			// A subtype requires fewer imports; its requirements must accept
			// each corresponding value promised by the supertype's caller.
			ok, err = r.members(expected.Imports, actual.Imports)
			if err != nil || !ok {
				return ok, err
			}
		}
		ok, err = r.members(actual.Exports, expected.Exports)
	case model.Function, model.AsyncFunction:
		if len(actual.Parameters) != len(expected.Parameters) {
			return false, nil
		}
		for i, p := range actual.Parameters {
			if p.Name != expected.Parameters[i].Name {
				return false, nil
			}
			ok, err = r.value(p.Type, expected.Parameters[i].Type)
			if err != nil || !ok {
				return ok, err
			}
		}
		ok, err = r.value(actual.Result, expected.Result)
	default:
		if actual.CoreModule != nil && expected.CoreModule != nil {
			ok, err = r.coreModule(actual.CoreModule, expected.CoreModule)
		} else {
			ok, err = r.value(actual.Value, expected.Value)
		}
	}
	if ok && err == nil {
		r.seenTypes[key] = true
	}
	return ok, err
}

func (r *relation) members(actual, expected []Extern) (bool, error) {
	if err := r.pool.charge(uint64(len(actual)) + uint64(len(expected))); err != nil {
		return false, err
	}
	byName := make(map[string]Extern, len(actual))
	for _, e := range actual {
		byName[e.Name.Text] = e
	}
	for _, e := range expected {
		a, ok := byName[e.Name.Text]
		if !ok {
			return false, nil
		}
		ok, err := r.external(a, e)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (r *relation) value(a, b *model.Type) (bool, error) {
	if a == nil || b == nil {
		return a == b, nil
	}
	if a.Kind != b.Kind || a.Length != b.Length || len(a.Fields) != len(b.Fields) || len(a.Cases) != len(b.Cases) || len(a.Labels) != len(b.Labels) {
		return false, nil
	}
	x, err := r.id(a.ResourceID)
	if err != nil {
		return false, err
	}
	y, err := r.id(b.ResourceID)
	if err != nil {
		return false, err
	}
	if x != y {
		return false, nil
	}
	key := valueComparison{a, b, r.generation}
	if r.seenValues[key] {
		return true, nil
	}
	if err := r.pool.charge(1 + uint64(len(a.Fields)) + uint64(len(a.Cases)) + uint64(len(a.Labels))); err != nil {
		return false, err
	}
	for i, n := range a.Labels {
		if n != b.Labels[i] {
			return false, nil
		}
	}
	for _, fs := range [][2][]model.Field{{a.Fields, b.Fields}, {a.Cases, b.Cases}} {
		for i, f := range fs[0] {
			if f.Name != fs[1][i].Name {
				return false, nil
			}
			ok, err := r.value(f.Type, fs[1][i].Type)
			if err != nil || !ok {
				return ok, err
			}
		}
	}
	if ok, err := r.value(a.Element, b.Element); err != nil || !ok {
		return ok, err
	}
	if ok, err := r.value(a.Key, b.Key); err != nil || !ok {
		return ok, err
	}
	r.seenValues[key] = true
	return true, nil
}

func (r *relation) coreModule(actual, expected *CoreModuleType) (bool, error) {
	if err := r.pool.charge(uint64(len(actual.Imports)) + uint64(len(expected.Imports)) + uint64(len(actual.Exports)) + uint64(len(expected.Exports))); err != nil {
		return false, err
	}
	imports := map[[2]string]module.Import{}
	for _, e := range expected.Imports {
		imports[[2]string{e.Module, e.Name}] = e.Descriptor
	}
	for _, a := range actual.Imports {
		e, ok := imports[[2]string{a.Module, a.Name}]
		if !ok || !module.ExternSubtype(&r.pool.registry, e, a.Descriptor) {
			return false, nil
		}
	}
	exports := map[string]module.Import{}
	for _, a := range actual.Exports {
		exports[a.Name] = a.Descriptor
	}
	for _, e := range expected.Exports {
		a, ok := exports[e.Name]
		if !ok || !module.ExternSubtype(&r.pool.registry, a, e.Descriptor) {
			return false, nil
		}
	}
	return true, nil
}

func (r *relation) rewrite() *substitution {
	return &substitution{pool: r.pool, bindings: r.bindings, symbols: r.symbols, typeBindings: r.types, valueBindings: r.values, types: map[*Type]*Type{}, values: map[*model.Type]*model.Type{}, cloneAll: true}
}
