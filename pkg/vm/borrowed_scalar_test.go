package vm

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestBorrowedScalarOverlapReentryAndCollection(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		for _, counts := range [][2]int{{0, 0}, {0, 5}, {5, 0}, {4, 4}, {17, 19}, {65, 2}} {
			t.Run(fmt.Sprint(counts), func(t *testing.T) {
				e, s, l := setupWithEngine(t, newEngine)
				types := []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64}
				sig := FunctionType{Params: make([]wasm.ValueType, counts[0]), Results: make([]wasm.ValueType, counts[1])}
				args := make([]Value, counts[0])
				for i := range sig.Params {
					sig.Params[i] = types[i%4]
					args[i] = scalarTestValue(sig.Params[i], uint64(i)+0x80000000)
				}
				for i := range sig.Results {
					sig.Results[i] = types[i%4]
				}
				var expired *Caller
				var copied Caller
				calls := 0
				host, err := s.NewBorrowedScalarFunction(sig, func(c *Caller, words []uint64) error {
					calls++
					expired, copied = c, *c
					if len(words) != max(counts[0], counts[1]) || cap(words) != len(words) {
						t.Fatal("borrow exposes storage outside its signature", len(words), cap(words))
					}
					for i := range words {
						want := uint64(0)
						if i < counts[0] {
							want = uint64(i) + 0x80000000
						}
						if words[i] != want {
							t.Fatal("overlapping parameter packing or result initialization", i, words[i], want)
						}
					}
					child, err := c.NewBorrowedScalarFunction(FunctionType{Results: []wasm.ValueType{wasm.I64}}, func(fresh *Caller, nested []uint64) error {
						if nested[0] != 0 {
							t.Fatal("nested result storage was not initialized")
						}
						if err := fresh.Collect(); err != nil {
							return err
						}
						runtime.GC()
						nested[0] = 123
						return nil
					})
					if err != nil {
						return err
					}
					result, err := c.Call(child, nil)
					if err != nil || bits(t, result.Values[0]) != 123 {
						t.Fatal(result, err)
					}
					for i := range sig.Params {
						if words[i] != uint64(i)+0x80000000 {
							t.Fatal("reentry changed parent storage", i, words[i])
						}
					}
					for i, typ := range sig.Results {
						words[i] = uint64(i) + 0x12340000
						if typ == wasm.I32 || typ == wasm.F32 {
							words[i] |= 0xdeadbeef00000000
						}
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := l.Define("h", "borrowed", host); err != nil {
					t.Fatal(err)
				}
				body := module.Expr{}
				for i := range sig.Params {
					body = append(body, module.Instr{Op: 0x20, A: uint32(i)})
				}
				body = append(body, module.Instr{Op: 0x10}, module.Instr{Op: 0xb})
				d := &module.Module{Types: []module.FuncType{module.FuncType(sig)}, Imports: []module.Import{{Module: "h", Name: "borrowed", Kind: 0}}, Functions: []module.Function{{Body: body}}, Exports: []module.Export{{Name: "run", Index: 1}}}
				guest := export(t, instantiate(t, l, compileDefinition(t, e, d)), "run")
				for round := range 3 {
					fn := guest
					if round == 1 {
						fn = host
					}
					result, err := fn.Call(nil, args, CallOptions{})
					if err != nil || len(result.Values) != counts[1] {
						t.Fatal(result, err)
					}
					for i, value := range result.Values {
						if bits(t, value) != uint64(i)+0x12340000 {
							t.Fatal("overlapping result unpacking", i, value)
						}
					}
					for _, c := range []*Caller{expired, &copied} {
						if _, err := c.NewBorrowedScalarFunction(FunctionType{}, func(*Caller, []uint64) error { return nil }); !errors.Is(err, wasm.ErrClosed) {
							t.Fatal("expired capability became usable", err)
						}
					}
				}
				if calls != 3 {
					t.Fatal(calls)
				}
			})
		}
	})
}

func scalarTestValue(typ wasm.ValueType, bits uint64) Value {
	switch typ {
	case wasm.I32:
		return I32(uint32(bits))
	case wasm.I64:
		return I64(bits)
	case wasm.F32:
		return F32Bits(uint32(bits))
	default:
		return F64Bits(bits)
	}
}

