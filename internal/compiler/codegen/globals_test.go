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

func globalFixture(typ wasm.ValueType) *slots.Function {
	f := &slots.Function{Signature: module.FuncType{Params: []wasm.ValueType{typ}, Results: []wasm.ValueType{typ, typ}},
		LocalCount: 4, Slots: []wasm.ValueType{typ, typ, typ, typ}, Args: []slots.Slot{2, 3}}
	for pc, op := range []uint32{0x23, 0x24, 0x23, 0x23, slots.ControlBase + uint32(ir.Return)} {
		index := uint32(0)
		if pc == 2 {
			index = 1
		}
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: index, Offset: uint64(100 + pc)}})
		in := slots.Instruction{Op: op, Source: uint32(pc), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
		switch pc {
		case 0:
			in.Dst = 1
		case 1:
			in.A = 0
		case 2:
			in.Dst = 2
		case 3:
			in.Dst = 3
		case 4:
			in.Count = 2
		}
		f.Code = append(f.Code, in)
	}
	return f
}

func TestDirectGlobalsTypesAliasesAndEveryContinuation(t *testing.T) {
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.F32, wasm.F64, wasm.V128} {
		for _, unmetered := range []bool{false, true} {
			for _, tables := range []uint32{0, 3000} {
				f := globalFixture(typ)
				config := Config{Unmetered: unmetered, TableCount: tables, FunctionCount: 3,
					Globals: []module.GlobalType{{Value: typ, Mutable: true}, {Value: typ, Mutable: true}}}
				for _, arch := range []string{"amd64", "arm64"} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
					if err != nil || c.NativeGlobals != 4 || c.Helpers != 0 || len(c.GlobalIndices) != 2 || c.GlobalIndices[0] != 0 || c.GlobalIndices[1] != 1 || !c.UsesRuntimeData {
						t.Fatal(arch, typ, c, err)
					}
					limits := DefaultLimits()
					limits.MaxCodeBytes = uint64(len(c.Entry.Code) - 1)
					if _, err := CompileFunctionWithConfig(arch, f, limits, config); !errors.Is(err, wasm.ErrLimit) {
						t.Fatal("global code exceeded its budget", arch, err)
					}
				}
				if !native.Available {
					continue
				}
				_, lease, ctx := configuredContext(t, f, config)
				offset, _ := nativeabi.GlobalPointerOffset(3, tables, 0)
				data, err := native.NewMemory(uint64(offset)+32, uint64(offset)+32)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(data.Close)
				buffer, _ := data.Buffer()
				address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
				words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
				words[offset/8], words[offset/8+1] = address+uint64(offset)+16, address+uint64(offset)+16
				cell := words[offset/8+2:]
				for pc := range f.Code {
					initial := value.Value{Lo: 0xfff8123456789abc, Hi: 0xfedcba9876543210}
					argument := value.Value{Lo: 0xfff8abcdef012345, Hi: 0x123456789abcdef0}
					if typ == wasm.I32 || typ == wasm.F32 {
						initial.Lo, argument.Lo = 0xffc01234, 0xffc0abcd
					}
					if unmetered && typ != wasm.V128 {
						initial.Hi, argument.Hi = 0, 0
					}
					cell[0], cell[1] = initial.Lo, initial.Hi
					want := []value.Value{argument, {}, {}, {}}
					for i, v := range want {
						writeSlot(t, ctx, 16, slots.Slot(i), v)
					}
					if pc == 0 {
						want[1] = initial
					}
					if pc <= 1 {
						initial = argument
					}
					if pc <= 2 {
						want[2] = initial
					}
					if pc <= 3 {
						want[3] = initial
					}
					s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 64, Fuel: 10}
					s.Payload[2] = address
					if err := ctx.SetState(s); err != nil {
						t.Fatal(err)
					}
					if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
						t.Fatal(typ, unmetered, tables, pc, exit, err)
					}
					if cell[0] != initial.Lo || cell[1] != initial.Hi {
						t.Fatal("aliased global bits", typ, pc, cell, initial)
					}
					for i, v := range want {
						if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
							t.Fatal("global cold frame", typ, unmetered, tables, pc, i, got, v)
						}
					}
					// Even a return-only continuation must validate the runtime view.
					s.Payload[2] = 0
					ctx.SetState(s)
					if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.BadState {
						t.Fatal("missing global view", exit, err)
					}
				}
			}
		}
	}
}

