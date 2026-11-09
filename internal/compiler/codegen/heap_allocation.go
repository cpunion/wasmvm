package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
)

// Reuse the bounded register-frequency hints for profitability only. Unknown
// flow keeps the previous small batch; generated quota/payload guards apply to
// every reservation size. Large straight-line sequences also amortize entry.
func (e *functionEmitter) prepareAllocationBatch(frequencies []uint64) {
	if !e.hasAllocations {
		return
	}
	e.allocationBatch = nativeabi.HeapAllocationSmallBatch
	var constructors uint32
	for pc, in := range e.f.Code {
		if _, ok := e.heapAllocationType(in); !ok {
			continue
		}
		constructors++
		if frequency(frequencies, pc) > 1 || constructors > nativeabi.HeapAllocationSmallBatch {
			e.allocationBatch = nativeabi.HeapAllocationBatch
			return
		}
	}
}

func (e *functionEmitter) heapAllocationType(in slots.Instruction) (module.TypeDef, bool) {
	if in.Op != 0xfb0000 && in.Op != 0xfb0001 {
		return module.TypeDef{}, false
	}
	source := e.f.Sources[in.Source].Wasm
	if source.Op != in.Op || uint64(source.A) >= uint64(len(e.allocationTypes)) {
		return module.TypeDef{}, false
	}
	t := e.allocationTypes[source.A]
	valid := t.Kind == module.StructKind && t.Descriptor == 0 && t.Describes == 0 && len(t.Fields) <= nativeabi.MaxNativeStructFields
	return t, valid && (in.Op == 0xfb0001 && in.Count == 0 || in.Op == 0xfb0000 && int(in.Count) == len(t.Fields))
}

func (e *functionEmitter) heapAllocation(in slots.Instruction, pc uint32) bool {
	t, ok := e.heapAllocationType(in)
	if !ok {
		return false
	}
	var payload uint32
	for _, f := range t.Fields {
		payload += uint32(f.HeapBytes())
	}
	slow, done := e.a.label(), e.a.label()
	header := e.runtimeBase(e.heapPair + 16)
	if e.a.arch == "arm64" {
		e.registerMove(11, header, true)
		e.allocateHeapARM64(payload, slow)
	} else {
		e.registerMove(1, header, true)
		e.allocateHeapAMD64(payload, slow)
	}
	e.constant(0, uint64(e.f.Sources[in.Source].Wasm.A)<<32|uint64(len(t.Fields)))
	e.allocationDescriptor(16)
	e.constant(0, 0)
	e.allocationDescriptor(24)
	// All guards precede publication. No exit or collection can occur while
	// fields are initialized; output slots may alias still-unread operands.
	offset := uint32(0)
	for i, field := range t.Fields {
		width := uint32(field.HeapBytes())
		vector := byte(0)
		if in.Op == 0xfb0001 {
			if width == 16 {
				if e.a.arch == "arm64" {
					e.a.word(0x4f00e400) // movi v0.16b, #0
				} else {
					e.a.xvector(0xef, 0, 0) // pxor xmm0, xmm0
				}
			} else {
				e.constant(0, 0)
			}
		} else {
			slot := e.f.Args[in.Extra+uint32(i)]
			if width == 16 {
				vector = e.vectorOperand(slot, 0)
			} else {
				e.slotLoad(slot, 0, false, true)
			}
		}
		e.allocationField(width, offset, vector)
		offset += width
	}
	if e.a.arch == "arm64" {
		e.a.word(0xf94001c9) // ldr x9, [x14] (complete handle)
	} else {
		e.a.bytes(0x48, 0x8b, 0x02) // mov rax, [rdx]
	}
	e.slotStore(in.Dst, 0, false)
	if e.a.arch == "arm64" {
		e.a.word(0xf9400169) // ldr x9, [x11]
		e.a.word(0x91000529) // add x9, x9, #1
		e.a.word(0xf9000169) // str x9, [x11]; commit last
	} else {
		e.a.bytes(0x48, 0xff, 0x01) // incq [rcx]
	}
	e.a.branch(done, false, 0)
	e.a.mark(slow)
	e.contextConst(nativeabi.SourceOffset, e.f.Sources[in.Source].Wasm.Offset)
	e.leave(nativeabi.RuntimeHelper, uint64(pc), uint64(pc+1))
	e.a.mark(done)
	return true
}

func (e *functionEmitter) allocationDescriptor(offset uint32) {
	if e.a.arch == "arm64" {
		e.a.word(0xf90001c9 | offset/8<<10)
	} else {
		e.a.bytes(0x48, 0x89, 0x42, byte(offset))
	}
}

