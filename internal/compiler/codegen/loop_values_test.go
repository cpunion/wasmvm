package codegen

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLoopValueVersionsObserveDeadWritesAndEveryFinalWord(t *testing.T) {
	for _, kind := range []string{"disjoint", "live-inputs", "dead-write", "same-instruction", "different-final"} {
		v := loopValues{first: 1, current: map[slots.Slot]int{}, bySlot: map[slots.Slot][]int{}}
		switch kind {
		case "disjoint":
			v.read(1, 1)
			v.write(0, v.fresh(), 2)
			v.write(1, v.read(0, 3), 3)
		case "live-inputs":
			v.read(0, 1)
			v.read(1, 1)
			v.write(1, v.read(0, 2), 2)
		case "dead-write":
			v.read(0, 1)
			v.write(1, v.fresh(), 2) // Never read, but its emitted write can clobber a register.
			v.read(0, 3)
			v.write(1, v.read(0, 4), 4)
		case "same-instruction":
			v.read(0, 1)
			v.write(1, v.fresh(), 1)
			v.write(0, v.read(1, 2), 2)
		case "different-final":
			v.read(0, 1)
			v.read(1, 1)
		}
		for _, i := range v.current {
			v.versions[i].last = 20
		}
		work := 100
		interferes, complete := v.interfere(0, 1, &work)
		if !complete || interferes != (kind != "disjoint") {
			t.Fatal(kind, interferes, complete, v)
		}
		if kind == "disjoint" {
			work = 1
			if _, complete := v.interfere(0, 1, &work); complete {
				t.Fatal("interference budget ignored")
			}
		}
	}
}

func coalescedValueLoopFixture(literals bool) *slots.Function {
	f := &slots.Function{LocalCount: 16, Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.F32, wasm.F32, wasm.F64, wasm.F64, wasm.V128, wasm.V128, wasm.ExternRef, wasm.ExternRef, wasm.I32, wasm.I32, wasm.I32, wasm.I32}}
	emit := func(op uint32, dst, a, b slots.Slot, local uint32, bits uint64) {
		s := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Bits: bits, Offset: uint64(100 + s)}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Dst: dst, A: a, B: b, C: slots.NoSlot, Source: s})
	}
	emit(1, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
	emit(0x36, slots.NoSlot, 0, 3, 0, 0)
	emit(0x41, 12, slots.NoSlot, slots.NoSlot, 0, 1)
	emit(0x6a, 2, 1, 12, 0, 0)
	emit(0x21, slots.NoSlot, 2, slots.NoSlot, 3, 0)
	for _, pair := range []struct {
		store, load    uint32
		source, target slots.Slot
		offset         uint32
	}{{0x38, 0x2a, 4, 5, 8}, {0x39, 0x2b, 6, 7, 16}, {0xfd000b, 0xfd0000, 8, 9, 32}} {
		emit(pair.store, slots.NoSlot, 0, pair.target, 0, 0)
		f.Sources[len(f.Sources)-1].Wasm.B = pair.offset
		op := pair.load
		if literals {
			op = map[uint32]uint32{0x2a: 0x43, 0x2b: 0x44, 0xfd0000: 0xfd000c}[op]
		}
		emit(op, pair.source, 0, slots.NoSlot, 0, 0x80000000)
		if op == pair.load {
			f.Sources[len(f.Sources)-1].Wasm.B = pair.offset
		}
		if op == 0x44 {
			f.Sources[len(f.Sources)-1].Wasm.Bits = 0x8000000000000000
		}
		if op == 0xfd000c {
			binary.LittleEndian.PutUint64(f.Sources[len(f.Sources)-1].Wasm.Vector[:8], 0x0123456789abcdef)
			binary.LittleEndian.PutUint64(f.Sources[len(f.Sources)-1].Wasm.Vector[8:], 0xfedcba9876543210)
		}
		emit(0x21, slots.NoSlot, pair.source, slots.NoSlot, uint32(pair.target), 0)
	}
	emit(0xd1, 15, 11, slots.NoSlot, 0, 0)
	emit(0xd0, 10, slots.NoSlot, slots.NoSlot, 0, 0)
	emit(0x21, slots.NoSlot, 10, slots.NoSlot, 11, 0)
	emit(0x21, slots.NoSlot, 2, slots.NoSlot, 1, 0)
	emit(0x41, 13, slots.NoSlot, slots.NoSlot, 0, 4)
	emit(0x47, 14, 1, 13, 0, 0)
	emit(slots.ControlBase+uint32(ir.If), slots.NoSlot, 14, slots.NoSlot, 0, 0)
	f.Code[len(f.Code)-1].Count = 2
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
	f.Code[len(f.Code)-1].Count = uint32(len(f.Slots))
	f.Edges = []slots.Edge{{PC: 1}, {PC: uint32(len(f.Code) - 1)}}
	for s := range f.Slots {
		f.Args = append(f.Args, slots.Slot(s))
	}
	return f
}

