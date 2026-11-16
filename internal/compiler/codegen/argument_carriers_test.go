package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPrivateArgumentCarriersPreserveWordsAndParallelMoves(t *testing.T) {
	want := [2]value.Value{{Lo: 0xfedcba9876543210, Hi: 0x1020304050607080}, {Lo: 0x89abcdef01234567, Hi: 0x8877665544332211}}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, swap := range []bool{false, true} {
			a := assembler{arch: arch, longBranches: true}
			e := functionEmitter{a: &a, f: &slots.Function{Signature: module.FuncType{Params: []wasm.ValueType{wasm.I64, wasm.I64}}, Slots: []wasm.ValueType{wasm.I64, wasm.I64}}, unmetered: true, frameRegisters: true, exit: a.label()}
			pool := e.integerRegisterPool()
			regs := [2]byte{pool[0], pool[1]}
			if arch == "amd64" {
				regs = [2]byte{e.callReg(5), e.callReg(6)}
			}
			if swap {
				regs[0], regs[1] = regs[1], regs[0]
			}
			plan := &registerPlan{homes: map[slots.Slot]byte{0: regs[0], 1: regs[1]}, slots: []slots.Slot{0, 1}}
			e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
			e.callEntryFrame(0)
			e.callConst(callUsed, 0)
			e.callConst(callDepth, 1)
			e.callConst(callTop, 0)
			e.callPair(5, 1, e.callFrame(), 0, false)
			e.callPair(6, 1, e.callFrame(), 16, false)
			// Isolate the private entry from its canonical-frame fallback. Real
			// callers retain both words; here only the stale low words are poisoned
			// after capturing them, so a hidden reload cannot pass this probe.
			for i := range want {
				e.callConst(0, uint64(i)+0xdead)
				e.callStore(0, e.callFrame(), uint32(i)*16)
			}
			pcs := []int{a.label()}
			e.emitCallStart(pcs, plan)
			a.mark(pcs[0])
			e.registerMove(e.callReg(0), regs[0], true)
			e.callStore(0, e.callContext(), uint32(nativeabi.ResultLoOffset))
			e.registerMove(e.callReg(0), regs[1], true)
			e.callStore(0, e.callContext(), uint32(nativeabi.ResultHiOffset))
			e.rawStatus(nativeabi.Return)
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
			ctx, err := native.NewContext(32, 32, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			for i, v := range want {
				writeSlot(t, ctx, 0, slots.Slot(i), v)
			}
			s := native.State{Continuation: 7}
			if status, err := ctx.RunState(lease, 0, &s); err != nil || status != native.Return || s.ResultLo != want[0].Lo || s.ResultHi != want[1].Lo || s.Continuation != 7 {
				t.Fatal(arch, swap, status, s, err)
			}
			for i, v := range want {
				if got := readSlot(t, ctx, 0, slots.Slot(i)); got.Hi != v.Hi {
					t.Fatal("carrier modified the opaque high word", arch, swap, i, got, v)
				}
			}
		}
	}
}