func TestDirectGlobalsRetainCheckedFallbacks(t *testing.T) {
	for _, kind := range []string{"missing", "index", "immutable", "reference", "unknown", "type", "no-slot", "offset"} {
		f := globalFixture(wasm.I32)
		config := Config{Globals: []module.GlobalType{{Value: wasm.I32, Mutable: true}, {Value: wasm.I32, Mutable: true}}}
		switch kind {
		case "missing":
			config.Globals = nil
		case "index":
			f.Sources[0].Wasm.A = 100
		case "immutable":
			config.Globals[0].Mutable = false
		case "reference":
			config.Globals[0].Value = wasm.ExternRef
		case "unknown":
			config.Globals[0].Value = 0
		case "type":
			config.Globals[0].Value = wasm.I64
		case "no-slot":
			f.Code[0].Dst = slots.NoSlot
		case "offset":
			config.FunctionCount = ^uint32(0)
		}
		c, err := CompileFunctionWithConfig(runtime.GOARCH, f, DefaultLimits(), config)
		if err != nil || c.Helpers == 0 {
			t.Fatal(kind, c, err)
		}
	}
}

func TestDirectGlobalsUnderRegisterPressure(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, typ := range []wasm.ValueType{wasm.I64, wasm.F32, wasm.F64, wasm.V128} {
		f := globalFixture(typ)
		ret := f.Code[len(f.Code)-1]
		f.Code = f.Code[:len(f.Code)-1]
		appendOp := func(op uint32, dst, operand slots.Slot) {
			source := uint32(len(f.Sources))
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: 1}})
			f.Code = append(f.Code, slots.Instruction{Op: op, Source: source, Dst: dst, A: operand, B: slots.NoSlot, C: slots.NoSlot})
		}
		for range 16 {
			slot := slots.Slot(len(f.Slots))
			f.Slots = append(f.Slots, typ)
			appendOp(0x23, slot, slots.NoSlot)
		}
		f.LocalCount = uint32(len(f.Slots))
		// Exercise stores from both allocated high XMM registers and spills.
		for slot := slots.Slot(4); int(slot) < len(f.Slots); slot++ {
			appendOp(0x24, slots.NoSlot, slot)
		}
		f.Code = append(f.Code, ret)
		config := Config{Unmetered: true, Globals: []module.GlobalType{{Value: typ, Mutable: true}, {Value: typ, Mutable: true}}}
		_, lease, ctx := configuredContext(t, f, config)
		data, err := native.NewMemory(32, 32)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(data.Close)
		buffer, _ := data.Buffer()
		address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
		words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), 4)
		words[0], words[1] = address+16, address+16
		v := value.Value{Lo: 0xfedcba9876543210}
		if typ == wasm.F32 {
			v.Lo = 0xffc01234
		}
		if typ == wasm.V128 {
			v.Hi = 0x123456789abcdef0
		}
		for i := range f.Slots {
			writeSlot(t, ctx, 16, slots.Slot(i), value.Value{})
		}
		writeSlot(t, ctx, 16, 0, v)
		s := native.State{Used: 16, Depth: 1}
		s.Payload[2] = address
		if err := ctx.SetState(s); err != nil {
			t.Fatal(err)
		}
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
			t.Fatal(exit, err)
		}
		if words[2] != v.Lo || words[3] != v.Hi {
			t.Fatal("spilled global store", typ, words, v)
		}
		for slot := slots.Slot(2); int(slot) < len(f.Slots); slot++ {
			if got := readSlot(t, ctx, 16, slot); got != v {
				t.Fatal("spilled global load", typ, slot, got, v)
			}
		}
	}
}