func coalescedArithmeticLoopFixture(reverse, flags, always bool) *slots.Function {
	f := coalescedValueLoopFixture(true)
	for pc := range f.Code {
		in := &f.Code[pc]
		if in.Op == 0x36 || in.Op == 0x38 || in.Op == 0x39 || in.Op == 0xfd000b {
			in.Op, in.A, in.B = 0x1a, in.B, slots.NoSlot
			f.Sources[in.Source].Wasm.Op = 0x1a
		}
		if reverse && in.Op == 0x47 {
			in.Op = 0x46
			f.Sources[in.Source].Wasm.Op = 0x46
		}
	}
	// Shared inputs remain eligible for shifted-operand selection. The final
	// frame still observes every intermediate, so overwrite the two temporaries.
	var rotations []slots.Instruction
	emit := func(op uint32, dst, a, b slots.Slot, bits uint64) {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: bits, Offset: uint64(1000 + source)}})
		rotations = append(rotations, slots.Instruction{Op: op, Dst: dst, A: a, B: b, C: slots.NoSlot, Source: source})
	}
	emit(0x41, 12, slots.NoSlot, slots.NoSlot, 13)
	emit(0x78, 13, 2, 12, 0)
	emit(0x73, 14, 2, 13, 0)
	emit(0x41, 12, slots.NoSlot, slots.NoSlot, 19)
	emit(0x78, 13, 2, 12, 0)
	emit(0x73, 14, 14, 13, 0)
	emit(0x73, 14, 1, 2, 0)
	emit(0x41, 12, slots.NoSlot, slots.NoSlot, 5)
	emit(0x78, 14, 14, 12, 0)
	emit(0x41, 12, slots.NoSlot, slots.NoSlot, 0)
	pc := len(f.Code) - 4
	f.Code = append(append(append([]slots.Instruction(nil), f.Code[:pc]...), rotations...), f.Code[pc:]...)
	for i := range f.Edges {
		if f.Edges[i].PC > 1 {
			f.Edges[i].PC += uint32(len(rotations))
		}
	}
	if reverse {
		f.Edges[0], f.Edges[1] = f.Edges[1], f.Edges[0]
	}
	if !flags {
		pc := len(f.Code) - 2
		branch, ret := f.Code[pc], f.Code[pc+1]
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: 14, Offset: 999}})
		copy := slots.Instruction{Op: 0x20, Dst: 15, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Source: source}
		branch.A = 15
		f.Code = append(append([]slots.Instruction(nil), f.Code[:pc]...), copy, branch, ret)
		for i := range f.Edges {
			if f.Edges[i].PC > 1 {
				f.Edges[i].PC++
			}
		}
	}
	if always {
		branch := &f.Code[len(f.Code)-2]
		branch.Op = slots.ControlBase + uint32(ir.Jump)
		branch.Count = 1
		branch.A = slots.NoSlot
		f.Sources[branch.Source].Wasm.Op = branch.Op
		f.Edges = []slots.Edge{{PC: 1}}
	}
	return f
}

