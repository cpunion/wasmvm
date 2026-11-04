package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Private single-memory bulk operations use the same stable base as scalar
// accesses. Shared/multiple memories and precise fuel retain runtime helpers.
func (e *functionEmitter) bulkMemoryOpcode(in slots.Instruction) bool {
	if !e.nativeBulk || in.Op != 0xfc000a && in.Op != 0xfc000b {
		return false
	}
	source := e.f.Sources[in.Source].Wasm
	return source.Op == in.Op && source.A == 0 && (in.Op != 0xfc000a || source.B == 0)
}

func (e *functionEmitter) bulkMemory(in slots.Instruction, pc uint32) {
	e.slotLoad(in.A, 0, false, e.f.Slots[in.A] == wasm.I64)
	e.slotLoad(in.B, 1, false, e.f.Slots[in.B] == wasm.I64)
	e.slotLoad(in.C, 2, false, e.f.Slots[in.C] == wasm.I64)
	trap := e.trapExit(wasm.TrapMemoryBounds, e.f.Sources[in.Source].Wasm.Offset)
	done := e.a.label()
	if e.a.arch == "arm64" {
		// Check complete ranges, including u64 addition overflow, before writes.
		e.a.word(0xab0b012d) // adds x13, x9, x11
		e.a.branch(trap, true, 2)
		e.a.word(0xeb1101bf) // cmp x13, x17
		e.a.branch(trap, true, 8)
		if in.Op == 0xfc000a {
			e.a.word(0xab0b014d) // adds x13, x10, x11
			e.a.branch(trap, true, 2)
			e.a.word(0xeb1101bf)
			e.a.branch(trap, true, 8)
		}
		e.a.compareBranch(done, 11, true, false)
		e.a.word(0x8b090209) // add x9, x16, x9
		if in.Op == 0xfc000b {
			e.a.word(0x4e010d40) // dup v0.16b, w10
			e.bulkARM64(pc, false, false)
		} else {
			e.a.word(0x8b0a020a) // add x10, x16, x10
			forward := e.a.label()
			e.a.word(0xeb0a013f) // cmp x9, x10
			e.a.branch(done, true, 0)
			e.a.branch(forward, true, 3)
			e.a.word(0x8b0b0129) // add x9, x9, x11
			e.a.word(0x8b0b014a) // add x10, x10, x11
			e.bulkARM64(pc, true, true)
			e.a.branch(done, false, 0)
			e.a.mark(forward)
			e.bulkARM64(pc, true, false)
		}
	} else {
		e.bulkAMD64(in.Op == 0xfc000a, pc, trap, done)
	}
	e.a.mark(done)
}

// x9/x10 are cursors, x11 the total remainder, and x14/x15 chunk counters.
// Only scratch vectors v0/v1 are touched. No guest slot changes during a chunk.
func (e *functionEmitter) bulkARM64(pc uint32, copying, backwards bool) {
	outer, block, vector, tail, next, poll := e.a.label(), e.a.label(), e.a.label(), e.a.label(), e.a.label(), e.a.label()
	e.pollExits = append(e.pollExits, pollExit{poll, pc, e.currentRegisterPlan()})
	e.a.mark(outer)
	e.a.word(0xf940000d | uint32(nativeabi.ControlOffset/8)<<10) // ldr x13, [x0, control]
	e.a.compareBranch(poll, 13, true, true)
	e.a.word(0xd282000e) // mov x14, #4096
	e.a.word(0xeb0e017f) // cmp x11, x14
	e.a.word(0x9a8e316e) // csel x14, x11, x14, lo
	e.a.word(0xaa0e03ef) // mov x15, x14
	e.a.mark(block)
	e.a.word(0xf10101ff) // cmp x15, #64
	e.a.branch(vector, true, 3)
	for i := 0; i < 2; i++ {
		if copying {
			if backwards {
				e.a.word(0xadff0540) // ldp q0, q1, [x10, #-32]!
				e.a.word(0xadbf0520) // stp q0, q1, [x9, #-32]!
			} else {
				e.a.word(0xacc10540) // ldp q0, q1, [x10], #32
				e.a.word(0xac810520) // stp q0, q1, [x9], #32
			}
		} else {
			e.a.word(0xac810120) // stp q0, q0, [x9], #32
		}
	}
	e.a.word(0xd10101ef) // sub x15, x15, #64
	e.a.branch(block, false, 0)
	e.a.mark(vector)
	e.a.word(0xf10041ff) // cmp x15, #16
	e.a.branch(tail, true, 3)
	if copying {
		if backwards {
			e.a.word(0x3cdf0d40) // ldr q0, [x10, #-16]!
		} else {
			e.a.word(0x3cc10540) // ldr q0, [x10], #16
		}
	}
	if backwards {
		e.a.word(0x3c9f0d20) // str q0, [x9, #-16]!
	} else {
		e.a.word(0x3c810520) // str q0, [x9], #16
	}
	e.a.word(0xd10041ef) // sub x15, x15, #16
	e.a.branch(vector, false, 0)
	e.a.mark(tail)
	e.a.compareBranch(next, 15, true, false)
	if copying {
		if backwards {
			e.a.word(0x385ffd4d) // ldrb w13, [x10, #-1]!
			e.a.word(0x381ffd2d) // strb w13, [x9, #-1]!
		} else {
			e.a.word(0x3840154d) // ldrb w13, [x10], #1
			e.a.word(0x3800152d) // strb w13, [x9], #1
		}
	} else {
		e.a.word(0x3800152a) // strb w10, [x9], #1
	}
	e.a.word(0xf10005ef) // subs x15, x15, #1
	e.a.branch(tail, true, 1)
	e.a.mark(next)
	e.a.word(0xeb0e016b) // subs x11, x11, x14
	e.a.branch(outer, true, 1)
}

