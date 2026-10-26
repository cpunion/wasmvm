package codegen

import (
	"encoding/binary"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/interp"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func functionModule(t *testing.T, m *module.Module) *slots.Module {
	t.Helper()
	for fi := range m.Functions {
		for i := range m.Functions[fi].Body {
			m.Functions[fi].Body[i].Offset = uint64(100 + i)
		}
	}
	v, err := validate.Check(m, validate.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	i, err := lower.Compile(v, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	s, err := slots.Compile(i, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func functionFixture(sig module.FuncType, locals uint32, body module.Expr) *module.Module {
	f := module.Function{Body: body, LocalCount: locals}
	if locals != 0 {
		f.Locals = []module.Local{{Count: locals, Type: wasm.I32}}
	}
	return &module.Module{Types: []module.FuncType{sig}, Functions: []module.Function{f}}
}
func sumFixture() *module.Module {
	return functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}, 1, module.Expr{
		{Op: 2, Block: -64}, {Op: 3, Block: -64}, {Op: 0x20}, {Op: 0x45}, {Op: 0xd, A: 1},
		{Op: 0x20, A: 1}, {Op: 0x20}, {Op: 0x6a}, {Op: 0x21, A: 1},
		{Op: 0x20}, {Op: 0x41, Bits: 1}, {Op: 0x6b}, {Op: 0x21}, {Op: 0xc},
		{Op: 0xb}, {Op: 0xb}, {Op: 0x20, A: 1}, {Op: 0xb},
	})
}
func writeSlot(t *testing.T, c *native.Context, base uint64, slot slots.Slot, v value.Value) {
	t.Helper()
	var b [16]byte
	binary.LittleEndian.PutUint64(b[:8], v.Lo)
	binary.LittleEndian.PutUint64(b[8:], v.Hi)
	if err := c.Write(base+uint64(slot)*16, b[:]); err != nil {
		t.Fatal(err)
	}
}
func readSlot(t *testing.T, c *native.Context, base uint64, slot slots.Slot) value.Value {
	t.Helper()
	var b [16]byte
	if err := c.Read(base+uint64(slot)*16, b[:]); err != nil {
		t.Fatal(err)
	}
	return value.Value{Lo: binary.LittleEndian.Uint64(b[:8]), Hi: binary.LittleEndian.Uint64(b[8:])}
}
func functionContext(t *testing.T, f *slots.Function, unmetered ...bool) (*CompiledFunction, *native.Lease, *native.Context) {
	t.Helper()
	config := Config{Unmetered: len(unmetered) != 0 && unmetered[0]}
	return configuredContext(t, f, config)
}
func configuredContext(t *testing.T, f *slots.Function, config Config) (*CompiledFunction, *native.Lease, *native.Context) {
	t.Helper()
	compiled, err := CompileFunctionWithConfig(runtime.GOARCH, f, DefaultLimits(), config)
	if err != nil {
		t.Fatal(err)
	}
	code, err := native.Publish(compiled.Entry.Code, []uint32{0}, DefaultLimits().MaxCodeBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(code.Close)
	lease, err := code.Acquire()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(lease.Close)
	size := uint64(len(f.Slots))*16 + 32
	ctx, err := native.NewContext(size, size, 1024)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ctx.Close)
	return compiled, lease, ctx
}
func runFunction(t *testing.T, compiled *CompiledFunction, lease *native.Lease, c *native.Context, args []value.Value, fuel uint64, metered bool, budget uint64, memories ...*native.Memory) ([]value.Value, native.Exit, native.State) {
	t.Helper()
	base := uint64(16)
	if err := c.Write(0, make([]byte, len(compiled.Function.Slots)*16+32)); err != nil {
		t.Fatal(err)
	}
	for i, v := range args {
		writeSlot(t, c, base, slots.Slot(i), v)
	}
	s := native.State{Used: base, Depth: 1, Fuel: fuel, Budget: budget}
	if metered {
		s.Payload[3] = 1
	}
	if err := c.SetState(s); err != nil {
		t.Fatal(err)
	}
	for slice := 0; slice < 100000; slice++ {
		exit, err := c.Enter(lease, 0)
		if err != nil {
			t.Fatal(err)
		}
		s, err = c.State()
		if err != nil {
			t.Fatal(err)
		}
		if exit == native.Poll {
			s.Budget = budget
			if err := c.SetState(s); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if exit == native.GrowMemory && len(memories) != 0 {
			memory := memories[0]
			if memory == nil {
				t.Fatal("grow without memory")
			}
			in := compiled.Function.Code[s.Payload[0]]
			delta := uint32(readSlot(t, c, base, in.A).Lo)
			size, err := memory.Size()
			if err != nil {
				t.Fatal(err)
			}
			result := size / 65536
			if err = memory.Grow(size + uint64(delta)*65536); err != nil {
				result = 0xffffffff
			}
			writeSlot(t, c, base, in.Dst, value.Value{Lo: result})
			continue
		}
		if exit != native.Return {
			return nil, exit, s
		}
		in := compiled.Function.Code[s.Payload[0]]
		out := make([]value.Value, in.Count)
		for i, slot := range compiled.Function.Args[in.Extra : in.Extra+in.Count] {
			out[i] = readSlot(t, c, base, slot)
		}
		return out, exit, s
	}
	t.Fatal("native function exceeded bounded slices")
	return nil, 0, s
}

func TestNativeFunctionControlAndFuel(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "baseline"
		if fast {
			name = "unmetered"
		}
		t.Run(name, func(t *testing.T) {
			testNativeFunctionControlAndFuel(t, fast)
		})
	}
}
func testNativeFunctionControlAndFuel(t *testing.T, fast bool) {
	if !native.Available {
		return
	}
	sig := module.FuncType{Params: []wasm.ValueType{wasm.I32}, Results: []wasm.ValueType{wasm.I32}}
	fixtures := []*module.Module{
		sumFixture(),
		functionFixture(sig, 0, module.Expr{{Op: 0x20}, {Op: 4, Block: -1}, {Op: 0x41, Bits: 42}, {Op: 5}, {Op: 0x41, Bits: 99}, {Op: 0xb}, {Op: 0xb}}),
		functionFixture(sig, 1, module.Expr{{Op: 2, Block: -64}, {Op: 2, Block: -64}, {Op: 0x41, Bits: 11}, {Op: 0x21, A: 1}, {Op: 0x20}, {Op: 0xe, Labels: []uint32{0, 1}, A: 1}, {Op: 0xb}, {Op: 0x41, Bits: 22}, {Op: 0x21, A: 1}, {Op: 0xb}, {Op: 0x20, A: 1}, {Op: 0xb}}),
		functionFixture(sig, 0, module.Expr{{Op: 0x20}, {Op: 0x41}, {Op: 0x6d}, {Op: 0xb}}),
		functionFixture(sig, 0, module.Expr{{Op: 0}, {Op: 0xb}}),
	}
	for index, m := range fixtures {
		code := functionModule(t, m)
		f := code.Functions[0]
		compiled, lease, c := functionContext(t, f, fast)
		store := state.New(state.DefaultLimits())
		defer store.Close()
		instance, err := store.Allocate(code, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := instance.Initialize(); err != nil {
			t.Fatal(err)
		}
		for _, arg := range []uint64{0, 1, 2, 17} {
			for _, budget := range []uint64{1, 64} {
				for _, fuel := range []uint64{0, 1, 2, 3, 5, 17, 1000} {
					args := []value.Value{{Lo: arg}}
					for _, metered := range []bool{false, true} {
						if fast && metered {
							continue
						}
						ref := interp.New(store, nil, interp.Config{Limits: interp.DefaultLimits(), Metered: metered, Fuel: fuel})
						if err := ref.Start(instance.Functions[0], args); err != nil {
							t.Fatal(err)
						}
						status := ref.Run()
						got, exit, s := runFunction(t, compiled, lease, c, args, fuel, metered, budget)
						if s.Fuel != ref.Fuel() {
							t.Fatalf("fixture %d fuel %d/%d (%d,%v,%d)", index, s.Fuel, ref.Fuel(), fuel, metered, budget)
						}
						if status == interp.Done {
							if exit != native.Return || len(got) != len(ref.Results()) {
								t.Fatal(index, exit, got, ref.Results(), s)
							}
							for i, v := range got {
								if v != ref.Results()[i] {
									t.Fatal(index, arg, v, ref.Results()[i])
								}
							}
						} else {
							var trap *wasm.Trap
							var limit *wasm.Error
							if errors.As(ref.Err(), &trap) {
								if exit != native.Trap || s.Payload[0] != uint64(trap.Code) || s.Source != trap.Offset {
									t.Fatal(index, exit, s, trap)
								}
							} else if errors.As(ref.Err(), &limit) {
								if exit != native.FuelExhausted || s.Source != limit.Offset {
									t.Fatal(index, exit, s, limit)
								}
							} else {
								t.Fatal(ref.Err())
							}
						}
					}
				}
			}
		}
	}
}

func TestNativeFunctionValuesAndExits(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "baseline"
		if fast {
			name = "unmetered"
		}
		t.Run(name, func(t *testing.T) {
			testNativeFunctionValuesAndExits(t, fast)
		})
	}
}
func testNativeFunctionValuesAndExits(t *testing.T, fast bool) {
	if !native.Available {
		return
	}
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.I64, wasm.V128, wasm.FuncRef, wasm.ExternRef} {
		m := functionFixture(module.FuncType{Params: []wasm.ValueType{typ, typ, wasm.I32}, Results: []wasm.ValueType{typ, typ}}, 0,
			module.Expr{{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 0x20, A: 2}, {Op: 0x1c, Types: []wasm.ValueType{typ}}, {Op: 0x20}, {Op: 0xb}})
		f := functionModule(t, m).Functions[0]
		compiled, lease, c := functionContext(t, f, fast)
		if typ.IsReference() && len(compiled.ReferenceSlots) == 0 {
			t.Fatal("missing reference map")
		}
		for _, condition := range []uint64{0, 1} {
			a, b := value.Value{Lo: 11, Hi: 22}, value.Value{Lo: 33, Hi: 44}
			if fast && typ != wasm.V128 {
				a.Hi, b.Hi = 0, 0
			}
			got, exit, _ := runFunction(t, compiled, lease, c, []value.Value{a, b, {Lo: condition}}, 0, false, 64)
			want := b
			if condition != 0 {
				want = a
			}
			if exit != native.Return || len(got) != 2 || got[0] != want || got[1] != a {
				t.Fatal(typ, got, exit)
			}
		}
	}
	// A suspended helper owns no hidden live registers; results written to the
	// declared frame slot survive re-entry through the continuation dispatcher.
	m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32}}, 0, module.Expr{{Op: 0x23}, {Op: 0x41, Bits: 1}, {Op: 0x6a}, {Op: 0xb}})
	m.Globals = []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: module.Expr{{Op: 0x41}, {Op: 0xb}}}}
	f := functionModule(t, m).Functions[0]
	compiled, lease, c := functionContext(t, f, fast)
	_, exit, s := runFunction(t, compiled, lease, c, nil, 10, !fast, 64)
	if exit != native.RuntimeHelper || compiled.Helpers != 1 || (s.Fuel != 9 && !fast || s.Fuel != 10 && fast) {
		t.Fatal(exit, s, compiled.Helpers)
	}
	in := f.Code[s.Payload[0]]
	writeSlot(t, c, s.Used, in.Dst, value.Value{Lo: 41})
	if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
		t.Fatal(exit, err)
	}
	s, _ = c.State()
	ret := f.Code[s.Payload[0]]
	if v := readSlot(t, c, s.Used, f.Args[ret.Extra]); v.Lo != 42 || (s.Fuel != 7 && !fast || s.Fuel != 10 && fast) {
		t.Fatal(v, s)
	}
	// Calls have a distinct exit and retain their complete argument/result plan.
	m = functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32}}, 0, module.Expr{{Op: 0x10}, {Op: 0xb}})
	m.Imports = []module.Import{{Kind: 0}}
	f = functionModule(t, m).Functions[0]
	compiled, lease, c = functionContext(t, f, fast)
	_, exit, s = runFunction(t, compiled, lease, c, nil, 1, !fast, 64)
	if exit != native.Call || compiled.Calls != 1 || s.Continuation != 1 {
		t.Fatal(exit, s)
	}
}

