package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type loopSpan struct {
	base   slots.Slot
	cursor bool
	bias   uint32
	extent uint64
}
type memoryLoop struct {
	cursor        slots.Slot
	limit, stride uint32
	yes           bool
	exit          int
	spans         []loopSpan
	proven        map[int]bool
	addresses     map[int]memoryAddress
}
type loopBackedge struct {
	header    int
	yes       bool
	exit      int
	registers *registerPlan
}

// One invariant local plus the entry cursor and a constant bias, modulo i32.
// Recycled temporaries carry values, never hidden local anchors.
type loopValue struct {
	base        slots.Slot
	bias        uint32
	step, valid bool
}

func (e *functionEmitter) memoryLoop(start int, r memoryRegion) *memoryLoop {
	end := r.end + 1
	if end >= len(e.f.Code) || start/256 != end/256 || end-start < 4 {
		return nil
	}
	branch, cmp, bound := e.f.Code[end], e.f.Code[end-1], e.f.Code[end-2]
	if branch.Op != slots.ControlBase+uint32(ir.If) || branch.Count != 2 || branch.A != cmp.Dst || bound.Op != 0x41 {
		return nil
	}
	yes, no := e.f.Edges[branch.Extra], e.f.Edges[branch.Extra+1]
	outside := func(pc uint32) bool { return int(pc) < start || int(pc) > end }
	backYes := int(yes.PC) == start && outside(no.PC)
	backNo := int(no.PC) == start && outside(yes.PC)
	if yes.Count != 0 || no.Count != 0 || !(backYes && cmp.Op == 0x47 || backNo && cmp.Op == 0x46) {
		return nil
	}
	cursor := cmp.A
	if cmp.B != bound.Dst {
		if cmp.A != bound.Dst {
			return nil
		}
		cursor = cmp.B
	}
	if cursor >= slots.Slot(e.f.LocalCount) || e.f.Slots[cursor] != wasm.I32 {
		return nil
	}
	limit := uint32(e.f.Sources[bound.Source].Wasm.Bits)
	if limit < 2 {
		return nil
	}
	values := map[slots.Slot]loopValue{}
	mutations := map[slots.Slot]int{}
	get := func(s slots.Slot) loopValue {
		if v, ok := values[s]; ok {
			return v
		}
		if s < slots.Slot(e.f.LocalCount) && e.f.Slots[s] == wasm.I32 {
			if s == cursor {
				return loopValue{base: slots.NoSlot, step: true, valid: true}
			}
			return loopValue{base: s, valid: true}
		}
		return loopValue{}
	}
	write := func(s slots.Slot, v loopValue) {
		if s == slots.NoSlot {
			return
		}
		if e.f.Slots[s] != wasm.I32 {
			v = loopValue{}
		}
		if s < slots.Slot(e.f.LocalCount) && get(s) != v {
			mutations[s]++
		}
		values[s] = v
	}
	type key struct {
		base   slots.Slot
		cursor bool
	}
	type group struct {
		key
		first, last int64
	}
	var groups []group
	indexes := map[key]int{}
	proven := map[int]bool{}
	addresses := map[int]memoryAddress{}
	for pc := start; pc < end; pc++ {
		in := e.f.Code[pc]
		if !plainOpcode(in.Op) || scalarCanTrap(in.Op) {
			return nil
		}
		src := e.f.Sources[in.Source].Wasm
		a, b, v := get(in.A), get(in.B), loopValue{}
		if isMemoryAccess(in.Op) {
			if !a.valid {
				return nil
			}
			_, size := memoryAccess(in.Op)
			// Memarg offsets do not wrap with their i32 address operand.
			if uint64(src.B)+uint64(size) > 1<<32 {
				return nil
			}
			k := key{a.base, a.step}
			first := int64(int32(a.bias))
			last := first + int64(src.B) + int64(size)
			index, ok := indexes[k]
			if !ok {
				index = len(groups)
				indexes[k] = index
				groups = append(groups, group{k, first, last})
			} else {
				groups[index].first = min(groups[index].first, first)
				groups[index].last = max(groups[index].last, last)
			}
			// Fold an address only if this guard also proves its canonical
			// local anchor plus the complete nonwrapping folded offset.
			if remap, ok := r.addresses[pc]; ok {
				anchor := get(remap.local)
				if anchor.valid && anchor.base == a.base && anchor.step == a.step {
					lo := min(groups[index].first, int64(int32(anchor.bias)))
					hi := max(groups[index].last, int64(int32(anchor.bias))+int64(remap.offset)+int64(size))
					if uint64(hi-lo) <= 1<<32 {
						groups[index].first, groups[index].last = lo, hi
						addresses[pc] = remap
					}
				}
			}
			proven[pc] = true
		}
		switch in.Op {
		case 0x20:
			v = get(slots.Slot(src.A))
		case 0x21, 0x22:
			v = a
			write(slots.Slot(src.A), v)
		case 0x41:
			v = loopValue{base: slots.NoSlot, bias: uint32(src.Bits), valid: true}
		case 0x6a:
			if a.valid && b.valid && !(a.step && b.step) && (a.base == slots.NoSlot || b.base == slots.NoSlot) {
				v = loopValue{base: a.base, bias: a.bias + b.bias, step: a.step || b.step, valid: true}
				if v.base == slots.NoSlot {
					v.base = b.base
				}
			}
		case 0x6b:
			if a.valid && b.valid && b.base == slots.NoSlot && !b.step {
				v = a
				v.bias -= b.bias
			}
		}
		write(in.Dst, v)
	}
	update := get(cursor)
	if len(groups) == 0 || mutations[cursor] != 1 || !update.valid || !update.step || update.base != slots.NoSlot || update.bias == 0 {
		return nil
	}
	spans := make([]loopSpan, 0, len(groups))
	for _, g := range groups {
		if g.base != slots.NoSlot && mutations[g.base] != 0 || uint64(g.last-g.first) > 1<<32 {
			return nil
		}
		spans = append(spans, loopSpan{g.base, g.cursor, uint32(g.first), uint64(g.last - g.first)})
	}
	exit := int(no.PC)
	if backNo {
		exit = int(yes.PC)
	}
	return &memoryLoop{cursor: cursor, limit: limit, stride: update.bias, yes: backYes, exit: exit, spans: spans, proven: proven, addresses: addresses}
}

