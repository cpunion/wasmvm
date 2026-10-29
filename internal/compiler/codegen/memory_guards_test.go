package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func guardedMemoryFunction(t *testing.T, op uint32, lane byte, biases, offsets []uint32) *slots.Function {
	t.Helper()
	var base *slots.Function
	if nativeVectorMemoryOpcode(op) {
		base, _ = vectorMemoryFunction(t, op, 0, lane, false, false)
	} else {
		base, _ = memoryFunction(t, op, 0)
	}
	load, _ := memoryAccess(op)
	var body module.Expr
	for i, bias := range biases {
		body = append(body, module.Instr{Op: 0x20})
		if bias != 0 {
			body = append(body, module.Instr{Op: 0x41, Bits: uint64(bias)}, module.Instr{Op: 0x6a})
		}
		if !load || op >= 0xfd0054 && op <= 0xfd0057 {
			body = append(body, module.Instr{Op: 0x20, A: 1})
		}
		body = append(body, module.Instr{Op: op, B: offsets[i], Lane: lane})
		if load && i != len(biases)-1 {
			body = append(body, module.Instr{Op: 0x1a})
		}
	}
	body = append(body, module.Instr{Op: 0xb})
	m := functionFixture(base.Signature, 0, body)
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
	return functionModule(t, m).Functions[0]
}

func compareGuardedExecution(t *testing.T, f *slots.Function, addresses []uint64) {
	t.Helper()
	if !native.Available {
		return
	}
	_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	_, refLease, refCtx := functionContext(t, f, true)
	mem, refMem := bindMemory(t, ctx, 65536, 65536), bindMemory(t, refCtx, 65536, 65536)
	initial := make([]byte, 65536)
	for i := range initial {
		initial[i] = byte(i*193 + 127)
	}
	for pc := range f.Code {
		for _, seed := range addresses {
			mem.Write(0, initial)
			refMem.Write(0, initial)
			for s, typ := range f.Slots {
				v := value.Value{Lo: seed + uint64(s)*17}
				if typ == wasm.V128 {
					v.Hi = ^v.Lo
				}
				writeSlot(t, ctx, 16, slots.Slot(s), v)
				writeSlot(t, refCtx, 16, slots.Slot(s), v)
			}
			state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}
			if err := ctx.SetState(state); err != nil {
				t.Fatal(err)
			}
			if err := refCtx.SetState(state); err != nil {
				t.Fatal(err)
			}
			got, err := ctx.Enter(lease, 0)
			if err != nil {
				t.Fatal(err)
			}
			want, err := refCtx.Enter(refLease, 0)
			if err != nil {
				t.Fatal(err)
			}
			gs, _ := ctx.State()
			ws, _ := refCtx.State()
			if got != want || gs != ws {
				t.Fatal("state/trap order", pc, seed, got, want, gs, ws)
			}
			for s := range f.Slots {
				g, w := readSlot(t, ctx, 16, slots.Slot(s)), readSlot(t, refCtx, 16, slots.Slot(s))
				if g != w {
					t.Fatal("cold frame", pc, seed, s, g, w)
				}
			}
			gb, wb := make([]byte, 65536), make([]byte, 65536)
			mem.Read(0, gb)
			refMem.Read(0, wb)
			if !bytes.Equal(gb, wb) {
				t.Fatal("memory effects before trap", pc, seed)
			}
		}
	}
}

func TestGuardedMemoryAllWidthsAndLanes(t *testing.T) {
	var ops []uint32
	for op := uint32(0x28); op <= 0x3e; op++ {
		ops = append(ops, op)
	}
	for op := uint32(0xfd0000); op <= 0xfd005d; op++ {
		if nativeVectorMemoryOpcode(op) {
			ops = append(ops, op)
		}
	}
	for _, op := range ops {
		_, size := memoryAccess(op)
		lanes := uint32(1)
		if op >= 0xfd0054 && op <= 0xfd005b {
			lanes = 16 / size
		}
		for lane := uint32(0); lane < lanes; lane++ {
			t.Run(fmt.Sprintf("%x/%d", op, lane), func(t *testing.T) {
				f := guardedMemoryFunction(t, op, byte(lane), []uint32{0, 0, 0}, []uint32{0, 1, 2})
				for _, arch := range []string{"amd64", "arm64"} {
					cfg := Config{Unmetered: true, MemoryMinimumPages: 1}
					c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), cfg)
					if err != nil || c.MemoryGuards != 1 || c.GuardedAccesses != 3 || c.BoundsElided != 3 {
						t.Fatal(arch, c, err)
					}
					limits := DefaultLimits()
					limits.MaxCodeBytes = uint64(len(c.Entry.Code))
					if _, err := CompileFunctionWithConfig(arch, f, limits, cfg); err != nil {
						t.Fatal(err)
					}
					limits.MaxCodeBytes--
					if c, err := CompileFunctionWithConfig(arch, f, limits, cfg); c != nil || !errors.Is(err, wasm.ErrLimit) {
						t.Fatal("partial limited code", c, err)
					}
					c, err = CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{MemoryMinimumPages: 1})
					if err != nil || c.MemoryGuards != 0 || c.BoundsElided != 0 {
						t.Fatal("precise tier", c, err)
					}
				}
				compareGuardedExecution(t, f, []uint64{0, 1, 65536 - uint64(size) - 2, 65536 - uint64(size) - 1, 65536 - uint64(size), 0xffffffff, 1 << 32})
			})
		}
	}
}

