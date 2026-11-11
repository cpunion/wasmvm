package codegen

import "github.com/cpunion/wasmvm/internal/compiler/nativeabi"

// Linked-call work registers may overwrite allocated caller registers only
// after they have been spilled. Context and current frame registers stay fixed.
func (e *functionEmitter) callReg(n byte) byte {
	if e.a.arch == "amd64" {
		return [...]byte{0, 1, 2, 8, 9, 10, 11, 3}[n]
	}
	return [...]byte{9, 10, 11, 14, 15, 16, 17, 13}[n]
}
func (e *functionEmitter) callContext() byte {
	if e.a.arch == "amd64" {
		return 7
	}
	return 0
}
func (e *functionEmitter) callFrame() byte {
	if e.a.arch == "amd64" {
		return 6
	}
	return 12
}
func (e *functionEmitter) callConst(n byte, v uint64) {
	if e.a.arch == "amd64" {
		e.a.xconst(e.callReg(n), v)
	} else {
		e.a.aconst(uint32(e.callReg(n)), v, true)
	}
}
func (e *functionEmitter) callMove(dst, src byte) {
	e.registerMove(e.callReg(dst), e.callReg(src), true)
}
func (e *functionEmitter) callMemory(n, base byte, offset uint32, store, narrow bool) {
	r := e.callReg(n)
	if e.a.arch == "amd64" {
		rex := byte(0x40) | (r>>3)<<2 | base>>3
		if !narrow {
			rex |= 8
		}
		if rex != 0x40 {
			e.a.bytes(rex)
		}
		op := byte(0x8b)
		if store {
			op = 0x89
		}
		e.a.bytes(op, 0x80|(r&7)<<3|base&7)
		e.a.word(offset)
		return
	}
	shift, op := uint32(3), uint32(0xf9400000)
	if narrow {
		shift, op = 2, 0xb9400000
	}
	if store {
		op &^= 1 << 22
	}
	if offset>>shift >= 4096 {
		e.a.aconst(uint32(e.callReg(7)), uint64(offset), true)
		e.a.word(0x8b000000 | uint32(e.callReg(7))<<16 | uint32(base)<<5 | uint32(e.callReg(7)))
		base, offset = e.callReg(7), 0
	}
	e.a.word(op | offset>>shift<<10 | uint32(base)<<5 | uint32(r))
}
func (e *functionEmitter) callLoad(n, base byte, off uint32) {
	e.callMemory(n, base, off, false, false)
}
func (e *functionEmitter) callStore(n, base byte, off uint32) {
	e.callMemory(n, base, off, true, false)
}
func (e *functionEmitter) callAdd(dst, left, right byte) {
	d, l, r := e.callReg(dst), e.callReg(left), e.callReg(right)
	if e.a.arch == "amd64" {
		if d == r {
			l, r = r, l
		}
		e.registerMove(d, l, true)
		e.a.bytes(0x48|(r>>3)<<2|d>>3, 0x01, 0xc0|(r&7)<<3|d&7)
	} else {
		e.a.word(0x8b000000 | uint32(r)<<16 | uint32(l)<<5 | uint32(d))
	}
}
func (e *functionEmitter) callAdvance(n byte, delta int) {
	r := e.callReg(n)
	if e.a.arch == "amd64" {
		group, imm := byte(0), delta
		if delta < 0 {
			group, imm = 5, -delta
		}
		e.a.bytes(0x48|r>>3, 0x81, 0xc0|group<<3|r&7)
		e.a.word(uint32(imm))
	} else {
		op, imm := uint32(0x91000000), delta
		if delta < 0 {
			op, imm = 0xd1000000, -delta
		}
		e.a.word(op | uint32(imm)<<10 | uint32(r)<<5 | uint32(r))
	}
}

// Subtract one work register from another after caller registers are spilled.
func (e *functionEmitter) callSubtract(left, right byte) {
	l, r := e.callReg(left), e.callReg(right)
	if e.a.arch == "amd64" {
		e.a.bytes(0x48|(r>>3)<<2|l>>3, 0x29, 0xc0|(r&7)<<3|l&7)
	} else {
		e.a.word(0xcb000000 | uint32(r)<<16 | uint32(l)<<5 | uint32(l))
	}
}

