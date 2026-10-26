package codegen

import (
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestAllocatedConditionUsesLow32Bits(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		registers := []byte{3, 9, 10, 11, 12, 13, 14, 15}
		if arch == "arm64" {
			registers = []byte{19, 20, 21, 22, 23, 24, 25, 26, 27, 28}
		}
		for _, reg := range registers {
			a := assembler{arch: arch, labels: map[int]int{}, longBranches: true}
			e := functionEmitter{a: &a, exit: a.label(), registers: map[slots.Slot]byte{0: reg}}
			zero, done := a.label(), a.label()
			e.contextLoad(0, nativeabi.ResultLoOffset)
			e.registerMove(reg, e.operandRegister(0), true)
			e.branchZero(0, zero)
			e.constant(0, 7)
			a.branch(done, false, 0)
			a.mark(zero)
			e.constant(0, 11)
			a.mark(done)
			e.contextStore(0, nativeabi.ResultLoOffset)
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
			c, err := native.NewContext(0, 0, 1)
			if err != nil {
				t.Fatal(err)
			}
			for _, condition := range []uint64{0, 1, 2, 0x80000000, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0)} {
				if err := c.SetState(native.State{ResultLo: condition}); err != nil {
					t.Fatal(err)
				}
				if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
					t.Fatal(exit, err)
				}
				state, err := c.State()
				want := uint64(7)
				if uint32(condition) == 0 {
					want = 11
				}
				if err != nil || state.ResultLo != want {
					t.Fatal(arch, reg, condition, state, err)
				}
			}
			c.Close()
			lease.Close()
			code.Close()
		}
	}
}
