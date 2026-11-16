package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFreshPrivateEntryLoadsFirstPlanWithoutContextDispatch(t *testing.T) {
	const want = uint64(0x87654321fedcba98)
	for _, arch := range []string{"amd64", "arm64"} {
		for _, allocated := range []bool{false, true} {
			a := assembler{arch: arch, longBranches: true}
			e := functionEmitter{a: &a, f: &slots.Function{Slots: []wasm.ValueType{wasm.I64}}, unmetered: true, frameRegisters: true, exit: a.label()}
			pool := e.integerRegisterPool()
			first := &registerPlan{homes: map[slots.Slot]byte{0: pool[0]}, slots: []slots.Slot{0}}
			initial := first
			if allocated {
				// The function-wide mapping is insufficient once allocation has
				// selected a distinct first block. The fresh entry must load it.
				initial = &registerPlan{homes: map[slots.Slot]byte{0: pool[1]}, slots: []slots.Slot{0}}
				e.allocationPlans = []*registerPlan{first, initial}
			}
			e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
			e.callEntryFrame(0)
			e.callConst(callUsed, 0)
			e.callConst(callDepth, 1)
			e.callConst(callTop, 0)
			if arch == "amd64" {
				a.xconst(pool[0], 0xdead)
				a.xconst(pool[1], 0xbeef)
			} else {
				a.aconst(uint32(pool[0]), 0xdead, true)
				a.aconst(uint32(pool[1]), 0xbeef, true)
			}
			pcs := []int{a.label(), a.label()}
			e.emitCallStart(pcs, initial)
			a.mark(pcs[0])
			e.registerMove(e.callReg(0), pool[0], true)
			e.callStore(0, e.callContext(), uint32(nativeabi.ResultLoOffset))
			e.rawStatus(nativeabi.Return)
			a.branch(e.exit, false, 0)
			a.mark(pcs[1])
			e.rawStatus(nativeabi.BadState)
			e.emitStatusExit()
			a.mark(e.exit)
			body, err := a.finish()
			if err != nil {
				t.Fatal(err)
			}
			entry, err := nativeabi.Wrap(arch, body, 16384)
			if err != nil {
				t.Fatal(err)
			}
			if arch != runtime.GOARCH || !native.Available {
				continue
			}
			code, err := native.Publish(entry.Code, []uint32{0}, uint64(len(entry.Code)))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(code.Close)
			lease, err := code.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			ctx, err := native.NewContext(16, 16, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			writeSlot(t, ctx, 0, 0, value.Value{Lo: want})
			state := native.State{Continuation: 1}
			if status, err := ctx.RunState(lease, 0, &state); err != nil || status != native.Return || state.ResultLo != want || state.Continuation != 1 {
				t.Fatal("fresh entry used a stale PC or wrong first mapping", allocated, status, state, err)
			}
		}
	}
}
