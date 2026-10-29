package codegen

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	enginememory "github.com/cpunion/wasmvm/internal/engine/memory"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func vectorMemoryFunction(t *testing.T, op, offset uint32, lane byte, bounded, pressure bool) (*slots.Function, int) {
	load, _ := memoryAccess(op)
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.V128}}
	if load {
		sig.Results = []wasm.ValueType{wasm.V128}
	}
	body := module.Expr{{Op: 0x20}}
	if bounded {
		body = append(body, module.Instr{Op: 0x41, Bits: 8191}, module.Instr{Op: 0x71})
	}
	if !load || op >= 0xfd0054 && op <= 0xfd0057 {
		body = append(body, module.Instr{Op: 0x20, A: 1})
	}
	body = append(body, module.Instr{Op: op, B: offset, Lane: lane})
	if pressure {
		// Keep all parameter vectors observable and more frequent than operands.
		for i := uint32(2); i < 14; i++ {
			sig.Params = append(sig.Params, wasm.V128)
			for j := 0; j < 3; j++ {
				body = append(body, module.Instr{Op: 0x20, A: i}, module.Instr{Op: 0x1a})
			}
		}
	}
	body = append(body, module.Instr{Op: 0xb})
	m := functionFixture(sig, 0, body)
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
	f := functionModule(t, m).Functions[0]
	for pc, in := range f.Code {
		if in.Op == op {
			return f, pc
		}
	}
	t.Fatal("missing vector memory operation")
	return nil, 0
}

func TestNativeVectorMemoryBoundsLanesAndSpills(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	initial := make([]byte, 65536)
	for i := range initial {
		initial[i] = byte((i*193 + 127) ^ (i >> 8))
	}
	for op := uint32(0xfd0000); op <= 0xfd005d; op++ {
		if !nativeVectorMemoryOpcode(op) {
			continue
		}
		load, size := memoryAccess(op)
		lanes := uint32(1)
		if op >= 0xfd0054 && op <= 0xfd005b {
			lanes = 16 / size
		}
		for lane := uint32(0); lane < lanes; lane++ {
			for _, offset := range []uint32{0, 1, 16, 4095, 4096, 8193, 65535, 0xffffffff} {
				for mode := 0; mode < 4; mode++ {
					fast, bounded, pressure := mode != 0, mode >= 2, mode == 3
					f, pc := vectorMemoryFunction(t, op, offset, byte(lane), bounded, pressure)
					in := f.Code[pc]
					config := Config{Unmetered: fast}
					if bounded {
						config.MemoryMinimumPages = 1
					}
					compiled, lease, c := configuredContext(t, f, config)
					if compiled.NativeVectors != 1 || compiled.Helpers != 0 || !compiled.UsesMemory {
						t.Fatal("vector access escaped", op, compiled)
					}
					if bounded && uint64(offset)+8191+uint64(size) <= 65536 && compiled.BoundsElided != 1 {
						t.Fatal("missing bounded vector proof", op, compiled.BoundsElided)
					}
					mem := bindMemory(t, c, 65536, 131072)
					ref, err := enginememory.New(module.Limits{Min: 1}, 2)
					if err != nil {
						t.Fatal(err)
					}
					for _, address := range []uint64{0, 1, 9, 65536 - uint64(size), 65537 - uint64(size), 65536, 0xffffffff, 1 << 32} {
						if err := mem.Write(0, initial); err != nil || !ref.Write(0, initial) {
							t.Fatal(err)
						}
						input := value.Value{Lo: 0xfff123458001abcd, Hi: 0x9876543210fedcba}
						args := []value.Value{{Lo: address}, input}
						if pressure {
							for i := uint64(2); i < 14; i++ {
								args = append(args, value.Value{Lo: input.Lo + i, Hi: input.Hi - i})
							}
						}
						actual := uint32(address)
						if bounded {
							actual &= 8191
						}
						var want value.Value
						var trap wasm.TrapCode
						if load {
							want, trap = ref.Load(op, actual, offset, byte(lane), input)
						} else {
							trap = ref.Store(op, actual, offset, byte(lane), input)
						}
						out, exit, state := runFunction(t, compiled, lease, c, args, 1000, !fast, 256)
						if trap != wasm.NoTrap {
							if exit != native.Trap || state.Payload[0] != uint64(trap) || state.Source != f.Sources[in.Source].Wasm.Offset {
								t.Fatal(op, lane, offset, address, mode, exit, state)
							}
						} else if exit != native.Return || load && (len(out) != 1 || out[0] != want) {
							t.Fatal(op, lane, offset, address, mode, exit, out, want)
						}
						for i, want := range args {
							if got := readSlot(t, c, 16, slots.Slot(i)); got != want {
								t.Fatal("input clobbered", op, mode, i, got, want)
							}
						}
						got, wantBytes := make([]byte, 65536), make([]byte, 65536)
						mem.Read(0, got)
						ref.Read(0, wantBytes)
						if !bytes.Equal(got, wantBytes) {
							t.Fatal("partial or incorrect store", op, lane, offset, address, mode)
						}
						cost := 2 // Address local.get and the access itself.
						if !load || op >= 0xfd0054 && op <= 0xfd0057 {
							cost++ // Vector local.get.
						}
						if bounded {
							cost += 2 // Constant and mask.
						}
						if pressure && exit == native.Return {
							cost += 72 // Twelve vectors, three get/drop pairs each.
						}
						if !fast && state.Fuel != 1000-uint64(cost) {
							t.Fatal("vector access fuel", op, state.Fuel, cost)
						}
						// Every original access must check a new address on resumption.
						writeSlot(t, c, 16, in.A, value.Value{Lo: 0xffffffff})
						if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 256}); err != nil {
							t.Fatal(err)
						}
						if exit, err := c.Enter(lease, 0); err != nil || exit != native.Trap {
							t.Fatal("unchecked cold vector access", op, mode, exit, err)
						}
					}
					ref.Release()
					c.Close()
					lease.Close()
					mem.Close()
				}
			}
		}
	}
}

