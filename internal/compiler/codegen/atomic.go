package codegen

import (
	"math"
	"math/bits"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Wait/notify retain checked queue helpers. Other atomic effects use the same
// target memory contract as the host adapter, including mixed-size overlaps.
func (e *functionEmitter) atomicInstruction(in slots.Instruction, pc uint32) bool {
	if !e.nativeThreads || in.Op>>16 != 0xfe {
		return false
	}
	sub := in.Op & 0xffff
	if sub == 3 {
		if e.a.arch == "arm64" {
			e.a.word(0xd5033bbf) // dmb ish
		} else {
			e.a.bytes(0x0f, 0xae, 0xf0) // mfence
		}
		return true
	}
	if sub < 0x10 || sub > 0x4e || in.A == slots.NoSlot {
		return false
	}
	width := [...]uint32{4, 8, 1, 2, 1, 2, 4}[(sub-0x10)%7]
	group := (sub - 0x10) / 7
	source := e.f.Sources[in.Source].Wasm
	if source.Op != in.Op || group != 0 && in.B == slots.NoSlot || group == 8 && in.C == slots.NoSlot {
		return false
	}
	if e.memoryViews {
		e.selectMemory(source.Memory)
	}
	bound := e.trapExit(wasm.TrapMemoryBounds, source.Offset)
	unaligned := e.trapExit(wasm.TrapAtomicUnaligned, source.Offset)
	e.atomicAddress(in, width, source.MemoryOffset(), bound, unaligned)
	if e.a.arch == "arm64" {
		e.atomicARM64(in, width, group, pc)
	} else {
		e.atomicAMD64(in, width, group, pc)
	}
	if group != 1 {
		e.slotStore(in.Dst, 0, false)
		if !e.unmetered {
			e.constant(0, 0)
			e.slotStore(in.Dst, 0, true)
		}
	}
	return true
}

func (e *functionEmitter) atomicAddress(in slots.Instruction, width uint32, offset uint64, bound, unaligned int) {
	e.slotLoad(in.A, 0, false, e.f.Slots[in.A] == wasm.I64)
	if e.a.arch == "arm64" {
		if offset != 0 {
			if offset < 4096 {
				e.a.word(0xb1000129 | uint32(offset)<<10) // adds x9, x9, #offset
			} else {
				e.constant(1, offset)
				e.a.word(0xab0a0129)
			}
			e.a.branch(bound, true, 2)
		}
		e.a.word(0xb100012a | width<<10) // adds x10, x9, #width
		e.a.branch(bound, true, 2)
		e.a.word(0xeb11015f) // cmp x10, x17
		e.a.branch(bound, true, 8)
		if width != 1 {
			e.constant(1, uint64(width-1))
			e.a.word(0xea0a013f) // tst x9, x10
			e.branch(unaligned, "ne")
		}
		e.a.word(0x8b09020d) // add x13, x16, x9
		return
	}
	if offset != 0 {
		if offset <= math.MaxInt32 {
			e.a.bytes(0x48, 0x05)
			e.a.word(uint32(offset))
		} else {
			e.constant(1, offset)
			e.a.bytes(0x48, 0x01, 0xc8)
		}
		e.a.branch(bound, true, 2)
	}
	e.registerMove(1, 0, true)
	e.a.bytes(0x48, 0x83, 0xc1, byte(width))
	e.a.branch(bound, true, 2)
	e.a.bytes(0x4c, 0x39, 0xd9) // cmp rcx, r11
	e.a.branch(bound, true, 7)
	if width != 1 {
		e.a.bytes(0xa8, byte(width-1)) // test al, width-1
		e.branch(unaligned, "ne")
	}
	e.a.bytes(0x49, 0x8d, 0x14, 0x02) // lea rdx, [r10+rax]
}

func (e *functionEmitter) atomicARM64(in slots.Instruction, width, group uint32, pc uint32) {
	size := uint32(bits.TrailingZeros32(width)) << 30
	if group == 0 {
		e.a.word(0x08dffda9 | size) // ldar[b/h] x/w9, [x13]
		return
	}
	e.slotLoad(in.B, 1, false, width == 8)
	if group == 1 {
		e.a.word(0x089ffdaa | size) // stlr[b/h] x/w10, [x13]
		return
	}
	if group == 8 {
		if width < 4 {
			e.a.word(0x5300014a | (width*8-1)<<10) // uxtb/uxth w10, w10
		}
		e.slotLoad(in.C, 2, false, width == 8)
	}
	loop := e.a.label()
	done := e.a.label()
	e.a.mark(loop)
	// Poll before the exclusive load; cancellation cannot retain a reservation.
	e.interrupt(pc)
	e.a.word(0x085ffda9 | size) // ldaxr[b/h] x/w9, [x13]
	if group == 8 {
		matched := e.a.label()
		e.compare(width == 8)
		e.branch(matched, "eq")
		e.a.word(0xd5033f5f) // clrex
		e.a.branch(done, false, 0)
		e.a.mark(matched)
		e.a.word(0x080ffdab | size) // stlxr[b/h] w15, x/w11, [x13]
	} else {
		var op uint32
		switch group {
		case 2:
			op = 0x0b0a012e
		case 3:
			op = 0x4b0a012e
		case 4:
			op = 0x0a0a012e
		case 5:
			op = 0x2a0a012e
		case 6:
			op = 0x4a0a012e
		case 7:
			e.registerMove(14, 10, width == 8)
		}
		if group != 7 {
			e.a.arm(width == 8, op)
		}
		e.a.word(0x080ffdae | size) // stlxr[b/h] w15, x/w14, [x13]
	}
	e.a.compareBranch(loop, 15, false, true)
	e.a.mark(done)
}

func (e *functionEmitter) atomicPrefix(width uint32, locked, high bool) {
	if locked {
		e.a.bytes(0xf0)
	}
	if width == 2 {
		e.a.bytes(0x66)
	}
	rex := byte(0x40)
	if width == 8 {
		rex |= 8
	}
	if high {
		rex |= 4
	}
	if rex != 0x40 {
		e.a.bytes(rex)
	}
}

func (e *functionEmitter) atomicLoadAMD64(width uint32) {
	switch width {
	case 1:
		e.a.bytes(0x0f, 0xb6, 0x02)
	case 2:
		e.a.bytes(0x0f, 0xb7, 0x02)
	case 4:
		e.a.bytes(0x8b, 0x02)
	case 8:
		e.a.bytes(0x48, 0x8b, 0x02)
	}
}

func (e *functionEmitter) atomicAMD64(in slots.Instruction, width, group uint32, pc uint32) {
	if group == 0 {
		e.atomicLoadAMD64(width)
		return
	}
	byteOp := width == 1
	if group == 1 || group == 7 {
		e.slotLoad(in.B, 0, false, width == 8)
		e.atomicPrefix(width, false, false)
		op := byte(0x87)
		if byteOp {
			op = 0x86
		}
		e.a.bytes(op, 0x02) // xchg [rdx], rax/eax/ax/al
	} else if group == 2 || group == 3 {
		e.slotLoad(in.B, 0, false, width == 8)
		if group == 3 {
			e.a.xop(width == 8, 0xf7, 0xd8) // neg eax/rax
		}
		e.atomicPrefix(width, true, false)
		op := byte(0xc1)
		if byteOp {
			op = 0xc0
		}
		e.a.bytes(0x0f, op, 0x02) // lock xadd [rdx], rax/eax/ax/al
	} else if group == 8 {
		e.slotLoad(in.B, 0, false, width == 8)
		e.slotLoad(in.C, 1, false, width == 8)
		e.atomicPrefix(width, true, false)
		op := byte(0xb1)
		if byteOp {
			op = 0xb0
		}
		e.a.bytes(0x0f, op, 0x0a) // lock cmpxchg [rdx], rcx/ecx/cx/cl
	} else {
		e.slotLoad(in.B, 1, false, width == 8)
		e.atomicLoadAMD64(width)
		loop := e.a.label()
		e.a.mark(loop)
		e.interrupt(pc)
		e.registerMove(8, 0, width == 8)
		rex := byte(0x41)
		if width == 8 {
			rex |= 8
		}
		op := [...]byte{0x21, 0x09, 0x31}[group-4]
		e.a.bytes(rex, op, 0xc8) // and/or/xor r8, rcx
		e.atomicPrefix(width, true, true)
		cas := byte(0xb1)
		if byteOp {
			cas = 0xb0
		}
		e.a.bytes(0x0f, cas, 0x02) // lock cmpxchg [rdx], r8[d/w/b]
		e.branch(loop, "ne")
	}
	if width == 1 {
		e.a.bytes(0x0f, 0xb6, 0xc0)
	} else if width == 2 {
		e.a.bytes(0x0f, 0xb7, 0xc0)
	} else if width == 4 {
		e.a.bytes(0x89, 0xc0)
	}
}
