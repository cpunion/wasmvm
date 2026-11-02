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
	loop      *memoryLoop
}

func (e *functionEmitter) codeLimit() int {
	if e.regionLimit != 0 {
		return e.regionLimit
	}
	return len(e.f.Code)
}

// A bounded local abstract interpretation, never a speculative loop invariant.
// Facts start unknown at every CFG entry, helper, growth and poll boundary. Each
// region has at most 256 instructions, so compiler work stays linear in input.
func (e *functionEmitter) memoryRegions() map[int]memoryRegion {
	if !e.unmetered || !e.usesMemory || e.memoryMinimum == 0 || e.memoryViews {
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
			a, b := get(in.A), get(in.B)
			v := boundedValue{max: math.MaxUint32}
			if isMemoryAccess(in.Op) {
				_, size := memoryAccess(in.Op)
				if a.max+uint64(src.B)+uint64(size) <= e.memoryMinimum {
					proven[pc] = true
				} else {
					candidates++
				}
			}
			switch in.Op {
			case 0x20:
				v = get(slots.Slot(src.A))
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
		var guards map[int]memoryGuard
		var addresses map[int]memoryAddress
		var guarded uint32
		if candidates >= 3 {
			guards, addresses, guarded = e.memoryGuards(start, end, proven)
		}
		if len(proven) != 0 {
			r := memoryRegion{end: end - 1, proven: proven, guards: guards, guarded: guarded, addresses: addresses}
			if len(guards) != 0 {
				r.elided = e.deadAddressTemps(start, end, r)
			}
			regions[start] = r
		}
		start = end
	}
	for start := 0; start < len(e.f.Code); start++ {
		if r, ok := regions[start]; ok {
			r.loop = e.memoryLoop(start, r)
			regions[start] = r
		}
	}
	return regions
}

// Only uninterrupted normal entry consumes these proofs. Every interior PC is
// emitted again on a cold path, with all bounds checks, until the next region.
// Thus direct resumption never inherits facts from instructions it did not run.
func (e *functionEmitter) emitMemoryRegion(start int, r memoryRegion, pcs []int, out *CompiledFunction) {
	out.GuardedAccesses += r.guarded
	if r.loop != nil {
		e.emitMemoryLoop(start, r, pcs, out)
		return
	}
	out.DeadScalarOps += uint32(len(r.elided))
	e.emitMemoryRegionRange(start, r.end, r, pcs, out)
}

func (e *functionEmitter) emitMemoryRegionRange(start, end int, r memoryRegion, pcs []int, out *CompiledFunction) {
	e.regionLimit = end + 1
	e.regionEnd = r.end + 1
	e.provenPCs = r.proven
	e.guardPCs = r.guards
	e.addresses, e.elidedPCs = r.addresses, r.elided
	e.vectorAliases = map[slots.Slot]slots.Slot{}
	defer func() {
		e.regionLimit = 0
		e.regionEnd = 0
		e.provenPCs = nil
		e.guardPCs = nil
		e.vectorAliases = nil
		e.addresses = nil
		e.elidedPCs = nil
	}()
	var boundaries []int
	for pc := start; pc <= end; pc++ {
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
		if guard, ok := r.guards[pc]; ok {
			cold := pcs[pc]
			if pc == start {
				cold = e.a.label()
				e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, pc})
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
		if address, ok := r.addresses[pc]; ok {
			in.A = address.local
			e.memoryAt(in, true, address.offset)
			continue
		}
		if end, ok := e.combineVectorAffine(pc); ok {
			out.FusedVectors++
			pc = end
			continue
		}
		if end, ok := e.combineMaskShift(pc); ok {
			out.FusedIntegers++
			pc = end
			continue
		}
		if end, ok := e.combineShiftedInteger(pc); ok {
			out.FusedIntegers++
			pc = end
			continue
		}
		if end, ok := e.combineDeadImmediate(pc); ok {
			out.ForwardedScalars++
			pc = end
			continue
		}
		if pc > start {
			prev := e.f.Code[pc-1]
			if (prev.Op == 0x41 || prev.Op == 0x42) && prev.Dst == in.B && in.B != slots.NoSlot && e.registerImmediate(in.Op, in.Dst, in.A, e.f.Sources[prev.Source].Wasm.Bits) {
				continue
			}
		}
		if end, ok := e.combineScalarCopies(pc); ok {
			out.ForwardedScalars++
			pc = end
			continue
		}
		if end, ok := e.combineMultiplyAdd(pc); ok {
			out.FusedIntegers++
			pc = end
			continue
		}
		if e.forwardScalarResult(pc) {
			out.ForwardedScalars++
			pc++
			continue
		}
		e.plain(in, r.proven[pc])
	}
}

// A constant needs no physical slot when the consuming instruction encodes it
// and normal execution overwrites that slot before any observation or exit.
func (e *functionEmitter) combineDeadImmediate(pc int) (int, bool) {
	constant := e.f.Code[pc]
	if (constant.Op != 0x41 && constant.Op != 0x42) || pc+1 >= e.codeLimit() {
		return 0, false
	}
	in := e.f.Code[pc+1]
	if in.B != constant.Dst || in.A == constant.Dst {
		return 0, false
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
		return 0, false
	}
	forwarded := in
	forwarded.Dst = dst
	v := e.f.Sources[constant.Source].Wasm.Bits
	if !e.registerImmediate(in.Op, dst, in.A, v) && !e.pooledInteger(forwarded, v) {
		return 0, false
	}
	if dst != in.Dst && !e.overwrittenBeforeExit(end+1, in.Dst) {
		e.move(in.Dst, dst)
	}
	if end != pc+1 && !e.overwrittenBeforeExit(end+1, e.f.Code[end].Dst) {
		e.move(e.f.Code[end].Dst, dst)
	}
	return end, true
}
