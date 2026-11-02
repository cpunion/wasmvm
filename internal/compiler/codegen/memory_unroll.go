package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type memoryUnroll struct {
	cursor, bound slots.Slot
	constant      uint32
	copies, exit  int
	backYes       bool
	keepPredicate bool
}

// Prove one unsigned, unit-stride local recurrence and an invariant bound.
// The existing region proof supplies every memory range; no observation or
// potentially trapping operation may occur inside an unrolled group.
func (e *functionEmitter) memoryUnroll(start int, r memoryRegion) *memoryUnroll {
	end := r.end + 1
	length := end - start + 1 // Include the original conditional backedge.
	if r.loop != nil || len(r.guards) != 0 || end >= len(e.f.Code) ||
		length < 4 || length > 128 || start/256 != end/256 {
		return nil
	}
	branch, cmp := e.f.Code[end], e.f.Code[r.end]
	if branch.Op != slots.ControlBase+uint32(ir.If) || branch.Count != 2 || branch.A != cmp.Dst {
		return nil
	}
	yes, no := e.f.Edges[branch.Extra], e.f.Edges[branch.Extra+1]
	outside := func(pc uint32) bool { return int(pc) < start || int(pc) > end }
	backYes := cmp.Op == 0x49 && int(yes.PC) == start && outside(no.PC)
	backNo := cmp.Op == 0x4f && int(no.PC) == start && outside(yes.PC)
	if yes.Count != 0 || no.Count != 0 || !(backYes || backNo) {
		return nil
	}
	type value struct {
		local slots.Slot
		bias  uint32
		valid bool
	}
	values := map[slots.Slot]value{}
	mutations := map[slots.Slot]int{}
	get := func(s slots.Slot) value {
		if v, ok := values[s]; ok {
			return v
		}
		if s < slots.Slot(e.f.LocalCount) && e.f.Slots[s] == wasm.I32 {
			return value{local: s, valid: true}
		}
		return value{}
	}
	write := func(s slots.Slot, v value) {
		if s == slots.NoSlot {
			return
		}
		if e.f.Slots[s] != wasm.I32 {
			v = value{}
		}
		if s < slots.Slot(e.f.LocalCount) && get(s) != v {
			mutations[s]++
		}
		values[s] = v
	}
	for pc := start; pc < r.end; pc++ {
		in := e.f.Code[pc]
		if !plainOpcode(in.Op) || scalarCanTrap(in.Op) || isMemoryAccess(in.Op) && !r.proven[pc] {
			return nil
		}
		if isMemoryAccess(in.Op) {
			// Keep subword regions compact; their expansion cost is not
			// covered by this conservative word-access profitability rule.
			if _, width := memoryAccess(in.Op); width < 4 {
				return nil
			}
		}
		src := e.f.Sources[in.Source].Wasm
		a, b, v := get(in.A), get(in.B), value{}
		switch in.Op {
		case 0x20:
			v = get(slots.Slot(src.A))
		case 0x21, 0x22:
			v = a
			write(slots.Slot(src.A), v)
		case 0x41:
			v = value{slots.NoSlot, uint32(src.Bits), true}
		case 0x6a:
			if a.valid && b.valid && (a.local == slots.NoSlot || b.local == slots.NoSlot) {
				v = value{a.local, a.bias + b.bias, true}
				if v.local == slots.NoSlot {
					v.local = b.local
				}
			}
		}
		write(in.Dst, v)
	}
	current, bound := get(cmp.A), get(cmp.B)
	if !current.valid || current.local == slots.NoSlot || current.bias != 1 ||
		mutations[current.local] != 1 || get(current.local) != current ||
		!bound.valid || bound.local != slots.NoSlot && (bound.bias != 0 || mutations[bound.local] != 0) {
		return nil
	}
	// Expansion is profitable only when successive accesses can overlap.
	// Invariant-address updates and pointer chasing retain their compact loop;
	// repeating their serial load/store chain can regress store forwarding.
	type origin struct{ independent, varying bool }
	origins := map[slots.Slot]origin{}
	input := func(s slots.Slot) origin {
		if s == slots.NoSlot {
			return origin{independent: true}
		}
		if v, ok := origins[s]; ok {
			return v
		}
		if s < slots.Slot(e.f.LocalCount) {
			v := get(s)
			if s == current.local || mutations[s] == 1 && v.valid && v.local == s && v.bias != 0 {
				return origin{independent: true, varying: true}
			}
			return origin{independent: mutations[s] == 0}
		}
		return origin{}
	}
	for pc := start; pc < r.end; pc++ {
		in := e.f.Code[pc]
		a, b, c := input(in.A), input(in.B), input(in.C)
		v := origin{a.independent && b.independent && c.independent, a.varying || b.varying || c.varying}
		if isMemoryAccess(in.Op) {
			if !a.independent || !a.varying {
				return nil
			}
			v = origin{} // A loaded value is not independent address generation.
		} else if in.Op == 0x20 {
			v = input(slots.Slot(e.f.Sources[in.Source].Wasm.A))
		} else if in.Op == 0x21 || in.Op == 0x22 {
			v = a
			origins[slots.Slot(e.f.Sources[in.Source].Wasm.A)] = v
		}
		if in.Dst != slots.NoSlot {
			origins[in.Dst] = v
		}
	}
	exit := int(no.PC)
	if backNo {
		exit = int(yes.PC)
	}
	copies := 4
	if length*copies > 256 {
		copies = 2
	}
	// No poll, trap or helper can observe the frame between these copies.
	// Keep a repeated comparison result only if the next body can read it
	// before overwriting it. The final comparison is always emitted normally.
	keepPredicate := true
	for pc := start; pc < r.end; pc++ {
		in := e.f.Code[pc]
		local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
		read, write := scalarSlotUse(in, local, cmp.Dst)
		if read {
			break
		}
		if write {
			keepPredicate = false
			break
		}
	}
	return &memoryUnroll{current.local, bound.local, bound.bias, copies, exit, backYes, keepPredicate}
}

