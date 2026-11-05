package codegen

import (
	"bytes"
	"fmt"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLoopAuxiliaryRegisterConstraints(t *testing.T) {
	for _, tc := range []struct {
		op   uint32
		want []byte
	}{{0x6a, []byte{2, 8, 11}}, {0x1b, []byte{8, 11}}, {0x1c, []byte{8, 11}}, {0x69, []byte{11}}, {0x7b, []byte{11}}, {0x28, []byte{2, 8, 11}}, {0x6d, nil}, {0xff, nil}} {
		f := &slots.Function{Slots: []wasm.ValueType{wasm.I32, wasm.I32}, Code: []slots.Instruction{{Op: tc.op, Dst: 0, A: 0, B: 1, C: slots.NoSlot}}}
		r := memoryRegion{end: 0, proven: map[int]bool{0: true}}
		e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, usesMemory: true}
		if got := e.loopAuxiliaryRegisters(0, r); !slices.Equal(got, tc.want) {
			t.Fatal(tc.op, got, tc.want)
		}
		e.memoryViews = true
		if slices.Contains(e.loopAuxiliaryRegisters(0, r), byte(11)) {
			t.Fatal("indexed view lends its length register")
		}
		e.sharedDefaultMemory = true
		if got := e.loopAuxiliaryRegisters(0, r); !slices.Equal(got, tc.want) {
			t.Fatal("stable shared default", tc.op, got)
		}
		e.a.arch = "arm64"
		if len(e.loopAuxiliaryRegisters(0, r)) != 0 {
			t.Fatal("x86 constraint leaked to arm64")
		}
	}
	f := &slots.Function{Slots: []wasm.ValueType{wasm.F64}, Code: []slots.Instruction{{Op: 0x20, Dst: 0, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}}}
	e := functionEmitter{a: &assembler{arch: "amd64"}, f: f, usesMemory: true}
	if len(e.loopAuxiliaryRegisters(0, memoryRegion{end: 0})) != 0 {
		t.Fatal("unclassified floating lowering borrowed integer scratch")
	}
}

func scratchLoopFixture(t *testing.T, wide bool, kind string) *slots.Function {
	t.Helper()
	typ, xor, pop := wasm.I32, uint32(0x73), uint32(0x69)
	if wide {
		typ, xor, pop = wasm.I64, 0x85, 0x7b
	}
	address, cursor := uint32(8), uint32(9)
	if kind == "unroll" {
		address, cursor = 9, 10
	}
	params := make([]wasm.ValueType, cursor)
	for i := range params {
		params[i] = typ
	}
	params[8] = wasm.I32
	params[address] = wasm.I32
	body := module.Expr{{Op: 3, Block: -64}}
	for i := uint32(0); i < 8; i++ {
		body = append(body, module.Instr{Op: 0x20, A: i}, module.Instr{Op: 0x20, A: (i + 1) % 8})
		if kind == "select" {
			body = append(body, module.Instr{Op: 0x20, A: 8}, module.Instr{Op: 0x1b})
		} else {
			body = append(body, module.Instr{Op: xor})
			if kind == "popcount" {
				body = append(body, module.Instr{Op: pop})
			}
		}
		body = append(body, module.Instr{Op: 0x21, A: i})
	}
	if kind == "unroll" {
		body = append(body, module.Instr{Op: 0x20, A: cursor}, module.Instr{Op: 0x41, Bits: 255}, module.Instr{Op: 0x71})
	} else {
		body = append(body, module.Instr{Op: 0x41})
	}
	body = append(body, module.Expr{
		{Op: 0x20, A: cursor}, {Op: 0x36, A: 2},
		{Op: 0x20, A: cursor}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x22, A: cursor},
	}...)
	if kind == "unroll" {
		body = append(body, module.Instr{Op: 0x20, A: 8}, module.Instr{Op: 0x49})
	} else {
		body = append(body, module.Instr{Op: 0x41, Bits: 4}, module.Instr{Op: 0x47})
	}
	body = append(body, module.Expr{
		{Op: 0xd}, {Op: 0xb},
		// Both checked access and memory.size must see the restored length.
		{Op: 0x20, A: address}, {Op: 0x28, A: 2}, {Op: 0x1a}, {Op: 0x3f}, {Op: 0x20}, {Op: 0xb},
	}...)
	m := functionFixture(module.FuncType{Params: params, Results: []wasm.ValueType{wasm.I32, typ}}, 1, body)
	m.Memories = []module.Limits{{Min: 1}}
	return functionModule(t, m).Functions[0]
}

func TestLoopScratchHomesRestoreEveryFrameMemoryAndContinuation(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, kind := range []string{"xor", "select", "popcount", "unroll"} {
			for _, shared := range []bool{false, true} {
				t.Run(fmt.Sprintf("wide=%v/%s/shared=%v", wide, kind, shared), func(t *testing.T) {
					checkLoopScratchFrames(t, scratchLoopFixture(t, wide, kind), kind, shared)
				})
			}
		}
	}
}