func coalescedNonresidentLoopFixture(literals bool) *slots.Function {
	f := coalescedValueLoopFixture(literals)
	var code []slots.Instruction
	for _, in := range f.Code {
		code = append(code, in)
		if in.Op != 0x21 || in.A != 2 && in.A != 4 && in.A != 6 && in.A != 8 && in.A != 10 || f.Sources[in.Source].Wasm.A != uint32(in.A)+1 {
			continue
		}
		// More equal final words than either target's entire register bank.
		// Their old values are dead, while the original target remains live-in.
		for i := 0; i < 24; i++ {
			dst := slots.Slot(len(f.Slots))
			f.Slots = append(f.Slots, f.Slots[in.A])
			f.Args = append(f.Args, dst)
			source := uint32(len(f.Sources))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x21, A: uint32(dst), Offset: uint64(2000 + source)}})
			code = append(code, slots.Instruction{Op: 0x21, Dst: slots.NoSlot, A: in.A, B: slots.NoSlot, C: slots.NoSlot, Source: source})
		}
	}
	f.Code = code
	f.LocalCount = uint32(len(f.Slots))
	f.Code[len(f.Code)-1].Count = uint32(len(f.Args))
	f.Edges[1].PC = uint32(len(f.Code) - 1)
	return f
}

func TestLoopValueCoalescingBudgetsAndUnsafeObservations(t *testing.T) {
	for _, kind := range []string{"accept", "precise", "header-poll", "trap", "unproved", "budget", "no-mapping"} {
		f := coalescedValueLoopFixture(false)
		a := assembler{arch: "arm64"}
		e := functionEmitter{a: &a, f: f, unmetered: true, usesMemory: true, memoryMinimum: 65536}
		regions := e.memoryRegions()
		r, ok := regions[1]
		if !ok || r.loop == nil {
			t.Fatal("fixture is not a guarded loop")
		}
		e.prepareRegisters(regions)
		start, work := 1, 100000
		switch kind {
		case "precise":
			e.unmetered = false
		case "header-poll":
			start = 0
		case "trap":
			f.Code[3].Op = 0x6d
		case "unproved":
			delete(r.proven, 1)
		case "budget":
			work = 1
		case "no-mapping":
			e.useRegisterPlan(nil)
		}
		old := e.currentRegisterPlan()
		full, entry, merges := e.coalesceLoopRegisters(e.loopValueVersions(start, r), work)
		if e.currentRegisterPlan() != old {
			t.Fatal("analysis mutated emitter state")
		}
		if kind == "accept" {
			if full == nil || entry == nil || merges < 5 {
				t.Fatal("all value classes not coalesced", merges)
			}
			shared := map[uint16]bool{}
			for _, s := range entry.slots {
				kind, reg := registerLocation(entry, s)
				key := uint16(kind)<<8 | uint16(reg)
				if shared[key] {
					t.Fatal("entry loaded different original words into one register")
				}
				shared[key] = true
			}
		} else if full != nil || entry != nil || merges != 0 {
			t.Fatal(kind, "unsafe coalescing")
		}
	}
}

func TestLoopValueCoalescingIncludesNonresidentEntryValues(t *testing.T) {
	for _, test := range []struct {
		source, target slots.Slot
		kind           byte
	}{{2, 3, 1}, {4, 5, 2}, {6, 7, 2}, {8, 9, 3}, {10, 11, 1}} {
		f := coalescedValueLoopFixture(false)
		e := functionEmitter{a: &assembler{arch: "arm64"}, f: f, unmetered: true, usesMemory: true, memoryMinimum: 65536}
		r := e.memoryRegions()[1]
		old := &registerPlan{homes: map[slots.Slot]byte{}, floats: map[slots.Slot]byte{}, vectors: map[slots.Slot]byte{}, slots: []slots.Slot{test.source}}
		setRegisterHome(old, test.source, test.kind, 4)
		e.useRegisterPlan(old)
		full, entry, merges := e.coalesceLoopRegisters(e.loopValueVersions(1, r), 100000)
		if full == nil || entry == nil || merges != 1 {
			t.Fatal(test, "nonresident value not coalesced", merges)
		}
		for _, s := range []slots.Slot{test.source, test.target} {
			if kind, reg := registerLocation(full, s); kind != test.kind || reg != 4 {
				t.Fatal(test, s, "did not reuse resident home", kind, reg)
			}
		}
		// The target's original value is read before the source's first write.
		// Transition must load that nonresident live-in, not the old cached word.
		if len(entry.slots) != 1 || entry.slots[0] != test.target {
			t.Fatal(test, "wrong live-in representative", entry.slots)
		}
		if kind, reg := registerLocation(entry, test.target); kind != test.kind || reg != 4 {
			t.Fatal(test, "entry did not load original nonresident bits")
		}
		if e.currentRegisterPlan() != old || len(old.slots) != 1 {
			t.Fatal("analysis modified the incoming plan")
		}
	}
	f := coalescedValueLoopFixture(false)
	e := functionEmitter{a: &assembler{arch: "arm64"}, f: f, unmetered: true, usesMemory: true, memoryMinimum: 65536}
	// The address never joins any copied-value group. Uncached groups must
	// remain in the frame when none has a resident member to supply a home.
	e.useRegisterPlan(&registerPlan{homes: map[slots.Slot]byte{0: 4}, slots: []slots.Slot{0}})
	if full, entry, merges := e.coalesceLoopRegisters(e.loopValueVersions(1, e.memoryRegions()[1]), 100000); full != nil || entry != nil || merges != 0 {
		t.Fatal("memory-only groups acquired registers")
	}
}

