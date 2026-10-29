package validate

import (
	"fmt"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Extern is a named component type declaration. Attribute values are retained
// as immutable metadata, but do not participate in structural type comparison.
type Extern struct {
	Name  model.Name
	Sort  model.Sort
	Type  *Type
	Fresh bool
}

func (s *TypeScope) interfaceType(def *model.TypeDef) (*Type, error) {
	body, err := s.pool.Scope(s, false)
	if err != nil {
		return nil, err
	}
	firstResource := s.pool.resources + 1
	if err := s.pool.charge(uint64(len(def.Declarations))); err != nil {
		return nil, err
	}
	for _, d := range def.Declarations {
		switch d.Kind {
		case 0:
			err = body.DefineCore(d.CoreType, def.CoreTypes)
		case 1:
			_, err = body.Define(d.Type, nil)
		case 2:
			err = body.aliasDeclaration(d.Alias)
		case 3:
			if def.Kind != model.Component {
				return nil, invalid("import in instance type")
			}
			_, err = body.Declare(d.Import, false)
		case 4:
			_, err = body.Declare(d.Export, true)
		default:
			return nil, invalid("unknown type declaration")
		}
		if err != nil {
			return nil, err
		}
	}
	t := &Type{Kind: def.Kind, owner: s.pool, Imports: body.imports.ordered, Exports: body.exports.ordered, depth: 1}
	for _, group := range [][]Extern{t.Imports, t.Exports} {
		for _, e := range group {
			t.depth = max(t.depth, e.Type.depth+1)
			// Locally bound resources are alpha-bound by this interface rather
			// than captured from its lexical environment.
			if id := e.Type.minimumResource; id < firstResource {
				t.minimumResource = minimum(t.minimumResource, id)
			}
		}
	}
	t.Captures, err = s.pool.captures(t, firstResource)
	if err != nil {
		return nil, err
	}
	t.minimumResource = 0
	if len(t.Captures) != 0 {
		t.minimumResource = t.Captures[0]
	}
	t.resource = t.minimumResource != 0
	if t.depth > s.pool.limits.MaxDepth {
		return nil, limit("interface type nesting")
	}
	return t, nil
}

// Reference resolves the sort ascription of an external declaration. It does
// not introduce a new index; Declare performs that operation after checking the
// name. Equality-bound type declarations preserve their existing nominal IDs.
func (s *TypeScope) Reference(ref model.ExternType) (*Type, error) {
	if ref.Sort == model.TypeSort && ref.Bound == 1 {
		return nil, invalid("fresh resource bound requires a declaration")
	}
	if ref.Sort == model.CoreModuleSort {
		m := s.core.modules[ref.Index]
		if m == nil {
			return nil, invalid("Core type index is not a module type")
		}
		return &Type{Kind: 0x50, CoreModule: m, owner: s.pool, depth: 1}, nil
	}
	if ref.Sort == model.ValueSort {
		return nil, fmt.Errorf("%w: component external sort %#x", wasm.ErrUnsupported, ref.Sort)
	}
	t, err := s.At(ref.Index)
	if err != nil {
		return nil, err
	}
	valid := ref.Sort == model.TypeSort && ref.Bound == 0 ||
		ref.Sort == model.FunctionSort && (t.Kind == model.Function || t.Kind == model.AsyncFunction) ||
		ref.Sort == model.ComponentSort && t.Kind == model.Component ||
		ref.Sort == model.InstanceSort && t.Kind == model.Instance
	if !valid {
		return nil, invalid("external type sort mismatch")
	}
	return t, nil
}

// Declare validates an import/export type and appends its introduced index.
// It is also used for component/instance type declarators, whose namespaces
// start empty. A failure aborts the compilation pool; allocation budgets and
// nominal IDs are intentionally not rolled back for speculative retries.
func (s *TypeScope) Declare(d *model.Import, exporting bool) (*Type, error) {
	if d == nil {
		return nil, invalid("missing external declaration")
	}
	var t *Type
	var err error
	fresh := d.Type.Sort == model.TypeSort && d.Type.Bound == 1
	if fresh {
		t, err = s.Abstract()
	} else {
		t, err = s.Reference(d.Type)
	}
	if err != nil {
		return nil, err
	}
	if d.Type.Sort == model.InstanceSort {
		t, err = s.pool.freshInstance(t)
		if err != nil {
			return nil, err
		}
	}
	names := &s.imports
	if exporting {
		names = &s.exports
	}
	if err = names.add(d.Name, d.Type.Sort, t, s.pool); err != nil {
		return nil, err
	}
	if fresh {
		i := len(names.ordered) - 1
		names.ordered[i].Fresh = true
		names.raw[d.Name.Text] = names.ordered[i]
	}
	if !fresh {
		if d.Type.Sort == model.TypeSort {
			_, err = s.Alias(t)
		} else if err = s.pool.charge(1); err == nil {
			s.items[d.Type.Sort] = append(s.items[d.Type.Sort], t)
		}
	}
	return t, err
}

func (s *TypeScope) aliasDeclaration(a *model.Alias) error {
	if a == nil {
		return invalid("missing type alias")
	}
	if a.Kind == 2 {
		if a.Sort == model.CoreTypeSort {
			return s.core.alias(a.Count, a.Index, true)
		}
		if a.Sort != model.TypeSort {
			return invalid("outer alias sort in type declaration")
		}
		_, err := s.Outer(a.Count, a.Index)
		return err
	}
	if a.Kind != 0 || a.Sort != model.TypeSort && a.Sort != model.InstanceSort {
		return invalid("export alias sort in type declaration")
	}
	instances := s.items[model.InstanceSort]
	if uint64(a.Instance) >= uint64(len(instances)) {
		return invalid("instance alias index out of bounds")
	}
	for _, e := range instances[a.Instance].Exports {
		if e.Name.Text != a.Name {
			continue
		}
		if e.Sort != a.Sort {
			return invalid("instance export alias sort mismatch")
		}
		if a.Sort == model.TypeSort {
			_, err := s.Alias(e.Type)
			return err
		}
		if err := s.pool.charge(1); err != nil {
			return err
		}
		s.items[model.InstanceSort] = append(instances, e.Type)
		return nil
	}
	return invalid("instance export alias name not found")
}
