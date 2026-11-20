package codegen

import (
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestStructLoopGuardPreservesPrefixAndTrapPC(t *testing.T) {
	f, config := heapFixture(module.FieldType{Value: wasm.I32}, false, false)
	config.Unmetered = true
	// The guest updates a scalar before touching its first field. A failed
	// guard must execute that update and trap at the original field access.
	source := uint32(len(f.Sources))
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x6a, Offset: 55}})
	prefix := slots.Instruction{Op: 0x6a, Source: source, Dst: 1, A: 1, B: 2, C: slots.NoSlot}
	f.Code = append([]slots.Instruction{prefix}, f.Code...)
	addHeapFixtureLoop(f, 0)
	for _, arch := range []string{"amd64", "arm64"} {
		c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
		if err != nil || c.GuardedStructLoops != 1 || len(c.PrivatePollOffsets) != 1 {
			t.Fatalf("%s: missing guarded loop: %v, %v", arch, c, err)
		}
		metered := config
		metered.Unmetered = false
		c, err = CompileFunctionWithConfig(arch, f, DefaultLimits(), metered)
		if err != nil || c.GuardedStructLoops != 0 {
			t.Fatalf("%s: changed precise fuel tier: %v, %v", arch, c, err)
		}
		mutating := *f
		mutating.Code = append([]slots.Instruction(nil), f.Code...)
		mutating.Sources = append(append([]ir.Source(nil), f.Sources...), ir.Source{Wasm: module.Instr{Op: 0xd0}})
		mutating.Code[0] = slots.Instruction{Op: 0xd0, Source: uint32(len(f.Sources)), Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
		c, err = CompileFunctionWithConfig(arch, &mutating, DefaultLimits(), config)
		if err != nil || c.GuardedStructLoops != 0 {
			t.Fatalf("%s: reused a changed reference: %v, %v", arch, c, err)
		}
	}
	if !native.Available {
		return
	}
	_, lease, ctx := configuredContext(t, f, config)
	storage, err := native.NewMemory(256, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	buf, _ := storage.Buffer()
	address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buf))))
	words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buf))), len(buf)/8)
	words[0], words[1] = 2, address+16
	const handle = 0x8000000100000001
	words[6], words[7] = handle, address+128
	for _, tc := range []struct {
		pc, reference, scalar uint64
		exit                  native.Exit
	}{
		{0, 0, 18, native.Trap},
		{1, 0, 7, native.Trap},
		{0, handle, 18, native.Return},
		{1, handle, 7, native.Return},
		{0, handle + 1<<32, 0, native.BadState},
	} {
		for i, lo := range []uint64{tc.reference, 7, 11, 0, 0, 0, 0} {
			writeSlot(t, ctx, 16, slots.Slot(i), value.Value{Lo: lo})
		}
		s := native.State{Used: 16, Depth: 1, Continuation: tc.pc}
		s.Payload[2] = address
		ctx.SetState(s)
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != tc.exit {
			t.Fatalf("%+v: exit=%v error=%v", tc, exit, err)
		}
		if tc.exit != native.BadState && readSlot(t, ctx, 16, 1).Lo != tc.scalar {
			t.Fatalf("prefix state lost at PC %d", tc.pc)
		}
		if tc.exit == native.Trap {
			s, _ = ctx.State()
			if s.Source != 100 || s.Payload[0] != uint64(wasm.TrapNullReference) {
				t.Fatalf("wrong field trap: %+v", s)
			}
		}
	}
}

func TestStructLoopForwardingKeepsColdI32Narrowing(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	f, config := heapFixture(module.FieldType{Value: wasm.I32}, false, false)
	config.Unmetered = true
	// Arithmetic supplies the store and the following load reuses its slot.
	// A store-PC entry skips arithmetic and must still zero-extend the load.
	source := uint32(len(f.Sources))
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x6a}})
	add := slots.Instruction{Op: 0x6a, Source: source, Dst: 2, A: 1, B: 2, C: slots.NoSlot}
	f.Code = append(f.Code[:1:1], append([]slots.Instruction{add}, f.Code[1:]...)...)
	f.Code[3].Dst = 2
	f.Args[len(f.Args)-1] = 2
	addHeapFixtureLoop(f, 0)
	compiled, lease, ctx := configuredContext(t, f, config)
	if compiled.GuardedStructLoops != 1 || compiled.ForwardedHeap != 1 {
		t.Fatal("missing guarded store/load forwarding")
	}
	storage, err := native.NewMemory(256, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	buf, _ := storage.Buffer()
	address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buf))))
	words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buf))), len(buf)/8)
	const handle = 0x8000000100000001
	words[0], words[1], words[6], words[7] = 2, address+16, handle, address+128
	for _, tc := range []struct{ pc, want uint64 }{{0, 4}, {2, 3}} {
		writeSlot(t, ctx, 16, 0, value.Value{Lo: handle})
		writeSlot(t, ctx, 16, 1, value.Value{Lo: 1})
		writeSlot(t, ctx, 16, 2, value.Value{Lo: 0x1234567800000003})
		writeSlot(t, ctx, 16, 6, value.Value{})
		s := native.State{Used: 16, Depth: 1, Continuation: tc.pc}
		s.Payload[2] = address
		ctx.SetState(s)
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
			t.Fatal(tc.pc, exit, err)
		}
		if got := readSlot(t, ctx, 16, 2); got != (value.Value{Lo: tc.want}) {
			t.Fatalf("PC %d: field result %v, want %#x", tc.pc, got, tc.want)
		}
	}
}
