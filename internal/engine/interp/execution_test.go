package interp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func instance(t *testing.T, s *state.Store, m *module.Module, resolve state.Resolver) *state.Instance {
	t.Helper()
	v, err := validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Compile(v, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	code, err := slots.Compile(ir, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	in, err := s.Allocate(code, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if err := in.Initialize(); err != nil {
		t.Fatal(err)
	}
	return in
}
func simple(body module.Expr, results ...wasm.ValueType) *module.Module {
	return &module.Module{Types: []module.FuncType{{Results: results}}, Functions: []module.Function{{Body: body}}}
}
func TestFuelBeforeEffectsAndReuse(t *testing.T) {
	s := state.New(state.DefaultLimits())
	m := simple(module.Expr{{Op: 0x41, Bits: 7, Offset: 10}, {Op: 0x24, Offset: 11}, {Op: 0x41, Bits: 9, Offset: 12}, {Op: 0x24, Offset: 13}, {Op: 0xb, Offset: 14}})
	m.Globals = []module.Global{{Type: module.GlobalType{Value: wasm.I32, Mutable: true}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}
	in := instance(t, s, m, nil)
	for fuel := uint64(0); fuel <= 4; fuel++ {
		in.Globals[0].Value = value.Value{}
		e := New(s, nil, Config{Limits: DefaultLimits(), Metered: true, Fuel: fuel})
		if err := e.Start(in.Functions[0], nil); err != nil {
			t.Fatal(err)
		}
		status := e.Run()
		want := uint64(0)
		if fuel >= 2 {
			want = 7
		}
		if fuel == 4 {
			want = 9
		}
		if in.Globals[0].Value.Lo != want || e.Fuel() != 0 {
			t.Fatal(fuel, e.Fuel(), in.Globals[0].Value)
		}
		if fuel < 4 {
			var limit *wasm.Error
			if status != Failed || !errors.As(e.Err(), &limit) || !errors.Is(limit, wasm.ErrLimit) || limit.Offset != 10+fuel {
				t.Fatal(status, e.Err())
			}
		} else if status != Done {
			t.Fatal(status, e.Err())
		}
	}
	constant := instance(t, s, simple(module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}, wasm.I32), nil)
	e := New(s, nil, Config{Limits: DefaultLimits()})
	if err := e.Start(0, nil); !errors.Is(err, state.ErrReference) {
		t.Fatal(err)
	}
	if err := e.Start(constant.Functions[0], []value.Value{{Lo: 1}}); err == nil {
		t.Fatal("argument count")
	}
	run := func() {
		if err := e.Start(constant.Functions[0], nil); err != nil {
			t.Fatal(err)
		}
		if e.Run() != Done || e.Results()[0].Lo != 42 {
			t.Fatal(e.Err(), e.Results())
		}
	}
	run()
	if got := testing.AllocsPerRun(100, run); got != 0 {
		t.Fatalf("warmed numeric execution allocated %g times", got)
	}
	if e.Run() != Done {
		t.Fatal("completed status changed")
	}
	for _, limits := range []Limits{{0, 100}, {100, 0}} {
		bad := New(s, nil, Config{Limits: limits})
		if err := bad.Start(constant.Functions[0], nil); !errors.Is(err, wasm.ErrLimit) || bad.Run() != Failed || len(bad.frames) != 0 {
			t.Fatal(err, bad.status)
		}
	}
}

func TestHostSuspendResumeRootsAndReentry(t *testing.T) {
	s := state.New(state.DefaultLimits())
	sig := module.FuncType{Params: []wasm.ValueType{wasm.ExternRef}, Results: []wasm.ValueType{wasm.ExternRef}}
	host, err := s.NewHost(sig, "echo")
	if err != nil {
		t.Fatal(err)
	}
	m := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}}, Functions: []module.Function{{Body: module.Expr{{Op: 0x20}, {Op: 0x10}, {Op: 0xb}}}}}
	in := instance(t, s, m, func(string, string) (state.Extern, bool) { return host, true })
	e := New(s, nil, Config{Limits: DefaultLimits()})
	arg := []value.Value{{Lo: 99}}
	if err := e.Start(in.Functions[1], arg); err != nil {
		t.Fatal(err)
	}
	arg[0].Lo = 5
	if err := e.Start(in.Functions[1], nil); err == nil {
		t.Fatal("active execution restarted")
	}
	if e.Run() != HostRequest || e.Run() != HostRequest || e.Host().Function.Host != "echo" || e.Host().Args[0].Lo != 99 {
		t.Fatal(e.Host(), e.Err())
	}
	roots := 0
	e.RetainedValues(func(typ wasm.ValueType, v value.Value) {
		if typ != wasm.ExternRef {
			t.Fatal(typ)
		}
		if v.Lo == 99 {
			roots++
		}
	})
	if roots < 2 {
		t.Fatal("missing suspended roots", roots)
	}
	// A driver can run a nested continuation while the outer host request stays
	// suspended. Its arguments/frames must not overwrite the outer continuation.
	nested := New(s, nil, Config{Limits: DefaultLimits()})
	if err := nested.Start(host.Function, []value.Value{{Lo: 7}}); err != nil {
		t.Fatal(err)
	}
	if nested.Run() != HostRequest {
		t.Fatal(nested.status)
	}
	if err := nested.ResumeHost([]value.Value{{Lo: 8}}, nil); err != nil || nested.Run() != Done || nested.Results()[0].Lo != 8 {
		t.Fatal(err)
	}
	if err := e.ResumeHost([]value.Value{{Lo: 123}}, nil); err != nil || e.Run() != Done || e.Results()[0].Lo != 123 {
		t.Fatal(err, e.Err(), e.Results())
	}
	if err := e.ResumeHost(nil, nil); err == nil {
		t.Fatal("resumed completed host call")
	}
	if err := e.Start(host.Function, []value.Value{{Lo: 2}}); err != nil {
		t.Fatal(err)
	}
	e.RetainedValues(func(typ wasm.ValueType, v value.Value) {
		if typ != wasm.ExternRef || v.Lo != 2 {
			t.Fatal(typ, v)
		}
	})
	if err := e.ResumeHost(nil, nil); err == nil || e.Run() != Failed || e.Host() != nil {
		t.Fatal(err, e.status)
	}
	if err := e.Start(in.Functions[1], []value.Value{{Lo: 1}}); err != nil {
		t.Fatal(err)
	}
	e.Run()
	sentinel := errors.New("host failure")
	if err := e.ResumeHost(nil, sentinel); err != sentinel || e.Err() != sentinel {
		t.Fatal(err)
	}
	if len(e.frames) != 0 || len(e.values) != 0 {
		t.Fatal("host failure leaked continuation")
	}
	if err := e.Start(in.Functions[1], []value.Value{{Lo: 1}}); err != nil {
		t.Fatal(err)
	}
	e.Run()
	func() {
		defer func() {
			if recover() != sentinel {
				t.Fatal("host panic changed")
			}
		}()
		defer e.Abort(sentinel)
		panic(sentinel)
	}()
	if e.Run() != Failed || len(e.frames) != 0 || e.Host() != nil {
		t.Fatal("panic cleanup")
	}
}

