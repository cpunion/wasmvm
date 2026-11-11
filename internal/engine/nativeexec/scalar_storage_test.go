package nativeexec

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarStorageGeneratedAccessAndExitPublication(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, metered := range []bool{false, true} {
		for _, trap := range []bool{false, true} {
			d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I64}}, {Results: []wasm.ValueType{wasm.I64}}, {Params: []wasm.ValueType{wasm.I64}, Results: []wasm.ValueType{wasm.I64}}},
				Imports:   []module.Import{{Kind: 0, Name: "set"}, {Kind: 0, Name: "get", Type: 1}},
				Functions: []module.Function{{Type: 2, Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0x10, A: 1}, {Op: 0xb}}}}}
			if trap {
				d.Functions[0].Body[2] = module.Instr{Op: 0}
			}
			m := callSlots(t, d)
			compiled, err := Compile(context.Background(), m, codegen.DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(compiled.Close)
			lease, err := compiled.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			s := state.NewNative(state.DefaultLimits())
			t.Cleanup(s.Close)
			cells, valid := []uint64{17, 0}, true
			storage := &state.ScalarStorage{Words: 2, Bind: func() ([]uint64, bool) { return cells, valid }}
			set, err := s.NewHost(d.Types[0], &state.ScalarAccess{Storage: storage, Index: 1, Type: wasm.I64, Write: true})
			if err != nil {
				t.Fatal(err)
			}
			get, err := s.NewHost(d.Types[1], &state.ScalarAccess{Storage: storage, Index: 1, Type: wasm.I64})
			if err != nil {
				t.Fatal(err)
			}
			in, err := s.Allocate(m, func(_, name string) (state.Extern, bool) {
				if name == "set" {
					return set, true
				}
				return get, true
			})
			if err != nil {
				t.Fatal(err)
			}
			fn, _ := s.Function(in.Functions[2])
			n, err := New(context.Background(), map[*slots.Module]*Lease{m: lease}, 16, 2, metered)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(n.Close)
			if _, err := n.Resize(len(fn.Code.Slots), 16); err != nil {
				t.Fatal(err)
			}
			for _, want := range []uint64{0, 1 << 63, 0xfedcba9876543210, ^uint64(0)} {
				n.values[0].Lo = want
				pc, _, err := n.Run(fn, 0, 1, 0, 100)
				if cells[0] != 17 || cells[1] != want {
					t.Fatal("native write was not published", metered, trap, cells)
				}
				if trap {
					if !errors.Is(err, wasm.ErrTrap) {
						t.Fatal("trap after write", pc, err)
					}
				} else {
					if err != nil || fn.Code.Code[pc].Op < slots.ControlBase {
						t.Fatal("accessor exited to common dispatch", pc, err)
					}
					result := fn.Code.Args[fn.Code.Code[pc].Extra]
					if n.values[result].Lo != want || n.values[result].Hi != 0 {
						t.Fatal("native result width", n.values[result])
					}
				}
			}
			valid = false
			pc, _, err := n.Run(fn, 0, 1, 0, 100)
			if err != nil || fn.Code.Code[pc].Op != 0x10 {
				t.Fatal("inactive storage did not use checked call", pc, err)
			}
			if len(n.scalarBindings) != 1 || len(n.scalarAccesses) != 2 {
				t.Fatal("aliases duplicated storage")
			}
			retained := n.scalarBindings[:cap(n.scalarBindings)]
			if !n.Suspend() {
				t.Fatal("suspend")
			}
			for _, b := range retained {
				if b.storage != nil || b.values != nil {
					t.Fatal("idle workspace retained task storage")
				}
			}
		}
	}
}
