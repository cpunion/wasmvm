package codegen

import (
	"math"
	"math/bits"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Memory bases are stable only during one entry. The owner excludes growth and
// concurrent access; every reentry refreshes the context and these registers.
// amd64 reserves r10/r11, arm64 x16/x17, outside the constrained scratch set.
func (e *functionEmitter) prepareMemory() {
	e.contextLoad(0, nativeabi.MemoryBaseOffset)
	e.test(true)
	e.branch(e.bad, "eq")
	base, length := byte(10), byte(11)
	if e.a.arch == "arm64" {
		base, length = 16, 17
	}
	e.registerMove(base, e.operandRegister(0), true)
	e.contextLoad(0, nativeabi.MemorySizeOffset)
	e.registerMove(length, e.operandRegister(0), true)
	if e.memoryMinimum != 0 {
		e.constant(1, e.memoryMinimum)
		e.compare(true)
		e.branch(e.bad, "lo")
	}
	e.contextLoad(0, nativeabi.ContinuationOffset)
}

func (e *functionEmitter) memoryInstruction(in slots.Instruction, proven bool) {
	if proven {
		e.boundsElided++
	}
	if in.Op == 0x3f {
		if e.a.arch == "arm64" {
			e.a.word(0xd350fe29) // lsr x9, x17, #16
		} else {
			e.registerMove(0, 11, true)
			e.a.bytes(0x48, 0xc1, 0xe8, 16)
		}
		e.slotStore(in.Dst, 0, false)
		if !e.unmetered {
			e.constant(0, 0)
			e.slotStore(in.Dst, 0, true)
		}
		return
	}
	load, size := memoryAccess(in.Op)
	source := e.f.Sources[in.Source].Wasm
	trap := 0
	if !proven {
		trap = e.trapExit(wasm.TrapMemoryBounds, source.Offset)
	}
	// Widen the dynamic i32 address before adding the u32 immediate. The end
	// address is at most 2^33+6, so host arithmetic cannot wrap before this check.
	offset := uint64(source.B)
	vector := nativeVectorMemoryOpcode(in.Op)
	if e.a.arch == "arm64" && proven {
		if r, ok := e.registers[in.A]; ok {
			if offset < 4096 && (!vector || vectorDirectMemory(in.Op)) {
				// Offset the stable base first, keeping the dynamic address off
				// the addition dependency chain. UXTW still widens the i32 index.
				base := byte(16)
				if offset != 0 {
					e.a.word(0x91000209 | uint32(offset)<<10)
					base = 9
				}
				if vector {
					e.vectorMemoryARM64(in, load, size, base, r, 0)
				} else {
					e.memoryARM64(in, load, size, base, r, 0)
				}
				return
			}
			if offset%uint64(size) == 0 && offset/uint64(size) < 4096 && (!vector || vectorDirectMemory(in.Op)) {
				// Larger scaled immediates use a widened complete address.
				e.a.word(0x8b204209 | uint32(r)<<16)
				if vector {
					e.vectorMemoryARM64(in, load, size, 9, 31, uint32(offset)/size)
				} else {
					e.memoryARM64(in, load, size, 9, 31, uint32(offset)/size)
				}
				return
			}
			if vector {
				e.a.word(0x8b204209 | uint32(r)<<16)
				e.memoryOffsetARM64(offset)
				e.vectorMemoryARM64(in, load, size, 9, 31, 0)
				return
			}
		}
	}
	e.slotLoad(in.A, 0, false, false)
	if e.a.arch == "arm64" {
		e.memoryOffsetARM64(offset)
		if !proven {
			e.a.word(0x9100012a | size<<10) // add x10, x9, #size
			e.a.word(0xeb11015f)            // cmp x10, x17
			e.a.branch(trap, true, 8)       // unsigned end > length
		}
		e.a.word(0x8b090209) // add x9, x16, x9
		if vector {
			e.vectorMemoryARM64(in, load, size, 9, 31, 0)
		} else {
			e.memoryARM64(in, load, size, 9, 31, 0)
		}
	} else {
		if offset != 0 {
			if offset <= math.MaxInt32 {
				e.a.bytes(0x48, 0x05)
				e.a.word(uint32(offset))
			} else {
				e.constant(1, offset)
				e.a.bytes(0x48, 0x01, 0xc8)
			}
		}
		if !proven {
			e.a.bytes(0x48, 0x8d, 0x48, byte(size)) // lea rcx,[rax+size]
			e.a.bytes(0x4c, 0x39, 0xd9)             // cmp rcx,r11
			e.a.branch(trap, true, 7)
		}
		if vector {
			e.vectorMemoryAMD64(in, load, size)
		} else {
			e.memoryAMD64(in, load, size)
		}
	}
}

func (e *functionEmitter) memoryOffsetARM64(offset uint64) {
	if offset == 0 {
		return
	}
	if offset < 4096 {
		e.a.word(0x91000129 | uint32(offset)<<10)
	} else {
		e.constant(1, offset)
		e.a.word(0x8b0a0129)
	}
}

func (e *functionEmitter) memoryAddressARM64(ins uint32, r, base, index byte, scaledOffset uint32) {
	if index != 31 {
		// Register offset with UXTW preserves Wasm i32 address wrapping.
		e.a.word(ins&^(1<<24) | 1<<21 | uint32(index)<<16 | 0x4800 | uint32(base)<<5 | uint32(r))
	} else {
		e.a.word(ins | scaledOffset<<10 | uint32(base)<<5 | uint32(r))
	}
}

func (e *functionEmitter) memoryARM64(in slots.Instruction, load bool, size uint32, base, index byte, scaledOffset uint32) {
	slot := in.B
	if load {
		slot = in.Dst
	}
	if r, ok := e.floatRegisters[slot]; ok {
		ins := uint32(0xbd000000)
		if size == 8 {
			ins = 0xfd000000
		}
		if load {
			ins |= 1 << 22
		}
		e.memoryAddressARM64(ins, r, base, index, scaledOffset)
		return
	}
	r, cached := e.registers[slot]
	if !cached {
		r = 10
		if !load {
			e.slotLoad(slot, 1, false, true)
		}
	}
	if load {
		ins := uint32(0x39400000) | uint32(bits.TrailingZeros32(size))<<30
		switch in.Op {
		case 0x2c:
			ins = 0x39c00000
		case 0x2e:
			ins = 0x79c00000
		case 0x30:
			ins = 0x39800000
		case 0x32:
			ins = 0x79800000
		case 0x34:
			ins = 0xb9800000
		}
		e.memoryAddressARM64(ins, r, base, index, scaledOffset)
		if !cached {
			e.slotStore(slot, 1, false)
		}
		if !e.unmetered {
			e.constant(0, 0)
			e.slotStore(slot, 0, true)
		}
	} else {
		ins := uint32(0x39000000) | uint32(bits.TrailingZeros32(size))<<30
		e.memoryAddressARM64(ins, r, base, index, scaledOffset)
	}
}

func (e *functionEmitter) memoryAMD64(in slots.Instruction, load bool, size uint32) {
	slot := in.B
	if load {
		slot = in.Dst
	}
	if r, ok := e.floatRegisters[slot]; ok {
		prefix := byte(0xf3)
		if size == 8 {
			prefix = 0xf2
		}
		e.a.bytes(prefix)
		e.a.bytes(0x41 | (r>>3)<<2)
		opcode := byte(0x11)
		if load {
			opcode = 0x10
		}
		e.a.bytes(0x0f, opcode, (r&7)<<3|4, 2)
		return
	}
	r, cached := e.registers[slot]
	if !cached {
		r = 1
		if !load {
			e.slotLoad(slot, 1, false, true)
		}
	}
	if load {
		signed := in.Op == 0x2c || in.Op == 0x2e || in.Op == 0x30 || in.Op == 0x32 || in.Op == 0x34
		wide := size == 8 || signed && in.Op >= 0x30
		rex := byte(0x41) | (r>>3)<<2
		if wide {
			rex |= 8
		}
		if rex != 0x40 {
			e.a.bytes(rex)
		}
		if size < 4 {
			opcode := byte(0xb6)
			if size == 2 {
				opcode++
			}
			if signed {
				opcode += 8
			}
			e.a.bytes(0x0f, opcode, (r&7)<<3|4, 2)
		} else if signed {
			e.a.bytes(0x63, (r&7)<<3|4, 2)
		} else {
			e.a.bytes(0x8b, (r&7)<<3|4, 2)
		}
		if !cached {
			e.slotStore(slot, 1, false)
		}
		if !e.unmetered {
			e.constant(0, 0)
			e.slotStore(slot, 0, true)
		}
	} else {
		if size == 2 {
			e.a.bytes(0x66)
		}
		rex := byte(0x41) | (r>>3)<<2
		if size == 8 {
			rex |= 8
		}
		if rex != 0x40 || size == 1 && r >= 4 {
			e.a.bytes(rex)
		}
		opcode := byte(0x89)
		if size == 1 {
			opcode = 0x88
		}
		e.a.bytes(opcode, (r&7)<<3|4, 2)
	}
}

// One width inventory drives both bounds proofs and target emission.
func memoryAccess(op uint32) (load bool, size uint32) {
	if nativeVectorMemoryOpcode(op) {
		sub := op & 0xffff
		if sub <= 11 {
			return sub != 11, [...]uint32{16, 8, 8, 8, 8, 8, 8, 1, 2, 4, 8, 16}[sub]
		}
		if sub <= 91 {
			return sub < 88, 1 << ((sub - 84) % 4)
		}
		return true, 4 << (sub - 92)
	}
	if op <= 0x35 {
		return true, [14]uint32{4, 8, 4, 8, 1, 1, 2, 2, 1, 1, 2, 2, 4, 4}[op-0x28]
	}
	return false, [9]uint32{4, 8, 4, 8, 1, 2, 1, 2, 4}[op-0x36]
}

func isMemoryAccess(op uint32) bool {
	return op >= 0x28 && op <= 0x3e || nativeVectorMemoryOpcode(op)
}
