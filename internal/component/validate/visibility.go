package validate

import "github.com/cpunion/wasmvm/internal/component/model"

// Nameability is stricter than structural equality. Import names are available
// to both boundaries; an exported alias is available only to later exports.
// Separate pointer identities preserve the original private binding.
type visibility struct {
	imports, exports map[nameable]bool
}
type nameable struct {
	value    *model.Type
	resource *Type
}

func (s *TypeScope) inheritNames(outer *TypeScope) error {
	if err := s.pool.charge(uint64(len(outer.visibility.imports)) + uint64(len(outer.visibility.exports))); err != nil {
		return err
	}
	for _, pair := range []struct {
		from map[nameable]bool
		to   *map[nameable]bool
	}{{outer.visibility.imports, &s.visibility.imports}, {outer.visibility.exports, &s.visibility.exports}} {
		if *pair.to == nil {
			*pair.to = map[nameable]bool{}
		}
		for key, value := range pair.from {
			(*pair.to)[key] = value
		}
	}
	return nil
}

func named(kind model.Kind) bool {
	return kind == model.Record || kind == model.Variant || kind == model.Enum || kind == model.Flags
}

func (s *TypeScope) checkVisible(t *Type, exporting, root bool) error {
	if err := s.pool.charge(1 + uint64(len(t.Parameters))); err != nil {
		return err
	}
	known := s.visibility.imports
	if exporting {
		known = s.visibility.exports
	}
	if root && !exporting && t.Kind == model.Resource && t.definingScope != nil {
		return invalid("local resource type found in imports")
	}
	seen := map[*model.Type]bool{}
	var value func(*model.Type, bool) error
	value = func(v *model.Type, root bool) error {
		if v == nil {
			return nil
		}
		if named(v.Kind) && !root {
			if !known[nameable{value: v}] {
				return invalid("value type is not externally nameable")
			}
			return nil
		}
		if seen[v] {
			return nil
		}
		if err := s.pool.charge(1 + uint64(len(v.Fields)) + uint64(len(v.Cases))); err != nil {
			return err
		}
		seen[v] = true
		if v.Kind == model.Own || v.Kind == model.Borrow {
			if !known[nameable{resource: s.pool.values[v].resourceName}] {
				return invalid("resource type is not externally nameable")
			}
		}
		for _, fs := range [][]model.Field{v.Fields, v.Cases} {
			for _, f := range fs {
				if err := value(f.Type, false); err != nil {
					return err
				}
			}
		}
		if err := value(v.Element, false); err != nil {
			return err
		}
		return value(v.Key, false)
	}
	if t.Kind == model.Resource {
		if !root && !known[nameable{resource: t}] {
			return invalid("resource type is not externally nameable")
		}
		return nil
	}
	// Component type declarations validate eagerly; instance types defer
	// external nameability until attached to an import or export.
	if t.Kind == model.Component {
		return nil
	}
	if t.Kind == model.Instance {
		if err := s.pool.charge(uint64(len(known)) + uint64(len(t.Exports))); err != nil {
			return err
		}
		visible := make(map[nameable]bool, len(known))
		for key, v := range known {
			visible[key] = v
		}
		body := &TypeScope{pool: s.pool, visibility: visibility{exports: visible}}
		for _, e := range t.Exports {
			if err := body.checkVisible(e.Type, true, e.Sort == model.TypeSort); err != nil {
				return err
			}
			if e.Sort == model.TypeSort || e.Sort == model.InstanceSort {
				if err := body.expose(e.Type, true); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := value(t.Value, root); err != nil {
		return err
	}
	for _, f := range t.Parameters {
		if err := value(f.Type, false); err != nil {
			return err
		}
	}
	return value(t.Result, false)
}

func (s *TypeScope) publish(t *Type, exporting bool) (*Type, error) {
	if err := s.pool.charge(1); err != nil {
		return nil, err
	}
	out := *t
	if t.Value != nil {
		value := *t.Value
		out.Value = &value
		s.pool.values[&value] = s.pool.values[t.Value]
	}
	if err := s.expose(&out, exporting); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *TypeScope) expose(t *Type, exporting bool) error {
	if s.visibility.exports == nil {
		s.visibility.exports = map[nameable]bool{}
	}
	mark := func(key nameable) {
		s.visibility.exports[key] = true
		if !exporting {
			if s.visibility.imports == nil {
				s.visibility.imports = map[nameable]bool{}
			}
			s.visibility.imports[key] = true
		}
	}
	seen := map[*Type]bool{}
	var visit func(*Type) error
	visit = func(t *Type) error {
		if seen[t] {
			return nil
		}
		if err := s.pool.charge(1 + uint64(len(t.Exports))); err != nil {
			return err
		}
		seen[t] = true
		if t.Kind == model.Resource {
			mark(nameable{resource: t})
		}
		if t.Value != nil && named(t.Value.Kind) {
			mark(nameable{value: t.Value})
		}
		if t.Kind == model.Instance {
			for _, e := range t.Exports {
				if e.Sort == model.TypeSort || e.Sort == model.InstanceSort {
					if err := visit(e.Type); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	return visit(t)
}
