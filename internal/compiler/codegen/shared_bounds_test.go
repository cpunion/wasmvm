package codegen

import (
	"bytes"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// The descriptor and its size cell have separate native owners, retained until
// the context has stopped. No Go address is written into either ABI record.
func bindSharedTestMemory(t *testing.T, ctx *native.Context, size, maximum uint64, missingBase ...bool) (*native.Memory, uint64) {
	t.Helper()
	m, err := native.NewSharedMemory(size, maximum)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	if err := ctx.BindMemory(m); err != nil {
		t.Fatal(err)
	}
	base, length, _, err := m.SharedView()
	if err != nil {
		t.Fatal(err)
	}
	view, err := native.NewMemory(16, 16)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(view.Close)
	buffer, err := view.Buffer()
	if err != nil {
		t.Fatal(err)
	}
	words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(buffer))), 2)
	words[0], words[1] = uint64(uintptr(length)), uint64(uintptr(base))
	if len(missingBase) != 0 && missingBase[0] {
		words[1] = 0
	}
	return m, uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
}

func TestSharedMemoryGuardsAllWidthsAndColdEntries(t *testing.T) {
	ops := []uint32{0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2f, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0xfd0000, 0xfd000b, 0xfd0054, 0xfd005b}
	for _, op := range ops {
		t.Run(fmt.Sprintf("%x", op), func(t *testing.T) {
			f := guardedMemoryFunction(t, op, 0, []uint32{0, 0, 0}, []uint32{0, 1, 2})
			for _, arch := range []string{"amd64", "arm64"} {
				cfg := Config{Unmetered: true, NativeThreads: true, SharedMemories: []bool{true}, MemoryCount: 1, MemoryMinimumPages: 1}
				compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
				if err != nil || compiled.BoundsElided != 3 || compiled.MemoryGuards != 1 || !compiled.UsesRuntimeData {
					t.Fatal(arch, compiled, err)
				}
				// Mixed memories must not inherit memory zero's entry base.
				cfg.MemoryCount, cfg.SharedMemories = 2, []bool{true, true}
				for _, in := range f.Code {
					if isMemoryAccess(in.Op) {
						f.Sources[in.Source].Wasm.Memory = 1
						break
					}
				}
				mixed, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
				if err != nil || mixed.BoundsElided != 0 {
					t.Fatal("mixed memory base reused", arch, mixed, err)
				}
				for i := range f.Sources {
					f.Sources[i].Wasm.Memory = 0
				}
			}
			_, size := memoryAccess(op)
			compareGuardedExecution(t, f, []uint64{0, 1, 65536 - uint64(size) - 2, 65536 - uint64(size) - 1, 65536 - uint64(size), 0xffffffff, 1 << 32}, true)
		})
	}
}

func TestSharedDefaultMemoryValidatesEntryMinimumAndBase(t *testing.T) {
	if !native.Available {
		return
	}
	f, _ := memoryFunction(t, 0x28, 0, true)
	for _, kind := range []string{"short", "missing-base"} {
		t.Run(kind, func(t *testing.T) {
			cfg := Config{Unmetered: true, NativeThreads: true, SharedMemories: []bool{true}, MemoryCount: 1, MemoryMinimumPages: 1}
			_, lease, ctx := configuredContext(t, f, cfg)
			size := uint64(65536)
			if kind == "short" {
				size--
			}
			_, view := bindSharedTestMemory(t, ctx, size, 65536, kind == "missing-base")
			state := native.State{Used: 16, Depth: 1}
			state.Payload[2] = view
			if err := ctx.SetState(state); err != nil {
				t.Fatal(err)
			}
			if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.BadState {
				t.Fatal("invalid shared binding admitted", kind, exit, err)
			}
		})
	}
}