// x11 is the header, x13 payload and x14 descriptor. All are outside the guest
// register inventory. x15 may be reused by large-frame operand addressing.
func (e *functionEmitter) allocateHeapARM64(payload uint32, slow int) {
	e.a.compareBranch(slow, 11, true, false)
	load := func(dst byte, offset int) { e.a.word(0xf9400160 | uint32(offset/8)<<10 | uint32(dst)) }
	store := func(src byte, offset int) { e.a.word(0xf9000160 | uint32(offset/8)<<10 | uint32(src)) }
	load(9, nativeabi.HeapAllocationCursor)
	load(10, nativeabi.HeapAllocationCapacity)
	e.compare(true)
	e.branch(slow, "hs")
	load(13, nativeabi.HeapAllocationUsed)
	e.a.word(0xb10001ae | payload<<10) // adds x14, x13, #payload
	e.a.branch(slow, true, 2)
	load(10, nativeabi.HeapAllocationLimit)
	e.a.word(0xeb0a01df) // cmp x14, x10
	e.a.branch(slow, true, 8)
	load(9, nativeabi.HeapAllocationBudget)
	e.a.word(0xf100013f | (payload+64)<<10)
	e.a.branch(slow, true, 3)
	store(14, nativeabi.HeapAllocationUsed)
	e.a.word(0xd1000129 | (payload+64)<<10)
	store(9, nativeabi.HeapAllocationBudget)
	load(9, nativeabi.HeapAllocationCursor)
	load(10, nativeabi.HeapAllocationHandles)
	e.a.word(0xf869794f) // ldr x15, [x10, x9, lsl #3]
	load(10, nativeabi.HeapAllocationDescriptors)
	e.registerMove(14, 15, false)
	e.a.word(0x8b0e154e) // add x14, x10, x14, lsl #5
	e.a.word(0xf90001cf) // str x15, [x14]
	if payload == 0 {
		e.a.aconst(13, 0, false)
	} else {
		load(9, nativeabi.HeapAllocationBase)
		e.a.word(0x8b0d012d) // add x13, x9, x13
	}
	e.a.word(0xf90005cd) // str x13, [x14, #8]
}

// rcx is the header, r8 payload and rdx descriptor. Loading an operand into
// rax (including FP bits) never clobbers these three scratch registers.
func (e *functionEmitter) allocateHeapAMD64(payload uint32, slow int) {
	e.a.bytes(0x48, 0x85, 0xc9)
	e.branch(slow, "eq")
	e.a.bytes(0x48, 0x8b, 0x01)
	e.a.bytes(0x48, 0x3b, 0x41, nativeabi.HeapAllocationCapacity)
	e.branch(slow, "hs")
	e.a.bytes(0x4c, 0x8b, 0x41, nativeabi.HeapAllocationUsed)
	e.a.bytes(0x49, 0x81, 0xc0)
	e.a.word(payload)
	e.a.branch(slow, true, 2)
	e.a.bytes(0x4c, 0x3b, 0x41, nativeabi.HeapAllocationLimit)
	e.a.branch(slow, true, 7)
	e.a.bytes(0x48, 0x81, 0x79, nativeabi.HeapAllocationBudget)
	e.a.word(payload + 64)
	e.a.branch(slow, true, 2)
	e.a.bytes(0x4c, 0x89, 0x41, nativeabi.HeapAllocationUsed)
	e.a.bytes(0x48, 0x81, 0x69, nativeabi.HeapAllocationBudget)
	e.a.word(payload + 64)
	e.a.bytes(0x48, 0x8b, 0x01)
	e.a.bytes(0x48, 0x8b, 0x51, nativeabi.HeapAllocationHandles)
	e.a.bytes(0x48, 0x8b, 0x04, 0xc2) // mov rax, [rdx+rax*8]
	e.a.bytes(0x48, 0x8b, 0x51, nativeabi.HeapAllocationDescriptors)
	e.registerMove(8, 0, false)
	e.a.bytes(0x49, 0xc1, 0xe0, 5, 0x4c, 0x01, 0xc2)
	e.a.bytes(0x48, 0x89, 0x02)
	if payload == 0 {
		e.a.bytes(0x45, 0x31, 0xc0)
	} else {
		e.a.bytes(0x4c, 0x8b, 0x41, nativeabi.HeapAllocationUsed)
		e.a.bytes(0x49, 0x81, 0xe8)
		e.a.word(payload)
		e.a.bytes(0x4c, 0x03, 0x41, nativeabi.HeapAllocationBase)
	}
	e.a.bytes(0x4c, 0x89, 0x42, 8)
}

func (e *functionEmitter) allocationField(width, offset uint32, vector byte) {
	if e.a.arch == "arm64" {
		base := byte(13)
		if offset != 0 {
			e.a.word(0x910001af | offset<<10) // add x15, x13, #offset
			base = 15
		}
		if width == 16 {
			e.a.word(0x3d800000 | uint32(base)<<5 | uint32(vector))
		} else {
			e.heapStore(width, base, 0, e.operandRegister(0))
		}
		return
	}
	if width == 16 {
		e.a.bytes(0xf3, 0x41|(vector>>3)<<2, 0x0f, 0x7f, 0x80|(vector&7)<<3)
	} else {
		if width == 2 {
			e.a.bytes(0x66)
		}
		rex := byte(0x41)
		if width == 8 {
			rex |= 8
		}
		op := byte(0x89)
		if width == 1 {
			op = 0x88
		}
		e.a.bytes(rex, op, 0x80)
	}
	e.a.word(offset)
}
