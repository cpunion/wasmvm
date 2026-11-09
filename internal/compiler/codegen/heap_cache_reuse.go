package codegen

import "github.com/cpunion/wasmvm/internal/compiler/slots"

// Reuse hints affect profitability only. Every hit still compares the complete
// handle, and every miss retains all existing validity checks. Keep the hint
// bounded and conservative at joins, tuple writes and loop calls.
func (e *functionEmitter) heapCacheReusePriorities(config Config, frequencies []uint64) map[uint64]uint64 {
	if !e.unmetered || len(e.heapAccesses) == 0 || frequencies == nil || len(e.f.Handlers) != 0 || len(e.f.Prompts) != 0 {
		return nil
	}
	written := map[slots.Slot]bool{}
	for pc, in := range e.f.Code {
		if frequency(frequencies, pc) <= 1 {
			continue
		}
		if in.Op >= 0x10 && in.Op <= 0x15 {
			return nil
		}
		if in.Dst != slots.NoSlot && !(in.Op == 0x20 && in.Dst == slots.Slot(e.f.Sources[in.Source].Wasm.A)) {
			written[in.Dst] = true
		}
		if in.Op == 0x21 || in.Op == 0x22 {
			written[slots.Slot(e.f.Sources[in.Source].Wasm.A)] = true
		}
	}
	for _, move := range e.f.Moves {
		written[move.Dst] = true
	}
	starts := make([]bool, len(e.f.Code))
	for _, edge := range e.f.Edges {
		starts[edge.PC] = true
	}
	origins := map[slots.Slot]uint64{}
	classes := map[uint64]uint64{}
	weights := map[uint64]uint64{}
	unstable := map[uint64]bool{}
	root := func(s slots.Slot) uint64 {
		if s == slots.NoSlot {
			return 0
		}
		if v := origins[s]; v != 0 {
			return v
		}
		if !written[s] {
			return uint64(s) + 1
		}
		return 0
	}
	for pc, in := range e.f.Code {
		if starts[pc] || pc%256 == 0 {
			clear(origins)
		}
		if access, ok := e.heapAccesses[in.Source]; ok && in.Op>>16 == 0xfb && frequency(frequencies, pc) > 1 {
			value := root(in.A)
			class := access.cacheClass
			if value == 0 || classes[class] != 0 && classes[class] != value {
				unstable[class] = true
			}
			classes[class] = value
			weights[class] += 2 * frequency(frequencies, pc)
		}
		value := uint64(0)
		src := e.f.Sources[in.Source].Wasm
		switch in.Op {
		case 0x20:
			value = root(slots.Slot(src.A))
		case 0x21, 0x22:
			value = root(in.A)
			origins[slots.Slot(src.A)] = value
		case 0x23:
			if uint64(src.A) < uint64(len(config.Globals)) && !config.Globals[src.A].Mutable {
				value = 1<<32 | uint64(src.A)
			}
		case 0x1b, 0x1c:
			if a, b := root(in.A), root(in.B); a == b {
				value = a
			}
		}
		if in.Dst != slots.NoSlot {
			origins[in.Dst] = value
		}
	}
	for class := range unstable {
		delete(weights, class)
	}
	return weights
}
