package codegen

import (
	"encoding/binary"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestScalarResultCarrierPreservesCompleteWordsAndLargeSources(t *testing.T) {
	want := value.Value{Lo: 0xfedcba9876543210, Hi: 0x123456789abcdef0}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, source := range []slots.Slot{0, 4097} {
			for _, cached := range []bool{false, true} {
				a := assembler{arch: arch, longBranches: true}
				e := functionEmitter{a: &a, exit: a.label()}
				cancelled := a.label()
				e.callLoad(3, e.callContext(), uint32(nativeabi.PayloadOffset))
				e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
				e.callLoad(0, e.callContext(), uint32(nativeabi.UsedOffset))
				e.callAdd(0, 0, 4)
				e.registerMove(e.callFrame(), e.callReg(0), true)
				e.callConst(5, want.Lo)
				e.callScalarResult(source, cached, cancelled)
				e.callStore(1, e.callContext(), uint32(nativeabi.ResultLoOffset))
				e.callStore(2, e.callContext(), uint32(nativeabi.ResultHiOffset))
				e.rawStatus(nativeabi.Return)
				a.mark(cancelled)
				e.rawStatus(nativeabi.Poll)
				a.mark(e.exit)
				body, err := a.finish()
				if err != nil {
					t.Fatal(err)
				}
				entry, err := nativeabi.Wrap(arch, body, 16384)
				if err != nil {
					t.Fatal(err)
				}
				if arch != runtime.GOARCH || !native.Available {
					continue
				}
				metadata, err := native.NewMemory(96, 96)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(metadata.Close)
				data, err := metadata.Buffer()
				if err != nil {
					t.Fatal(err)
				}
				address := uint64(uintptr(unsafe.Pointer(unsafe.SliceData(data))))
				binary.LittleEndian.PutUint64(data[nativeabi.CallDestinationsOffset:], address+64)
				code, err := native.Publish(entry.Code, []uint32{0}, uint64(len(entry.Code)))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(code.Close)
				lease, err := code.Acquire()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(lease.Close)
				capacity := uint64(source+2) * 16
				ctx, err := native.NewContext(capacity, capacity, 2)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(ctx.Close)
				writeSlot(t, ctx, 16, source, want)
				state := native.State{Used: 16, Depth: 2, Payload: [4]uint64{address}}
				if status, err := ctx.RunState(lease, 0, &state); err != nil || status != native.Return || state.ResultLo != want.Lo || state.ResultHi != want.Hi || readSlot(t, ctx, 0, 0) != want {
					t.Fatal("incomplete result or carrier", arch, source, cached, status, state, err)
				}
			}
		}
	}
}

func TestPrivateScalarResultsSurviveEntryScratchAndBorrowedHomes(t *testing.T) {
	const original, incoming = uint64(77), uint64(0x87654321fedcba98)
	for _, arch := range []string{"amd64", "arm64"} {
		for _, memory := range []bool{false, true} {
			probe := functionEmitter{a: &assembler{arch: arch}, frameRegisters: true, usesMemory: memory}
			pool := probe.integerRegisterPool()
			regs := slices.Clone(pool)
			if arch == "amd64" {
				// Loop plans can borrow constrained scratch. An ordinary
				// private stub must reload those homes after initialization.
				regs = append(regs, 2, 8, 11)
			}
			for _, reg := range regs {
				a := assembler{arch: arch, longBranches: true}
				e := functionEmitter{a: &a, f: &slots.Function{Slots: []wasm.ValueType{wasm.I64}}, unmetered: true, frameRegisters: true, usesMemory: memory, memoryMinimum: 65536, exit: a.label(), bad: a.label()}
				resume, done := a.label(), a.label()
				plan := &registerPlan{homes: map[slots.Slot]byte{0: reg}, slots: []slots.Slot{0}}
				e.callLoad(callUsed, e.callContext(), uint32(nativeabi.UsedOffset))
				e.callLoad(callDepth, e.callContext(), uint32(nativeabi.DepthOffset))
				e.callLoad(4, e.callContext(), uint32(nativeabi.ArenaOffset))
				e.callConst(1, incoming)
				a.branch(resume, false, 0)
				e.callResumes = []nativeCallResume{{label: resume, plan: plan, result: 0}}
				e.emitCallResumes([]int{done})
				a.mark(done)
				e.useRegisterPlan(plan)
				e.spillRegisters()
				e.rawStatus(nativeabi.Return)
				a.mark(e.bad)
				e.rawStatus(nativeabi.BadState)
				a.mark(e.exit)
				body, err := a.finish()
				if err != nil {
					t.Fatal(arch, memory, reg, err)
				}
				entry, err := nativeabi.Wrap(arch, body, 16384)
				if err != nil {
					t.Fatal(err)
				}
				if arch != runtime.GOARCH || !native.Available {
					continue
				}
				mem, err := native.NewMemory(65536, 65536)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(mem.Close)
				code, err := native.Publish(entry.Code, []uint32{0}, uint64(len(entry.Code)))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(code.Close)
				lease, err := code.Acquire()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(lease.Close)
				ctx, err := native.NewContext(16, 16, 1)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(ctx.Close)
				if err := ctx.BindMemory(mem); err != nil {
					t.Fatal(err)
				}
				writeSlot(t, ctx, 0, 0, value.Value{Lo: original, Hi: 123})
				if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
					t.Fatal(arch, memory, reg, exit, err)
				}
				want := original
				if slices.Contains(pool, reg) {
					want = incoming
				}
				if got := readSlot(t, ctx, 0, 0); got != (value.Value{Lo: want, Hi: 123}) {
					t.Fatal("private result or high word corrupted", arch, memory, reg, got, want)
				}
			}
		}
	}
}
