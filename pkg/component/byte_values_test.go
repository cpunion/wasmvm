package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
)

func TestByteValueBindingsNormalizeEquivalentRepresentations(t *testing.T) {
	e, err := NewEngine(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	legacy := Value{Elements: []Value{{Bits: 0}, {Bits: 255}, {Bits: 128}, {Bits: 1}}}
	packed := ByteList([]byte{0, 255, 128, 1})
	for _, name := range []string{"byte-value-imports", "byte-value-constant-bound"} {
		m, err := e.Compile(nil, fixture(t, name))
		if err != nil {
			t.Fatal(name, err)
		}
		for _, a := range []Value{packed, legacy} {
			for _, b := range []Value{legacy, packed} {
				values := map[string]Value{"input": a}
				path := []string{"input"}
				if name == "byte-value-imports" {
					values = map[string]Value{"first": {Elements: []Value{{Bits: 7}, a}}, "same": {Elements: []Value{{Bits: 7}, b}}}
					path = []string{"values", "same"}
				}
				in, err := m.Instantiate(nil, Imports{Values: values}, CallOptions{})
				if err != nil {
					t.Fatal(name, err)
				}
				v, err := in.Value(nil, path...)
				if err != nil {
					t.Fatal(err)
				}
				if name == "byte-value-imports" {
					v = v.Elements[1]
				}
				if v.Text != packed.Text || v.Elements != nil {
					t.Fatal("binding did not retain canonical packed bytes", v)
				}
				in.Close()
			}
		}
		values := map[string]Value{"input": ByteList([]byte{0, 255, 128, 2})}
		if name == "byte-value-imports" {
			values = map[string]Value{"first": {Elements: []Value{{Bits: 7}, packed}}, "same": {Elements: []Value{{Bits: 7}, values["input"]}}}
		}
		if in, err := m.Instantiate(nil, Imports{Values: values}, CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal(in, err)
		}
		m.Close()
	}
}
