package validate

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func bindingType(t *testing.T) *Type {
	t.Helper()
	input, err := os.ReadFile("../../../pkg/component/testdata/host-module-rich.wasm")
	if err != nil {
		t.Fatal(err)
	}
	d, err := decode.Decode(input, decode.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	p, err := Check(d, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return p.Type.Imports[0].Type
}

func TestCoreBindingSessionBoundsAndSharing(t *testing.T) {
	typ := bindingType(t)
	session, err := typ.CoreBindings()
	if err != nil {
		t.Fatal(err)
	}
	// The rich module requires the Core 3.0 profile for memory64.
	cfg := vm.DefaultEngineConfig()
	cfg.Profile = "core3"
	core, err := vm.NewEngineWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer core.Close()
	input, err := os.ReadFile("../../../pkg/component/testdata/host-core-rich.wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := core.Compile(nil, input)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	schema, err := m.Interface()
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := session.Check(typ, schema); !ok || err != nil {
		t.Fatal(ok, err)
	}
	nodes, count := session.pool.nodes, session.pool.registry.Count()
	if ok, err := session.Check(typ, schema); !ok || err != nil || session.pool.nodes != nodes {
		t.Fatal("repeated match expanded metadata", ok, err)
	}
	alias := *typ
	if ok, err := session.Check(&alias, schema); !ok || err != nil || session.pool.registry.Count() != count || len(session.actual) != 1 {
		t.Fatal("actual schema was not reused", ok, err)
	}
	for _, other := range []*Type{nil, {}, bindingType(t)} {
		if ok, err := session.Check(other, schema); ok || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(ok, err)
		}
	}
	if ok, err := session.Check(typ, module.Interface{}); ok || !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(ok, err)
	}
	var zero *CoreBindings
	if ok, err := zero.Check(typ, schema); ok || !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(ok, err)
	}
	for _, other := range []*Type{nil, {}} {
		if s, err := other.CoreBindings(); s != nil || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(s, err)
		}
	}
	// Independent linking attempts consume their own bounded work budget.
	session.pool.limits.MaxNodes = session.pool.nodes
	copy := *typ
	if ok, err := session.Check(&copy, schema); ok || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(ok, err)
	}
}