type cancelAfter struct{ calls, at int }

func TestSuspendedUnpublishedInstanceSurvivesCollection(t *testing.T) {
	s := state.New(state.DefaultLimits())
	host, _ := s.NewHost(module.FuncType{}, "pause")
	m := simple(module.Expr{{Op: 0x10}, {Op: 0x41, Bits: 42}, {Op: 0xb}}, wasm.I32)
	m.Types = append(m.Types, module.FuncType{})
	m.Imports = []module.Import{{Kind: 0, Type: 1}}
	in := instance(t, s, m, func(string, string) (state.Extern, bool) { return host, true })
	e := New(s, nil, Config{Limits: DefaultLimits()})
	if err := e.Start(in.Functions[1], nil); err != nil {
		t.Fatal(err)
	}
	if e.Run() != HostRequest {
		t.Fatal(e.Err())
	}
	var active []*state.Instance
	e.RetainedInstances(func(in *state.Instance) { active = append(active, in) })
	s.CollectFailed(nil, active)
	if s.Usage().Instances != 1 || len(active) != 1 {
		t.Fatal("lost suspended owner")
	}
	if err := e.ResumeHost(nil, nil); err != nil || e.Run() != Done || e.Results()[0].Lo != 42 {
		t.Fatal(err, e.Err())
	}
	active = nil
	e.RetainedInstances(func(in *state.Instance) { active = append(active, in) })
	s.CollectFailed(nil, active)
	if len(active) != 0 || s.Usage().Instances != 0 {
		t.Fatal("completed frame retained owner")
	}
}

