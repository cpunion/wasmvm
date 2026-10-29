package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type memoryLoop struct {
	first  int
	cursor slots.Slot
	limit  uint32
	yes    bool
	exit   int
}

type loopBackedge struct {
	header int
	yes    bool
	exit   int
}

// A value is an optional invariant local plus one induction variable and a
// constant bias, modulo i32. Unknown/nonlinear values cannot prove a loop span.
type loopValue struct {
	base  slots.Slot
	bias  uint32
	step  bool
	valid bool
}

func (e *functionEmitter) memoryLoop(start int, r memoryRegion) *memoryLoop {
	end := r.end + 1
	if len(r.guards) != 1 || end >= len(e.f.Code) || start/256 != end/256 || end-start < 4 {
		return nil
	}
	branch, cmp, bound := e.f.Code[end], e.f.Code[end-1], e.f.Code[end-2]
	if branch.Op != slots.ControlBase+uint32(ir.If) || branch.Count != 2 || branch.A != cmp.Dst || bound.Op != 0x41 {
		return nil
	}
	yes, no := e.f.Edges[branch.Extra], e.f.Edges[branch.Extra+1]
	if yes.Count != 0 || no.Count != 0 {
		return nil
	}
	outside := func(pc uint32) bool { return int(pc) < start || int(pc) > end }
	backYes := int(yes.PC) == start && outside(no.PC)
	backNo := int(no.PC) == start && outside(yes.PC)
	if !(backYes && cmp.Op == 0x47 || backNo && cmp.Op == 0x46) {
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
	first := start
	for pc := range r.guards {
		first = pc
	}
	if first-start > 16 {
		return nil
	}
	root := r.addresses[first].local
	values := map[slots.Slot]loopValue{}
	mutations := map[slots.Slot]int{}
	get := func(s slots.Slot) loopValue {
		if v, found := values[s]; found {
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
	update, lastMemory := -1, -1
	write := func(s slots.Slot, v loopValue, pc int) {
		if s == slots.NoSlot {
			return
		}
		if e.f.Slots[s] != wasm.I32 {
			v = loopValue{}
		}
		if s < slots.Slot(e.f.LocalCount) && get(s) != v {
			mutations[s]++
			if s == cursor {
				update = pc
			}
		}
		values[s] = v
	}
	var address loopValue
	for pc := start; pc < end; pc++ {
		in := e.f.Code[pc]
		if !plainOpcode(in.Op) || scalarCanTrap(in.Op) {
			return nil
		}
		if pc < first && in.Op != 0x20 && in.Op != 0x21 && in.Op != 0x22 && in.Op != 0x41 && in.Op != 0x6a {
			return nil
		}
		if isMemoryAccess(in.Op) {
			a, remapped := r.addresses[pc]
			if !r.proven[pc] || !remapped || a.local != root || pc < first {
				return nil
			}
			lastMemory = pc
			if pc == first {
				address = get(in.A)
			}
		}
		src := e.f.Sources[in.Source].Wasm
		a, b, v := get(in.A), get(in.B), loopValue{}
		switch in.Op {
		case 0x20:
			v = get(slots.Slot(src.A))
		case 0x21, 0x22:
			v = a
			write(slots.Slot(src.A), v, pc)
		case 0x41:
			v = loopValue{base: slots.NoSlot, bias: uint32(src.Bits), valid: true}
		case 0x6a:
			if a.valid && b.valid && !(a.step && b.step) && (a.base == slots.NoSlot || b.base == slots.NoSlot) {
				v = loopValue{base: a.base, bias: a.bias + b.bias, step: a.step || b.step, valid: true}
				if v.base == slots.NoSlot {
					v.base = b.base
				}
			}
		}
		write(in.Dst, v, pc)
	}
	if !address.valid || !address.step || address.base != slots.NoSlot && mutations[address.base] != 0 ||
		mutations[cursor] != 1 || update <= lastMemory || get(cursor) != (loopValue{base: slots.NoSlot, bias: 1, step: true, valid: true}) {
		return nil
	}
	exit := int(no.PC)
	if backNo {
		exit = int(yes.PC)
	}
	return &memoryLoop{first: first, cursor: cursor, limit: limit, yes: backYes, exit: exit}
}

func (e *functionEmitter) emitMemoryLoop(start int, r memoryRegion, pcs []int, out *CompiledFunction) {
	loop := r.loop
	// The original prefix supplies the exact frame expected by a failed guard.
	// Its scalar-only clone is the sole target of the private successful backedge.
	e.emitMemoryRegionRange(start, loop.first-1, r, pcs, &CompiledFunction{})
	cold := pcs[loop.first]
	if loop.first == start {
		cold = e.a.label()
		e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, loop.first})
	}
	e.emitLoopGuard(*loop, r.guards[loop.first], cold)
	out.MemoryGuards++
	out.LoopGuards++
	header, body := e.a.label(), e.a.label()
	e.a.branch(body, false, 0)
	e.a.mark(header)
	if start%256 == 0 {
		e.interrupt(uint32(start))
	}
	warm := r
	warm.guards, warm.loop = nil, nil
	warm.elided = e.deadAddressTemps(start, r.end+1, warm)
	out.DeadScalarOps += uint32(len(warm.elided))
	e.emitMemoryRegionRange(start, loop.first-1, warm, pcs, out)
	e.a.mark(body)
	e.emitMemoryRegionRange(loop.first, r.end, warm, pcs, out)
	if e.loopBackedges == nil {
		e.loopBackedges = map[int]loopBackedge{}
	}
	e.loopBackedges[r.end+1] = loopBackedge{header: header, yes: loop.yes, exit: loop.exit}
}

// For i < limit, future roots advance by at most limit-1-i. Check that complete
// remaining span with widened arithmetic, then repeat without guards. No loop
// proof survives entry through an original PC or a return to the host driver.
func (e *functionEmitter) emitLoopGuard(loop memoryLoop, g memoryGuard, cold int) {
	e.slotLoad(loop.cursor, 0, false, false)
	e.constant(1, uint64(loop.limit))
	e.compare(false)
	e.branch(cold, "hs")
	e.constant(1, uint64(loop.limit-1))
	if e.a.arch == "arm64" {
		e.a.word(0x4b09014a) // sub w10, last, cursor
	} else {
		e.a.bytes(0x29, 0xc1) // sub ecx, eax
	}
	e.slotLoad(g.address, 0, false, false)
	if g.before != 0 {
		if e.a.arch == "arm64" {
			if g.before < 4096 {
				e.a.word(0x51000129 | uint32(g.before)<<10)
			} else {
				e.constant(2, g.before)
				e.a.word(0x4b0b0129)
			}
		} else {
			e.a.bytes(0x2d) // sub eax, first bias (mod i32)
			e.a.word(uint32(g.before))
		}
	}
	span := g.before + g.after
	hi := byte(8)
	if e.a.arch == "arm64" {
		e.a.word(0x8b0a0129) // add x9, root, remaining iterations
		e.memoryOffsetARM64(span)
	} else {
		hi = 7
		e.a.bytes(0x48, 0x01, 0xc8)
		if span <= math.MaxInt32 {
			e.a.bytes(0x48, 0x05)
			e.a.word(uint32(span))
		} else {
			e.constant(1, span)
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
