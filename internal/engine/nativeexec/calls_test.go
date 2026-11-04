package nativeexec

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func callSlots(t *testing.T, d *module.Module) *slots.Module {
	t.Helper()
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
	return m
}

func callExecution(t *testing.T, s *state.Store, m *slots.Module, in *state.Instance, metered bool, limits execute.Limits) (*execute.Execution, *Execution) {
	t.Helper()
	compiled, err := Compile(context.Background(), m, codegen.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(compiled.Close)
	if !compiled.linked {
		t.Fatal("fixture did not enable native call linking")
	}
	lease, err := compiled.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(lease.Close)
	n, err := New(context.Background(), map[*slots.Module]*Lease{m: lease}, 2048, 128, metered)
	if err != nil {
		t.Fatal(err)
	}
	// Reserve space so the test exercises native pushes before any common
	// frame-machine growth. Other cases explicitly test bounded exhaustion.
	if _, err = n.Resize(0, 2048); err != nil {
		t.Fatal(err)
	}
	e := execute.New(s, nil, execute.Config{Limits: limits, Backend: n, Metered: metered, Fuel: 1000})
	t.Cleanup(e.Close)
	return e, n
}

func TestGeneratedCallsPreserveLargeMixedTuplesAndHostRoots(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	types := []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.ExternRef, wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.ExternRef}
	results := append([]wasm.ValueType(nil), types...)
	for i := 0; i < len(results)/2; i++ {
		results[i], results[len(results)-1-i] = results[len(results)-1-i], results[i]
	}
	args := []value.Value{{Lo: 7}, {Lo: 0xfedcba9876543210}, {Lo: 0x7f812345}, {Lo: 0x8000000000000000}, {Lo: 7, Hi: 19}, {Lo: 99}, {Lo: 11}, {Lo: 17}, {Lo: 0xff812345}, {Lo: 0x7ff123456789abcd}, {Lo: 37, Hi: 41}, {Lo: 101}}
	for _, hostBoundary := range []bool{false, true} {
		d := &module.Module{Types: []module.FuncType{{Params: types, Results: results}}}
		if hostBoundary {
			d.Imports = []module.Import{{Kind: 0}}
		}
		for f := 0; f < 3; f++ {
			var body module.Expr
			for i := range types {
				index := i
				if f == 2 && !hostBoundary {
					index = len(types) - 1 - i
				}
				body = append(body, module.Instr{Op: 0x20, A: uint32(index)})
			}
			if f < 2 {
				target := uint32(f + 1)
				if hostBoundary {
					target++
				}
				body = append(body, module.Instr{Op: 0x10, A: target})
			} else if hostBoundary {
				body = append(body, module.Instr{Op: 0x10})
			}
			body = append(body, module.Instr{Op: 0xb})
			d.Functions = append(d.Functions, module.Function{Body: body})
		}
		m := callSlots(t, d)
		for _, metered := range []bool{false, true} {
			s := state.NewNative(state.DefaultLimits())
			t.Cleanup(s.Close)
			var resolve state.Resolver
			if hostBoundary {
				host, err := s.NewHost(d.Types[0], "tuple")
				if err != nil {
					t.Fatal(err)
				}
				resolve = func(string, string) (state.Extern, bool) { return host, true }
			}
			in, err := s.Allocate(m, resolve)
			if err != nil {
				t.Fatal(err)
			}
			e, n := callExecution(t, s, m, in, metered, execute.Limits{MaxFrames: 128, MaxSlots: 2048})
			index := 0
			if hostBoundary {
				index++
			}
			for repeat := 0; repeat < 2; repeat++ {
				args[0].Lo = uint64(7 + repeat)
				want := append([]value.Value(nil), args...)
				for i := 0; i < len(want)/2; i++ {
					want[i], want[len(want)-1-i] = want[len(want)-1-i], want[i]
				}
				if err := e.Start(in.Functions[index], args); err != nil {
					t.Fatal(err)
				}
				status := e.Run()
				if hostBoundary {
					if status != execute.HostRequest || !reflect.DeepEqual(e.Host().Args, args) || e.Footprint().MaxFrames != 3 {
						t.Fatal("lost native callers at host boundary", status, e.Err(), e.Footprint())
					}
					roots := 0
					e.RetainedValues(func(typ wasm.ValueType, v value.Value) {
						if typ == wasm.ExternRef && v.Lo == 99 {
							roots++
						}
					})
					if roots < 4 {
						t.Fatal("missing native caller roots", roots)
					}
					if err := e.ResumeHost(want, nil); err != nil {
						t.Fatal(err)
					}
					status = e.Run()
				}
				if status != execute.Done || !reflect.DeepEqual(e.Results(), want) {
					t.Fatal("native tuple return", hostBoundary, metered, status, e.Err(), e.Results(), want)
				}
				if len(n.callWords) == 0 || n.callWords[nativeabi.CallHeaderBytes/8+nativeabi.CallResumeOffset/8] == 0 {
					t.Fatal("no generated native push executed")
				}
				if n.callWords[0] != 0 {
					t.Fatal("retained native parents")
				}
				oracle := execute.New(s, nil, execute.Config{Limits: execute.Limits{MaxFrames: 128, MaxSlots: 2048}, Metered: metered, Fuel: 1000})
				if err := oracle.Start(in.Functions[index], args); err != nil {
					t.Fatal(err)
				}
				oracleStatus := oracle.Run()
				if hostBoundary {
					if oracleStatus != execute.HostRequest {
						t.Fatal(oracleStatus, oracle.Err())
					}
					if err := oracle.ResumeHost(want, nil); err != nil {
						t.Fatal(err)
					}
					oracleStatus = oracle.Run()
				}
				if oracleStatus != execute.Done || oracle.Fuel() != e.Fuel() || !reflect.DeepEqual(oracle.Results(), e.Results()) {
					t.Fatal("native calls changed exact fuel", metered, oracleStatus, oracle.Err(), oracle.Fuel(), e.Fuel())
				}
				oracle.Close()
			}
		}
	}
}