func TestVectorMemoryTrapPreservesIntermediateSlots(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	literal := module.Instr{Op: 0xfd000c, Vector: [16]byte{1, 2, 3, 4, 5, 6, 7, 8}}
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.V128, wasm.I32}}, 0, module.Expr{
		{Op: 0x20}, literal, {Op: 0xfd00ae}, {Op: 0x21},
		{Op: 0x20, A: 1}, {Op: 0xfd0000}, {Op: 0x1a},
		{Op: 0x20}, literal, {Op: 0xfd00ae}, {Op: 0x21}, {Op: 0xb},
	})
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	fast, fl, fc := functionContext(t, f, true)
	slow, sl, sc := functionContext(t, f)
	bindMemory(t, fc, 65536, 65536)
	bindMemory(t, sc, 65536, 65536)
	args := []value.Value{{Lo: 0xfedcba9876543210, Hi: 0x12345678deadbeef}, {Lo: 0xffffffff}}
	_, actual, a := runFunction(t, fast, fl, fc, args, 100, false, 100)
	_, expected, b := runFunction(t, slow, sl, sc, args, 100, false, 100)
	if actual != native.Trap || expected != actual || a.Source != b.Source {
		t.Fatal(actual, expected, a, b)
	}
	for s := range f.Slots {
		if got, want := readSlot(t, fc, 16, slots.Slot(s)), readSlot(t, sc, 16, slots.Slot(s)); got != want {
			t.Fatal("copy omitted before trapping vector access", s, got, want)
		}
	}
}

func TestVectorMemoryFuelCancelAndCodeLimits(t *testing.T) {
	for _, op := range []uint32{0xfd0000, 0xfd000b, 0xfd0054, 0xfd005b} {
		f, pc := vectorMemoryFunction(t, op, 0, 0, false, false)
		for _, arch := range []string{"amd64", "arm64"} {
			compiled, err := CompileFunction(arch, f, DefaultLimits())
			if err != nil {
				t.Fatal(err)
			}
			limit := DefaultLimits()
			limit.MaxCodeBytes = uint64(len(compiled.Entry.Code))
			if _, err := CompileFunction(arch, f, limit); err != nil {
				t.Fatal(err)
			}
			limit.MaxCodeBytes--
			if _, err := CompileFunction(arch, f, limit); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(err)
			}
		}
		if !native.Available {
			continue
		}
		compiled, lease, c := functionContext(t, f)
		args := []value.Value{{}, {Lo: ^uint64(0), Hi: ^uint64(0)}}
		if _, exit, _ := runFunction(t, compiled, lease, c, args, 100, true, 64); exit != native.BadState {
			t.Fatal(exit)
		}
		mem := bindMemory(t, c, 65536, 65536)
		for fuel := uint64(0); fuel <= uint64(pc); fuel++ {
			if _, exit, _ := runFunction(t, compiled, lease, c, args, fuel, true, 64); exit != native.FuelExhausted {
				t.Fatal(op, fuel, exit)
			}
			var got [16]byte
			if err := mem.Read(0, got[:]); err != nil || got != ([16]byte{}) {
				t.Fatal("write before fuel", got, err)
			}
		}
		c.Cancel()
		if exit, err := c.Enter(lease, 0); err != nil || exit != native.Poll {
			t.Fatal(exit, err)
		}
	}
}

