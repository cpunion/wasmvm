package codegen

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestIntegerRegisterInventoryExcludesReservedAndImmutableRegisters(t *testing.T) {
	for _, memory := range []bool{false, true} {
		e := functionEmitter{a: &assembler{arch: "amd64"}, usesMemory: memory}
		want := []byte{3, 12, 13, 14, 15, 9, 10, 11}
		if memory {
			want = want[:6]
		}
		if got := e.integerRegisterPool(); !reflect.DeepEqual(got, want) {
			t.Fatal("changed x86 inventory", memory, got, want)
		}
		e.a.arch = "arm64"
		e.registers = map[slots.Slot]byte{0: 19, 1: 1, 2: 3}
		e.constantRegisters = map[uint64]byte{123: 2, 456: 5}
		e.runtimeRegisters = map[uint32]byte{0: 4, 8: 6}
		want = []byte{19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 1, 3, 7, 8}
		if got := e.integerRegisterPool(); !reflect.DeepEqual(got, want) {
			t.Fatal("reserved or immutable register in mutable inventory", memory, got, want)
		}
		if got := e.spareCallerRegisters(); !reflect.DeepEqual(got, []byte{7, 8}) {
			t.Fatal("overlapping caller register classes", memory, got)
		}
	}
}

func TestCallerRegistersPreserveLargeFrameReferencesAndColdEntries(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.FuncRef, wasm.ExternRef} {
		const base, count = 4096, 24
		f := &slots.Function{LocalCount: base + count}
		for i := 0; i < base+count; i++ {
			f.Slots = append(f.Slots, wasm.I32)
		}
		for i := 0; i < count; i++ {
			s := slots.Slot(base + i)
			f.Slots[s] = typ
			f.Args = append(f.Args, s)
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(base + (i+1)%count)}})
			f.Code = append(f.Code, slots.Instruction{Op: 0x20, Source: uint32(i), Dst: s, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
		f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: count, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		for _, arch := range []string{"amd64", "arm64"} {
			if _, err := CompileUnmeteredFunction(arch, f, DefaultLimits()); err != nil {
				t.Fatal(arch, typ, err)
			}
		}
		if !native.Available {
			continue
		}
		compiled, lease, ctx := functionContext(t, f, true)
		if typ.IsReference() && len(compiled.ReferenceSlots) != count {
			t.Fatal("missing frame roots", typ, compiled.ReferenceSlots)
		}
		for pc := range f.Code {
			want := make([]value.Value, count)
			for i := range want {
				want[i].Lo = 0x8000000000000001 + uint64(i)
				if typ == wasm.I32 {
					want[i].Lo = uint64(uint32(want[i].Lo))
				}
				writeSlot(t, ctx, 16, slots.Slot(base+i), want[i])
			}
			for i := pc; i < count; i++ {
				want[i] = want[(i+1)%count]
			}
			if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
				t.Fatal(err)
			}
			if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
				t.Fatal(typ, pc, exit, err)
			}
			for i, v := range want {
				if got := readSlot(t, ctx, 16, slots.Slot(base+i)); got != v {
					t.Fatal("caller register or spill corrupted", typ, pc, i, got, v)
				}
			}
		}
	}
}

func TestLeafExpansionUsesCallerRegistersWithoutClobberingLiveSlots(t *testing.T) {
	// This pressure level specifically exercises the expanded AArch64 pool.
	if !native.Available || runtime.GOARCH != "arm64" {
		return
	}
	literal := uint64(0x123456789abcdef0)
	callee := leafFixture(0x85, &literal)
	f := leafCaller(callee, 12, 14)
	// Keep ten other locals above the two inputs in the allocation priority.
	// The inputs and output then compete for the caller-saved inventory.
	tail := append([]slots.Instruction(nil), f.Code[14:]...)
	f.Code = f.Code[:14]
	for i := slots.Slot(2); i < 12; i++ {
		f.Code = append(f.Code, slots.Instruction{Op: 0x84, Dst: i, A: i, B: i, C: slots.NoSlot})
	}
	f.Code = append(f.Code, tail...)
	compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, DirectCallees: map[uint32]*slots.Function{7: callee}})
	if compiled.InlinedCalls != 1 {
		t.Fatal("call not expanded with available caller registers")
	}
	for pc := range f.Code {
		want := make([]value.Value, len(f.Slots))
		for i := range want {
			want[i].Lo = 0xfedcba9876543210 + uint64(i)
			writeSlot(t, ctx, 16, slots.Slot(i), want[i])
		}
		if pc <= len(f.Code)-2 {
			want[12].Lo = want[0].Lo ^ literal
		}
		if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
			t.Fatal(err)
		}
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
			t.Fatal(pc, exit, err)
		}
		for i, v := range want {
			if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
				t.Fatal(pc, i, got, v)
			}
		}
	}
}
