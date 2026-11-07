package codegen

import (
	"encoding/binary"
	"math"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLinearScanSplitsAndReentersSpilledIntervals(t *testing.T) {
	b := scanBank{pool: []byte{19}}
	b.advance(0, []registerInterval{{0, 0, 4, 0}})
	b.advance(1, []registerInterval{{1, 1, 1, 0}, {2, 1, 2, 0}})
	if b.resident[0].interval.slot != 1 || len(b.pending) != 2 {
		t.Fatal("short interval did not displace the long interval", b)
	}
	b.advance(2, nil)
	if b.resident[0].interval.slot != 2 {
		t.Fatal(b)
	}
	b.advance(3, nil)
	if b.resident[0].interval.slot != 0 {
		t.Fatal("spilled value never reentered", b)
	}
	b.advance(5, nil)
	if b.resident[0].valid || len(b.pending) != 0 {
		t.Fatal(b)
	}
	b.advance(6, []registerInterval{{3, 6, 6, 0}, {4, 6, 6, 0}})
	if b.resident[0].interval.slot != 3 {
		t.Fatal("nondeterministic equal-end tie", b)
	}
	b.advance(8, nil)
	if b.resident[0].valid || len(b.pending) != 0 {
		t.Fatal(b)
	}
	var unavailable scanBank
	unavailable.advance(0, []registerInterval{{0, 0, 1, 0}})
	if len(unavailable.resident) != 0 || len(unavailable.pending) != 1 {
		t.Fatal(unavailable)
	}
}

func TestLinearScanKeepsCostlyLoopValuesAcrossShorterColdIntervals(t *testing.T) {
	b := scanBank{pool: []byte{19}}
	b.advance(0, []registerInterval{{0, 0, 10, 1000}})
	b.advance(1, []registerInterval{{1, 1, 1, 10}})
	if b.resident[0].interval.slot != 0 {
		t.Fatal("loop accumulator evicted for a cheap temporary")
	}
	b.advance(2, []registerInterval{{2, 2, 3, 2000}})
	if b.resident[0].interval.slot != 2 {
		t.Fatal("higher-cost value did not gain residency")
	}
	b.advance(4, nil)
	if b.resident[0].interval.slot != 0 {
		t.Fatal("spilled accumulator did not regain residency")
	}
	b.advance(11, nil)
	if b.resident[0].valid || len(b.pending) != 0 {
		t.Fatal("expired values retained residency")
	}
}

func TestLinearScanBudgetsRetainConservativeAllocation(t *testing.T) {
	for _, kind := range []string{"map-cells", "arguments", "edges", "moves"} {
		f := globalRegisterFixture("return")
		cells := 1 << 18
		switch kind {
		case "map-cells":
			cells = 1
		case "arguments":
			f = &slots.Function{Slots: []wasm.ValueType{wasm.I64}, Sources: []ir.Source{{}}}
			for range 100 {
				f.Args = append(f.Args, 0)
			}
			for range 100 {
				f.Code = append(f.Code, slots.Instruction{Op: 0x10, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Count: 100})
			}
		case "edges", "moves":
			f = &slots.Function{Slots: []wasm.ValueType{wasm.I32}, Sources: []ir.Source{{}}}
			for range 20 {
				edge := slots.Edge{PC: 0}
				if kind == "moves" {
					edge.Count = 20
					f.Moves = append(f.Moves, slots.Move{Dst: 0, Src: 0})
				}
				f.Edges = append(f.Edges, edge)
			}
			for range 100 {
				f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Table), A: 0, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Count: 20})
			}
		}
		a := assembler{arch: "arm64", labels: map[int]int{}}
		e := functionEmitter{a: &a, f: f, unmetered: true}
		e.prepareRegisters()
		baseline := e.currentRegisterPlan()
		e.prepareLinearScan(nil, cells)
		if e.allocationPlans != nil || e.currentRegisterPlan() != baseline {
			t.Fatal(kind, "budget changed the conservative mapping")
		}
	}
	// A planner with no previously initialized homes still partitions both
	// banks; reserved immutable words must never appear in the integer bank.
	f := globalRegisterFixture("return")
	a := assembler{arch: "arm64", labels: map[int]int{}}
	e := functionEmitter{a: &a, f: f, unmetered: true, constantRegisters: map[uint64]byte{1: 1, 2: 2}}
	e.prepareLinearScan(nil, 1<<18)
	if len(e.allocationPlans) != len(f.Code) {
		t.Fatal("missing standalone plans")
	}
	for _, p := range e.allocationPlans {
		used := map[byte]bool{}
		for _, r := range p.homes {
			if r == 1 || r == 2 || used[r] {
				t.Fatal("overlapping or reserved integer register", r)
			}
			used[r] = true
		}
		used = map[byte]bool{}
		for _, bank := range []map[slots.Slot]byte{p.floats, p.vectors} {
			for _, r := range bank {
				if used[r] {
					t.Fatal("FP/vector bank overlap", r)
				}
				used[r] = true
			}
		}
	}
}

