package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Combine local copies/constants with their consuming FP instruction. Every
// interior PC retains its original cold entry, so entering after a copy cannot
// incorrectly inherit the fallthrough operand-equivalence proof.
func (e *functionEmitter) combineFloatCopies(start int) (int, bool) {
	if !e.unmetered || e.f.Code[start].Op != 0x20 {
		return 0, false
	}
	pc := start + 1
	for ; pc < len(e.f.Code) && pc-start <= 8 && pc/256 == start/256; pc++ {
		op := e.f.Code[pc].Op
		if isFloat(op) {
			break
		}
		if !(op == 0x20 || op >= 0x41 && op <= 0x44) {
			return 0, false
		}
	}
	if pc >= len(e.f.Code) || pc-start > 8 || pc/256 != start/256 || !isFloat(e.f.Code[pc].Op) {
		return 0, false
	}
	original, in := e.f.Code[pc], e.f.Code[pc]
	in.A, in.B = e.floatOperand(start, pc, in.A), e.floatOperand(start, pc, in.B)
	if in.A == original.A && in.B == original.B {
		return 0, false
	}
	end := pc
	if pc+1 < len(e.f.Code) && (pc+1)/256 == start/256 {
		next := e.f.Code[pc+1]
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == original.Dst {
			candidate := in
			candidate.Dst = slots.Slot(e.f.Sources[next.Source].Wasm.A)
			if e.canRegisterFloat(candidate) {
				in, end = candidate, pc+1
			}
		}
	}
	if !e.canRegisterFloat(in) {
		return 0, false
	}
	for _, copy := range e.f.Code[start:pc] {
		if copy.Op == 0x20 {
			e.move(copy.Dst, slots.Slot(e.f.Sources[copy.Source].Wasm.A))
		} else {
			e.literal(copy)
		}
	}
	e.registerFloat(in)
	if end != pc {
		e.move(original.Dst, in.Dst)
		e.move(e.f.Code[end].Dst, in.Dst)
	}
	return end, true
}

func (e *functionEmitter) writesSlot(in slots.Instruction, slot slots.Slot) bool {
	return in.Dst == slot || (in.Op == 0x21 || in.Op == 0x22) && slots.Slot(e.f.Sources[in.Source].Wasm.A) == slot
}
func (e *functionEmitter) floatOperand(start, pc int, slot slots.Slot) slots.Slot {
	if slot == slots.NoSlot {
		return slot
	}
	for i := pc - 1; i >= start; i-- {
		in := e.f.Code[i]
		// Restrict the proof to nontrapping, nonpolling scalar operations.
		if !(in.Op >= 0x20 && in.Op <= 0x22 || in.Op >= 0x41 && in.Op <= 0x44 || isFloat(in.Op)) {
			break
		}
		if !e.writesSlot(in, slot) {
			continue
		}
		source := slots.NoSlot
		if in.Op == 0x20 {
			source = slots.Slot(e.f.Sources[in.Source].Wasm.A)
		} else if in.Op == 0x21 || in.Op == 0x22 {
			source = in.A
		}
		if source == slots.NoSlot || e.f.Slots[source] != e.f.Slots[slot] {
			break
		}
		for j := i + 1; j < pc; j++ {
			if e.writesSlot(e.f.Code[j], source) {
				return slot
			}
		}
		return source
	}
	return slot
}

// Write a result into the following local.set's destination, then retain its
// temporary slot with an off-chain copy. No fuel or poll boundary is removed;
// the original local instruction still has an independent cold continuation.
func (e *functionEmitter) forwardFloatResult(pc int) bool {
	if !e.unmetered || pc+1 >= len(e.f.Code) || (pc+1)%256 == 0 {
		return false
	}
	in, next := e.f.Code[pc], e.f.Code[pc+1]
	if !isFloat(in.Op) || (next.Op != 0x21 && next.Op != 0x22) || next.A != in.Dst {
		return false
	}
	dst := slots.Slot(e.f.Sources[next.Source].Wasm.A)
	if dst == in.Dst {
		return false
	}
	forwarded := in
	forwarded.Dst = dst
	if !e.registerFloat(forwarded) {
		return false
	}
	e.move(in.Dst, dst)
	e.move(next.Dst, dst)
	return true
}
