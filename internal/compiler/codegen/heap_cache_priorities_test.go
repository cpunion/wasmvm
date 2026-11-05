package codegen

import (
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
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