func (e *functionEmitter) emitMemoryLoop(start int, r memoryRegion, pcs []int, out *CompiledFunction) {
	loop := r.loop
	cold := e.a.label()
	e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, start, e.currentRegisterPlan()})
	e.emitLoopProgress(*loop, cold)
	for i, span := range loop.spans {
		// Guards alter only scratch registers. A canceled guard resumes the fully
		// checked entry with unchanged canonical slots, including every raw bit.
		if i%16 == 0 {
			e.interrupt(uint32(start))
		}
		e.emitLoopSpan(*loop, span, cold)
	}
	out.MemoryGuards += uint32(len(loop.spans))
	out.LoopGuards++
	previous := e.currentRegisterPlan()
	plan := e.loopRegisterPlan(start, r)
	if plan != nil {
		e.transitionRegisters(plan)
		out.RegisterRegions++
	}
	coalesced, entry, merges := e.coalesceLoopRegisters(start, r, (r.end-start+1)*128)
	if coalesced != nil {
		e.transitionRegisters(entry)
		e.useRegisterPlan(coalesced)
		out.CoalescedSlotGroups += merges
	}
	header := e.a.label()
	e.a.mark(header)
	if start%256 == 0 {
		e.interrupt(uint32(start))
	}
	warm := r
	warm.guards, warm.loop = nil, nil
	warm.proven = loop.proven
	e.emitMemoryRegionRange(start, r.end, warm, pcs, out)
	if e.loopBackedges == nil {
		e.loopBackedges = map[int]loopBackedge{}
	}
	edge := loopBackedge{header: header, yes: loop.yes, exit: loop.exit}
	if plan != nil || coalesced != nil {
		edge.registers = previous
	}
	e.loopBackedges[r.end+1] = edge
}

