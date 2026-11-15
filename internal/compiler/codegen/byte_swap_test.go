package codegen

import (
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func byteSwapFixture() *slots.Function {
	f := &slots.Function{LocalCount: 1, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32}}
	emit := func(op uint32, dst, a, b slots.Slot, bits uint64) {
		index := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: bits}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: index, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	get := func() { emit(0x20, 1, slots.NoSlot, slots.NoSlot, 0) }
	constant := func(v uint64) { emit(0x41, 2, slots.NoSlot, slots.NoSlot, v) }
	get()
	constant(24)
	emit(0x74, 3, 1, 2, 0)
	get()
	constant(8)
	emit(0x74, 4, 1, 2, 0)
	constant(0xff0000)
	emit(0x71, 4, 4, 2, 0)
	emit(0x72, 3, 3, 4, 0)
	get()
	constant(8)
	emit(0x76, 4, 1, 2, 0)
	constant(0xff00)
	emit(0x71, 4, 4, 2, 0)
	emit(0x72, 3, 3, 4, 0)
	get()
	constant(24)
	emit(0x76, 4, 1, 2, 0)
	emit(0x72, 5, 3, 4, 0)
	for slot := slots.Slot(1); slot < 5; slot++ {
		emit(0x41, slot, slots.NoSlot, slots.NoSlot, 0)
	}
	for slot := range f.Slots {
		f.Args = append(f.Args, slots.Slot(slot))
	}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
	return f
}

func TestByteSwapSelectionAndColdFrames(t *testing.T) {
	for _, arch := range []string{"amd64"} {
		f := byteSwapFixture()
		e := functionEmitter{f: f, a: &assembler{arch: arch}, unmetered: true, registers: map[slots.Slot]byte{0: 3, 5: 12}}
		if end, ok := e.combineByteSwap(0); !ok || end != 18 {
			t.Fatal("byte permutation not selected", arch, end, ok)
		}
		for _, mode := range []string{"fuel", "live", "mask", "shift", "foreign", "region", "pressure"} {
			f := byteSwapFixture()
			e := functionEmitter{f: f, a: &assembler{arch: arch}, unmetered: true, registers: map[slots.Slot]byte{0: 3, 5: 12}}
			switch mode {
			case "fuel":
				e.unmetered = false
			case "live":
				f.Code = append(f.Code[:19], f.Code[len(f.Code)-1])
			case "mask":
				f.Sources[6].Wasm.Bits = 0x00fe0000
			case "shift":
				f.Sources[1].Wasm.Bits = 23
			case "foreign":
				f.Sources[3].Wasm.A = 5
			case "region":
				e.regionLimit = 18
			case "pressure":
				delete(e.registers, 5)
			}
			if _, ok := e.combineByteSwap(0); ok || len(e.a.code) != 0 {
				t.Fatal("unsafe byte permutation", arch, mode)
			}
		}
	}
	checkLinearIntegerEntries(t, byteSwapFixture())
	// The rotate/mask form retains two observable masked projections and its
	// final mask literal. Check their complete frame words at every cold PC.
	f := &slots.Function{LocalCount: 1, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32}}
	emit := func(op uint32, dst, a, b slots.Slot, value uint64) {
		index := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: value}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: index, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	emit(0x41, 1, slots.NoSlot, slots.NoSlot, 0xff00ff)
	emit(0x71, 1, 0, 1, 0)
	emit(0x41, 2, slots.NoSlot, slots.NoSlot, 8)
	emit(0x78, 1, 1, 2, 0)
	emit(0x20, 0, slots.NoSlot, slots.NoSlot, 0)
	emit(0x41, 2, slots.NoSlot, slots.NoSlot, 24)
	emit(0x78, 2, 0, 2, 0)
	emit(0x41, 3, slots.NoSlot, slots.NoSlot, 0xff00ff)
	emit(0x71, 2, 2, 3, 0)
	emit(0x72, 4, 1, 2, 0)
	for s := range f.Slots {
		f.Args = append(f.Args, slots.Slot(s))
	}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
	for _, arch := range []string{"amd64"} {
		e := functionEmitter{f: f, a: &assembler{arch: arch}, unmetered: true, registers: map[slots.Slot]byte{0: 3, 1: 12, 2: 13, 3: 14, 4: 15}}
		if end, ok := e.combineByteSwap(0); !ok || end != 9 {
			t.Fatal("live projections", arch, end, ok)
		}
	}
	checkLinearIntegerEntries(t, f)
}

func TestByteSwapApplicationSelection(t *testing.T) {
	for _, producer := range []string{"clang", "rust"} {
		data, err := os.ReadFile("../../../benchmarks/applications/fixtures/sha256-" + producer + ".wasm")
		if err != nil {
			t.Fatal(err)
		}
		m, err := decode.Decode(data, decode.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		compiled := functionModule(t, m)
		for _, arch := range []string{"amd64", "arm64"} {
			var swaps uint32
			for _, f := range compiled.Functions {
				c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 2})
				if err != nil {
					t.Fatal(err)
				}
				swaps += c.ByteSwaps
			}
			t.Log(producer, arch, "byte swaps", swaps)
			if (swaps != 0) != (arch == "amd64") {
				t.Fatal("producer byte permutations were not selected", producer, arch)
			}
		}
	}
}