func TestFunctionLimitsAndInvalidState(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "baseline"
		if fast {
			name = "unmetered"
		}
		t.Run(name, func(t *testing.T) {
			testFunctionLimitsAndInvalidState(t, fast)
		})
	}
}
func testFunctionLimitsAndInvalidState(t *testing.T, fast bool) {
	compile := CompileFunction
	if fast {
		compile = CompileUnmeteredFunction
	}
	f := functionModule(t, sumFixture()).Functions[0]
	for _, arch := range []string{"amd64", "arm64"} {
		compiled, err := compile(arch, f, DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		for _, limits := range []Limits{{0, 1000, 1000}, {100000, 0, 1000}, {100000, 1000, 0}} {
			if _, err := compile(arch, f, limits); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal(limits, err)
			}
		}
		limits := DefaultLimits()
		limits.MaxCodeBytes = uint64(len(compiled.Entry.Code))
		if _, err := compile(arch, f, limits); err != nil {
			t.Fatal(err)
		}
		limits.MaxCodeBytes--
		if _, err := compile(arch, f, limits); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	if _, err := compile("bad", f, DefaultLimits()); err != wasm.ErrUnsupported {
		t.Fatal(err)
	}
	if _, err := compile("amd64", nil, DefaultLimits()); err != wasm.ErrInvalid {
		t.Fatal(err)
	}
	if !native.Available {
		return
	}
	_, lease, c := functionContext(t, f, fast)
	for _, s := range []native.State{{Used: 1}, {Used: uint64(len(f.Slots))*16 + 32}, {Continuation: uint64(len(f.Code))}} {
		if err := c.SetState(s); err != nil {
			t.Fatal(err)
		}
		if exit, err := c.Enter(lease, 0); err != nil || exit != native.BadState {
			t.Fatal(exit, err, s)
		}
	}
	if err := c.SetState(native.State{Budget: 64}); err != nil {
		t.Fatal(err)
	}
	c.Cancel()
	if exit, err := c.Enter(lease, 0); err != nil || exit != native.Poll {
		t.Fatal(exit, err)
	}
}

func TestFunctionLargeOffsetsAndBranches(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "baseline"
		if fast {
			name = "unmetered"
		}
		t.Run(name, func(t *testing.T) {
			testFunctionLargeOffsetsAndBranches(t, fast)
		})
	}
}
func testFunctionLargeOffsetsAndBranches(t *testing.T, fast bool) {
	compile := CompileFunction
	if fast {
		compile = CompileUnmeteredFunction
	}
	m := functionFixture(module.FuncType{Results: []wasm.ValueType{wasm.I32}}, 4097, module.Expr{
		{Op: 0x41, Bits: 42}, {Op: 0x22, A: 4096}, {Op: 0x1a}, {Op: 0x20, A: 4096}, {Op: 0xb},
	})
	f := functionModule(t, m).Functions[0]
	for _, arch := range []string{"amd64", "arm64"} {
		if _, err := compile(arch, f, DefaultLimits()); err != nil {
			t.Fatal(err)
		}
	}
	if native.Available {
		compiled, lease, c := functionContext(t, f, fast)
		got, exit, _ := runFunction(t, compiled, lease, c, nil, 5, !fast, 64)
		if exit != native.Return || len(got) != 1 || got[0].Lo != 42 {
			t.Fatal(exit, got)
		}
	}
	// Conditional branches to the shared exit must reach beyond AArch64's
	// short conditional range without narrowing the supported function profile.
	long := &slots.Function{Sources: []ir.Source{{Wasm: module.Instr{Op: 1}}, {Wasm: module.Instr{Op: 0xb}}}}
	for i := 0; i < 12000; i++ {
		long.Code = append(long.Code, slots.Instruction{Op: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot})
	}
	long.Code = append(long.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot})
	for _, arch := range []string{"amd64", "arm64"} {
		if _, err := compile(arch, long, DefaultLimits()); err != nil {
			t.Fatal(arch, err)
		}
	}
	if native.Available {
		compiled, lease, c := functionContext(t, long, fast)
		_, exit, s := runFunction(t, compiled, lease, c, nil, 12000, !fast, 64)
		if exit != native.Return || (s.Fuel != 0 && !fast || s.Fuel != 12000 && fast) {
			t.Fatal(exit, s)
		}
	}
}

