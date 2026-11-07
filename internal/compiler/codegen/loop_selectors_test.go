package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func liveLiteralLoopFixture(bits uint64) *slots.Function {
	f := &slots.Function{LocalCount: 3, Slots: make([]wasm.ValueType, 8)}
	for i := range f.Slots {
		f.Slots[i] = wasm.I32
		f.Args = append(f.Args, slots.Slot(i))
	}
	emit := func(op uint32, dst, a, b slots.Slot, local uint32, bits uint64) {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Bits: bits, Offset: uint64(source) + 100}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	emit(0x41, 4, slots.NoSlot, slots.NoSlot, 0, 1)
	emit(0x6a, 3, 0, 4, 0, 0)
	emit(0x21, slots.NoSlot, 3, slots.NoSlot, 0, 0)
	emit(0x41, 5, slots.NoSlot, slots.NoSlot, 0, bits)
	emit(0x6a, 6, 3, 5, 0, 0)
	emit(0x21, slots.NoSlot, 6, slots.NoSlot, 2, 0)
	emit(0x41, 4, slots.NoSlot, slots.NoSlot, 0, 3)
	emit(0x49, 7, 0, 4, 0, 0)
	emit(slots.ControlBase+uint32(ir.If), slots.NoSlot, 7, slots.NoSlot, 0, 0)
	f.Code[len(f.Code)-1].Count = 2
	f.Edges = []slots.Edge{{PC: 0}, {PC: uint32(len(f.Code))}}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
	return f
}

func TestNativeLoopLiveLiteralSelectionAndEveryOriginalFrame(t *testing.T) {
	for _, bits := range []uint64{0x123, 0x12345678} {
		checkLiveLiteralLoop(t, bits)
	}
}

func checkLiveLiteralLoop(t *testing.T, bits uint64) {
	t.Helper()
	f := liveLiteralLoopFixture(bits)
	for _, arch := range []string{"amd64", "arm64"} {
		compiled, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
		if err != nil {
			t.Fatal(arch, bits, err)
		}
		if compiled.NativeLoops != 1 || (bits == 0x123 || arch == "amd64") && compiled.FusedIntegers == 0 {
			t.Fatal(arch, bits, "native region lost live-literal selection", compiled.NativeLoops, compiled.FusedIntegers)
		}
	}
	if !native.Available {
		return
	}
	compiled, lease, ctx := functionContext(t, f, true)
	_, reference, ref := functionContext(t, f, false)
	patched, err := native.Publish(forcePrivateLoopPoll(t, compiled, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(patched.Close)
	pollLease, err := patched.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pollLease.Close)
	for start := range f.Code {
		for _, poll := range []bool{false, true} {
			if poll && start != 0 {
				continue
			}
			for slot := range f.Slots {
				v := value.Value{Lo: uint64(slot) + 1}
				if slot == 0 {
					v.Lo = 0
				}
				writeSlot(t, ctx, 16, slots.Slot(slot), v)
				writeSlot(t, ref, 16, slots.Slot(slot), v)
			}
			state := native.State{Used: 16, Depth: 1, Budget: 1000, Continuation: uint64(start)}
			entry := lease
			if poll {
				state.Budget = uint64(len(f.Code) - 2)
				entry = pollLease
			}
			ctx.SetState(state)
			ref.SetState(state)
			got, err := ctx.Enter(entry, 0)
			want, werr := ref.Enter(reference, 0)
			gs, gerr := ctx.State()
			ws, serr := ref.State()
			if err != nil || werr != nil || gerr != nil || serr != nil || got != want || gs.Continuation != ws.Continuation || gs.Payload[0] != ws.Payload[0] {
				t.Fatal(start, poll, got, want, gs, ws, err, werr, gerr, serr)
			}
			for slot := range f.Slots {
				if got, want := readSlot(t, ctx, 16, slots.Slot(slot)), readSlot(t, ref, 16, slots.Slot(slot)); got != want {
					t.Fatal(start, poll, slot, got, want)
				}
			}
		}
	}
}
