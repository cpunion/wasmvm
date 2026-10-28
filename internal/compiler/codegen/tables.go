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
	if in.Op != 0x25 && in.Op != 0x26 || index >= config.TableCount || in.A == slots.NoSlot || f.Slots[in.A] != wasm.I32 {
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
		if _, cached := e.runtimeRegisters[offset]; !cached {
			e.contextLoad(2, nativeabi.RuntimeDataOffset)
		}
		e.runtimeLoad(0, offset)
	} else {
		if _, cached := e.runtimeRegisters[offset]; !cached {
			e.contextLoad(2, nativeabi.RuntimeDataOffset)
		}
		e.slotLoad(in.A, 0, false, false)
		e.runtimeCompare(offset)
		e.branch(e.trapExit(wasm.TrapTableBounds, e.f.Sources[in.Source].Wasm.Offset), "hs")
		base := e.runtimeBase(offset + 8)
		if in.Op == 0x25 {
			if e.a.arch == "arm64" {
				e.a.word(0xf8697809 | uint32(base)<<5) // ldr x9, [base, x9, lsl #3]
			} else {
				e.a.bytes(0x48, 0x8b, 0x04, 0xc2) // mov rax, [rdx + rax*8]
			}
		} else {
			e.slotLoad(in.B, 1, false, true)
			if e.a.arch == "arm64" {
				e.a.word(0xf829780a | uint32(base)<<5) // str x10, [base, x9, lsl #3]
			} else {
				e.a.bytes(0x48, 0x89, 0x0c, 0xc2) // mov [rdx + rax*8], rcx
			}
			return true
		}
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