// Replace only the private-backedge control observation, keeping every code
// offset/relocation intact. This forces the generated exit after the first full
// iteration instead of letting the C adapter cancel before guest execution.
func forcePrivateLoopPoll(t *testing.T, c *CompiledFunction, arch string) []byte {
	t.Helper()
	code := append([]byte(nil), c.Entry.Code...)
	if len(c.PrivatePollOffsets) != 1 {
		t.Fatal(arch, "private poll sites", c.PrivatePollOffsets)
	}
	pc := int(c.PrivatePollOffsets[0])
	if arch == "arm64" {
		load := uint32(0xf9400009) | uint32(nativeabi.ControlOffset/8)<<10
		if binary.LittleEndian.Uint32(code[pc:]) != load || binary.LittleEndian.Uint32(code[pc+4:])&0xff00001f != 0xb5000009 {
			t.Fatal("private poll metadata did not identify the actual control load")
		}
		binary.LittleEndian.PutUint32(code[pc:], 0xd2800029)
	} else {
		pattern := []byte{0x48, 0x83, 0xbf}
		pattern = binary.LittleEndian.AppendUint32(pattern, uint32(nativeabi.ControlOffset))
		pattern = append(pattern, 0)
		if !bytes.Equal(code[pc:pc+8], pattern) || code[pc+8] != 0x0f || code[pc+9] != 0x85 {
			t.Fatal("private poll metadata did not identify the actual control comparison")
		}
		copy(code[pc:pc+8], []byte{0x48, 0x83, 0xff, 1, 0x90, 0x90, 0x90, 0x90})
	}
	return code
}

