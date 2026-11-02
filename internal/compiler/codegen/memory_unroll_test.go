package codegen

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func unrollFixture(padding int, reversed, constant bool, limit uint32) *slots.Function {
	f := &slots.Function{LocalCount: 3, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32}}
	add := func(op uint32, dst, a, b slots.Slot, bits uint64) {
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: uint32(len(f.Sources)), Dst: dst, A: a, B: b, C: slots.NoSlot})
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: bits, Offset: uint64(len(f.Sources)) + 100}})
	}
	add(0x20, 3, slots.NoSlot, slots.NoSlot, 0)
	add(0x41, 4, slots.NoSlot, slots.NoSlot, 255)
	add(0x71, 3, 3, 4, 0)
	add(0x28, 4, 3, slots.NoSlot, 0)
	add(0x41, 5, slots.NoSlot, slots.NoSlot, 1)
	add(0x6a, 4, 4, 5, 0)
	add(0x36, slots.NoSlot, 3, 4, 0)
	add(0x6a, 2, 2, 4, 0)
	add(0x6a, 0, 0, 5, 0)
	for range padding {
		add(1, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	}
	bound := slots.Slot(1)
	if constant {
		add(0x41, 5, slots.NoSlot, slots.NoSlot, uint64(limit))
		bound = 5
	}
	op := uint32(0x49)
	if reversed {
		op = 0x4f
	}
	add(op, 6, 0, bound, 0)
	add(slots.ControlBase+uint32(ir.If), slots.NoSlot, 6, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = 2
	f.Edges = []slots.Edge{{PC: 0}, {PC: uint32(len(f.Code))}}
	if reversed {
		f.Edges[0], f.Edges[1] = f.Edges[1], f.Edges[0]
	}
	add(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Slots))
	for i := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
	}
	return f
}

func TestMemoryUnrollSelectionAndCodeBudget(t *testing.T) {
	for _, padding := range []int{0, 53, 54, 116, 117} {
		f := unrollFixture(padding, false, false, 0)
		for _, arch := range []string{"amd64", "arm64"} {
			config := Config{Unmetered: true, MemoryMinimumPages: 1}
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			copies := uint32(4)
			if len(f.Code)-1 > 64 {
				copies = 2
			}
			if err != nil || c.UnrolledLoops != 1 || c.UnrolledCopies != copies || copies*uint32(len(f.Code)-1) > 256 {
				t.Fatal(arch, padding, c, err)
			}
			limits := DefaultLimits()
			limits.MaxCodeBytes = uint64(len(c.Entry.Code))
			if _, err := CompileFunctionWithConfig(arch, f, limits, config); err != nil {
				t.Fatal(err)
			}
			limits.MaxCodeBytes--
			if c, err := CompileFunctionWithConfig(arch, f, limits, config); c != nil || err == nil {
				t.Fatal("unrolled code exceeded budget", c, err)
			}
		}
	}
}

func TestMemoryUnrollColdEntriesRemaindersAndWrap(t *testing.T) {
	if !native.Available {
		return
	}
	for _, reversed := range []bool{false, true} {
		for _, constant := range []bool{false, true} {
			padding := 0
			if reversed {
				padding = 54
			}
			f := unrollFixture(padding, reversed, constant, 4)
			c, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
			_, refLease, ref := functionContext(t, f, true)
			if c.UnrolledLoops != 1 {
				t.Fatal("unroll not exercised")
			}
			memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
			initial, gotBytes, wantBytes := make([]byte, 65536), make([]byte, 65536), make([]byte, 65536)
			for i := range initial {
				initial[i] = byte(i*193 + 127)
			}
			for pc := range f.Code {
				for _, input := range [][3]uint64{{0, 0, 0}, {0, 1, 100}, {0, 2, 255}, {0, 3, 0}, {0, 4, 0}, {0, 5, 0}, {0, 8, 0}, {3, 8, 9}, {7, 3, 3}, {0xffffffff, 4, 200}, {1<<32 | 2, 1<<32 | 5, 0xffffffff}} {
					memory.Write(0, initial)
					refMemory.Write(0, initial)
					for s := range f.Slots {
						bits := uint64(s*17 + 5)
						if s < 2 {
							bits = input[s]
						} else if s == 3 {
							bits = input[2]
						}
						writeSlot(t, ctx, 16, slots.Slot(s), value.Value{Lo: bits})
						writeSlot(t, ref, 16, slots.Slot(s), value.Value{Lo: bits})
					}
					state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
					ctx.SetState(state)
					ref.SetState(state)
					got, err := ctx.Enter(lease, 0)
					want, refErr := ref.Enter(refLease, 0)
					gs, _ := ctx.State()
					ws, _ := ref.State()
					if err != nil || refErr != nil || got != want || gs != ws {
						t.Fatal("unroll state/trap", reversed, constant, pc, input, got, want, gs, ws, err, refErr)
					}
					for s := range f.Slots {
						if got, want := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); got != want {
							t.Fatal("unroll frame", reversed, constant, pc, input, s, got, want)
						}
					}
					memory.Read(0, gotBytes)
					refMemory.Read(0, wantBytes)
					if !bytes.Equal(gotBytes, wantBytes) {
						t.Fatal("unroll memory effects", reversed, constant, pc, input)
					}
				}
			}
		}
	}
}

