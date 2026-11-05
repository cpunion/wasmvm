package codegen

import (
	"encoding/binary"
	"math"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func splitValueFixture(memory bool) (*slots.Function, int) {
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.I32, wasm.V128}, Args: []slots.Slot{0}}
	emit := func(op uint32, dst, a, b slots.Slot, bits uint64) {
		source := uint32(len(f.Sources))
		wasm := module.Instr{Op: op, Bits: bits, Offset: uint64(1000 + source)}
		if op == 0xfd000c {
			for lane := 0; lane < 4; lane++ {
				binary.LittleEndian.PutUint32(wasm.Vector[lane*4:], uint32(bits))
			}
		}
		f.Sources = append(f.Sources, ir.Source{Wasm: wasm})
		f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: a, B: b, C: slots.NoSlot})
	}
	if !memory {
		f.Slots[4], f.Slots[5] = wasm.F32, wasm.F64
		emit(0x43, 4, slots.NoSlot, slots.NoSlot, uint64(math.Float32bits(1)))
		emit(0x44, 5, slots.NoSlot, slots.NoSlot, math.Float64bits(1))
		emit(0xfd000c, 6, slots.NoSlot, slots.NoSlot, 1)
	} else {
		emit(1, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	}
	start := len(f.Code)
	if memory {
		emit(0x41, 3, slots.NoSlot, slots.NoSlot, 0)
		emit(0x28, 5, 3, slots.NoSlot, 0)
	}
	for group := 0; group < 3; group++ {
		for i := 0; i < 8; i++ {
			typ, literal, add, one, bits := wasm.I32, uint32(0x41), uint32(0x6a), slots.Slot(2), uint64(i+1+group*8)
			if !memory {
				switch group {
				case 0:
					typ, literal, add, one, bits = wasm.F32, 0x43, 0x92, 4, uint64(math.Float32bits(float32(i+1)))
				case 1:
					typ, literal, add, one, bits = wasm.F64, 0x44, 0xa0, 5, math.Float64bits(float64(i+1))
				case 2:
					typ, literal, add, one = wasm.V128, 0xfd000c, 0xfd00ae, 6
				}
			}
			dst := slots.Slot(len(f.Slots))
			f.Slots = append(f.Slots, typ)
			emit(literal, dst, slots.NoSlot, slots.NoSlot, bits)
			for range 7 {
				emit(add, dst, dst, one, 0)
			}
		}
	}
	emit(0x6a, 0, 0, 2, 0)
	emit(0x49, 3, 0, 1, 0)
	branch := len(f.Code)
	emit(slots.ControlBase+uint32(ir.If), slots.NoSlot, 3, slots.NoSlot, 0)
	f.Code[branch].Count = 2
	f.Edges = []slots.Edge{{PC: uint32(start)}, {PC: uint32(branch + 1)}}
	if memory {
		emit(0x28, 5, 4, slots.NoSlot, 0) // Checked after the borrowed-length loop.
	}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	f.Code[len(f.Code)-1].Count = 1
	f.LocalCount = uint32(len(f.Slots))
	f.Signature = module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	return f, branch
}

func modelSplitValues(f *slots.Function, initial []value.Value, start, branch int, poll bool, memory []byte) ([]value.Value, native.Exit, uint64, uint64) {
	words := append([]value.Value(nil), initial...)
	pc := start
	for steps := 0; steps < len(f.Code)*10; steps++ {
		if poll && pc == branch {
			return words, native.Poll, uint64(pc), 0
		}
		in := f.Code[pc]
		src := f.Sources[in.Source].Wasm
		a, b := value.Value{}, value.Value{}
		if in.A != slots.NoSlot {
			a = words[in.A]
		}
		if in.B != slots.NoSlot {
			b = words[in.B]
		}
		switch in.Op {
		case 1:
		case 0x41, 0x43, 0x44:
			words[in.Dst] = value.Value{Lo: src.Bits}
		case 0xfd000c:
			words[in.Dst] = value.Value{Lo: binary.LittleEndian.Uint64(src.Vector[:8]), Hi: binary.LittleEndian.Uint64(src.Vector[8:])}
		case 0x6a:
			words[in.Dst] = value.Value{Lo: uint64(uint32(a.Lo) + uint32(b.Lo))}
		case 0x49:
			words[in.Dst] = value.Value{}
			if uint32(a.Lo) < uint32(b.Lo) {
				words[in.Dst].Lo = 1
			}
		case 0x92:
			words[in.Dst] = value.Value{Lo: uint64(math.Float32bits(math.Float32frombits(uint32(a.Lo)) + math.Float32frombits(uint32(b.Lo))))}
		case 0xa0:
			words[in.Dst] = value.Value{Lo: math.Float64bits(math.Float64frombits(a.Lo) + math.Float64frombits(b.Lo))}
		case 0xfd00ae:
			lo := uint64(uint32(a.Lo)+uint32(b.Lo)) | uint64(uint32(a.Lo>>32)+uint32(b.Lo>>32))<<32
			hi := uint64(uint32(a.Hi)+uint32(b.Hi)) | uint64(uint32(a.Hi>>32)+uint32(b.Hi>>32))<<32
			words[in.Dst] = value.Value{Lo: lo, Hi: hi}
		case 0x28:
			address := uint64(uint32(a.Lo))
			if address+4 > uint64(len(memory)) {
				return words, native.Trap, uint64(pc), uint64(wasm.TrapMemoryBounds)
			}
			words[in.Dst] = value.Value{Lo: uint64(binary.LittleEndian.Uint32(memory[address:]))}
		case slots.ControlBase + uint32(ir.If):
			edge := 1
			if uint32(a.Lo) != 0 {
				edge = 0
			}
			pc = int(f.Edges[edge].PC)
			continue
		case slots.ControlBase + uint32(ir.Return):
			return words, native.Return, uint64(pc), uint64(pc)
		default:
			panic("unmodeled split-value instruction")
		}
		pc++
	}
	panic("split-value fixture did not terminate")
}

