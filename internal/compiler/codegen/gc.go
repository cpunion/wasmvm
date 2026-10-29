package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

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
