package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarStorageWidthsAliasesRebindingAndFallback(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		first, second := []uint64{0, 17}, []uint64{0, 29}
		current, valid, panicBinding := first, true, false
		sentinel := errors.New("binding panic")
		storage, err := NewScalarStorage(2, func() ([]uint64, bool) {
			if panicBinding {
				panic(sentinel)
			}
			return current, valid
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range []struct {
			name  string
			index uint32
			typ   wasm.ValueType
			write bool
		}{{"get", 0, wasm.I64, false}, {"set", 0, wasm.I64, true}, {"get32", 0, wasm.I32, false}, {"set32", 0, wasm.I32, true}, {"other", 1, wasm.I64, false}} {
			fn, err := s.NewScalarStorageFunction(storage, test.index, test.typ, test.write)
			if err != nil {
				t.Fatal(err)
			}
			if err := l.Define("h", test.name, fn); err != nil {
				t.Fatal(err)
			}
		}
		hook, err := s.NewFunction(FunctionType{}, func(c *Caller, _ []Value) ([]Value, error) {
			if first[0] != 0xfedcba9876543210 {
				t.Fatalf("write was not published before callback: %x", first[0])
			}
			current = second
			get, err := c.NewScalarStorageFunction(storage, 1, wasm.I64, false)
			if err != nil {
				return nil, err
			}
			result, err := c.Call(get, nil)
			if err != nil || bits(t, result.Values[0]) != 29 {
				t.Fatal("rebound direct accessor", result, err)
			}
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := l.Define("h", "hook", hook); err != nil {
			t.Fatal(err)
		}
		fallbackCalls := 0
		fallback, err := s.NewFunction(FunctionType{Results: []wasm.ValueType{wasm.I64}}, func(*Caller, []Value) ([]Value, error) {
			fallbackCalls++
			return []Value{I64(73)}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := l.Define("h", "fallback", fallback); err != nil {
			t.Fatal(err)
		}
		d := &module.Module{Types: []module.FuncType{
			{Results: []wasm.ValueType{wasm.I64}}, {Params: []wasm.ValueType{wasm.I64}},
			{Results: []wasm.ValueType{wasm.I32}}, {Params: []wasm.ValueType{wasm.I32}}, {},
			{Results: []wasm.ValueType{wasm.I64, wasm.I32, wasm.I64, wasm.I64}},
		}, Imports: []module.Import{
			{Module: "h", Name: "get", Kind: 0}, {Module: "h", Name: "set", Kind: 0, Type: 1},
			{Module: "h", Name: "get32", Kind: 0, Type: 2}, {Module: "h", Name: "set32", Kind: 0, Type: 3},
			{Module: "h", Name: "other", Kind: 0}, {Module: "h", Name: "hook", Kind: 0, Type: 4},
			{Module: "h", Name: "fallback", Kind: 0},
		}, Functions: []module.Function{
			{Type: 5, Body: module.Expr{{Op: 0x42, Bits: 0xfedcba9876543210}, {Op: 0x10, A: 1}, {Op: 0x10, A: 5},
				{Op: 0x41, Bits: 0xffffffff}, {Op: 0x10, A: 3}, {Op: 0x10}, {Op: 0x10, A: 2}, {Op: 0x10, A: 4}, {Op: 0x10, A: 6}, {Op: 0xb}}},
			{Body: module.Expr{{Op: 0x12}, {Op: 0xb}}},
			{Body: module.Expr{{Op: 0x10}, {Op: 0xb}}},
		}, Exports: []module.Export{{Name: "run", Index: 7}, {Name: "tail", Index: 8}, {Name: "direct", Index: 0}, {Name: "get", Index: 9}}}
		in := instantiate(t, l, compileDefinitionWithFeatures(t, e, d, wasm.TailCall))
		for _, options := range []CallOptions{{}, {Metered: true, Fuel: 100}} {
			current, valid = first, true
			result, err := export(t, in, "run").Call(nil, nil, options)
			if err != nil {
				t.Fatal(err)
			}
			for i, want := range []uint64{0xffffffff, 0xffffffff, 29, 73} {
				if got := bits(t, result.Values[i]); got != want {
					t.Fatalf("result %d: %x != %x", i, got, want)
				}
			}
			if first[0] != 0xfedcba9876543210 || second[0] != 0xffffffff {
				t.Fatal("aliased writes or binding isolation", first, second)
			}
			for _, name := range []string{"direct", "tail"} {
				result, err := export(t, in, name).Call(nil, nil, options)
				if err != nil || bits(t, result.Values[0]) != 0xffffffff {
					t.Fatal(name, result, err)
				}
			}
		}
		if fallbackCalls != 2 {
			t.Fatal("ordinary import bypassed", fallbackCalls)
		}
		for _, words := range [][]uint64{nil, {1}, {1, 2, 3}} {
			current = words
			for _, name := range []string{"get", "tail", "direct"} {
				if _, err := export(t, in, name).Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal("invalid storage length", name, words, err)
				}
			}
		}
		current, valid = first, false
		for _, name := range []string{"get", "tail", "direct"} {
			if _, err := export(t, in, name).Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("inactive storage", name, err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := export(t, in, "tail").Call(ctx, nil, CallOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled accessor", err)
		}
		get := export(t, in, "get")
		panicBinding = true
		func() {
			defer func() {
				if got := recover(); got != sentinel {
					t.Fatal("binding panic was not propagated", got)
				}
			}()
			get.Call(nil, nil, CallOptions{})
		}()
		panicBinding, valid = false, true
		if result, err := get.Call(nil, nil, CallOptions{}); err != nil || bits(t, result.Values[0]) != first[0] {
			t.Fatal("binding panic retained active workspace", result, err)
		}
	})
}

func TestScalarStorageRegistrationContracts(t *testing.T) {
	for _, size := range []uint32{0, 1} {
		if _, err := NewScalarStorage(size, nil); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
	}
	if _, err := NewScalarStorage(0, func() ([]uint64, bool) { return nil, false }); !errors.Is(err, ErrType) {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		storage, err := NewScalarStorage(1, func() ([]uint64, bool) { return []uint64{1}, true })
		if err != nil {
			t.Fatal(err)
		}
		for _, typ := range []wasm.ValueType{0, wasm.F32, wasm.F64, wasm.V128, wasm.ExternRef} {
			if _, err := s.NewScalarStorageFunction(storage, 0, typ, false); !errors.Is(err, ErrType) {
				t.Fatal(typ, err)
			}
		}
		for _, st := range []*ScalarStorage{nil, storage} {
			if _, err := s.NewScalarStorageFunction(st, 1, wasm.I32, true); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
		}
		if _, err := new(Caller).NewScalarStorageFunction(storage, 0, wasm.I64, false); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
		s.Close()
		if _, err := s.NewScalarStorageFunction(storage, 0, wasm.I64, false); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	})
}

func TestScalarStorageOrdinaryImportKeepsDirectCallIdentity(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, l := setupWithEngine(t, newEngine)
		var calls int
		var received uint64
		plain, err := s.NewFunction(FunctionType{Params: []wasm.ValueType{wasm.I64}}, func(_ *Caller, args []Value) ([]Value, error) {
			calls++
			received = bits(t, args[0])
			return nil, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		storage, err := NewScalarStorage(1, func() ([]uint64, bool) { return []uint64{0}, true })
		if err != nil {
			t.Fatal(err)
		}
		get, err := s.NewScalarStorageFunction(storage, 0, wasm.I64, false)
		if err != nil {
			t.Fatal(err)
		}
		for name, fn := range map[string]Extern{"plain": plain, "get": get} {
			if err := l.Define("h", name, fn); err != nil {
				t.Fatal(err)
			}
		}
		d := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I64}}, {Results: []wasm.ValueType{wasm.I64}}},
			Imports:   []module.Import{{Module: "h", Name: "plain", Kind: 0}, {Module: "h", Name: "get", Kind: 0, Type: 1}},
			Functions: []module.Function{{Body: module.Expr{{Op: 0xb}}}, {Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0x20}, {Op: 0x10, A: 2}, {Op: 0xb}}}},
			Exports:   []module.Export{{Name: "noop", Index: 2}, {Name: "run", Index: 3}}}
		in := instantiate(t, l, compileDefinition(t, e, d))
		// The integer argument deliberately equals a same-signature guest
		// function identity. A direct import must never dispatch through it.
		identity := export(t, in, "noop").external.Function
		for _, options := range []CallOptions{{}, {Metered: true, Fuel: 100}} {
			if _, err := export(t, in, "run").Call(nil, []Value{I64(identity)}, options); err != nil {
				t.Fatal(err)
			}
			if received != identity {
				t.Fatal("ordinary import lost its integer argument", received)
			}
		}
		if calls != 2 {
			t.Fatal("ordinary import was dispatched as an indirect call", calls)
		}
	})
}