func TestMemoryUnrollRejectsUnprovedGroups(t *testing.T) {
	for _, kind := range []string{"metered", "minimum", "shared", "long", "step", "limit-write", "predicate", "edge-move", "unproven-load", "mask", "trap", "invariant-address", "dependent-address", "subword-load", "subword-store"} {
		f := unrollFixture(0, false, false, 0)
		config := Config{Unmetered: true, MemoryMinimumPages: 1}
		switch kind {
		case "metered":
			config.Unmetered = false
		case "minimum":
			config.MemoryMinimumPages = 0
		case "shared":
			config.NativeThreads, config.MemoryCount, config.SharedMemories = true, 1, []bool{true}
		case "long":
			f = unrollFixture(118, false, false, 0)
		case "step":
			f.Sources[4].Wasm.Bits = 2
		case "limit-write":
			f.Code[7].Dst = 1
		case "predicate":
			f.Code[len(f.Code)-3].Op = 0x47
		case "edge-move":
			f.Edges[0].Count = 1
			f.Moves = []slots.Move{{Dst: 0, Src: 0}}
		case "unproven-load":
			f.Sources[3].Wasm.B = 65536
		case "mask":
			f.Sources[1].Wasm.Bits = 0xffffffff
		case "trap":
			f.Code[7].Op = 0x6d
		case "invariant-address":
			f.Code[0].Op, f.Sources[0].Wasm.Op = 0x41, 0x41
		case "dependent-address":
			f.Code[6].A = 4
		case "subword-load":
			f.Code[3].Op, f.Sources[3].Wasm.Op = 0x2d, 0x2d
		case "subword-store":
			f.Code[6].Op, f.Sources[6].Wasm.Op = 0x3a, 0x3a
		}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			if err != nil || c.UnrolledLoops != 0 {
				t.Fatal(kind, arch, c, err)
			}
		}
	}
}

func TestMemoryUnrollCancellation(t *testing.T) {
	if !native.Available {
		return
	}
	f := unrollFixture(54, false, false, 0)
	c, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	if c.UnrolledLoops != 1 {
		t.Fatal("cancellation fixture not unrolled")
	}
	bindMemory(t, ctx, 65536, 65536)
	writeSlot(t, ctx, 16, 1, value.Value{Lo: 0xffffffff})
	ctx.SetState(native.State{Used: 16, Depth: 1})
	done := make(chan native.Exit, 1)
	go func() {
		exit, err := ctx.Enter(lease, 0)
		if err != nil {
			t.Error(err)
		}
		done <- exit
	}()
	defer ctx.Cancel()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := ctx.State(); err == native.ErrBusy {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("unrolled loop did not start")
		}
		runtime.Gosched()
	}
	ctx.Cancel()
	select {
	case exit := <-done:
		state, err := ctx.State()
		if err != nil || exit != native.Poll || state.Continuation >= uint64(len(f.Code)) {
			t.Fatal("unrolled loop lost original cancellation PC", exit, state, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("unrolled loop missed cancellation")
	}
}