func (c *cancelAfter) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *cancelAfter) Done() <-chan struct{}       { return nil }
func (c *cancelAfter) Value(any) any               { return nil }
func (c *cancelAfter) Err() error {
	c.calls++
	if c.calls >= c.at {
		return context.Canceled
	}
	return nil
}
func TestCancellationBoundaries(t *testing.T) {
	s := state.New(state.DefaultLimits())
	loop := instance(t, s, simple(module.Expr{{Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}}), nil)
	for _, at := range []int{1, 3} {
		ctx := &cancelAfter{at: at}
		e := New(s, ctx, Config{Limits: DefaultLimits()})
		e.Start(loop.Functions[0], nil)
		if e.Run() != Failed || e.Err() != context.Canceled || len(e.frames) != 0 {
			t.Fatal(at, e.Err())
		}
	}
	body := make(module.Expr, 200)
	for i := range body {
		body[i].Op = 1
	}
	body = append(body, module.Instr{Op: 0xb})
	straight := instance(t, s, simple(body), nil)
	ctx := &cancelAfter{at: 2}
	e := New(s, ctx, Config{Limits: DefaultLimits()})
	e.Start(straight.Functions[0], nil)
	if e.Run() != Failed || e.Err() != context.Canceled || e.steps != 64 {
		t.Fatal(e.steps, e.Err())
	}
	mem := simple(module.Expr{{Op: 0x41}, {Op: 0x41, Bits: 0x77}, {Op: 0x41, Bits: 8192}, {Op: 0xfc000b}, {Op: 0xb}})
	mem.Memories = []module.Limits{{Min: 1}}
	in := instance(t, s, mem, nil)
	e = New(s, &cancelAfter{at: 3}, Config{Limits: DefaultLimits()})
	e.Start(in.Functions[0], nil)
	if e.Run() != Failed || e.Err() != context.Canceled {
		t.Fatal(e.Err())
	}
	b := make([]byte, 8192)
	in.Memories[0].Read(0, b)
	for i, v := range b {
		want := byte(0)
		if i < 4096 {
			want = 0x77
		}
		if v != want {
			t.Fatalf("bulk cancellation byte %d = %d", i, v)
		}
	}
	// Table fill uses the same bounded polling contract with 512-element chunks.
	tab := simple(module.Expr{{Op: 0x41}, {Op: 0xd0, A: uint32(wasm.ExternRef)}, {Op: 0x41, Bits: 1024}, {Op: 0xfc0011}, {Op: 0xb}})
	tab.Tables = []module.TableType{{Element: wasm.ExternRef, Limits: module.Limits{Min: 1024}}}
	tin := instance(t, s, tab, nil)
	for i := range tin.Tables[0].Values {
		tin.Tables[0].Values[i] = 1
	}
	e = New(s, &cancelAfter{at: 3}, Config{Limits: DefaultLimits()})
	e.Start(tin.Functions[0], nil)
	if e.Run() != Failed || e.Err() != context.Canceled {
		t.Fatal(e.Err())
	}
	for i, v := range tin.Tables[0].Values {
		want := uint64(1)
		if i < 512 {
			want = 0
		}
		if v != want {
			t.Fatal(i, v)
		}
	}
}

func FuzzExecution(f *testing.F) {
	f.Add([]byte("\x00asm\x01\x00\x00\x00"))
	f.Add([]byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 4, 1, 0x60, 0, 0, 3, 2, 1, 0, 10, 7, 1, 5, 0, 0x41, 1, 0x1a, 0xb})
	f.Fuzz(func(t *testing.T, data []byte) {
		limits := decode.DefaultLimits()
		limits.MaxBytes = 1 << 16
		limits.MaxNodes = 4096
		limits.MaxLocals = 128
		limits.MaxNesting = 64
		m, err := decode.Decode(data, limits)
		if err != nil {
			return
		}
		validated, err := validate.Check(m, validate.Limits{MaxLocals: 128, MaxStack: 128, MaxControl: 65})
		if err != nil {
			return
		}
		lowered, err := lower.Compile(validated, lower.Limits{MaxNodes: 1 << 16, MaxVerifyWork: 1 << 20})
		if err != nil {
			return
		}
		code, err := slots.Compile(lowered, slots.Limits{MaxSlots: 2048, MaxNodes: 1 << 16, MaxVerifyWork: 1 << 20})
		if err != nil {
			return
		}
		s := state.New(state.Limits{MaxTypes: 256, MaxFunctions: 128, MaxGlobals: 128, MaxInstances: 4, MaxMemories: 4, MaxTables: 8, MaxMemoryPages: 4, MaxTableElements: 1024, MaxSegmentBytes: 1 << 16, MaxSegmentElements: 1024})
		in, err := s.Allocate(code, nil)
		if err != nil {
			if !errors.Is(err, state.ErrLink) && !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
			return
		}
		if err := in.Initialize(); err != nil {
			if !errors.Is(err, wasm.ErrTrap) {
				t.Fatal(err)
			}
			return
		}
		for _, ref := range in.Functions {
			fn, _ := s.Function(ref)
			e := New(s, nil, Config{Limits: Limits{MaxFrames: 32, MaxSlots: 8192}, Metered: true, Fuel: 1000})
			err := e.Start(ref, make([]value.Value, len(fn.Signature.Params)))
			if err == nil {
				status := e.Run()
				if status != Done && status != Failed {
					t.Fatal(status)
				}
				err = e.Err()
			}
			if err != nil && !errors.Is(err, wasm.ErrTrap) && !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
		}
	})
}
