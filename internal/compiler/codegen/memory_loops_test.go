package codegen

import (
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

func memoryLoopFixture(limit uint32, firstPC, reversed bool) *slots.Function {
	f := &slots.Function{LocalCount: 3, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32},
		Sources: []ir.Source{
			{Wasm: module.Instr{Op: 0x6a}}, {Wasm: module.Instr{Op: 0x22, A: 2}},
			{Wasm: module.Instr{Op: 0x41, Bits: 33}}, {Wasm: module.Instr{Op: 0x6a}},
			{Wasm: module.Instr{Op: 0x3a, Offset: 100}}, {Wasm: module.Instr{Op: 0x3a, B: 1, Offset: 101}}, {Wasm: module.Instr{Op: 0x3a, B: 2, Offset: 102}},
			{Wasm: module.Instr{Op: 0x41, Bits: 1}}, {Wasm: module.Instr{Op: 0x6a}}, {Wasm: module.Instr{Op: 0x22, A: 1}},
			{Wasm: module.Instr{Op: 0x41, Bits: uint64(limit)}}, {Wasm: module.Instr{Op: 0x47}}, {}, {},
		},
		Code: []slots.Instruction{
			{Op: 0x6a, Dst: 3, A: 0, B: 1, C: slots.NoSlot},
			{Op: 0x22, Source: 1, Dst: 2, A: 3, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x41, Source: 2, Dst: 4, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x6a, Source: 3, Dst: 4, A: 2, B: 4, C: slots.NoSlot},
			{Op: 0x3a, Source: 4, Dst: slots.NoSlot, A: 4, B: 1, C: slots.NoSlot},
			{Op: 0x3a, Source: 5, Dst: slots.NoSlot, A: 2, B: 1, C: slots.NoSlot},
			{Op: 0x3a, Source: 6, Dst: slots.NoSlot, A: 2, B: 1, C: slots.NoSlot},
			{Op: 0x41, Source: 7, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x6a, Source: 8, Dst: 3, A: 1, B: 3, C: slots.NoSlot},
			{Op: 0x22, Source: 9, Dst: 1, A: 3, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x41, Source: 10, Dst: 3, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x47, Source: 11, Dst: 3, A: 1, B: 3, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.If), Source: 12, Dst: slots.NoSlot, A: 3, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
			{Op: slots.ControlBase + uint32(ir.Return), Source: 13, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 7},
		}}
	if firstPC {
		f.Code = append([]slots.Instruction(nil), f.Code[4:]...)
		for i := 0; i < 3; i++ {
			f.Code[i].A = 1
		}
		cmp := &f.Code[len(f.Code)-3]
		cmp.A, cmp.B = cmp.B, cmp.A
	}
	f.Edges = []slots.Edge{{PC: 0}, {PC: uint32(len(f.Code) - 1)}}
	if reversed {
		f.Code[len(f.Code)-3].Op = 0x46
		f.Edges[0], f.Edges[1] = f.Edges[1], f.Edges[0]
		// The false backedge also covers an exit beyond the fallthrough PC.
		f.Edges[0].PC = uint32(len(f.Code))
		f.Code = append(f.Code, f.Code[len(f.Code)-1])
	}
	for i := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
	}
	return f
}

func TestMemoryLoopGuardEffectsColdPCsAndRebinding(t *testing.T) {
	for _, firstPC := range []bool{false, true} {
		for _, reversed := range []bool{false, true} {
			f := memoryLoopFixture(4, firstPC, reversed)
			for _, arch := range []string{"amd64", "arm64"} {
				config := Config{Unmetered: true, MemoryMinimumPages: 1}
				c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
				if err != nil || c.LoopGuards != 1 || c.MemoryGuards != 1 {
					t.Fatal("loop not selected", arch, firstPC, reversed, c, err)
				}
				limits := DefaultLimits()
				limits.MaxCodeBytes = uint64(len(c.Entry.Code))
				if _, err := CompileFunctionWithConfig(arch, f, limits, config); err != nil {
					t.Fatal(err)
				}
				limits.MaxCodeBytes--
				if c, err := CompileFunctionWithConfig(arch, f, limits, config); c != nil || err == nil {
					t.Fatal("code limit ignored", c, err)
				}
			}
			if !native.Available {
				continue
			}
			_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
			_, reference, ref := functionContext(t, f, true)
			for _, size := range []uint64{65536, 131072} {
				memory := bindMemory(t, ctx, size, size)
				refMemory := bindMemory(t, ref, size, size)
				initial := make([]byte, size)
				for i := range initial {
					initial[i] = byte(i*193 + 127)
				}
				for pc := range f.Code {
					for _, input := range [][2]uint64{{0, 0}, {0, 2}, {0, 3}, {0, 4}, {size - 36, 0}, {size - 35, 0}, {size - 34, 0}, {0, 0xffffffff}, {0xffffffff, 1}, {1 << 32, 1<<32 | 2}} {
						memory.Write(0, initial)
						refMemory.Write(0, initial)
						for s := range f.Slots {
							v := value.Value{Lo: uint64(s) + 7}
							if s < 2 {
								v.Lo = input[s]
							}
							writeSlot(t, ctx, 16, slots.Slot(s), v)
							writeSlot(t, ref, 16, slots.Slot(s), v)
						}
						state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
						ctx.SetState(state)
						ref.SetState(state)
						got, err := ctx.Enter(lease, 0)
						want, refErr := ref.Enter(reference, 0)
						gs, _ := ctx.State()
						ws, _ := ref.State()
						if err != nil || refErr != nil || got != want || gs != ws {
							t.Fatal("loop state/trap", firstPC, reversed, size, pc, input, got, want, gs, ws, err, refErr)
						}
						for s := range f.Slots {
							if got, want := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); got != want {
								t.Fatal("loop frame", firstPC, reversed, size, pc, input, s, got, want)
							}
						}
						gb, wb := make([]byte, size), make([]byte, size)
						memory.Read(0, gb)
						refMemory.Read(0, wb)
						for i := range gb {
							if gb[i] != wb[i] {
								t.Fatal("effects before a later-iteration trap", pc, input, i, gb[i], wb[i])
							}
						}
					}
				}
			}
		}
	}
}