// Equality exits must be reached before the cursor wraps. Nondivisible strides
// and all other progress failures take the original checked loop, not a trap.
func (e *functionEmitter) emitLoopProgress(loop memoryLoop, cold int) {
	e.slotLoad(loop.cursor, 0, false, false)
	e.constant(1, uint64(loop.limit))
	e.compare(false)
	e.branch(cold, "hs")
	if loop.stride == 1 {
		return
	}
	if e.a.arch == "arm64" {
		e.a.word(0x4b09014a)
	} else {
		e.a.bytes(0x29, 0xc1)
	}
	if loop.stride&(loop.stride-1) == 0 {
		if e.a.arch == "arm64" {
			e.constant(2, uint64(loop.stride-1))
			e.a.word(0x6a0b014b)
		} else {
			e.a.bytes(0xf7, 0xc1)
			e.a.word(loop.stride - 1)
		}
	} else if e.a.arch == "arm64" {
		e.constant(2, uint64(loop.stride))
		e.a.word(0x1ac00800 | 11<<16 | 10<<5 | 9)         // udiv w9,w10,w11
		e.a.word(0x1b008000 | 11<<16 | 10<<10 | 9<<5 | 9) // msub w9,w9,w11,w10
		e.a.word(0x7100013f)                              // cmp w9,#0
	} else {
		e.a.bytes(0x89, 0xc8)
		e.constant(1, uint64(loop.stride))
		e.a.bytes(0x31, 0xd2, 0xf7, 0xf1, 0x85, 0xd2)
	}
	e.branch(cold, "ne")
}

// The initial address wraps exactly once as an i32 operand. Its whole future
// span and nonwrapping memarg are then checked with widened arithmetic. Reading
// basis locals at the entry creates no dependency on interior temporary slots.
func (e *functionEmitter) emitLoopSpan(loop memoryLoop, span loopSpan, cold int) {
	if span.cursor {
		e.slotLoad(loop.cursor, 0, false, false)
		e.constant(1, uint64(loop.limit-loop.stride))
		if e.a.arch == "arm64" {
			e.a.word(0x4b09014a)
		} else {
			e.a.bytes(0x29, 0xc1)
		}
	}
	if span.base == slots.NoSlot {
		e.constant(0, 0)
	} else {
		e.slotLoad(span.base, 0, false, false)
	}
	if span.cursor {
		e.slotLoad(loop.cursor, 2, false, false)
		if e.a.arch == "arm64" {
			e.a.word(0x0b0b0129)
		} else {
			e.a.bytes(0x01, 0xd0)
		}
	}
	if span.bias != 0 {
		e.constant(2, uint64(span.bias))
		if e.a.arch == "arm64" {
			e.a.word(0x0b0b0129)
		} else {
			e.a.bytes(0x01, 0xd0)
		}
	}
	if span.cursor {
		if e.a.arch == "arm64" {
			e.a.word(0x8b0a0129)
		} else {
			e.a.bytes(0x48, 0x01, 0xc8)
		}
	}
	hi := byte(8)
	if e.a.arch == "arm64" {
		e.memoryOffsetARM64(span.extent)
	} else {
		hi = 7
		if span.extent <= math.MaxInt32 {
			e.a.bytes(0x48, 0x05)
			e.a.word(uint32(span.extent))
		} else {
			e.constant(1, span.extent)
			e.a.bytes(0x48, 0x01, 0xc8)
		}
	}
	e.constant(1, 1<<32)
	e.compare(true)
	e.a.branch(cold, true, hi)
	if e.a.arch == "arm64" {
		e.a.word(0xeb11013f)
	} else {
		e.a.bytes(0x4c, 0x39, 0xd8)
	}
	e.a.branch(cold, true, hi)
}