// Disjoint phases reuse all target register classes. Long-lived operands and
// early overlapping phases force spilling; forward control skips entire phases.
// The oracle below models frame words directly and never invokes a compiler.
func globalRegisterFixture(mode string) *slots.Function {
	f := &slots.Function{LocalCount: 58}
	for s := 0; s < 58; s++ {
		t := wasm.I64
		switch {
		case s >= 24 && s < 30 || s == 55:
			t = wasm.F32
		case s >= 30 && s < 36 || s == 53:
			t = wasm.F64
		case s >= 36 && s < 48 || s == 54:
			t = wasm.V128
		case s >= 48 && s < 52:
			t = wasm.ExternRef
		case s == 56:
			t = wasm.I32
		}
		f.Slots = append(f.Slots, t)
	}
	emit := func(op uint32, dst, a, b slots.Slot, local uint32) {
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Offset: uint64(1000 + source)}})
		f.Code = append(f.Code, slots.Instruction{Op: op, Dst: dst, A: a, B: b, C: slots.NoSlot, Source: source})
	}
	starts, ends := make([]int, 7), make([]int, 6)
	for block := 0; block < 6; block++ {
		starts[block] = len(f.Code)
		for s := slots.Slot(0); s < 48; s++ {
			if block != 0 && int(s)%3 != (block-1)%3 {
				continue
			}
			op, operand := uint32(0x7c), slots.Slot(52)
			switch f.Slots[s] {
			case wasm.F32:
				op, operand = 0x92, 55
			case wasm.F64:
				op, operand = 0xa0, 53
			case wasm.V128:
				op, operand = 0xfd0051, 54
			}
			emit(op, s, s, operand, 0)
		}
		emit(0x20, slots.Slot(48+(block+1)%4), slots.NoSlot, slots.NoSlot, uint32(48+block%4))
		ends[block] = len(f.Code)
		emit(slots.ControlBase+uint32(ir.Jump), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	}
	starts[6] = len(f.Code)
	switch mode {
	case "divide":
		emit(0x7f, 23, 52, 57, 0)
	case "call":
		emit(0x10, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	case "helper":
		emit(0xe0, slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	}
	emit(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0)
	for block, pc := range ends {
		in := &f.Code[pc]
		in.Extra = uint32(len(f.Edges))
		if block == 1 || block == 3 {
			in.Op, in.Count, in.A = slots.ControlBase+uint32(ir.If), 2, 56
			if block == 3 {
				in.Op, in.A = slots.ControlBase+uint32(ir.IfNull), 48
			}
			f.Edges = append(f.Edges, slots.Edge{PC: uint32(starts[block+1])}, slots.Edge{PC: uint32(starts[block+2])})
		} else if block == 2 {
			in.Op, in.Count, in.A = slots.ControlBase+uint32(ir.Table), 3, 56
			for target := 3; target < 6; target++ {
				f.Edges = append(f.Edges, slots.Edge{PC: uint32(starts[target])})
			}
		} else {
			in.Count = 1
			f.Edges = append(f.Edges, slots.Edge{PC: uint32(starts[block+1]), Start: uint32(len(f.Moves)), Count: 1})
			f.Moves = append(f.Moves, slots.Move{Dst: slots.Slot(block), Src: slots.Slot(block + 1)})
		}
	}
	return f
}

func modelGlobalRegisters(f *slots.Function, initial []value.Value, start int, cancel bool) ([]value.Value, native.Exit, uint64, uint64, uint64) {
	words := append([]value.Value(nil), initial...)
	for pc := start; pc < len(f.Code); {
		in := f.Code[pc]
		if cancel && (pc%256 == 0 || in.Op == 0x10 || in.Op >= slots.ControlBase && !(in.Op == slots.ControlBase+uint32(ir.Jump) && f.Edges[in.Extra].PC == uint32(pc+1) && f.Edges[in.Extra].Count == 0)) {
			return words, native.Poll, uint64(pc), 0, 0
		}
		if in.Op >= slots.ControlBase {
			kind := ir.TermKind(in.Op - slots.ControlBase)
			if kind == ir.Return {
				return words, native.Return, uint64(pc), uint64(pc), 0
			}
			edge := in.Extra
			switch kind {
			case ir.If:
				if uint32(words[in.A].Lo) == 0 {
					edge++
				}
			case ir.IfNull:
				if words[in.A].Lo != 0 {
					edge++
				}
			case ir.Table:
				edge += min(uint32(words[in.A].Lo), in.Count-1)
			}
			e := f.Edges[edge]
			for _, m := range f.Moves[e.Start : e.Start+e.Count] {
				words[m.Dst] = words[m.Src]
			}
			pc = int(e.PC)
			continue
		}
		source := uint64(f.Sources[in.Source].Wasm.Offset)
		switch in.Op {
		case 0x20:
			words[in.Dst] = words[f.Sources[in.Source].Wasm.A]
		case 0x7c:
			words[in.Dst].Lo = words[in.A].Lo + words[in.B].Lo
		case 0x92:
			words[in.Dst].Lo = uint64(math.Float32bits(math.Float32frombits(uint32(words[in.A].Lo)) + math.Float32frombits(uint32(words[in.B].Lo))))
		case 0xa0:
			words[in.Dst].Lo = math.Float64bits(math.Float64frombits(words[in.A].Lo) + math.Float64frombits(words[in.B].Lo))
		case 0xfd0051:
			words[in.Dst] = value.Value{Lo: words[in.A].Lo ^ words[in.B].Lo, Hi: words[in.A].Hi ^ words[in.B].Hi}
		case 0x7f:
			a, b := int64(words[in.A].Lo), int64(words[in.B].Lo)
			if b == 0 {
				return words, native.Trap, 0, uint64(wasm.TrapIntegerDivideByZero), source
			}
			if a == math.MinInt64 && b == -1 {
				return words, native.Trap, 0, uint64(wasm.TrapIntegerOverflow), source
			}
			words[in.Dst].Lo = uint64(a / b)
		case 0x10:
			return words, native.Call, uint64(pc + 1), uint64(pc), source
		case 0xe0:
			return words, native.RuntimeHelper, uint64(pc + 1), uint64(pc), source
		default:
			panic("unmodeled fixture instruction")
		}
		pc++
	}
	panic("fixture fell through")
}

func TestGlobalRegisterPlansColdEntriesControlSnapshotsAndTraps(t *testing.T) {
	for _, mode := range []string{"return", "divide", "call", "helper"} {
		f := globalRegisterFixture(mode)
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
			if err != nil {
				t.Fatal(arch, mode, err)
			}
			if c.AllocationBlocks < 6 {
				t.Fatal("global allocation was not exercised", arch, mode, c.AllocationBlocks)
			}
			limits := DefaultLimits()
			limits.MaxCodeBytes = uint64(len(c.Entry.Code)) - 1
			if c, err := CompileUnmeteredFunction(arch, f, limits); c != nil || err == nil {
				t.Fatal("code budget ignored", arch, mode)
			}
		}
		if !native.Available {
			continue
		}
		compiled, lease, plainContext := functionContext(t, f, true)
		// Publish cancellation after the C adapter's early check, exercising
		// generated polls and their actual allocation mapping at each cold PC.
		var prefix []byte
		if runtime.GOARCH == "arm64" {
			for _, word := range []uint32{0xd2800029, 0x9100000a | uint32(nativeabi.ControlOffset)<<10, 0xc89ffd49} {
				prefix = binary.LittleEndian.AppendUint32(prefix, word)
			}
		} else {
			prefix = []byte{0x48, 0xc7, 0x87}
			prefix = binary.LittleEndian.AppendUint32(prefix, uint32(nativeabi.ControlOffset))
			prefix = binary.LittleEndian.AppendUint32(prefix, 1)
		}
		code, err := native.Publish(append(prefix, compiled.Entry.Code...), []uint32{0}, DefaultLimits().MaxCodeBytes)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(code.Close)
		interrupted, err := code.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(interrupted.Close)
		for _, input := range []uint64{0, 1, 2, math.MaxUint64} {
			initial := make([]value.Value, len(f.Slots))
			for s, typ := range f.Slots {
				initial[s].Lo = uint64(s)*0x123456789abcdef + input
				switch typ {
				case wasm.F32:
					initial[s].Lo = uint64(math.Float32bits(float32(s) / 4))
				case wasm.F64:
					initial[s].Lo = math.Float64bits(float64(s) / 4)
				case wasm.V128:
					initial[s].Hi = 0xfedcba9876543210 ^ uint64(s)
				case wasm.ExternRef:
					if input == 0 {
						initial[s].Lo = 0
					}
				}
			}
			initial[52].Lo, initial[56].Lo, initial[57].Lo = 1<<63, input, input
			for start := range f.Code {
				for _, cancel := range []bool{false, true} {
					func() {
						ctx := plainContext
						entry := lease
						if cancel {
							size := uint64(len(f.Slots))*16 + 32
							ctx, err = native.NewContext(size, size, 1024)
							if err != nil {
								t.Fatal(err)
							}
							defer ctx.Close()
							entry = interrupted
						}
						for s, v := range initial {
							writeSlot(t, ctx, 16, slots.Slot(s), v)
						}
						if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start)}); err != nil {
							t.Fatal(err)
						}
						got, err := ctx.Enter(entry, 0)
						state, stateErr := ctx.State()
						want, exit, pc, payload, source := modelGlobalRegisters(f, initial, start, cancel)
						if err != nil || stateErr != nil || got != exit || got != native.Trap && state.Continuation != pc || state.Payload[0] != payload || source != 0 && state.Source != source {
							t.Fatal(runtime.GOARCH, mode, input, start, cancel, got, exit, state, pc, payload, source, err, stateErr)
						}
						for s, expected := range want {
							if got := readSlot(t, ctx, 16, slots.Slot(s)); got != expected {
								t.Fatal(mode, input, start, cancel, "frame", s, got, expected)
							}
						}
					}()
				}
			}
		}
	}
}
