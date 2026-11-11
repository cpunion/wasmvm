package nativeexec

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCrossInstanceVeneerRequiresOwnedLiveArtifacts(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	veneer := callSlots(t, &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}},
		Tables:    []module.TableType{{Element: wasm.FuncRef, Limits: module.Limits{Min: 1}}},
		Elements:  []module.Element{{Type: wasm.FuncRef, Offset: module.Expr{{Op: 0x41}, {Op: 0xb}}, Functions: []uint32{0}}},
		Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x41}, {Op: 0x11}, {Op: 0xb}}}}})
	caller := callSlots(t, &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}},
		Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0xb}}}}})
	compile := func(m *slots.Module) *Lease {
		c, err := Compile(context.Background(), m, codegen.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		lease, err := c.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		c.Close() // The acquired lease remains an independent pin.
		t.Cleanup(lease.Close)
		return lease
	}
	cl, vl := compile(caller), compile(veneer)
	s := state.NewNative(state.DefaultLimits())
	t.Cleanup(s.Close)
	host, err := s.NewHost(sig, "echo")
	if err != nil {
		t.Fatal(err)
	}
	vin, err := s.Allocate(veneer, func(string, string) (state.Extern, bool) { return host, true })
	if err != nil {
		t.Fatal(err)
	}
	if err := vin.Initialize(); err != nil {
		t.Fatal(err)
	}
	in, err := s.Allocate(caller, func(string, string) (state.Extern, bool) {
		return state.Extern{Owner: s, Kind: 0, Function: vin.Functions[1]}, true
	})
	if err != nil {
		t.Fatal(err)
	}
	modules := map[*slots.Module]*Lease{caller: cl}
	n, err := New(context.Background(), modules, 64, 8, false)
	if err != nil {
		t.Fatal(err)
	}
	e := execute.New(s, nil, execute.Config{Backend: n, Limits: execute.Limits{MaxFrames: 8, MaxSlots: 64}})
	t.Cleanup(e.Close)
	args := []value.Value{{Lo: 77}}
	if err := e.Start(in.Functions[1], args); err != nil || e.Run() != execute.Failed || !errors.Is(e.Err(), wasm.ErrUnsupported) {
		t.Fatal("missing native artifact was bypassed", err, e.Err())
	}
	modules[veneer] = vl
	if err := e.Start(in.Functions[1], args); err != nil || e.Run() != execute.HostRequest || e.Host().Caller != vin {
		t.Fatal("retired code lost its held lease", err, e.Err())
	}
	vl.Close()
	if err := e.ResumeHost(args, nil); err != nil || e.Run() != execute.Failed || !errors.Is(e.Err(), native.ErrClosed) || e.Footprint() != (execute.Limits{}) {
		t.Fatal("closed artifact skipped continuation admission", err, e.Err())
	}
}
