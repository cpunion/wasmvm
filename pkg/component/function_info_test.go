package component

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func TestFunctionContractsPrecedeGuestInitialization(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "resource-start-failure"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	called := 0
	sentinel := errors.New("host contract mismatch")
	imports := Imports{Functions: map[string]HostFunction{"fail": func(*Caller, []Value) (*Value, error) { called++; return nil, nil }}, FunctionChecks: map[string]FunctionCheck{
		"fail": func(f FunctionInfo) error {
			if f.ParameterCount() != 0 || f.Result().Valid() {
				t.Fatal("wrong checked schema")
			}
			return sentinel
		},
	}}
	if in, err := m.Instantiate(nil, imports, CallOptions{}); in != nil || !errors.Is(err, sentinel) || called != 0 {
		t.Fatal("contract ran after guest effects", in, err, called)
	}
	ctx, cancel := context.WithCancel(context.Background())
	imports.FunctionChecks["fail"] = func(FunctionInfo) error { cancel(); return nil }
	if in, err := m.Instantiate(ctx, imports, CallOptions{}); in != nil || !errors.Is(err, context.Canceled) || called != 0 {
		t.Fatal(in, err, called)
	}
}

func TestFunctionViewsRetainNominalAndAggregateInformation(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "ownership-containers"))
	if err != nil {
		t.Fatal(err)
	}
	typ, other := NewHostResourceType(nil), NewHostResourceType(nil)
	var saved FunctionInfo
	check := func(f FunctionInfo) error {
		saved = f
		if f.ParameterCount() != 1 {
			return vm.ErrType
		}
		name, input := f.Parameter(0)
		if name != "input" || input.Kind() != KindTuple || input.FieldCount() != 6 || f.Result().Kind() != KindTuple {
			return vm.ErrType
		}
		_, list := input.Field(0)
		_, mapping := input.Field(1)
		_, option := input.Field(2)
		_, variant := input.Field(3)
		_, fixed := input.Field(4)
		_, record := input.Field(5)
		if list.Kind() != KindList || !list.Element().IsResource(typ) || list.Element().IsResource(other) || list.Element().IsResource(nil) {
			return vm.ErrType
		}
		if mapping.Kind() != KindMap || mapping.Key().Kind() != KindU32 || !mapping.Element().IsResource(typ) {
			return vm.ErrType
		}
		if option.Kind() != KindOption || !option.Element().IsResource(typ) || variant.CaseCount() != 2 {
			return vm.ErrType
		}
		label, empty := variant.Case(0)
		_, payload := variant.Case(1)
		if label != "empty" || empty.Valid() || !payload.IsResource(typ) {
			return vm.ErrType
		}
		if fixed.Kind() != KindFixedList || fixed.Length() != 8 || !fixed.Element().IsResource(typ) {
			return vm.ErrType
		}
		field, value := record.Field(0)
		if record.Kind() != KindRecord || field != "value" || !value.IsResource(typ) {
			return vm.ErrType
		}
		return nil
	}
	in, err := m.Instantiate(nil, Imports{Instances: map[string]Imports{"host": {Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{"echo": func(*Caller, []Value) (*Value, error) { t.Fatal("binding executed callback"); return nil, nil }}, FunctionChecks: map[string]FunctionCheck{"echo": check}}}}, CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	exported, err := in.Function("echo")
	if err != nil {
		t.Fatal(err)
	}
	if err := check(exported.Info()); err != nil {
		t.Fatal("exported signature lost nominal information", err)
	}
	m.Close()
	in.Close()
	_, input := saved.Parameter(0)
	_, list := input.Field(0)
	if !list.Element().IsResource(typ) {
		t.Fatal("retained view lost its immutable bindings")
	}
}

func TestEmptyAndBoundedTypeViews(t *testing.T) {
	var missing *Function
	if missing.Info().Result().Valid() || missing.Info().ParameterCount() != 0 {
		t.Fatal("nil function has a signature")
	}
	var f FunctionInfo
	if f.ParameterCount() != 0 || f.Result().Valid() {
		t.Fatal(f)
	}
	for _, i := range []int{-1, 0, 99} {
		if name, typ := f.Parameter(i); name != "" || typ.Valid() {
			t.Fatal(name, typ)
		}
	}
	var zero ValueType
	if zero.Valid() || zero.Kind() != 0 || zero.Element().Valid() || zero.Key().Valid() || zero.Length() != 0 || zero.FieldCount() != 0 || zero.CaseCount() != 0 || zero.LabelCount() != 0 {
		t.Fatal(zero)
	}
	for _, i := range []int{-1, 0, 99} {
		if name, typ := zero.Field(i); name != "" || typ.Valid() {
			t.Fatal(name, typ)
		}
		if name, typ := zero.Case(i); name != "" || typ.Valid() {
			t.Fatal(name, typ)
		}
		if zero.Label(i) != "" {
			t.Fatal("invalid label")
		}
	}
	labels := ValueType{typ: &model.Type{Kind: model.Enum, Labels: []string{"first", "second"}}}
	if labels.LabelCount() != 2 || labels.Label(1) != "second" || labels.IsResource(NewHostResourceType(nil)) {
		t.Fatal(labels)
	}
	// A schema without a bound nominal declaration cannot acquire authority.
	own := ValueType{typ: &model.Type{Kind: model.Own, ResourceID: 1}}
	if own.IsResource(NewHostResourceType(nil)) || own.IsResource(new(HostResourceType)) {
		t.Fatal("unbound nominal type matched")
	}
}

func TestHostContractRejectsOwnInPlaceOfBorrow(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	typ := NewHostResourceType(nil)
	called := 0
	imports := Imports{Resources: map[string]*HostResourceType{"r": typ}, Functions: map[string]HostFunction{}, FunctionChecks: map[string]FunctionCheck{
		"get": func(f FunctionInfo) error {
			_, p := f.Parameter(0)
			if f.ParameterCount() != 1 || p.Kind() != KindBorrow || !p.IsResource(typ) {
				return vm.ErrType
			}
			return nil
		},
	}}
	for _, name := range []string{"get", "consume", "create"} {
		imports.Functions[name] = func(*Caller, []Value) (*Value, error) { called++; return &Value{}, nil }
	}
	for _, name := range []string{"ownership-host", "function-check-own"} {
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(err)
		}
		in, err := m.Instantiate(nil, imports, CallOptions{})
		if name == "ownership-host" {
			if err != nil {
				t.Fatal(err)
			}
			in.Close()
		} else if in != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal("lifetime mismatch linked", in, err)
		}
		m.Close()
	}
	if called != 0 {
		t.Fatal("signature checking executed callbacks", called)
	}
}