func TestGeneratedSelfRecursionResumesLargeMovesAndGrowsMetadata(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	m := callSlots(t, &module.Module{Types: []module.FuncType{sig}, Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0xb}}}}})
	f := &slots.Function{Signature: sig, LocalCount: 1, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32}, Args: []slots.Slot{1, 4},
		Sources: []ir.Source{{Wasm: module.Instr{Op: 0x20}}, {Wasm: module.Instr{Op: 0x45}}, {Wasm: module.Instr{Op: 0x41, Bits: 1}}, {Wasm: module.Instr{Op: 0x6b}}, {Wasm: module.Instr{Op: 0x10}}, {Wasm: module.Instr{Op: 0x41, Bits: 42}}, {Wasm: module.Instr{Op: 0x04}}}}
	makeIn := func(op uint32, dst, a, b slots.Slot, source uint32) slots.Instruction {
		return slots.Instruction{Op: op, Dst: dst, A: a, B: b, C: slots.NoSlot, Source: source}
	}
	f.Code = []slots.Instruction{
		makeIn(0x20, 1, slots.NoSlot, slots.NoSlot, 0), makeIn(0x45, 2, 1, slots.NoSlot, 1), makeIn(slots.ControlBase+uint32(ir.If), slots.NoSlot, 2, slots.NoSlot, 6),
		makeIn(0x41, 1, slots.NoSlot, slots.NoSlot, 5), makeIn(slots.ControlBase+uint32(ir.Jump), slots.NoSlot, slots.NoSlot, slots.NoSlot, 6),
		makeIn(0x20, 1, slots.NoSlot, slots.NoSlot, 0), makeIn(0x41, 3, slots.NoSlot, slots.NoSlot, 2), makeIn(0x6b, 1, 1, 3, 3),
		makeIn(0x10, slots.NoSlot, slots.NoSlot, slots.NoSlot, 4), makeIn(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 6), makeIn(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 6),
	}
	f.Code[2].Count = 2
	f.Code[4].Extra = 2
	f.Code[4].Count = 1
	f.Code[8].Count = 1
	for _, pc := range []int{9, 10} {
		f.Code[pc].Extra = 1
		f.Code[pc].Count = 1
	}
	f.Edges = []slots.Edge{{PC: 3}, {PC: 5}, {PC: 10, Count: 300}}
	for range 300 {
		f.Moves = append(f.Moves, slots.Move{Src: 1, Dst: 4})
	}
	m.Functions[0] = f
	for _, metered := range []bool{false, true} {
		s := state.NewNative(state.DefaultLimits())
		t.Cleanup(s.Close)
		in, err := s.Allocate(m, nil)
		if err != nil {
			t.Fatal(err)
		}
		e, n := callExecution(t, s, m, in, metered, execute.Limits{MaxFrames: 128, MaxSlots: 2048})
		if err := e.Start(in.Functions[0], []value.Value{{Lo: 40}}); err != nil || e.Run() != execute.Done || e.Results()[0].Lo != 42 {
			t.Fatal("recursive native move", metered, err, e.Err(), e.Results())
		}
		if n.callWords[1] < 32 {
			t.Fatal("metadata did not grow", n.callWords[1])
		}
		if err := e.SetLimits(execute.Limits{MaxFrames: 16, MaxSlots: 2048}); err != nil {
			t.Fatal(err)
		}
		if err := e.Start(in.Functions[0], []value.Value{{Lo: 40}}); err != nil || e.Run() != execute.Failed || !errors.Is(e.Err(), wasm.ErrLimit) {
			t.Fatal("native depth limit", err, e.Err())
		}
		if err := e.SetLimits(execute.Limits{MaxFrames: 128, MaxSlots: 2048}); err != nil {
			t.Fatal(err)
		}
		if err := e.Start(in.Functions[0], []value.Value{{Lo: 20}}); err != nil || e.Run() != execute.Done || e.Results()[0].Lo != 42 {
			t.Fatal("reuse after native depth exhaustion", err, e.Err(), e.Results())
		}
	}
}

