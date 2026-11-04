package execute

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// This model exercises the shared boundary contract, not generated native code.
// Its relocating arena deliberately invalidates old views on every hidden push.
type batchFrameBackend struct {
	movingFrameBackend
	store   *state.Store
	metered bool
	batches int
	limits  Limits
}

func (b *batchFrameBackend) RunFrames(out *state.NativeFrames) error {
	b.limits = Limits{out.MaxFrames, out.MaxSlots}
	charge := func(packet *state.NativeFrames, fn *state.Function, pc uint32) error {
		if !b.metered {
			return nil
		}
		src := fn.Code.Sources[fn.Code.Code[pc].Source]
		cost := src.Fuel()
		if packet.Fuel < cost {
			return &wasm.Error{Kind: wasm.ErrLimit, Offset: src.Wasm.Offset, Detail: "execution fuel"}
		}
		packet.Fuel -= cost
		return nil
	}
	current := *out
	var frames []state.NativeFrame
	for {
		err := b.advance(&current, charge)
		out.Values, out.PC, out.Fuel = current.Values, current.PC, current.Fuel
		if err != nil {
			return err
		}
		fn, pc := current.Function, current.PC
		in := fn.Code.Code[pc]
		var callee *state.Function
		if in.Op == 0x10 {
			index := fn.Code.Sources[in.Source].Wasm.A
			callee, _ = b.store.Function(fn.Instance.Functions[index])
		}
		if callee == nil || callee.Code == nil {
			if len(frames) != 0 {
				out.Frames = append(frames, state.NativeFrame{Function: fn, Base: current.Base, PC: pc})
			}
			return nil
		}
		args := make([]value.Value, in.Count)
		for i, slot := range fn.Code.Args[in.Extra : in.Extra+in.Count] {
			args[i] = current.Values[current.Base+int(slot)]
		}
		frames = append(frames, state.NativeFrame{Function: fn, Base: current.Base, PC: pc + 1, CallPC: pc})
		base := current.Base + len(fn.Code.Slots)
		current.Values, err = b.Resize(base+len(callee.Code.Slots), base+len(callee.Code.Slots))
		if err != nil {
			return err
		}
		copy(current.Values[base:], args)
		current.Function, current.Base, current.PC = callee, base, 0
		b.batches++
	}
}

type failingBatchBackend struct {
	batchFrameBackend
	fail error
}

func (b *failingBatchBackend) RunFrames(out *state.NativeFrames) error {
	before := b.batches
	if err := b.batchFrameBackend.RunFrames(out); err != nil {
		return err
	}
	if b.batches != before && b.fail != nil {
		err := b.fail
		b.fail = nil
		return err
	}
	return nil
}

func TestNativeFrameBatchFailureKeepsRelocatedArenaReusable(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	in := instance(t, s, &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{
		{Body: module.Expr{{Op: 0x20}, {Op: 0x10, A: 1}, {Op: 0x20}, {Op: 0x6a}, {Op: 0xb}}},
		{Body: module.Expr{{Op: 0x20}, {Op: 0xb}}},
	}}, nil)
	sentinel := errors.New("failure after arena relocation")
	b := &failingBatchBackend{batchFrameBackend: batchFrameBackend{store: s}, fail: sentinel}
	e := New(s, nil, Config{Limits: DefaultLimits(), Backend: b})
	defer e.Close()
	if err := e.Start(in.Functions[0], []value.Value{{Lo: 7}}); err != nil || e.Run() != Failed || e.Err() != sentinel {
		t.Fatal(err, e.Err())
	}
	limits := Limits{MaxFrames: 3, MaxSlots: 100}
	if err := e.SetLimits(limits); err != nil {
		t.Fatal(err)
	}
	for _, arg := range []uint64{19, 31} {
		if err := e.Start(in.Functions[0], []value.Value{{Lo: arg}}); err != nil || e.Run() != Done || e.Results()[0].Lo != 2*arg || b.limits != limits {
			t.Fatal("lost arena ownership or invocation limits", arg, err, e.Err(), e.Results(), b.limits)
		}
	}
}

