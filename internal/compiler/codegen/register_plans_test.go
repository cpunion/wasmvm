package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func pressuredLoopFixture() *slots.Function {
	f := affineLoopFixture(4, 0)
	f.Sources[10].Wasm.Bits = 256
	for len(f.Slots) < 70 {
		f.Slots = append(f.Slots, wasm.I32)
	}
	f.Slots[64], f.Slots[65], f.Slots[66] = wasm.F64, wasm.V128, wasm.ExternRef
	source := uint32(len(f.Sources))
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x6a, Offset: 999}})
	end := f.Code[len(f.Code)-1]
	f.Code = f.Code[:len(f.Code)-1]
	for s := slots.Slot(6); s < 58; s++ {
		for i := 0; i < 20; i++ {
			f.Code = append(f.Code, slots.Instruction{Op: 0x6a, Source: source, Dst: s, A: s, B: 63, C: slots.NoSlot})
		}
	}
	end.Count = uint32(len(f.Slots))
	f.Code = append(f.Code, end)
	f.Args = nil
	for s := range f.Slots {
		f.Args = append(f.Args, slots.Slot(s))
	}
	return f
}

func TestLoopRegisterPlansPressureColdEntriesAndCanonicalTuples(t *testing.T) {
	f := pressuredLoopFixture()
	for _, arch := range []string{"amd64", "arm64"} {
		c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
		if err != nil {
			t.Fatal(arch, err)
		}
		if c.RegisterRegions == 0 && c.AllocationBlocks < 2 {
			t.Fatal(arch, "neither loop nor global allocation selected")
		}
	}
	var entries []int
	for pc := 0; pc <= 14; pc++ {
		entries = append(entries, pc)
	}
	entries = append(entries, len(f.Code)-1)
	compareAffineLoop(t, f, entries...)
}

func TestRegisterPlanDeferredPollsAndTransitionsKeepEveryWord(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, where := range []string{"before", "inside", "return", "trap", "call"} {
			a := assembler{arch: arch, longBranches: true}
			f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.V128}}
			e := functionEmitter{a: &a, f: f, unmetered: true, exit: a.label()}
			e.contextLoad(0, nativeabi.ArenaOffset)
			frame := byte(6)
			if arch == "arm64" {
				frame = 12
			}
			e.registerMove(frame, e.operandRegister(0), true)
			pool := e.integerRegisterPool()
			old := &registerPlan{homes: map[slots.Slot]byte{0: pool[0], 1: pool[1], 2: pool[2]}, slots: []slots.Slot{0, 1, 2}}
			region := &registerPlan{homes: map[slots.Slot]byte{1: pool[1], 3: pool[0], 4: pool[2]}, slots: []slots.Slot{1, 3, 4}}
			for _, s := range old.slots {
				e.slotLoad(s, 0, false, true)
				e.registerMove(old.homes[s], e.operandRegister(0), true)
			}
			e.useRegisterPlan(old)
			e.constant(0, 0xfedcba9876543210)
			e.slotStore(0, 0, false)
			if where == "before" {
				e.contextConst(nativeabi.ControlOffset, 1)
			}
			e.interrupt(11)
			e.transitionRegisters(region)
			for _, item := range []struct {
				slot slots.Slot
				bits uint64
			}{{1, 0x8000000000000001}, {3, 0x1122334455667788}, {4, 0xffffffffffffffff}} {
				e.constant(0, item.bits)
				e.slotStore(item.slot, 0, false)
			}
			if where == "trap" {
				a.branch(e.trapExit(wasm.TrapUnreachable, 777), false, 0)
			}
			if where == "call" {
				label := a.label()
				e.callExits = append(e.callExits, nativeCallExit{label, 22, 888, e.currentRegisterPlan()})
				a.branch(label, false, 0)
			}
			if where == "inside" {
				e.contextConst(nativeabi.ControlOffset, 1)
			}
			e.interrupt(22)
			e.transitionRegisters(old)
			e.leave(nativeabi.Return, 0, 33)
			e.emitPollExits()
			e.emitCallExits(Config{})
			e.emitTrapExits()
			a.mark(e.exit)
			body, err := a.finish()
			if err != nil {
				t.Fatal(err)
			}
			wrapped, err := nativeabi.Wrap(arch, body, 1<<16)
			if err != nil {
				t.Fatal(err)
			}
			if !native.Available || arch != runtime.GOARCH {
				continue
			}
			code, err := native.Publish(wrapped.Code, []uint32{0}, 1<<16)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(code.Close)
			lease, err := code.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			ctx, err := native.NewContext(96, 96, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			expected := make([]value.Value, 6)
			for s := range expected {
				expected[s] = value.Value{Lo: 0x9876543210000000 | uint64(s)}
				if s == 5 {
					expected[s].Hi = 0xaabbccddeeff0011
				}
				writeSlot(t, ctx, 0, slots.Slot(s), expected[s])
			}
			ctx.SetState(native.State{Depth: 1})
			exit, err := ctx.Enter(lease, 0)
			state, stateErr := ctx.State()
			want, pc := native.Poll, uint64(11)
			expected[0].Lo = 0xfedcba9876543210
			if where != "before" {
				expected[1].Lo = 0x8000000000000001
				expected[3].Lo = 0x1122334455667788
				expected[4].Lo = 0xffffffffffffffff
				pc = 22
			}
			if where == "return" {
				want, pc = native.Return, 33
			}
			if where == "trap" {
				want, pc = native.Trap, 0
				if state.Source != 777 || state.Payload[0] != uint64(wasm.TrapUnreachable) {
					t.Fatal(state)
				}
			}
			if where == "call" {
				want, pc = native.Call, 23
				if state.Source != 888 || state.Payload[0] != 22 {
					t.Fatal(state)
				}
			}
			if err != nil || stateErr != nil || exit != want || state.Continuation != pc {
				t.Fatal(arch, where, exit, state, err, stateErr)
			}
			for s, v := range expected {
				if got := readSlot(t, ctx, 0, slots.Slot(s)); got != v {
					t.Fatal(arch, where, s, got, v)
				}
			}
		}
	}
}

func TestLoopRegisterPlansKeepShortLoopPlacement(t *testing.T) {
	f := pressuredLoopFixture()
	f.Sources[10].Wasm.Bits = 2
	for _, arch := range []string{"amd64", "arm64"} {
		c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
		if err != nil || c.LoopGuards != 1 || c.RegisterRegions != 0 {
			t.Fatal(arch, c, err)
		}
	}
}
