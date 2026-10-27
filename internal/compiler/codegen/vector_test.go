package codegen

import (
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/simd"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func vectorTypes(op uint32) ([]wasm.ValueType, wasm.ValueType) {
	p, r := []wasm.ValueType{wasm.V128, wasm.V128}, wasm.V128
	switch {
	case nativeVectorFloatOpcode(0xfd0000 + op):
		if op == 94 || op == 95 || op >= 248 {
			p = p[:1]
		} else if _, _, unary := vectorFloatOperation(op); unary {
			p = p[:1]
		}
	case op <= 20:
		p = []wasm.ValueType{[]wasm.ValueType{wasm.I32, wasm.I32, wasm.I32, wasm.I64, wasm.F32, wasm.F64}[op-15]}
	case op <= 34:
		size, replace, _ := vectorLane(op)
		typ := wasm.I32
		if size == 3 {
			typ = wasm.I64
		}
		if op >= 31 {
			typ = wasm.F32
			if size == 3 {
				typ = wasm.F64
			}
		}
		p = []wasm.ValueType{wasm.V128}
		if replace {
			p = append(p, typ)
		} else {
			r = typ
		}
	case op == 77:
		p = p[:1]
	case op == 82:
		p = append(p, wasm.V128)
	case op == 83:
		p, r = p[:1], wasm.I32
	case op >= 96 && op < 214:
		kind := (op - 96) % 32
		if kind == 0 || kind == 1 {
			p = p[:1]
		}
		if kind == 3 || kind == 4 {
			p, r = p[:1], wasm.I32
		}
		if kind >= 11 && kind <= 13 {
			p[1] = wasm.I32
		}
	}
	return p, r
}

func TestVectorForwardingKeepsEveryColdContinuation(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, op := range []uint32{0xfd00ae, 0xfd00b5, 0xfd0051, 0xfd0052, 0xfd006e, 0xfd008e,
		0xfd00e4, 0xfd00e6, 0xfd00e7, 0xfd00e8, 0xfd00ea, 0xfd00eb,
		0xfd00f0, 0xfd00f2, 0xfd00f3, 0xfd00f4, 0xfd00f6, 0xfd00f7} {
		literal := module.Instr{Op: 0xfd000c}
		binary.LittleEndian.PutUint64(literal.Vector[:8], 0x0123456789abcdef)
		binary.LittleEndian.PutUint64(literal.Vector[8:], 0xfedcba9876543210)
		body := module.Expr{{Op: 0x20}, literal, {Op: op}, {Op: 0x21, A: 1}, {Op: 0x20, A: 1}, {Op: 0xb}}
		if op == 0xfd0052 {
			body = append(module.Expr{{Op: 0x20, A: 1}}, body...)
		} else {
			// The second literal starts a separate combination without local.get.
			// Compare every slot after every possible independent continuation.
			body = append(module.Expr{{Op: 0x20}, literal, {Op: op}}, body[1:]...)
		}
		m := functionModule(t, functionFixture(module.FuncType{Params: []wasm.ValueType{wasm.V128, wasm.V128}, Results: []wasm.ValueType{wasm.V128}}, 0, body))
		f := m.Functions[0]
		fast, fl, fc := functionContext(t, f, true)
		_, sl, sc := functionContext(t, f, false)
		if fast.ForwardedScalars == 0 {
			t.Fatal("vector copies not forwarded", op)
		}
		for pc := range f.Code {
			for i := range f.Slots {
				v := value.Value{Lo: 0xdeadbeef12345678 + uint64(i), Hi: 0xfedcba9876543210 - uint64(i)}
				writeSlot(t, fc, 16, slots.Slot(i), v)
				writeSlot(t, sc, 16, slots.Slot(i), v)
			}
			s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1000}
			if err := fc.SetState(s); err != nil {
				t.Fatal(err)
			}
			if err := sc.SetState(s); err != nil {
				t.Fatal(err)
			}
			if exit, err := fc.Enter(fl, 0); err != nil || exit != native.Return {
				t.Fatal(op, pc, exit, err)
			}
			if exit, err := sc.Enter(sl, 0); err != nil || exit != native.Return {
				t.Fatal(op, pc, exit, err)
			}
			for i := range f.Slots {
				if got, want := readSlot(t, fc, 16, slots.Slot(i)), readSlot(t, sc, 16, slots.Slot(i)); got != want {
					t.Fatal(op, pc, i, got, want)
				}
			}
		}
	}
}

