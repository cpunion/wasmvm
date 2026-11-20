package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

type structLoop struct {
	start     int
	reference slots.Slot
	class     uint64
	cache     heapCache
}

func (e *functionEmitter) isLoopStructAccess(in slots.Instruction, loop *structLoop) bool {
	a, ok := e.heapAccesses[in.Source]
	return ok && in.Op == e.f.Sources[in.Source].Wasm.Op && !a.array &&
		in.A == loop.reference && a.cacheClass == loop.class
}

func structLoopTerminator(op uint32) bool {
	return op == slots.ControlBase+uint32(ir.Jump) || op == slots.ControlBase+uint32(ir.If)
}

// Every independent entry can resolve the invariant handle afresh. The private
// backedge can keep the payload address for the duration of this native entry.
func (e *functionEmitter) addStructLoops(regions map[int]memoryRegion) map[int]memoryRegion {
	if len(e.heapCaches) == 0 {
		return regions
	}
	code := e.f.Code
	entries := make([]bool, len(code))
	occupied := make([]bool, len(code))
	for _, edge := range e.f.Edges {
		entries[edge.PC] = true
	}
	for start, region := range regions {
		for pc := start; pc <= region.end; pc++ {
			occupied[pc] = true
		}
	}
	for start := 0; start < len(code); {
		end := start
		for end < len(code) && !occupied[end] && (end == start || !entries[end] && end%256 != 0) {
			in := code[end]
			if a, ok := e.heapAccesses[in.Source]; ok && in.Op == e.f.Sources[in.Source].Wasm.Op && !a.array {
				end++
				continue
			}
			if !plainOpcode(in.Op) || scalarCanTrap(in.Op) || isMemoryAccess(in.Op) || in.Op == 0x3f {
				break
			}
			end++
		}
		first := start
		start = max(start+1, end)
		if end == first || !entries[first] || end >= len(code) || !structLoopTerminator(code[end].Op) {
			continue
		}
		r := memoryRegion{end: end - 1, heap: e.proveStructLoop(code[first:end])}
		if r.heap == nil {
			continue
		}
		r.heap.start = first
		r.loop = e.nativeLoop(first, r)
		if r.loop == nil {
			continue
		}
		if regions == nil {
			regions = map[int]memoryRegion{}
		}
		regions[first] = r
	}
	return regions
}

func (e *functionEmitter) proveStructLoop(code []slots.Instruction) *structLoop {
	var proof structLoop
	found := false
	for _, in := range code {
		if a, ok := e.heapAccesses[in.Source]; ok && in.Op == e.f.Sources[in.Source].Wasm.Op {
			cache, cached := e.heapCaches[a.cacheClass]
			if !cached || a.array || in.A >= slots.Slot(e.f.LocalCount) {
				return nil
			}
			if found && (proof.reference != in.A || proof.class != a.cacheClass) {
				return nil
			}
			proof = structLoop{reference: in.A, class: a.cacheClass, cache: cache}
			found = true
		}
	}
	if !found {
		return nil
	}
	for _, in := range code {
		// Self local.get retains the canonical handle. Other writes to the
		// slot would invalidate the proof even if they happen after all uses.
		local := slots.Slot(e.f.Sources[in.Source].Wasm.A)
		if in.Dst == proof.reference && !(in.Op == 0x20 && local == proof.reference) ||
			(in.Op == 0x21 || in.Op == 0x22) && local == proof.reference {
			return nil
		}
	}
	return &proof
}

func (e *functionEmitter) guardStructLoop(loop *structLoop, start int) {
	cold := e.a.label()
	e.memoryFallbacks = append(e.memoryFallbacks, memoryFallback{cold, start, e.currentRegisterPlan()})
	// Resolve touches only scratch registers until the guard succeeds.
	e.slotLoad(loop.reference, 0, false, true)
	e.resolveHeapHandle(false, cold, cold)
	e.registerMove(loop.cache.address, e.operandRegister(2), true)
	e.registerMove(loop.cache.key, e.operandRegister(0), true)
}
