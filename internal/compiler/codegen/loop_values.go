package codegen

import (
	"encoding/binary"
	"slices"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopSlotVersion struct {
	value       uint32
	first, last int
}
type loopSlotCopy struct{ dst, src slots.Slot }
type loopLiteral struct {
	typ    wasm.ValueType
	lo, hi uint64
}
type loopValues struct {
	first    int
	next     uint32
	versions []loopSlotVersion
	current  map[slots.Slot]int
	bySlot   map[slots.Slot][]int
	literals map[loopLiteral]uint32
	copies   []loopSlotCopy
}

func (v *loopValues) initial(s slots.Slot) int {
	if index, ok := v.current[s]; ok {
		return index
	}
	index := len(v.versions)
	v.versions = append(v.versions, loopSlotVersion{v.next, v.first*2 - 1, -2})
	v.next++
	v.current[s] = index
	v.bySlot[s] = append(v.bySlot[s], index)
	return index
}
func (v *loopValues) read(s slots.Slot, pc int) uint32 {
	if s == slots.NoSlot {
		return 0
	}
	index := v.initial(s)
	v.versions[index].last = max(v.versions[index].last, pc*2)
	return v.versions[index].value
}
func (v *loopValues) write(s slots.Slot, value uint32, pc int) {
	if s == slots.NoSlot {
		return
	}
	v.initial(s)
	index := len(v.versions)
	// Closed instruction points keep different input/output values apart.
	// Thus coalescing does not add new physical operand aliases to encoders.
	v.versions = append(v.versions, loopSlotVersion{value, pc * 2, pc * 2})
	v.current[s] = index
	v.bySlot[s] = append(v.bySlot[s], index)
}
func (v *loopValues) fresh() uint32 {
	result := v.next
	v.next++
	return result
}
func (v *loopValues) interfere(a, b slots.Slot, work *int) (bool, bool) {
	if v.versions[v.current[a]].value != v.versions[v.current[b]].value {
		return true, true // Every final frame word is observed at the backedge.
	}
	for _, ai := range v.bySlot[a] {
		x := v.versions[ai]
		if x.last < x.first {
			continue
		}
		for _, bi := range v.bySlot[b] {
			if *work == 0 {
				return false, false
			}
			*work--
			y := v.versions[bi]
			if y.last >= y.first && x.value != y.value && max(x.first, y.first) <= min(x.last, y.last) {
				return true, true
			}
		}
	}
	return false, true
}

type loopValueGroups struct {
	parent  []int
	members [][]slots.Slot
}

func (g *loopValueGroups) root(i int) int {
	for g.parent[i] != i {
		g.parent[i] = g.parent[g.parent[i]]
		i = g.parent[i]
	}
	return i
}

func cloneRegisterPlan(p *registerPlan) *registerPlan {
	clone := func(source map[slots.Slot]byte) map[slots.Slot]byte {
		result := make(map[slots.Slot]byte, len(source))
		for s, r := range source {
			result[s] = r
		}
		return result
	}
	return &registerPlan{homes: clone(p.homes), floats: clone(p.floats), vectors: clone(p.vectors), slots: append([]slots.Slot(nil), p.slots...), shared: p.shared}
}
func setRegisterHome(p *registerPlan, s slots.Slot, kind, reg byte) {
	switch kind {
	case 1:
		p.homes[s] = reg
	case 2:
		p.floats[s] = reg
	case 3:
		p.vectors[s] = reg
	}
}

func (e *functionEmitter) coalescedRegisterSlots() map[slots.Slot]bool {
	p := e.currentRegisterPlan()
	if p == nil || !p.shared {
		return nil
	}
	owners := map[uint16]slots.Slot{}
	var shared map[slots.Slot]bool
	for _, s := range p.slots {
		kind, reg := registerLocation(p, s)
		key := uint16(kind)<<8 | uint16(reg)
		if previous, ok := owners[key]; ok {
			if shared == nil {
				shared = map[slots.Slot]bool{}
			}
			shared[previous], shared[s] = true, true
		} else {
			owners[key] = s
		}
	}
	return shared
}

// Version liveness includes every final frame word, not just future guest
// reads. Only proven nontrapping loop bodies have no intervening observation.
// One version graph serves coalescing and invariant literal materialization.
func (e *functionEmitter) loopValueVersions(start int, r memoryRegion) *loopValues {
	if !e.unmetered || r.loop == nil || start%256 == 0 {
		return nil
	}
	v := loopValues{first: start, current: map[slots.Slot]int{}, bySlot: map[slots.Slot][]int{}, literals: map[loopLiteral]uint32{}}
	for pc := start; pc <= r.end; pc++ {
		if r.elided[pc] {
			continue
		}
		in := e.f.Code[pc]
		if !plainOpcode(in.Op) || scalarCanTrap(in.Op) || isMemoryAccess(in.Op) && !r.proven[pc] {
			return nil
		}
		local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
		switch in.Op {
		case 0x20:
			value := v.read(local, pc)
			v.write(in.Dst, value, pc)
			v.copies = append(v.copies, loopSlotCopy{in.Dst, local})
		case 0x21, 0x22:
			if in.A == slots.NoSlot {
				return nil
			}
			value := v.read(in.A, pc)
			v.write(local, value, pc)
			v.write(in.Dst, value, pc)
			v.copies = append(v.copies, loopSlotCopy{local, in.A}, loopSlotCopy{in.Dst, in.A})
		case 0x41, 0x42, 0x43, 0x44, 0xfd000c, 0xd0:
			if in.Dst == slots.NoSlot {
				continue
			}
			source := e.f.Sources[in.Source].Wasm
			key := loopLiteral{typ: e.f.Slots[in.Dst], lo: source.Bits}
			if in.Op == 0x41 || in.Op == 0x43 {
				key.lo = uint64(uint32(key.lo))
			} else if in.Op == 0xd0 {
				key.lo = 0
			} else if in.Op == 0xfd000c {
				key.lo, key.hi = binary.LittleEndian.Uint64(source.Vector[:8]), binary.LittleEndian.Uint64(source.Vector[8:])
			}
			value, ok := v.literals[key]
			if !ok {
				value = v.fresh()
				v.literals[key] = value
			}
			v.write(in.Dst, value, pc)
		default:
			if address, ok := r.addresses[pc]; ok {
				in.A = address.local
			}
			v.read(in.A, pc)
			v.read(in.B, pc)
			v.read(in.C, pc)
			if in.Dst != slots.NoSlot {
				v.write(in.Dst, v.fresh(), pc)
			}
		}
	}
	for _, index := range v.current {
		v.versions[index].last = (r.end + 1) * 2
	}
	return &v
}

// Sharing all three value classes retains fixed homes throughout the body.
// Every overlapping version must have identical bits.
func (e *functionEmitter) coalesceLoopRegisters(v *loopValues, work int) (*registerPlan, *registerPlan, uint32) {
	old := e.currentRegisterPlan()
	if v == nil || old == nil || len(old.slots) == 0 {
		return nil, nil, 0
	}
	// A copied value need not already own a register. Include nonresident
	// versions in the same interference graph, then reuse a resident member's
	// existing home. This adds aliases, never an additional physical register.
	tracked := append([]slots.Slot(nil), old.slots...)
	resident := map[slots.Slot]bool{}
	for _, s := range old.slots {
		resident[s] = true
	}
	var extra []slots.Slot
	for s := range v.bySlot {
		if !resident[s] {
			extra = append(extra, s)
		}
	}
	slices.Sort(extra)
	tracked = append(tracked, extra...)
	indices := map[slots.Slot]int{}
	g := loopValueGroups{}
	for i, s := range tracked {
		indices[s] = i
		g.parent = append(g.parent, i)
		g.members = append(g.members, []slots.Slot{s})
	}
	for _, copy := range v.copies {
		a, aok := indices[copy.dst]
		b, bok := indices[copy.src]
		if !aok || !bok || e.f.Slots[copy.dst] != e.f.Slots[copy.src] {
			continue
		}
		a, b = g.root(a), g.root(b)
		if a == b {
			continue
		}
		conflict := false
		for _, x := range g.members[a] {
			for _, y := range g.members[b] {
				interferes, complete := v.interfere(x, y, &work)
				if !complete {
					return nil, nil, 0
				}
				conflict = conflict || interferes
			}
		}
		if conflict {
			continue
		}
		if a > b {
			a, b = b, a
		}
		g.parent[b] = a
		g.members[a] = append(g.members[a], g.members[b]...)
		g.members[b] = nil
	}
	var full, entry *registerPlan
	merges := uint32(0)
	for _, members := range g.members {
		if len(members) < 2 {
			continue
		}
		var kind, reg byte
		for _, s := range members {
			if kind, reg = registerLocation(old, s); kind != 0 {
				break
			}
		}
		if kind == 0 {
			continue // A memory-only group does not acquire a new register.
		}
		if full == nil {
			full, entry = cloneRegisterPlan(old), cloneRegisterPlan(old)
			full.shared, entry.shared = true, false
		}
		representative := members[0]
		for _, s := range members {
			initial := v.versions[v.bySlot[s][0]]
			if initial.last >= initial.first {
				representative = s // Interference permits at most one live-in.
				break
			}
		}
		if k, r := registerLocation(old, representative); k != 0 {
			kind, reg = k, r
		}
		for _, s := range members {
			setRegisterHome(full, s, kind, reg)
			delete(entry.homes, s)
			delete(entry.floats, s)
			delete(entry.vectors, s)
		}
		setRegisterHome(entry, representative, kind, reg)
		merges += uint32(len(members) - 1)
	}
	if merges == 0 {
		return nil, nil, 0
	}
	full.slots = full.slots[:0]
	entry.slots = entry.slots[:0]
	for _, s := range tracked {
		if kind, _ := registerLocation(full, s); kind != 0 {
			full.slots = append(full.slots, s)
		}
		if kind, _ := registerLocation(entry, s); kind != 0 {
			entry.slots = append(entry.slots, s)
		}
	}
	return full, entry, merges
}
