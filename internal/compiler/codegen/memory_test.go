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

func memoryFunction(t *testing.T, op, offset uint32, bounded ...bool) (*slots.Function, int) {
	typ := wasm.I32
	switch op {
	case 0x29, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x37, 0x3c, 0x3d, 0x3e:
		typ = wasm.I64
	case 0x2a, 0x38:
		typ = wasm.F32
	case 0x2b, 0x39:
		typ = wasm.F64
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}}
	body := module.Expr{{Op: 0x20}}
	if len(bounded) != 0 && bounded[0] {
		body = append(body, module.Instr{Op: 0x41, Bits: 8191}, module.Instr{Op: 0x71})
	}
	if op <= 0x35 {
		sig.Results = []wasm.ValueType{typ}
	} else {
		sig.Params = append(sig.Params, typ)
		body = append(body, module.Instr{Op: 0x20, A: 1})
	}
	body = append(body, module.Instr{Op: op, B: offset}, module.Instr{Op: 0xb})
	m := functionFixture(sig, 0, body)
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
	f := functionModule(t, m).Functions[0]
	for pc, in := range f.Code {
		if in.Op == op {
			return f, pc
		}
	}
	t.Fatal("missing memory instruction")
	return nil, 0
}
func bindMemory(t *testing.T, c *native.Context, size, limit uint64) *native.Memory {
	t.Helper()
	m, err := native.NewMemory(size, limit)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	if err = c.BindMemory(m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMemoryTargetsAndLimits(t *testing.T) {
	for op := uint32(0x28); op <= 0x3e; op++ {
		f, _ := memoryFunction(t, op, 0xffffffff)
		for _, arch := range []string{"amd64", "arm64"} {
			for _, compile := range []func(string, *slots.Function, Limits) (*CompiledFunction, error){CompileFunction, CompileUnmeteredFunction} {
				compiled, err := compile(arch, f, DefaultLimits())
				if err != nil || !compiled.UsesMemory || compiled.Helpers != 0 {
					t.Fatal(arch, op, err)
				}
				limit := DefaultLimits()
				limit.MaxCodeBytes = uint64(len(compiled.Entry.Code))
				if _, err = compile(arch, f, limit); err != nil {
					t.Fatal(err)
				}
				limit.MaxCodeBytes--
				if _, err = compile(arch, f, limit); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestNativeScalarMemorySemantics(t *testing.T) {
	if !native.Available {
		return
	}
	initial := make([]byte, 65536)
	for i := range initial {
		initial[i] = byte((i*193 + 127) ^ (i >> 8))
	}
	for op := uint32(0x28); op <= 0x3e; op++ {
		for _, offset := range []uint32{0, 1, 2, 8, 4095, 4096, 65535, 0xffffffff} {
			for mode := 0; mode < 3; mode++ {
				fast, bounded := mode != 0, mode == 2
				f, pc := memoryFunction(t, op, offset, bounded)
				in := f.Code[pc]
				config := Config{Unmetered: fast}
				if bounded {
					config.MemoryMinimumPages = 1
				}
				compiled, lease, c := configuredContext(t, f, config)
				mem := bindMemory(t, c, 65536, 131072)
				reference, err := enginememory.New(module.Limits{Min: 1}, 2)
				if err != nil {
					t.Fatal(err)
				}
				for _, address := range []uint64{0, 1, 9, 65528, 65535, 65536, 0xffffffff, 1 << 32} {
					if err = mem.Write(0, initial); err != nil {
						t.Fatal(err)
					}
					reference.Write(0, initial)
					bits := uint64(0xfff123458001abcd) // Includes signaling NaN bits and sign bits.
					actual := uint32(address)
					if bounded {
						actual &= 8191
					}
					var want value.Value
					var trap wasm.TrapCode
					if op <= 0x35 {
						want, trap = reference.Load(op, uint64(actual), uint64(offset), 0, value.Value{})
					} else {
						trap = reference.Store(op, uint64(actual), uint64(offset), 0, value.Value{Lo: bits})
					}
					args := []value.Value{{Lo: address}}
					if op >= 0x36 {
						args = append(args, value.Value{Lo: bits})
					}
					out, exit, state := runFunction(t, compiled, lease, c, args, 100, !fast, 64)
					if trap != wasm.NoTrap {
						if exit != native.Trap || state.Payload[0] != uint64(trap) || state.Source != f.Sources[in.Source].Wasm.Offset {
							t.Fatal(op, offset, address, fast, exit, state)
						}
					} else if exit != native.Return || op <= 0x35 && (len(out) != 1 || out[0] != want) {
						t.Fatal(op, offset, address, fast, exit, out, want)
					}
					got, wantBytes := make([]byte, 65536), make([]byte, 65536)
					if err = mem.Read(0, got); err != nil {
						t.Fatal(err)
					}
					reference.Read(0, wantBytes)
					if !bytes.Equal(got, wantBytes) {
						t.Fatal("memory side effect differs", op, offset, address, fast)
					}
					// Resume directly at the original load/store with different slot data.
					writeSlot(t, c, 16, in.A, value.Value{Lo: 0xffffffff})
					if err = c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 64}); err != nil {
						t.Fatal(err)
					}
					if exit, err = c.Enter(lease, 0); err != nil || exit != native.Trap {
						t.Fatal("independent access bypassed bounds", op, exit, err)
					}
				}
				c.Close()
				lease.Close()
				mem.Close()
			}
		}
	}
}

func TestMemoryFuelBeforeEffectAndMissingBinding(t *testing.T) {
	if !native.Available {
		return
	}
	f, _ := memoryFunction(t, 0x36, 0)
	compiled, lease, c := functionContext(t, f)
	if _, exit, _ := runFunction(t, compiled, lease, c, []value.Value{{}, {Lo: 42}}, 10, true, 64); exit != native.BadState {
		t.Fatal(exit)
	}
	mem := bindMemory(t, c, 65536, 65536)
	for fuel := uint64(0); fuel <= 2; fuel++ {
		_, exit, _ := runFunction(t, compiled, lease, c, []value.Value{{}, {Lo: 42}}, fuel, true, 64)
		if exit != native.FuelExhausted {
			t.Fatal(fuel, exit)
		}
		var got [4]byte
		if err := mem.Read(0, got[:]); err != nil || got != ([4]byte{}) {
			t.Fatal(got, err)
		}
	}
	if _, exit, _ := runFunction(t, compiled, lease, c, []value.Value{{}, {Lo: 42}}, 3, true, 64); exit != native.Return {
		t.Fatal(exit)
	}
	var got [4]byte
	mem.Read(0, got[:])
	if got[0] != 42 {
		t.Fatal(got)
	}
}

func TestMemorySizeGrowthAndResume(t *testing.T) {
	if !native.Available {
		return
	}
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32, wasm.I32}}, 0,
		module.Expr{{Op: 0x20}, {Op: 0x40}, {Op: 0x3f}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 0, Max: 2, HasMax: true}}
	f := functionModule(t, m).Functions[0]
	for _, fast := range []bool{false, true} {
		for _, delta := range []uint32{0, 1, 2, 3, 0xffffffff} {
			compiled, lease, c := functionContext(t, f, fast)
			mem := bindMemory(t, c, 0, 131072)
			_, exit, s := runFunction(t, compiled, lease, c, []value.Value{{Lo: uint64(delta)}}, 3, !fast, 64)
			if exit != native.GrowMemory || compiled.MemoryGrows != 1 || s.Fuel != 3 && fast || s.Fuel != 1 && !fast {
				t.Fatal(exit, s)
			}
			in := f.Code[s.Payload[0]]
			if s.Continuation != s.Payload[0]+1 || s.Source != f.Sources[in.Source].Wasm.Offset {
				t.Fatal(s)
			}
			// The idle driver commits growth once, writes the typed result, and resumes
			// after the charged grow instruction. Cached bases/lengths must reload.
			result, pages := uint64(0), uint64(delta)
			if err := mem.Grow(uint64(delta) * 65536); err != nil {
				result, pages = 0xffffffff, 0
			}
			writeSlot(t, c, s.Used, in.Dst, value.Value{Lo: result})
			if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
				t.Fatal(exit, err)
			}
			state, err := c.State()
			if err != nil {
				t.Fatal(err)
			}
			returned := f.Code[state.Payload[0]]
			outputs := f.Args[returned.Extra : returned.Extra+returned.Count]
			if readSlot(t, c, 16, outputs[0]).Lo != result || readSlot(t, c, 16, outputs[1]).Lo != pages {
				t.Fatal(delta, result, pages, state)
			}
			if !fast && state.Fuel != 0 {
				t.Fatal("growth charged twice", state)
			}
			c.Close()
			lease.Close()
			mem.Close()
		}
	}
}