func TestVectorMemoryReloadsBindingAfterGrowth(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.V128}, Results: []wasm.ValueType{wasm.V128}}, 0,
		module.Expr{{Op: 0x41, Bits: 1}, {Op: 0x40}, {Op: 0x1a},
			{Op: 0x41, Bits: 65536}, {Op: 0x20}, {Op: 0xfd000b},
			{Op: 0x41, Bits: 65536}, {Op: 0xfd0000}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
	f := functionModule(t, m).Functions[0]
	input := value.Value{Lo: 0x0123456789abcdef, Hi: 0xfedcba9876543210}
	for _, fast := range []bool{false, true} {
		compiled, lease, c := configuredContext(t, f, Config{Unmetered: fast, MemoryMinimumPages: 1})
		mem := bindMemory(t, c, 65536, 131072)
		out, exit, state := runFunction(t, compiled, lease, c, []value.Value{input}, 100, !fast, 64, mem)
		if exit != native.Return || len(out) != 1 || out[0] != input || !fast && state.Fuel != 92 {
			t.Fatal(fast, exit, out, state)
		}
	}
}

func TestVectorAliasesPreserveEverySlotAndColdEntry(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	a := module.Instr{Op: 0xfd000c, Vector: [16]byte{1, 3, 5, 7, 9, 11, 13, 15}}
	b := module.Instr{Op: 0xfd000c, Vector: [16]byte{3, 5, 7, 11, 13, 17, 19, 23}}
	bodies := []module.Expr{
		{{Op: 0x20}, {Op: 0x41}, {Op: 0xfd0000}, {Op: 0xfd0051}, {Op: 0x21},
			{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0xfd00ae}, {Op: 0x21, A: 1},
			{Op: 0x20, A: 1}, {Op: 0x41, Bits: 16}, {Op: 0xfd0000}, {Op: 0xfd0051}, {Op: 0xb}},
		{{Op: 0x20}, {Op: 0x21, A: 1}, {Op: 0x20, A: 1}, {Op: 0x41}, {Op: 0xfd0000}, {Op: 0xfd0051}, {Op: 0x21},
			{Op: 0x20, A: 1}, a, {Op: 0xfd00b5}, {Op: 0x21, A: 1},
			{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0xfd00ae}, {Op: 0xb}},
		{{Op: 0x20}, {Op: 0x21, A: 1}, {Op: 0x41}, {Op: 0xfd0000}, {Op: 0x1a},
			{Op: 0x20, A: 1}, a, {Op: 0xfd00ae}, b, {Op: 0xfd00b5}, {Op: 0x21},
			{Op: 0x20, A: 1}, {Op: 0x20}, {Op: 0xfd0051}, {Op: 0xb}},
	}
	for bi, body := range bodies {
		m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.V128, wasm.V128}, Results: []wasm.ValueType{wasm.V128}}, 0, body)
		m.Memories = []module.Limits{{Min: 1}}
		f := functionModule(t, m).Functions[0]
		fast, fl, fc := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
		_, sl, sc := functionContext(t, f)
		fm, sm := bindMemory(t, fc, 65536, 65536), bindMemory(t, sc, 65536, 65536)
		data := make([]byte, 65536)
		for i := range data {
			data[i] = byte(i*193 + 127)
		}
		fm.Write(0, data)
		sm.Write(0, data)
		// Bodies 0 and 2 now coalesce the consumed copies. Body 1 still reads
		// a materialized copy before either local changes, exercising aliases.
		if bi == 1 && fast.AliasedVectorReads == 0 {
			t.Fatal("copy chain not bypassed", bi)
		}
		for pc := range f.Code {
			for i, typ := range f.Slots {
				v := value.Value{Lo: uint64(i)}
				if typ == wasm.V128 {
					v = value.Value{Lo: 0x0123456789abcdef + uint64(i), Hi: 0xfedcba9876543210 - uint64(i)}
				}
				writeSlot(t, fc, 16, slots.Slot(i), v)
				writeSlot(t, sc, 16, slots.Slot(i), v)
			}
			s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1000}
			if err := fc.SetState(s); err != nil {
				t.Fatal(err)
			}
			if err := sc.SetState(s); err != nil {
				t.Fatal(err)
			}
			if exit, err := fc.Enter(fl, 0); err != nil || exit != native.Return {
				t.Fatal(bi, pc, exit, err)
			}
			if exit, err := sc.Enter(sl, 0); err != nil || exit != native.Return {
				t.Fatal(bi, pc, exit, err)
			}
			for i := range f.Slots {
				if got, want := readSlot(t, fc, 16, slots.Slot(i)), readSlot(t, sc, 16, slots.Slot(i)); got != want {
					t.Fatal(bi, pc, i, got, want)
				}
			}
		}
	}
}
