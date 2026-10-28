package nativeexec

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestDirectTablesBindBothTiersAndRejectNonNativeStorage(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.FuncRef}}},
		Tables:    []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}},
		Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x25}, {Op: 0xb}}}}}
	v, err := validate.Check(d, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	i, err := lower.Compile(v, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	m, err := slots.Compile(i, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	e, fn, lease := prepare(t, m)
	if !lease.entries[fn.Code].directRuntimeData || lease.entries[fn.Code].hasExpansion {
		t.Fatal("ordinary table entry missing its runtime view")
	}
	original := fn.Instance.Tables[0]
	for _, metered := range []bool{false, true} {
		e.metered = metered
		pc, _, err := e.Run(fn, 0, 1, 0, 10)
		if err != nil || fn.Code.Code[pc].Op != slots.ControlBase+uint32(ir.Return) {
			t.Fatal(metered, pc, err)
		}
	}
	goStore := state.New(state.DefaultLimits())
	defer goStore.Close()
	for _, length := range []uint32{0, 1} {
		table, err := goStore.NewTable(module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: length}})
		if err != nil {
			t.Fatal(err)
		}
		fn.Instance.Tables[0] = table.Table
		if _, _, err := e.Run(fn, 0, 1, 0, 10); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("Go table entered native view", length, err)
		}
	}
	fn.Instance.Tables[0] = nil
	if _, _, err := e.Run(fn, 0, 1, 0, 10); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("nil table entered native view", err)
	}
	fn.Instance.Tables[0] = original
	fn.Instance.Owner.Close()
	if _, err := e.bindRuntimeData(&state.Instance{Tables: []*state.Table{original}}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("retired table entered native metadata", err)
	}
	if _, _, err := e.Run(fn, 0, 1, 0, 10); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal("retired instance entered native execution", err)
	}
}