func TestGeneratedCallPublishesParentRootsBeforeGuestCollection(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	d := &module.Module{Types: []module.FuncType{{}, {Results: []wasm.ValueType{wasm.I32}}, {}}, TypeDefs: []module.TypeDef{
		{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32}}}, {Kind: module.FunctionKind}, {Kind: module.FunctionKind}}, Functions: []module.Function{
		{Type: 1, Locals: []module.Local{{Count: 1, Type: wasm.IndexedReference(wasm.StructRef, 0, true)}}, LocalCount: 1, Body: module.Expr{{Op: 0x41, Bits: 79}, {Op: 0xfb0000}, {Op: 0x21}, {Op: 0x10, A: 1}, {Op: 0x20}, {Op: 0xfb0002}, {Op: 0xb}}},
		{Type: 2, Body: module.Expr{{Op: 0x41, Bits: 1}, {Op: 0xfb0000}, {Op: 0x1a}, {Op: 0x41, Bits: 2}, {Op: 0xfb0000}, {Op: 0x1a}, {Op: 0x41, Bits: 3}, {Op: 0xfb0000}, {Op: 0x1a}, {Op: 0xb}}},
	}}
	v, err := validate.CheckWithFeatures(d, validate.DefaultLimits(), wasm.Core3Features)
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
	for _, metered := range []bool{false, true} {
		limits := state.DefaultLimits()
		limits.MaxHeapObjects = 4
		s := state.NewNative(limits)
		t.Cleanup(s.Close)
		s.CollectEveryAllocation = true
		in, err := s.Allocate(m, nil)
		if err != nil {
			t.Fatal(err)
		}
		e, n := callExecution(t, s, m, in, metered, execute.Limits{MaxFrames: 128, MaxSlots: 2048})
		for range 2 {
			if err := e.Start(in.Functions[0], nil); err != nil || e.Run() != execute.Done || e.Results()[0].Lo != 79 {
				t.Fatal("lost parent-only GC root", metered, err, e.Err(), e.Results())
			}
			if len(n.callWords) == 0 || n.callWords[nativeabi.CallHeaderBytes/8+nativeabi.CallResumeOffset/8] == 0 {
				t.Fatal("GC fixture did not use a native call")
			}
		}
	}
}

