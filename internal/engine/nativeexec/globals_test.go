package nativeexec

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestRuntimeGlobalBindingAliasesRefreshAndCleanup(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	e, fn, _ := prepare(t, fixture(t, false))
	g := &state.Global{Type: module.GlobalType{Value: wasm.V128, Mutable: true}, Value: value.Value{Lo: 42, Hi: 17}}
	in := &state.Instance{Functions: []uint64{10}, Globals: []*state.Global{g, g}}
	indices := []uint32{0, 1}
	for range 2 {
		address, err := e.bindRuntimeGlobals(in, fn.Code, indices)
		if err != nil {
			t.Fatal(err)
		}
		if len(e.runtimeGlobals) != 1 || e.runtimeWords[1] != e.runtimeWords[2] || e.runtimeWords[1] != address+24 {
			t.Fatal("aliases did not share one native cell", e.runtimeWords, e.runtimeGlobals)
		}
		if e.runtimeWords[3] != g.Value.Lo || e.runtimeWords[4] != g.Value.Hi {
			t.Fatal("snapshot was not refreshed")
		}
		e.runtimeWords[3], e.runtimeWords[4] = 100, 200
		e.flushRuntimeGlobals()
		if g.Value != (value.Value{Lo: 100, Hi: 200}) {
			t.Fatal("writeback lost bits", g.Value)
		}
		g.Value = value.Value{Lo: 43, Hi: 18}
	}
	// A larger instance reallocates the native view and refreshes cell pointers.
	other := &state.Instance{Functions: make([]uint64, 100), Globals: []*state.Global{g, g}}
	if address, err := e.bindRuntimeGlobals(other, fn.Code, indices); err != nil || e.runtimeWords[100] != address+102*8 {
		t.Fatal(address, err)
	}
	// Switching function rebuilds bindings, including a function with no globals.
	if _, err := e.bindRuntimeGlobals(in, &slots.Function{}, []uint32{1}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.bindRuntimeData(in); err != nil || len(e.runtimeGlobals) != 0 {
		t.Fatal("stale global bindings", err)
	}
	for _, bad := range [][]*state.Global{{nil}, {{Type: module.GlobalType{Value: wasm.ExternRef}}}} {
		in := &state.Instance{Globals: bad}
		if _, err := e.bindRuntimeGlobals(in, fn.Code, []uint32{0}); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	if _, err := e.bindRuntimeGlobals(in, fn.Code, []uint32{99}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	for _, close := range []bool{false, true} {
		if _, err := e.bindRuntimeGlobals(in, fn.Code, indices); err != nil {
			t.Fatal(err)
		}
		retained := e.runtimeGlobals[:cap(e.runtimeGlobals)]
		if close {
			e.Close()
		} else if !e.Suspend() {
			t.Fatal("suspend")
		}
		if e.runtimeFunction != nil || len(e.runtimeGlobals) != 0 {
			t.Fatal("retained function/global roots")
		}
		for _, binding := range retained {
			if binding.global != nil {
				t.Fatal("retained tail root")
			}
		}
		if !close {
			if err := e.Begin(nil, false); err != nil {
				t.Fatal(err)
			}
		}
	}
}
