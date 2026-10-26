package codegen

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestMemoryRangeProofs(t *testing.T) {
	get := module.Instr{Op: 0x20}
	constant := func(n uint64) module.Instr { return module.Instr{Op: 0x41, Bits: n} }
	tests := []struct {
		name    string
		address module.Expr
		offset  uint32
		proven  bool
	}{
		{"constant", module.Expr{constant(65532)}, 0, true},
		{"offset", module.Expr{constant(0)}, 65532, true},
		{"past-end", module.Expr{constant(65533)}, 0, false},
		{"large-offset", module.Expr{constant(1)}, 0xffffffff, false},
		{"mask", module.Expr{get, constant(255), {Op: 0x71}}, 3, true},
		{"add", module.Expr{get, constant(255), {Op: 0x71}, constant(100), {Op: 0x6a}}, 0, true},
		{"multiply", module.Expr{get, constant(8191), {Op: 0x71}, constant(4), {Op: 0x6c}}, 0, true},
		{"shift-left", module.Expr{get, constant(8191), {Op: 0x71}, constant(2), {Op: 0x74}}, 0, true},
		{"shift-right", module.Expr{get, constant(18), {Op: 0x76}}, 0, true},
		{"wrapping-add", module.Expr{get, constant(255), {Op: 0x71}, constant(0xffffffff), {Op: 0x6a}}, 0, false},
		{"wrapping-multiply", module.Expr{get, constant(255), {Op: 0x71}, constant(0xffffffff), {Op: 0x6c}}, 0, false},
		{"wrapping-shift", module.Expr{get, constant(2), {Op: 0x74}}, 0, false},
		{"variable-shift", module.Expr{get, get, {Op: 0x76}}, 0, false},
		{"comparison", module.Expr{get, constant(42), {Op: 0x46}}, 0, true},
		{"narrow-load", module.Expr{constant(0), {Op: 0x2d}}, 0, true},
		{"wide-narrow-load", module.Expr{constant(0), {Op: 0x2f}, constant(3), {Op: 0x76}}, 0, true},
		{"unknown", module.Expr{get}, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := append(module.Expr{}, tc.address...)
			body = append(body, module.Instr{Op: 0x22, A: 2}, module.Instr{Op: 0x20, A: 1}, module.Instr{Op: 0x36, B: tc.offset}, module.Instr{Op: 0x20, A: 2}, module.Instr{Op: 0x28, B: tc.offset}, module.Instr{Op: 0xb})
			m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, 1, body)
			m.Memories = []module.Limits{{Min: 1}}
			f := functionModule(t, m).Functions[0]
			config := Config{Unmetered: true, MemoryMinimumPages: 1}
			for _, arch := range []string{"amd64", "arm64"} {
				c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
				if err != nil || (c.BoundsElided >= 2) != tc.proven {
					t.Fatal(arch, "range proof", err)
				}
				limit := DefaultLimits()
				limit.MaxCodeBytes = uint64(len(c.Entry.Code)) - 1
				if _, err = CompileFunctionWithConfig(arch, f, limit, config); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal(err)
				}
			}
			if !native.Available {
				return
			}
			optimized, lease, ctx := configuredContext(t, f, config)
			mem := bindMemory(t, ctx, 65536, 65536)
			reference, refLease, refCtx := functionContext(t, f, true)
			refMem := bindMemory(t, refCtx, 65536, 65536)
			initial := make([]byte, 65536)
			for i := range initial {
				initial[i] = byte(i*151 + 29)
			}
			for _, arg := range []uint64{0, 1, 3, 255, 8191, 65532, 65535, 65536, 0xfffffffe, 0xffffffff, 1 << 32, 1 << 63} {
				mem.Write(0, initial)
				refMem.Write(0, initial)
				args := []value.Value{{Lo: arg}, {Lo: arg ^ 0xabcd1234}}
				got, exit, state := runFunction(t, optimized, lease, ctx, args, 100, false, 64)
				want, wantExit, wantState := runFunction(t, reference, refLease, refCtx, args, 100, false, 64)
				if exit != wantExit || len(got) != len(want) || len(got) != 0 && got[0] != want[0] || exit == native.Trap && (state.Source != wantState.Source || state.Payload[0] != wantState.Payload[0]) {
					t.Fatal(arg, got, want, exit, wantExit, state, wantState)
				}
				gotBytes, wantBytes := make([]byte, 65536), make([]byte, 65536)
				mem.Read(0, gotBytes)
				refMem.Read(0, wantBytes)
				if !bytes.Equal(gotBytes, wantBytes) {
					t.Fatal("side effects", arg)
				}
			}
			// Each original memory PC is a checked continuation, even when normal
			// entry proves its address safe. No proof survives arbitrary resumption.
			for pc, in := range f.Code {
				if in.Op < 0x28 || in.Op > 0x3e {
					continue
				}
				writeSlot(t, ctx, 16, in.A, value.Value{Lo: 0xffffffff})
				if err := ctx.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(pc)}); err != nil {
					t.Fatal(err)
				}
				if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Trap {
					t.Fatal(pc, exit, err)
				}
			}
			// A smaller binding cannot satisfy the compilation assumption.
			small := bindMemory(t, ctx, 65535, 65535)
			if _, exit, _ := runFunction(t, optimized, lease, ctx, nil, 0, false, 0); exit != native.BadState {
				t.Fatal(exit)
			}
			small.Close()
		})
	}
}

