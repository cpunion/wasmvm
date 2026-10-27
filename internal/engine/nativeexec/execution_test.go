package nativeexec

import (
	"context"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/codegen"
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func fixture(t *testing.T, largeEdge bool) *slots.Module {
	t.Helper()
	d := &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}}}}
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
	if largeEdge {
		f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}, Args: []slots.Slot{1}, Sources: []ir.Source{{Wasm: module.Instr{Op: 0x41, Bits: 42}}},
			Code: []slots.Instruction{
				{Op: 0x41, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
				{Op: slots.ControlBase + uint32(ir.Jump), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
				{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1},
			}, Edges: []slots.Edge{{PC: 2, Count: 300}}}
		for range 300 {
			f.Moves = append(f.Moves, slots.Move{Src: 0, Dst: 1})
		}
		m.Functions[0] = f
	}
	return m
}

func prepare(t *testing.T, m *slots.Module) (*Execution, *state.Function, *Lease) {
	t.Helper()
	c, err := Compile(context.Background(), m, codegen.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	l, err := c.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	t.Cleanup(l.Close)
	s := state.NewNative(state.DefaultLimits())
	t.Cleanup(s.Close)
	in, err := s.Allocate(m, nil)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := s.Function(in.Functions[0])
	e, err := New(context.Background(), map[*slots.Module]*Lease{m: l}, 16, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	if _, err = e.Resize(len(f.Code.Slots), len(f.Code.Slots)); err != nil {
		t.Fatal(err)
	}
	return e, f, l
}

func TestNativeFrameGrowthAndLargeEdges(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	m := fixture(t, true)
	e, fn, _ := prepare(t, m)
	for _, capacity := range []int{2, 8, 16} {
		if _, err := e.Resize(2, capacity); err != nil {
			t.Fatal(err)
		}
		pc, _, err := e.Run(fn, 0, 1, 0, 0)
		if err != nil || pc != 2 || e.values[1].Lo != 42 {
			t.Fatal(pc, e.values, err)
		}
	}
	for _, pair := range [][2]int{{-1, 1}, {2, 1}, {1, 17}} {
		if _, err := e.Resize(pair[0], pair[1]); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	e.Close()
	if _, err := e.Resize(1, 1); !errors.Is(err, native.ErrClosed) {
		t.Fatal(err)
	}
}

func TestNativeRejectsStaleOrMismatchedArtifacts(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, kind := range []string{"module", "function", "lease", "arena", "frame", "pc", "exit", "edge", "go-memory", "retired-memory"} {
		t.Run(kind, func(t *testing.T) {
			m := fixture(t, true)
			e, fn, l := prepare(t, m)
			base, want := 0, error(wasm.ErrInvalid)
			switch kind {
			case "module":
				delete(e.modules, m)
				want = wasm.ErrUnsupported
			case "function":
				delete(l.entries, fn.Code)
			case "lease":
				l.Close()
				want = native.ErrClosed
			case "arena":
				base = 3
				want = native.ErrLimit
			case "frame":
				base = 1
			case "pc":
				fn.Code.Code = fn.Code.Code[:1]
			case "exit":
				fn.Code.Code[2].Op = 0x41
			case "edge":
				fn.Code.Edges = nil
			case "go-memory", "retired-memory":
				s := state.New(state.DefaultLimits())
				if kind == "retired-memory" {
					s = state.NewNative(state.DefaultLimits())
				}
				defer s.Close()
				ex, err := s.NewMemory(module.Limits{Min: 1})
				if err != nil {
					t.Fatal(err)
				}
				fn.Instance.Memories = []*state.Memory{ex.Memory}
				if kind == "retired-memory" {
					ex.Memory.Native().Close()
					want = native.ErrClosed
				}
			}
			if _, _, err := e.Run(fn, base, 1, 0, 0); !errors.Is(err, want) {
				t.Fatal(err, want)
			}
		})
	}
}

func TestNativeCompileLimitsAndRetirement(t *testing.T) {
	m := fixture(t, false)
	if !native.Available {
		if _, err := Compile(context.Background(), m, codegen.DefaultLimits()); !errors.Is(err, wasm.ErrUnsupported) {
			t.Fatal(err)
		}
		if _, err := New(context.Background(), nil, 0, 0, false); !errors.Is(err, native.ErrUnavailable) {
			t.Fatal(err)
		}
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Compile(ctx, m, codegen.DefaultLimits()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	limits := codegen.DefaultLimits()
	limits.MaxCodeBytes = 0
	if _, err := Compile(context.Background(), m, limits); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	c, err := Compile(context.Background(), m, codegen.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	if _, err = c.Acquire(); !errors.Is(err, native.ErrClosed) {
		t.Fatal(err)
	}
}
