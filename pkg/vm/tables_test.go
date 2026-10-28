package vm

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTablesAliasesReentryGrowthRootsAndOrderedExits(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, typ := range []wasm.ValueType{wasm.FuncRef, wasm.ExternRef} {
			t.Run(fmt.Sprintf("%x", typ), func(t *testing.T) {
				e, s, l := setupWithEngine(t, newEngine)
				tableType := module.TableType{Element: typ, Limits: module.Limits{Min: 1, Max: 32, HasMax: true}}
				table, err := s.NewTable(TableType(tableType))
				if err != nil {
					t.Fatal(err)
				}
				null, _ := NullRef(typ)
				var first, second Value
				var change, read, grow, mutate Extern
				host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
					id := first.ref.id
					if table.external.Table.Values[0] != id {
						t.Fatal("host did not see native table write")
					}
					if r, err := c.Call(read, []Value{I32(0)}); err != nil || r.Values[0].ref == nil || r.Values[0].ref.id != id {
						t.Fatal("cross-instance alias read", r, err)
					} else {
						r.Values[0].Release()
					}
					if _, err := c.Call(change, []Value{I32(0), null}); err != nil {
						t.Fatal(err)
					}
					first.Release()
					if err := c.Collect(); err != nil {
						t.Fatal(err)
					}
					if _, err := c.Call(grow, nil); err != nil {
						t.Fatal(err)
					}
					if _, err := c.Call(change, []Value{I32(0), second}); err != nil {
						t.Fatal(err)
					}
					if _, err := c.Call(mutate, nil); err != nil {
						t.Fatal(err)
					}
					return nil, c.Collect()
				})
				if err != nil {
					t.Fatal(err)
				}
				for name, ex := range map[string]Extern{"host": host, "a": table, "b": table} {
					if err := l.Define("h", name, ex); err != nil {
						t.Fatal(err)
					}
				}
				d := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{wasm.I32, typ}, Results: []wasm.ValueType{typ, typ, wasm.I32}}, {Params: []wasm.ValueType{wasm.I32, typ}}, {Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{typ}}, {Results: []wasm.ValueType{wasm.FuncRef, wasm.FuncRef}}},
					Imports: []module.Import{{Module: "h", Name: "host", Kind: 0}, {Module: "h", Name: "a", Kind: 1, Table: tableType}, {Module: "h", Name: "b", Kind: 1, Table: tableType}},
					Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I32, Mutable: true}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}},
					Functions: []module.Function{
						{Type: 1, Locals: []module.Local{{Count: 1, Type: typ}}, Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x26}, {Op: 0x20}, {Op: 0x25, A: 1}, {Op: 0x21, A: 2}, {Op: 0x41, Bits: 41}, {Op: 0x24}, {Op: 0x10}, {Op: 0x20, A: 2}, {Op: 0x20}, {Op: 0x25}, {Op: 0x23}, {Op: 0xb}}},
						{Type: 2, Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x26, A: 1}, {Op: 0xb}}},
						{Type: 3, Body: module.Expr{{Op: 0x20}, {Op: 0x25}, {Op: 0xb}}},
						{Body: module.Expr{{Op: 0xd0, A: uint32(typ)}, {Op: 0x41, Bits: 1}, {Op: 0xfc000f}, {Op: 0x1a}, {Op: 0xb}}},
						{Body: module.Expr{{Op: 0x41, Bits: 99}, {Op: 0x24}, {Op: 0xb}}},
						{Type: 2, Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x26, A: 1}, {Op: 0x41, Bits: 0xffffffff}, {Op: 0x25, Offset: 501}, {Op: 0x1a}, {Op: 0xb}}},
						{Type: 2, Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x26}, {Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}},
						{Type: 4, Body: module.Expr{{Op: 0xd2}, {Op: 0xd2, A: 5}, {Op: 0xb}}},
					}, Elements: []module.Element{{Mode: module.Declarative, Type: wasm.FuncRef, Functions: []uint32{0, 5}}},
					Exports: []module.Export{{Name: "run", Index: 1}, {Name: "change", Index: 2}, {Name: "read", Index: 3}, {Name: "grow", Index: 4}, {Name: "mutate", Index: 5}, {Name: "trap", Index: 6}, {Name: "spin", Index: 7}, {Name: "refs", Index: 8}}}
				m := compileDefinition(t, e, d)
				in, other := instantiate(t, l, m), instantiate(t, l, m)
				change, read, grow, mutate = export(t, other, "change"), export(t, other, "read"), export(t, other, "grow"), export(t, in, "mutate")
				if typ == wasm.FuncRef {
					second, err = host.Reference()
				} else {
					second, err = s.NewExternRef("second")
				}
				if err != nil {
					t.Fatal(err)
				}
				defer second.Release()
				for _, metered := range []bool{false, true} {
					if typ == wasm.FuncRef {
						first, err = mutate.Reference()
					} else {
						first, err = s.NewExternRef("first")
					}
					if err != nil {
						t.Fatal(err)
					}
					id := first.ref.id
					r, err := export(t, in, "run").Call(nil, []Value{I32(0), first}, CallOptions{Metered: metered, Fuel: 100})
					if err != nil || len(r.Values) != 3 || r.Values[0].ref == nil || r.Values[0].ref.id != id || r.Values[1].ref == nil || r.Values[1].ref.id != second.ref.id || bits(t, r.Values[2]) != 99 {
						t.Fatal("table reentry/growth/global view", r, err)
					}
					if typ == wasm.ExternRef {
						if object, err := s.ExternRef(r.Values[0]); err != nil || object != "first" {
							t.Fatal("suspended native frame lost root", object, err)
						}
					}
					for _, v := range r.Values {
						v.Release()
					}
					if _, err := export(t, in, "trap").Call(nil, []Value{I32(0), second}, CallOptions{Metered: metered, Fuel: 100}); err == nil {
						t.Fatal("missing bounds trap")
					} else {
						var trap *wasm.Trap
						if !errors.As(err, &trap) || trap.Code != wasm.TrapTableBounds || trap.Offset != 501 {
							t.Fatal("bounds trap/source", err)
						}
					}
					if table.external.Table.Values[0] != second.ref.id {
						t.Fatal("write before trap was lost")
					}
					for _, instance := range []*Instance{in, other} {
						refs, err := export(t, instance, "refs").Call(nil, nil, CallOptions{Metered: metered, Fuel: 10})
						if err != nil || refs.Values[0].ref.id != host.external.Function || refs.Values[1].ref.id != export(t, instance, "mutate").external.Function {
							t.Fatal("imported/defined function reference identities", refs, err)
						}
						for _, v := range refs.Values {
							v.Release()
						}
					}
				}
				for fuel := uint64(0); fuel <= 3; fuel++ {
					if err := table.TableSet(0, null); err != nil {
						t.Fatal(err)
					}
					_, err := change.Call(nil, []Value{I32(0), second}, CallOptions{Metered: true, Fuel: fuel})
					if fuel < 3 && !errors.Is(err, wasm.ErrLimit) || fuel == 3 && err != nil {
						t.Fatal(fuel, err)
					}
					want := uint64(0)
					if fuel == 3 {
						want = second.ref.id
					}
					if table.external.Table.Values[0] != want {
						t.Fatal("table fuel ordering", fuel)
					}
				}
				for _, index := range []uint32{uint32(len(table.external.Table.Values)), 0xffffffff} {
					if _, err := read.Call(nil, []Value{I32(index)}, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal(index, err)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				if _, err := export(t, in, "spin").Call(ctx, []Value{I32(0), second}, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
				if table.external.Table.Values[0] != second.ref.id {
					t.Fatal("write before cancellation was lost")
				}
				for i := range table.external.Table.Values {
					if err := table.TableSet(uint32(i), null); err != nil {
						t.Fatal(err)
					}
				}
				second.Release()
				s.Collect()
				if typ == wasm.ExternRef {
					if usage, _ := s.Usage(); usage.ExternRefs != 0 || usage.HostRoots != 0 {
						t.Fatal("table roots retained after cleanup", usage)
					}
				}
			})
		}
	})
}
