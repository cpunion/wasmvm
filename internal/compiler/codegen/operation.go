package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// All straight-line instructions share one emitter on hot and cold paths.
func plainOpcode(op uint32) bool {
	return op == 1 || op >= 0x1a && op <= 0x1c || op >= 0x20 && op <= 0x22 || isMemoryAccess(op) || op == 0x3f || op >= 0x41 && op <= 0x44 || isInteger(op) || isFloat(op) || isConversion(op) || op == 0xd0 || op == 0xd1 || op == 0xfd000c || nativeVectorOpcode(op)
}
func (e *functionEmitter) plain(in slots.Instruction, proven bool) {
	a := e.a
	src := e.f.Sources[in.Source]
	switch op := in.Op; {
	case op == 1 || op == 0x1a:
	case op == 0x20:
		e.move(in.Dst, slots.Slot(src.Wasm.A))
	case op == 0x21 || op == 0x22:
		e.move(slots.Slot(src.Wasm.A), in.A)
		if in.Dst != slots.NoSlot {
			e.move(in.Dst, in.A)
		}
	case op == 0x1b || op == 0x1c:
		other, done := a.label(), a.label()
		e.branchZero(in.C, other)
		e.move(in.Dst, in.A)
		a.branch(done, false, 0)
		a.mark(other)
		e.move(in.Dst, in.B)
		a.mark(done)
	case op >= 0x41 && op <= 0x44 || op == 0xd0 || op == 0xfd000c:
		e.literal(in)
	case isInteger(op) || op == 0xd1:
		e.integer(in)
	case isFloat(op):
		e.floating(in)
	case isConversion(op):
		e.conversion(in)
	case nativeVectorOpcode(op):
		e.vectorInstruction(in)
	case isMemoryAccess(op) || op == 0x3f:
		e.memoryInstruction(in, proven)
	}
}