func TestGuardedAffineWrapAndOffsets(t *testing.T) {
	for _, offsets := range [][]uint32{{0, 0, 0, 0}, {3, 1, 4096, 2}, {0xffffffff, 0, 0, 0}} {
		f := guardedMemoryFunction(t, 0x2d, 0, []uint32{33, 1, 65, 2}, offsets)
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			want := uint32(1)
			if offsets[0] == 0xffffffff {
				want = 0
			}
			if err != nil || c.MemoryGuards != want {
				t.Fatal(c, err)
			}
		}
		compareGuardedExecution(t, f, []uint64{0, 1, 65470, 65500, 0xffffffbc, 0xffffffde, 0xfffffffe, 0xffffffff, 1 << 32, 1 << 63})
	}
	// A span wider than the entire memory32 space cannot be guarded.
	f := guardedMemoryFunction(t, 0x28, 0, []uint32{0, 0, 0}, []uint32{0xffffffff, 0, 1})
	c, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
	if err != nil || c.MemoryGuards != 0 {
		t.Fatal(c, err)
	}
	f = guardedMemoryFunction(t, 0x2d, 0, []uint32{0, 0, 0}, []uint32{0xffffffff, 0xffffffff, 0xffffffff})
	compareGuardedExecution(t, f, []uint64{0, 1, 0xffffffff})
}

func TestGuardedFirstPCAndCanonicalPrefix(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		f := &slots.Function{LocalCount: 1, Signature: module.FuncType{Params: []wasm.ValueType{wasm.I32}}, Slots: make([]wasm.ValueType, 7),
			Sources: []ir.Source{{Wasm: module.Instr{Op: 0x41, Bits: 5}}, {Wasm: module.Instr{Op: 0x41, Bits: 7}}, {Wasm: module.Instr{Op: 0x6a}},
				{Wasm: module.Instr{Op: 0x28, Offset: 123}}, {Wasm: module.Instr{Op: 0x28, B: 4, Offset: 124}}, {Wasm: module.Instr{Op: 0x28, B: 8, Offset: 125}}, {Wasm: module.Instr{Op: 0x41}}, {Wasm: module.Instr{Op: 0xb}}}}
		for i := range f.Slots {
			f.Slots[i] = wasm.I32
		}
		if mode == 1 {
			f.Code = []slots.Instruction{{Op: 0x41, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
				{Op: 0x41, Source: 1, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}, {Op: 0x6a, Source: 2, Dst: 3, A: 1, B: 2, C: slots.NoSlot}}
		}
		for i := 0; i < 3; i++ {
			f.Code = append(f.Code, slots.Instruction{Op: 0x28, Source: uint32(3 + i), Dst: slots.Slot(4 + i), A: 0, B: slots.NoSlot, C: slots.NoSlot})
			if mode == 2 && i == 0 {
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x6d, Offset: 333}})
				f.Code = append(f.Code,
					slots.Instruction{Op: 0x41, Source: 0, Dst: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
					slots.Instruction{Op: 0x41, Source: 6, Dst: 2, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot},
					slots.Instruction{Op: 0x6d, Source: 8, Dst: 3, A: 1, B: 2, C: slots.NoSlot})
			}
		}
		for i := 1; i <= 3; i++ {
			f.Code = append(f.Code, slots.Instruction{Op: 0x41, Source: 6, Dst: slots.Slot(i), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
		f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: 7, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		c, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
		if err != nil || c.MemoryGuards != 1 {
			t.Fatal(c, err)
		}
		compareGuardedExecution(t, f, []uint64{0, 1, 65524, 65528, 65532, 0xffffffff})
	}
}

func TestGuardedAddressFoldingKeepsExitSlots(t *testing.T) {
	f := guardedMemoryFunction(t, 0x2d, 0, []uint32{1, 2, 3, 4, 5}, []uint32{0, 0, 0, 0, 0})
	for _, arch := range []string{"amd64", "arm64"} {
		c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
		if err != nil || c.DeadScalarOps == 0 {
			t.Fatal("address arithmetic not removed", arch, c, err)
		}
	}
	compareGuardedExecution(t, f, []uint64{0, 1, 65530, 65532, 0xfffffffe, 0xffffffff})
}

func TestGuardedVectorConstantStoresUseRemappedAddresses(t *testing.T) {
	body := module.Expr{}
	for bias := uint64(1); bias <= 3; bias++ {
		body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x41, Bits: bias}, module.Instr{Op: 0x6a},
			module.Instr{Op: 0xfd000c, Vector: [16]byte{byte(bias), 3, 5, 7, 8: 17, 19, 23, 29}}, module.Instr{Op: 0xfd000b})
	}
	body = append(body, module.Instr{Op: 0xb})
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32}}, 0, body)
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	c, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
	if err != nil || c.MemoryGuards != 1 || c.DeadScalarOps == 0 {
		t.Fatal(c, err)
	}
	compareGuardedExecution(t, f, []uint64{0, 1, 65515, 65517, 0xffffffff})
}

