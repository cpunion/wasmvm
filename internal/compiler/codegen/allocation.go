package codegen

import (
	"sort"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Backward CFG edges provide bounded priority hints, not correctness proofs.
// Merge latches with the same header; cap overlapping/nested interval weights.
// Pathological shared edge lists fall back to uniform priorities.
func instructionFrequencies(f *slots.Function) []uint64 {
	if len(f.Edges) == 0 {
		return nil
	}
	ends := make([]uint32, len(f.Code))
	work := uint64(len(f.Code)+len(f.Edges)) * 4
	found := false
	for pc, in := range f.Code {
		if in.Op < slots.ControlBase+uint32(ir.Jump) || in.Op > slots.ControlBase+uint32(ir.Table) {
			continue
		}
		if uint64(in.Count) > work {
			return nil
		}
		work -= uint64(in.Count)
		for _, edge := range f.Edges[in.Extra : in.Extra+in.Count] {
			if edge.PC <= uint32(pc) {
				ends[edge.PC] = max(ends[edge.PC], uint32(pc+1))
				found = true
			}
		}
	}
	if !found {
		return nil
	}
	delta := make([]int32, len(f.Code)+1)
	for start, end := range ends {
		if end != 0 {
			delta[start]++
			delta[end]--
		}
	}
	frequencies := make([]uint64, len(f.Code))
	depth := int32(0)
	for pc := range frequencies {
		depth += delta[pc]
		frequencies[pc] = uint64(1) << uint(min(depth, 3)*3)
	}
	return frequencies
}

func frequency(weights []uint64, pc int) uint64 {
	if weights == nil {
		return 1
	}
	return weights[pc]
}

func slotPriorities(f *slots.Function, regions map[int]memoryRegion, frequencies []uint64) []uint64 {
	weights := make([]uint64, len(f.Slots))
	use := func(s slots.Slot, w uint64) {
		if s != slots.NoSlot {
			weights[s] += w
		}
	}
	move := func(dst, src slots.Slot, w uint64) {
		if dst != slots.NoSlot && dst != src {
			use(dst, w)
			use(src, w)
		}
	}
	// Account for all tuples/moves once, including diagnostic-only entries.
	for _, s := range f.Args {
		use(s, 1)
	}
	for _, m := range f.Moves {
		use(m.Src, 1)
		use(m.Dst, 1)
	}
	work := uint64(len(f.Code)+len(f.Args)+len(f.Moves)+len(f.Edges)) * 8
	region := memoryRegion{end: -1}
	for pc, in := range f.Code {
		if next, ok := regions[pc]; ok {
			region = next
		} else if pc > region.end {
			region = memoryRegion{end: -1}
		}
		if region.elided[pc] {
			continue
		}
		w := frequency(frequencies, pc)
		if guard, ok := region.guards[pc]; ok {
			use(guard.address, w)
		}
		if address, ok := region.addresses[pc]; ok {
			in.A = address.local
		}
		if in.Op >= 0x20 && in.Op <= 0x22 {
			local := slots.Slot(f.Sources[in.Source].Wasm.A)
			if in.Op == 0x20 {
				move(in.Dst, local, w)
			} else {
				move(local, in.A, w)
				if in.Dst != local {
					move(in.Dst, in.A, w)
				}
			}
			continue
		}
		use(in.A, w)
		use(in.B, w)
		use(in.C, w)
		use(in.Dst, w)
		if w == 1 {
			continue
		}
		if in.Op >= 0x10 && in.Op <= 0x13 || in.Op == slots.ControlBase+uint32(ir.Return) {
			count := uint64(in.Count)
			if in.Op == 0x10 || in.Op == 0x11 {
				count += uint64(len(f.Sources[in.Source].Outputs))
			}
			if count <= work {
				work -= count
				for _, s := range f.Args[uint64(in.Extra) : uint64(in.Extra)+count] {
					use(s, w-1)
				}
			}
		} else if in.Op >= slots.ControlBase+uint32(ir.Jump) && in.Op <= slots.ControlBase+uint32(ir.Table) {
			if uint64(in.Count) > work {
				continue
			}
			work -= uint64(in.Count)
			for _, edge := range f.Edges[in.Extra : in.Extra+in.Count] {
				if uint64(edge.Count) > work {
					break
				}
				work -= uint64(edge.Count)
				for _, m := range f.Moves[edge.Start : edge.Start+edge.Count] {
					use(m.Src, w-1)
					use(m.Dst, w-1)
				}
			}
		}
	}
	return weights
}

type registerCandidate struct {
	kind        byte
	key, weight uint64
}

// All immutable integer words compete with scalar homes before the inventory
// fills. Constants/metadata remain in ARM caller registers; FP/vector homes use
// their separate shared physical pool. Every class is stable at every cold PC.
func (e *functionEmitter) allocateRegisters(config Config, regions map[int]memoryRegion) {
	frequencies := instructionFrequencies(e.f)
	weights := slotPriorities(e.f, regions, frequencies)
	var candidates []registerCandidate
	for s, w := range weights {
		if w != 0 {
			candidates = append(candidates, registerCandidate{0, uint64(s), w})
		}
	}
	if e.a.arch == "arm64" {
		for offset, w := range e.runtimePriorities(config, frequencies) {
			candidates = append(candidates, registerCandidate{1, uint64(offset), w})
		}
		for bits, w := range e.constantPriorities(config, frequencies) {
			candidates = append(candidates, registerCandidate{2, bits, w})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.weight != b.weight {
			return a.weight > b.weight
		}
		if a.kind != b.kind {
			return a.kind < b.kind
		}
		return a.key < b.key
	})
	available := e.integerRegisterPool()
	floatAvailable := []byte{3, 4, 5, 6, 7, 8, 9, 10}
	if e.a.arch == "arm64" {
		floatAvailable = []byte{8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	}
	integers, floats, vectors := map[slots.Slot]byte{}, map[slots.Slot]byte{}, map[slots.Slot]byte{}
	e.constantRegisters, e.runtimeRegisters = map[uint64]byte{}, map[uint32]byte{}
	var used [32]bool
	for _, candidate := range candidates {
		if candidate.kind == 0 {
			s := slots.Slot(candidate.key)
			t := e.f.Slots[s]
			if t == wasm.F32 || t == wasm.F64 || t == wasm.V128 {
				if len(floatAvailable) == 0 {
					continue
				}
				r := floatAvailable[0]
				floatAvailable = floatAvailable[1:]
				if t == wasm.V128 {
					e.vectorMemory(s, r, false)
					vectors[s] = r
				} else {
					e.slotLoad(s, 0, false, t == wasm.F64)
					e.floatFromInteger(r, e.operandRegister(0), t == wasm.F64)
					floats[s] = r
				}
				e.registerSlots = append(e.registerSlots, s)
				continue
			}
		}
		reg, ok := byte(0), false
		for _, r := range available {
			if !used[r] && (candidate.kind == 0 || r >= 1 && r <= 8) {
				reg, ok = r, true
				break
			}
		}
		if !ok {
			continue
		}
		used[reg] = true
		switch candidate.kind {
		case 0:
			s := slots.Slot(candidate.key)
			e.slotLoad(s, 0, false, true)
			e.registerMove(reg, e.operandRegister(0), true)
			integers[s] = reg
			e.registerSlots = append(e.registerSlots, s)
		case 1:
			e.runtimeRegisters[uint32(candidate.key)] = reg
		case 2:
			e.constantRegisters[candidate.key] = reg
		}
	}
	e.registers, e.floatRegisters, e.vectorRegisters = integers, floats, vectors
}