func vectorFixture(op uint32, lane byte, alias, pressure int) (*slots.Function, int) {
	params, result := vectorTypes(op)
	f := &slots.Function{Slots: append([]wasm.ValueType(nil), params...)}
	dst := slots.Slot(len(f.Slots))
	f.Slots = append(f.Slots, result)
	if alias < len(params) && params[alias] == result {
		dst = slots.Slot(alias)
	}
	for i := 0; i < pressure; i++ {
		f.Slots = append(f.Slots, wasm.V128, wasm.F64)
	}
	f.LocalCount = uint32(len(f.Slots))
	for i := range f.Slots {
		f.Args = append(f.Args, slots.Slot(i))
		source := uint32(len(f.Sources))
		f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0x20, A: uint32(i)}})
		count := 1
		if i > len(params) {
			count = 4
		}
		for j := 0; j < count; j++ {
			f.Code = append(f.Code, slots.Instruction{Op: 0x20, Source: source, Dst: slots.Slot(i), A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot})
		}
	}
	pc := len(f.Code)
	in := slots.Instruction{Op: 0xfd0000 + op, Source: uint32(len(f.Sources)), Dst: dst, A: 0, B: slots.NoSlot, C: slots.NoSlot}
	if len(params) >= 2 {
		in.B = 1
	}
	if len(params) == 3 {
		in.C = 2
	}
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: in.Op, Lane: lane, Offset: 123}, Inputs: params, Outputs: []wasm.ValueType{result}})
	f.Code = append(f.Code, in)
	f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: 0xb}})
	f.Code = append(f.Code, slots.Instruction{Op: slots.ControlBase + uint32(ir.Return), Source: uint32(len(f.Sources) - 1), Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: uint32(len(f.Args))})
	return f, pc
}

