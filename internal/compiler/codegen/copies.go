package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Forward nontrapping copies into one consuming scalar operation. Retain values in
// their original slots, but remove copies from the arithmetic dependency chain.
// Every interior instruction keeps an independent cold continuation.
func (e *functionEmitter) combineScalarCopies(start int) (int, bool) {
	if !e.unmetered || e.f.Code[start].Op != 0x20 {
		return 0, false
	}
	aliases := map[slots.Slot]slots.Slot{}
	pc := start
	for ; pc < e.codeLimit() && pc-start <= 8 && pc/256 == start/256; pc++ {
		in := e.f.Code[pc]
		if in.Op == 0x20 {
			local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
			for dst, source := range aliases {
				if source == in.Dst {
					delete(aliases, dst)
				}
			}
			aliases[in.Dst] = local
		} else if in.Op >= 0x41 && in.Op <= 0x44 {
			delete(aliases, in.Dst)
			for dst, source := range aliases {
				if source == in.Dst {
					delete(aliases, dst)
				}
			}
		} else {
			break
		}
	}
	if pc >= e.codeLimit() || pc-start > 8 || pc/256 != start/256 {
		return 0, false
	}
	if e.a.arch == "arm64" {
		if _, _, _, ok := e.multiplyAddTail(pc); ok {
			return 0, false
		}
	}
	original, in := e.f.Code[pc], e.f.Code[pc]
	if source, ok := aliases[in.A]; ok {
		in.A = source
	}
	if source, ok := aliases[in.B]; ok {
		in.B = source
	}
	if in.A == original.A && in.B == original.B {
		return 0, false
	}
	end := pc
	if pc+1 < e.codeLimit() && (pc+1)/256 == start/256 {
		next := e.f.Code[pc+1]
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == original.Dst {
			candidate := in
			candidate.Dst = slots.Slot(e.f.Sources[next.Source].Wasm.A)
			if e.canRegisterScalar(candidate, pc) {
				in, end = candidate, pc+1
			}
		}
	}
	if !e.canRegisterScalar(in, pc) {
		return 0, false
	}
	previous := e.f.Code[pc-1]
	deferConstant := (previous.Op == 0x41 || previous.Op == 0x42) && previous.Dst == in.B && in.A != previous.Dst
	for offset, copy := range e.f.Code[start:pc] {
		if deferConstant && start+offset == pc-1 {
			continue
		}
		if copy.Op == 0x20 {
			if !scalarCanTrap(in.Op) && copy.Dst != in.A && copy.Dst != in.B && copy.Dst != in.C &&
				(copy.Dst == in.Dst || e.overwrittenBeforeExit(end+1, copy.Dst)) {
				continue
			}
			e.move(copy.Dst, slots.Slot(e.f.Sources[copy.Source].Wasm.A))
		} else {
			e.literal(copy)
		}
	}
	if deferConstant && e.registerImmediate(in.Op, in.Dst, in.A, e.f.Sources[previous.Source].Wasm.Bits) {
		if previous.Dst != in.Dst && previous.Dst != original.Dst && (end == pc || e.f.Code[end].Dst != previous.Dst) && !e.overwrittenBeforeExit(end+1, previous.Dst) {
			e.literal(previous)
		}
	} else {
		if deferConstant {
			e.literal(previous)
		}
		e.registerScalar(in)
	}
	if end != pc {
		if !e.overwrittenBeforeExit(end+1, original.Dst) {
			e.move(original.Dst, in.Dst)
		}
		if !e.overwrittenBeforeExit(end+1, e.f.Code[end].Dst) {
			e.move(e.f.Code[end].Dst, in.Dst)
		}
	}
	return end, true
}

func (e *functionEmitter) canRegisterScalar(in slots.Instruction, pc int) bool {
	if in.Op >= 0x28 && in.Op <= 0x3e {
		return e.provenPCs[pc]
	}
	if isInteger(in.Op) {
		return e.canRegisterInteger(in.Op, in.Dst, in.A, in.B)
	}
	if isFloat(in.Op) {
		return e.canRegisterFloat(in) || e.canReinterpretRegister(in)
	}
	return e.canRegisterConversion(in)
}

func (e *functionEmitter) registerScalar(in slots.Instruction) {
	if in.Op >= 0x28 && in.Op <= 0x3e {
		e.memoryInstruction(in, true)
	} else if isInteger(in.Op) {
		e.registerInteger(in.Op, in.Dst, in.A, in.B)
	} else if isFloat(in.Op) {
		e.floating(in)
	} else {
		e.conversion(in)
	}
}

// Redirect a result into its following local destination, retaining its temporary
// with an off-chain copy. The original local instruction keeps a cold entry.
func (e *functionEmitter) forwardScalarResult(pc int) bool {
	if !e.unmetered || pc+1 >= e.codeLimit() || (pc+1)%256 == 0 {
		return false
	}
	in, next := e.f.Code[pc], e.f.Code[pc+1]
	if (next.Op != 0x21 && next.Op != 0x22) || next.A != in.Dst {
		return false
	}
	dst := slots.Slot(e.f.Sources[next.Source].Wasm.A)
	if dst == in.Dst {
		return false
	}
	forwarded := in
	forwarded.Dst = dst
	if !e.canRegisterScalar(forwarded, pc) {
		return false
	}
	e.registerScalar(forwarded)
	if !e.overwrittenBeforeExit(pc+2, in.Dst) {
		e.move(in.Dst, dst)
	}
	if !e.overwrittenBeforeExit(pc+2, next.Dst) {
		e.move(next.Dst, dst)
	}
	return true
}

// These operations can exit before producing their result. Copies needed by
// a trap frame must remain materialized ahead of them.
func scalarCanTrap(op uint32) bool {
	return op >= 0x6d && op <= 0x70 || op >= 0x7f && op <= 0x82 || op >= 0xa8 && op <= 0xab || op >= 0xae && op <= 0xb1
}