// REP strings use only the scratch register set and temporarily borrow the
// context/frame registers. Every normal/cancel exit restores them and clears DF.
func (e *functionEmitter) bulkAMD64(copying bool, pc uint32, trap, done int) {
	e.registerMove(8, 0, true)
	e.a.bytes(0x49, 0x01, 0xd0) // add r8, rdx
	e.a.branch(trap, true, 2)
	e.a.bytes(0x4d, 0x39, 0xd8) // cmp r8, r11
	e.a.branch(trap, true, 7)
	if copying {
		e.registerMove(8, 1, true)
		e.a.bytes(0x49, 0x01, 0xd0)
		e.a.branch(trap, true, 2)
		e.a.bytes(0x4d, 0x39, 0xd8)
		e.a.branch(trap, true, 7)
	}
	e.a.bytes(0x48, 0x85, 0xd2) // test rdx, rdx
	e.branch(done, "eq")
	e.a.bytes(0x4c, 0x01, 0xd0) // add rax, r10
	if copying {
		e.a.bytes(0x4c, 0x01, 0xd1) // add rcx, r10
	}
	e.registerMove(8, 7, true) // r8 = context
	e.registerMove(7, 0, true) // rdi = destination
	if copying {
		e.registerMove(0, 6, true) // rax = saved frame
		e.registerMove(6, 1, true) // rsi = source
	} else {
		e.registerMove(0, 1, true) // al = fill byte
	}
	e.a.bytes(0xfc) // cld
	loop, cleanup, poll, interrupted := e.a.label(), e.a.label(), e.a.label(), e.a.label()
	if copying {
		e.a.bytes(0x48, 0x39, 0xf7)             // cmp rdi, rsi
		e.a.branch(loop, true, 6)               // jbe forward
		e.a.bytes(0x48, 0x8d, 0x7c, 0x17, 0xff) // lea rdi, [rdi+rdx-1]
		e.a.bytes(0x48, 0x8d, 0x74, 0x16, 0xff) // lea rsi, [rsi+rdx-1]
		e.a.bytes(0xfd)                         // std
	}
	e.a.mark(loop)
	e.a.bytes(0x49, 0x83, 0xb8) // cmpq $0, control(r8)
	e.a.word(uint32(nativeabi.ControlOffset))
	e.a.bytes(0)
	e.branch(interrupted, "ne")
	e.a.bytes(0xb9)
	e.a.word(4096)                    // mov ecx, 4096
	e.a.bytes(0x48, 0x39, 0xca)       // cmp rdx, rcx
	e.a.bytes(0x48, 0x0f, 0x42, 0xca) // cmovb rcx, rdx
	e.a.bytes(0x48, 0x29, 0xca)       // sub rdx, rcx
	if copying {
		e.a.bytes(0xf3, 0xa4) // rep movsb
	} else {
		e.a.bytes(0xf3, 0xaa) // rep stosb
	}
	e.a.bytes(0x48, 0x85, 0xd2)
	e.branch(loop, "ne")
	e.a.branch(cleanup, false, 0)
	e.a.mark(interrupted)
	e.a.bytes(0xfc)
	if copying {
		e.registerMove(6, 0, true)
	}
	e.registerMove(7, 8, true)
	e.a.branch(poll, false, 0)
	e.pollExits = append(e.pollExits, pollExit{poll, pc, e.currentRegisterPlan()})
	e.a.mark(cleanup)
	e.a.bytes(0xfc)
	if copying {
		e.registerMove(6, 0, true)
	}
	e.registerMove(7, 8, true)
}
