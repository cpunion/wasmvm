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

func heapFixture(field module.FieldType, array, signed bool) (*slots.Function, Config) {
	typ := field.Unpacked().Storage()
	f := &slots.Function{Slots: []wasm.ValueType{wasm.AnyRef, wasm.I32, typ, typ, typ, wasm.I32}, LocalCount: 6,
		Signature: module.FuncType{Params: []wasm.ValueType{wasm.AnyRef, wasm.I32, typ}, Results: []wasm.ValueType{typ}}}
	t := module.TypeDef{Kind: module.StructKind, Fields: []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I64}, field}}
	get, set, index := uint32(0xfb0002), uint32(0xfb0005), uint32(2)
	if array {
		t.Kind, t.Fields = module.ArrayKind, []module.FieldType{field}
		get, set, index = 0xfb000b, 0xfb000e, 0
	}
	if field.Packed != 0 {
		get += 2
		if signed {
			get--
		}
	}
	for pc, op := range []uint32{get, set, get} {
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, B: index, Offset: uint64(100 + pc)}})
		in := slots.Instruction{Op: op, Source: uint32(pc), A: 0, B: 1, C: slots.NoSlot, Dst: slots.NoSlot}
		if op == set {
			if !array {
				in.B = 2
			}
		} else {
			in.Dst = 3
			if pc == 2 {
				in.Dst = 4
			}
		}
		in.Extra = uint32(len(f.Args))
		f.Args = append(f.Args, 0)
		if array {
			f.Args = append(f.Args, 1)
		}
		if op == set {
			f.Args = append(f.Args, 2)
		}
		in.Count = uint32(len(f.Args)) - in.Extra
		f.Code = append(f.Code, in)
	}
	if array {
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0xfb000f, Offset: 103}})
		f.Code = append(f.Code, slots.Instruction{Op: 0xfb000f, Source: 3, A: 0, B: slots.NoSlot, C: slots.NoSlot, Dst: 5, Extra: 0, Count: 1})
	}
	f.Sources = append(f.Sources, ir.Source{})
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: uint32(len(f.Sources) - 1),
		A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Extra: uint32(len(f.Args)), Count: 1})
	f.Args = append(f.Args, 4)
	return f, Config{HeapTypes: []module.TypeDef{t}}
}

