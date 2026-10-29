package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func referenceCell(f *slots.Function, in slots.Instruction, config Config) (uint32, bool) {
	index := f.Sources[in.Source].Wasm.A
	if in.Op == 0xd2 {
		offset := uint64(index) * 8
		return uint32(offset), index < config.FunctionCount && offset <= math.MaxInt32 && in.Dst != slots.NoSlot && f.Slots[in.Dst] == wasm.FuncRef
	}
	if in.Op != 0x25 && in.Op != 0x26 || index >= config.TableCount || in.A == slots.NoSlot || f.Slots[in.A] != wasm.I32 && f.Slots[in.A] != wasm.I64 {
		return 0, false
	}
	ref := in.Dst
	if in.Op == 0x26 {
		ref = in.B
	}
	if ref == slots.NoSlot || !f.Slots[ref].IsReference() {
		return 0, false
	}
	return nativeabi.TablePairOffset(config.FunctionCount, index)
}

func (e *functionEmitter) referenceInstruction(in slots.Instruction, config Config) bool {
	offset, ok := referenceCell(e.f, in, config)
	if !ok {
		return false
	}
	if in.Op == 0xd2 {
		if dst, ok := e.registers[in.Dst]; ok {
			if source, cached := e.runtimeRegisters[offset]; cached {
				e.registerMove(dst, source, true)
				return true
			}
		}
		if _, cached := e.runtimeRegisters[offset]; !cached {
			e.contextLoad(2, nativeabi.RuntimeDataOffset)
		}
		e.runtimeLoad(0, offset)
	} else {
		if e.unmetered {
			e.registerTable(in, offset)
			return true
		}
		if _, cached := e.runtimeRegisters[offset]; !cached {
			e.contextLoad(2, nativeabi.RuntimeDataOffset)
		}
		e.slotLoad(in.A, 0, false, e.f.Slots[in.A] == wasm.I64)
		e.runtimeCompare(offset)
		e.branch(e.trapExit(wasm.TrapTableBounds, e.f.Sources[in.Source].Wasm.Offset), "hs")
		base := e.runtimeBase(offset + 8)
		if in.Op == 0x26 {
			e.slotLoad(in.B, 1, false, true)
			e.tableElement(e.operandRegister(1), base, e.operandRegister(0), true, false)
			return true
		}
		e.tableElement(e.operandRegister(0), base, e.operandRegister(0), false, false)
	}
	// Reference storage contains a complete handle word, with canonical Hi=0.
	// The ordinary spill/root-map protocol publishes it at every native exit.
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(in.Dst, 0, true)
	}
	return true
}

// Access allocated reference registers directly. ARM's extended compare and
// UXTW addressing consume the original i32 home without changing its upper
// word. This also preserves raw cold-frame values while truncating guest indices.
func (e *functionEmitter) registerTable(in slots.Instruction, offset uint32) {
	wide := e.f.Slots[in.A] == wasm.I64
	index := e.operandRegister(0)
	if r, ok := e.registers[in.A]; ok && e.a.arch == "arm64" {
		index = r
	} else {
		e.slotLoad(in.A, 0, false, wide)
	}
	if _, cached := e.runtimeRegisters[offset]; !cached {
		e.contextLoad(2, nativeabi.RuntimeDataOffset)
	}
	trap := e.trapExit(wasm.TrapTableBounds, e.f.Sources[in.Source].Wasm.Offset)
	if e.a.arch == "arm64" {
		length, cached := e.runtimeRegisters[offset]
		if !cached {
			e.runtimeLoad(1, offset)
			length = e.operandRegister(1)
		}
		ins := uint32(0xeb20401f) // cmp length, windex, uxtw
		if wide {
			ins = 0xeb00001f // cmp length, xindex
		}
		e.a.word(ins | uint32(index)<<16 | uint32(length)<<5)
		e.a.branch(trap, true, 9) // b.ls: length <= uint32(index)
	} else {
		e.runtimeCompare(offset)
		e.branch(trap, "hs")
	}
	base := e.runtimeBase(offset + 8)
	slot := in.Dst
	store := in.Op == 0x26
	if store {
		slot = in.B
	}
	reg, cached := e.registers[slot]
	if !cached {
		reg = e.operandRegister(0)
		if store {
			e.slotLoad(slot, 1, false, true)
			reg = e.operandRegister(1)
		}
	}
	e.tableElement(reg, base, index, store, !wide)
	if !store && !cached {
		e.slotStore(slot, 0, false)
	}
}

// Element accesses share one encoder with guarded indirect calls. AMD callers
// supply a scratch base (rcx/rdx) and the zero-extended index in rax.
func (e *functionEmitter) tableElement(reg, base, index byte, store, narrow bool) {
	if e.a.arch == "arm64" {
		op := uint32(0xf8607800) // ldr value, [base, index, lsl #3]
		if narrow {
			op &^= 1 << 13
		} // uxtw #3
		if store {
			op &^= 1 << 22
		}
		e.a.word(op | uint32(index)<<16 | uint32(base)<<5 | uint32(reg))
	} else {
		op := byte(0x8b)
		if store {
			op = 0x89
		}
		e.a.bytes(0x48|(reg>>3)<<2, op, (reg&7)<<3|4, 0xc0|base)
	}
}
