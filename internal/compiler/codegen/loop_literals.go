package codegen

import "slices"

func literalOpcode(op uint32) bool {
	return op >= 0x41 && op <= 0x44 || op == 0xd0 || op == 0xfd000c
}

// A literal can move to entry only when its old word is never read and every
// write in the nontrapping body assigns the same bits. Shared physical homes
// remain excluded: another member may still need its original entry value.
func (e *functionEmitter) loopLiteralDefinitions(v *loopValues) []int {
	if v == nil {
		return nil
	}
	shared := e.coalescedRegisterSlots()
	var definitions []int
	for s, versions := range v.bySlot {
		if len(versions) < 2 || shared[s] {
			continue
		}
		initial := v.versions[versions[0]]
		if initial.last >= initial.first {
			continue
		}
		first := v.versions[versions[1]]
		valid := true
		for _, index := range versions[1:] {
			version := v.versions[index]
			in := e.f.Code[version.first/2]
			if version.value != first.value || !literalOpcode(in.Op) || in.Dst != s {
				valid = false
				break
			}
		}
		if valid {
			definitions = append(definitions, first.first/2)
		}
	}
	slices.Sort(definitions)
	return definitions
}

func (e *functionEmitter) hoistLoopLiterals(v *loopValues, r memoryRegion, out *CompiledFunction) memoryRegion {
	definitions := e.loopLiteralDefinitions(v)
	if len(definitions) == 0 {
		return r
	}
	// Only the private body uses this map. Every original cold PC and guard
	// fallback keeps its literal and complete original instruction semantics.
	elided := make(map[int]bool, len(r.elided)+len(definitions))
	for pc, skip := range r.elided {
		elided[pc] = skip
	}
	for _, pc := range definitions {
		in := e.f.Code[pc]
		e.literal(in)
		out.HoistedLiterals++
		versions := v.bySlot[in.Dst][1:]
		for _, index := range versions {
			elided[v.versions[index].first/2] = true
		}
		out.DeadScalarOps += uint32(len(versions) - 1)
	}
	r.elided = elided
	return r
}
