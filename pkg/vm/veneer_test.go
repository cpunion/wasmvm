package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCrossInstanceVeneerLiveTableRootsFuelAndCaller(t *testing.T) {
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		e, s, linker := setupWithEngine(t, newEngine)
		types := []wasm.ValueType{wasm.I32, wasm.V128, wasm.ExternRef}
		sig := FunctionType{Params: types, Results: types}
		calls := 0
		var saved *Caller
		callback := func(delta uint32) HostFunction {
			return func(c *Caller, args []Value) ([]Value, error) {
				calls++
				saved = c
				var byteAtSecondPage [1]byte
				if err := c.ReadMemory(0, 65536, byteAtSecondPage[:]); err != nil {
					t.Fatal("forwarder caller instance lost", err)
				}
				if err := c.Collect(); err != nil {
					return nil, err
				}
				if object, err := s.ExternRef(args[2]); err != nil || object != "veneer root" {
					t.Fatal("parked forwarding frame lost reference", object, err)
				}
				args[0] = I32(uint32(bits(t, args[0])) + delta)
				return args, nil
			}
		}
		host, err := s.NewFunction(sig, callback(0))
		if err != nil {
			t.Fatal(err)
		}
		if err := linker.Define("h", "echo", host); err != nil {
			t.Fatal(err)
		}
		definition := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Module: "h", Name: "echo", Kind: 0}},
			Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}, Memories: []module.Limits{{Min: 2}},
			Elements:  []module.Element{{Type: wasm.FuncRef, Offset: module.Expr{{Op: 0x41}, {Op: 0xb}}, Functions: []uint32{0}}},
			Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20, A: 2}, {Op: 0x41}, {Op: 0x11}, {Op: 0xb}}}},
			Exports:   []module.Export{{Name: "run", Index: 1}, {Name: "table", Kind: 1}}}
		m := compileDefinition(t, e, definition)
		if m.code.Functions[0].Veneer == nil {
			t.Fatal("forwarding function was not lowered")
		}
		forwarder := instantiate(t, linker, m)
		if err := linker.Define("v", "run", export(t, forwarder, "run")); err != nil {
			t.Fatal(err)
		}
		callerDefinition := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Module: "v", Name: "run", Kind: 0}}, Memories: []module.Limits{{Min: 1}},
			Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20, A: 2}, {Op: 0x10}, {Op: 0xb}}}}, Exports: []module.Export{{Name: "run", Index: 1}}}
		caller := instantiate(t, linker, compileDefinition(t, e, callerDefinition))
		run := export(t, caller, "run")
		root, err := s.NewExternRef("veneer root")
		if err != nil {
			t.Fatal(err)
		}
		defer root.Release()
		args := []Value{I32(77), V128(0xfedcba9876543210, 0x0123456789abcdef), root}
		for fuel := uint64(0); fuel <= 10; fuel++ {
			before := calls
			result, err := run.Call(nil, args, CallOptions{Metered: true, Fuel: fuel})
			if fuel < 9 {
				if !errors.Is(err, wasm.ErrLimit) || calls != before {
					t.Fatal("fuel crossed host effect", fuel, calls, err)
				}
				continue
			}
			if err != nil || result.FuelRemaining != fuel-9 || calls != before+1 {
				t.Fatal("exact forwarding fuel", fuel, result, err)
			}
			if lo, hi, err := result.Values[1].Bits(); err != nil || lo != 0xfedcba9876543210 || hi != 0x0123456789abcdef {
				t.Fatal("vector tuple lost bits", lo, hi, err)
			}
			result.Values[2].Release()
		}
		table := export(t, forwarder, "table")
		next, err := s.NewFunction(sig, callback(1))
		if err != nil {
			t.Fatal(err)
		}
		ref, err := next.Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer ref.Release()
		if err := table.TableSet(0, ref); err != nil {
			t.Fatal(err)
		}
		result, err := run.Call(nil, args, CallOptions{})
		if err != nil || bits(t, result.Values[0]) != 78 {
			t.Fatal("table target was cached", result, err)
		}
		result.Values[2].Release()
		if err := saved.ReadMemory(0, 0, nil); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal("parked caller did not expire", err)
		}
		wrong, err := s.NewFunction(FunctionType{}, func(*Caller, []Value) ([]Value, error) { t.Fatal("wrong type called"); return nil, nil })
		if err != nil {
			t.Fatal(err)
		}
		wrongRef, err := wrong.Reference()
		if err != nil {
			t.Fatal(err)
		}
		defer wrongRef.Release()
		null, _ := NullRef(wasm.FuncRef)
		for _, value := range []Value{wrongRef, null} {
			if err := table.TableSet(0, value); err != nil {
				t.Fatal(err)
			}
			if _, err := run.Call(nil, args, CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
				t.Fatal("live table checks bypassed", err)
			}
		}
		if err := table.TableSet(0, ref); err != nil {
			t.Fatal(err)
		}
		before := calls
		frames, slots := s.config.Execution.MaxFrames, s.config.Execution.MaxSlots
		for _, limitFrames := range []bool{false, true} {
			if limitFrames {
				s.config.Execution.MaxFrames = 1
			} else {
				s.config.Execution.MaxSlots = uint32(len(m.code.Functions[0].Slots))
			}
			if _, err := run.Call(nil, args, CallOptions{}); !errors.Is(err, wasm.ErrLimit) || calls != before {
				t.Fatal("forwarding discarded logical quotas", err)
			}
			s.config.Execution.MaxFrames, s.config.Execution.MaxSlots = frames, slots
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := run.Call(ctx, args, CallOptions{}); !errors.Is(err, context.Canceled) || calls != before {
			t.Fatal("cancelled veneer called host", err)
		}
	})
}