// The row supplies the callee's slot count in register 5. Poll every slot so
// even a large dynamically selected frame has bounded cancellation latency.
func (e *functionEmitter) callZeroDynamic(cancelled int) {
	done, loop := e.a.label(), e.a.label()
	e.callTest(5)
	e.branch(done, "eq")
	e.callMove(6, 4)
	e.callConst(1, 0)
	e.a.mark(loop)
	e.callStore(1, e.callReg(6), 0)
	e.callStore(1, e.callReg(6), 8)
	e.callCancelled(cancelled)
	e.callAdvance(6, 16)
	e.callAdvance(5, -1)
	e.callTest(5)
	e.branch(loop, "ne")
	e.a.mark(done)
}
func (e *functionEmitter) callShift(n, bits byte) {
	r := e.callReg(n)
	if e.a.arch == "amd64" {
		e.a.bytes(0x48|r>>3, 0xc1, 0xe0|r&7, bits)
	} else {
		e.a.word(0xd3400000 | uint32(64-bits)<<16 | uint32(63-bits)<<10 | uint32(r)<<5 | uint32(r))
	}
}
func (e *functionEmitter) callCompare(left, right byte) {
	l, r := e.callReg(left), e.callReg(right)
	if e.a.arch == "amd64" {
		e.a.bytes(0x48|(r>>3)<<2|l>>3, 0x39, 0xc0|(r&7)<<3|l&7)
	} else {
		e.a.word(0xeb00001f | uint32(r)<<16 | uint32(l)<<5)
	}
}
func (e *functionEmitter) callTest(n byte) {
	r := e.callReg(n)
	if e.a.arch == "amd64" {
		e.a.bytes(0x48|(r>>3)<<2|r>>3, 0x85, 0xc0|(r&7)<<3|r&7)
	} else {
		e.a.word(0xf100001f | uint32(r)<<5)
	}
}
func (e *functionEmitter) callAddress(label int, dst byte) {
	if e.a.arch == "amd64" {
		r := e.callReg(dst)
		e.a.bytes(0x48|(r>>3)<<2, 0x8d, 0x05|(r&7)<<3)
		e.a.fixups = append(e.a.fixups, relocation{position: len(e.a.code), label: label})
		e.a.word(0)
	} else {
		e.a.addresses = append(e.a.addresses, addressRelocation{len(e.a.code), label})
		e.a.word(0)
		e.registerMove(e.callReg(dst), 13, true)
	}
}
func (e *functionEmitter) callJump(n byte) {
	r := e.callReg(n)
	if e.a.arch == "amd64" {
		if r >= 8 {
			e.a.bytes(0x41)
		}
		e.a.bytes(0xff, 0xe0|r&7)
	} else {
		e.a.word(0xd61f0000 | uint32(r)<<5)
	}
}
func (e *functionEmitter) callMeta(parent bool) {
	e.callLoad(2, e.callContext(), uint32(nativeabi.CallViewOffset))
	e.callLoad(0, e.callReg(2), nativeabi.CallTopOffset)
	if parent {
		e.callAdvance(0, -1)
	}
	e.callShift(0, 6)
	e.callAdd(3, 2, 0)
	e.callAdvance(3, nativeabi.CallHeaderBytes)
}
func (e *functionEmitter) callCancelled(label int) {
	e.callLoad(0, e.callContext(), uint32(nativeabi.ControlOffset))
	e.callTest(0)
	e.branch(label, "ne")
}
func (e *functionEmitter) callRawLeave(status nativeabi.Exit, payload, pc uint64) {
	e.contextConst(nativeabi.ContinuationOffset, pc)
	e.contextConst(nativeabi.PayloadOffset, payload)
	e.rawStatus(status)
}

// Zero at most 4KiB between cancellation checks. The caller has spilled all
// registers; a canceled partial child is not yet part of the live frame chain.
func (e *functionEmitter) callZero(slots uint32, cancelled int) {
	e.callMove(6, 4)
	chunks, tail := slots/256, slots%256
	if e.a.arch == "amd64" {
		e.registerMove(e.callReg(7), e.callContext(), true)
		chunk := func(n uint32) {
			e.registerMove(7, e.callReg(6), true)
			e.a.xconst(1, uint64(n)*2)
			e.a.xconst(0, 0)
			e.a.bytes(0xf3, 0x48, 0xab) // rep stosq
			e.registerMove(e.callReg(6), 7, true)
			e.registerMove(7, e.callReg(7), true)
		}
		if chunks != 0 {
			e.callConst(5, uint64(chunks))
			loop := e.a.label()
			e.a.mark(loop)
			chunk(256)
			e.callCancelled(cancelled)
			e.callAdvance(5, -1)
			e.callTest(5)
			e.branch(loop, "ne")
		}
		if tail != 0 {
			chunk(tail)
		}
		return
	}
	chunk := func(n uint32) {
		e.a.word(0x4f00e400) // movi v0.16b, #0; all guest vectors were spilled.
		if blocks := n / 8; blocks != 0 {
			e.callConst(5, uint64(blocks))
			loop := e.a.label()
			e.a.mark(loop)
			for pair := uint32(0); pair < 4; pair++ {
				e.a.word(0xad000000 | pair*2<<15 | uint32(e.callReg(6))<<5) // stp q0,q0,[cursor,#pair*32]
			}
			e.callAdvance(6, 128)
			e.callAdvance(5, -1)
			e.callTest(5)
			e.branch(loop, "ne")
		}
		remaining := n % 8
		for pair := uint32(0); pair < remaining/2; pair++ {
			e.a.word(0xad000000 | pair*2<<15 | uint32(e.callReg(6))<<5)
		}
		if remaining%2 != 0 {
			e.a.word(0x3d800000 | (remaining-1)<<10 | uint32(e.callReg(6))<<5) // str q0,[cursor,#slot*16]
		}
		if remaining != 0 {
			e.callAdvance(6, int(remaining)*16)
		}
	}
	if chunks != 0 {
		e.callConst(7, uint64(chunks))
		loop := e.a.label()
		e.a.mark(loop)
		chunk(256)
		e.callCancelled(cancelled)
		e.callAdvance(7, -1)
		e.callTest(7)
		e.branch(loop, "ne")
	}
	if tail != 0 {
		chunk(tail)
	}
}