func TestCoalescedLoopEveryPCFullFramesMemoryAndGeneratedPoll(t *testing.T) {
	type loopCase struct {
		name                                         string
		literals, arithmetic, reverse, flags, always bool
	}
	cases := []loopCase{{name: "memory"}, {name: "literals", literals: true}, {name: "nonresident"}, {name: "nonresident", literals: true}, {name: "invariant-literals", literals: true}}
	for _, reverse := range []bool{false, true} {
		for _, flags := range []bool{false, true} {
			cases = append(cases, loopCase{"arithmetic", true, true, reverse, flags, false})
		}
	}
	cases = append(cases, loopCase{"unconditional", true, true, false, true, true})
	for _, test := range cases {
		f := coalescedValueLoopFixture(test.literals)
		if test.arithmetic {
			f = coalescedArithmeticLoopFixture(test.reverse, test.flags, test.always)
		}
		if test.name == "nonresident" {
			f = coalescedNonresidentLoopFixture(test.literals)
		}
		if test.name == "invariant-literals" {
			f = invariantLiteralLoopFixture()
		}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			if err != nil {
				t.Fatal(arch, err)
			}
			if c.CoalescedSlotGroups < 5 {
				t.Fatal(arch, "value classes not exercised", c.CoalescedSlotGroups)
			}
			if test.name == "nonresident" && c.CoalescedSlotGroups < 120 {
				t.Fatal(arch, "nonresident value classes not exercised", c.CoalescedSlotGroups)
			}
			if test.name == "invariant-literals" && c.HoistedLiterals < 6 {
				t.Fatal(arch, "literal classes not exercised", c.HoistedLiterals)
			}
			if c.NativeLoops != 1 || test.arithmetic && (c.UsesMemory || c.LoopGuards != 0) {
				t.Fatal(arch, "loop classification", test.name, c.NativeLoops, c.UsesMemory, c.LoopGuards)
			}
			if test.arithmetic && test.flags && arch == "arm64" && c.FusedIntegers < 2 {
				t.Fatal("shared input prevented shifted-operand selection", c.FusedIntegers)
			}
			if test.arithmetic && test.flags && arch == "amd64" && c.FusedIntegers == 0 {
				t.Fatal("native loop lost logical/shift selection")
			}
			forcePrivateLoopPoll(t, c, arch)
		}
		if !native.Available {
			continue
		}
		compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
		_, reference, ref := functionContext(t, f, false)
		patched, err := native.Publish(forcePrivateLoopPoll(t, compiled, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(patched.Close)
		pollLease, err := patched.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pollLease.Close)
		memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
		initialMemory := bytes.Repeat([]byte{0xa5}, 65536)
		for _, address := range []uint64{0, 65536 - 48, 65536 - 47, 0xffffffff, 1 << 32} {
			if test.arithmetic && address != 0 {
				continue
			}
			for start := range f.Code {
				for _, poll := range []bool{false, true} {
					if test.always && !poll {
						continue
					}
					if !test.always && poll && (start != 0 || address != 0 && address != 65536-48) {
						continue
					}
					memory.Write(0, initialMemory)
					refMemory.Write(0, initialMemory)
					for s, typ := range f.Slots {
						v := value.Value{Lo: 0x12345678 + uint64(s)}
						switch typ {
						case wasm.F32:
							v.Lo = 0x7fc12345
						case wasm.F64:
							v.Lo = 0xfff0000000000001
						case wasm.V128:
							v.Lo, v.Hi = 0x0123456789abcdef, 0xfedcba9876543210
						}
						switch s {
						case 0:
							v.Lo = address
						case 1:
							v.Lo = 0
						case 2, 12, 14:
							v.Lo = 1
						case 13:
							v.Lo = 4
						}
						if test.reverse && (s == 14 || s == 15) {
							v.Lo = 0
						}
						writeSlot(t, ctx, 16, slots.Slot(s), v)
						writeSlot(t, ref, 16, slots.Slot(s), v)
					}
					state := native.State{Used: 16, Depth: 1, Continuation: uint64(start), Budget: 10000}
					entry := lease
					if poll {
						state.Budget = uint64(len(f.Code) - 2)
						if test.always {
							branch := len(f.Code) - 2
							if start <= 1 {
								state.Budget = uint64(branch - start)
							} else {
								state.Budget = uint64(2*branch - start)
							}
						}
						entry = pollLease
					}
					ctx.SetState(state)
					ref.SetState(state)
					got, err := ctx.Enter(entry, 0)
					want, werr := ref.Enter(reference, 0)
					gs, gerr := ctx.State()
					ws, serr := ref.State()
					if err != nil || werr != nil || gerr != nil || serr != nil || got != want || gs.Continuation != ws.Continuation || gs.Payload[0] != ws.Payload[0] || got == native.Trap && gs.Source != ws.Source {
						t.Fatal(test.name, test.reverse, test.flags, address, start, poll, got, want, gs, ws, err, werr, gerr, serr)
					}
					for s := range f.Slots {
						if a, b := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); a != b {
							t.Fatal(test.name, test.reverse, test.flags, address, start, poll, "slot", s, a, b)
						}
					}
					actual, expected := make([]byte, 65536), make([]byte, 65536)
					memory.Read(0, actual)
					refMemory.Read(0, expected)
					if !bytes.Equal(actual, expected) {
						t.Fatal(test.name, address, start, poll, "memory effects differ")
					}
				}
			}
		}
	}
}

func TestVectorAliasesForgetEverySharedPhysicalSource(t *testing.T) {
	f := &slots.Function{Slots: []wasm.ValueType{wasm.V128, wasm.V128, wasm.V128}}
	e := functionEmitter{f: f, vectorRegisters: map[slots.Slot]byte{0: 8, 1: 8, 2: 9}, vectorAliases: map[slots.Slot]slots.Slot{2: 0}}
	e.forgetVectorAliases(1)
	if len(e.vectorAliases) != 0 {
		t.Fatal("alias survived a shared source-register write")
	}
}
