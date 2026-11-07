package codegen

import (
	"cmp"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopFrameCopies struct {
	first    int
	aliases  []map[slots.Slot]slots.Slot
	deferred map[slots.Slot]bool
	boundary []bool
	next     []int
	final    []loopSlotCopy
}

// Read identities precede writes at the same instruction. The existing local
// graph retains raw copy identities, including versions with no future reads.
func (v *loopValues) before(s slots.Slot, pc int) (uint32, bool) {
	versions := v.bySlot[s]
	i, _ := slices.BinarySearchFunc(versions, pc*2, func(index, point int) int {
		return cmp.Compare(v.versions[index].first, point)
	})
	if i == 0 {
		return 0, false
	}
	return v.versions[versions[i-1]].value, i > 1
}

// Keep an independent materialized anchor at every point. Reserving anchors
// makes later selections unable to introduce reconstruction cycles. Targets
// have no entry read, so a private backedge can start the next iteration without
// restoring their old frame words. Every actual exit reconstructs them first.
func (e *functionEmitter) planLoopFrameCopies(v *loopValues, r memoryRegion, work int) *loopFrameCopies {
	if v == nil || r.values != nil {
		return nil
	}
	shared := e.coalescedRegisterSlots()
	var tracked, candidates []slots.Slot
	for s, versions := range v.bySlot {
		if shared[s] {
			continue // New reads must not extend a shared home's proven lifetime.
		}
		tracked = append(tracked, s)
		kind := e.f.Slots[s]
		if kind != wasm.I32 && kind != wasm.I64 || len(versions) < 2 || v.versions[versions[0]].last >= v.versions[versions[0]].first {
			continue
		}
		if kind, _ := registerLocation(e.currentRegisterPlan(), s); kind != 0 {
			continue
		}
		valid := true
		for _, index := range versions[1:] {
			in := e.f.Code[v.versions[index].first/2]
			local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
			valid = valid && (in.Op == 0x20 && in.Dst == s || (in.Op == 0x21 || in.Op == 0x22) && (local == s || in.Dst == s))
		}
		if valid {
			candidates = append(candidates, s)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	plan := e.currentRegisterPlan()
	slices.SortFunc(tracked, func(a, b slots.Slot) int {
		ak, _ := registerLocation(plan, a)
		bk, _ := registerLocation(plan, b)
		if (ak != 0) != (bk != 0) {
			if ak != 0 {
				return -1
			}
			return 1
		}
		return cmp.Compare(a, b)
	})
	slices.Sort(candidates)
	n := r.end - v.first + 2 // Include the branch/poll snapshot.
	result := &loopFrameCopies{first: v.first, aliases: make([]map[slots.Slot]slots.Slot, n), deferred: map[slots.Slot]bool{}, boundary: make([]bool, n)}
	reserved := map[slots.Slot]bool{}
	for _, target := range candidates {
		if reserved[target] {
			continue
		}
		anchors := make([]slots.Slot, n)
		valid := true
		for i := range anchors {
			anchors[i] = slots.NoSlot
			value, written := v.before(target, v.first+i)
			if !written {
				continue
			}
			for _, source := range tracked {
				work--
				if work < 0 {
					valid = false
					break
				}
				if source == target || result.deferred[source] || e.f.Slots[source] != e.f.Slots[target] {
					continue
				}
				if other, _ := v.before(source, v.first+i); other == value {
					anchors[i] = source
					break
				}
			}
			if anchors[i] == slots.NoSlot {
				valid = false
				break
			}
		}
		if !valid {
			if work < 0 {
				break // Previously selected targets retain independent anchors.
			}
			continue
		}
		result.deferred[target] = true
		for i, source := range anchors {
			if source != slots.NoSlot {
				if result.aliases[i] == nil {
					result.aliases[i] = map[slots.Slot]slots.Slot{}
				}
				result.aliases[i][target] = source
				reserved[source] = true
			}
		}
		for _, index := range v.bySlot[target][1:] {
			i := v.versions[index].first/2 - v.first
			result.boundary[i], result.boundary[i+1] = true, true
		}
		result.final = append(result.final, loopSlotCopy{target, anchors[n-1]})
	}
	if len(result.final) == 0 {
		return nil
	}
	result.next = make([]int, n)
	for i := 1; i < n; i++ {
		if !equalFrameAliases(result.aliases[i-1], result.aliases[i]) {
			result.boundary[i] = true
		}
	}
	boundary := v.first + n
	for i := n - 1; i >= 0; i-- {
		result.next[i] = boundary
		if result.boundary[i] {
			boundary = v.first + i
		}
	}
	return result
}

func equalFrameAliases(a, b map[slots.Slot]slots.Slot) bool {
	if len(a) != len(b) {
		return false
	}
	for target, source := range a {
		if other, ok := b[target]; !ok || other != source {
			return false
		}
	}
	return true
}

func (e *functionEmitter) frameRead(slot slots.Slot) slots.Slot {
	if e.frameAliases == nil {
		return slot
	}
	if source, ok := e.frameAliases[slot]; ok {
		return source
	}
	return slot
}

func (e *functionEmitter) materializeFrameCopies(copies []loopSlotCopy) {
	for _, copy := range copies {
		e.move(copy.dst, copy.src)
	}
}
