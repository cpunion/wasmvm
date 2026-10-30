package component

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTypedHostValueEqualityAndOwnedExports(t *testing.T) {
	e, _ := NewEngine(DefaultConfig())
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "value-imports"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	makeValue := func(first uint64, text string) Value {
		return Value{Elements: []Value{{Bits: first}, {Elements: []Value{{Text: text}, {Text: "界"}}}}}
	}
	input := makeValue(7, "hello")
	in, err := m.Instantiate(nil, Imports{Values: map[string]Value{"first": input, "same": makeValue(7, "hello")}}, vm.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	input.Elements[0].Bits = 99
	value, err := in.Value(nil, "values", "first")
	if err != nil || value.Elements[0].Bits != 7 || value.Elements[1].Elements[0].Text != "hello" {
		t.Fatal(value, err)
	}
	value.Elements[1].Elements[0].Text = "changed"
	if again, err := in.Value(nil, "values", "first"); err != nil || again.Elements[1].Elements[0].Text != "hello" {
		t.Fatal("export aliases runtime aggregate", again, err)
	}
	for _, path := range [][]string{nil, {"values", "missing"}, {"values", "first", "invalid"}} {
		if _, err := in.Value(nil, path...); err == nil {
			t.Fatal(path)
		}
	}
	for _, other := range []Value{makeValue(8, "hello"), makeValue(7, "different"), {Elements: []Value{{Bits: 7}, {Elements: []Value{{Text: "hello"}}}}}} {
		if result, err := m.Instantiate(nil, Imports{Values: map[string]Value{"first": makeValue(7, "hello"), "same": other}}, vm.CallOptions{}); result != nil || !errors.Is(err, vm.ErrType) {
			t.Fatal(result, err)
		}
	}
	if _, err := m.Instantiate(nil, Imports{}, vm.CallOptions{}); !errors.Is(err, vm.ErrLink) {
		t.Fatal(err)
	}
	bad := makeValue(256, "hello")
	if _, err := m.Instantiate(nil, Imports{Values: map[string]Value{"first": bad}}, vm.CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestHostConstantBoundsAndCloneLimits(t *testing.T) {
	for _, bytes := range []uint64{1024, 1} {
		config := DefaultConfig()
		config.MaxBytes = bytes
		e, _ := NewEngine(config)
		m, err := e.Compile(nil, fixture(t, "value-constant-bound"))
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range []uint64{42, 43} {
			in, err := m.Instantiate(nil, Imports{Values: map[string]Value{"input": {Bits: n}}}, vm.CallOptions{})
			if n == 43 {
				if in != nil || !errors.Is(err, vm.ErrType) {
					t.Fatal(in, err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				v, err := in.Value(nil, "input")
				if err != nil || v.Bits != 42 {
					t.Fatal(v, err)
				}
				in.Close()
			}
		}
		m.Close()
		e.Close()
	}
	config := DefaultConfig()
	config.MaxValues = 1
	e, _ := NewEngine(config)
	defer e.Close()
	m, err := e.Compile(nil, fixture(t, "value-imports"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Instantiate(nil, Imports{Values: map[string]Value{"first": {Elements: []Value{{}, {}}}}}, vm.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
