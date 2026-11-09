package codegen

import (
	"math"
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
					e.prepareHeapCaches(Config{})
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

// Equal-frequency types can have very different hit rates. An invariant array
// root must win the last register pair over freshly loaded object identities.
func TestARMHeapCachePrefersInvariantIdentityWithinTheCallerBudget(t *testing.T) {
	for _, mode := range []string{"immutable-global", "local", "mutable-global", "mixed-roots", "changing-local", "move", "join", "loop-call", "handler", "prompt", "no-loop"} {
		for _, occupied := range []int{4, 6} {
			f := &slots.Function{}
			e := functionEmitter{a: &assembler{arch: "arm64"}, f: f, unmetered: true, heapAccesses: map[uint32]heapAccess{}, registers: map[slots.Slot]byte{}}
			for i := 1; i <= occupied; i++ {
				e.registers[slots.Slot(i)] = byte(i)
			}
			cfg := Config{Globals: []module.GlobalType{{Value: wasm.AnyRef}, {Value: wasm.AnyRef}}}
			emit := func(op uint32, dst, a, b slots.Slot, index uint32) {
				source := uint32(len(f.Sources))
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: index}})
				f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
				if op == 0xfb000b {
					e.heapAccesses[source] = heapAccess{array: true, cacheClass: 1}
				} else if op == 0xfb0002 {
					e.heapAccesses[source] = heapAccess{cacheClass: 0}
				}
			}
			for pass := range 2 {
				op, index := uint32(0x23), uint32(0)
				if mode == "local" || mode == "changing-local" || mode == "move" {
					op = 0x20
				}
				if mode == "mixed-roots" {
					index = uint32(pass)
				}
				emit(op, 10, slots.NoSlot, slots.NoSlot, index)
				emit(0xfb000b, 11, 10, slots.NoSlot, 0)
				emit(0xfb0002, 12, 11, slots.NoSlot, 0)
			}
			switch mode {
			case "mutable-global":
				cfg.Globals[0].Mutable = true
			case "changing-local":
				emit(0x21, slots.NoSlot, 11, slots.NoSlot, 0)
			case "move":
				f.Moves = []slots.Move{{Dst: 0, Src: 11}}
			case "join":
				f.Edges = append(f.Edges, slots.Edge{PC: 1})
			case "loop-call":
				emit(0x10, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
			case "handler":
				f.Handlers = []slots.Handler{{}}
			case "prompt":
				f.Prompts = map[uint32][]slots.Catch{0: {}}
			}
			emit(slots.ControlBase+uint32(ir.Jump), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
			f.Code[len(f.Code)-1].Extra = uint32(len(f.Edges))
			f.Code[len(f.Code)-1].Count = 1
			f.Edges = append(f.Edges, slots.Edge{PC: 0})
			if mode == "no-loop" {
				f.Code[len(f.Code)-1].Count = 0
				f.Edges = nil
			}
			e.prepareHeapCaches(cfg)
			first := uint64(0)
			if mode == "immutable-global" || mode == "local" {
				first = 1
			}
			if len(e.heapCaches) != (8-occupied)/2 || e.heapCaches[first] != (heapCache{byte(occupied + 1), byte(occupied + 2)}) {
				t.Fatal(mode, occupied, "identity priority or budget", e.heapCaches)
			}
			for _, cache := range e.heapCaches {
				if cache.key <= byte(occupied) || cache.address > 8 {
					t.Fatal("caller budget overlap", e.heapCaches)
				}
			}
		}
	}
}
