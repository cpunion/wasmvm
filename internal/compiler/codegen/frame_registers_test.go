package codegen

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPhysicalFrameTransfersPreserveRawWordsAndLargeOffsets(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, typ := range []wasm.ValueType{wasm.I64, wasm.F32, wasm.F64} {
			for _, index := range []uint32{0, 1, 1023, 1024, 2047, 2048, 4095, 4096} {
				for _, extended := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%x/%d/%v", arch, typ, index, extended), func(t *testing.T) {
						f := &slots.Function{Slots: make([]wasm.ValueType, index+2)}
						for i := range f.Slots {
							f.Slots[i] = typ
						}
						floating := typ == wasm.F32 || typ == wasm.F64
						reg := byte(3)
						if extended {
							reg = 15
						}
						if arch == "arm64" {
							if floating {
								reg = 8
								if extended {
									reg = 31
								}
							} else {
								reg = 19
								if extended {
									reg = 28
								}
							}
						}
						a := &assembler{arch: arch, longBranches: true}
						e := &functionEmitter{a: a, f: f, unmetered: true, exit: a.label()}
						e.contextLoad(0, nativeabi.ArenaOffset)
						frame := byte(6)
						if arch == "arm64" {
							frame = 12
						}
						e.registerMove(frame, e.operandRegister(0), true)
						src, dst := slots.Slot(index), slots.Slot(index+1)
						e.frameRegister(src, reg, floating, false)
						e.frameRegister(dst, reg, floating, true)
						e.rawStatus(nativeabi.Return)
						a.mark(e.exit)
						body, err := a.finish()
						if err != nil {
							t.Fatal(err)
						}
						wrapped, err := nativeabi.Wrap(arch, body, 1<<16)
						if err != nil {
							t.Fatal(err)
						}
						if arch != runtime.GOARCH || !native.Available {
							return
						}
						code, err := native.Publish(wrapped.Code, []uint32{0}, 1<<16)
						if err != nil {
							t.Fatal(err)
						}
						defer code.Close()
						lease, err := code.Acquire()
						if err != nil {
							t.Fatal(err)
						}
						defer lease.Close()
						size := uint64(len(f.Slots)) * 16
						ctx, err := native.NewContext(size, size, 1)
						if err != nil {
							t.Fatal(err)
						}
						defer ctx.Close()
						original := value.Value{Lo: 0xfedcba987fc01234, Hi: 0x1122334455667788}
						if typ == wasm.F64 {
							original.Lo = 0xfff8000000001234
						}
						untouched := value.Value{Lo: 0xdeadbeef01234567, Hi: 0x8877665544332211}
						writeSlot(t, ctx, 0, src, original)
						writeSlot(t, ctx, 0, dst, untouched)
						state := native.State{Depth: 1}
						if exit, err := ctx.RunState(lease, 0, &state); err != nil || exit != native.Return {
							t.Fatal(exit, err)
						}
						want := untouched
						want.Lo = original.Lo
						if typ == wasm.F32 {
							want.Lo = uint64(uint32(original.Lo))
						}
						if got := readSlot(t, ctx, 0, dst); got != want {
							t.Fatal("lost payload or changed high word", got, want)
						}
						if got := readSlot(t, ctx, 0, src); got != original {
							t.Fatal("source changed", got, original)
						}
					})
				}
			}
		}
	}
}