func TestSplitValuePlansAllColdWordsAndPrivatePoll(t *testing.T) {
	for _, mode := range []struct{ memory, header bool }{{}, {header: true}, {memory: true}, {memory: true, header: true}} {
		memory := mode.memory
		f, branch := splitValueFixture(memory)
		if mode.header {
			f.Edges[0].PC = 0
		}
		config := Config{Unmetered: true, MemoryMinimumPages: 1}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
			if err != nil {
				t.Fatal(arch, memory, err)
			}
			if c.ValueLoops != 1 || c.ValuePhis == 0 || c.ValueVersions == 0 {
				var unsupported []uint32
				for _, in := range f.Code[3:branch] {
					if !plainOpcode(in.Op) || scalarCanTrap(in.Op) {
						unsupported = append(unsupported, in.Op)
					}
				}
				t.Fatal(arch, memory, "value splitting not exercised", c.NativeLoops, c.ValueLoops, c.ValuePhis, c.ValueVersions, "unsupported", unsupported, "branch", branch)
			}
			forcePrivateLoopPoll(t, c, arch)
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
		if memory {
			for _, target := range []*native.Context{ctx, ref} {
				m := bindMemory(t, target, 65536, 65536)
				if err := m.Write(0, data); err != nil {
					t.Fatal(err)
				}
			}
		}
		for _, address := range []uint64{0, 65532, 65533} {
			if !memory && address != 0 {
				continue
			}
			for start := range f.Code {
				for _, poll := range []bool{false, true} {
					if poll && start != 0 {
						continue
					}
					initial := make([]value.Value, len(f.Slots))
					for slot, typ := range f.Slots {
						initial[slot].Lo = uint64(20 + slot)
						if typ == wasm.F32 {
							initial[slot].Lo = uint64(math.Float32bits(float32(slot) / 4))
						} else if typ == wasm.F64 {
							initial[slot].Lo = math.Float64bits(float64(slot) / 4)
						} else if typ == wasm.V128 {
							initial[slot].Lo, initial[slot].Hi = 0x123456789abcdef, 0xfedcba9876543210
						}
					}
					initial[0].Lo, initial[1].Lo, initial[2].Lo = 0, 3, 1
					if memory {
						initial[4].Lo = address
					}
					want, exit, continuation, payload := modelSplitValues(f, initial, start, branch, poll, data)
					entries := []*native.Lease{lease, precise}
					contexts := []*native.Context{ctx, ref}
					if poll {
						entries, contexts = []*native.Lease{polled}, []*native.Context{ctx}
					}
					for i, entry := range entries {
						target := contexts[i]
						for slot, word := range initial {
							writeSlot(t, target, 16, slots.Slot(slot), word)
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
							t.Fatal(memory, address, start, poll, i, "exit", got, exit, state, continuation, payload, err, stateErr)
						}
						for slot, word := range want {
							if got := readSlot(t, target, 16, slots.Slot(slot)); got != word {
								t.Fatal(memory, address, start, poll, i, "word", slot, got, word)
							}
						}
						if got == native.Trap && state.Source != f.Sources[f.Code[continuation].Source].Wasm.Offset {
							t.Fatal("late checked memory source changed", state)
						}
					}
				}
			}
		}
	}
}
