package codegen

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func guardedScaledLoopFixture(t *testing.T) *slots.Function {
	t.Helper()
	raw, err := os.ReadFile("../../../benchmarks/native/memory-fixtures/dependent-address.wasm")
	if err != nil {
		t.Fatal(err)
	}
	d, err := decode.Decode(raw, decode.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	f := functionModule(t, d).Functions[0]
	e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, unmetered: true, usesMemory: true, memoryMinimum: 65536}
	for start, r := range e.addNativeLoops(e.memoryRegions()) {
		if r.loop == nil || start+5 > r.end || f.Code[start].Op != 0x20 || f.Code[start+2].Op != 0x71 {
			continue
		}
		f.Code = append([]slots.Instruction(nil), f.Code[start:r.end+2]...)
		f.Edges = []slots.Edge{{PC: 0}, {PC: uint32(len(f.Code))}}
		f.Code[len(f.Code)-1].Extra = 0
		f.Args, f.Signature.Results = nil, nil
		for i, typ := range f.Slots {
			f.Args = append(f.Args, slots.Slot(i))
			f.Signature.Results = append(f.Signature.Results, typ)
		}
		f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: f.Code[len(f.Code)-1].Source,
			Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: uint32(len(f.Args))})
		return f
	}
	t.Fatal("missing indexed loop")
	return nil
}

func TestLoopScaledMemoryRequiresGuardedInductionAndRecoverableWords(t *testing.T) {
	for _, mode := range []string{"valid", "non-low-mask", "unbounded-backedge", "live-product", "local-mask", "same-temporaries", "missing-index", "shared-home", "views", "copies", "versions", "arm64", "metered", "unproven", "remapped"} {
		f := guardedScaledLoopFixture(t)
		e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, unmetered: true, usesMemory: true, memoryMinimum: 65536,
			registers: map[slots.Slot]byte{4: 9}}
		want := mode == "valid"
		switch mode {
		case "non-low-mask":
			f.Sources[f.Code[1].Source].Wasm.Bits = 4094
		case "unbounded-backedge":
			f.Code[7].Op = 0x6a
		case "live-product":
			f.Code[6].Op, f.Code[6].Dst = 1, slots.NoSlot
		case "local-mask":
			f.Code[1].Dst, f.Code[2].Dst, f.Code[2].B, f.Code[4].A, f.Code[4].Dst, f.Code[5].A = 4, 4, 4, 4, 4, 4
		case "same-temporaries":
			f.Code[3].Dst, f.Code[4].B, f.Code[5].Dst = f.Code[2].Dst, f.Code[2].Dst, f.Code[2].Dst
		case "missing-index":
			e.registers = nil
		case "shared-home":
			e.registers[6] = 9
		case "views":
			e.memoryViews = true
		}
		r := e.addNativeLoops(e.memoryRegions())[0]
		switch mode {
		case "copies":
			r.copies = &loopFrameCopies{}
		case "versions":
			r.values = &loopValuePlans{}
		case "arm64":
			e.a.arch = "arm64"
		case "metered":
			e.unmetered = false
		case "unproven":
			r.proven = nil
		case "remapped":
			r.addresses = map[int]memoryAddress{5: {local: 4}}
		}
		if got := e.planLoopScaledMemory(0, r); (got != nil) != want {
			t.Fatal(mode, got, want)
		}
	}
}

func TestLoopScaledMemoryEveryEntryFullFramesAndCancellation(t *testing.T) {
	for _, factor := range []uint64{2, 4, 8} {
		f := guardedScaledLoopFixture(t)
		f.Sources[f.Code[3].Source].Wasm.Bits = factor
		for _, arch := range []string{"amd64", "arm64"} {
			compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			if err != nil {
				t.Fatal(arch, err)
			}
			if arch == "amd64" && compiled.LoopGuards != 1 || arch == "arm64" && compiled.LoopGuards != 0 {
				t.Fatal("guarded scaled loop not selected as declared", arch, compiled.LoopGuards)
			}
		}
		if !native.Available {
			continue
		}
		compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
		_, reference, ref := configuredContext(t, f, Config{Unmetered: true})
		memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
		initial := make([]byte, 65536)
		for i := range initial {
			initial[i] = byte(i*37 + 193)
		}
		memory.Write(0, initial)
		refMemory.Write(0, initial)
		initialize := func(raw uint64) {
			for s := range f.Slots {
				bits := raw ^ uint64(s*17)
				switch s {
				case 1:
					bits = 5
				case 2:
					bits = 0
				case 4:
					bits = raw
				}
				writeSlot(t, ctx, 16, slots.Slot(s), value.Value{Lo: bits})
				writeSlot(t, ref, 16, slots.Slot(s), value.Value{Lo: bits})
			}
		}
		compare := func() {
			for s := range f.Slots {
				if a, b := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); a != b {
					t.Fatal("loop range frame", factor, s, a, b)
				}
			}
		}
		for pc := range f.Code {
			for _, raw := range []uint64{0, 4095, 4096, 0xffffffff, 0xdeadbeef00000042} {
				initialize(raw)
				state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 100000}
				wantState := state
				got, err := ctx.RunState(lease, 0, &state)
				want, refErr := ref.RunState(reference, 0, &wantState)
				if err != nil || refErr != nil || got != want || state != wantState {
					t.Fatal(factor, pc, raw, got, want, state, wantState, err, refErr)
				}
				compare()
			}
		}
		// A generated backedge exit must publish the complete original frame.
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
		once := *f
		once.Edges = []slots.Edge{{PC: uint32(len(f.Code) - 1)}, {PC: uint32(len(f.Code) - 1)}}
		_, onceLease, _ := configuredContext(t, &once, Config{Unmetered: true})
		initialize(17)
		state := native.State{Used: 16, Depth: 1, Budget: 100000}
		wantState := state
		got, err := ctx.RunState(pollLease, 0, &state)
		want, refErr := ref.RunState(onceLease, 0, &wantState)
		if err != nil || refErr != nil || got != native.Poll || want != native.Return {
			t.Fatal("generated loop range poll", got, want, err, refErr)
		}
		compare()
		// Entry cancellation takes precedence over a failed range guard.
		initialize(0xdeadbeef00000042)
		ctx.Cancel()
		ref.Cancel()
		state, wantState = native.State{Used: 16, Depth: 1, Budget: 100000}, native.State{Used: 16, Depth: 1, Budget: 100000}
		got, err = ctx.RunState(lease, 0, &state)
		want, refErr = ref.RunState(reference, 0, &wantState)
		if err != nil || refErr != nil || got != want || state != wantState {
			t.Fatal("range guard changed entry cancellation", got, want, state, wantState, err, refErr)
		}
		compare()
		data := make([]byte, 65536)
		memory.Read(0, data)
		if !bytes.Equal(data, initial) {
			t.Fatal("range loop changed memory")
		}
	}
}
