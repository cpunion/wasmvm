package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type boundedValue struct {
	max      uint64
	constant bool
}
type memoryRegion struct {
	end       int
	proven    map[int]bool
	guards    map[int]memoryGuard
	guarded   uint32
	addresses map[int]memoryAddress
	elided    map[int]bool
	loop      *nativeLoop
	unroll    *memoryUnroll
	values    *loopValuePlans
	copies    *loopFrameCopies
	predicate *loopPredicate
	operands  map[int]memoryOperand
	cache     *loopMemoryValue
}

func (e *functionEmitter) codeLimit() int {
	limit := len(e.f.Code)
	if e.regionLimit != 0 {
		limit = min(limit, e.regionLimit)
	}
	if e.nativeRegionLimit != 0 {
		limit = min(limit, e.nativeRegionLimit)
	}
	if e.frameCopyLimit != 0 {
		limit = min(limit, e.frameCopyLimit)
	}
	if e.allocationLimit != 0 {
		limit = min(limit, e.allocationLimit)
	}
	return limit
}

// A bounded local abstract interpretation, never a speculative loop invariant.
// Facts start unknown at every CFG entry, helper, growth and poll boundary. Each
// region has at most 256 instructions, so compiler work stays linear in input.
func (e *functionEmitter) memoryRegions() map[int]memoryRegion {
	if !e.unmetered || !e.usesMemory || e.memoryMinimum == 0 || e.memoryViews && !e.sharedDefaultMemory {
		return nil
	}
	for _, in := range e.f.Code {
		if isMemoryAccess(in.Op) && (e.f.Slots[in.A] == wasm.I64 || e.f.Sources[in.Source].Wasm.MemoryOffsetHigh != 0) {
			return nil // Existing affine proofs are specifically modulo i32.
		}
	}
	starts := make([]bool, len(e.f.Code))
	starts[0] = true
	for _, edge := range e.f.Edges {
		starts[edge.PC] = true
	}
	for i, in := range e.f.Code {
		if i%256 == 0 {
			starts[i] = true
		}
		if !plainOpcode(in.Op) && i+1 < len(starts) {
			starts[i+1] = true
		}
	}
	regions := map[int]memoryRegion{}
	for start := 0; start < len(e.f.Code); {
		end := start
		for end < len(e.f.Code) && plainOpcode(e.f.Code[end].Op) && (end == start || !starts[end]) {
			end++
		}
		if end == start {
			start++
			continue
		}
		values := map[slots.Slot]boundedValue{}
		get := func(s slots.Slot) boundedValue {
			if v, ok := values[s]; ok {
				return v
			}
			return boundedValue{max: math.MaxUint32}
		}
		proven := map[int]bool{}
		candidates := 0
		for pc := start; pc < end; pc++ {
			in := e.f.Code[pc]
			src := e.f.Sources[in.Source].Wasm
			a := get(in.A)
			if isMemoryAccess(in.Op) {
				_, size := memoryAccess(in.Op)
				if a.max+uint64(src.B)+uint64(size) <= e.memoryMinimum {
					proven[pc] = true
				} else {
					candidates++
				}
			}
			updateI32Bounds(e.f, &e.f.Code[pc], values)
		}
		var guards map[int]memoryGuard
		var addresses map[int]memoryAddress
		var guarded uint32
		if candidates >= 3 {
			guards, addresses, guarded = e.memoryGuards(start, end, proven)
		}
		r := memoryRegion{end: end - 1, proven: proven, guards: guards, guarded: guarded, addresses: addresses}
		if proof := e.memoryLoop(start, r); proof != nil {
			r.loop = &nativeLoop{yes: proof.yes, exit: proof.exit, flags: true, memory: proof}
			r.proven, r.guards, r.addresses = proof.proven, nil, proof.addresses
			r.elided = e.deadAddressTemps(start, end, r)
			r.guarded = uint32(len(r.proven))
		} else {
			if len(guards) != 0 {
				r.elided = e.deadAddressTemps(start, end, r)
			}
			r.unroll = e.memoryUnroll(start, r)
		}
		r.operands = e.memoryOperandFacts(start, end, r)
		if len(proven) != 0 || r.loop != nil {
			regions[start] = r
		}
		start = end
	}
	return regions
}

