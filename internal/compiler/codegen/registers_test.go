package codegen

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/nativeabi"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestGeneratedPollContinuationsAndSpills(t *testing.T) {
	if !native.Available {
		return
	}
	f := &slots.Function{Slots: []wasm.ValueType{wasm.I64}, Args: []slots.Slot{0},
		Sources: []ir.Source{{Wasm: module.Instr{Op: 1}}, {Wasm: module.Instr{Op: 0x42, Bits: 42}}}}
	for i := 0; i < 300; i++ {
		f.Code = append(f.Code, slots.Instruction{Op: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	}
	f.Code[1].Op, f.Code[1].Source, f.Code[1].Dst = 0x42, 1, 0
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Count: 1,
		Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
	for _, fast := range []bool{false, true} {
		compile := CompileFunction
		if fast {
			compile = CompileUnmeteredFunction
		}
		compiled, err := compile(runtime.GOARCH, f, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		// Set cancellation after the adapter's entry check, exercising generated
		// polling rather than the adapter's early return. These are release stores.
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
		lease, err := code.Acquire()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(lease.Close)
		for _, start := range []uint64{0, 1, 300} {
			c, err := native.NewContext(16, 16, 1)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(c.Close)
			writeSlot(t, c, 0, 0, value.Value{Lo: 7})
			if err := c.SetState(native.State{Depth: 1, Continuation: start, Budget: 600}); err != nil {
				t.Fatal(err)
			}
			exit, err := c.Enter(lease, 0)
			state, stateErr := c.State()
			wantPC, wantValue := uint64(300), uint64(7)
			if start < 2 {
				wantValue = 42
			}
			if fast && start == 0 {
				wantPC, wantValue = 0, 7
			} else if fast && start == 1 {
				wantPC = 256
			}
			if err != nil || stateErr != nil || exit != native.Poll || state.Continuation != wantPC || readSlot(t, c, 0, 0) != (value.Value{Lo: wantValue}) {
				t.Fatalf("fast=%v start=%d: %v %v %v %+v", fast, start, err, stateErr, exit, state)
			}
		}
	}
}

func TestAllocatedIntegerSemantics(t *testing.T) {
	if !native.Available {
		return
	}
	// Include destination/source aliasing in independently constructed slot IR,
	// not just the non-overlapping operands produced by today's slot lowering.
	for _, dst := range []slots.Slot{0, 1, 2} {
		for _, op := range integerOps() {
			f := &slots.Function{
				Slots:   []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64},
				Sources: []ir.Source{{Wasm: module.Instr{Op: op, Offset: 99}}, {Wasm: module.Instr{Op: 0xb}}},
				Code: []slots.Instruction{
					{Op: op, A: 0, B: 1, C: slots.NoSlot, Dst: dst},
					{Op: slots.ControlBase + uint32(ir.Return), Source: 1, Count: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot},
				}, Args: []slots.Slot{dst},
			}
			compiled, lease, c := functionContext(t, f, true)
			edges := []uint64{0, 1, 31, 32, 63, 64, 0x80000000, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0)}
			for _, a := range edges {
				for _, b := range edges {
					want, trap := numeric.Eval(op, a, b)
					got, exit, state := runFunction(t, compiled, lease, c, []value.Value{{Lo: a}, {Lo: b}}, 77, false, 0)
					if trap != wasm.NoTrap {
						if exit != native.Trap || state.Payload[0] != uint64(trap) || state.Source != 99 {
							t.Fatalf("op %x: %v %+v", op, exit, state)
						}
					} else if exit != native.Return || len(got) != 1 || got[0] != (value.Value{Lo: want}) {
						t.Fatalf("op %x dst %d (%x,%x): %v %v want %x", op, dst, a, b, exit, got, want)
					}
					if state.Fuel != 77 {
						t.Fatal("unmetered entry changed fuel")
					}
				}
			}
			c.Close()
			lease.Close() // Cleanup is idempotent; do not retain 198 contexts.
		}
	}
}

func TestImmediateShiftsAndIndependentContinuation(t *testing.T) {
	if !native.Available {
		return
	}
	for _, first := range []uint32{0x74, 0x86} {
		for op := first; op < first+5; op++ {
			for _, count := range []uint64{0, 1, 31, 32, 33, 63, 64, 65, ^uint64(0)} {
				for _, dst := range []slots.Slot{0, 1, 2} {
					f := &slots.Function{Slots: []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64}, Args: []slots.Slot{dst},
						Sources: []ir.Source{{Wasm: module.Instr{Op: 0x42, Bits: count}}, {Wasm: module.Instr{Op: op}}},
						Code: []slots.Instruction{
							{Op: 0x42, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
							{Op: op, Source: 1, Dst: dst, A: 0, B: 1, C: slots.NoSlot},
							{Op: slots.ControlBase + uint32(ir.Return), Count: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
						}}
					compiled, lease, c := functionContext(t, f, true)
					for _, input := range []uint64{0, ^uint64(0), 1 << 63, 0x123456789abcdef0} {
						want, _ := numeric.Eval(op, input, count)
						got, exit, _ := runFunction(t, compiled, lease, c, []value.Value{{Lo: input}}, 0, false, 64)
						if exit != native.Return || len(got) != 1 || got[0] != (value.Value{Lo: want}) {
							t.Fatalf("op=%x count=%d dst=%d input=%x: %v %v want=%x", op, count, dst, input, exit, got, want)
						}
						// Enter directly at the shift with a different operand. The
						// preceding constant did not execute and cannot be assumed.
						writeSlot(t, c, 16, 0, value.Value{Lo: input})
						writeSlot(t, c, 16, 1, value.Value{Lo: ^count})
						if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: 1}); err != nil {
							t.Fatal(err)
						}
						want, _ = numeric.Eval(op, input, ^count)
						exit, err := c.Enter(lease, 0)
						if err != nil || exit != native.Return || readSlot(t, c, 16, dst) != (value.Value{Lo: want}) {
							t.Fatalf("independent entry op=%x count=%d dst=%d: %v %v", op, count, dst, exit, err)
						}
					}
					c.Close()
					lease.Close()
				}
			}
		}
	}
}