func TestNativeFrameBatchHostRootsTuplesFuelAndReuse(t *testing.T) {
	s := state.New(state.DefaultLimits())
	defer s.Close()
	types := []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128, wasm.FuncRef, wasm.ExternRef}
	sig := module.FuncType{Params: types, Results: types}
	host, err := s.NewHost(sig, "echo")
	if err != nil {
		t.Fatal(err)
	}
	m := &module.Module{Types: []module.FuncType{sig}, Imports: []module.Import{{Kind: 0}}}
	for _, target := range []uint32{2, 3, 0} {
		body := module.Expr{}
		for i := range types {
			body = append(body, module.Instr{Op: 0x20, A: uint32(i)})
		}
		body = append(body, module.Instr{Op: 0x10, A: target}, module.Instr{Op: 0xb})
		m.Functions = append(m.Functions, module.Function{Body: body})
	}
	in := instance(t, s, m, func(string, string) (state.Extern, bool) { return host, true })
	args := []value.Value{{Lo: 0xffffffff}, {Lo: 0xfedcba9876543210}, {Lo: 0x7f812345}, {Lo: 0x8000000000000000}, {Lo: 7, Hi: 19}, {}, {Lo: 99}}
	for fuel := uint64(0); fuel <= 32; fuel++ {
		var expectedStatus Status
		var expectedFuel uint64
		for _, batched := range []bool{false, true} {
			model := &batchFrameBackend{store: s, metered: true}
			config := Config{Limits: DefaultLimits(), Metered: true, Fuel: fuel}
			if batched {
				config.Backend = model
			}
			e := New(s, context.Background(), config)
			for repeat := 0; repeat < 2; repeat++ {
				if err := e.Start(in.Functions[1], args); err != nil {
					t.Fatal(err)
				}
				status := e.Run()
				if status == HostRequest {
					if !reflect.DeepEqual(e.Host().Args, args) || len(e.frames) != 3 {
						t.Fatal("lost hidden frame or tuple", fuel, batched, e.Host(), len(e.frames))
					}
					roots := 0
					e.RetainedValues(func(typ wasm.ValueType, v value.Value) {
						if typ == wasm.ExternRef && v.Lo == 99 {
							roots++
						}
					})
					if roots < 4 {
						t.Fatal("hidden caller roots missing", roots)
					}
					for _, f := range e.frames {
						if !f.suspended || f.pc != f.callPC+1 || f.function.Code.Code[f.callPC].Op != 0x10 {
							t.Fatal("lost suspended call PC", f)
						}
					}
					if err := e.ResumeHost(args, nil); err != nil {
						t.Fatal(err)
					}
					status = e.Run()
				}
				if e.nativeFrames.Function != nil || e.nativeFrames.Values != nil || e.nativeFrames.Frames != nil {
					t.Fatal("retained borrowed native frame packet")
				}
				if !batched {
					expectedStatus, expectedFuel = status, e.Fuel()
				} else if status != expectedStatus || e.Fuel() != expectedFuel || model.limits != config.Limits {
					t.Fatal("changed exact fuel or limits", fuel, status, e.Fuel(), expectedStatus, expectedFuel)
				}
				if status == Done {
					if !reflect.DeepEqual(e.Results(), args) || batched && model.batches < 2 {
						t.Fatal("lost tuple or skipped frame batching", batched, model.batches, e.Results())
					}
				} else if status != Failed || !errors.Is(e.Err(), wasm.ErrLimit) {
					t.Fatal(status, e.Err())
				}
			}
			e.Close()
		}
	}
}

func frameTransfer(t *testing.T) (*Execution, state.NativeFrames) {
	t.Helper()
	s := state.New(state.DefaultLimits())
	t.Cleanup(s.Close)
	in := instance(t, s, &module.Module{Types: []module.FuncType{{Results: []wasm.ValueType{wasm.I32}}}, Functions: []module.Function{
		{Body: module.Expr{{Op: 0x10, A: 1}, {Op: 0xb}}},
		{Body: module.Expr{{Op: 0x41, Bits: 42}, {Op: 0xb}}},
	}}, nil)
	e := New(s, nil, Config{Limits: DefaultLimits()})
	t.Cleanup(e.Close)
	if err := e.Start(in.Functions[0], nil); err != nil {
		t.Fatal(err)
	}
	root, _ := s.Function(in.Functions[0])
	child, _ := s.Function(in.Functions[1])
	base := len(root.Code.Slots)
	return e, state.NativeFrames{PC: 0, Frames: []state.NativeFrame{{Function: root, PC: 1}, {Function: child, Base: base}}, Values: make([]value.Value, base+len(child.Code.Slots))}
}

func TestNativeFrameBatchRejectsInvalidTransfersBeforeMutation(t *testing.T) {
	for _, name := range []string{"root", "root-base", "nil-child", "child-base", "child-pc", "final-pc", "parent-pc", "call-pc", "opcode", "source", "target-index", "target", "values", "depth", "slots"} {
		t.Run(name, func(t *testing.T) {
			e, batch := frameTransfer(t)
			want := wasm.ErrInvalid
			switch name {
			case "root":
				batch.Frames[0].Function = batch.Frames[1].Function
			case "root-base":
				batch.Frames[0].Base++
			case "nil-child":
				batch.Frames[1].Function = nil
			case "child-base":
				batch.Frames[1].Base++
			case "child-pc":
				batch.Frames[1].PC = ^uint32(0)
				batch.PC = ^uint32(0)
			case "final-pc":
				batch.PC++
			case "parent-pc":
				batch.Frames[0].PC = 0
			case "call-pc":
				batch.Frames[0].CallPC = ^uint32(0)
			case "opcode":
				batch.Frames[0].Function.Code.Code[0].Op = 0x11
			case "source":
				batch.Frames[0].Function.Code.Code[0].Source = ^uint32(0)
			case "target-index":
				batch.Frames[0].Function.Code.Sources[0].Wasm.A = ^uint32(0)
			case "target":
				batch.Frames[0].Function.Instance.Functions[1] = 0
			case "values":
				batch.Values = batch.Values[:len(batch.Values)-1]
			case "depth":
				e.config.Limits.MaxFrames = 1
				want = wasm.ErrLimit
			case "slots":
				e.config.Limits.MaxSlots = uint32(len(batch.Values) - 1)
				want = wasm.ErrLimit
			}
			before := e.frames[0]
			length := len(e.values)
			if err := e.adoptNativeFrames(&batch); !errors.Is(err, want) || len(e.frames) != 1 || e.frames[0].function != before.function || e.frames[0].pc != before.pc || len(e.values) != length {
				t.Fatal("accepted invalid/partial frame transfer", err, want, len(e.frames), len(e.values))
			}
		})
	}
	e, batch := frameTransfer(t)
	batch.Frames = nil
	batch.Values = batch.Values[:len(e.values)]
	batch.PC = ^uint32(0)
	if err := e.adoptNativeFrames(&batch); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("invalid frameless PC", err)
	}
	batch.PC = 0
	batch.Values = nil
	if err := e.adoptNativeFrames(&batch); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("invalid frameless value count", err)
	}
	e.frames = nil
	if err := e.adoptNativeFrames(&batch); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal("accepted transfer without a live root", err)
	}
}
