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
	for ; pc < len(e.f.Code) && pc-start <= 8 && pc/256 == start/256; pc++ {
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
	if pc >= len(e.f.Code) || pc-start > 8 || pc/256 != start/256 {
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
	if pc+1 < len(e.f.Code) && (pc+1)/256 == start/256 {
		next := e.f.Code[pc+1]
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == original.Dst {
			candidate := in
			candidate.Dst = slots.Slot(e.f.Sources[next.Source].Wasm.A)
			if e.canRegisterScalar(candidate) {
				in, end = candidate, pc+1
			}
		}
	}
	if !e.canRegisterScalar(in) {
		return 0, false
	}
	for _, copy := range e.f.Code[start:pc] {
		if copy.Op == 0x20 {
			e.move(copy.Dst, slots.Slot(e.f.Sources[copy.Source].Wasm.A))
		} else {
			e.literal(copy)
		}
	}
	e.registerScalar(in)
	if end != pc {
		e.move(original.Dst, in.Dst)
		e.move(e.f.Code[end].Dst, in.Dst)
	}
	return end, true
}

func (e *functionEmitter) canRegisterScalar(in slots.Instruction) bool {
	if isInteger(in.Op) {
		return e.canRegisterInteger(in.Op, in.Dst, in.A, in.B)
	}
	if isFloat(in.Op) {
		return e.canRegisterFloat(in) || e.canReinterpretRegister(in)
	}
	return e.canRegisterConversion(in)
}

func (e *functionEmitter) registerScalar(in slots.Instruction) {
	if isInteger(in.Op) {
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
	if !e.unmetered || pc+1 >= len(e.f.Code) || (pc+1)%256 == 0 {
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
	if !e.canRegisterScalar(forwarded) {
		return false
	}
	e.registerScalar(forwarded)
	e.move(in.Dst, dst)
	e.move(next.Dst, dst)
	return true
}