func TestMemoryLoopRejectsUnprovedRecurrencesAndControl(t *testing.T) {
	for _, kind := range []string{"metered", "minimum", "term", "bound", "limit", "polarity", "cursor", "moves", "step", "updates", "base", "nonlinear", "trap", "groups", "overflow", "split", "prefix", "poll"} {
		f := memoryLoopFixture(4, false, false)
		config := Config{Unmetered: true, MemoryMinimumPages: 1}
		switch kind {
		case "metered":
			config.Unmetered = false
		case "minimum":
			config.MemoryMinimumPages = 0
		case "term":
			f.Code[12] = f.Code[13]
		case "bound":
			f.Code[10].Op = 0x20
		case "limit":
			f.Sources[10].Wasm.Bits = 1
		case "polarity":
			f.Code[11].Op = 0x46
		case "cursor":
			f.Code[11].A = 4
		case "moves":
			f.Edges[0].Count = 1
			f.Moves = []slots.Move{{Dst: 0, Src: 0}}
		case "step":
			f.Sources[7].Wasm.Bits = 2
		case "updates":
			f.Code[0].Dst = 1
		case "base":
			f.Code[2].Dst, f.Code[3].B = 0, 0
		case "nonlinear":
			f.Code[0].Op = 0x6c
		case "trap":
			f.Code[7] = slots.Instruction{Op: 0x6d, Dst: 3, A: 0, B: 1, C: slots.NoSlot}
		case "groups":
			f.Code[5].A = 1
		case "overflow":
			f.Sources[5].Wasm.B = 0xffffffff
			f.Code[5].Op = 0x36 // A four-byte access exceeds the entire memory32 span.
		case "split":
			f.Edges[1].PC = 5
		case "prefix", "poll":
			count := 17
			if kind == "poll" {
				count = 250
			}
			prefix := make([]slots.Instruction, count)
			for i := range prefix {
				prefix[i] = slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
			}
			f.Code = append(prefix, f.Code...)
			f.Edges[1].PC += uint32(count)
			if kind == "poll" {
				f.Edges[0].PC = uint32(count)
			}
		}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			if err != nil {
				t.Fatal(kind, arch, err)
			}
			if c.LoopGuards != 0 {
				t.Fatal("unproved loop selected", kind, arch, c.LoopGuards)
			}
		}
	}
}

func TestMemoryLoopPrivateBackedgeKeepsCancellation(t *testing.T) {
	if !native.Available {
		return
	}
	f := memoryLoopFixture(1000000, false, false)
	tail := append([]slots.Instruction(nil), f.Code[7:]...)
	f.Code = f.Code[:7]
	for i := 0; i < 100; i++ {
		f.Code = append(f.Code, slots.Instruction{Op: 0x6c, Dst: 6, A: 6, B: 6, C: slots.NoSlot})
	}
	f.Code = append(f.Code, tail...)
	f.Edges[1].PC = uint32(len(f.Code) - 1)
	c, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	if c.LoopGuards != 1 {
		t.Fatal("cancellation fixture not versioned")
	}
	bindMemory(t, ctx, 1<<20, 1<<20)
	ctx.SetState(native.State{Used: 16, Depth: 1})
	writeSlot(t, ctx, 16, 6, value.Value{Lo: 17})
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
			t.Fatal("loop did not start")
		}
		runtime.Gosched()
	}
	ctx.Cancel()
	select {
	case exit := <-done:
		state, err := ctx.State()
		if err != nil || exit != native.Poll || state.Continuation >= uint64(len(f.Code)) {
			t.Fatal("private edge lost original continuation", exit, state, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("private backedge missed cancellation")
	}
}