func TestNativeHeapFieldsEveryWidthAndContinuation(t *testing.T) {
	fields := []module.FieldType{{Value: wasm.I32, Packed: 8}, {Value: wasm.I32, Packed: 16},
		{Value: wasm.I32}, {Value: wasm.I64}, {Value: wasm.F32}, {Value: wasm.F64}, {Value: wasm.AnyRef}, {Value: wasm.V128}}
	for _, field := range fields {
		for _, array := range []bool{false, true} {
			for _, signed := range []bool{false, true} {
				for _, unmetered := range []bool{false, true} {
					f, config := heapFixture(field, array, signed)
					config.Unmetered = unmetered
					config.FunctionCount, config.TableCount, config.MemoryCount = 3, 3000, 2
					for _, arch := range []string{"amd64", "arm64"} {
						c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
						if err != nil || c.Helpers != 0 || c.NativeHeap != uint32(len(f.Code)-1) || !c.UsesRuntimeData {
							t.Fatal(arch, field, array, signed, c, err)
						}
						if (c.HeapCaches != 0) != (unmetered && arch == "arm64") {
							t.Fatal("heap cache policy", arch, unmetered, c.HeapCaches)
						}
						if (c.ForwardedHeap == 1) != unmetered {
							t.Fatal("heap forwarding policy", arch, unmetered, c.ForwardedHeap)
						}
					}
					if !native.Available {
						continue
					}
					_, lease, ctx := configuredContext(t, f, config)
					pair, _ := nativeabi.HeapPairOffset(3, 3000, 0, 0, 2)
					storage, err := native.NewMemory(uint64(pair)+256, uint64(pair)+256)
					if err != nil {
						t.Fatal(err)
					}
					defer storage.Close()
					buffer, _ := storage.Buffer()
					address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
					words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
					words[pair/8], words[pair/8+1] = 2, address+uint64(pair)+16
					view := words[pair/8+6:]
					ref := uint64(0x8000000100000001)
					view[0], view[1], view[2] = ref, address+uint64(pair)+128, 2
					width, offset := field.HeapBytes(), 9
					if array {
						offset = width
					}
					cell := buffer[int(pair)+128+offset:]
					mask := func(v value.Value) value.Value {
						if width < 8 {
							v.Lo &= (uint64(1) << (width * 8)) - 1
						}
						if width != 16 {
							v.Hi = 0
						}
						return v
					}
					unpack := func(v value.Value) value.Value {
						if signed && field.Packed != 0 {
							v.Lo = uint64(uint32(int32(uint32(v.Lo)<<(32-field.Packed)) >> (32 - field.Packed)))
						}
						return v
					}
					for pc := range f.Code {
						initial := mask(value.Value{Lo: 0x89abcdef87658081, Hi: 0x123456789abcdef0})
						argument := mask(value.Value{Lo: 0xfedcba987654def0, Hi: 0xf0e1d2c3b4a59687})
						binary.LittleEndian.PutUint64(cell, initial.Lo)
						binary.LittleEndian.PutUint64(cell[8:], initial.Hi)
						want := []value.Value{{Lo: ref}, {Lo: 0x100000001}, argument, {}, {}, {}}
						// i32 indices truncate even on cold continuations.
						for i, v := range want {
							writeSlot(t, ctx, 16, slots.Slot(i), v)
						}
						if pc == 0 {
							want[3] = unpack(initial)
							if array {
								want[1].Lo = 0x100000001
							}
						}
						if pc <= 1 {
							initial = argument
						}
						if pc <= 2 {
							want[4] = unpack(initial)
						}
						if array && pc <= 3 {
							want[5].Lo = 2
						}
						s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 64}
						s.Payload[2] = address
						ctx.SetState(s)
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(field, array, signed, unmetered, pc, exit, err)
						}
						for i, v := range want {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != v {
								t.Fatal("heap cold frame", field, array, signed, unmetered, pc, i, got, v)
							}
						}
						if got := mask(value.Value{Lo: binary.LittleEndian.Uint64(cell), Hi: binary.LittleEndian.Uint64(cell[8:])}); got != initial {
							t.Fatal("heap store lost bits")
						}
					}
					for _, bad := range []struct {
						ref, index uint64
						exit       native.Exit
						trap       wasm.TrapCode
					}{{0, 0, native.Trap, wasm.TrapNullReference}, {ref + 1, 0, native.BadState, 0},
						{ref + 1<<32, 0, native.BadState, 0}, {ref &^ (1 << 63), 0, native.BadState, 0},
						{ref, 2, native.Trap, wasm.TrapArrayBounds}, {ref, 0xffffffff, native.Trap, wasm.TrapArrayBounds}} {
						if !array && bad.trap == wasm.TrapArrayBounds {
							continue
						}
						writeSlot(t, ctx, 16, 0, value.Value{Lo: bad.ref})
						writeSlot(t, ctx, 16, 1, value.Value{Lo: bad.index})
						s := native.State{Used: 16, Depth: 1, Budget: 64}
						s.Payload[2] = address
						ctx.SetState(s)
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != bad.exit {
							t.Fatal("unchecked handle/index", bad, exit, err)
						}
						state, _ := ctx.State()
						if bad.trap != 0 && (state.Payload[0] != uint64(bad.trap) || state.Source != 100) {
							t.Fatal("heap trap source", bad, state)
						}
					}
				}
			}
		}
	}
}

