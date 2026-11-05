package codegen

import (
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func frameCopyFixture(wide, memory, entryRead bool) (*slots.Function, int) {
	typ, add := wasm.I32, uint32(0x6a)
	if wide {
		typ, add = wasm.I64, 0x7c
	}
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, typ, wasm.I32, wasm.I32, wasm.I32, wasm.I32, typ}, Args: []slots.Slot{3}, Signature: module.FuncType{Results: []wasm.ValueType{typ}}}
	emit := func(op uint32, dst, a, b slots.Slot, local uint32, bits uint64) {
		s := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Bits: bits, Offset: uint64(1000 + s)}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: s, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	emit(1, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
	if memory {
		emit(0x41, 7, slots.NoSlot, slots.NoSlot, 0, 0)
		emit(0x28, 6, 7, slots.NoSlot, 0, 0)
	}
	for group := 0; group < 20; group++ {
		source := slots.Slot(len(f.Slots))
		target, anchor := source+1, source+2
		f.Slots = append(f.Slots, typ, typ, typ)
		if entryRead {
			emit(add, 3, 3, target, 0, 0)
		}
		if group%2 == 0 {
			emit(0x20, target, slots.NoSlot, slots.NoSlot, uint32(source), 0)
		} else {
			emit(0x21, slots.NoSlot, source, slots.NoSlot, uint32(target), 0)
		}
		// The old anchor is still used after the target is born. They cannot
		// share a fixed home, but the target can later read through this anchor.
		emit(add, 3, 3, anchor, 0, 0)
		emit(0x22, source, target, slots.NoSlot, uint32(anchor), 0)
		emit(add, source, source, 8, 0, 0)
		emit(add, 3, 3, target, 0, 0)
		if group%3 == 0 {
			emit(0x21, slots.NoSlot, source, slots.NoSlot, uint32(target), 0)
		}
	}
	emit(0x6a, 0, 0, 2, 0, 0)
	emit(0x41, 1, slots.NoSlot, slots.NoSlot, 0, 3)
	emit(0x47, 4, 0, 1, 0, 0)
	branch := len(f.Code)
	emit(slots.ControlBase+uint32(ir.If), slots.NoSlot, 4, slots.NoSlot, 0, 0)
	f.Code[branch].Count = 2
	f.Edges = []slots.Edge{{PC: 1}, {PC: uint32(branch + 1)}}
	if memory {
		emit(0x28, 6, 5, slots.NoSlot, 0, 0)
	}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
	f.Code[len(f.Code)-1].Count = 1
	f.LocalCount = uint32(len(f.Slots))
	return f, branch
}

func modelFrameCopies(f *slots.Function, initial []value.Value, start, branch int, poll bool, memory []byte) ([]value.Value, native.Exit, uint64, uint64) {
	words := append([]value.Value(nil), initial...)
	for pc, steps := start, 0; steps < len(f.Code)*10; steps++ {
		if poll && pc == branch {
			return words, native.Poll, uint64(pc), 0
		}
		in := f.Code[pc]
		src := f.Sources[in.Source].Wasm
		a, b := uint64(0), uint64(0)
		if in.A != slots.NoSlot {
			a = words[in.A].Lo
		}
		if in.B != slots.NoSlot {
			b = words[in.B].Lo
		}
		switch in.Op {
		case 1:
		case 0x20:
			words[in.Dst].Lo = words[src.A].Lo
		case 0x21, 0x22:
			words[src.A].Lo = a
			if in.Dst != slots.NoSlot {
				words[in.Dst].Lo = a
			}
		case 0x41:
			words[in.Dst].Lo = uint64(uint32(src.Bits))
		case 0x6a, 0x6b, 0x6c, 0x71, 0x72, 0x73, 0x7c, 0x7d, 0x7e, 0x83, 0x84, 0x85:
			words[in.Dst].Lo, _ = numeric.Eval(in.Op, a, b)
		case 0x47:
			words[in.Dst].Lo = 0
			if uint32(a) != uint32(b) {
				words[in.Dst].Lo = 1
			}
		case 0x28:
			address := uint64(uint32(a))
			if address+4 > uint64(len(memory)) {
				return words, native.Trap, uint64(pc), uint64(wasm.TrapMemoryBounds)
			}
			words[in.Dst].Lo = uint64(binary.LittleEndian.Uint32(memory[address:]))
		case slots.ControlBase + uint32(ir.If):
			edge := 1
			if uint32(a) != 0 {
				edge = 0
			}
			pc = int(f.Edges[edge].PC)
			continue
		case slots.ControlBase + uint32(ir.Return):
			return words, native.Return, uint64(pc), uint64(pc)
		default:
			panic("unmodeled frame-copy instruction")
		}
		pc++
	}
	panic("frame-copy fixture did not terminate")
}

