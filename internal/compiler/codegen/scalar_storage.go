package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func scalarStorageCall(f *slots.Function, in slots.Instruction, config Config) (slots.Slot, wasm.ValueType, bool) {
	if in.Op != 0x10 || f.Sources[in.Source].Wasm.A >= config.ScalarImports {
		return 0, 0, false
	}
	outputs := f.Sources[in.Source].Outputs
	var slot slots.Slot
	if in.Count == 0 && len(outputs) == 1 {
		slot = f.Args[in.Extra]
	} else if in.Count == 1 && len(outputs) == 0 {
		slot = f.Args[in.Extra]
	} else {
		return 0, 0, false
	}
	typ := f.Slots[slot]
	return slot, typ, typ == wasm.I32 || typ == wasm.I64
}

// The view contains (value address, validity address) for each module function.
// Only Store-created accessors have nonzero addresses. Ordinary imports keep
// the original checked exit. Scratch registers never clobber allocated values.
func (e *functionEmitter) scalarStorageCall(pc int, config Config, globals uint32) bool {
	in := e.f.Code[pc]
	slot, typ, ok := scalarStorageCall(e.f, in, config)
	if !ok {
		return false
	}
	offset, valid := nativeabi.HeapPairOffset(config.FunctionCount, config.TableCount, uint32(len(config.Globals)), globals, config.MemoryCount)
	index := uint64(e.f.Sources[in.Source].Wasm.A) * 16
	if !valid || offset > 1<<31-33 || index > 1<<31-9 {
		return false
	}
	fallback := e.a.label()
	e.contextLoad(2, nativeabi.RuntimeDataOffset)
	e.pointerLoad(2, 2, offset+24)
	e.registerMove(e.operandRegister(0), e.operandRegister(2), true)
	e.test(true)
	e.branch(fallback, "eq")
	e.pointerLoad(0, 2, uint32(index))
	e.test(true)
	e.branch(fallback, "eq")
	e.pointerLoad(1, 2, uint32(index)+8)
	e.pointerLoad(1, 1, 0)
	e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
	e.registerMove(e.operandRegister(0), e.operandRegister(1), true)
	e.test(true)
	e.branch(fallback, "eq")
	if in.Count == 0 {
		e.callMemory(0, e.operandRegister(2), 0, false, typ == wasm.I32)
		e.slotStore(slot, 0, false)
		if !e.unmetered {
			e.constant(0, 0)
			e.slotStore(slot, 0, true)
		}
	} else {
		e.slotLoad(slot, 0, false, typ == wasm.I64)
		e.callStore(0, e.operandRegister(2), 0)
	}
	e.callExits = append(e.callExits, nativeCallExit{fallback, pc, e.f.Sources[in.Source].Wasm.Offset, e.currentRegisterPlan()})
	return true
}
