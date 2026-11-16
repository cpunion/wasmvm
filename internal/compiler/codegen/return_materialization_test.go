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

func TestPrivateReturnMaterializesResultsWorkHomesAndCanceledRoots(t *testing.T) {
	const poison = uint64(0xdead)
	want := []value.Value{{Lo: 0x87654321fedcba98, Hi: 31}, {Lo: 0x123456789abcdef0, Hi: 47}, {Lo: 0x3f000000, Hi: 59}, {Lo: 0x1020304050607080, Hi: 0xfedcba9876543210}}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, cancel := range []bool{false, true} {
			a := assembler{arch: arch, longBranches: true}
			f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.ExternRef, wasm.F32, wasm.V128}, Args: []slots.Slot{3}}
			e := functionEmitter{a: &a, f: f, unmetered: true, frameRegisters: true, exit: a.label()}
			var work [32]bool
			for n := byte(0); n < 8; n++ {
				work[e.callReg(n)] = true
			}
			var safe byte
			for _, reg := range e.integerRegisterPool() {
				if !work[reg] {
					safe = reg
					break
				}
			}
			plan := &registerPlan{homes: map[slots.Slot]byte{0: e.callReg(3), 1: safe}, floats: map[slots.Slot]byte{2: 3}, vectors: map[slots.Slot]byte{3: 4}, slots: []slots.Slot{0, 1, 2, 3}}
			e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
			e.callEntryFrame(0)
			e.callConst(callUsed, 0)
			e.callConst(callDepth, 1)
			e.callConst(callTop, 0)
			e.transitionRegisters(plan)
			// Keep the current complete values in registers while making their
			// old frame homes observably stale. Scalar high words remain owned
			// by their canonical frames; vectors have two register-owned words.
			e.callConst(0, poison)
			for slot := range want {
				e.callStore(0, e.callFrame(), uint32(slot)*16)
			}
			e.callStore(0, e.callFrame(), 3*16+8)
			deferred := e.callReturnSpills(slots.Instruction{Count: 1}, false)
			for n := byte(0); n < 8; n++ {
				e.callConst(n, poison+uint64(n))
			}
			if cancel {
				e.contextConst(nativeabi.ControlOffset, 1)
			}
			canceled := a.label()
			e.callCancelled(canceled)
			e.rawStatus(nativeabi.Return)
			a.mark(canceled)
			e.spillRegisterSlots(deferred)
			e.rawStatus(nativeabi.Poll)
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
			ctx, err := native.NewContext(64, 64, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			for slot, v := range want {
				writeSlot(t, ctx, 0, slots.Slot(slot), v)
			}
			status, err := ctx.RunState(lease, 0, &native.State{})
			expected := native.Return
			if cancel {
				expected = native.Poll
			}
			if err != nil || status != expected {
				t.Fatal(cancel, status, err)
			}
			for slot, v := range want {
				if !cancel && (slot == 1 || slot == 2) {
					v.Lo = poison
				}
				if got := readSlot(t, ctx, 0, slots.Slot(slot)); got != v {
					t.Fatal("return materialization lost a word or stored a retired home", cancel, slot, got, v)
				}
			}
		}
	}
}
