package codegen

import (
	"encoding/binary"
	"errors"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLinkedCallBodiesAndRelocations(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, unmetered := range []bool{false, true} {
			callee := leafFixture(0x6a, nil)
			for range 700 {
				callee.Slots = append(callee.Slots, wasm.I64)
			}
			caller := leafCaller(callee, 0, 5)
			config := Config{Unmetered: unmetered, LinkedCallees: map[uint32]*slots.Function{7: callee}}
			bodies := make([]*CompiledFunction, 2)
			for i, f := range []*slots.Function{caller, callee} {
				var err error
				bodies[i], err = CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
				if err != nil {
					t.Fatal(arch, unmetered, err)
				}
			}
			if len(bodies[0].CallRelocations) != 1 {
				t.Fatal("missing linked call", arch)
			}
			code := append(append([]byte{}, bodies[0].Entry.Code...), bodies[1].Entry.Code...)
			target := uint32(len(bodies[0].Entry.Code)) + bodies[1].LinkedEntry
			if err := LinkCalls(arch, code, bodies[0].CallRelocations, map[uint32]uint32{7: target}); err != nil {
				t.Fatal(err)
			}
			if arch != runtime.GOARCH || !native.Available {
				continue
			}
			published, err := native.Publish(code, []uint32{0}, uint64(len(code)))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(published.Close)
			lease, err := published.Acquire()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(lease.Close)
			ctx, err := native.NewContext(2048*16, 2048*16, 64)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(ctx.Close)
			metadata, err := native.NewMemory(nativeabi.CallHeaderBytes+16*nativeabi.CallFrameBytes, nativeabi.CallHeaderBytes+16*nativeabi.CallFrameBytes)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(metadata.Close)
			data, err := metadata.Buffer()
			if err != nil {
				t.Fatal(err)
			}
			words := unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(data))), len(data)/8)
			view := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(data))))
			for _, test := range []struct {
				capacity, slots uint64
				linked          bool
			}{{16, 2048, true}, {0, 2048, false}, {16, 5, false}} {
				clear(words)
				words[1], words[2], words[3] = test.capacity, test.slots*16, 64
				for i := range 2048 {
					writeSlot(t, ctx, 0, slots.Slot(i), value.Value{Lo: 99, Hi: 88})
				}
				initial := []value.Value{{Lo: 7}, {Lo: 19}, {Lo: 17}, {Lo: 19}, {Lo: 23}}
				for i, v := range initial {
					writeSlot(t, ctx, 0, slots.Slot(i), v)
				}
				state := native.State{Depth: 1, Fuel: 100, Budget: 64, Payload: [4]uint64{0, view, view, 0}}
				if !unmetered {
					state.Payload[3] = 1
				}
				exit, err := ctx.RunState(lease, 0, &state)
				if err != nil {
					t.Fatal(arch, unmetered, test, err)
				}
				if !test.linked {
					if exit != native.Call || words[0] != 0 {
						t.Fatal("missing bounded fallback", exit, state, words[0])
					}
					continue
				}
				if exit != native.Return || state.Function != 0 || state.Depth != 1 || words[0] != 0 || readSlot(t, ctx, 0, 0).Lo != 26 {
					t.Fatal("linked guest return", unmetered, exit, state, words[:12], readSlot(t, ctx, 0, 0))
				}
				for i := 5; i < len(callee.Slots); i++ {
					if got := readSlot(t, ctx, uint64(len(caller.Slots))*16, slots.Slot(i)); got != (value.Value{}) {
						t.Fatal("uncleared child", i, got)
					}
				}
				if unmetered && state.Fuel != 100 {
					t.Fatal("unmetered fuel", state.Fuel)
				}
			}
		}
	}
}

func TestCallRelocationRejectsInvalidTargetsWithoutPublication(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		code := make([]byte, 20)
		if arch == "amd64" {
			code[0] = 0xe9
		} else {
			binary.LittleEndian.PutUint32(code, 0x14000000)
		}
		for _, test := range []struct {
			offset, target uint32
			missing        bool
		}{{0, 20, false}, {20, 0, false}, {0, 4, true}, {1, 4, false}} {
			before := append([]byte(nil), code...)
			targets := map[uint32]uint32{7: test.target}
			if test.missing {
				delete(targets, 7)
			}
			if err := LinkCalls(arch, code, []CallRelocation{{Offset: test.offset, Function: 7}}, targets); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(arch, test, err)
			}
			for i := range code {
				if code[i] != before[i] {
					t.Fatal("partial relocation on invalid input")
				}
			}
		}
	}
}

func TestLinkedBodiesKeepOrdinaryHostEntryChecksAndResumption(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, unmetered := range []bool{false, true} {
		f := leafFixture(0x6a, nil)
		_, lease, ctx := configuredContext(t, f, Config{Unmetered: unmetered, LinkedCallees: map[uint32]*slots.Function{7: f}})
		data, err := native.NewMemory(8, 8)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(data.Close)
		buffer, err := data.Buffer()
		if err != nil {
			t.Fatal(err)
		}
		view := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(buffer))))
		writeSlot(t, ctx, 16, 0, value.Value{Lo: 7})
		writeSlot(t, ctx, 16, 1, value.Value{Lo: 19})
		state := native.State{Used: 16, Depth: 1, Fuel: 100, Budget: 1, Payload: [4]uint64{0, 0, view, 0}}
		if !unmetered {
			state.Payload[3] = 1
		}
		done := false
		for attempt := 0; attempt < 8; attempt++ {
			status, err := ctx.RunState(lease, 0, &state)
			if err != nil {
				t.Fatal(err)
			}
			if status == native.Poll {
				state.Budget = 1
				continue
			}
			if status != native.Return || readSlot(t, ctx, 16, 4).Lo != 26 {
				t.Fatal("ordinary linked host entry or poll resume failed", unmetered, status, state)
			}
			done = true
			break
		}
		if !done {
			t.Fatal("ordinary entry did not finish")
		}
		for _, state := range []native.State{
			{Used: 8, Depth: 1, Budget: 64, Payload: [4]uint64{0, 0, view, 0}},
			{Used: 48, Depth: 1, Budget: 64, Payload: [4]uint64{0, 0, view, 0}},
			{Used: 16, Depth: 1, Continuation: uint64(len(f.Code)), Budget: 64, Payload: [4]uint64{0, 0, view, 0}},
		} {
			status, err := ctx.RunState(lease, 0, &state)
			if err != nil || status != native.Exit(nativeabi.BadState) {
				t.Fatal("ordinary linked entry bypassed header", unmetered, state, status, err)
			}
		}
	}
}
