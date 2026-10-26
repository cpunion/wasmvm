package codegen

import (
	"bytes"
	"encoding/binary"
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

// Exercise both short-range boundaries, long fallbacks, and forward/backward
// targets. Decode the final branch independently of relocation metadata.
func TestARMConditionalRelaxation(t *testing.T) {
	for _, target := range []int{-(1 << 20) - 4, -(1 << 20), -4, 0, 8, (1 << 20) - 4, 1 << 20} {
		for _, known := range []bool{false, true} {
			for _, wide := range []bool{false, true} {
				for _, nonzero := range []bool{false, true} {
					for _, flags := range []bool{false, true} {
						a := assembler{arch: "arm64", labels: map[int]int{}, longBranches: true}
						if known {
							a.labels[1] = target
						}
						if flags {
							a.branch(1, true, 10)
						} else {
							a.compareBranch(1, 28, wide, nonzero)
						}
						a.labels[1] = target
						code, err := a.finish()
						if err != nil {
							t.Fatal(err)
						}
						first := binary.LittleEndian.Uint32(code)
						near := target >= -(1<<20) && target < 1<<20
						destination := int(int32(first<<8)>>13) * 4
						if near {
							if destination != target {
								t.Fatal(target, first, destination)
							}
							if len(code) == 8 && binary.LittleEndian.Uint32(code[4:]) != 0xd503201f {
								t.Fatal("unstable relaxed padding")
							}
						} else {
							second := binary.LittleEndian.Uint32(code[4:])
							if destination != 8 || second>>26 != 5 || 4+int(int32(second<<6)>>6)*4 != target {
								t.Fatal("bad long fallback", target, code)
							}
						}
						if flags {
							expected := uint32(10)
							if !near {
								expected ^= 1
							}
							if first&15 != expected {
								t.Fatal("condition inverted", first)
							}
						} else {
							expected := uint32(0x34000000) | 28
							if wide {
								expected |= 1 << 31
							}
							if nonzero != !near {
								expected |= 1 << 24
							}
							if first&^(0x7ffff<<5) != expected {
								t.Fatal("register branch changed width/condition", first, expected)
							}
						}
						before := append([]byte(nil), code...)
						if _, err = a.finish(); err != nil || !bytes.Equal(before, a.code) {
							t.Fatal("non-idempotent finish", err)
						}
					}
				}
			}
		}
	}
}

func TestConditionalEdgesPreserveCopiesAndContinuations(t *testing.T) {
	if !native.Available {
		return
	}
	for _, fast := range []bool{false, true} {
		for yesPC := uint32(1); yesPC <= 2; yesPC++ {
			for noPC := uint32(1); noPC <= 2; noPC++ {
				for copies := 0; copies < 4; copies++ {
					f := &slots.Function{
						Slots:   []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32},
						Sources: []ir.Source{{Wasm: module.Instr{Op: 1}}},
						Args:    []slots.Slot{1, 2}, Moves: []slots.Move{{Dst: 1, Src: 3}, {Dst: 2, Src: 3}},
						Edges: []slots.Edge{{PC: yesPC, Count: uint32(copies & 1)}, {PC: noPC, Start: 1, Count: uint32(copies >> 1)}},
						Code: []slots.Instruction{
							{Op: slots.ControlBase + uint32(ir.If), A: 0, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Count: 2},
							{Op: slots.ControlBase + uint32(ir.Return), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Count: 1},
							{Op: slots.ControlBase + uint32(ir.Return), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Extra: 1, Count: 1},
						},
					}
					compiled, lease, c := functionContext(t, f, fast)
					for _, cond := range []uint64{0, 1, 0xffffffff, 1 << 32} {
						got, exit, _ := runFunction(t, compiled, lease, c, []value.Value{{Lo: cond}, {Lo: 11}, {Lo: 22}, {Lo: 77}}, 100, false, 1)
						wantPC := noPC
						a, b := uint64(11), uint64(22)
						if uint32(cond) != 0 {
							wantPC = yesPC
							if copies&1 != 0 {
								a = 77
							}
						} else if copies&2 != 0 {
							b = 77
						}
						want := a
						if wantPC == 2 {
							want = b
						}
						if exit != native.Return || len(got) != 1 || got[0].Lo != want || readSlot(t, c, 16, 1).Lo != a || readSlot(t, c, 16, 2).Lo != b {
							t.Fatal(fast, yesPC, noPC, copies, cond, got, exit)
						}
						// Every original return remains a valid independent entry. Neither arm's
						// copies nor the condition may be inferred when entering here directly.
						for pc := uint64(1); pc <= 2; pc++ {
							writeSlot(t, c, 16, 1, value.Value{Lo: 11})
							writeSlot(t, c, 16, 2, value.Value{Lo: 22})
							if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: pc, Budget: 1}); err != nil {
								t.Fatal(err)
							}
							if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
								t.Fatal(exit, err)
							}
							if readSlot(t, c, 16, 1).Lo != 11 || readSlot(t, c, 16, 2).Lo != 22 {
								t.Fatal("entry inherited edge copies")
							}
						}
					}
					c.Close()
					lease.Close()
				}
			}
		}
	}
}
