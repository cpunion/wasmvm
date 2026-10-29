package codegen

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func frequencyFixture(size int, edges map[int][]uint32) *slots.Function {
	f := &slots.Function{Code: make([]slots.Instruction, size)}
	for pc := range f.Code {
		in := &f.Code[pc]
		in.Op = 1
		if targets := edges[pc]; len(targets) != 0 {
			in.Op = slots.ControlBase + uint32(ir.Table)
			in.Extra = uint32(len(f.Edges))
			in.Count = uint32(len(targets))
			for _, target := range targets {
				f.Edges = append(f.Edges, slots.Edge{PC: target})
			}
		}
	}
	return f
}

func TestAllocationFrequencyHintsAreBoundedAndDeterministic(t *testing.T) {
	for _, test := range []struct {
		edges map[int][]uint32
		want  []uint64
	}{
		{nil, nil}, {map[int][]uint32{0: {11}}, nil},
		{map[int][]uint32{0: {0}}, []uint64{8, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		{map[int][]uint32{5: {2}, 8: {2}}, []uint64{1, 1, 8, 8, 8, 8, 8, 8, 8, 1, 1, 1}},
		{map[int][]uint32{6: {3}, 10: {1}}, []uint64{1, 8, 8, 64, 64, 64, 64, 8, 8, 8, 8, 1}},
		{map[int][]uint32{6: {1}, 9: {4}}, []uint64{1, 8, 8, 8, 64, 64, 64, 8, 8, 8, 1, 1}},
		{map[int][]uint32{4: {3}, 5: {2}, 6: {1}, 7: {0}}, []uint64{8, 64, 512, 512, 512, 512, 64, 8, 1, 1, 1, 1}},
	} {
		f := frequencyFixture(12, test.edges)
		for range 3 {
			if got := instructionFrequencies(f); !reflect.DeepEqual(got, test.want) {
				t.Fatal(test.edges, got, test.want)
			}
		}
	}
	// Shared edge arrays can otherwise turn a linear input into quadratic work.
	f := frequencyFixture(128, map[int][]uint32{0: make([]uint32, 128)})
	for i := range f.Code {
		f.Code[i] = f.Code[0]
	}
	if got := instructionFrequencies(f); got != nil {
		t.Fatal("analysis budget did not fall back", got)
	}
}

func TestLoopPrioritiesIncludeTuplesAndEdgeMoves(t *testing.T) {
	f := &slots.Function{Slots: make([]wasm.ValueType, 8), Args: []slots.Slot{3, 4, 7}, Moves: []slots.Move{{Dst: 5, Src: 6}}, Edges: []slots.Edge{{PC: 0, Count: 1}},
		Sources: []ir.Source{{Outputs: []wasm.ValueType{wasm.I64}}}, Code: []slots.Instruction{
			{Op: 0x10, Count: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.Jump), Count: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.Return), Extra: 2, Count: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}}
	got := slotPriorities(f, nil, instructionFrequencies(f))
	if !reflect.DeepEqual(got, []uint64{0, 0, 0, 8, 8, 8, 8, 1}) {
		t.Fatal("lost hot tuple/move", got)
	}
}

func allocationLoopFixture() (*slots.Function, Config, slots.Slot) {
	f := tableFixture(wasm.FuncRef, 0, 20)
	condition := slots.Slot(len(f.Slots))
	f.Slots = append(f.Slots, wasm.I32)
	branch := slots.Instruction{Op: slots.ControlBase + uint32(ir.If), Count: 2, Dst: slots.NoSlot, A: condition, B: slots.NoSlot, C: slots.NoSlot}
	f.Code = append(f.Code[:4:4], append([]slots.Instruction{branch}, f.Code[4:]...)...)
	f.Edges = []slots.Edge{{PC: 0}, {PC: 5}}
	return f, Config{Unmetered: true, FunctionCount: 3, TableCount: 2}, condition
}

func TestSharedAllocationKeepsHotMetadataAndOperandsAheadOfColdTail(t *testing.T) {
	f, config, condition := allocationLoopFixture()
	for _, arch := range []string{"amd64", "arm64"} {
		e := functionEmitter{a: &assembler{arch: arch, labels: map[int]int{}}, f: f, unmetered: true}
		e.allocateRegisters(config, nil)
		for _, slot := range []slots.Slot{0, 1, 2, 3, 4, condition} {
			if _, ok := e.registers[slot]; !ok {
				t.Fatal("cold tail displaced hot home", arch, slot, e.registers)
			}
		}
		used := map[byte]bool{}
		for _, r := range e.registers {
			if used[r] {
				t.Fatal("aliased mutable homes")
			}
			used[r] = true
		}
		for _, r := range e.runtimeRegisters {
			if used[r] || r < 1 || r > 8 {
				t.Fatal("overlapping or reserved metadata register", r)
			}
			used[r] = true
		}
		for _, r := range e.constantRegisters {
			if used[r] || r < 1 || r > 8 {
				t.Fatal("overlapping literal register", r)
			}
			used[r] = true
		}
		if arch == "arm64" {
			for _, offset := range []uint32{16, 24, 32, 40, 48} {
				if _, ok := e.runtimeRegisters[offset]; !ok {
					t.Fatal("hot runtime word spilled", offset, e.runtimeRegisters)
				}
			}
		}
		first, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			next, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			if err != nil || !bytes.Equal(first.Entry.Code, next.Entry.Code) {
				t.Fatal("nondeterministic shared allocation", arch, err)
			}
		}
		limits := DefaultLimits()
		limits.MaxCodeBytes = uint64(len(first.Entry.Code) - 1)
		if _, err := CompileFunctionWithConfig(arch, f, limits, config); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("allocation code budget", err)
		}
	}
}

