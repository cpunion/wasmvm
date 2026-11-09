package codegen

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func loopMemoryValueFixture(t *testing.T, wide bool, address uint32) *slots.Function {
	t.Helper()
	data, err := os.ReadFile("../../../benchmarks/modern/threads-private-memory.wasm")
	if err != nil {
		t.Fatal(err)
	}
	m, err := decode.Decode(data, decode.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if wide {
		m.Types[0].Params[0], m.Types[0].Results[0] = wasm.I64, wasm.I64
	}
	body := m.Functions[0].Body
	firstAdd := true
	for pc := range body {
		in := &body[pc]
		if in.Op == 0x41 && in.Bits == 0 {
			in.Bits = uint64(address)
		}
		if !wide {
			continue
		}
		switch in.Op {
		case 0x28:
			in.Op = 0x29
		case 0x36:
			in.Op = 0x37
		case 0x6a:
			if firstAdd {
				in.Op = 0x7c
				firstAdd = false
			}
		case 0x41:
			if pc > 0 && body[pc-1].Op == 0x29 {
				in.Op = 0x42
			}
		}
	}
	return functionModule(t, m).Functions[0]
}

func TestLoopMemoryValueBothTargetsColdEntriesAndPollFrames(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, address := range []uint32{0, 1, 127, 128, 255, 256, 16380, 32760, 65529} {
			t.Run(fmt.Sprintf("wide=%v/address=%d", wide, address), func(t *testing.T) {
				f := loopMemoryValueFixture(t, wide, address)
				for _, arch := range []string{"amd64", "arm64"} {
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
					if err != nil {
						t.Fatal(err)
					}
					supported := address <= 65528 || !wide
					if arch == "arm64" {
						width := uint32(4)
						if wide {
							width = 8
						}
						supported = supported && (address <= 255 || address%width == 0 && address/width < 4096)
					}
					if (c.ForwardedMemory == 1) != supported {
						t.Fatalf("%s forwarded%d want%v", arch, c.ForwardedMemory, supported)
					}
					if c.NativeLoops != 0 {
						forcePrivateLoopPoll(t, c, arch)
					}
				}
				if !native.Available {
					return
				}
				compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
				_, reference, ref := functionContext(t, f, false)
				var pollLease *native.Lease
				if compiled.NativeLoops != 0 {
					patched, err := native.Publish(forcePrivateLoopPoll(t, compiled, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(patched.Close)
					pollLease, err = patched.Acquire()
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(pollLease.Close)
				}
				memory, refMemory := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
				initial := bytes.Repeat([]byte{0xa5}, 65536)
				for start := range f.Code {
					for _, poll := range []bool{false, true} {
						if poll && (start != 0 || pollLease == nil) {
							continue
						}
						if err := memory.Write(0, initial); err != nil {
							t.Fatal(err)
						}
						if err := refMemory.Write(0, initial); err != nil {
							t.Fatal(err)
						}
						for s := range f.Slots {
							bits := value.Value{Lo: 0xdeadbeef00000100 + uint64(s)}
							switch s {
							case 0:
								bits.Lo = 0xffffffffffffffff
							case 1:
								bits.Lo = 3
							case 2:
								bits.Lo = 0
							case 3:
								bits.Hi = 0x123456789abcdef0
							}
							writeSlot(t, ctx, 16, slots.Slot(s), bits)
							writeSlot(t, ref, 16, slots.Slot(s), bits)
						}
						state := native.State{Used: 16, Depth: 1, Continuation: uint64(start), Budget: 10000}
						entry := lease
						if poll {
							state.Budget = 16
							entry = pollLease
						}
						if err := ctx.SetState(state); err != nil {
							t.Fatal(err)
						}
						if err := ref.SetState(state); err != nil {
							t.Fatal(err)
						}
						got, err := ctx.Enter(entry, 0)
						want, werr := ref.Enter(reference, 0)
						gs, gerr := ctx.State()
						ws, serr := ref.State()
						if err != nil || werr != nil || gerr != nil || serr != nil || got != want || gs.Continuation != ws.Continuation || gs.Payload[0] != ws.Payload[0] || got == native.Trap && gs.Source != ws.Source {
							t.Fatal(start, poll, got, want, gs, ws, err, werr, gerr, serr)
						}
						for s := range f.Slots {
							if a, b := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); a != b {
								t.Fatal(start, poll, "frame", s, a, b)
							}
						}
						actual, expected := make([]byte, 65536), make([]byte, 65536)
						if err := memory.Read(0, actual); err != nil {
							t.Fatal(err)
						}
						if err := refMemory.Read(0, expected); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(actual, expected) {
							t.Fatal(start, poll, "memory effects")
						}
					}
				}
			})
		}
	}
}

func TestLoopMemoryValueRejectsUnprovedMemoryAndRegisterOwnership(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, kind := range []string{"accept", "shared", "views", "precise", "header", "affine", "copies", "remap", "unchecked", "memory64", "other-memory", "high-offset", "large-offset", "different-address", "narrow-store", "store-first", "extra-read", "no-store", "unknown-address", "no-free-register", "future-register-plan"} {
			t.Run(arch+"/"+kind, func(t *testing.T) {
				f := loopMemoryValueFixture(t, false, 0)
				e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: true, usesMemory: true}
				r := memoryRegion{end: 16, loop: &nativeLoop{}, proven: map[int]bool{7: true, 10: true}}
				start := 5
				switch kind {
				case "shared":
					e.sharedDefaultMemory = true
				case "views":
					e.memoryViews = true
				case "precise":
					e.unmetered = false
				case "header":
					start = 0
				case "affine":
					r.loop.memory = &memoryLoop{}
				case "copies":
					r.copies = &loopFrameCopies{}
				case "remap":
					r.addresses = map[int]memoryAddress{7: {}}
				case "unchecked":
					delete(r.proven, 7)
				case "memory64":
					f.Slots[f.Code[7].A] = wasm.I64
				case "other-memory":
					f.Sources[f.Code[10].Source].Wasm.Memory = 1
				case "high-offset":
					f.Sources[f.Code[7].Source].Wasm.MemoryOffsetHigh = 1
				case "large-offset":
					f.Sources[f.Code[7].Source].Wasm.B = 0x80000000
				case "different-address":
					f.Sources[f.Code[10].Source].Wasm.B = 1
				case "narrow-store":
					f.Code[10].Op = 0x3a
				case "store-first":
					f.Code[7], f.Code[10] = f.Code[10], f.Code[7]
				case "extra-read":
					f.Code[8] = f.Code[7]
					r.proven[8] = true
				case "no-store":
					f.Code[10].Op = 1
				case "unknown-address":
					f.Code[7].A = 0
				case "no-free-register", "future-register-plan":
					p := &registerPlan{homes: map[slots.Slot]byte{}}
					for reg := byte(0); reg < 32; reg++ {
						p.homes[slots.Slot(reg)] = reg
					}
					if kind == "no-free-register" {
						e.useRegisterPlan(p)
					} else {
						r.values = &loopValuePlans{plans: []*registerPlan{p}}
					}
				}
				if got := e.planLoopMemoryValue(start, r); (got != nil) != (kind == "accept") {
					t.Fatal("cache selection", got)
				}
			})
		}
	}
}