func TestFunctionCancelAndLargeMoves(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "baseline"
		if fast {
			name = "unmetered"
		}
		t.Run(name, func(t *testing.T) {
			testFunctionCancelAndLargeMoves(t, fast)
		})
	}
}
func testFunctionCancelAndLargeMoves(t *testing.T, fast bool) {
	if !native.Available {
		return
	}
	m := functionFixture(module.FuncType{}, 0, module.Expr{{Op: 3, Block: -64}, {Op: 0xc}, {Op: 0xb}, {Op: 0xb}})
	f := functionModule(t, m).Functions[0]
	_, lease, c := functionContext(t, f, fast)
	if err := c.SetState(native.State{Budget: ^uint64(0)}); err != nil {
		t.Fatal(err)
	}
	done := make(chan native.Exit, 1)
	go func() {
		exit, err := c.Enter(lease, 0)
		if err != nil {
			t.Error(err)
		}
		done <- exit
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := c.State(); err == native.ErrBusy {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("native loop did not enter")
		}
		runtime.Gosched()
	}
	c.Cancel()
	select {
	case exit := <-done:
		if exit != native.Poll {
			t.Fatal(exit)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("native backedge did not observe cancellation")
	}
	// Large tuple copies leave native code before copying, preserving a bounded
	// native slice even when one guest branch carries thousands of values.
	f = &slots.Function{Slots: make([]wasm.ValueType, 258), Sources: []ir.Source{{Wasm: module.Instr{Op: 0xc}}, {Wasm: module.Instr{Op: 0xb}}},
		Code: []slots.Instruction{{Op: slots.ControlBase + uint32(ir.Jump), Count: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot}, {Op: slots.ControlBase + uint32(ir.Return), Source: 1, Count: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot}}, Args: []slots.Slot{257}, Edges: []slots.Edge{{PC: 1, Count: 257}}}
	for i := range f.Slots {
		f.Slots[i] = wasm.I64
	}
	for i := 1; i < len(f.Slots); i++ {
		f.Moves = append(f.Moves, slots.Move{Dst: slots.Slot(i), Src: slots.Slot(i - 1)})
	}
	compiled, lease, c := functionContext(t, f, fast)
	if compiled.EdgeCopies != 1 {
		t.Fatal(compiled.EdgeCopies)
	}
	writeSlot(t, c, 0, 0, value.Value{Lo: 42})
	if err := c.SetState(native.State{Budget: 64}); err != nil {
		t.Fatal(err)
	}
	if exit, err := c.Enter(lease, 0); err != nil || exit != native.Move {
		t.Fatal(exit, err)
	}
	s, _ := c.State()
	if s.Payload[0] != 0 || s.Continuation != 1 {
		t.Fatal(s)
	}
	for _, move := range f.Moves {
		writeSlot(t, c, 0, move.Dst, readSlot(t, c, 0, move.Src))
	}
	if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
		t.Fatal(exit, err)
	}
	if v := readSlot(t, c, 0, 257); v.Lo != 42 {
		t.Fatal(v)
	}
}

