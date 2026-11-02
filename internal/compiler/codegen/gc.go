package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// A newly constructed i31 is non-null. Forward its payload into an adjacent
// extraction while retaining the boxed slot if a later exit can observe it.
// The interior PC still gets the ordinary, independently checked entry.
func (e *functionEmitter) combineI31(pc int) (int, bool) {
	if !e.unmetered || pc+1 >= e.codeLimit() || pc/256 != (pc+1)/256 {
		return 0, false
	}
	box, get := e.f.Code[pc], e.f.Code[pc+1]
	if box.Op != 0xfb001c || (get.Op != 0xfb001d && get.Op != 0xfb001e) || get.A != box.Dst {
		return 0, false
	}
	if box.Dst != get.Dst && !e.overwrittenBeforeExit(pc+2, box.Dst) {
		e.gcScalar(box)
	}
	e.slotLoad(box.A, 0, false, false)
	if e.a.arch == "arm64" {
		reg := uint32(e.operandRegister(0))
		if get.Op == 0xfb001d {
			e.a.arm(false, 0x13007800|reg<<5|reg) // sbfx wN, wN, #0, #31
		} else {
			e.a.arm(false, 0x12007800|reg<<5|reg) // and wN, wN, #0x7fffffff
		}
	} else if get.Op == 0xfb001d {
		e.a.bytes(0xd1, 0xe0, 0xd1, 0xf8) // shl eax, 1; sar eax, 1
	} else {
		e.a.bytes(0x25)
		e.a.word(0x7fffffff) // and eax, 0x7fffffff
	}
	e.slotStore(get.Dst, 0, false)
	return pc + 1, true
}

// Immediate i31 values and reference equality need no heap or helper handoff.
func (e *functionEmitter) gcScalar(in slots.Instruction) {
	core := func(op uint32, wide bool) {
		if e.a.arch == "amd64" {
			integerAMD64Core(e.a, op, wide, e.zero, e.overflow)
		} else {
			integerARM64Core(e.a, op, wide, e.zero, e.overflow)
		}
	}
	e.slotLoad(in.A, 0, false, true)
	if in.Op == 0xd3 {
		e.slotLoad(in.B, 1, false, true)
		core(0x51, true)
	} else if in.Op == 0xfb001c {
		e.constant(1, 0x7fffffff)
		core(0x83, true)
		e.constant(1, 1<<62)
		core(0x84, true)
	} else {
		e.test(true)
		e.branch(e.trapExit(wasm.TrapNullReference, e.f.Sources[in.Source].Wasm.Offset), "eq")
		e.constant(1, 0x7fffffff)
		core(0x83, true)
		if in.Op == 0xfb001d {
			e.constant(1, 1)
			core(0x74, false)
			core(0x75, false)
		}
	}
	e.slotStore(in.Dst, 0, false)
	if !e.unmetered {
		e.constant(0, 0)
		e.slotStore(in.Dst, 0, true)
	}
}
