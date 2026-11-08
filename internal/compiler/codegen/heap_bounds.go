package codegen

import (
	"math"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// These are profitability hints, not permission to trust a frame's index.
// Every hit checks the actual unsigned index against the cached length bound;
// cold entries and unexpected values retain the complete ordinary checks.
func (e *functionEmitter) prepareHeapIndexBounds() {
	if !e.unmetered {
		return
	}
	arrays := false
	for _, access := range e.heapAccesses {
		arrays = arrays || access.array && !access.length
	}
	if !arrays {
		return
	}
	starts := make([]bool, len(e.f.Code))
	for _, edge := range e.f.Edges {
		starts[edge.PC] = true
	}
	values := map[slots.Slot]boundedValue{}
	bounds := map[uint64]uint32{}
	for pc, in := range e.f.Code {
		if starts[pc] || pc%256 == 0 {
			clear(values)
		}
		access, heap := e.heapAccesses[in.Source]
		heap = heap && in.Op == e.f.Sources[in.Source].Wasm.Op
		if heap && access.array && !access.length {
			bound := uint32(math.MaxUint32)
			if value, ok := values[in.B]; ok {
				bound = uint32(value.max)
			}
			bounds[access.cacheClass] = max(bounds[access.cacheClass], bound)
		}
		updateI32Bounds(e.f, &e.f.Code[pc], values)
		if !plainOpcode(in.Op) && !heap {
			clear(values)
		}
	}
	for class, bound := range bounds {
		if bound == math.MaxUint32 {
			delete(bounds, class)
		}
	}
	if len(bounds) != 0 {
		e.heapIndexBounds = bounds
	}
}

func (e *functionEmitter) arrayCacheHit(in slots.Instruction, cache heapCache, bound uint32, ready int) {
	miss := e.a.label()
	e.test(true)
	e.branch(miss, "eq")
	e.registerCompareFlags(e.operandRegister(0), cache.key, 0x51)
	e.branch(miss, "ne")
	e.slotLoad(in.B, 0, false, false)
	if !e.registerCompareImmediateFlags(e.operandRegister(0), 0x49, uint64(bound)) {
		e.constant(1, uint64(bound))
		e.compare(false)
	}
	// Moves preserve the index comparison flags. Branch directly to the hit
	// continuation rather than adding a miss branch and another hot jump.
	e.registerMove(e.operandRegister(2), cache.address, true)
	e.branch(ready, "ls")
	e.a.mark(miss)
	// A failed index hint may have replaced the reference operand.
	e.slotLoad(in.A, 0, false, true)
}
