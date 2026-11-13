package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type memoryOperand struct {
	bits                uint32
	canonical, constant bool
}

// Entry slots may retain arbitrary upper Lo bits. Only definitions executed in
// this bounded region establish a zero-extended i32 index or exact constant.
// Interior cold entries retain the original checked address calculation.
func (e *functionEmitter) memoryOperandFacts(start, end int, r memoryRegion) map[int]memoryOperand {
	if e.a.arch != "amd64" || len(r.proven) == 0 {
		return nil
	}
	values := map[slots.Slot]memoryOperand{}
	var operands map[int]memoryOperand
	for pc := start; pc < end; pc++ {
		in := e.f.Code[pc]
		src := e.f.Sources[in.Source].Wasm
		a, b, v := values[in.A], values[in.B], memoryOperand{}
		if r.proven[pc] && isMemoryAccess(in.Op) && !nativeVectorMemoryOpcode(in.Op) {
			address := in.A
			if remap, ok := r.addresses[pc]; ok {
				address = remap.local
			}
			operand := values[address]
			if address != slots.NoSlot && e.f.Slots[address] == wasm.I32 && operand.canonical {
				if operands == nil {
					operands = map[int]memoryOperand{}
				}
				operands[pc] = operand
			}
		}
		switch in.Op {
		case 0x20:
			v = values[slots.Slot(src.A)]
		case 0x21, 0x22:
			v = a
			values[slots.Slot(src.A)] = v
		case 0x41:
			v = memoryOperand{uint32(src.Bits), true, true}
		case 0x1b, 0x1c:
			v.canonical = a.canonical && b.canonical
			v.constant = a.constant && b.constant && a.bits == b.bits
			v.bits = a.bits
		default:
			if in.Dst != slots.NoSlot && e.f.Slots[in.Dst] == wasm.I32 {
				v.canonical = isInteger(in.Op) || isFloat(in.Op) || isConversion(in.Op)
				if isMemoryAccess(in.Op) {
					load, _ := memoryAccess(in.Op)
					v.canonical = load
				}
			}
		}
		if r.elided[pc] {
			v = memoryOperand{}
			if in.Op == 0x21 || in.Op == 0x22 {
				values[slots.Slot(src.A)] = v
			}
		}
		if in.Dst != slots.NoSlot {
			values[in.Dst] = v
		}
	}
	return operands
}

// R10 is the stable memory base. Index 4 denotes no index, since RSP is never
// allocated to a guest value. All other indexes use the ordinary SIB form.
func (a *assembler) xmemoryAddress(reg, index byte, offset uint32) {
	a.xscaledMemoryAddress(reg, index, offset, 0)
}

func (a *assembler) xscaledMemoryAddress(reg, index byte, offset uint32, scale byte) {
	mode := byte(0)
	if offset != 0 {
		mode = 0x40
		if offset >= 128 {
			mode = 0x80
		}
	}
	if index == 4 {
		a.bytes(mode | (reg&7)<<3 | 2)
	} else {
		a.bytes(mode|(reg&7)<<3|4, scale<<6|(index&7)<<3|2)
	}
	if mode == 0x40 {
		a.bytes(byte(offset))
	} else if mode == 0x80 {
		a.word(offset)
	}
}

// Fold a proven nonwrapping i32 scale into the load's address generator. Only
// the normal bounded path can consume the three instructions; independent PCs
// retain their original checked bodies. Every unobserved temporary must already
// be overwritten before the next effect, trap, poll or region exit.
func (e *functionEmitter) combineScaledMemory(pc int) (int, bool) {
	if !e.unmetered || e.a.arch != "amd64" || e.f.Code[pc].Op != 0x41 || pc+2 >= e.codeLimit() {
		return 0, false
	}
	scale := e.memoryScales[pc+1]
	if scale == 0 {
		return 0, false
	}
	p, ok := e.planDeadImmediate(pc)
	if !ok || p.end != pc+1 {
		return 0, false
	}
	load := e.f.Code[pc+2]
	if load.Op < 0x28 || load.Op > 0x35 || load.A != p.original || !e.provenPCs[pc+2] || e.elidedPCs[pc+1] || e.elidedPCs[pc+2] {
		return 0, false
	}
	if _, remapped := e.addresses[pc+2]; remapped {
		return 0, false
	}
	offset := e.f.Sources[load.Source].Wasm.MemoryOffset()
	if offset > 0x7fffffff || p.original != load.Dst && !e.overwrittenBeforeExit(pc+3, p.original) {
		return 0, false
	}
	// The range proof came from a normal i32 definition. Use a narrow scratch
	// copy when no home is resident; incoming raw frame upper bits never become
	// part of the 64-bit effective address.
	index, resident := e.registers[e.frameRead(p.in.A)]
	if !resident {
		e.slotLoad(p.in.A, 0, false, false)
		index = 0
	}
	_, size := memoryAccess(load.Op)
	e.memoryAMD64Indexed(load, true, size, index, uint32(offset), scale)
	e.boundsElided++
	e.directMemoryOperands++
	return pc + 2, true
}