func TestNativeCallMetadataValidationAndRetirement(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	m := callSlots(t, &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{
		{Body: module.Expr{{Op: 0x10, A: 1}, {Op: 0xb}}}, {Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}},
	}})
	s := state.NewNative(state.DefaultLimits())
	t.Cleanup(s.Close)
	in, err := s.Allocate(m, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, n := callExecution(t, s, m, in, false, execute.Limits{MaxFrames: 128, MaxSlots: 2048})
	root, _ := s.Function(in.Functions[0])
	child, _ := s.Function(in.Functions[1])
	callPC := uint64(0)
	for pc, op := range root.Code.Code {
		if op.Op == 0x10 {
			callPC = uint64(pc)
			break
		}
	}
	for _, name := range []string{"valid", "top", "limit", "depth", "parent-function", "parent-used", "parent-pc", "current-function", "current-used", "current-end", "slot-limit"} {
		if _, err := n.prepareCalls(2048, 128); err != nil {
			t.Fatal(err)
		}
		n.callWords[0] = 1
		word := nativeabi.CallHeaderBytes / 8
		n.callWords[word+nativeabi.CallFunctionOffset/8] = 0
		n.callWords[word+nativeabi.CallUsedOffset/8] = 0
		n.callWords[word+nativeabi.CallPCOffset/8] = callPC
		n.entryState = native.State{Function: 1, Used: uint64(len(root.Code.Slots)) * 16, Depth: 2}
		packet := state.NativeFrames{Function: root, Depth: 1, MaxFrames: 128, MaxSlots: 2048}
		switch name {
		case "top":
			n.callWords[0] = n.callWords[1] + 1
		case "limit":
			packet.MaxFrames = 1
		case "depth":
			n.entryState.Depth = 1
		case "parent-function":
			n.callWords[word+nativeabi.CallFunctionOffset/8] = 100
		case "parent-used":
			n.callWords[word+nativeabi.CallUsedOffset/8] = 1
		case "parent-pc":
			n.callWords[word+nativeabi.CallPCOffset/8] = uint64(len(root.Code.Code))
		case "current-function":
			n.entryState.Function = 100
		case "current-used":
			n.entryState.Used = 1
		case "current-end":
			n.entryState.Used = 2048 * 16
		case "slot-limit":
			packet.MaxSlots = uint32(len(root.Code.Slots))
		}
		err := n.publishCallFrames(&packet)
		if name == "valid" {
			if err != nil || len(packet.Frames) != 2 || packet.Frames[0].Function != root || packet.Frames[1].Function != child {
				t.Fatal(name, err, packet.Frames)
			}
		} else if !errors.Is(err, wasm.ErrInvalid) || packet.Frames != nil {
			t.Fatal("accepted corrupt native call metadata", name, err)
		}
	}
	retained := n.callFrames[:cap(n.callFrames)]
	if !n.Suspend() {
		t.Fatal("native metadata blocked suspension")
	}
	for _, f := range retained {
		if f.Function != nil {
			t.Fatal("retained invocation root")
		}
	}
	if len(n.callFrames) != 0 || n.callWords[0] != 0 {
		t.Fatal("retained native call state")
	}
	n.Close()
	if n.callData != nil || n.callWords != nil || n.callFrames != nil {
		t.Fatal("unreleased call metadata")
	}
}

type nativeCallCancelContext struct {
	context.Context
	polls int
}

func (c *nativeCallCancelContext) Err() error {
	c.polls++
	if c.polls >= 3 {
		return context.Canceled
	}
	return nil
}

func TestGeneratedCallCancellationDropsHiddenFrames(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	m := callSlots(t, &module.Module{Types: []module.FuncType{{}}, Functions: []module.Function{
		{Body: module.Expr{{Op: 0x10, A: 1}, {Op: 0xb}}},
		{Body: module.Expr{{Op: 3, Block: -64}, {Op: 0x41, Bits: 1}, {Op: 0x1a}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}},
	}})
	s := state.NewNative(state.DefaultLimits())
	t.Cleanup(s.Close)
	in, err := s.Allocate(m, nil)
	if err != nil {
		t.Fatal(err)
	}
	limits := execute.Limits{MaxFrames: 128, MaxSlots: 2048}
	e, n := callExecution(t, s, m, in, true, limits)
	ctx := &nativeCallCancelContext{Context: context.Background()}
	if !n.Suspend() {
		t.Fatal("initial suspension")
	}
	if err := n.Begin(ctx, true); err != nil {
		t.Fatal(err)
	}
	if err := e.Rebind(ctx, limits, true, 1000); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(in.Functions[0], nil); err != nil || e.Run() != execute.Failed || !errors.Is(e.Err(), context.Canceled) {
		t.Fatal("hidden-frame cancellation", err, e.Err())
	}
	if e.Footprint() != (execute.Limits{}) || n.callWords[0] != 0 || n.callWords[nativeabi.CallHeaderBytes/8+nativeabi.CallResumeOffset/8] == 0 {
		t.Fatal("did not enter or retire native call chain", e.Footprint(), n.callWords)
	}
}
