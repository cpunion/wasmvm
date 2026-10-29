package codegen

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMemoryLoopsScalarVectorWidthsAndLanes(t *testing.T) {
	var ops []uint32
	for op := uint32(0x28); op <= 0x3e; op++ {
		ops = append(ops, op)
	}
	for op := uint32(0xfd0000); op <= 0xfd005d; op++ {
		if nativeVectorMemoryOpcode(op) {
			ops = append(ops, op)
		}
	}
	for _, op := range ops {
		load, size := memoryAccess(op)
		lanes := uint32(1)
		if op >= 0xfd0054 && op <= 0xfd005b {
			lanes = 16 / size
		}
		for lane := uint32(0); lane < lanes; lane++ {
			t.Run(fmt.Sprintf("%x/%d", op, lane), func(t *testing.T) {
				var original *slots.Function
				if nativeVectorMemoryOpcode(op) {
					original, _ = vectorMemoryFunction(t, op, 0, byte(lane), false, false)
				} else {
					original, _ = memoryFunction(t, op, 0)
				}
				sig := original.Signature
				cursor := uint32(len(sig.Params))
				sig.Params = append(append([]wasm.ValueType(nil), sig.Params...), wasm.I32)
				sig.Results = []wasm.ValueType{wasm.I32}
				root := uint32(len(sig.Params))
				body := module.Expr{{Op: 3, Block: -64}, {Op: 0x20}, {Op: 0x20, A: cursor}, {Op: 0x6a}, {Op: 0x21, A: root}}
				for _, bias := range []uint64{33, 1, 2} {
					body = append(body, module.Instr{Op: 0x20, A: root}, module.Instr{Op: 0x41, Bits: bias}, module.Instr{Op: 0x6a})
					if !load || op >= 0xfd0054 && op <= 0xfd0057 {
						body = append(body, module.Instr{Op: 0x20, A: 1})
					}
					body = append(body, module.Instr{Op: op, Lane: byte(lane)})
					if load {
						body = append(body, module.Instr{Op: 0x1a})
					}
				}
				body = append(body, module.Instr{Op: 0x20, A: cursor}, module.Instr{Op: 0x41, Bits: 1}, module.Instr{Op: 0x6a}, module.Instr{Op: 0x22, A: cursor},
					module.Instr{Op: 0x41, Bits: 4}, module.Instr{Op: 0x47}, module.Instr{Op: 0xd}, module.Instr{Op: 0xb}, module.Instr{Op: 0x20, A: cursor}, module.Instr{Op: 0xb})
				m := functionFixture(sig, 1, body)
				m.Memories = []module.Limits{{Min: 1}}
				f := functionModule(t, m).Functions[0]
				for _, arch := range []string{"amd64", "arm64"} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
					if err != nil {
						t.Fatal(err)
					}
					if c.LoopGuards != 1 {
						t.Fatal("typed loop not versioned", arch, c.LoopGuards)
					}
				}
				if !native.Available {
					return
				}
				fast, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
				checked, refLease, ref := functionContext(t, f, true)
				mem, refMem := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
				initial := make([]byte, 65536)
				for i := range initial {
					initial[i] = byte(i*193 + 127)
				}
				for _, base := range []uint64{0, 65536 - 33 - uint64(size) - 2, 0xffffffff} {
					mem.Write(0, initial)
					refMem.Write(0, initial)
					args := make([]value.Value, len(sig.Params))
					args[0].Lo = base
					if cursor > 1 {
						args[1] = value.Value{Lo: 0x123456789abcdef0}
						if sig.Params[1] == wasm.V128 {
							args[1].Hi = 0xfedcba9876543210
						}
					}
					got, exit, state := runFunction(t, fast, lease, ctx, args, 0, false, 0)
					want, refExit, refState := runFunction(t, checked, refLease, ref, args, 0, false, 0)
					if exit != refExit || state != refState || len(got) != len(want) || len(got) > 0 && got[0] != want[0] {
						t.Fatal("typed loop result/trap", base, got, want, exit, refExit, state, refState)
					}
					for s := range f.Slots {
						if readSlot(t, ctx, 16, slots.Slot(s)) != readSlot(t, ref, 16, slots.Slot(s)) {
							t.Fatal("typed loop frame", base, s)
						}
					}
					gb, wb := make([]byte, 65536), make([]byte, 65536)
					mem.Read(0, gb)
					refMem.Read(0, wb)
					if !bytes.Equal(gb, wb) {
						t.Fatal("typed loop memory effects", base)
					}
				}
			})
		}
	}
}
