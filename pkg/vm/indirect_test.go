package vm

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestIndirectLeavesObserveHostReentryGrowthAndTargetChanges(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		typ := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
		d := &module.Module{
			Types:   []module.FuncType{typ, typ, {Params: []wasm.ValueType{wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, {}},
			Imports: []module.Import{{Module: "h", Name: "change", Kind: 0, Type: 3}, {Module: "h", Name: "table", Kind: 1, Table: module.TableType{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}},
			Functions: []module.Function{
				{Body: module.Expr{{Op: 0x20}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0xb}}},
				{Type: 1, Body: module.Expr{{Op: 0x20}, {Op: 0x41, Bits: 3}, {Op: 0x6c}, {Op: 0xb}}},
				{Type: 2, Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x11, A: 1}, {Op: 0xb}}},
				{Type: 2, Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x11}, {Op: 0x10}, {Op: 0x20, A: 1}, {Op: 0x11, A: 1}, {Op: 0xb}}},
				{Type: 3, Body: module.Expr{{Op: 0xd2, A: 2}, {Op: 0x41, Bits: 1}, {Op: 0xfc000f}, {Op: 0x1a}, {Op: 0x41}, {Op: 0xd2, A: 2}, {Op: 0x26}, {Op: 0xb}}},
			},
			Elements: []module.Element{{Mode: module.Declarative, Type: wasm.FuncRef, Functions: []uint32{1, 2}}},
			Exports:  []module.Export{{Name: "a", Kind: 0, Index: 1}, {Name: "b", Kind: 0, Index: 2}, {Name: "run", Kind: 0, Index: 3}, {Name: "reenter", Kind: 0, Index: 4}, {Name: "mutate", Kind: 0, Index: 5}},
		}
		table, err := s.NewTable(TableType{Element: wasm.FuncRef, Limits: MemoryType{Min: 1}})
		if err != nil {
			t.Fatal(err)
		}
		var mutate, other Extern
		host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			if _, err := c.Call(other, []Value{I32(7), I32(0)}); err != nil {
				return nil, err
			}
			if _, err := c.Call(mutate, nil); err != nil {
				return nil, err
			}
			return nil, c.Collect()
		})
		if err != nil {
			t.Fatal(err)
		}
		if err = l.Define("h", "table", table); err != nil {
			t.Fatal(err)
		}
		if err = l.Define("h", "change", host); err != nil {
			t.Fatal(err)
		}
		m := compileDefinition(t, e, d)
		in := instantiate(t, l, m)
		other = export(t, instantiate(t, l, m), "run")
		mutate = export(t, in, "mutate")
		run := export(t, in, "run")
		a, _ := export(t, in, "a").Reference()
		defer a.Release()
		if err := table.TableSet(0, a); err != nil {
			t.Fatal(err)
		}
		for fuel := uint64(0); fuel <= 7; fuel++ {
			r, err := run.Call(nil, []Value{I32(5), I32(0)}, CallOptions{Metered: true, Fuel: fuel})
			if fuel < 6 {
				if !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(fuel, r, err)
				}
			} else if err != nil || bits(t, r.Values[0]) != 6 || r.FuelRemaining != fuel-6 {
				t.Fatal(fuel, r, err)
			}
		}
		if r, err := export(t, in, "reenter").Call(nil, []Value{I32(5), I32(0)}, CallOptions{}); err != nil || bits(t, r.Values[0]) != 18 {
			t.Fatal(r, err)
		}
		// The new element and replaced old element both point at the second leaf.
		for _, index := range []uint32{0, 1} {
			if r, err := run.Call(nil, []Value{I32(5), I32(index)}, CallOptions{}); err != nil || bits(t, r.Values[0]) != 15 {
				t.Fatal(index, r, err)
			}
		}
		null, _ := NullRef(wasm.FuncRef)
		wrong, _ := mutate.Reference()
		defer wrong.Release()
		foreign, _ := other.Reference()
		defer foreign.Release()
		for _, test := range []struct {
			ref   Value
			index uint32
			trap  wasm.TrapCode
		}{
			{null, 0, wasm.TrapUninitializedElement}, {wrong, 0, wasm.TrapIndirectCallType},
			{foreign, 0, wasm.TrapIndirectCallType}, {a, 2, wasm.TrapTableBounds}, {a, ^uint32(0), wasm.TrapTableBounds},
		} {
			if err := table.TableSet(0, test.ref); err != nil {
				t.Fatal(err)
			}
			_, err := run.Call(nil, []Value{I32(5), I32(test.index)}, CallOptions{})
			var trap *wasm.Trap
			if !errors.As(err, &trap) || trap.Code != test.trap {
				t.Fatal(test.index, err, test.trap)
			}
		}
		// A matching host target remains on the checked host-call path.
		h, _ := s.NewFunction(FunctionType(typ), func(*Caller, []Value) ([]Value, error) { return []Value{I32(99)}, nil })
		r, _ := h.Reference()
		defer r.Release()
		if err := table.TableSet(0, r); err != nil {
			t.Fatal(err)
		}
		if result, err := run.Call(nil, []Value{I32(5), I32(0)}, CallOptions{}); err != nil || bits(t, result.Values[0]) != 99 {
			t.Fatal(result, err)
		}
	})
}