func checkLoopScratchFrames(t *testing.T, f *slots.Function, kind string, shared bool) {
	t.Helper()
	cfg := Config{Unmetered: true, MemoryCount: 1, MemoryMinimumPages: 1, NativeThreads: shared}
	if shared {
		cfg.SharedMemories = []bool{true}
	}
	branchPC := -1
	for pc, in := range f.Code {
		if in.Op == slots.ControlBase+uint32(ir.If) {
			branchPC = pc
		}
	}
	if branchPC < 0 {
		t.Fatal("missing loop branch")
	}
	for _, arch := range []string{"amd64", "arm64"} {
		c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
		if err != nil || kind != "unroll" && c.NativeLoops != 1 || kind == "unroll" && c.UnrolledLoops != 1 {
			t.Fatal(arch, "loop not exercised", err)
		}
		if arch == "amd64" {
			e := functionEmitter{a: &assembler{arch: arch}, f: f, unmetered: true, usesMemory: true, memoryMinimum: 65536, memoryViews: shared, sharedDefaultMemory: shared}
			regions := e.addNativeLoops(e.memoryRegions())
			e.selectRegisters(cfg, regions)
			e.prepareLinearScan(regions, 1<<18)
			for start, r := range regions {
				if r.loop == nil && r.unroll == nil {
					continue
				}
				if e.allocationPlans != nil {
					e.useRegisterPlan(e.allocationPlans[start])
				}
				plan := e.loopRegisterPlan(start, r)
				if plan == nil {
					t.Fatal(kind, "no additional loop homes selected")
				}
				var registers []byte
				for _, reg := range plan.homes {
					registers = append(registers, reg)
				}
				for _, reg := range []byte{2, 8, 11} {
					want := reg == 11 || kind == "xor" || kind == "unroll" || reg == 8 && kind == "select"
					if slices.Contains(registers, reg) != want {
						t.Fatal(kind, "scratch home", reg, registers)
					}
				}
			}
		}
	}
	if !native.Available {
		return
	}
	compiled, lease, ctx := configuredContext(t, f, cfg)
	_, reference, ref := functionContext(t, f, false)
	watchdog := time.AfterFunc(30*time.Second, func() { ctx.Cancel(); ref.Cancel() })
	defer watchdog.Stop()
	mem, refMem := bindMemory(t, ctx, 65536, 65536), bindMemory(t, ref, 65536, 65536)
	var view uint64
	if shared {
		mem, view = bindSharedTestMemory(t, ctx, 65536, 65536)
	}
	pollLease := lease
	if kind != "unroll" {
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
	pollBudget := uint64(0)
	for pc := 0; pc != branchPC; pollBudget++ {
		if pollBudget > uint64(len(f.Code)) {
			t.Fatal("unexpected prefix control flow")
		}
		in := f.Code[pc]
		if in.Op == slots.ControlBase+uint32(ir.Jump) {
			pc = int(f.Edges[in.Extra].PC)
		} else {
			pc++
		}
	}
	initial, gotMemory, wantMemory := make([]byte, 65536), make([]byte, 65536), make([]byte, 65536)
	for pc := range f.Code {
		for _, address := range []uint64{0, 65532, 65533} {
			for _, poll := range []bool{false, true} {
				if poll && (pc != 0 || address != 0 || kind == "unroll") {
					continue
				}
				mem.Write(0, initial)
				refMem.Write(0, initial)
				for s, typ := range f.Slots {
					// Finite cold states keep interrupted recurrence temporaries
					// at zero; arbitrary data parameters still exercise all bits.
					bits := uint64(0)
					if s < 8 {
						bits = uint64(s*73 + 1)
						if typ == wasm.I64 {
							bits |= 0xabcdef0000000000
						}
					}
					if kind == "unroll" && s == 8 {
						bits = 17
					} else if kind == "unroll" && s == 9 || kind != "unroll" && s == 8 {
						bits = address
					} else if s == 9 {
						bits = 0
					}
					writeSlot(t, ctx, 16, slots.Slot(s), value.Value{Lo: bits})
					writeSlot(t, ref, 16, slots.Slot(s), value.Value{Lo: bits})
				}
				state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 10000}
				entry := lease
				if poll {
					state.Budget, entry = pollBudget, pollLease
				}
				state.Payload[2] = view
				gs, ge, err := ctx.Run(entry, 0, state)
				state.Payload[2] = 0
				ws, we, werr := ref.Run(reference, 0, state)
				if err != nil || werr != nil || ge != we || gs.Continuation != ws.Continuation || gs.Payload[0] != ws.Payload[0] || ge == native.Trap && gs.Source != ws.Source {
					t.Fatal(pc, address, poll, ge, we, gs, ws, err, werr)
				}
				for s := range f.Slots {
					if readSlot(t, ctx, 16, slots.Slot(s)) != readSlot(t, ref, 16, slots.Slot(s)) {
						t.Fatal("canonical frame", pc, address, poll, s)
					}
				}
				mem.Read(0, gotMemory)
				refMem.Read(0, wantMemory)
				if !bytes.Equal(gotMemory, wantMemory) {
					t.Fatal("memory before exit", pc, address, poll)
				}
			}
		}
	}
}
