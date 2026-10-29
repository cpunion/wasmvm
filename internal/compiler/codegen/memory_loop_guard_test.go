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

func TestRemainingLoopSpanArithmeticAndOriginalOperands(t *testing.T) {
	// Only guard arithmetic runs: memory32 endpoints need no multi-GiB mapping.
	for _, arch := range []string{"amd64", "arm64"} {
		for mask := 0; mask < 4; mask++ {
			for _, interval := range [][2]uint64{{0, 1}, {33, 34}, {4096, 17}, {0, 1 << 32}, {(1 << 32) - 1, 1}} {
				for _, limit := range []uint32{2, 30, 65536, 0xffffffff} {
					a := assembler{arch: arch, labels: map[int]int{}, longBranches: true}
					e := functionEmitter{a: &a, f: &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}}, unmetered: true, exit: a.label()}
					cold, done := a.label(), a.label()
					e.contextLoad(0, nativeabi.ArenaOffset)
					frame, length := byte(6), byte(11)
					if arch == "arm64" {
						frame, length = 12, 17
					}
					e.registerMove(frame, e.operandRegister(0), true)
					e.contextLoad(0, nativeabi.ResultHiOffset)
					e.registerMove(length, e.operandRegister(0), true)
					registers := map[slots.Slot]byte{}
					pool := e.integerRegisterPool()
					for s := slots.Slot(0); s < 2; s++ {
						if mask&(1<<s) != 0 {
							e.slotLoad(s, 0, false, true)
							e.registerMove(pool[s], e.operandRegister(0), true)
							registers[s] = pool[s]
						}
					}
					e.registers = registers
					e.emitLoopGuard(memoryLoop{cursor: 1, limit: limit}, memoryGuard{address: 0, before: interval[0], after: interval[1]}, cold)
					e.contextConst(nativeabi.PayloadOffset, 1)
					a.branch(done, false, 0)
					a.mark(cold)
					e.contextConst(nativeabi.PayloadOffset, 0)
					a.mark(done)
					// Publish cached raw words to prove that guard failure and
					// success both preserve their complete mutable input bits.
					for s, offset := range []uintptr{nativeabi.ResultLoOffset, nativeabi.PayloadOffset + 8} {
						e.slotLoad(slots.Slot(s), 0, false, true)
						e.contextStore(0, offset)
					}
					e.rawStatus(nativeabi.Return)
					a.mark(e.exit)
					body, err := a.finish()
					if err != nil {
						t.Fatal(err)
					}
					entry, err := nativeabi.Wrap(arch, body, 4096)
					if err != nil {
						t.Fatal(err)
					}
					if !native.Available || arch != runtime.GOARCH {
						continue
					}
					code, err := native.Publish(entry.Code, []uint32{0}, 4096)
					if err != nil {
						t.Fatal(err)
					}
					lease, err := code.Acquire()
					if err != nil {
						t.Fatal(err)
					}
					ctx, err := native.NewContext(32, 32, 1)
					if err != nil {
						t.Fatal(err)
					}
					for _, cursor := range []uint64{0, 1, uint64(limit) - 1, uint64(limit), uint64(limit) + 1, 0xffffffff, 1<<63 | 1} {
						for _, address := range []uint64{0, interval[0] - 1, interval[0], interval[0] + 1, 0xffffffde, 0xffffffff, 1 << 32, ^uint64(0)} {
							low, i := uint64(uint32(address)), uint64(uint32(cursor))
							end := uint64(uint32(low-interval[0])) + uint64(limit) - 1 - i + interval[0] + interval[1]
							for _, bound := range []uint64{0, 65536, (1 << 32) - 1, 1 << 32, (1 << 32) + 1} {
								writeSlot(t, ctx, 0, 0, value.Value{Lo: address})
								writeSlot(t, ctx, 0, 1, value.Value{Lo: cursor})
								ctx.SetState(native.State{ResultHi: bound})
								if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
									t.Fatal(exit, err)
								}
								state, err := ctx.State()
								want := i < uint64(limit) && low >= interval[0] && end <= 1<<32 && end <= bound
								if err != nil || (state.Payload[0] == 1) != want || state.ResultLo != address || state.Payload[1] != cursor {
									t.Fatal(arch, mask, interval, limit, cursor, address, bound, state, want, err)
								}
								for s, bits := range []uint64{address, cursor} {
									if readSlot(t, ctx, 0, slots.Slot(s)) != (value.Value{Lo: bits}) {
										t.Fatal("guard wrote a mutable operand", s)
									}
								}
							}
						}
					}
					ctx.Close()
					lease.Close()
					code.Close()
				}
			}
		}
	}
}