func TestLoopFrameCopiesCompleteWordsColdPCsPollAndLateTrap(t *testing.T) {
	for _, mode := range []struct{ wide, memory, header, entryRead, rawHigh bool }{{}, {wide: true}, {memory: true}, {header: true}, {entryRead: true}, {wide: true, rawHigh: true}, {memory: true, rawHigh: true}} {
		f, branch := frameCopyFixture(mode.wide, mode.memory, mode.entryRead)
		if mode.header {
			f.Edges[0].PC = 0
		}
		config := Config{Unmetered: true, MemoryMinimumPages: 1}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			if err != nil {
				t.Fatal(arch, mode, err)
			}
			if (c.DeferredFrameCopies != 0) != (!mode.header && !mode.entryRead) {
				t.Fatal(arch, mode, "wrong reconstruction selection", c.DeferredFrameCopies, c.NativeLoops, c.ValueLoops)
			}
		}
		if !native.Available {
			continue
		}
		c, lease, ctx := configuredContext(t, f, config)
		_, precise, ref := functionContext(t, f, false)
		code, err := native.Publish(forcePrivateLoopPoll(t, c, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(code.Close)
		polled, err := code.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(polled.Close)
		data := make([]byte, 65536)
		for i := range data {
			data[i] = byte(i*17 + 3)
		}
		if mode.memory {
			for _, target := range []*native.Context{ctx, ref} {
				if err := bindMemory(t, target, 65536, 65536).Write(0, data); err != nil {
					t.Fatal(err)
				}
			}
		}
		for _, address := range []uint64{0, 65532, 65533} {
			if !mode.memory && address != 0 {
				continue
			}
			for start := range f.Code {
				for _, poll := range []bool{false, true} {
					if poll && start != 0 {
						continue
					}
					initial := make([]value.Value, len(f.Slots))
					for s := range initial {
						initial[s].Lo = 0x1234567800000000 | uint64(s+1)
						if mode.rawHigh {
							initial[s].Hi = 0x89abcdef00000000 | uint64(s+1)
						}
					}
					initial[0].Lo, initial[2].Lo, initial[8].Lo, initial[5].Lo = 0, 1, 1, address
					want, exit, continuation, payload := modelFrameCopies(f, initial, start, branch, poll, data)
					entries, contexts := []*native.Lease{lease, precise}, []*native.Context{ctx, ref}
					if mode.rawHigh {
						entries, contexts = entries[:1], contexts[:1]
					}
					if poll {
						entries, contexts = []*native.Lease{polled}, []*native.Context{ctx}
					}
					for i, entry := range entries {
						target := contexts[i]
						for s, word := range initial {
							writeSlot(t, target, 16, slots.Slot(s), word)
						}
						state := native.State{Used: 16, Depth: 1, Continuation: uint64(start), Fuel: 1 << 20, Budget: 1 << 20}
						if i == 1 {
							state.Payload[3] = 1
						}
						if err := target.SetState(state); err != nil {
							t.Fatal(err)
						}
						got, err := target.Enter(entry, 0)
						state, stateErr := target.State()
						if err != nil || stateErr != nil || got != exit || got != native.Trap && state.Continuation != continuation || state.Payload[0] != payload {
							t.Fatal(mode, address, start, poll, i, got, exit, state, continuation, payload, err, stateErr)
						}
						for s, word := range want {
							if got := readSlot(t, target, 16, slots.Slot(s)); got != word {
								t.Fatal(mode, address, start, poll, i, "word", s, got, word)
							}
						}
						if got == native.Trap && state.Source != f.Sources[f.Code[continuation].Source].Wasm.Offset {
							t.Fatal("late trap source changed", state)
						}
					}
				}
			}
		}
	}
}

func TestLoopFrameCopiesBoundedIndependentAnchors(t *testing.T) {
	f, branch := frameCopyFixture(false, false, false)
	e := &functionEmitter{a: &assembler{arch: "arm64"}, f: f, unmetered: true}
	r := memoryRegion{end: branch - 1, loop: &nativeLoop{}}
	v := e.loopValueVersions(1, r)
	if e.planLoopFrameCopies(v, r, 0) != nil {
		t.Fatal("copy reconstruction ignored its work budget")
	}
	p := e.planLoopFrameCopies(v, r, 1<<20)
	if p == nil {
		t.Fatal("missing changing-anchor reconstruction")
	}
	for _, copy := range p.final {
		if copy.src == copy.dst || p.deferred[copy.src] {
			t.Fatal("reconstruction depends on an unmaterialized frame word")
		}
	}
	for pc := 1; pc < branch; pc++ {
		i := pc - p.first
		if p.next[i] <= pc {
			t.Fatal("selector boundary did not advance")
		}
	}
	// Destroy the only original before creating its replacement anchor. A
	// final matching copy cannot justify forgetting the intervening value.
	for pc := 1; pc+1 < branch; pc++ {
		if f.Code[pc].Op == 0x22 {
			f.Code[pc], f.Code[pc+1] = f.Code[pc+1], f.Code[pc]
			pc++
		}
	}
	v = e.loopValueVersions(1, r)
	if e.planLoopFrameCopies(v, r, 1<<20) != nil {
		t.Fatal("copy reconstruction crossed an anchorless interval")
	}
}