// Only uninterrupted normal entry consumes these proofs. Every interior PC is
// emitted again on a cold path, with all bounds checks, until the next region.
// Thus direct resumption never inherits facts from instructions it did not run.
func (e *functionEmitter) emitMemoryRegion(start int, r memoryRegion, pcs []int, out *CompiledFunction) {
	out.GuardedAccesses += r.guarded
	if r.unroll != nil {
		e.emitMemoryUnroll(start, r, pcs, out)
		return
	}
	if r.loop != nil {
		e.emitNativeLoop(start, r, pcs, out)
		return
	}
	out.DeadScalarOps += uint32(len(r.elided))
	e.emitMemoryRegionRange(start, r.end, r, pcs, out)
}

func (e *functionEmitter) emitMemoryRegionRange(start, end int, r memoryRegion, pcs []int, out *CompiledFunction) {
	previousLimit := e.allocationLimit
	e.regionLimit = end + 1
	e.regionEnd = r.end + 1
	e.provenPCs = r.proven
	e.memoryOperands = r.operands
	e.guardPCs = r.guards
	e.addresses, e.elidedPCs = r.addresses, r.elided
	e.regionPredicate = r.predicate
	e.vectorAliases = map[slots.Slot]slots.Slot{}
	defer func() {
		e.frameAliases, e.deferredFrameSlots = nil, nil
		e.frameCopyLimit = 0
		e.allocationLimit = previousLimit
		e.regionLimit = 0
		e.regionEnd = 0
		e.provenPCs = nil
		e.memoryOperands = nil
		e.guardPCs = nil
		e.vectorAliases = nil
		e.addresses = nil
		e.elidedPCs = nil
		e.regionPredicate = nil
	}()
	var boundaries []int
	coalesced := e.coalescedRegisterSlots()
	for pc := start; pc <= end; pc++ {
		if r.copies != nil && r.copies.boundary[pc-r.copies.first] {
			boundaries = append(boundaries, pc)
		}
		if r.values != nil && pc > start && r.values.at(pc) != r.values.at(pc-1) {
			boundaries = append(boundaries, pc)
		}
		in := e.f.Code[pc]
		local := slots.NoSlot
		if in.Op == 0x21 || in.Op == 0x22 {
			local = slots.Slot(e.f.Sources[in.Source].Wasm.A)
		}
		// Selectors may extend copy reads or advance a following local write.
		// Split at every shared-home write. Read-only sequences may still fuse:
		// their inputs cannot change physical value before the next boundary.
		if coalesced[in.Dst] || coalesced[local] {
			boundaries = append(boundaries, pc, pc+1)
			continue
		}
		if r.cache != nil && (pc == r.cache.load || pc == r.cache.store) {
			boundaries = append(boundaries, pc, pc+1)
		}
		_, guard := r.guards[pc]
		_, address := r.addresses[pc]
		// Every remapped access must reach memoryAt; earlier copy/vector
		// selection cannot consume it using the original address operand.
		if guard || address || r.elided[pc] {
			boundaries = append(boundaries, pc)
		}
	}
	next := 0
	for pc := start; pc <= end; pc++ {
		if r.copies != nil {
			e.frameAliases = r.copies.aliases[pc-r.copies.first]
			e.deferredFrameSlots = r.copies.deferred
			e.frameCopyLimit = min(end+1, r.copies.next[pc-r.copies.first])
		}
		if r.values != nil {
			e.transitionRegisters(r.values.at(pc))
			e.allocationLimit = min(end+1, r.values.next[pc-r.values.first])
		}
		if guard, ok := r.guards[pc]; ok {
			cold := e.targetPC(pcs, pc)
			if pc == start {
				cold = e.a.label()
				e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, pc, e.currentRegisterPlan()})
			}
			e.emitMemoryGuard(guard, cold)
			out.MemoryGuards++
		}
		for next < len(boundaries) && boundaries[next] <= pc {
			next++
		}
		e.regionLimit = end + 1
		if next < len(boundaries) {
			e.regionLimit = boundaries[next]
		}
		in := e.f.Code[pc]
		if r.elided[pc] {
			continue
		}
		if r.cache != nil && pc == r.cache.load {
			e.loadLoopMemoryValue(in, r.cache)
			continue
		}
		if r.cache != nil && pc == r.cache.store {
			e.memoryAt(in, true, e.f.Sources[in.Source].Wasm.MemoryOffset(), r.operands[pc])
			e.storeLoopMemoryValue(in, r.cache)
			continue
		}
		if address, ok := r.addresses[pc]; ok {
			in.A = address.local
			e.memoryAt(in, true, address.offset, r.operands[pc])
			continue
		}
		if end, ok := e.combinePlain(pc, start, out); ok {
			pc = end
			continue
		}
		if isMemoryAccess(in.Op) {
			e.memoryAt(in, r.proven[pc], e.f.Sources[in.Source].Wasm.MemoryOffset(), r.operands[pc])
			continue
		}
		e.plain(in, r.proven[pc])
	}
}

