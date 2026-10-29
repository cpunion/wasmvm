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

func TestMemoryGuardNormalizedIntervalsAndFullMemory32End(t *testing.T) {
	// The probe only tests the guard, so a 4 GiB endpoint needs no 4 GiB
	// allocation. Compare emitted branches with the original interval contract.
	for _, arch := range []string{"amd64", "arm64"} {
		for _, cached := range []byte{0, 1, 19} {
			if arch == "amd64" && cached != 0 {
				continue
			}
			for _, interval := range [][2]uint64{{0, 1}, {0, 4095}, {0, 4096}, {33, 34}, {4095, 1}, {0, 4097}, {4096, 1}, {65536, 17}, {0, 1 << 32}, {(1 << 32) - 1, 1}} {
				g := memoryGuard{address: 0, before: interval[0], after: interval[1]}
				a := assembler{arch: arch, labels: map[int]int{}, longBranches: true}
				e := functionEmitter{a: &a, f: &slots.Function{Slots: []wasm.ValueType{wasm.I32}}, unmetered: true, exit: a.label()}
				cold, done := a.label(), a.label()
				e.contextLoad(0, nativeabi.ArenaOffset)
				frame, length := byte(6), byte(11)
				if arch == "arm64" {
					frame, length = 12, 17
				}
				e.registerMove(frame, e.operandRegister(0), true)
				e.contextLoad(0, nativeabi.ResultHiOffset)
				e.registerMove(length, e.operandRegister(0), true)
				if cached != 0 {
					e.contextLoad(0, nativeabi.ResultLoOffset)
					e.registerMove(cached, e.operandRegister(0), true)
					e.registers = map[slots.Slot]byte{0: cached}
				}
				e.emitMemoryGuard(g, cold)
				e.contextConst(nativeabi.PayloadOffset, 1)
				a.branch(done, false, 0)
				a.mark(cold)
				e.contextConst(nativeabi.PayloadOffset, 0)
				a.mark(done)
				if cached != 0 {
					e.registerMove(e.operandRegister(0), cached, true)
					e.contextStore(0, nativeabi.ResultLoOffset)
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
				ctx, err := native.NewContext(16, 16, 1)
				if err != nil {
					t.Fatal(err)
				}
				upper := (1 << 32) - g.after
				addresses := []uint64{0, 1, g.before - 1, g.before, g.before + 1, upper - 1, upper, upper + 1, 0xffffffff, 1 << 32, 1<<63 | g.before, ^uint64(0)}
				for _, address := range addresses {
					for _, bound := range []uint64{0, 1, g.before + g.after - 1, g.before + g.after, 65536, (1 << 32) - 1, 1 << 32, (1 << 32) + 1} {
						writeSlot(t, ctx, 0, 0, value.Value{Lo: address})
						if err := ctx.SetState(native.State{ResultLo: address, ResultHi: bound}); err != nil {
							t.Fatal(err)
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(exit, err)
						}
						state, err := ctx.State()
						low := uint64(uint32(address))
						want := low >= g.before && low <= upper && low+g.after <= bound
						if err != nil || (state.Payload[0] == 1) != want || state.ResultLo != address || readSlot(t, ctx, 0, 0) != (value.Value{Lo: address}) {
							t.Fatal("interval or original operand changed", arch, cached, g, address, bound, state, want, err)
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

func TestImmediateObservationAcrossProvenAccessesAndLaterGuards(t *testing.T) {
	for _, laterGuard := range []bool{false, true} {
		f := &slots.Function{LocalCount: 2, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32},
			Sources: []ir.Source{{Wasm: module.Instr{Op: 0x2d, Offset: 101}}, {Wasm: module.Instr{Op: 0x41, Bits: 2}}, {Wasm: module.Instr{Op: 0x74}},
				{Wasm: module.Instr{Op: 0x2d, B: 1, Offset: 102}}, {Wasm: module.Instr{Op: 0x41, Bits: 7}}, {Wasm: module.Instr{Op: 0x2d, B: 2, Offset: 103}}}}
		f.Code = []slots.Instruction{
			{Op: 0x2d, Dst: 4, A: 0, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x41, Source: 1, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x74, Source: 2, Dst: 3, A: 3, B: 2, C: slots.NoSlot},
			{Op: 0x2d, Source: 3, Dst: 4, A: 0, B: slots.NoSlot, C: slots.NoSlot},
		}
		if laterGuard {
			for _, source := range []uint32{0, 3, 5} {
				f.Code = append(f.Code, slots.Instruction{Op: 0x2d, Source: source, Dst: 5, A: 1, B: slots.NoSlot, C: slots.NoSlot})
			}
		}
		f.Code = append(f.Code,
			slots.Instruction{Op: 0x41, Source: 4, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			slots.Instruction{Op: 0x2d, Source: 5, Dst: 4, A: 0, B: slots.NoSlot, C: slots.NoSlot},
			slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			if err != nil || (c.ForwardedScalars > 0) == laterGuard {
				t.Fatal("unsafe or missing dead immediate", arch, laterGuard, c, err)
			}
		}
		compareGuardedExecution(t, f, []uint64{0, 1, 65518, 65532, 65534, 0xffffffff, 1 << 32})
	}
}