func TestSharedAllocationPreservesAllColdFramesAndTableEffects(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	f, config, condition := allocationLoopFixture()
	_, fast, ctx := configuredContext(t, f, config)
	config.Unmetered = false
	_, checked, ref := configuredContext(t, f, config)
	memories := make([]*native.Memory, 2)
	buffers := make([][]uint64, 2)
	addresses := make([]uint64, 2)
	for i := range memories {
		m, err := native.NewMemory(72, 72)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(m.Close)
		memories[i] = m
		buffer, _ := m.Buffer()
		addresses[i] = uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
		buffers[i] = unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), 9)
		w := buffers[i]
		w[0], w[1], w[2] = 1, 2, 0x123456789abcdef0
		w[3], w[4], w[5], w[6] = 2, addresses[i]+56, 2, addresses[i]+56
	}
	for _, index := range []uint64{0, 1, 2, 0xffffffff, 0x100000000} {
		for pc := range f.Code {
			for i, c := range []*native.Context{ctx, ref} {
				for slot := range f.Slots {
					writeSlot(t, c, 16, slots.Slot(slot), value.Value{Lo: uint64(slot) + 99})
				}
				writeSlot(t, c, 16, 0, value.Value{Lo: index})
				writeSlot(t, c, 16, condition, value.Value{})
				buffers[i][7], buffers[i][8] = 0, 0xfedcba9876543210
				s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1024, Fuel: 7}
				s.Payload[2] = addresses[i]
				if err := c.SetState(s); err != nil {
					t.Fatal(err)
				}
			}
			exit, err := ctx.Enter(fast, 0)
			want, werr := ref.Enter(checked, 0)
			if err != nil || werr != nil || exit != want {
				t.Fatal(index, pc, exit, want, err, werr)
			}
			gotState, _ := ctx.State()
			wantState, _ := ref.State()
			if gotState.Source != wantState.Source || gotState.Payload[0] != wantState.Payload[0] || gotState.Continuation != wantState.Continuation || gotState.Fuel != wantState.Fuel {
				t.Fatal("cold state", index, pc, gotState, wantState)
			}
			for slot := range f.Slots {
				if got, want := readSlot(t, ctx, 16, slots.Slot(slot)), readSlot(t, ref, 16, slots.Slot(slot)); got != want {
					t.Fatal("cold frame", index, pc, slot, got, want)
				}
			}
			if !reflect.DeepEqual(buffers[0][7:], buffers[1][7:]) {
				t.Fatal("table effects", index, pc, buffers)
			}
		}
	}
}

func TestAllocationSharedTupleAndLeafBudgets(t *testing.T) {
	// Reusing one large tuple from every instruction must not multiply analysis
	// work by input length. The baseline count remains when the hint budget ends.
	f := &slots.Function{Slots: make([]wasm.ValueType, 4), Args: make([]slots.Slot, 64), Sources: []ir.Source{{}}, Code: make([]slots.Instruction, 64)}
	weights := make([]uint64, 64)
	for i := range f.Args {
		f.Args[i] = 3
		weights[i] = 8
		f.Code[i] = slots.Instruction{Op: 0x10, Count: 64, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
	}
	if got := slotPriorities(f, nil, weights)[3]; got != 64+16*64*7 {
		t.Fatal("shared tuple budget", got)
	}
	f.Args = nil
	f.Edges = []slots.Edge{{Count: 256}}
	f.Moves = make([]slots.Move, 256)
	for i := range f.Moves {
		f.Moves[i] = slots.Move{Dst: 2, Src: 3}
	}
	for i := range f.Code {
		f.Code[i].Op = slots.ControlBase + uint32(ir.Jump)
		f.Code[i].Count = 1
	}
	if got := slotPriorities(f, nil, weights); got[2] != 256+9*256*7 || got[3] != got[2] {
		t.Fatal("shared edge-move budget", got)
	}

	literal := uint64(0x123456789abcdef0)
	callee := &slots.Function{Signature: module.FuncType{Results: []wasm.ValueType{wasm.I64}}, Slots: []wasm.ValueType{wasm.I64}, Args: []slots.Slot{0},
		Sources: []ir.Source{{Wasm: module.Instr{Op: 0x42, Bits: literal}}}, Code: []slots.Instruction{
			{Op: 0x42, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
			{Op: slots.ControlBase + uint32(ir.Return), Count: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}}
	caller := &slots.Function{Sources: []ir.Source{{}}, Code: []slots.Instruction{{Op: 0x11}}}
	config := Config{DirectCallees: map[uint32]*slots.Function{0: callee, 1: callee, 2: callee, 3: callee, 4: callee}, IndirectCallees: map[uint32][]uint32{0: {0, 1, 2, 3, 4}}}
	e := functionEmitter{f: caller}
	if got := e.constantPriorities(config, nil)[literal]; got != 6 {
		t.Fatal("leaf literal work budget", got)
	}
}