func TestFunctionTupleCycles(t *testing.T) {
	for _, fast := range []bool{false, true} {
		name := "baseline"
		if fast {
			name = "unmetered"
		}
		t.Run(name, func(t *testing.T) {
			testFunctionTupleCycles(t, fast)
		})
	}
}
func testFunctionTupleCycles(t *testing.T, fast bool) {
	if !native.Available {
		return
	}
	for _, typ := range []wasm.ValueType{wasm.I32, wasm.V128, wasm.ExternRef} {
		m := &module.Module{Types: []module.FuncType{{Params: []wasm.ValueType{typ, typ, wasm.I32}, Results: []wasm.ValueType{typ, typ}}, {Params: []wasm.ValueType{typ, typ}, Results: []wasm.ValueType{typ, typ}}}, Functions: []module.Function{{Locals: []module.Local{{Count: 2, Type: typ}}, LocalCount: 2, Body: module.Expr{
			{Op: 0x20}, {Op: 0x20, A: 1}, {Op: 3, Block: 1}, {Op: 0x21, A: 4}, {Op: 0x21, A: 3}, {Op: 0x20, A: 4}, {Op: 0x20, A: 3},
			{Op: 0x20, A: 2}, {Op: 0x41, Bits: 1}, {Op: 0x6b}, {Op: 0x22, A: 2}, {Op: 0xd}, {Op: 0xb}, {Op: 0xb},
		}}}}
		compiled, lease, c := functionContext(t, functionModule(t, m).Functions[0], fast)
		a, b := value.Value{Lo: 11, Hi: 22}, value.Value{Lo: 33, Hi: 44}
		if fast && typ != wasm.V128 {
			a.Hi, b.Hi = 0, 0
		}
		for n := uint64(1); n <= 4; n++ {
			got, exit, _ := runFunction(t, compiled, lease, c, []value.Value{a, b, {Lo: n}}, 1000, !fast, 1)
			want := []value.Value{a, b}
			if n%2 != 0 {
				want = []value.Value{b, a}
			}
			if exit != native.Return || len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
				t.Fatal(typ, n, exit, got, want)
			}
		}
	}
}