func TestNativeVectorOperationsAliasesAndSpills(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	seed := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 { seed ^= seed << 13; seed ^= seed >> 7; seed ^= seed << 17; return seed }
	for op := uint32(15); op < 256; op++ {
		if !nativeVectorOpcode(0xfd0000 + op) {
			continue
		}
		lanes := uint32(1)
		if op >= 21 && op <= 34 {
			size, _, _ := vectorLane(op)
			lanes = 16 >> size
		}
		for lane := uint32(0); lane < lanes; lane++ {
			for _, mode := range []struct {
				fast            bool
				alias, pressure int
			}{{false, 3, 0}, {true, 0, 0}, {true, 1, 0}, {true, 2, 0}, {true, 3, 12}} {
				f, pc := vectorFixture(op, byte(lane), mode.alias, mode.pressure)
				compiled, lease, ctx := functionContext(t, f, mode.fast)
				if compiled.NativeVectors != 1 || compiled.Helpers != 0 {
					t.Fatal("vector instruction escaped", op, compiled)
				}
				for sample := 0; sample < 7; sample++ {
					for _, entry := range []int{0, pc, pc + 1} {
						initial := make([]value.Value, len(f.Slots))
						for i, typ := range f.Slots {
							v := value.Value{Lo: next(), Hi: next()}
							if sample == 0 {
								v = value.Value{}
							}
							if sample == 1 {
								v = value.Value{Lo: ^uint64(0), Hi: ^uint64(0)}
							}
							if typ != wasm.V128 {
								v.Hi = 0
							}
							if typ == wasm.I32 || typ == wasm.F32 {
								v.Lo = uint64(uint32(v.Lo))
							}
							initial[i] = v
							writeSlot(t, ctx, 16, slots.Slot(i), v)
						}
						s := native.State{Used: 16, Depth: 1, Continuation: uint64(entry), Budget: 1000, Fuel: 1000}
						if !mode.fast {
							s.Payload[3] = 1
						}
						if err := ctx.SetState(s); err != nil {
							t.Fatal(err)
						}
						if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
							t.Fatal(op, lane, mode, entry, exit, err)
						}
						state, err := ctx.State()
						cost := uint64(0)
						if !mode.fast && entry <= pc {
							cost = uint64(pc - entry + 1)
						}
						if err != nil || state.Fuel != 1000-cost {
							t.Fatal("vector fuel charged per machine instruction", op, entry, state.Fuel, cost, err)
						}
						inputA, inputB := initial[0], value.Value{}
						if f.Code[pc].B != slots.NoSlot {
							inputB = initial[f.Code[pc].B]
						}
						if entry <= pc {
							in := f.Code[pc]
							b, c := value.Value{}, value.Value{}
							if in.B != slots.NoSlot {
								b = initial[in.B]
							}
							if in.C != slots.NoSlot {
								c = initial[in.C]
							}
							initial[in.Dst] = simd.Eval(f.Sources[in.Source].Wasm, initial[in.A], b, c)
						}
						for i, want := range initial {
							if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
								if entry <= pc && slots.Slot(i) == f.Code[pc].Dst && vectorFloatMatches(op, inputA, inputB, got, want) {
									continue
								}
								t.Fatalf("op=%x lane=%d mode=%+v entry=%d slot=%d got=%x want=%x", op, lane, mode, entry, i, got, want)
							}
						}
					}
				}
			}
		}
	}
}

func TestVectorCancellationSpillsAll128Bits(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	f, _ := vectorFixture(174, 0, 3, 12)
	_, lease, ctx := functionContext(t, f, true)
	initial := make([]value.Value, len(f.Slots))
	for i, typ := range f.Slots {
		v := value.Value{Lo: ^uint64(0) - uint64(i), Hi: 0xfedcba9876543210 + uint64(i)}
		if typ != wasm.V128 {
			v.Hi = 0
		}
		initial[i] = v
		writeSlot(t, ctx, 16, slots.Slot(i), v)
	}
	if err := ctx.SetState(native.State{Used: 16, Depth: 1}); err != nil {
		t.Fatal(err)
	}
	ctx.Cancel()
	if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Poll {
		t.Fatal(exit, err)
	}
	for i, want := range initial {
		if got := readSlot(t, ctx, 16, slots.Slot(i)); got != want {
			t.Fatal(i, got, want)
		}
	}
}

func TestVectorLargeFrameAndLiteralBits(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	f := &slots.Function{Slots: make([]wasm.ValueType, 4100), LocalCount: 4100, Args: []slots.Slot{4099}}
	for i := range f.Slots {
		f.Slots[i] = wasm.V128
	}
	v := value.Value{Lo: 0x0123456789abcdef, Hi: 0xfedcba9876543210}
	f.Sources = []ir.Source{{Wasm: module.Instr{Op: 0x20, A: 4098}}, {Wasm: module.Instr{Op: 0xb}}}
	f.Code = []slots.Instruction{{Op: 0x20, Dst: 4099, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}, {Op: slots.ControlBase + uint32(ir.Return), Source: 1, Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Count: 1}}
	for _, fast := range []bool{false, true} {
		_, lease, ctx := functionContext(t, f, fast)
		writeSlot(t, ctx, 16, 4098, v)
		if err := ctx.SetState(native.State{Used: 16, Depth: 1, Budget: 100}); err != nil {
			t.Fatal(err)
		}
		if exit, err := ctx.Enter(lease, 0); err != nil || exit != native.Return {
			t.Fatal(exit, err)
		}
		if got := readSlot(t, ctx, 16, 4099); got != v {
			t.Fatal(runtime.GOARCH, fast, got, v)
		}
	}
}