func TestGuardedLocalVersionsAndBoundaries(t *testing.T) {
	load := module.Expr{{Op: 0x20}, {Op: 0x28}, {Op: 0x1a}}
	for _, separator := range []module.Expr{
		{{Op: 0x20, A: 1}, {Op: 0x21}},
		{{Op: 0x23}, {Op: 0x1a}},
		{{Op: 0x41}, {Op: 0x40}, {Op: 0x1a}},
		{{Op: 2, Block: -64}, {Op: 0xb}},
	} {
		body := append(append(module.Expr{}, load...), load...)
		body = append(body, separator...)
		body = append(body, load...)
		body = append(body, load...)
		body = append(body, module.Instr{Op: 0xb})
		m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I32}}, 0, body)
		m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
		m.Globals = []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}
		f := functionModule(t, m).Functions[0]
		c, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
		if err != nil || c.MemoryGuards != 0 {
			t.Fatal("cross-boundary group", c, err)
		}
	}
	// An address computed before mutation cannot inherit its local's new version.
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I32}}, 0, module.Expr{
		{Op: 0x20}, {Op: 0x41, Bits: 32}, {Op: 0x6a}, {Op: 0x20, A: 1}, {Op: 0x21}, {Op: 0x28}, {Op: 0x1a},
		{Op: 0x20}, {Op: 0x28}, {Op: 0x1a}, {Op: 0x20}, {Op: 0x28}, {Op: 0x1a}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	c, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
	if err != nil || c.MemoryGuards != 0 {
		t.Fatal("stale address", c, err)
	}
}

func TestGuardedMemoryRefreshAfterGrowth(t *testing.T) {
	if !native.Available {
		return
	}
	body := module.Expr{}
	for _, offset := range []uint32{0, 1, 2} {
		body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x2d, B: offset}, module.Instr{Op: 0x1a})
	}
	body = append(body, module.Instr{Op: 0x41, Bits: 1}, module.Instr{Op: 0x40}, module.Instr{Op: 0x1a})
	for _, offset := range []uint32{65536, 65537, 65538} {
		body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x2d, B: offset}, module.Instr{Op: 0x1a})
	}
	body = append(body, module.Instr{Op: 0xb})
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32}}, 0, body)
	m.Memories = []module.Limits{{Min: 1, Max: 2, HasMax: true}}
	f := functionModule(t, m).Functions[0]
	c, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	if c.MemoryGuards != 2 {
		t.Fatal(c.MemoryGuards)
	}
	mem := bindMemory(t, ctx, 65536, 131072)
	if _, exit, state := runFunction(t, c, lease, ctx, []value.Value{{}}, 0, false, 64, mem); exit != native.Return {
		t.Fatal(exit, state)
	}
}

func TestGuardedCommutedAddressesAndCancellation(t *testing.T) {
	if !native.Available {
		return
	}
	body := module.Expr{{Op: 3, Block: -64}}
	for i := 0; i < 3; i++ {
		body = append(body, module.Instr{Op: 0x41, Bits: uint64(i)}, module.Instr{Op: 0x20}, module.Instr{Op: 0x6a}, module.Instr{Op: 0x2d}, module.Instr{Op: 0x1a})
	}
	body = append(body, module.Instr{Op: 0xc}, module.Instr{Op: 0xb}, module.Instr{Op: 0xb})
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32}}, 0, body)
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	c, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	if c.MemoryGuards != 1 {
		t.Fatal(c.MemoryGuards)
	}
	bindMemory(t, ctx, 65536, 65536)
	if err := ctx.SetState(native.State{Used: 16, Depth: 1}); err != nil {
		t.Fatal(err)
	}
	type result struct {
		exit native.Exit
		err  error
	}
	done := make(chan result, 1)
	go func() { exit, err := ctx.Enter(lease, 0); done <- result{exit, err} }()
	time.Sleep(10 * time.Millisecond)
	ctx.Cancel()
	select {
	case r := <-done:
		if r.err != nil || r.exit != native.Poll {
			t.Fatal(r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("guarded loop did not cancel")
	}
}

func TestGuardedMemoryBindingRequirements(t *testing.T) {
	if !native.Available {
		return
	}
	f := guardedMemoryFunction(t, 0x2d, 0, []uint32{0, 0, 0}, []uint32{0, 1, 2})
	c, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	bindMemory(t, ctx, 65535, 65536)
	if _, exit, _ := runFunction(t, c, lease, ctx, []value.Value{{}}, 0, false, 64); exit != native.BadState {
		t.Fatal(exit)
	}
	c, lease, ctx = functionContext(t, f, true)
	if c.MemoryGuards != 0 {
		t.Fatal("zero-minimum guard", c.MemoryGuards)
	}
	bindMemory(t, ctx, 0, 65536)
	if _, exit, _ := runFunction(t, c, lease, ctx, []value.Value{{}}, 0, false, 64); exit != native.Trap {
		t.Fatal(exit)
	}
}
