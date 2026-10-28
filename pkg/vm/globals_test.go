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

func TestGlobalsAliasesHostReentryAndExitVisibility(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, first := range []Value{I32(0xffc01234), I64(0xfedcba9876543210), F32Bits(0xffc01234), F64Bits(0xfff8123456789abc), V128(0xfedcba9876543210, 0x123456789abcdef0)} {
			t.Run(fmt.Sprintf("%x", first.Type()), func(t *testing.T) {
				e, s, l := setupWithEngine(t, newEngine)
				typ := first.Type()
				second := first
				second.bits.Lo ^= 0x12345678
				second.bits.Hi ^= first.bits.Hi
				gtype := module.GlobalType{Value: typ, Mutable: true}
				global, err := s.NewGlobal(GlobalType(gtype), first)
				if err != nil {
					t.Fatal(err)
				}
				var read, change, plain Extern
				var expected Value
				calls := 0
				host, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
					calls++
					if global.external.Global.Value != expected.bits {
						t.Fatal("host did not see preceding guest write", global.external.Global.Value, expected.bits)
					}
					if r, err := c.Call(read, nil); err != nil || r.Values[0].bits != expected.bits {
						t.Fatal("same-instance read", r, err)
					}
					if r, err := c.Call(change, []Value{second}); err != nil || r.Values[0].bits != second.bits {
						t.Fatal("cross-instance alias", r, err)
					}
					if _, err := c.Call(plain, nil); err != nil {
						t.Fatal(err)
					}
					if r, err := c.Call(read, nil); err != nil || r.Values[0].bits != second.bits {
						t.Fatal("no-global entry restored stale cells", r, err)
					}
					return nil, c.Collect()
				})
				if err != nil {
					t.Fatal(err)
				}
				for name, ex := range map[string]Extern{"a": global, "b": global, "host": host} {
					if err := l.Define("h", name, ex); err != nil {
						t.Fatal(err)
					}
				}
				d := &module.Module{Types: []module.FuncType{{}, {Params: []wasm.ValueType{typ}, Results: []wasm.ValueType{typ}}, {Results: []wasm.ValueType{typ}}, {Params: []wasm.ValueType{typ}}},
					Imports: []module.Import{{Module: "h", Name: "host", Kind: 0}, {Module: "h", Name: "a", Kind: 3, Global: gtype}, {Module: "h", Name: "b", Kind: 3, Global: gtype}},
					Functions: []module.Function{
						{Type: 1, Body: module.Expr{{Op: 0x20}, {Op: 0x24}, {Op: 0x23, A: 1}, {Op: 0x1a}, {Op: 0x10}, {Op: 0x23}, {Op: 0xb}}},
						{Type: 1, Body: module.Expr{{Op: 0x20}, {Op: 0x24, A: 1}, {Op: 0x23}, {Op: 0xb}}},
						{Type: 2, Body: module.Expr{{Op: 0x23}, {Op: 0xb}}},
						{Body: module.Expr{{Op: 0xb}}},
						{Type: 3, Body: module.Expr{{Op: 0x20}, {Op: 0x24}, {Op: 0}, {Op: 0xb}}},
						{Type: 3, Body: module.Expr{{Op: 0x20}, {Op: 0x24}, {Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}},
					}, Exports: []module.Export{{Name: "run", Index: 1}, {Name: "change", Index: 2}, {Name: "read", Index: 3}, {Name: "plain", Index: 4}, {Name: "trap", Index: 5}, {Name: "spin", Index: 6}}}
				m := compileDefinition(t, e, d)
				in, other := instantiate(t, l, m), instantiate(t, l, m)
				read, change, plain = export(t, in, "read"), export(t, other, "change"), export(t, other, "plain")
				run := export(t, in, "run")
				for _, metered := range []bool{false, true} {
					for range 2 {
						expected = first
						if r, err := run.Call(nil, []Value{first}, CallOptions{Metered: metered, Fuel: 100}); err != nil || r.Values[0].bits != second.bits {
							t.Fatal("resumed global snapshot", r, err)
						}
						if got, err := global.GlobalGet(); err != nil || got.bits != second.bits {
							t.Fatal(got, err)
						}
					}
					if _, err := export(t, in, "trap").Call(nil, []Value{first}, CallOptions{Metered: metered, Fuel: 100}); !errors.Is(err, wasm.ErrTrap) {
						t.Fatal(err)
					}
					if got, err := global.GlobalGet(); err != nil || got.bits != first.bits {
						t.Fatal("write before trap lost", got, err)
					}
				}
				if calls != 4 {
					t.Fatal("host callback count", calls)
				}
				for fuel := uint64(0); fuel <= 2; fuel++ {
					if err := global.GlobalSet(second); err != nil {
						t.Fatal(err)
					}
					if _, err := export(t, in, "trap").Call(nil, []Value{first}, CallOptions{Metered: true, Fuel: fuel}); !errors.Is(err, wasm.ErrLimit) {
						t.Fatal(fuel, err)
					}
					want := second.bits
					if fuel == 2 {
						want = first.bits
					}
					if got, err := global.GlobalGet(); err != nil || got.bits != want {
						t.Fatal("fuel write visibility", fuel, got, want, err)
					}
				}
				if err := global.GlobalSet(second); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				defer cancel()
				if _, err := export(t, in, "spin").Call(ctx, []Value{first}, CallOptions{}); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
				if got, err := global.GlobalGet(); err != nil || got.bits != first.bits {
					t.Fatal("write before cancellation lost", got, err)
				}
			})
		}
	})
}
