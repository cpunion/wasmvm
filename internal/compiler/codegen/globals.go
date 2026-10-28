package codegen

import (
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func globalCell(f *slots.Function, in slots.Instruction, config Config) (uint32, wasm.ValueType, bool) {
	if in.Op != 0x23 && in.Op != 0x24 {
		return 0, 0, false
	}
	index := f.Sources[in.Source].Wasm.A
	if uint64(index) >= uint64(len(config.Globals)) {
		return 0, 0, false
	}
	g := config.Globals[index]
	if g.Value != wasm.I32 && g.Value != wasm.I64 && g.Value != wasm.F32 && g.Value != wasm.F64 && g.Value != wasm.V128 || in.Op == 0x24 && !g.Mutable {
		return 0, 0, false
	}
	slot := in.Dst
	if in.Op == 0x24 {
		slot = in.A
	}
	if slot == slots.NoSlot || uint64(slot) >= uint64(len(f.Slots)) || f.Slots[slot] != g.Value {
		return 0, 0, false
	}
	offset, ok := nativeabi.GlobalPointerOffset(config.FunctionCount, config.TableCount, index)
	return offset, g.Value, ok
}

func (e *functionEmitter) prepareGlobals(config Config) []uint32 {
	seen := map[uint32]bool{}
	for _, in := range e.f.Code {
		if _, _, ok := globalCell(e.f, in, config); ok {
			seen[e.f.Sources[in.Source].Wasm.A] = true
		}
	}
	var indices []uint32
	for index := range seen {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })
	if len(indices) != 0 {
		// The driver supplies this trusted view before any guest instruction,
		// including arbitrary cold continuations. No Go pointer enters the ABI.
		e.contextLoad(0, nativeabi.RuntimeDataOffset)
		e.test(true)
		e.branch(e.bad, "eq")
		e.contextLoad(0, nativeabi.ContinuationOffset)
	}
	return indices
}

func (e *functionEmitter) globalInstruction(in slots.Instruction, config Config) bool {
	offset, typ, ok := globalCell(e.f, in, config)
	if !ok {
		return false
	}
	base := e.operandRegister(2)
	if r, cached := e.runtimeRegisters[offset]; cached {
		base = r
	} else {
		e.contextLoad(2, nativeabi.RuntimeDataOffset)
		e.runtimeLoad(2, offset)
	}
	if typ == wasm.V128 {
		reg := byte(0)
		if in.Op == 0x23 {
			reg = e.vectorDestination(in.Dst)
		} else {
			reg = e.vectorOperand(in.A, 0)
		}
		if e.a.arch == "arm64" {
			op := uint32(0x3dc00000) // ldr q, [cell]
			if in.Op == 0x24 {
				op = 0x3d800000
			}
			e.a.word(op | uint32(base)<<5 | uint32(reg))
		} else {
			e.a.bytes(0xf3)
			if reg >= 8 {
				e.a.bytes(0x44)
			}
			op := byte(0x6f)
			if in.Op == 0x24 {
				op = 0x7f
			}
			e.a.bytes(0x0f, op, (reg&7)<<3|2) // movdqu xmm, [rdx] / [rdx], xmm
		}
		if in.Op == 0x23 {
			e.vectorStore(in.Dst, reg)
		}
		return true
	}
	if e.unmetered {
		slot := in.Dst
		if in.Op == 0x24 {
			slot = in.A
		}
		if r, ok := e.floatRegisters[slot]; ok {
			e.globalRegisterMemory(r, base, in.Op == 0x24, typ)
			return true
		}
		if r, ok := e.registers[slot]; ok {
			e.globalRegisterMemory(r, base, in.Op == 0x24, typ)
			return true
		}
	}
	e.registerMove(e.operandRegister(2), base, true)
	if in.Op == 0x23 {
		e.pointerLoad(0, 2, 0)
		e.slotStore(in.Dst, 0, false)
		if !e.unmetered {
			e.pointerLoad(0, 2, 8)
			e.slotStore(in.Dst, 0, true)
		}
	} else {
		e.slotLoad(in.A, 0, false, true)
		e.globalWordStore(0)
		if !e.unmetered {
			e.slotLoad(in.A, 0, true, true)
			e.globalWordStore(8)
		}
	}
	return true
}

// Scalar cells are canonical in this tier: Hi and the upper f32 word start at
// zero and remain zero. Access allocated registers directly, avoiding FP/GPR
// transfers and redundant zero stores. Precise entries retain two-word copies.
func (e *functionEmitter) globalRegisterMemory(reg, base byte, store bool, typ wasm.ValueType) {
	floating := typ == wasm.F32 || typ == wasm.F64
	if e.a.arch == "arm64" {
		op := uint32(0xf9400000)
		if floating {
			op = 0xbd400000
			if typ == wasm.F64 {
				op = 0xfd400000
			}
		}
		if store {
			op &^= 1 << 22
		}
		e.a.word(op | uint32(base)<<5 | uint32(reg))
	} else {
		op, rex := byte(0x8b), byte(0x48)|(reg>>3)<<2
		if store {
			op = 0x89
		}
		if floating {
			prefix := byte(0xf3)
			if typ == wasm.F64 {
				prefix = 0xf2
			}
			e.a.bytes(prefix)
			rex, op = 0x40|(reg>>3)<<2, 0x10
			if store {
				op = 0x11
			}
		}
		if rex != 0x40 {
			e.a.bytes(rex)
		}
		if floating {
			e.a.bytes(0x0f)
		}
		e.a.bytes(op, (reg&7)<<3|base)
	}
}

func (e *functionEmitter) globalWordStore(offset uint32) {
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x89, 0x42, byte(offset)) // mov [rdx+offset], rax
	} else {
		e.a.word(0xf9000169 | offset/8<<10) // str x9, [x11, offset]
	}
}
