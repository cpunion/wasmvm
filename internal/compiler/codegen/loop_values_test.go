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

func TestLoopValueCoalescingBudgetsAndUnsafeObservations(t *testing.T) {
	for _, kind := range []string{"accept", "precise", "header-poll", "trap", "unproved", "budget", "no-mapping"} {
		f := coalescedValueLoopFixture(false)
		a := assembler{arch: "arm64", labels: map[int]int{}}
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
			delete(r.loop.proven, 1)
		case "budget":
			work = 1
		case "no-mapping":
			e.useRegisterPlan(nil)
		}
		old := e.currentRegisterPlan()
		full, entry, merges := e.coalesceLoopRegisters(start, r, work)
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

// Replace only the private-backedge control observation, keeping every code
// offset/relocation intact. This forces the generated exit after the first full
// iteration instead of letting the C adapter cancel before guest execution.
func forcePrivateLoopPoll(t *testing.T, c *CompiledFunction, arch string) []byte {
	t.Helper()
	code := append([]byte(nil), c.Entry.Code...)
	found := 0
	// The original conditional PC has a later independent cold entry. Its
	// checked backedge is not the private loop whose aliased snapshot we test.
	limit := int(c.Offsets[len(c.Function.Code)-2])
	if arch == "arm64" {
		load := uint32(0xf9400009) | uint32(nativeabi.ControlOffset/8)<<10
		for pc := 0; pc+12 <= limit; pc += 4 {
			a, b, d := binary.LittleEndian.Uint32(code[pc:]), binary.LittleEndian.Uint32(code[pc+4:]), binary.LittleEndian.Uint32(code[pc+8:])
			if a == load && b&0xff00001f == 0xb5000009 && d&0xff000010 == 0x54000000 && d&0x00800000 != 0 {
				binary.LittleEndian.PutUint32(code[pc:], 0xd2800029)
				found++
			}
		}
	} else {
		pattern := []byte{0x48, 0x83, 0xbf}
		pattern = binary.LittleEndian.AppendUint32(pattern, uint32(nativeabi.ControlOffset))
		pattern = append(pattern, 0)
		for pc := 0; pc+24 <= limit; pc++ {
			if !bytes.Equal(code[pc:pc+8], pattern) || code[pc+8] != 0x0f || code[pc+9] != 0x85 {
				continue
			}
			for branch := pc + 14; branch < pc+23 && branch+6 <= len(code); branch++ {
				if code[branch] == 0x0f && (code[branch+1] == 0x84 || code[branch+1] == 0x85) && int32(binary.LittleEndian.Uint32(code[branch+2:])) < 0 {
					copy(code[pc:pc+8], []byte{0x48, 0x83, 0xff, 1, 0x90, 0x90, 0x90, 0x90})
					found++
					break
				}
			}
		}
	}
	if found != 1 {
		t.Fatal(arch, "private poll sites", found)
	}
	return code
}

func TestCoalescedLoopEveryPCFullFramesMemoryAndGeneratedPoll(t *testing.T) {
	for _, literals := range []bool{false, true} {
		f := coalescedValueLoopFixture(literals)
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			if err != nil {
				t.Fatal(arch, err)
			}
			if c.CoalescedSlotGroups < 5 {
				t.Fatal(arch, "value classes not exercised", c.CoalescedSlotGroups)
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
			for start := range f.Code {
				for _, poll := range []bool{false, true} {
					if poll && (start != 0 || address != 0 && address != 65536-48) {
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
						writeSlot(t, ctx, 16, slots.Slot(s), v)
						writeSlot(t, ref, 16, slots.Slot(s), v)
					}
					state := native.State{Used: 16, Depth: 1, Continuation: uint64(start), Budget: 10000}
					entry := lease
					if poll {
						state.Budget = uint64(len(f.Code) - 2)
						entry = pollLease
					}
					ctx.SetState(state)
					ref.SetState(state)
					got, err := ctx.Enter(entry, 0)
					want, werr := ref.Enter(reference, 0)
					gs, gerr := ctx.State()
					ws, serr := ref.State()
					if err != nil || werr != nil || gerr != nil || serr != nil || got != want || gs.Continuation != ws.Continuation || gs.Payload[0] != ws.Payload[0] || got == native.Trap && gs.Source != ws.Source {
						t.Fatal(literals, address, start, poll, got, want, gs, ws, err, werr, gerr, serr)
					}
					for s := range f.Slots {
						if a, b := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); a != b {
							t.Fatal(literals, address, start, poll, "slot", s, a, b)
						}
					}
					actual, expected := make([]byte, 65536), make([]byte, 65536)
					memory.Read(0, actual)
					refMemory.Read(0, expected)
					if !bytes.Equal(actual, expected) {
						t.Fatal(literals, address, start, poll, "memory effects differ")
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