func TestVectorAffineKeepsIntermediateValuesAndColdEntries(t *testing.T) {
	if !native.Available {
		t.Skip("native platform unavailable")
	}
	for _, size := range []uint32{1, 2} {
		for _, addFirst := range []bool{false, true} {
			for _, prefix := range []bool{false, true} {
				for _, dst := range []uint32{0, 1} {
					f := &slots.Function{LocalCount: 8, Slots: make([]wasm.ValueType, 8)}
					for i := range f.Slots {
						f.Slots[i] = wasm.V128
						f.Args = append(f.Args, slots.Slot(i))
					}
					appendCode := func(source module.Instr, in slots.Instruction) {
						in.Source = uint32(len(f.Sources))
						f.Sources = append(f.Sources, ir.Source{Wasm: source})
						f.Code = append(f.Code, in)
					}
					empty := slots.Instruction{Dst: slots.NoSlot, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot}
					input := slots.Slot(0)
					if prefix {
						in := empty
						in.Op, in.Dst = 0x20, 3
						appendCode(module.Instr{Op: 0x20}, in)
						input = 3
					}
					literal := func(slot slots.Slot, lo, hi uint64) {
						source := module.Instr{Op: 0xfd000c}
						binary.LittleEndian.PutUint64(source.Vector[:8], lo)
						binary.LittleEndian.PutUint64(source.Vector[8:], hi)
						in := empty
						in.Op, in.Dst = source.Op, slot
						appendCode(source, in)
					}
					first, second := uint32(0xfd0095), uint32(0xfd008e)
					if size == 2 {
						first, second = 0xfd00b5, 0xfd00ae
					}
					if addFirst {
						first, second = second, first
					}
					literal(4, 0x123456789abcdef0, 0xfedcba9876543210)
					in := empty
					in.Op, in.Dst, in.A, in.B = first, 5, input, 4
					appendCode(module.Instr{Op: first}, in)
					literal(6, 0x80000001ffffffff, 0x0102030405060708)
					in = empty
					in.Op, in.Dst, in.A, in.B = second, 7, 5, 6
					appendCode(module.Instr{Op: second}, in)
					in = empty
					in.Op, in.A = 0x21, 7
					appendCode(module.Instr{Op: 0x21, A: dst}, in)
					in = empty
					in.Op, in.Count = slots.ControlBase+uint32(ir.Return), 8
					appendCode(module.Instr{Op: 0xb}, in)
					compiled, fl, fc := functionContext(t, f, true)
					_, sl, sc := functionContext(t, f, false)
					if runtime.GOARCH == "arm64" && compiled.FusedVectors != 1 {
						t.Fatal("affine vector not fused", size, addFirst, prefix, dst)
					}
					for pc := range f.Code {
						for i := range f.Slots {
							v := value.Value{Lo: ^uint64(0) - uint64(i), Hi: 0x8000000012345678 + uint64(i)}
							writeSlot(t, fc, 16, slots.Slot(i), v)
							writeSlot(t, sc, 16, slots.Slot(i), v)
						}
						s := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: 1000}
						if err := fc.SetState(s); err != nil {
							t.Fatal(err)
						}
						if err := sc.SetState(s); err != nil {
							t.Fatal(err)
						}
						if exit, err := fc.Enter(fl, 0); err != nil || exit != native.Return {
							t.Fatal(exit, err)
						}
						if exit, err := sc.Enter(sl, 0); err != nil || exit != native.Return {
							t.Fatal(exit, err)
						}
						for i := range f.Slots {
							if got, want := readSlot(t, fc, 16, slots.Slot(i)), readSlot(t, sc, 16, slots.Slot(i)); got != want {
								t.Fatal(size, addFirst, prefix, dst, pc, i, got, want)
							}
						}
					}
				}
			}
		}
	}
}
