package codegen

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestImmediatePrioritiesPreserveObservationsAndBoundaries(t *testing.T) {
	for _, kind := range []string{"live", "dead", "trap", "guard", "alias", "poll", "region", "elided", "wide", "shift", "arm64", "metered"} {
		f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32}, Sources: []ir.Source{{Wasm: module.Instr{Bits: 17}}}, Code: []slots.Instruction{
			{Op: 0x41, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 0x6a, Dst: 2, A: 0, B: 1, C: slots.NoSlot},
		}}
		regions := map[int]memoryRegion{}
		arch, unmetered, discount := "amd64", true, uint64(7)
		switch kind {
		case "dead", "trap", "guard":
			f.Code = append(f.Code, slots.Instruction{Op: 0x28, Dst: 2, A: 0, B: slots.NoSlot, C: slots.NoSlot}, f.Code[0])
			r := memoryRegion{end: 3, proven: map[int]bool{2: true}}
			if kind == "dead" {
				discount = 14
			} else if kind == "trap" {
				r.proven = nil
			} else {
				r.guards = map[int]memoryGuard{3: {}}
			}
			regions[0] = r
		case "alias":
			f.Code[1].A = 1
			discount = 0
		case "poll":
			prefix := make([]slots.Instruction, 255)
			for i := range prefix {
				prefix[i] = slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
			}
			f.Code = append(prefix, f.Code...)
			discount = 0
		case "region":
			regions[1] = memoryRegion{end: 1}
			discount = 0
		case "elided":
			regions[0] = memoryRegion{end: 1, elided: map[int]bool{1: true}}
			discount = 0
		case "wide":
			f.Code[0].Op, f.Code[1].Op = 0x42, 0x7c
			f.Sources[0].Wasm.Bits = 0x100000000
			for i := range f.Slots {
				f.Slots[i] = wasm.I64
			}
			discount = 0
		case "shift":
			f.Code[1].Op = 0x74
		case "arm64":
			arch, discount = "arm64", 0
		case "metered":
			unmetered, discount = false, 0
		}
		frequencies := make([]uint64, len(f.Code))
		for i := range frequencies {
			frequencies[i] = 8
		}
		before := slotPriorities(f, regions, frequencies)
		after := append([]uint64(nil), before...)
		e := functionEmitter{f: f, a: &assembler{arch: arch}, unmetered: unmetered}
		e.immediateRegisterCosts(after, frequencies, regions)
		if before[1]-after[1] != discount || after[0] != before[0] || after[2] != before[2] {
			t.Fatal("observation cost", kind, before, after, discount)
		}
		cold := slotPriorities(f, regions, nil)
		want := append([]uint64(nil), cold...)
		e.immediateRegisterCosts(cold, nil, regions)
		if !reflect.DeepEqual(cold, want) {
			t.Fatal("changed cold credits", kind, cold, want)
		}
	}
}

func immediateMemoryFixture(t *testing.T) *slots.Function {
	body := module.Expr{
		{Op: 0x20}, {Op: 0x21, A: 5}, {Op: 0x03, Block: -64},
		{Op: 0x20, A: 2}, {Op: 0x41, Bits: 3}, {Op: 0x6c}, {Op: 0x41, Bits: 63}, {Op: 0x71}, {Op: 0x41, Bits: 4}, {Op: 0x6c}, {Op: 0x21, A: 3},
		{Op: 0x20, A: 3}, {Op: 0x28}, {Op: 0x20}, {Op: 0x73}, {Op: 0x41, Bits: 12345}, {Op: 0x6a}, {Op: 0x21, A: 4},
		{Op: 0x20, A: 3}, {Op: 0x20, A: 4}, {Op: 0x36},
		{Op: 0x20, A: 5}, {Op: 0x20, A: 4}, {Op: 0x6a}, {Op: 0x21, A: 5},
		{Op: 0x20, A: 2}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x22, A: 2}, {Op: 0x20, A: 1}, {Op: 0x49}, {Op: 0x0d},
		{Op: 0x0b}, {Op: 0x20, A: 5}, {Op: 0x0b},
	}
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, 4, body)
	m.Memories = []module.Limits{{Min: 1}}
	return functionModule(t, m).Functions[0]
}

func TestImmediatePrioritiesKeepRunningMemoryValue(t *testing.T) {
	f := immediateMemoryFixture(t)
	e := functionEmitter{f: f, a: &assembler{arch: "amd64"}, unmetered: true, usesMemory: true, memoryMinimum: 65536}
	regions := e.memoryRegions()
	e.allocateRegisters(Config{Unmetered: true, MemoryMinimumPages: 1}, regions)
	if _, ok := e.registers[4]; !ok {
		t.Fatal("running value displaced by folded constants", e.registers)
	}
	if !native.Available {
		return
	}
	_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	_, refLease, ref := functionContext(t, f, true)
	memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
	initial, gotBytes, wantBytes := make([]byte, 65536), make([]byte, 65536), make([]byte, 65536)
	for i := range initial {
		initial[i] = byte(i*151 + 29)
	}
	for pc := range f.Code {
		for _, seed := range []uint64{0, 0x80000000, 0xffffffff} {
			for _, count := range []uint64{1, 3, 4, 5, 8} {
				memory.Write(0, initial)
				refMemory.Write(0, initial)
				for s := range f.Slots {
					bits := uint64(s*17 + 5)
					if s == 0 {
						bits = seed
					} else if s == 1 {
						bits = count
					} else if s == 2 {
						bits = 0
					}
					writeSlot(t, ctx, 16, slots.Slot(s), value.Value{Lo: bits})
					writeSlot(t, ref, 16, slots.Slot(s), value.Value{Lo: bits})
				}
				state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
				ctx.SetState(state)
				ref.SetState(state)
				got, err := ctx.Enter(lease, 0)
				want, refErr := ref.Enter(refLease, 0)
				gs, _ := ctx.State()
				ws, _ := ref.State()
				if err != nil || refErr != nil || got != want || gs != ws {
					t.Fatal("priority state", pc, seed, count, got, want, gs, ws, err, refErr)
				}
				for s := range f.Slots {
					if got, want := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); got != want {
						t.Fatal("priority frame", pc, seed, count, s, got, want)
					}
				}
				memory.Read(0, gotBytes)
				refMemory.Read(0, wantBytes)
				if !bytes.Equal(gotBytes, wantBytes) {
					t.Fatal("priority memory", pc, seed, count)
				}
			}
		}
	}
}
