package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Tuple inputs are consumed before either result is published. Scratch 2 is
// reserved by both allocators, so carry chains and multiply-high cannot clobber
// live homes. Slot loads preserve flags, including large ARM frame offsets.
func (e *functionEmitter) wideArithmetic(in slots.Instruction) {
	args := e.f.Args[in.Extra : in.Extra+in.Count+2]
	a := e.a
	lo, hi := byte(2), byte(0)
	if in.Op <= 0xfc0014 {
		e.slotLoad(args[0], 0, false, true)
		e.slotLoad(args[2], 1, false, true)
		if a.arch == "amd64" {
			if in.Op == 0xfc0013 {
				a.bytes(0x48, 0x01, 0xc8)
			} else {
				a.bytes(0x48, 0x29, 0xc8)
			}
		} else {
			if in.Op == 0xfc0013 {
				a.word(0xab0a0129)
			} else {
				a.word(0xeb0a0129)
			}
		}
		e.registerMove(e.operandRegister(2), e.operandRegister(0), true)
		e.slotLoad(args[1], 0, false, true)
		e.slotLoad(args[3], 1, false, true)
		if a.arch == "amd64" {
			if in.Op == 0xfc0013 {
				a.bytes(0x48, 0x11, 0xc8)
			} else {
				a.bytes(0x48, 0x19, 0xc8)
			}
		} else {
			if in.Op == 0xfc0013 {
				a.word(0x9a0a0129)
			} else {
				a.word(0xda0a0129)
			}
		}
	} else {
		e.slotLoad(args[0], 0, false, true)
		e.slotLoad(args[1], 1, false, true)
		if a.arch == "amd64" {
			lo, hi = 0, 2
			if in.Op == 0xfc0015 {
				a.bytes(0x48, 0xf7, 0xe9)
			} else {
				a.bytes(0x48, 0xf7, 0xe1)
			}
		} else {
			a.word(0x9b0a7d2b) // mul x11, x9, x10
			if in.Op == 0xfc0015 {
				a.word(0x9b4a7d29)
			} else {
				a.word(0x9bca7d29)
			}
		}
	}
	e.slotStore(args[in.Count], lo, false)
	e.slotStore(args[in.Count+1], hi, false)
	if !e.unmetered {
		e.constant(1, 0)
		e.slotStore(args[in.Count], 1, true)
		e.slotStore(args[in.Count+1], 1, true)
	}
}