func TestMemoryAccessAfterRelocatingGrowth(t *testing.T) {
	if !native.Available {
		return
	}
	// Grow necessarily replaces the allocation. The resumed load must observe
	// both the copied old bytes and zeroed extension through a refreshed base.
	m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32, wasm.I32}}, 0,
		module.Expr{{Op: 0x41}, {Op: 0x41, Bits: 42}, {Op: 0x36},
			{Op: 0x41, Bits: 1}, {Op: 0x40}, {Op: 0x1a},
			{Op: 0x41}, {Op: 0x28}, {Op: 0x41, Bits: 65536}, {Op: 0x28}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
	f := functionModule(t, m).Functions[0]
	for _, fast := range []bool{false, true} {
		compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: fast, MemoryMinimumPages: 1})
		memory := bindMemory(t, ctx, 65536, 131072)
		out, exit, state := runFunction(t, compiled, lease, ctx, nil, 10, !fast, 64, memory)
		if exit != native.Return || len(out) != 2 || out[0].Lo != 42 || out[1] != (value.Value{}) || !fast && state.Fuel != 0 {
			t.Fatal(out, exit, state)
		}
	}
}

func TestMemoryRegisterPressure(t *testing.T) {
	if !native.Available {
		return
	}
	for _, typ := range []wasm.ValueType{wasm.I64, wasm.F32, wasm.F64} {
		load, store := uint32(0x29), uint32(0x37)
		if typ == wasm.F32 {
			load, store = 0x2a, 0x38
		}
		if typ == wasm.F64 {
			load, store = 0x2b, 0x39
		}
		sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}}
		body := module.Expr{}
		for i := 0; i < 20; i++ {
			sig.Params = append(sig.Params, typ)
			sig.Results = append(sig.Results, typ)
			body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x41, Bits: 31}, module.Instr{Op: 0x71}, module.Instr{Op: 0x20, A: uint32(i + 1)}, module.Instr{Op: store, B: uint32(i * 8)})
		}
		for i := 0; i < 20; i++ {
			body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x41, Bits: 31}, module.Instr{Op: 0x71}, module.Instr{Op: load, B: uint32(i * 8)})
		}
		body = append(body, module.Instr{Op: 0xb})
		m := functionFixture(sig, 0, body)
		m.Memories = []module.Limits{{Min: 1}}
		f := functionModule(t, m).Functions[0]
		patterns := []uint64{0, 1, 0x80000000, 0x7f800001, 0xffc12345, 0x7ff0000000000001, 0xfff123456789abcd, 1 << 63, ^uint64(0)}
		for mode := 0; mode < 3; mode++ {
			config := Config{Unmetered: mode != 0}
			if mode == 2 {
				config.MemoryMinimumPages = 1
			}
			compiled, lease, ctx := configuredContext(t, f, config)
			bindMemory(t, ctx, 65536, 65536)
			for _, address := range []uint64{0, 1, 31, 0xffffffff} {
				args := []value.Value{{Lo: address}}
				for i := 0; i < 20; i++ {
					v := patterns[i%len(patterns)]
					if typ == wasm.F32 {
						v = uint64(uint32(v))
					}
					args = append(args, value.Value{Lo: v})
				}
				got, exit, _ := runFunction(t, compiled, lease, ctx, args, 1000, mode == 0, 512)
				if exit != native.Return || len(got) != 20 {
					t.Fatal(typ, mode, exit, len(got))
				}
				for i, v := range got {
					if v != args[i+1] {
						t.Fatal(typ, mode, i, v, args[i+1])
					}
				}
			}
		}
	}
}
