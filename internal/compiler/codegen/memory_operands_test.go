package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMemoryOperandFactsDefinitionsCopiesSelectsAndBoundaries(t *testing.T) {
	for _, mode := range []string{"literal", "copy", "arithmetic", "load", "select-equal", "select-different", "select-unknown", "unknown-write", "elided", "elided-local", "remap-unknown", "remap-known", "memory64", "fresh-entry", "not-proven", "arm64"} {
		f := &slots.Function{LocalCount: 2, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I64}}
		emit := func(op uint32, dst, a, b slots.Slot, local uint32, bits uint64) {
			s := uint32(len(f.Sources))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Bits: bits}})
			f.Code = append(f.Code, slots.Instruction{Op: op, Source: s, Dst: dst, A: a, B: b, C: 0})
		}
		emit(0x41, 2, slots.NoSlot, slots.NoSlot, 0, 9)
		emit(0x21, slots.NoSlot, 2, slots.NoSlot, 1, 0)
		emit(0x20, 3, slots.NoSlot, slots.NoSlot, 1, 0)
		emit(0x6a, 4, 0, 3, 0, 0)
		address, want, constant := slots.Slot(4), true, false
		r := memoryRegion{proven: map[int]bool{}, elided: map[int]bool{}}
		arch, start := "amd64", 0
		switch mode {
		case "literal":
			address, constant = 2, true
		case "copy":
			address, constant = 3, true
		case "load":
			emit(0x2d, 5, 2, slots.NoSlot, 0, 0)
			address = 5
		case "select-equal", "select-different", "select-unknown":
			left := slots.Slot(2)
			if mode == "select-different" {
				emit(0x41, 3, slots.NoSlot, slots.NoSlot, 0, 10)
			} else if mode == "select-unknown" {
				left, want = 0, false
			} else {
				constant = true
			}
			emit(0x1b, 5, left, 3, 0, 0)
			address = 5
		case "unknown-write":
			emit(0x23, 4, slots.NoSlot, slots.NoSlot, 0, 0)
			want = false
		case "elided":
			r.elided[0], address, want = true, 2, false
		case "elided-local":
			r.elided[1], address, want = true, 3, false
		case "remap-unknown":
			want = false
		case "remap-known":
			emit(0x41, 6, slots.NoSlot, slots.NoSlot, 0, 0)
			emit(0x6a, 4, 3, 6, 0, 0)
			constant = true
		case "memory64":
			emit(0x42, 7, slots.NoSlot, slots.NoSlot, 0, 9)
			address, want = 7, false
		case "fresh-entry", "not-proven", "arm64":
			want = false
		}
		pc := len(f.Code)
		emit(0x28, 6, address, slots.NoSlot, 0, 0)
		for i, in := range f.Code {
			if isMemoryAccess(in.Op) && mode != "not-proven" {
				r.proven[i] = true
			}
		}
		if mode == "remap-unknown" || mode == "remap-known" {
			local := slots.Slot(0)
			if mode == "remap-known" {
				local = 1
			}
			r.addresses = map[int]memoryAddress{pc: {local: local}}
		}
		if mode == "fresh-entry" {
			start = pc
		} else if mode == "arm64" {
			arch = "arm64"
		}
		e := functionEmitter{a: &assembler{arch: arch}, f: f}
		facts := e.memoryOperandFacts(start, len(f.Code), r)
		got, ok := facts[pc]
		if ok != want || ok && (!got.canonical || got.constant != constant || constant && got.bits != 9) {
			t.Fatal(mode, got, ok, want, constant)
		}
	}
}

func TestDirectMemoryOperandWidthsDisplacementsAndFallback(t *testing.T) {
	for op := uint32(0x28); op <= 0x3e; op++ {
		for _, offset := range []uint32{0, 1, 127, 128, 0x7fffffff, 0x80000000} {
			f, _ := memoryFunction(t, op, offset, true)
			for _, arch := range []string{"amd64", "arm64"} {
				for _, fast := range []bool{false, true} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: fast, MemoryMinimumPages: 65536})
					if err != nil {
						t.Fatal(arch, op, offset, fast, err)
					}
					want := fast && arch == "amd64" && offset <= 0x7fffffff
					if (c.DirectMemoryOperands != 0) != want {
						t.Fatal(arch, op, offset, fast, c.DirectMemoryOperands, want)
					}
				}
			}
		}
	}
	// A minimum-length proof does not canonicalize an incoming frame word.
	f, _ := memoryFunction(t, 0x2d, 0)
	c, err := CompileFunctionWithConfig("amd64", f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 65536})
	if err != nil || c.BoundsElided == 0 || c.DirectMemoryOperands != 0 {
		t.Fatal("unknown upper Lo bits inherited a direct index", err)
	}
}