func (e *functionEmitter) emitMemoryUnroll(start int, r memoryRegion, pcs []int, out *CompiledFunction) {
	u := r.unroll
	cold := e.a.label()
	// Each complete group executes at most 256 original slot instructions.
	// Original dispatcher PCs always retain their ordinary checked entries.
	if start%256 != 0 {
		e.interrupt(uint32(start))
	}
	if u.bound == slots.NoSlot {
		e.constant(0, uint64(u.constant))
	} else {
		e.slotLoad(u.bound, 0, false, false)
	}
	e.slotLoad(u.cursor, 1, false, false)
	e.compare(false)
	e.branch(cold, "lo") // Reject unsigned wrap before subtracting the cursor.
	if e.a.arch == "arm64" {
		e.a.word(0x4b0a0129) // sub w9, bound, cursor
	} else {
		e.a.bytes(0x29, 0xc8) // sub eax, ecx
	}
	e.constant(1, uint64(u.copies))
	e.compare(false)
	e.branch(cold, "lo")
	cmp := e.f.Code[r.end]
	for i := 0; i < u.copies; i++ {
		if i+1 == u.copies {
			e.emitMemoryRegionRange(start, r.end, r, pcs, out)
		} else {
			e.emitMemoryRegionRange(start, r.end-1, r, pcs, out)
			// The guard proves these comparisons continue the loop. A next
			// iteration that can read the scalar still receives its exact value.
			if u.keepPredicate {
				predicate := uint64(1)
				if !u.backYes {
					predicate = 0
				}
				e.constant(0, predicate)
				e.slotStore(cmp.Dst, 0, false)
			}
		}
	}
	e.slotLoad(cmp.Dst, 0, false, false)
	e.test(false)
	condition := "eq"
	if !u.backYes {
		condition = "ne"
	}
	e.branch(pcs[u.exit], condition)
	e.a.branch(pcs[start], false, 0)
	e.a.mark(cold)
	// Short remainders, unusual initial cursors and independent entries use
	// the original one-iteration body and conditional branch.
	e.emitMemoryRegionRange(start, r.end, r, pcs, out)
	out.UnrolledLoops++
	out.UnrolledCopies += uint32(u.copies)
}