// A constant needs no physical slot when the consuming instruction encodes it
// and normal execution overwrites that slot before any observation or exit.
type deadImmediate struct {
	in       slots.Instruction
	end      int
	original slots.Slot
	value    uint64
}

func (e *functionEmitter) planDeadImmediate(pc int) (deadImmediate, bool) {
	constant := e.f.Code[pc]
	if (constant.Op != 0x41 && constant.Op != 0x42) || pc+1 >= e.codeLimit() {
		return deadImmediate{}, false
	}
	in := e.f.Code[pc+1]
	if in.B != constant.Dst || in.A == constant.Dst {
		return deadImmediate{}, false
	}
	end, dst := pc+1, in.Dst
	if pc+2 < e.codeLimit() {
		next := e.f.Code[pc+2]
		local := slots.Slot(e.f.Sources[next.Source].Wasm.A)
		if (next.Op == 0x21 || next.Op == 0x22) && next.A == in.Dst {
			if _, ok := e.registers[local]; ok {
				end, dst = pc+2, local
			}
		}
	}
	if constant.Dst != dst && !e.overwrittenBeforeExit(end+1, constant.Dst) {
		return deadImmediate{}, false
	}
	forwarded := in
	forwarded.Dst = dst
	v := e.f.Sources[constant.Source].Wasm.Bits
	return deadImmediate{forwarded, end, in.Dst, v}, true
}

func (e *functionEmitter) combineDeadImmediate(pc int) (int, bool) {
	p, ok := e.planDeadImmediate(pc)
	if !ok {
		return 0, false
	}
	if !e.registerImmediate(p.in.Op, p.in.Dst, p.in.A, p.value) && !e.pooledInteger(p.in, p.value) {
		return 0, false
	}
	if p.in.Dst != p.original && !e.overwrittenBeforeExit(p.end+1, p.original) {
		e.move(p.original, p.in.Dst)
	}
	if p.end != pc+1 && !e.overwrittenBeforeExit(p.end+1, e.f.Code[p.end].Dst) {
		e.move(e.f.Code[p.end].Dst, p.in.Dst)
	}
	return p.end, true
}

// Both memory proofs and guarded heap-cache hints use the same i32 transfer.
// The caller owns CFG/helper boundaries. Pass the existing instruction by
// address rather than creating an aggregate argument for each compiler step.
func updateI32Bounds(f *slots.Function, in *slots.Instruction, values map[slots.Slot]boundedValue) {
	src := f.Sources[in.Source].Wasm
	a, ok := values[in.A]
	if !ok {
		a = boundedValue{max: math.MaxUint32}
	}
	b, ok := values[in.B]
	if !ok {
		b = boundedValue{max: math.MaxUint32}
	}
	v := boundedValue{max: math.MaxUint32}
	switch in.Op {
	case 0x20:
		v, ok = values[slots.Slot(src.A)]
		if !ok {
			v = boundedValue{max: math.MaxUint32}
		}
	case 0x21, 0x22:
		v = a
		values[slots.Slot(src.A)] = v
	case 0x41:
		v = boundedValue{uint64(uint32(src.Bits)), true}
	case 0x6a:
		if a.max+b.max <= math.MaxUint32 {
			v = boundedValue{a.max + b.max, a.constant && b.constant}
		}
	case 0x6c:
		if a.max*b.max <= math.MaxUint32 {
			v = boundedValue{a.max * b.max, a.constant && b.constant}
		}
	case 0x71:
		if a.constant && b.constant {
			v = boundedValue{a.max & b.max, true}
		} else {
			v.max = min(a.max, b.max)
		}
	case 0x74:
		if b.constant {
			n := b.max & 31
			if a.max <= math.MaxUint32>>n {
				v = boundedValue{a.max << n, a.constant}
			}
		}
	case 0x76:
		if b.constant {
			v = boundedValue{a.max >> (b.max & 31), a.constant}
		}
	case 0x2d:
		v.max = 255
	case 0x2f:
		v.max = 65535
	case 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f:
		v.max = 1
	}
	if in.Dst != slots.NoSlot {
		values[in.Dst] = v
	}
}
