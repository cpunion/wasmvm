package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Forward copies along a short, nontrapping fallthrough. Retain their values in
// their original slots, but remove copies from the arithmetic dependency chain.
// Every interior instruction keeps an independent cold continuation.
func (e *functionEmitter) combineIntegerCopies(start int) (int, bool) {
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
		} else if in.Op == 0x41 || in.Op == 0x42 {
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
			if e.canRegisterInteger(candidate.Op, candidate.Dst, candidate.A, candidate.B) {
				in, end = candidate, pc+1
			}
		}
	}
	if !e.canRegisterInteger(in.Op, in.Dst, in.A, in.B) {
		return 0, false
	}
	for i, copy := range e.f.Code[start:pc] {
		// Forwarded arithmetic no longer reads this temporary. Elide the
		// copy only if no later prefix instruction needs it and normal
		// execution replaces it before any poll, trap or other exit.
		needed := copy.Dst == in.A || copy.Dst == in.B || copy.Dst == in.C
		for _, later := range e.f.Code[start+i+1 : pc] {
			if later.Op == 0x20 && slots.Slot(e.f.Sources[later.Source].Wasm.A) == copy.Dst {
				needed = true
			}
		}
		if !needed && (copy.Dst == original.Dst || e.overwrittenBeforeExit(pc+1, copy.Dst)) {
			continue
		}
		if copy.Op == 0x20 {
			e.move(copy.Dst, slots.Slot(e.f.Sources[copy.Source].Wasm.A))
		} else {
			e.literal(copy)
		}
	}
	e.registerInteger(in.Op, in.Dst, in.A, in.B)
	if end != pc {
		if !e.overwrittenBeforeExit(end+1, original.Dst) {
			e.move(original.Dst, in.Dst)
		}
		e.move(e.f.Code[end].Dst, in.Dst)
	}
	return end, true
}