func TestRegisterPressureAndDeterminism(t *testing.T) {
	// More simultaneously live locals and return values than available registers.
	var body module.Expr
	var results []wasm.ValueType
	for i := uint32(0); i < 32; i++ {
		body = append(body, module.Instr{Op: 0x42, Bits: 1<<63 | uint64(i)}, module.Instr{Op: 0x21, A: i})
	}
	for i := uint32(0); i < 32; i++ {
		body = append(body, module.Instr{Op: 0x20, A: i})
		results = append(results, wasm.I64)
	}
	body = append(body, module.Instr{Op: 0xb})
	m := &module.Module{Types: []module.FuncType{{Results: results}}, Functions: []module.Function{{LocalCount: 32, Locals: []module.Local{{Count: 32, Type: wasm.I64}}, Body: body}}}
	f := functionModule(t, m).Functions[0]
	for _, arch := range []string{"amd64", "arm64"} {
		a, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		b, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
		if err != nil || !bytes.Equal(a.Entry.Code, b.Entry.Code) {
			t.Fatal("unstable code", err)
		}
	}
	if !native.Available {
		return
	}
	compiled, lease, c := functionContext(t, f, true)
	got, exit, _ := runFunction(t, compiled, lease, c, nil, 0, false, 0)
	if exit != native.Return || len(got) != 32 {
		t.Fatal(exit, got)
	}
	for i, v := range got {
		if v != (value.Value{Lo: 1<<63 | uint64(i)}) {
			t.Fatal(i, v)
		}
	}
	if err := c.SetState(native.State{Payload: [4]uint64{0, 0, 0, 1}}); err != nil {
		t.Fatal(err)
	}
	if exit, err := c.Enter(lease, 0); err != nil || exit != native.BadState {
		t.Fatal("metered state accepted by unmetered code", exit, err)
	}
}

func TestUnmeteredHostProgressAndCancellation(t *testing.T) {
	if !native.Available {
		return
	}
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)
	m := functionFixture(module.FuncType{}, 1, module.Expr{{Op: 3, Block: -64}, {Op: 0x20}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0x21}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}})
	f := functionModule(t, m).Functions[0]
	_, lease, c := functionContext(t, f, true)
	done := make(chan native.Exit, 1)
	go func() {
		exit, err := c.Enter(lease, 0)
		if err != nil {
			t.Error(err)
		}
		done <- exit
	}()
	defer c.Cancel()
	deadline := time.After(5 * time.Second)
	for {
		if _, err := c.State(); err == native.ErrBusy {
			break
		}
		select {
		case <-deadline:
			t.Fatal("native entry did not start")
		default:
			runtime.Gosched()
		}
	}
	// These host operations must complete while the guest remains inside one
	// native entry. GOMAXPROCS=1 catches accidental dependence on a spare Go P.
	progress := make(chan struct{})
	go func() {
		for i := 0; i < 3; i++ {
			runtime.GC()
			<-time.After(time.Millisecond)
		}
		close(progress)
	}()
	select {
	case <-progress:
	case <-deadline:
		t.Fatal("native entry starved host GC/timers")
	}
	started := time.Now()
	c.Cancel()
	select {
	case exit := <-done:
		if exit != native.Poll {
			t.Fatal(exit)
		}
		s, err := c.State()
		if err != nil {
			t.Fatal(err)
		}
		if s.Continuation >= uint64(len(f.Code)) || readSlot(t, c, 0, 0).Lo == 0 {
			t.Fatal("missing continuation/register spill", s)
		}
		t.Logf("cancel-to-exit: %s; diagnostic, not a controlled latency acceptance", time.Since(started))
	case <-deadline:
		t.Fatal("native loop did not acknowledge cancellation")
	}
}

func TestNativeConstantMaterialization(t *testing.T) {
	if !native.Available {
		return
	}
	bits := []uint64{0, ^uint64(0), 0xffffffff, 0xffff0000, 0xffffffffffff0000, 0xffff0000ffffffff, 1 << 63}
	seed := uint64(0x123456789abcdef0)
	for i := 0; i < 128; i++ {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		bits = append(bits, seed)
	}
	var body module.Expr
	var results []wasm.ValueType
	for _, v := range bits {
		body = append(body, module.Instr{Op: 0x42, Bits: v})
		results = append(results, wasm.I64)
	}
	body = append(body, module.Instr{Op: 0xb})
	f := functionModule(t, functionFixture(module.FuncType{Results: results}, 0, body)).Functions[0]
	for _, fast := range []bool{false, true} {
		compiled, lease, c := functionContext(t, f, fast)
		got, exit, _ := runFunction(t, compiled, lease, c, nil, 0, false, 64)
		if exit != native.Return || len(got) != len(bits) {
			t.Fatal(exit, got)
		}
		for i, v := range got {
			if v != (value.Value{Lo: bits[i]}) {
				t.Fatalf("constant %x: %v", bits[i], v)
			}
		}
	}
}
