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
	if _, cached := e.runtimeRegisters[offset]; !cached {
		e.contextLoad(2, nativeabi.RuntimeDataOffset)
	}
	e.runtimeLoad(2, offset)
	if typ == wasm.V128 {
		reg := byte(0)
		if in.Op == 0x23 {
			reg = e.vectorDestination(in.Dst)
		} else {
			reg = e.vectorOperand(in.A, 0)
		}
		if e.a.arch == "arm64" {
			op := uint32(0x3dc00000) // ldr q, [x11]
			if in.Op == 0x24 {
				op = 0x3d800000
			}
			e.a.word(op | 11<<5 | uint32(reg))
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
		e.slotLoad(in.A, 0, true, true)
		e.globalWordStore(8)
	}
	return true
}

func (e *functionEmitter) globalWordStore(offset uint32) {
	if e.a.arch == "amd64" {
		e.a.bytes(0x48, 0x89, 0x42, byte(offset)) // mov [rdx+offset], rax
	} else {
		e.a.word(0xf9000169 | offset/8<<10) // str x9, [x11, offset]
	}
}
