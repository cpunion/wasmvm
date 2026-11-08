package codegen

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func boundedArrayFixture(field module.FieldType, mask uint32) (*slots.Function, Config) {
	f, config := heapFixture(field, true, true)
	f.Slots = append(f.Slots, wasm.I32, wasm.AnyRef)
	f.LocalCount += 2
	constant := func(bits uint64) slots.Instruction {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x41, Bits: bits}})
		return slots.Instruction{Op: 0x41, Source: source, Dst: 6, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
	}
	numeric := func(op uint32) slots.Instruction {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op}})
		return slots.Instruction{Op: op, Source: source, Dst: 1, A: 1, B: 6, C: slots.NoSlot}
	}
	code := []slots.Instruction{constant(uint64(mask)), numeric(0x71), f.Code[0], constant(1), numeric(0x6a)}
	code = append(code, f.Code[1:]...)
	f.Code = code
	// A different raw reference after the first access exercises full identity
	// checks even when a cache has already been promoted in this entry.
	f.Code[5].A, f.Code[6].A = 7, 7
	f.Args[f.Code[5].Extra], f.Args[f.Code[6].Extra] = 7, 7
	addHeapFixtureLoop(f, 0)
	return f, config
}

// Compare complete frames, payload bytes, exits and trap locations with the
// ordinary checked emitter for defined exits. Invalid internal states require the
// same BadState exit and no payload effects, not a resumable frame. Every PC is
// an independent cold continuation; the
// raw indices deliberately violate the compile-time hints on those entries.
func TestNativeBoundedArrayCacheAgainstCheckedAccess(t *testing.T) {
	fields := []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I32, Packed: 16},
		{Value: wasm.I32}, {Value: wasm.I64}, {Value: wasm.F32}, {Value: wasm.F64}, {Value: wasm.AnyRef}, {Value: wasm.V128}}
	for _, field := range fields {
		for _, mask := range []uint32{0, 1, 4109} {
			t.Run(fmt.Sprintf("%v/%d", field, mask), func(t *testing.T) {
				f, config := boundedArrayFixture(field, mask)
				for _, arch := range []string{"amd64", "arm64"} {
					e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: true}
					e.prepareHeapAccesses(config)
					if bound, ok := e.heapIndexBounds[0]; !ok || bound != mask+1 {
						t.Fatal("bounded array class not selected", e.heapIndexBounds)
					}
					config.Unmetered = true
					compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || compiled.HeapCaches == 0 {
						t.Fatal(arch, compiled, err)
					}
				}
				if !native.Available {
					return
				}
				config.Unmetered = false
				_, checked, checkedContext := configuredContext(t, f, config)
				config.Unmetered = true
				_, cached, cachedContext := configuredContext(t, f, config)
				size := uint64(128 + (mask+4)*uint32(field.HeapBytes()))
				storage, err := native.NewMemory(size, size)
				if err != nil {
					t.Fatal(err)
				}
				defer storage.Close()
				buffer, _ := storage.Buffer()
				address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
				words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
				const ref = uint64(0x8000000100000001)
				tests := []struct {
					first, second, index uint64
					length               uint32
				}{
					{ref, ref, 0, 0}, {ref, ref, 0, 1}, {ref, ref, 1, 2},
					{ref, ref, uint64(mask + 1), mask + 2}, // Promote, then reject a hit above the hinted bound.
					{ref, ref, uint64(mask + 2), mask + 4}, // Unexpected but valid indices use ordinary checks.
					{ref, ref, 0x100000001, mask + 4},
					{ref, ref, 0xffffffff, mask + 4},
					{ref, ref, 0x80000000, mask + 4}, {ref, ref, 0x80000001, mask + 4},
					{ref, 0, 0, mask + 4}, {ref, ref + 1<<32, 0, mask + 4},
					{ref, ref + 1, 0, mask + 4}, {ref, ref &^ (1 << 63), 0, mask + 4},
					{0, ref, 0, mask + 4}, {ref + 1<<32, ref, 0, mask + 4},
				}
				for _, tc := range tests {
					for pc := range f.Code {
						var wantFrame, wantMemory []byte
						var wantExit native.Exit
						var wantState native.State
						for variant, context := range []*native.Context{checkedContext, cachedContext} {
							clear(buffer)
							words[0], words[1], words[6], words[7], words[8] = 2, address+16, ref, address+128, 0xfedcba98<<32|uint64(tc.length)
							if tc.length == 0 {
								words[7] = 0 // Empty arrays must trap before loading a payload.
							}
							for i := 128; i < len(buffer); i++ {
								buffer[i] = byte(i*37 + 19)
							}
							for slot := range f.Slots {
								writeSlot(t, context, 16, slots.Slot(slot), value.Value{})
							}
							argument := value.Value{Lo: 0x8765432180008081}
							if field.Value == wasm.I32 || field.Value == wasm.F32 {
								argument.Lo = uint64(uint32(argument.Lo))
							}
							if field.Value == wasm.V128 {
								argument.Hi = 0xfedcba9876543210
							}
							for slot, v := range map[slots.Slot]value.Value{0: {Lo: tc.first}, 1: {Lo: tc.index},
								2: argument, 6: {Lo: 0xffffffff}, 7: {Lo: tc.second}} {
								writeSlot(t, context, 16, slot, v)
							}
							s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 64}
							s.Payload[2] = address
							if err := context.SetState(s); err != nil {
								t.Fatal(err)
							}
							lease := checked
							if variant != 0 {
								lease = cached
							}
							exit, err := context.Enter(lease, 0)
							if err != nil {
								t.Fatal(err)
							}
							state, _ := context.State()
							frame := make([]byte, len(f.Slots)*16)
							if err := context.Read(16, frame); err != nil {
								t.Fatal(err)
							}
							if variant == 0 {
								wantExit, wantState = exit, state
								wantFrame, wantMemory = frame, append([]byte(nil), buffer...)
							} else if exit != wantExit || state.Source != wantState.Source ||
								(exit == native.Trap && state.Payload[0] != wantState.Payload[0]) ||
								(exit != native.BadState && !bytes.Equal(frame, wantFrame)) || !bytes.Equal(buffer, wantMemory) {
								t.Fatal("cached access differs", tc, pc, exit, wantExit, state, wantState,
									"frame equal", bytes.Equal(frame, wantFrame), "payload equal", bytes.Equal(buffer, wantMemory))
							}
						}
					}
				}
			})
		}
	}
}

func TestHeapBoundsUnknownIndicesDisablePayloadCaching(t *testing.T) {
	for _, mask := range []uint32{1, 0xffffffff} {
		f, config := boundedArrayFixture(module.FieldType{Value: wasm.I32}, mask)
		if mask == 1 {
			// One access whose index is unconstrained disables the entire class,
			// including accesses whose own indices have a useful hint.
			f.Code[6].B = 6
			f.Code[3].Op = 0x20
			f.Sources[f.Code[3].Source].Wasm.A = 2
		}
		for _, unmetered := range []bool{false, true} {
			e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, unmetered: unmetered}
			e.prepareHeapAccesses(config)
			if len(e.heapIndexBounds) != 0 {
				t.Fatal("unbounded or metered class has payload hints", mask, unmetered, e.heapIndexBounds)
			}
		}
	}
}
