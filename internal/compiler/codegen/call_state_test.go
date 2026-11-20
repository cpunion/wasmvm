package codegen

import (
	"bytes"
	"encoding/binary"
	"errors"
	"runtime"
	"slices"
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

func TestNativeFrameRegistersRequireUnmeteredLinkedBodies(t *testing.T) {
	f := leafFixture(0x6a, nil)
	for _, arch := range []string{"amd64", "arm64"} {
		for _, config := range []Config{
			{NativeCallLimits: true},
			{NativeCallLimits: true, NativeFrameRegisters: true, Unmetered: true, LinkedCallees: map[uint32]*slots.Function{7: f}},
			{NativeFrameRegisters: true},
			{NativeFrameRegisters: true, Unmetered: true},
			{NativeFrameRegisters: true, LinkedCallees: map[uint32]*slots.Function{7: f}},
			{NativeFunction: 1, Unmetered: true},
			{NativeFunction: 1<<32 + 1, Unmetered: true, NativeFrameRegisters: true, LinkedCallees: map[uint32]*slots.Function{7: f}},
		} {
			if _, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(arch, config, err)
			}
		}
	}
}

func TestGuardedLoopPreservesLinkedCallState(t *testing.T) {
	f := coalescedValueLoopFixture(false)
	tail := slices.Clone(f.Code[len(f.Code)-4:])
	f.Code = f.Code[:len(f.Code)-4]
	source := uint32(len(f.Sources))
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x6a}})
	// Keep independent live-in integers under pressure across the backedge.
	for range 24 {
		s := slots.Slot(len(f.Slots))
		f.Slots = append(f.Slots, wasm.I32)
		f.Args = append(f.Args, s)
		f.Code = append(f.Code, slots.Instruction{Op: 0x6a, Source: source, Dst: s, A: s, B: 1, C: slots.NoSlot})
	}
	f.LocalCount = uint32(len(f.Slots))
	tail[len(tail)-1].Count = uint32(len(f.Args))
	f.Code = append(f.Code, tail...)
	f.Edges[1].PC = uint32(len(f.Code) - 1)

	config := Config{Unmetered: true, MemoryMinimumPages: 1, NativeFrameRegisters: true, NativeFunction: 2, LinkedCallees: map[uint32]*slots.Function{}}
	for _, arch := range []string{"amd64", "arm64"} {
		compiled, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
		if err != nil {
			t.Fatal(arch, err)
		}
		save := binary.LittleEndian.AppendUint32(nil, 0xa9be73fb)
		save = binary.LittleEndian.AppendUint32(save, 0xf9000bfa)
		if bytes.Contains(compiled.Entry.Code, save) != (arch == "arm64") {
			t.Fatal(arch, "unexpected call-state reservation")
		}
	}
	if !native.Available || runtime.GOARCH != "arm64" {
		return
	}
	compiled, run, ctx := configuredContext(t, f, config)
	pollCode, err := native.Publish(forcePrivateLoopPoll(t, compiled, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pollCode.Close)
	poll, err := pollCode.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(poll.Close)
	_, reference, ref := functionContext(t, f)
	memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
	const viewBytes = nativeabi.CallHeaderBytes + nativeabi.CallFrameBytes
	metadata, err := native.NewMemory(viewBytes, viewBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(metadata.Close)
	data, err := metadata.Buffer()
	if err != nil {
		t.Fatal(err)
	}
	view := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(data))))
	for _, tc := range []struct {
		entry        *native.Lease
		address, top uint64
		budget       uint64
		want         native.Exit
	}{
		{run, 16, 0, 1000, native.Return},
		{poll, 65536 - 48, 1, uint64(len(f.Code) - 2), native.Poll},
		{run, 65536 - 47, 1, 1000, native.Trap},
	} {
		binary.LittleEndian.PutUint64(data, tc.top)
		binary.LittleEndian.PutUint64(data[8:], 1)
		for i := range f.Slots {
			v := value.Value{Lo: uint64(100 + i)}
			if i == 0 {
				v.Lo = tc.address
			} else if i == 1 {
				v.Lo = 0
			}
			writeSlot(t, ctx, 16, slots.Slot(i), v)
			writeSlot(t, ref, 16, slots.Slot(i), v)
		}
		state := native.State{Used: 16, Depth: tc.top + 1, Fuel: 42, Budget: tc.budget, Payload: [4]uint64{0, view, view}}
		oracle := state
		got, err := ctx.RunState(tc.entry, 0, &state)
		want, refErr := ref.RunState(reference, 0, &oracle)
		if err != nil || refErr != nil || got != tc.want || got != want {
			t.Fatal(got, want, tc.want, err, refErr)
		}
		if state.Used != 16 || state.Depth != tc.top+1 || state.Function != 1 || state.Fuel != 42 || binary.LittleEndian.Uint64(data) != tc.top {
			t.Fatal(got, "call state", state, binary.LittleEndian.Uint64(data))
		}
		if state.Continuation != oracle.Continuation || state.Payload[0] != oracle.Payload[0] || state.Source != oracle.Source {
			t.Fatal(got, "continuation and trap state", state, oracle)
		}
		for i := range f.Slots {
			if a, b := readSlot(t, ctx, 16, slots.Slot(i)), readSlot(t, ref, 16, slots.Slot(i)); a != b {
				t.Fatal(got, "slot", i, a, b)
			}
		}
		var actual, expected [48]byte
		offset := min(tc.address, 65536-48)
		if err := memory.Read(offset, actual[:]); err != nil {
			t.Fatal(err)
		}
		if err := refMemory.Read(offset, expected[:]); err != nil || actual != expected {
			t.Fatal(got, actual, expected, err)
		}
	}
}

func TestNativeFrameStateIsExcludedFromGuestAllocation(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, memory := range []bool{false, true} {
			e := functionEmitter{a: &assembler{arch: arch}, usesMemory: memory}
			ordinary := slices.Clone(e.integerRegisterPool())
			e.frameRegisters = true
			reserved := []byte{e.callReg(callUsed), e.callReg(callDepth), e.callReg(callTop)}
			want := slices.DeleteFunc(ordinary, func(r byte) bool { return slices.Contains(reserved, r) })
			if got := e.integerRegisterPool(); !slices.Equal(got, want) {
				t.Fatal(arch, memory, got, want)
			}
			for _, r := range e.spareCallerRegisters() {
				if slices.Contains(reserved, r) {
					t.Fatal("frame state overlaps leaf scratch", arch, r)
				}
			}
		}
	}
}
