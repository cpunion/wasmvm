package preview3

import (
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
)

// Generated WIT schemas describe values, not instance-bound runtime type views.
// Nominal resource checks are supplied by each provider at binding time.
type witShape struct {
	kind     component.ValueKind
	element  *witShape
	fields   []witField
	labels   []string
	resource string
}
type witField struct {
	name string
	typ  *witShape
}
type witFunction struct {
	async      bool
	parameters []*witShape
	result     *witShape
}

func (s *witShape) matches(t component.ValueType, resources map[string]*component.HostResourceType) bool {
	if s == nil {
		return !t.Valid()
	}
	if !t.Valid() || t.Kind() != s.kind {
		return false
	}
	switch s.kind {
	case component.KindOwn, component.KindBorrow:
		return t.IsResource(resources[s.resource])
	case component.KindList, component.KindFuture, component.KindStream, component.KindOption:
		return s.element.matches(t.Element(), resources)
	case component.KindRecord, component.KindTuple:
		if t.FieldCount() != len(s.fields) {
			return false
		}
		for i, f := range s.fields {
			name, field := t.Field(i)
			if s.kind == component.KindRecord && name != f.name || !f.typ.matches(field, resources) {
				return false
			}
		}
	case component.KindResult, component.KindVariant:
		if t.CaseCount() != len(s.fields) {
			return false
		}
		for i, f := range s.fields {
			name, field := t.Case(i)
			if s.kind == component.KindVariant && name != f.name || !f.typ.matches(field, resources) {
				return false
			}
		}
	case component.KindFlags, component.KindEnum:
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
func (f *witFunction) check(resources map[string]*component.HostResourceType) component.FunctionCheck {
	return func(info component.FunctionInfo) error {
		if info.Async() != f.async || info.ParameterCount() != len(f.parameters) || !f.result.matches(info.Result(), resources) {
			return vm.ErrType
		}
		for i, expected := range f.parameters {
			_, actual := info.Parameter(i)
			if !expected.matches(actual, resources) {
				return vm.ErrType
			}
		}
		return nil
	}
}
