package codegen

import (
	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
)

// The private backedge is independent of an optional whole-loop memory proof.
// Pure arithmetic and already-proven accesses use the same emitter, mapping
// transitions, canonical exits and bounded polling contract.
type nativeLoop struct {
	yes, always, flags bool
	exit               int
	memory             *memoryLoop
}

func (e *functionEmitter) addNativeLoops(regions map[int]memoryRegion) map[int]memoryRegion {
	if !e.unmetered {
		return regions
	}
	starts := make([]bool, len(e.f.Code))
	starts[0] = true
	for _, edge := range e.f.Edges {
		starts[edge.PC] = true
	}
	for pc, in := range e.f.Code {
		if pc%256 == 0 {
			starts[pc] = true
		}
		if !plainOpcode(in.Op) && pc+1 < len(starts) {
			starts[pc+1] = true
		}
	}
	for start := 0; start < len(e.f.Code); {
		end := start
		for end < len(e.f.Code) && plainOpcode(e.f.Code[end].Op) && (end == start || !starts[end]) {
			end++
		}
		if end == start {
			start++
			continue
		}
		r, exists := regions[start]
		if !exists {
			r = memoryRegion{end: end - 1}
		}
		if r.loop == nil && r.unroll == nil && r.end == end-1 && len(r.guards) == 0 {
			if loop := e.nativeLoop(start, r); loop != nil {
				if regions == nil {
					regions = map[int]memoryRegion{}
				}
				r.loop = loop
				regions[start] = r
			}
		}
		start = end
	}
	return regions
}

func (e *functionEmitter) nativeLoop(start int, r memoryRegion) *nativeLoop {
	branchPC := r.end + 1
	if branchPC >= len(e.f.Code) || start/256 != branchPC/256 || r.end < start {
		return nil
	}
	branch := e.f.Code[branchPC]
	loop := &nativeLoop{exit: -1}
	if branch.Op == slots.ControlBase+uint32(ir.Jump) && branch.Count == 1 {
		edge := e.f.Edges[branch.Extra]
		if edge.PC != uint32(start) || edge.Count != 0 {
			return nil
		}
		loop.always = true
	} else if branch.Op == slots.ControlBase+uint32(ir.If) && branch.Count == 2 {
		yes, no := e.f.Edges[branch.Extra], e.f.Edges[branch.Extra+1]
		if yes.Count != 0 || no.Count != 0 {
			return nil
		}
		outside := func(pc uint32) bool { return int(pc) < start || int(pc) > branchPC }
		if yes.PC == uint32(start) && outside(no.PC) {
			loop.yes, loop.exit = true, int(no.PC)
		} else if no.PC == uint32(start) && outside(yes.PC) {
			loop.exit = int(yes.PC)
		} else {
			return nil
		}
		previous := e.f.Code[r.end]
		loop.flags = previous.Op >= 0x45 && previous.Op <= 0x5a && previous.Dst == branch.A
	} else {
		return nil
	}
	for pc := start; pc <= r.end; pc++ {
		in := e.f.Code[pc]
		if !plainOpcode(in.Op) || scalarCanTrap(in.Op) || in.Op == 0x3f || isMemoryAccess(in.Op) && !r.proven[pc] {
			return nil
		}
	}
	return loop
}
