package codegen

import (
	"math"
	"slices"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHeapCachePriorityAndCallerRegisterBudget(t *testing.T) {
	for _, tc := range []struct {
		name   string
		counts map[uint64]int
		want   []uint64
	}{
		{"single-use", map[uint64]int{0: 1, 9: 1}, nil},
		{"one-class", map[uint64]int{17: 2}, []uint64{17}},
		{"equal-frequency", map[uint64]int{9: 3, 0: 3, 1: 3}, []uint64{0, 1}},
		{"weighted", map[uint64]int{0: 4, 8: 6, math.MaxUint64: 6, 3: 1}, []uint64{8, math.MaxUint64}},
	} {
		for occupied := 0; occupied <= 8; occupied++ {
			for _, arch := range []string{"arm64", "amd64"} {
				for _, unmetered := range []bool{false, true} {
					f := &slots.Function{}
					e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: unmetered,
						heapAccesses: map[uint32]heapAccess{}, registers: map[slots.Slot]byte{}}
					for r := 1; r <= occupied; r++ {
						e.registers[slots.Slot(r)] = byte(r)
					}
					// Input order and Go map order must not change equal-cost choices.
					for class, count := range tc.counts {
						for i := 0; i < count; i++ {
							source := uint32(len(f.Code))
							f.Code = append(f.Code, slots.Instruction{Op: 0xfb0002, Source: source})
							e.heapAccesses[source] = heapAccess{cacheClass: class}
						}
					}
					weights := e.heapCachePrioritiesARM64(instructionFrequencies(f))
					var candidates []registerCandidate
					for class, weight := range weights {
						candidates = append(candidates, registerCandidate{kind: 3, key: class, weight: weight})
					}
					slices.SortFunc(candidates, compareRegisterCandidates)
					var used [32]bool
					for _, r := range e.registers {
						used[r] = true
					}
					for _, candidate := range candidates {
						e.selectHeapCache(candidate.key, []byte{1, 2, 3, 4, 5, 6, 7, 8}, &used)
					}
					e.prepareHeapCaches()
					want := min(len(tc.want), (8-occupied)/2, 2)
					if arch != "arm64" || !unmetered {
						want = 0
					}
					if len(e.heapCaches) != want {
						t.Fatal(tc.name, arch, unmetered, occupied, "cache count", e.heapCaches, want)
					}
					for i, class := range tc.want[:want] {
						cache, ok := e.heapCaches[class]
						first := byte(occupied + 1 + i*2)
						if !ok || cache != (heapCache{first, first + 1}) {
							t.Fatal(tc.name, occupied, "priority or register assignment", class, cache, ok)
						}
					}
				}
			}
		}
	}
}

func TestARMHeapCacheCompetesWithLiteralsInTheSharedCallerBudget(t *testing.T) {
	f, config := heapFixture(module.FieldType{Value: wasm.I32}, false, false)
	config.HeapTypes = append(config.HeapTypes, config.HeapTypes[0])
	ret := f.Code[len(f.Code)-1]
	f.Code = f.Code[:len(f.Code)-1]
	for _, in := range slices.Clone(f.Code) {
		src := f.Sources[in.Source]
		src.Wasm.A = 1
		in.Source = uint32(len(f.Sources))
		f.Sources = append(f.Sources, src)
		f.Code = append(f.Code, in)
	}
	// Independent, expensive literals would otherwise consume every caller
	// word before either checked cache was considered.
	for i := range 8 {
		s := slots.Slot(len(f.Slots))
		f.Slots = append(f.Slots, wasm.I64)
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x42, Bits: 0x123456789abcdef0 + uint64(i)}})
		f.Code = append(f.Code, slots.Instruction{Op: 0x42, Source: source, Dst: s, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	f.Code = append(f.Code, ret)
	addHeapFixtureLoop(f, 0)
	e := functionEmitter{a: &assembler{arch: "arm64"}, f: f, unmetered: true}
	e.prepareHeapAccesses(config)
	e.selectRegisters(config, nil)
	if len(e.heapCaches) != 2 || len(e.constantRegisters) == 0 || len(e.constantRegisters) >= 8 {
		t.Fatal("cache/literal opportunity cost was not considered", e.heapCaches, e.constantRegisters)
	}
	used := map[byte]bool{}
	claim := func(r byte, caller bool) {
		if used[r] || caller && (r < 1 || r > 8) {
			t.Fatal("overlapping or out-of-budget register", r)
		}
		used[r] = true
	}
	for _, r := range e.registers {
		claim(r, false)
	}
	for _, r := range e.constantRegisters {
		claim(r, true)
	}
	for _, r := range e.runtimeRegisters {
		claim(r, true)
	}
	for _, cache := range e.heapCaches {
		claim(cache.key, true)
		claim(cache.address, true)
	}
	// Reset code is stable despite randomized Go map iteration, and neither
	// pair survives an independent native entry.
	var first []byte
	for range 8 {
		e.a.code = nil
		e.prepareHeapCaches()
		if first == nil {
			first = slices.Clone(e.a.code)
		} else if !slices.Equal(first, e.a.code) {
			t.Fatal("non-deterministic cache reset")
		}
	}
	if len(first) != 8 {
		t.Fatal("both cache keys must reset", first)
	}
}