func TestNativeHeapForwardingPreservesReusedSlotBits(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	const raw = uint64(0xffffbeef00008081)
	for _, tc := range []struct {
		field  module.FieldType
		signed bool
		want   value.Value
	}{
		{module.FieldType{Value: wasm.I32, Packed: 8}, false, value.Value{Lo: 0x81}},
		{module.FieldType{Value: wasm.I32, Packed: 8}, true, value.Value{Lo: 0xffffff81}},
		{module.FieldType{Value: wasm.I32, Packed: 16}, false, value.Value{Lo: 0x8081}},
		{module.FieldType{Value: wasm.I32, Packed: 16}, true, value.Value{Lo: 0xffff8081}},
		{module.FieldType{Value: wasm.I32}, false, value.Value{Lo: 0x8081}},
		{module.FieldType{Value: wasm.F32}, false, value.Value{Lo: 0x8081}},
		{module.FieldType{Value: wasm.I64}, false, value.Value{Lo: raw}},
		{module.FieldType{Value: wasm.F64}, false, value.Value{Lo: raw}},
		{module.FieldType{Value: wasm.AnyRef}, false, value.Value{Lo: raw}},
		{module.FieldType{Value: wasm.V128}, false, value.Value{Lo: raw, Hi: 0x87654321}},
	} {
		f, config := heapFixture(tc.field, false, tc.signed)
		config.Unmetered = true
		config.HeapTypes[0].Fields = config.HeapTypes[0].Fields[2:]
		for i := range f.Sources {
			f.Sources[i].Wasm.B = 0
		}
		f.Code[2].Dst, f.Args[len(f.Args)-1] = 2, 2
		compiled, lease, ctx := configuredContext(t, f, config)
		if compiled.ForwardedHeap != 1 {
			t.Fatal("store/load was not forwarded")
		}
		storage, err := native.NewMemory(256, 256)
		if err != nil {
			t.Fatal(err)
		}
		defer storage.Close()
		buffer, _ := storage.Buffer()
		address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
		words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
		ref := uint64(0x8000000100000001)
		words[0], words[1], words[6], words[7], words[8] = 2, address+16, ref, address+128, 1
		input := value.Value{Lo: raw}
		if tc.field.Value == wasm.V128 {
			input.Hi = tc.want.Hi
		}
		writeSlot(t, ctx, 16, 0, value.Value{Lo: ref})
		writeSlot(t, ctx, 16, 2, input)
		s := native.State{Used: 16, Depth: 1, Budget: 64}
		s.Payload[2] = address
		ctx.SetState(s)
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
			t.Fatal(exit, err)
		}
		if got := readSlot(t, ctx, 16, 2); got != tc.want {
			t.Fatal("reused field value lost bits", tc.field, tc.signed, got, tc.want)
		}
	}
}

func TestNativeHeapCacheRechecksChangedReferences(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	f, config := heapFixture(module.FieldType{Value: wasm.I32}, false, false)
	config.Unmetered = true
	config.HeapTypes[0].Fields = config.HeapTypes[0].Fields[2:]
	for i := range f.Sources {
		f.Sources[i].Wasm.B = 0
	}
	f.Slots = append(f.Slots, wasm.AnyRef)
	f.LocalCount++
	source := uint32(len(f.Sources))
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: 6}})
	move := slots.Instruction{Op: 0x20, Source: source, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
	f.Code = append(f.Code[:1:1], append([]slots.Instruction{move}, f.Code[1:]...)...)
	_, lease, ctx := configuredContext(t, f, config)
	storage, err := native.NewMemory(256, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	buffer, _ := storage.Buffer()
	address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
	words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), len(buffer)/8)
	words[0], words[1] = 3, address+16
	first, second := uint64(0x8000000100000001), uint64(0x8000000100000002)
	words[6], words[7], words[8] = first, address+128, 3
	words[10], words[11], words[12] = second, address+192, 3
	binary.LittleEndian.PutUint32(buffer[128:], 11)
	binary.LittleEndian.PutUint32(buffer[192:], 22)
	for i, v := range []value.Value{{Lo: first}, {}, {Lo: 777}, {}, {}, {}, {Lo: second}} {
		writeSlot(t, ctx, 16, slots.Slot(i), v)
	}
	s := native.State{Used: 16, Depth: 1, Budget: 64}
	s.Payload[2] = address
	ctx.SetState(s)
	if exit, err := ctx.Enter(lease, 0); exit != native.Return || err != nil {
		t.Fatal(exit, err)
	}
	if readSlot(t, ctx, 16, 3).Lo != 11 || readSlot(t, ctx, 16, 4).Lo != 777 || binary.LittleEndian.Uint32(buffer[128:]) != 11 || binary.LittleEndian.Uint32(buffer[192:]) != 777 {
		t.Fatal("changed handle reused a former object's payload")
	}
}

func TestNativeHeapCheckedFallbackAndOffsetLimits(t *testing.T) {
	for _, kind := range []string{"missing", "struct-kind", "struct-field", "array-kind", "array-field", "width", "pair"} {
		f, config := heapFixture(module.FieldType{Value: wasm.I32}, kind == "array-kind" || kind == "array-field", false)
		switch kind {
		case "missing":
			config.HeapTypes = nil
		case "struct-kind", "array-kind":
			config.HeapTypes[0].Kind = module.FunctionKind
		case "struct-field":
			f.Sources[0].Wasm.B = 99
		case "array-field":
			config.HeapTypes[0].Fields = nil
		case "width":
			config.HeapTypes[0].Fields[2].Value = 0
		case "pair":
			config.FunctionCount = ^uint32(0)
		}
		c, err := CompileFunctionWithConfig(runtime.GOARCH, f, DefaultLimits(), config)
		if kind == "pair" {
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(kind, err)
			}
		} else if err != nil || c.Helpers == 0 {
			t.Fatal(kind, c, err)
		}
	}
}