func TestBorrowedScalarErrorsPanicAndCancellation(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		if _, err := s.NewBorrowedScalarFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		for _, typ := range []wasm.ValueType{wasm.V128, wasm.FuncRef, wasm.ExternRef, 0} {
			if _, err := s.NewBorrowedScalarFunction(FunctionType{Params: []wasm.ValueType{typ}}, func(*Caller, []uint64) error { return nil }); !errors.Is(err, ErrType) {
				t.Fatal(err)
			}
		}
		sentinel := errors.New("borrowed host failure")
		for _, mode := range []string{"error", "panic", "cancel"} {
			ctx, cancel := context.WithCancel(context.Background())
			var saved *Caller
			fn, err := s.NewBorrowedScalarFunction(FunctionType{}, func(c *Caller, _ []uint64) error {
				saved = c
				if _, err := c.NewExternRef("temporary borrowed root"); err != nil {
					return err
				}
				if mode == "panic" {
					panic(sentinel)
				}
				if mode == "cancel" {
					cancel()
					return nil
				}
				return sentinel
			})
			if err != nil {
				t.Fatal(err)
			}
			func() {
				defer func() {
					if p := recover(); (mode == "panic" && p != sentinel) || (mode != "panic" && p != nil) {
						t.Fatal("panic changed", p)
					}
				}()
				_, err := fn.Call(ctx, nil, CallOptions{})
				if mode == "error" && !errors.Is(err, sentinel) || mode == "cancel" && !errors.Is(err, context.Canceled) || mode == "panic" {
					t.Fatal(mode, err)
				}
			}()
			cancel()
			if err := saved.Collect(); !errors.Is(err, wasm.ErrClosed) {
				t.Fatal(err)
			}
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			usage, err := s.Usage()
			if err != nil || usage.HostRoots != 0 || usage.ExternRefs != 0 {
				t.Fatal("borrowed callback leaked roots", usage, err)
			}
		}
	})
}

func TestBorrowedSuspendableScalarImmediateAndPending(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		_, s, _ := setupWithEngine(t, newEngine)
		if _, err := s.NewBorrowedSuspendableScalarFunction(FunctionType{}, nil); !errors.Is(err, ErrType) {
			t.Fatal(err)
		}
		for _, immediate := range []bool{false, true} {
			ready := false
			var expired *Caller
			var temporary Value
			sig := FunctionType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I64, wasm.I32}}
			_, err := s.WithCaller(nil, CallOptions{}, func(c *Caller) error {
				fn, err := c.NewBorrowedSuspendableScalarFunction(sig, func(original *Caller, words []uint64) (*HostSuspension, error) {
					expired = original
					n := words[0] // Pending work owns scalar copies, never the borrow.
					var failed error
					temporary, failed = original.NewExternRef("parked scalar root")
					if failed != nil {
						return nil, failed
					}
					if immediate {
						words[0], words[1] = n+7, n+9|0xdeadbeef00000000
						return nil, nil
					}
					return &HostSuspension{Ready: func() bool { return ready }, Resume: func(fresh *Caller) (HostResult, error) {
						if fresh == original {
							t.Fatal("suspended scalar reused Caller")
						}
						if object, err := s.ExternRef(temporary); err != nil || object != "parked scalar root" {
							t.Fatal("suspended scalar lost its temporary root", object, err)
						}
						return HostResult{Values: []Value{I64(n + 7), I32(uint32(n + 9))}}, nil
					}}, nil
				})
				if err != nil {
					return err
				}
				call, err := c.Start(fn, []Value{I32(35)})
				if err != nil {
					return err
				}
				result, done, err := c.Poll(call)
				if err != nil || done != immediate {
					t.Fatal(done, err)
				}
				if _, err := expired.NewBorrowedSuspendableScalarFunction(FunctionType{}, nil); !errors.Is(err, wasm.ErrClosed) {
					t.Fatal("scalar starter did not expire", err)
				}
				if !immediate {
					if err := c.Collect(); err != nil {
						return err
					}
					runtime.GC()
					ready = true
					result, done, err = c.Poll(call)
				}
				if err != nil || !done || len(result.Values) != 2 || bits(t, result.Values[0]) != 42 || bits(t, result.Values[1]) != 44 {
					t.Fatal(result, done, err)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Collect(); err != nil {
				t.Fatal(err)
			}
			usage, err := s.Usage()
			if err != nil || usage.HostRoots != 0 || usage.ExternRefs != 0 {
				t.Fatal(usage, err)
			}
		}
		fn, err := s.NewBorrowedSuspendableScalarFunction(FunctionType{}, func(*Caller, []uint64) (*HostSuspension, error) {
			return &HostSuspension{Ready: func() bool { return true }, Resume: func(*Caller) (HostResult, error) { return HostResult{}, nil }}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fn.Call(nil, nil, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("synchronous call accepted pending work", err)
		}
	})
}
