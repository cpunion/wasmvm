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
		a.bytes(mode|(reg&7)<<3|4, (index&7)<<3|2)
	}
	if mode == 0x40 {
		a.bytes(byte(offset))
	} else if mode == 0x80 {
		a.word(offset)
	}
}
