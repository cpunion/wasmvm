package codegen

import (
	"encoding/binary"
	"errors"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func allocationFunction(fields []module.FieldType, defaults, large bool) (*slots.Function, Config, []slots.Slot) {
	op := uint32(0xfb0000)
	if defaults {
		op++
	}
	f := &slots.Function{Sources: []ir.Source{{Wasm: module.Instr{Op: op}}, {Wasm: module.Instr{Op: 0xb}}}}
	if large {
		f.Slots = make([]wasm.ValueType, 5000)
		for i := range f.Slots {
			f.Slots[i] = wasm.I32
		}
	}
	var inputs []slots.Slot
	for _, field := range fields {
		inputs = append(inputs, slots.Slot(len(f.Slots)))
		f.Slots = append(f.Slots, field.Unpacked().Storage())
	}
	// Reuse a reference input as the constructor output. Its original value
	// must be stored before the slot is assigned the new handle.
	dst := slots.Slot(len(f.Slots))
	f.Slots = append(f.Slots, wasm.AnyRef)
	if len(fields) != 0 && fields[len(fields)-1].Value == wasm.AnyRef {
		dst = inputs[len(inputs)-1]
	}
	if !defaults {
		f.Args = append(f.Args, inputs...)
	}
	f.Code = []slots.Instruction{{Op: op, Dst: dst, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: uint32(len(f.Args))}}
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: 1, Extra: uint32(len(f.Args)), Count: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	f.Args = append(f.Args, dst)
	return f, Config{Unmetered: true, NativeAllocations: true, HeapTypes: []module.TypeDef{{Kind: module.StructKind, Fields: fields}}}, inputs
}

func TestNativeAllocationTargetsFieldsAliasesAndSlowGuards(t *testing.T) {
	fields := []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I32, Packed: 16}, {Value: wasm.I32}, {Value: wasm.I64}, {Value: wasm.F32}, {Value: wasm.F64}, {Value: wasm.V128}, {Value: wasm.AnyRef}}
	values := []value.Value{{Lo: 0xabcdefff}, {Lo: 0xf123ffff}, {Lo: 0x89abcdef}, {Lo: 0xfedcba9876543210}, {Lo: 0x7f801234}, {Lo: 0x7ff0123456789abc}, {Lo: 0x0123456789abcdef, Hi: 0xfedcba9876543210}, {Lo: 1<<62 | 7}}
	for _, defaults := range []bool{false, true} {
		for _, large := range []bool{false, true} {
			f, cfg, inputs := allocationFunction(fields, defaults, large)
			for _, arch := range []string{"amd64", "arm64"} {
				compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
				if err != nil || compiled.NativeAllocations != 1 || !compiled.UsesRuntimeData {
					t.Fatal(arch, compiled, err)
				}
				limit := DefaultLimits()
				limit.MaxCodeBytes = uint64(len(compiled.Entry.Code)) - 1
				if _, err := CompileFunctionWithConfig(arch, f, limit, cfg); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("constructor code limit", err)
				}
			}
			if !native.Available {
				continue
			}
			compiled, lease, ctx := configuredContext(t, f, cfg)
			// One native buffer holds the immutable view, batch, descriptors and
			// payload. No Go pointer is stored in the generated-code protocol.
			storage, err := native.NewMemory(4096, 4096)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(storage.Close)
			buffer, _ := storage.Buffer()
			words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
			address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
			for _, guard := range []string{"success", "header", "capacity", "space", "overflow", "quota"} {
				clear(buffer)
				words[0], words[1], words[2] = 4, address+1024, address+64
				header := words[8:]
				header[nativeabi.HeapAllocationCapacity/8] = 1
				header[nativeabi.HeapAllocationHandles/8] = address + 512
				header[nativeabi.HeapAllocationBase/8] = address + 2048
				header[nativeabi.HeapAllocationLimit/8] = 1024
				header[nativeabi.HeapAllocationBudget/8] = 1024
				header[nativeabi.HeapAllocationDescriptors/8] = address + 1024
				words[64] = 1<<63 | 1<<32 | 1
				switch guard {
				case "header":
					words[2] = 0
				case "capacity":
					header[nativeabi.HeapAllocationCapacity/8] = 0
				case "space":
					header[nativeabi.HeapAllocationLimit/8] = 1
				case "overflow":
					header[nativeabi.HeapAllocationUsed/8] = ^uint64(0) - 1
				case "quota":
					header[nativeabi.HeapAllocationBudget/8] = 63
				}
				for i, slot := range inputs {
					writeSlot(t, ctx, 16, slot, values[i])
				}
				state := native.State{Used: 16, Depth: 1}
				state.Payload[2] = address
				if err := ctx.SetState(state); err != nil {
					t.Fatal(err)
				}
				exit, err := ctx.Enter(lease, 0)
				if err != nil {
					t.Fatal(err)
				}
				if guard != "success" {
					if exit != native.RuntimeHelper || header[0] != 0 {
						t.Fatal("guard did not preserve slow allocation", runtime.GOARCH, guard, exit)
					}
					continue
				}
				if exit != native.Return || header[0] != 1 {
					t.Fatal("native allocation failed", exit, header[:8], compiled.NativeAllocations)
				}
				offset := 2048
				for i, field := range fields {
					var expected [16]byte
					if !defaults {
						binary.LittleEndian.PutUint64(expected[:8], values[i].Lo)
						binary.LittleEndian.PutUint64(expected[8:], values[i].Hi)
					}
					for j := 0; j < field.HeapBytes(); j++ {
						if buffer[offset+j] != expected[j] {
							t.Fatal("constructor field bits", defaults, large, i, j, buffer[offset+j], expected[j])
						}
					}
					offset += field.HeapBytes()
				}
			}
		}
	}
}
