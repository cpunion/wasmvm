package codegen

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func frameIntegerKernel(t *testing.T, op uint32, dst slots.Slot, mask int, aliases bool) nativeabi.Entry {
	t.Helper()
	a := assembler{arch: "amd64"}
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64, wasm.I64, wasm.I64}}
	e := functionEmitter{a: &a, f: f, unmetered: true, exit: a.label()}
	e.contextLoad(0, nativeabi.ArenaOffset)
	e.registerMove(6, 0, true)
	plan := &registerPlan{homes: map[slots.Slot]byte{}}
	pool := []byte{3, 12, 15, 9, 10}
	for s, reg := range pool {
		if mask&(1<<s) != 0 {
			plan.homes[slots.Slot(s)] = reg
			plan.slots = append(plan.slots, slots.Slot(s))
		}
	}
	e.transitionRegisters(plan)
	if aliases {
		e.frameAliases = map[slots.Slot]slots.Slot{0: 3, 1: 4}
	}
	if !e.frameInteger(op, dst, 0, 1) {
		t.Fatal("supported frame arithmetic rejected", op, dst, mask, aliases)
	}
	e.frameAliases = nil
	e.leave(nativeabi.Return, 0, 17)
	a.mark(e.exit)
	body, err := a.finish()
	if err != nil {
		t.Fatal(err)
	}
	entry, err := nativeabi.Wrap("amd64", body, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestFrameIntegerResidencyAliasesAndCompleteWords(t *testing.T) {
	for _, op := range []uint32{0x6a, 0x6b, 0x6c, 0x71, 0x72, 0x73, 0x7c, 0x7d, 0x7e, 0x83, 0x84, 0x85} {
		for _, dst := range []slots.Slot{0, 1, 2} {
			for _, aliases := range []bool{false, true} {
				masks := []int{0, 1, 2, 3, 4, 5, 6, 7}
				if aliases {
					masks = []int{0, 1, 2, 4, 8, 16, 24, 25, 26, 28, 31}
				}
				for _, mask := range masks {
					t.Run(fmt.Sprintf("%x/dst%d/%t/%d", op, dst, aliases, mask), func(t *testing.T) {
						entry := frameIntegerKernel(t, op, dst, mask, aliases)
						if !native.Available || runtime.GOARCH != "amd64" {
							return
						}
						code, err := native.Publish(entry.Code, []uint32{0}, 1<<16)
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(code.Close)
						lease, err := code.Acquire()
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(lease.Close)
						ctx, err := native.NewContext(80, 80, 1)
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(ctx.Close)
						edges := []uint64{0, 1, 0x80000000, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0), 0x123456789abcdef0}
						for _, left := range edges {
							for _, right := range edges {
								initial := []value.Value{{Lo: left, Hi: 11}, {Lo: right, Hi: 22}, {Lo: 77, Hi: 33}, {Lo: left, Hi: 44}, {Lo: right, Hi: 55}}
								if aliases {
									initial[0].Lo, initial[1].Lo = ^left, ^right
								}
								for s, v := range initial {
									writeSlot(t, ctx, 0, slots.Slot(s), v)
								}
								if err := ctx.SetState(native.State{Depth: 1, Fuel: 42, Budget: 37}); err != nil {
									t.Fatal(err)
								}
								exit, err := ctx.Enter(lease, 0)
								state, stateErr := ctx.State()
								if err != nil || stateErr != nil || exit != native.Return || state.Continuation != 17 || state.Fuel != 42 || state.Budget != 37 {
									t.Fatal(exit, state, err, stateErr)
								}
								want, trap := numeric.Eval(op, left, right)
								if trap != wasm.NoTrap {
									t.Fatal(trap)
								}
								initial[dst].Lo = want
								for s, v := range initial {
									if got := readSlot(t, ctx, 0, slots.Slot(s)); got != v {
										t.Fatal(left, right, "word", s, got, v)
									}
								}
							}
						}
					})
				}
			}
		}
	}
}

func TestFrameIntegerRejectsWithoutEmission(t *testing.T) {
	for _, mode := range []string{"metered", "arm64", "divide", "shift", "no-input", "float", "vector"} {
		a := assembler{arch: "amd64"}
		e := functionEmitter{a: &a, unmetered: true, registers: map[slots.Slot]byte{0: 3}}
		op, left := uint32(0x6a), slots.Slot(0)
		switch mode {
		case "metered":
			e.unmetered = false
		case "arm64":
			a.arch = "arm64"
		case "divide":
			op = 0x6d
		case "shift":
			op = 0x74
		case "no-input":
			left = slots.NoSlot
		case "float":
			e.floatRegisters = map[slots.Slot]byte{1: 3}
		case "vector":
			e.vectorRegisters = map[slots.Slot]byte{1: 3}
		}
		if e.canRegisterScalar(slots.Instruction{Op: op, Dst: 2, A: left, B: 1}, 0) || e.frameInteger(op, 2, left, 1) || len(a.code) != 0 {
			t.Fatal(mode, "fallback emitted partial code")
		}
	}
}