func TestSharedMemoryLoopFramesAndEffectsAtEveryEntry(t *testing.T) {
	for _, unroll := range []bool{false, true} {
		f := memoryLoopFixture(4, false, false)
		inputs := [][3]uint64{{0, 0, 0}, {65535, 0, 0}, {0xffffffff, 0, 0}}
		if unroll {
			f = unrollFixture(0, false, false, 0)
			inputs = [][3]uint64{{0, 17, 0}, {1, 17, 0}, {0xffffffff, 17, 0}}
		}
		cfg := Config{Unmetered: true, NativeThreads: true, SharedMemories: []bool{true}, MemoryCount: 1, MemoryMinimumPages: 1}
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
			if err != nil || unroll && c.UnrolledLoops != 1 || !unroll && c.LoopGuards != 1 {
				t.Fatal("shared loop was not exercised", arch, unroll, c, err)
			}
		}
		if !native.Available {
			continue
		}
		_, lease, ctx := configuredContext(t, f, cfg)
		cfg.MemoryMinimumPages = 0
		_, refLease, ref := configuredContext(t, f, cfg)
		mem, view := bindSharedTestMemory(t, ctx, 65536, 65536)
		refMem, refView := bindSharedTestMemory(t, ref, 65536, 65536)
		initial, got, want := make([]byte, 65536), make([]byte, 65536), make([]byte, 65536)
		for pc := range f.Code {
			for _, input := range inputs {
				if err := mem.Write(0, initial); err != nil {
					t.Fatal(err)
				}
				if err := refMem.Write(0, initial); err != nil {
					t.Fatal(err)
				}
				for slot := range f.Slots {
					bits := uint64(0)
					if slot < 3 {
						bits = input[slot]
					}
					writeSlot(t, ctx, 16, slots.Slot(slot), value.Value{Lo: bits})
					writeSlot(t, ref, 16, slots.Slot(slot), value.Value{Lo: bits})
				}
				state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
				state.Payload[2] = view
				gs, ge, err := ctx.Run(lease, 0, state)
				if err != nil {
					t.Fatal(err)
				}
				state.Payload[2] = refView
				ws, we, err := ref.Run(refLease, 0, state)
				gs.Payload[2], ws.Payload[2] = 0, 0
				if err != nil || ge != we || gs != ws {
					t.Fatal("shared loop snapshot", unroll, pc, input, ge, we, gs, ws, err)
				}
				for slot := range f.Slots {
					if readSlot(t, ctx, 16, slots.Slot(slot)) != readSlot(t, ref, 16, slots.Slot(slot)) {
						t.Fatal("shared loop frame", unroll, pc, input, slot)
					}
				}
				if err := mem.Read(0, got); err != nil {
					t.Fatal(err)
				}
				if err := refMem.Read(0, want); err != nil || !bytes.Equal(got, want) {
					t.Fatal("shared loop effects", unroll, pc, input, err)
				}
			}
		}
	}
}

func TestSharedGrowthIsObservedWithinOneNativeEntry(t *testing.T) {
	if !native.Available {
		return
	}
	m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32}}, 0, module.Expr{
		{Op: 0x41}, {Op: 0x41, Bits: 1}, {Op: 0xfe0017, A: 2},
		{Op: 3, Block: -64}, {Op: 0x3f}, {Op: 0x41, Bits: 1}, {Op: 0x46}, {Op: 0xd}, {Op: 0xb},
		{Op: 0x41, Bits: 65536}, {Op: 0x28, A: 2}, {Op: 0xb},
	})
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true, Shared: true}}
	f := functionModule(t, m, wasm.Threads).Functions[0]
	cfg := Config{Unmetered: true, NativeThreads: true, SharedMemories: []bool{true}, MemoryCount: 1, MemoryMinimumPages: 1}
	_, lease, ctx := configuredContext(t, f, cfg)
	memory, view := bindSharedTestMemory(t, ctx, 65536, 131072)
	base, _, _, err := memory.SharedView()
	if err != nil {
		t.Fatal(err)
	}
	state := native.State{Used: 16, Depth: 1}
	state.Payload[2] = view
	if err := ctx.SetState(state); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	var exit native.Exit
	var entryErr error
	go func() { exit, entryErr = ctx.Enter(lease, 0); close(done) }()
	defer func() {
		ctx.Cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("shared-growth native entry did not stop")
		}
	}()
	deadline := time.Now().Add(5 * time.Second)
	for atomic.LoadUint32((*uint32)(base)) == 0 {
		select {
		case <-done:
			t.Fatal("entry returned before growth", exit, entryErr)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("native entry did not publish its atomic marker")
		}
		runtime.Gosched()
	}
	if err := memory.Grow(131072); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("memory.size retained a stale shared length")
	}
	if entryErr != nil || exit != native.Return {
		t.Fatal("new-page access retained a stale shared length", exit, entryErr)
	}
	s, _ := ctx.State()
	ret := f.Code[s.Payload[0]]
	if got := readSlot(t, ctx, 16, slots.Slot(f.Args[ret.Extra])); got.Lo != 0 || got.Hi != 0 {
		t.Fatal("newly grown page was not zeroed", got)
	}
}
