package vm

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCompiledInterfaceOutlivesPublication(t *testing.T) {
	var missing *CompiledModule
	if typ, err := missing.Interface(); typ.Valid() || !errors.Is(err, ErrReference) {
		t.Fatal(typ, err)
	}
	e := NewEngine()
	defer e.Close()
	m := compileDefinition(t, e, &module.Module{})
	snapshot, err := m.Interface()
	if err != nil || !snapshot.Valid() {
		t.Fatal(snapshot, err)
	}
	m.Close()
	if snapshot.Nodes() != 0 {
		t.Fatal("empty module interface")
	}
	if _, ok := snapshot.Intern(&module.TypeRegistry{}, 1); !ok {
		t.Fatal("closing code invalidated metadata")
	}
	if typ, err := m.Interface(); typ.Valid() || !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(typ, err)
	}
}