func TestMemoryProofBoundaries(t *testing.T) {
	// A helper, growth or CFG transition must discard facts. Even a no-op
	// helper can return to a frame the runtime has changed while suspended.
	poll := module.Expr{}
	for i := 0; i < 260; i++ {
		poll = append(poll, module.Instr{Op: 0x41}, module.Instr{Op: 0x1a})
	}
	for _, separator := range []module.Expr{
		poll,
		{{Op: 0x23}, {Op: 0x1a}},
		{{Op: 0x41}, {Op: 0x40}, {Op: 0x1a}},
		{{Op: 2, Block: -64}, {Op: 0xc}, {Op: 0xb}},
	} {
		body := module.Expr{{Op: 0x41, Bits: 32}, {Op: 0x21}}
		body = append(body, separator...)
		body = append(body, module.Instr{Op: 0x20}, module.Instr{Op: 0x28}, module.Instr{Op: 0xb})
		m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32}}, 1, body)
		m.Memories = []module.Limits{{Min: 1}}
		m.Globals = []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}
		f := functionModule(t, m).Functions[0]
		for _, arch := range []string{"amd64", "arm64"} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{Unmetered: true, MemoryMinimumPages: 1})
			if err != nil || c.BoundsElided != 0 {
				t.Fatal("cross-boundary proof", err)
			}
		}
	}
	f, _ := memoryFunction(t, 0x28, 0)
	if _, err := CompileFunctionWithConfig("arm64", f, DefaultLimits(), Config{MemoryMinimumPages: 65537}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestMemoryRegionEveryContinuation(t *testing.T) {
	if !native.Available {
		return
	}
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32, wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, 2, module.Expr{
		{Op: 0x20}, {Op: 0x41, Bits: 1023}, {Op: 0x71}, {Op: 0x41, Bits: 4}, {Op: 0x6c}, {Op: 0x22, A: 2},
		{Op: 0x28}, {Op: 0x20, A: 1}, {Op: 0x73}, {Op: 0x41, Bits: 0x9e3779b9}, {Op: 0x6a}, {Op: 0x21, A: 3},
		{Op: 0x20, A: 2}, {Op: 0x20, A: 3}, {Op: 0x36}, {Op: 0x20, A: 2}, {Op: 0x28}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	_, lease, ctx := configuredContext(t, f, Config{Unmetered: true, MemoryMinimumPages: 1})
	_, refLease, refCtx := functionContext(t, f, true)
	mem, refMem := bindMemory(t, ctx, 65536, 65536), bindMemory(t, refCtx, 65536, 65536)
	initial := make([]byte, 65536)
	for i := range initial {
		initial[i] = byte(i*31 + 7)
	}
	for pc := range f.Code {
		for _, seed := range []uint64{0, 3, 1024, 0xffffffff, 1 << 32, 0x123456789abcdef} {
			for slot := range f.Slots {
				v := value.Value{Lo: seed + uint64(slot)*17}
				writeSlot(t, ctx, 16, slots.Slot(slot), v)
				writeSlot(t, refCtx, 16, slots.Slot(slot), v)
			}
			if err := mem.Write(0, initial); err != nil {
				t.Fatal(err)
			}
			if err := refMem.Write(0, initial); err != nil {
				t.Fatal(err)
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
				t.Fatal(pc, seed, got, want, gs, ws)
			}
			for slot := range f.Slots {
				g, w := readSlot(t, ctx, 16, slots.Slot(slot)), readSlot(t, refCtx, 16, slots.Slot(slot))
				if g != w {
					t.Fatal("continuation frame", pc, seed, slot, g, w)
				}
			}
			gb, wb := make([]byte, 65536), make([]byte, 65536)
			mem.Read(0, gb)
			refMem.Read(0, wb)
			if !bytes.Equal(gb, wb) {
				t.Fatal("continuation memory", pc, seed)
			}
		}
	}
}

func TestMemoryProofBelongsToInstruction(t *testing.T) {
	// Different lowered instructions can retain the same source location. A
	// proof for one PC must never authorize a later unknown address at that source.
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, 0,
		module.Expr{{Op: 0x41}, {Op: 0x28}, {Op: 0x1a}, {Op: 0x20}, {Op: 0x28}, {Op: 0xb}})
	m.Memories = []module.Limits{{Min: 1}}
	f := functionModule(t, m).Functions[0]
	first := -1
	for pc, in := range f.Code {
		if in.Op == 0x28 {
			if first < 0 {
				first = pc
			} else {
				f.Code[pc].Source = f.Code[first].Source
			}
		}
	}
	config := Config{Unmetered: true, MemoryMinimumPages: 1}
	for _, arch := range []string{"arm64", "amd64"} {
		c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), config)
		if err != nil || c.BoundsElided != 1 {
			t.Fatal(arch, err)
		}
	}
	if !native.Available {
		return
	}
	c, lease, ctx := configuredContext(t, f, config)
	bindMemory(t, ctx, 65536, 65536)
	if _, exit, _ := runFunction(t, c, lease, ctx, []value.Value{{Lo: 0xffffffff}}, 0, false, 0); exit != native.Trap {
		t.Fatal(exit)
	}
}
