package nativeexec

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSharedIndirectTableFitsLinearCodeBudget(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	const count = 128
	d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}},
		Tables: []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}}}
	for range count {
		d.Functions = append(d.Functions, module.Function{Body: module.Expr{{Op: 0x41}, {Op: 0x11}, {Op: 0xb}}})
	}
	m := callSlots(t, d)
	config := codegen.Config{SharedDispatch: true, FunctionCount: count, TableCount: 1,
		LinkedCallees: map[uint32]*slots.Function{}, LinkedSignatures: make([]uint32, count), LinkedTypeLabels: m.IR.Validated.TypeLabels()}
	for i, f := range m.Functions {
		config.LinkedCallees[uint32(i)] = f
		config.LinkedSignatures[i] = config.LinkedTypeLabels[0]
	}
	var size uint64 = 2 * count * 32
	for _, unmetered := range []bool{false, true} {
		config.Unmetered, config.NativeAllocations = unmetered, unmetered
		body, err := codegen.CompileFunctionWithConfig(runtime.GOARCH, m.Functions[0], codegen.DefaultLimits(), config)
		if err != nil || len(body.DispatchRelocations) != 1 {
			t.Fatal(body, err)
		}
		size += uint64(len(body.Entry.Code)) * count
	}
	limits := codegen.DefaultLimits()
	limits.MaxCodeBytes = size
	c, err := Compile(context.Background(), m, limits)
	if err != nil {
		t.Fatal("shared rows did not fit exact image budget", size, err)
	}
	c.Close()
	if !c.linked || size > count*4096 {
		t.Fatal("indirect metadata is not linear", c.linked, size)
	}
	limits.MaxCodeBytes--
	if c, err := compileModule(context.Background(), m, limits, true); !errors.Is(err, wasm.ErrLimit) {
		if c != nil {
			c.Close()
		}
		t.Fatal("shared table escaped aggregate code budget", err)
	}
	// Linking still yields to complete ordinary tiers under a smaller quota.
	c, err = Compile(context.Background(), m, limits)
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	if c.linked {
		t.Fatal("expected bounded ordinary fallback")
	}
}

func TestLinkedIndirectRetargetingFallbacksAndQuotas(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	d := &module.Module{Types: []module.FuncType{sig, {Results: []wasm.ValueType{wasm.I64}}},
		Tables:    []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}},
		Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x41}, {Op: 0x11}, {Op: 0xb}}}}}
	for i := uint64(1); i <= 64; i++ {
		d.Functions = append(d.Functions, module.Function{Body: module.Expr{{Op: 0x20}, {Op: 0x41, Bits: i}, {Op: 0x6a}, {Op: 0xb}}})
	}
	d.Functions = append(d.Functions, module.Function{Type: 1, Body: module.Expr{{Op: 0x42}, {Op: 0xb}}})
	m := callSlots(t, d)
	for _, metered := range []bool{false, true} {
		s := state.NewNative(state.DefaultLimits())
		t.Cleanup(s.Close)
		in, err := s.Allocate(m, nil)
		if err != nil {
			t.Fatal(err)
		}
		other, err := s.Allocate(m, nil)
		if err != nil {
			t.Fatal(err)
		}
		host, err := s.NewHost(sig, "echo")
		if err != nil {
			t.Fatal(err)
		}
		limits := execute.Limits{MaxFrames: 128, MaxSlots: 2048}
		e, n := callExecution(t, s, m, in, metered, limits)
		for _, target := range []struct {
			ref, want uint64
			host      bool
			trap      wasm.TrapCode
		}{
			{ref: in.Functions[64], want: 71},
			{ref: in.Functions[63], want: 70},
			{ref: in.Functions[1], want: 8},
			{trap: wasm.TrapUninitializedElement},
			{ref: in.Functions[65], trap: wasm.TrapIndirectCallType},
			{ref: host.Function, host: true, want: 20},
			{ref: other.Functions[64], want: 71},
			{ref: in.Functions[64], want: 71},
		} {
			in.Tables[0].Values[0] = target.ref
			if err := e.Start(in.Functions[0], []value.Value{{Lo: 7}}); err != nil {
				t.Fatal(err)
			}
			status := e.Run()
			if target.host {
				if status != execute.HostRequest || e.Host().Args[0].Lo != 7 {
					t.Fatal("lost imported target", status, e.Err())
				}
				if err := e.ResumeHost([]value.Value{{Lo: target.want}}, nil); err != nil {
					t.Fatal(err)
				}
				status = e.Run()
			}
			if target.trap != wasm.NoTrap {
				var trap *wasm.Trap
				if status != execute.Failed || !errors.As(e.Err(), &trap) || trap.Code != target.trap {
					t.Fatal("changed indirect trap", target, status, e.Err())
				}
			} else if status != execute.Done || e.Results()[0].Lo != target.want {
				t.Fatal("changed indirect target", metered, target, status, e.Err(), e.Results())
			}
		}
		if n.callWords[nativeabi.CallHeaderBytes/8+nativeabi.CallResumeOffset/8] == 0 {
			t.Fatal("the distant indirect target never used generated frames")
		}
		for _, limited := range []execute.Limits{{MaxFrames: 1, MaxSlots: 2048}, {MaxFrames: 128, MaxSlots: uint32(len(m.Functions[0].Slots))}} {
			if err := e.SetLimits(limited); err != nil {
				t.Fatal(err)
			}
			if err := e.Start(in.Functions[0], []value.Value{{Lo: 7}}); err != nil || e.Run() != execute.Failed || !errors.Is(e.Err(), wasm.ErrLimit) {
				t.Fatal("indirect call exceeded invocation quota", metered, limited, err, e.Err())
			}
		}
	}
}
