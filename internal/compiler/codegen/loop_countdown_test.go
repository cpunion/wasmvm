package codegen

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestLoopCountdownFullFramesColdPCsAndGuards(t *testing.T) {
	for _, address := range []uint32{0, 1, 255} {
		f := loopMemoryValueFixture(t, false, address)
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			if err != nil || c.DeferredLoopCounters != 1 {
				t.Fatal(arch, "counter not selected", c, err)
			}
			forcePrivateLoopPoll(t, c, arch)
			limits := DefaultLimits()
			limits.MaxCodeBytes = uint64(len(c.Entry.Code))
			if _, err = CompileFunctionWithConfig(arch, f, limits, Config{Unmetered: true, MemoryMinimumPages: 1}); err != nil {
				t.Fatal(err)
			}
			limits.MaxCodeBytes--
			if c, err = CompileFunctionWithConfig(arch, f, limits, Config{Unmetered: true, MemoryMinimumPages: 1}); c != nil || err == nil {
				t.Fatal("code limit ignored", arch, c, err)
			}
		}
		if !native.Available {
			continue
		}
		compiled, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
		_, reference, ref := functionContext(t, f, false)
		code, err := native.Publish(forcePrivateLoopPoll(t, compiled, runtime.GOARCH), []uint32{0}, DefaultLimits().MaxCodeBytes)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(code.Close)
		polled, err := code.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(polled.Close)
		mem, refMem := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
		initial := bytes.Repeat([]byte{0xa5}, 65536)
		cases := [][2]uint32{{0, 3}, {0, 0}, {1, 1}, {2, 1}, {0, 1}, {2, 3}, {0xfffffffe, 0xffffffff}, {0xffffffff, 2}}
		for _, inputs := range cases {
			for start := range f.Code {
				if inputs != cases[0] && start != 0 {
					continue
				}
				for _, poll := range []bool{false, true} {
					if poll && (start != 0 || inputs[0] >= inputs[1]) {
						continue
					}
					t.Run(fmt.Sprintf("address%d/i%d/n%d/pc%d/poll%v", address, inputs[0], inputs[1], start, poll), func(t *testing.T) {
						if err := mem.Write(0, initial); err != nil {
							t.Fatal(err)
						}
						if err := refMem.Write(0, initial); err != nil {
							t.Fatal(err)
						}
						for s := range f.Slots {
							bits := value.Value{Lo: 0xdeadbeef00000100 + uint64(s)}
							switch s {
							case 0:
								bits.Lo = 0xffffffff
							case 1:
								bits.Lo = 0x1234567800000000 | uint64(inputs[1])
							case 2:
								bits.Lo = 0x9abcdef000000000 | uint64(inputs[0])
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
							entry = polled
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
							t.Fatal(got, want, gs, ws, err, werr, gerr, serr)
						}
						for s := range f.Slots {
							if a, b := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, ref, 16, slots.Slot(s)); a != b {
								t.Fatal("frame", s, a, b)
							}
						}
						actual, expected := make([]byte, 65536), make([]byte, 65536)
						if err := mem.Read(0, actual); err != nil {
							t.Fatal(err)
						}
						if err := refMem.Read(0, expected); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(actual, expected) {
							t.Fatal("memory effects")
						}
					})
				}
			}
		}
	}
}

func TestLoopCountdownRejectsObservableCountersAndMappingChanges(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		for _, kind := range []string{"accept", "precise", "header", "reverse", "always", "span", "copies", "no-predicate", "literal", "signed", "same-input", "result-alias", "not-one", "not-add", "wrong-local", "counter-read", "counter-copy", "bound-write", "missing-counter", "missing-bound", "shared-home", "other-owner", "future-home", "borrowed-length"} {
			if kind == "borrowed-length" && arch != "amd64" {
				continue
			}
			t.Run(arch+"/"+kind, func(t *testing.T) {
				f := loopMemoryValueFixture(t, false, 0)
				e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: true, usesMemory: true}
				p := &registerPlan{homes: map[slots.Slot]byte{2: 13, 1: 14}}
				e.useRegisterPlan(p)
				r := memoryRegion{end: 16, loop: &nativeLoop{yes: true, flags: true}, predicate: &loopPredicate{in: f.Code[16]}}
				start := 5
				switch kind {
				case "precise":
					e.unmetered = false
				case "header":
					start = 0
				case "reverse":
					r.loop.yes = false
				case "always":
					r.loop.always = true
				case "span":
					r.loop.memory = &memoryLoop{}
				case "copies":
					r.copies = &loopFrameCopies{}
				case "no-predicate":
					r.predicate = nil
				case "literal":
					r.predicate.literal = &f.Code[15]
				case "signed":
					r.predicate.in.Op = 0x48
				case "same-input":
					r.predicate.in.B = r.predicate.in.A
				case "result-alias":
					r.predicate.in.Dst = r.predicate.in.A
				case "not-one":
					f.Sources[f.Code[12].Source].Wasm.Bits = 2
				case "not-add":
					f.Code[13].Op = 0x6b
				case "wrong-local":
					f.Sources[f.Code[14].Source].Wasm.A = 3
				case "counter-read":
					f.Code[7].A = 2
				case "counter-copy":
					f.Code[11].Dst = 7
				case "bound-write":
					f.Code[5].Dst = 1
				case "missing-counter":
					delete(p.homes, 2)
				case "missing-bound":
					delete(p.homes, 1)
				case "shared-home":
					p.homes[1] = 13
				case "other-owner":
					p.homes[3] = 13
				case "future-home":
					q := cloneRegisterPlan(p)
					q.homes[2] = 15
					r.values = &loopValuePlans{plans: []*registerPlan{q}}
				case "borrowed-length":
					p.homes[3] = 11
				}
				if got := e.planLoopCountdown(start, r); (got != nil) != (kind == "accept") {
					t.Fatal("counter selection", got)
				}
			})
		}
	}
}
