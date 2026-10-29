package codegen

import (
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

func tableFixture(typ wasm.ValueType, table uint32, pressure int) *slots.Function {
	f := &slots.Function{Signature: module.FuncType{Params: []wasm.ValueType{wasm.I32, typ}, Results: []wasm.ValueType{typ, typ, wasm.FuncRef}},
		LocalCount: 5, Slots: []wasm.ValueType{wasm.I32, typ, typ, typ, wasm.FuncRef}, Args: []slots.Slot{2, 3, 4}}
	for pc, op := range []uint32{0x25, 0x26, 0x25, 0xd2} {
		index := table
		if pc == 1 {
			index++
		}
		if pc == 3 {
			index = 2
		}
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: index, Offset: uint64(100 + pc)}})
		in := slots.Instruction{Op: op, Source: uint32(pc), Dst: slots.NoSlot, A: 0, B: slots.NoSlot, C: slots.NoSlot}
		switch pc {
		case 0:
			in.Dst = 2
		case 1:
			in.B = 1
		case 2:
			in.Dst = 3
		case 3:
			in.Dst, in.A = 4, slots.NoSlot
		}
		f.Code = append(f.Code, in)
	}
	for i := range pressure {
		source, dst := uint32(len(f.Sources)), slots.Slot(len(f.Slots))
		f.Slots = append(f.Slots, wasm.I64)
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x42, Bits: 0x123456789abcdef0 + uint64(i)}})
		for range 5 {
			f.Code = append(f.Code, slots.Instruction{Op: 0x42, Source: source, Dst: dst, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
	}
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: 3, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	return f
}

func TestTablesHandlesAliasesBoundsAndEveryContinuation(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.FuncRef, wasm.ExternRef} {
		for _, unmetered := range []bool{false, true} {
			for _, table := range []uint32{0, 3000} {
				for _, pressure := range []int{0, 20} {
					f := tableFixture(typ, table, pressure)
					config := Config{Unmetered: unmetered, FunctionCount: 3, TableCount: table + 2}
					for _, arch := range []string{"amd64", "arm64"} {
						c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
						if err != nil || c.NativeTables != 3 || c.NativeFunctions != 1 || c.Helpers != 0 || !c.UsesRuntimeData || len(c.ReferenceSlots) != 4 {
							t.Fatal(arch, c, err)
						}
						limits := DefaultLimits()
						limits.MaxCodeBytes = uint64(len(c.Entry.Code) - 1)
						if _, err := CompileFunctionWithConfig(arch, f, limits, config); !errors.Is(err, wasm.ErrLimit) {
							t.Fatal("table code budget", err)
						}
					}
					if !native.Available {
						continue
					}
					_, lease, ctx := configuredContext(t, f, config)
					offset, _ := nativeabi.TablePairOffset(3, table)
					size := uint64(offset) + 48
					data, err := native.NewMemory(size, size)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(data.Close)
					buffer, _ := data.Buffer()
					address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
					words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
					words[0], words[1], words[2] = 0, 0xfedcba9876543210, 0x123456789abcdef0
					cell := words[offset/8+4:]
					for _, length := range []uint64{0, 1, 2} {
						pointer := address + uint64(offset) + 32
						if length == 0 {
							pointer = 0
						}
						words[offset/8], words[offset/8+1], words[offset/8+2], words[offset/8+3] = length, pointer, length, pointer
						for _, index := range []uint64{0, 1, 2, 0xffffffff, 0x100000000} {
							for pc := range f.Code {
								cell[0], cell[1] = 0, 0xfedcba9876543210
								elements := [2]uint64{cell[0], cell[1]}
								want := make([]value.Value, len(f.Slots))
								want[0].Lo = index
								want[1].Lo = 0xdeadbeefcafebabe
								for i := range want {
									if !unmetered {
										want[i].Hi = uint64(i+1) * 17
									}
									writeSlot(t, ctx, 16, slots.Slot(i), want[i])
								}
								expected, source := native.Return, uint64(0)
								for _, in := range f.Code[pc:] {
									idx := uint64(uint32(index))
									if (in.Op == 0x25 || in.Op == 0x26) && idx >= length {
										expected, source = native.Trap, f.Sources[in.Source].Wasm.Offset
										break
									}
									switch in.Op {
									case 0x25:
										want[in.Dst] = value.Value{Lo: elements[idx]}
									case 0x26:
										elements[idx] = want[1].Lo
									case 0xd2:
										want[in.Dst] = value.Value{Lo: words[2]}
									case 0x42:
										want[in.Dst] = value.Value{Lo: f.Sources[in.Source].Wasm.Bits}
									}
								}
								s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1024, Fuel: 10}
								s.Payload[2] = address
								if err := ctx.SetState(s); err != nil {
									t.Fatal(err)
								}
								if exit, err := ctx.Enter(lease, 0); err != nil || exit != expected {
									t.Fatal(typ, unmetered, table, pressure, length, index, pc, exit, expected, err)
								}
								state, _ := ctx.State()
								if expected == native.Trap && (state.Source != source || wasm.TrapCode(state.Payload[0]) != wasm.TrapTableBounds) {
									t.Fatal("table trap location", state, source)
								}
								if cell[0] != elements[0] || cell[1] != elements[1] {
									t.Fatal("aliased table effects", cell, elements)
								}
								for i, v := range want {
									if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
										t.Fatal("table frame", typ, unmetered, length, index, pc, i, got, v)
									}
								}
							}
						}
					}
					for pc := range f.Code {
						if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
							t.Fatal(err)
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.BadState {
							t.Fatal("missing table view", pc, exit, err)
						}
					}
				}
			}
		}
	}
}

func TestReferencesRetainCheckedFallbacks(t *testing.T) {
	for _, kind := range []string{"tables", "index", "float-index", "get-type", "set-type", "no-get", "no-set", "functions", "function-type", "function-destination", "table-offset", "function-offset"} {
		f := tableFixture(wasm.FuncRef, 0, 0)
		config := Config{FunctionCount: 3, TableCount: 2}
		switch kind {
		case "tables":
			config.TableCount = 0
		case "index":
			f.Sources[0].Wasm.A = 2
		case "float-index":
			f.Slots[0] = wasm.F64
		case "get-type":
			f.Slots[2] = wasm.I32
		case "set-type":
			f.Slots[1] = wasm.I64
		case "no-get":
			f.Code[0].Dst = slots.NoSlot
		case "no-set":
			f.Code[1].B = slots.NoSlot
		case "functions":
			config.FunctionCount = 2
		case "function-type":
			f.Slots[4] = wasm.ExternRef
		case "function-destination":
			f.Code[3].Dst = slots.NoSlot
		case "table-offset":
			config.TableCount = ^uint32(0)
			f.Sources[0].Wasm.A = ^uint32(0) - 1
		case "function-offset":
			config.FunctionCount = ^uint32(0)
			f.Sources[3].Wasm.A = ^uint32(0) - 1
		}
		c, err := CompileFunctionWithConfig(runtime.GOARCH, f, DefaultLimits(), config)
		if err != nil || c.Helpers == 0 {
			t.Fatal(kind, c, err)
		}
	}
}
