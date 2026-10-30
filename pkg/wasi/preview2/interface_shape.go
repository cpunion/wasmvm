package preview2

import (
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// interfaceShape is a provider-owned immutable contract. Matching follows the
// finite expected graph, never an unbounded guest type expansion.
type interfaceShape struct {
	kind     component.ValueKind
	resource *component.HostResourceType
	element  *interfaceShape
	fields   []shapeField
	labels   []string
}
type shapeField struct {
	name string
	typ  *interfaceShape
}

func scalarShape(kind component.ValueKind) *interfaceShape { return &interfaceShape{kind: kind} }
func aggregateShape(kind component.ValueKind, fields ...shapeField) *interfaceShape {
	return &interfaceShape{kind: kind, fields: fields}
}
func elementShape(kind component.ValueKind, element *interfaceShape) *interfaceShape {
	return &interfaceShape{kind: kind, element: element}
}
func resourceShape(kind component.ValueKind, resource *component.HostResourceType) *interfaceShape {
	return &interfaceShape{kind: kind, resource: resource}
}
func (s *interfaceShape) matches(t component.ValueType) bool {
	if s == nil {
		return !t.Valid()
	}
	if t.Kind() != s.kind {
		return false
	}
	if s.resource != nil {
		return t.IsResource(s.resource)
	}
	if s.element != nil && !s.element.matches(t.Element()) {
		return false
	}
	switch s.kind {
	case component.KindRecord, component.KindTuple:
		if t.FieldCount() != len(s.fields) {
			return false
		}
		for i, want := range s.fields {
			name, typ := t.Field(i)
			if s.kind == component.KindRecord && name != want.name || !want.typ.matches(typ) {
				return false
			}
		}
	case component.KindVariant, component.KindResult:
		if t.CaseCount() != len(s.fields) {
			return false
		}
		for i, want := range s.fields {
			name, typ := t.Case(i)
			if s.kind == component.KindVariant && name != want.name || !want.typ.matches(typ) {
				return false
			}
		}
	case component.KindEnum, component.KindFlags:
		if t.LabelCount() != len(s.labels) {
			return false
		}
		for i, label := range s.labels {
			if t.Label(i) != label {
				return false
			}
		}
	}
	return true
}
func shapeContract(parameters []*interfaceShape, result *interfaceShape) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		if f.ParameterCount() != len(parameters) || !result.matches(f.Result()) {
			return vm.ErrType
		}
		for i, want := range parameters {
			_, typ := f.Parameter(i)
			if !want.matches(typ) {
				return vm.ErrType
			}
		}
		return nil
	}
}
