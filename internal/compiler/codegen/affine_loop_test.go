package codegen

import (
	"bytes"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func affineLoopFixture(stride uint32, prefix int) *slots.Function {
	f := &slots.Function{LocalCount: 3, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32}}
	f.Code = []slots.Instruction{
		{Op: 0x6a, Dst: 4, A: 0, B: 2, C: slots.NoSlot},
		{Op: 0x3a, Dst: slots.NoSlot, A: 4, B: 2, C: slots.NoSlot},
		{Op: 0x41, Dst: 5, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		{Op: 0x6a, Dst: 4, A: 4, B: 5, C: slots.NoSlot},
		{Op: 0x3a, Dst: slots.NoSlot, A: 4, B: 2, C: slots.NoSlot},
		{Op: 0x6a, Dst: 4, A: 1, B: 2, C: slots.NoSlot},
		{Op: 0x3a, Dst: slots.NoSlot, A: 4, B: 2, C: slots.NoSlot},
		{Op: 0x41, Dst: 5, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		{Op: 0x6a, Dst: 4, A: 2, B: 5, C: slots.NoSlot},
		{Op: 0x22, Dst: 2, A: 4, B: slots.NoSlot, C: slots.NoSlot},
		{Op: 0x41, Dst: 5, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
		{Op: 0x47, Dst: 4, A: 2, B: 5, C: slots.NoSlot},
		{Op: slots.ControlBase + uint32(ir.If), Dst: slots.NoSlot, A: 4, B: slots.NoSlot, C: slots.NoSlot, Count: 2},
		{Op: slots.ControlBase + uint32(ir.Return), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 6},
	}
	for pc := range f.Code {
		f.Code[pc].Source = uint32(pc)
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: f.Code[pc].Op, Offset: uint64(100 + pc)}})
	}
	f.Sources[2].Wasm.Bits = 8
	f.Sources[7].Wasm.Bits = uint64(stride)
	f.Sources[9].Wasm.A = 2
	f.Sources[10].Wasm.Bits = 12
	for i := 0; i < prefix; i++ {
		f.Code = append([]slots.Instruction{{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}, f.Code...)
	}
	f.Edges = []slots.Edge{{PC: 0}, {PC: uint32(len(f.Code) - 1)}}
	for s := range f.Slots {
		f.Args = append(f.Args, slots.Slot(s))
	}
	return f
}

func TestAffineLoopMultipleBasesStridesAliasesWrapAndInteriorPCs(t *testing.T) {
	for _, stride := range []uint32{1, 3, 4, 0x80000000, 0xffffffff} {
		for _, prefix := range []int{0, 17} {
			f := affineLoopFixture(stride, prefix)
			for _, arch := range []string{"arm64", "amd64"} {
				c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
				if err != nil || c.LoopGuards != 1 || c.MemoryGuards != 2 {
					t.Fatal(arch, stride, prefix, c, err)
				}
			}
			compareAffineLoop(t, f)
		}
	}
}

func compareAffineLoop(t *testing.T, f *slots.Function, entries ...int) {
	t.Helper()
	if !native.Available {
		return
	}
	_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	_, reference, ref := functionContext(t, f, true)
	const size = 65536
	memory := bindMemory(t, ctx, size, size)
	refMemory := bindMemory(t, ref, size, size)
	initial := bytes.Repeat([]byte{0xa5}, size)
	if len(entries) == 0 {
		for pc := range f.Code {
			entries = append(entries, pc)
		}
	}
	for _, pc := range entries {
		for _, input := range [][3]uint64{{0, 80, 0}, {5, 5, 3}, {0, 80, 10}, {0, 80, 12}, {size - 10, 0, 0}, {0, size - 4, 0}, {0xfffffffd, 80, 3}, {0xffffffff, 1, 0}, {1 << 32, 80, 1<<63 | 3}} {
			memory.Write(0, initial)
			refMemory.Write(0, initial)
			for s := range f.Slots {
				v := value.Value{Lo: uint64(s) + 7}
				if f.Slots[s] == wasm.V128 {
					v.Hi = 0x123456789abcdef0
				}
				if s < 3 {
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
				t.Fatal("state/trap", pc, input, got, want, gs, ws, err, refErr)
			}
			for s := range f.Slots {
				if got, want := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); got != want {
					t.Fatal("canonical frame", pc, input, s, got, want)
				}
			}
			gb, wb := make([]byte, size), make([]byte, size)
			memory.Read(0, gb)
			refMemory.Read(0, wb)
			if !bytes.Equal(gb, wb) {
				t.Fatal("effects before trap", pc, input)
			}
		}
	}
}

func TestAffineLoopInvariantAndNegativeOffsets(t *testing.T) {
	for _, kind := range []string{"invariant", "negative-bias", "negative-fold"} {
		t.Run(kind, func(t *testing.T) {
			f := affineLoopFixture(3, 0)
			switch kind {
			case "invariant":
				f.Code[5].Op = 0x20
				f.Sources[5].Wasm.A = 1
			case "negative-bias":
				f.Sources[2].Wasm.Bits = 0xfffffff8
			case "negative-fold":
				f = memoryLoopFixture(12, false, false)
				f.Sources[2].Wasm.Bits = 0xffffffde
			}
			for _, arch := range []string{"arm64", "amd64"} {
				c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
				if err != nil || c.LoopGuards != 1 {
					t.Fatal(kind, arch, c, err)
				}
			}
			compareAffineLoop(t, f)
		})
	}
}
