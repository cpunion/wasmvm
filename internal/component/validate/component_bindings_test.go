package validate

import (
	"errors"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
	"github.com/cpunion/wasmvm/internal/component/model"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func bindingSchema(t *testing.T, name string) *Type {
	t.Helper()
	input, err := os.ReadFile("../../../pkg/component/testdata/" + name + ".wasm")
	if err != nil {
		t.Fatal(err)
	}
	d, err := decode.Decode(input, decode.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	c, err := Check(d, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return c.Type
}

func TestComponentBindingsCloneIndependentSchemas(t *testing.T) {
	for _, name := range []string{"ownership-containers", "map", "value-imports", "value-constant-bound", "host-components-core", "type-imports"} {
		t.Run(name, func(t *testing.T) {
			expected, actual := bindingSchema(t, name), bindingSchema(t, name)
			eNodes, aNodes := expected.owner.nodes, actual.owner.nodes
			b, err := expected.ComponentBindings()
			if err != nil {
				t.Fatal(err)
			}
			if ok, err := b.Check(expected, actual); !ok || err != nil {
				t.Fatal(ok, err)
			}
			b.pool.limits.MaxNodes = b.pool.nodes
			if ok, err := b.Check(expected, actual); !ok || err != nil {
				t.Fatal("cached match consumed quota", ok, err)
			}
			another := bindingSchema(t, name)
			if ok, err := b.Check(expected, another); ok || !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("aggregate schema budget", ok, err)
			}
			if expected.owner.nodes != eNodes || actual.owner.nodes != aNodes {
				t.Fatal("compiled pools mutated")
			}
		})
	}
}

func TestComponentBindingLiteralConstraintsAndInvalidHandles(t *testing.T) {
	makeLiteral := func(n byte) *Type {
		c, err := Check(&model.Binary{Definitions: []model.Definition{constant(model.U8, n), exportValue("v", 0)}}, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		return c.Type
	}
	expected, matching, wrong := makeLiteral(42), makeLiteral(42), makeLiteral(43)
	b, err := expected.ComponentBindings()
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := b.Check(expected, matching); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if ok, err := b.Check(expected, wrong); ok || err != nil {
		t.Fatal("literal constraint ignored", ok, err)
	}
	for _, v := range []*Type{nil, {}, {Kind: model.Function}} {
		if _, err := v.ComponentBindings(); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	for _, pair := range [][2]*Type{{nil, matching}, {matching, expected}, {expected, nil}, {expected, {Kind: model.Component}}} {
		if ok, err := b.Check(pair[0], pair[1]); ok || !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(ok, err)
		}
	}
	var zero *ComponentBindings
	if ok, err := zero.Check(expected, matching); ok || !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(ok, err)
	}
}
