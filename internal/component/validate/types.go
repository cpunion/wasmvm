// Package validate checks component static semantics without instantiating or
// executing Core modules. Resolved graphs are immutable and retain sharing.
package validate

import (
	"fmt"
	"math"
	"strings"

	"github.com/cpunion/wasmvm/internal/component/canonical"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type Limits struct {
	MaxNodes, MaxDepth uint32
}

func DefaultLimits() Limits { return Limits{1 << 20, 128} }

// Type describes a validated value, function or resource type. Resource IDs
// are compilation identities, not runtime handles; instantiation must freshen
// resource binders before exposing the resulting instance's exports.
type Type struct {
	Kind             model.Kind
	Value            *model.Type
	Parameters       []model.Field
	Result           *model.Type
	Resource         uint64
	Representation   wasm.ValueType
	Destructor       *uint32
	Imports, Exports []Extern
	CoreModule       *CoreModuleType
	Captures         []uint64
	owner            *TypePool
	borrow           bool
	resource         bool
	depth            uint32
	minimumResource  uint64
	definingScope    *TypeScope
	valueBinding     *valueBinding
}

type traits struct {
	borrow, resource bool
	depth            uint32
	minimumResource  uint64
	resourceName     *Type
}

// TypePool bounds total metadata across lexical scopes. It caches primitive
// values and layouts, so repeated aliases never expand shared value graphs.
// A pool belongs to one component compilation and is not concurrent mutable.
type TypePool struct {
	limits     Limits
	nodes      uint32
	resources  uint64
	values     map[*model.Type]traits
	primitives map[model.Kind]*Type
	layouts    *canonical.Builder
	layouts32  *canonical.Builder
	registry   module.TypeRegistry
}

func NewTypePool(limits Limits) *TypePool {
	return &TypePool{limits: limits, values: map[*model.Type]traits{}, primitives: map[model.Kind]*Type{},
		layouts: canonical.NewBuilder(true, canonical.Limits{MaxNodes: limits.MaxNodes, MaxDepth: limits.MaxDepth, MaxBytes: math.MaxUint64})}
}

// TypeScope is an ordered type index space. Component boundaries are distinct
// from component/instance type declarations, which may capture outer resource
// types but must not introduce concrete resource definitions.
type TypeScope struct {
	pool             *TypePool
	parent           *TypeScope
	types            []*Type
	component        bool
	declarator       bool
	depth            uint32
	items            map[model.Sort][]*Type
	imports, exports nameSet
	core             *coreScope
	visibility       visibility
	deferVisibility  bool
}

func (p *TypePool) Scope(parent *TypeScope, component bool) (*TypeScope, error) {
	depth := uint32(1)
	if parent != nil {
		if parent.pool != p {
			return nil, invalid("foreign lexical scope")
		}
		depth += parent.depth
	}
	if depth > p.limits.MaxDepth {
		return nil, limit("lexical nesting")
	}
	if err := p.charge(1); err != nil {
		return nil, err
	}
	s := &TypeScope{pool: p, parent: parent, component: component, declarator: !component, depth: depth, items: map[model.Sort][]*Type{}}
	s.core = &coreScope{pool: p, types: new(module.Module), modules: map[uint32]*CoreModuleType{}}
	if parent != nil {
		s.core.parent = parent.core
	}
	return s, nil
}

func invalid(detail string) error { return fmt.Errorf("%w: component %s", wasm.ErrInvalid, detail) }
func limit(detail string) error   { return fmt.Errorf("%w: component %s", wasm.ErrLimit, detail) }

func (p *TypePool) charge(n uint64) error {
	if n > uint64(p.limits.MaxNodes-p.nodes) {
		return limit("type metadata")
	}
	p.nodes += uint32(n)
	return nil
}

func (s *TypeScope) At(index uint32) (*Type, error) {
	if uint64(index) >= uint64(len(s.types)) {
		return nil, invalid("type index out of bounds")
	}
	return s.types[index], nil
}

func (s *TypeScope) bind(t *Type) (*Type, error) {
	if err := s.pool.charge(1); err != nil {
		return nil, err
	}
	s.types = append(s.types, t)
	return t, nil
}

// Abstract introduces a fresh nominal resource binder, as used by a sub
// resource import/export. Equality bounds instead retain the existing identity.
func (s *TypeScope) Abstract() (*Type, error) {
	t, err := s.pool.abstract()
	if err != nil {
		return nil, err
	}
	return s.bind(t)
}

func (p *TypePool) abstract() (*Type, error) {
	if err := p.charge(1); err != nil {
		return nil, err
	}
	p.resources++
	return &Type{Kind: model.Resource, Resource: p.resources, resource: true, owner: p, minimumResource: p.resources}, nil
}

func (s *TypeScope) Alias(t *Type) (*Type, error) {
	if t == nil || t.owner != s.pool {
		return nil, invalid("foreign type alias")
	}
	return s.bind(t)
}

func (s *TypeScope) Outer(count, index uint32) (*Type, error) {
	at, crossed := s, false
	for i := uint32(0); i < count; i++ {
		crossed = crossed || at.component
		at = at.parent
		if at == nil {
			return nil, invalid("outer scope out of bounds")
		}
	}
	t, err := at.At(index)
	if err != nil {
		return nil, err
	}
	if crossed && t.resource {
		return nil, invalid("outer alias captures resource across component boundary")
	}
	if at != s && !crossed {
		if err := s.inheritNames(at); err != nil {
			return nil, err
		}
	}
	return s.Alias(t)
}

func primitive(kind model.Kind) bool {
	return kind >= model.String && kind <= model.Bool || kind == model.ErrorContext
}

func (s *TypeScope) Value(ref model.TypeRef) (*model.Type, error) {
	if ref.Indexed {
		t, err := s.At(ref.Index)
		if err != nil {
			return nil, err
		}
		if t.Value == nil {
			return nil, invalid("type index is not a value type")
		}
		return t.Value, nil
	}
	if !primitive(ref.Primitive) {
		return nil, invalid("unknown primitive value type")
	}
	if t := s.pool.primitives[ref.Primitive]; t != nil {
		return t.Value, nil
	}
	t, err := s.resolve(&model.TypeDef{Kind: ref.Primitive})
	if err != nil {
		return nil, err
	}
	s.pool.primitives[ref.Primitive] = t
	return t.Value, nil
}

func (s *TypeScope) optional(ref *model.TypeRef) (*model.Type, error) {
	if ref == nil {
		return nil, nil
	}
	return s.Value(*ref)
}

// Define appends a type only after its entire expression has validated. The
// caller supplies the already validated, ordered Core function signatures for
// checking a resource destructor. No destructor is executed during validation.
func (s *TypeScope) Define(def *model.TypeDef, functions []module.FuncType) (*Type, error) {
	if def == nil {
		return nil, invalid("missing type definition")
	}
	if def.Kind == model.Resource {
		if s.declarator {
			return nil, invalid("concrete resource definition in type declaration")
		}
		if def.Representation != wasm.I32 && def.Representation != wasm.I64 {
			return nil, invalid("resource representation must be i32 or i64")
		}
		if def.Destructor != nil {
			if uint64(*def.Destructor) >= uint64(len(functions)) {
				return nil, invalid("resource destructor index out of bounds")
			}
			f := functions[*def.Destructor]
			if len(f.Params) != 1 || f.Params[0] != def.Representation || len(f.Results) != 0 {
				return nil, invalid("resource destructor signature")
			}
		}
		t, err := s.Abstract()
		if err != nil {
			return nil, err
		}
		t.Representation = def.Representation
		t.definingScope = s
		if def.Destructor != nil {
			index := *def.Destructor
			t.Destructor = &index
		}
		return t, nil
	}
	t, err := s.resolve(def)
	if err != nil {
		return nil, err
	}
	return s.bind(t)
}

func (s *TypeScope) resolve(def *model.TypeDef) (*Type, error) {
	if err := s.pool.charge(1); err != nil {
		return nil, err
	}
	t := &Type{Kind: def.Kind, owner: s.pool, depth: 1}
	v := &model.Type{Kind: def.Kind}
	var err error
	var resourceName *Type
	include := func(child *model.Type) {
		if child != nil {
			info := s.pool.values[child]
			t.borrow, t.resource = t.borrow || info.borrow, t.resource || info.resource
			t.depth = max(t.depth, info.depth+1)
			t.minimumResource = minimum(t.minimumResource, info.minimumResource)
		}
	}
	resolve := func(ref *model.TypeRef) *model.Type {
		if err != nil {
			return nil
		}
		var child *model.Type
		child, err = s.optional(ref)
		include(child)
		return child
	}
	fields := func(in []model.NamedType, labels bool) []model.Field {
		if err = s.pool.charge(uint64(len(in))); err != nil {
			return nil
		}
		out := make([]model.Field, len(in))
		seen := map[string]bool{}
		for i, f := range in {
			if labels && err == nil {
				err = label(f.Name, seen)
			}
			out[i] = model.Field{Name: f.Name, Type: resolve(&f.Type)}
		}
		return out
	}
	switch def.Kind {
	case model.Bool, model.U8, model.S8, model.U16, model.S16, model.U32, model.S32, model.U64, model.S64, model.F32, model.F64, model.Char, model.String, model.ErrorContext:
	case model.Record:
		if len(def.Fields) == 0 {
			return nil, invalid("empty record")
		}
		v.Fields = fields(def.Fields, true)
	case model.Tuple:
		if len(def.Tuple) == 0 {
			return nil, invalid("empty tuple")
		}
		if err = s.pool.charge(uint64(len(def.Tuple))); err != nil {
			return nil, err
		}
		v.Fields = make([]model.Field, len(def.Tuple))
		for i := range def.Tuple {
			v.Fields[i].Type = resolve(&def.Tuple[i])
		}
	case model.Variant, model.Result:
		if len(def.Cases) == 0 || def.Kind == model.Result && len(def.Cases) != 2 {
			return nil, invalid("variant/result case count")
		}
		if err = s.pool.charge(uint64(len(def.Cases))); err != nil {
			return nil, err
		}
		v.Cases = make([]model.Field, len(def.Cases))
		seen := map[string]bool{}
		for i, c := range def.Cases {
			if def.Kind == model.Variant && err == nil {
				err = label(c.Name, seen)
			}
			v.Cases[i] = model.Field{Name: c.Name, Type: resolve(c.Type)}
		}
	case model.Flags, model.Enum:
		if len(def.Labels) == 0 || def.Kind == model.Flags && len(def.Labels) > 32 {
			return nil, invalid("flags/enum label count")
		}
		if err = s.pool.charge(uint64(len(def.Labels))); err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, name := range def.Labels {
			if err = label(name, seen); err != nil {
				return nil, err
			}
		}
		v.Labels = append([]string(nil), def.Labels...)
	case model.List, model.FixedList, model.Option, model.Map:
		if def.Element == nil || def.Kind == model.Map && def.Key == nil {
			return nil, invalid("missing container element type")
		}
		if def.Kind == model.FixedList && def.Length == 0 {
			return nil, invalid("empty fixed list")
		}
		v.Element, v.Length = resolve(def.Element), def.Length
		if def.Kind == model.Map {
			v.Key = resolve(def.Key)
			if err == nil && !mapKey(v.Key.Kind) {
				return nil, invalid("map key type")
			}
		}
	case model.Stream, model.Future:
		v.Element = resolve(def.Element)
		if t.borrow {
			return nil, invalid("async element contains borrow")
		}
		if v.Element != nil && def.Kind == model.Stream && v.Element.Kind == model.Char {
			return nil, invalid("stream of char")
		}
	case model.Own, model.Borrow:
		r, e := s.At(def.Resource)
		if e != nil {
			return nil, e
		}
		if r.Kind != model.Resource {
			return nil, invalid("handle type index is not a resource")
		}
		v.ResourceID, t.resource, t.borrow = r.Resource, true, def.Kind == model.Borrow
		t.minimumResource = r.Resource
		resourceName = r
	case model.Function, model.AsyncFunction:
		t.Parameters, t.Result = fields(def.Parameters, true), resolve(def.Result)
		if err == nil && t.Result != nil && s.pool.values[t.Result].borrow {
			return nil, invalid("function result contains borrow")
		}
		v = nil
	case model.Component, model.Instance:
		return s.interfaceType(def)
	default:
		// Component/instance declaration checking is supplied by the lexical
		// definition validator; accepting an unvisited body would be unsound.
		return nil, fmt.Errorf("%w: component type constructor %#x", wasm.ErrUnsupported, def.Kind)
	}
	if err != nil {
		return nil, err
	}
	if t.depth > s.pool.limits.MaxDepth {
		return nil, limit("type nesting")
	}
	if v != nil {
		layout, err := s.pool.layouts.Build(v)
		if err != nil {
			return nil, err
		}
		if layout.Size >= 1<<28 {
			return nil, invalid("value type is at least 2^28 bytes in the memory64 ABI")
		}
		t.Value = v
		s.pool.values[v] = traits{borrow: t.borrow, resource: t.resource, depth: t.depth, minimumResource: t.minimumResource, resourceName: resourceName}
	}
	return t, nil
}

func minimum(a, b uint64) uint64 {
	if a == 0 || b != 0 && b < a {
		return b
	}
	return a
}

func mapKey(k model.Kind) bool {
	return k == model.Bool || k == model.Char || k == model.String || k >= model.U64 && k <= model.S8
}

// label validates ASCII kebab case and strong uniqueness, including acronym
// case and hyphen collisions that language binding generators cannot preserve.
func label(name string, seen map[string]bool) error {
	if name == "" {
		return invalid("empty label")
	}
	lower, upper, fragment := false, false, false
	for i := 0; i < len(name); i++ {
		b := name[i]
		switch {
		case b == '-':
			if !fragment {
				return invalid("label is not kebab case")
			}
			lower, upper, fragment = false, false, false
			continue
		case b >= 'a' && b <= 'z':
			lower = true
		case b >= 'A' && b <= 'Z':
			upper = true
		case b >= '0' && b <= '9' && i != 0:
		default:
			return invalid("label is not kebab case")
		}
		fragment = true
		if lower && upper {
			return invalid("label is not kebab case")
		}
	}
	if !fragment {
		return invalid("label is not kebab case")
	}
	key := strings.ReplaceAll(strings.ToLower(name), "-", "")
	if seen[key] {
		return invalid("conflicting label")
	}
	seen[key] = true
	return nil
}

// ContainsResource reports nominal resource occurrences in this checked type.
func (t *Type) ContainsResource() bool { return t.resource }

// ResultContainsResource reports the cached traits of a function's result.
func (t *Type) ResultContainsResource() bool {
	return t.Result != nil && t.owner.values[t.Result].resource
}
