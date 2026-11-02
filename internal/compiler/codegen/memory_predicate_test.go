package codegen

import (
	"bytes"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func predicateFixture(read, reversed bool) *slots.Function {
	f := unrollFixture(0, reversed, false, 0)
	if !read {
		// The next iteration replaces this comparison slot with its literal.
		f.Code[len(f.Code)-3].Dst = 4
		f.Code[len(f.Code)-2].A = 4
		return f
	}
	// Make the previous comparison observable in the accumulator, before any
	// write to it. A missing intermediate predicate changes the returned frame.
	in := slots.Instruction{Op: 0x6a, Source: uint32(len(f.Sources)), Dst: 2, A: 2, B: 6, C: slots.NoSlot}
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x6a, Offset: 5000}})
	f.Code = append([]slots.Instruction{in}, f.Code...)
	for i := range f.Edges {
		if f.Edges[i].PC != 0 {
			f.Edges[i].PC++
		}
	}
	return f
}

func TestMemoryUnrollComparisonLiveness(t *testing.T) {
	for _, test := range []struct{ read, reversed bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		read := test.read
		f := predicateFixture(read, test.reversed)
		for _, arch := range []string{"amd64", "arm64"} {
			e := functionEmitter{f: f, a: &assembler{arch: arch}, unmetered: true, usesMemory: true, memoryMinimum: 65536}
			u := e.memoryRegions()[0].unroll
			if u == nil || u.keepPredicate != read {
				t.Fatal("comparison liveness", arch, read, u)
			}
		}
		if !native.Available {
			continue
		}
		_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
		_, refLease, ref := functionContext(t, f, true)
		memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
		initial, gotBytes, wantBytes := make([]byte, 65536), make([]byte, 65536), make([]byte, 65536)
		for i := range initial {
			initial[i] = byte(i*193 + 127)
		}
		for pc := range f.Code {
			for _, input := range [][2]uint64{{0, 1}, {0, 3}, {0, 4}, {0, 5}, {0, 8}, {0, 9}, {3, 8}, {7, 3}, {0xffffffff, 4}} {
				memory.Write(0, initial)
				refMemory.Write(0, initial)
				for s := range f.Slots {
					bits := uint64(s*17 + 5)
					if s < 2 {
						bits = input[s]
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
					t.Fatal("comparison state", read, pc, input, got, want, gs, ws, err, refErr)
				}
				for s := range f.Slots {
					if got, want := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); got != want {
						t.Fatal("comparison frame", read, pc, input, s, got, want)
					}
				}
				memory.Read(0, gotBytes)
				refMemory.Read(0, wantBytes)
				if !bytes.Equal(gotBytes, wantBytes) {
					t.Fatal("comparison memory", read, pc, input)
				}
			}
		}
	}
}

func TestScalarSlotReadPrecedesWrite(t *testing.T) {
	for _, in := range []slots.Instruction{
		{Op: 0x6a, A: 2, B: 3, C: slots.NoSlot, Dst: 2},
		{Op: 0x20, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: 2},
		{Op: 0x22, A: 2, B: slots.NoSlot, C: slots.NoSlot, Dst: 2},
	} {
		if read, write := scalarSlotUse(in, 2, 2); !read || !write {
			t.Fatal("lost aliased read", in, read, write)
		}
	}
}
