package codegen

import (
	"encoding/binary"
	"errors"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func atomicFunction(sub uint32, large bool) (*slots.Function, []slots.Slot) {
	variant, group := (sub-0x10)%7, (sub-0x10)/7
	typ := wasm.I64
	if variant == 0 || variant == 2 || variant == 3 {
		typ = wasm.I32
	}
	f := &slots.Function{Sources: []ir.Source{{Wasm: module.Instr{Op: 0xfe0000 | sub, Offset: 123}}, {Wasm: module.Instr{Op: 0xb}}}}
	if large {
		f.Slots = make([]wasm.ValueType, 5000)
		for i := range f.Slots {
			f.Slots[i] = wasm.I32
		}
	}
	var inputs []slots.Slot
	for _, typ := range []wasm.ValueType{wasm.I64, typ, typ} {
		inputs = append(inputs, slots.Slot(len(f.Slots)))
		f.Slots = append(f.Slots, typ)
	}
	in := slots.Instruction{Op: 0xfe0000 | sub, A: inputs[0], B: inputs[1], C: slots.NoSlot, Dst: inputs[1]}
	if group == 0 {
		in.B = slots.NoSlot
	} else if group == 1 {
		in.Dst = slots.NoSlot
	} else if group == 8 {
		in.C, in.Dst = inputs[2], inputs[2]
	}
	f.Code = []slots.Instruction{in, {Op: slots.ControlBase + uint32(ir.Return), Source: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}
	if group != 1 {
		f.Code[1].Count = 1
		f.Args = []slots.Slot{in.Dst}
	}
	return f, inputs
}

func TestNativeAtomicTargetsWidthsAliasesBoundsAndFuel(t *testing.T) {
	for sub := uint32(0x10); sub <= 0x4e; sub++ {
		width := [...]uint32{4, 8, 1, 2, 1, 2, 4}[(sub-0x10)%7]
		group := (sub - 0x10) / 7
		for _, fast := range []bool{false, true} {
			f, inputs := atomicFunction(sub, sub == 0x49)
			cfg := Config{Unmetered: fast, NativeThreads: true, MemoryCount: 1}
			for _, arch := range []string{"amd64", "arm64"} {
				compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
				if err != nil || compiled.NativeAtomics != 1 || compiled.Helpers != 0 || !compiled.UsesMemory {
					t.Fatal(arch, sub, compiled, err)
				}
				limit := DefaultLimits()
				limit.MaxCodeBytes = uint64(len(compiled.Entry.Code)) - 1
				if _, err := CompileFunctionWithConfig(arch, f, limit, cfg); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("atomic code budget", err)
				}
			}
			if !native.Available {
				continue
			}
			_, lease, ctx := configuredContext(t, f, cfg)
			mem := bindMemory(t, ctx, 64, 64)
			ref, err := native.NewSharedMemory(64, 64)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ref.Close)
			refBase, _, _, err := ref.SharedView()
			if err != nil {
				t.Fatal(err)
			}
			initial := make([]byte, 64)
			for i := range initial {
				initial[i] = byte(i*17 + 11)
			}
			for _, address := range []uint64{0, 8, 1, 63, 64, ^uint64(0)} {
				for _, a := range []uint64{0xffffffffffffffff, 0x82716050aabbccdd} {
					if err := mem.Write(0, initial); err != nil {
						t.Fatal(err)
					}
					if err := ref.Write(0, initial); err != nil {
						t.Fatal(err)
					}
					const b = uint64(0x1928374655647382)
					for i, v := range []uint64{address, a, b} {
						writeSlot(t, ctx, 16, inputs[i], value.Value{Lo: v})
					}
					state := native.State{Used: 16, Depth: 1, Budget: 64, Fuel: 1}
					if !fast {
						state.Payload[3] = 1
					}
					if err := ctx.SetState(state); err != nil {
						t.Fatal(err)
					}
					exit, err := ctx.Enter(lease, 0)
					if err != nil {
						t.Fatal(err)
					}
					gotState, _ := ctx.State()
					wantTrap := wasm.NoTrap
					if address > 64 || uint64(width) > 64-address {
						wantTrap = wasm.TrapMemoryBounds
					} else if address%uint64(width) != 0 {
						wantTrap = wasm.TrapAtomicUnaligned
					}
					if wantTrap != wasm.NoTrap {
						if exit != native.Trap || wasm.TrapCode(gotState.Payload[0]) != wantTrap || gotState.Source != 123 {
							t.Fatal("atomic trap order", sub, address, exit, gotState)
						}
					} else {
						want, err := native.SharedAtomic(unsafe.Add(refBase, uintptr(address)), width, native.AtomicOperation(group), a, b)
						if err != nil || exit != native.Return {
							t.Fatal(sub, address, exit, err)
						}
						if group != 1 && readSlot(t, ctx, 16, f.Code[0].Dst) != (value.Value{Lo: want}) {
							t.Fatal("atomic old bits or aliased input", runtime.GOARCH, sub, address, readSlot(t, ctx, 16, f.Code[0].Dst), want)
						}
					}
					var actual, expected [64]byte
					mem.Read(0, actual[:])
					ref.Read(0, expected[:])
					if actual != expected || !fast && gotState.Fuel != 0 {
						t.Fatal("atomic memory or fuel", sub, address, binary.LittleEndian.Uint64(actual[:8]), binary.LittleEndian.Uint64(expected[:8]), gotState.Fuel)
					}
				}
			}
		}
	}
}
