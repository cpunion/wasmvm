package codegen

import (
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHeapCacheAMD64ReuseHintsAndConservativeBoundaries(t *testing.T) {
	for _, mode := range []string{"local", "same-slot-local", "copy", "immutable-global", "mutable-global", "mixed-roots", "changing-local", "join", "move", "select-same", "select-mixed", "loop-call", "handler", "prompt", "precise", "arm64", "no-loop"} {
		f := &slots.Function{Slots: make([]wasm.ValueType, 8)}
		e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, unmetered: true, heapAccesses: map[uint32]heapAccess{}}
		config := Config{Globals: []module.GlobalType{{Value: wasm.AnyRef}}}
		emit := func(op uint32, dst, a, b slots.Slot, local uint32) {
			source := uint32(len(f.Sources))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local}})
			f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
			if op == 0xfb0002 {
				e.heapAccesses[source] = heapAccess{cacheClass: 17}
			}
		}
		want := uint64(32)
		for pass := range 2 {
			if mode == "changing-local" {
				emit(0xfb000b, 0, 2, 3, 0)
			}
			op, local := uint32(0x20), uint32(0)
			if mode == "immutable-global" || mode == "mutable-global" {
				op = 0x23
			}
			if mode == "mixed-roots" {
				local = uint32(pass)
			}
			if mode == "same-slot-local" {
				emit(0x20, 0, slots.NoSlot, slots.NoSlot, 0)
			}
			emit(op, 4, slots.NoSlot, slots.NoSlot, local)
			if mode == "copy" {
				emit(0x22, 5, 4, slots.NoSlot, 6)
				emit(0x20, 4, slots.NoSlot, slots.NoSlot, 6)
			}
			if mode == "select-same" {
				emit(0x1b, 4, 4, 4, 0)
			}
			if mode == "select-mixed" {
				emit(0x1c, 4, 4, 1, 0)
			}
			emit(0xfb0002, 7, 4, slots.NoSlot, 0)
			if mode == "join" {
				f.Edges = append(f.Edges, slots.Edge{PC: uint32(len(f.Code) - 1)})
			}
		}
		switch mode {
		case "mutable-global":
			config.Globals[0].Mutable = true
			want = 0
		case "mixed-roots", "changing-local", "join", "select-mixed":
			want = 0
		case "move":
			f.Moves = []slots.Move{{Dst: 0, Src: 1}}
			want = 0
		case "loop-call":
			emit(0x10, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
			want = 0
		case "handler":
			f.Handlers = []slots.Handler{{}}
			want = 0
		case "prompt":
			f.Prompts = map[uint32][]slots.Catch{0: {}}
			want = 0
		case "precise":
			e.unmetered = false
			want = 0
		case "arm64":
			e.a.arch = "arm64"
			want = 0
		case "no-loop":
			want = 0
		}
		frequencies := make([]uint64, len(f.Code))
		for i := range frequencies {
			frequencies[i] = 8
		}
		if mode == "no-loop" {
			frequencies = nil
		}
		got := e.heapCachePrioritiesAMD64(config, frequencies)
		if got[17] != want || want == 0 && len(got) != 0 {
			t.Fatal(mode, got, want)
		}
	}
}

func TestHeapCacheAMD64PairBudgetAndReset(t *testing.T) {
	for occupied := 0; occupied <= 8; occupied++ {
		e := functionEmitter{a: &assembler{arch: "amd64"}, unmetered: true, registers: map[slots.Slot]byte{}}
		used := [32]bool{}
		pool := []byte{3, 12, 13, 14, 15, 9, 10, 11}
		for _, r := range pool[:occupied] {
			used[r] = true
		}
		available := pool
		e.selectHeapCacheAMD64(17, available, &used)
		if occupied > 6 {
			if len(e.heapCaches) != 0 {
				t.Fatal(occupied, e.heapCaches)
			}
			continue
		}
		want := heapCache{pool[occupied], pool[occupied+1]}
		if e.heapCaches[17] != want {
			t.Fatal(occupied, e.heapCaches, want)
		}
		e.selectHeapCacheAMD64(19, available, &used)
		if len(e.heapCaches) != 1 || !used[want.key] || !used[want.address] {
			t.Fatal(occupied, e.heapCaches, used)
		}
		for _, r := range e.integerRegisterPool() {
			if r == want.key || r == want.address {
				t.Fatal("cache overlaps slot bank", r)
			}
		}
		e.prepareHeapCaches()
		reference := assembler{arch: "amd64"}
		reference.xconst(want.key, 0)
		if !reflect.DeepEqual(e.a.code, reference.code) {
			t.Fatal("entry key not reset", e.a.code, reference.code)
		}
	}
}
