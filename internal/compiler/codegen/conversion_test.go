package codegen

import (
	"errors"
	"math"
	"math/rand"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func conversionOps() []uint32 {
	ops := []uint32{0xa8, 0xa9, 0xaa, 0xab, 0xae, 0xaf, 0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb}
	for op := uint32(0xfc0000); op <= 0xfc0007; op++ {
		ops = append(ops, op)
	}
	return ops
}
func conversionTypes(op uint32) (wasm.ValueType, wasm.ValueType) {
	if op == 0xb6 {
		return wasm.F64, wasm.F32
	}
	if op == 0xbb {
		return wasm.F32, wasm.F64
	}
	if op >= 0xb2 && op <= 0xba {
		output, variant := wasm.F32, op-0xb2
		if op >= 0xb7 {
			output, variant = wasm.F64, op-0xb7
		}
		input := wasm.I32
		if variant >= 2 {
			input = wasm.I64
		}
		return input, output
	}
	variant := op - 0xa8
	if op >= 0xfc0000 {
		variant = op - 0xfc0000
	} else if op >= 0xae {
		variant = op - 0xae + 4
	}
	input, output := wasm.F32, wasm.I32
	if variant&2 != 0 {
		input = wasm.F64
	}
	if variant&4 != 0 {
		output = wasm.I64
	}
	return input, output
}
func conversionFunction(t *testing.T, op uint32) (*slots.Function, int) {
	t.Helper()
	input, output := conversionTypes(op)
	m := functionFixture(module.FuncType{Params: []wasm.ValueType{input}, Results: []wasm.ValueType{output}}, 0,
		module.Expr{{Op: 0x20}, {Op: op}, {Op: 0xb}})
	f := functionModule(t, m).Functions[0]
	for pc, in := range f.Code {
		if in.Op == op {
			return f, pc
		}
	}
	t.Fatal("missing conversion")
	return nil, 0
}
func TestConversionTargetsAndBudgets(t *testing.T) {
	if len(conversionOps()) != 26 {
		t.Fatal("conversion inventory")
	}
	for _, op := range conversionOps() {
		f, _ := conversionFunction(t, op)
		for _, arch := range []string{"amd64", "arm64"} {
			for _, compile := range []func(string, *slots.Function, Limits) (*CompiledFunction, error){CompileFunction, CompileUnmeteredFunction} {
				compiled, err := compile(arch, f, DefaultLimits())
				if err != nil || compiled.Helpers != 0 {
					t.Fatal(arch, op, err)
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
		}
	}
}
func conversionInputs(typ wasm.ValueType) []uint64 {
	values := []uint64{0, 1, 2, 0x7fffffff, 0x80000000, 0xffffffff, 0x100000000, 1 << 63, (1 << 63) - 1, ^uint64(0),
		(1 << 24) - 1, 1 << 24, (1 << 24) + 1, (1 << 53) - 1, 1 << 53, (1 << 53) + 1, 0x7f800000, 0xff800000, 0x7fc00000, 0x7fa12345,
		0x7ff0000000000000, 0xfff0000000000000, 0x7ff8000000000000, 0x7ff123456789abcd, 0xfff8123456789abc}
	if typ == wasm.F32 || typ == wasm.F64 {
		for _, v := range []float64{-1.5, -1, -0.9999999999999999, -0.5, 0, 0.5, 1, 1.5,
			-math.Ldexp(1, 31) - 1, -math.Ldexp(1, 31) - 0.5, -math.Ldexp(1, 31), math.Ldexp(1, 31) - 1, math.Ldexp(1, 31),
			math.Ldexp(1, 32) - 1, math.Ldexp(1, 32), -math.Ldexp(1, 63), math.Ldexp(1, 63), math.Ldexp(1, 64), math.MaxFloat64} {
			if typ == wasm.F32 {
				x := float32(v)
				for _, n := range []float32{x, math.Nextafter32(x, float32(math.Inf(-1))), math.Nextafter32(x, float32(math.Inf(1)))} {
					values = append(values, uint64(math.Float32bits(n)))
				}
			} else {
				for _, n := range []float64{v, math.Nextafter(v, math.Inf(-1)), math.Nextafter(v, math.Inf(1))} {
					values = append(values, math.Float64bits(n))
				}
			}
		}
	}
	rng := rand.New(rand.NewSource(20261027))
	for i := 0; i < 256; i++ {
		values = append(values, rng.Uint64())
	}
	if typ == wasm.I32 || typ == wasm.F32 {
		for i := range values {
			values[i] = uint64(uint32(values[i]))
		}
	}
	return values
}
func nanKind(typ wasm.ValueType, bits uint64) int {
	exponent, mantissa, quiet := uint64(0x7f800000), uint64(0x7fffff), uint64(0x400000)
	if typ == wasm.F64 {
		exponent, mantissa, quiet = 0x7ff0000000000000, 0xfffffffffffff, 0x8000000000000
	} else if typ != wasm.F32 {
		return 0
	}
	if bits&exponent != exponent || bits&mantissa == 0 {
		return 0
	}
	if bits&mantissa == quiet {
		return 1
	}
	if bits&quiet != 0 {
		return 2
	}
	return 3
}
func conversionMatches(input, output wasm.ValueType, arg, got, want uint64) bool {
	if got == want {
		return true
	}
	if nanKind(output, want) == 0 {
		return false
	}
	kind := nanKind(output, got)
	if kind != 1 && kind != 2 {
		return false
	}
	return nanKind(input, arg) > 1 || kind == 1
}
func TestConversionNativeSemanticsAndTraps(t *testing.T) {
	if !native.Available {
		return
	}
	for _, op := range conversionOps() {
		input, output := conversionTypes(op)
		f, pc := conversionFunction(t, op)
		in := f.Code[pc]
		for _, fast := range []bool{false, true} {
			compiled, lease, c := functionContext(t, f, fast)
			for _, arg := range conversionInputs(input) {
				want, trap := numeric.Eval(op, arg, 0)
				for slot := range f.Slots {
					writeSlot(t, c, 16, slots.Slot(slot), value.Value{Lo: 7})
				}
				writeSlot(t, c, 16, in.A, value.Value{Lo: arg})
				state := native.State{Used: 16, Depth: 1, Continuation: uint64(pc), Budget: uint64(len(f.Code)), Fuel: 1}
				if !fast {
					state.Payload[3] = 1
				}
				if err := c.SetState(state); err != nil {
					t.Fatal(err)
				}
				exit, err := c.Enter(lease, 0)
				if err != nil {
					t.Fatal(err)
				}
				state, err = c.State()
				if err != nil {
					t.Fatal(err)
				}
				got := readSlot(t, c, 16, in.Dst)
				if !fast && state.Fuel != 0 {
					t.Fatal("fuel boundary", op, state)
				}
				if trap != wasm.NoTrap {
					if exit != native.Trap || state.Payload[0] != uint64(trap) || state.Source != f.Sources[in.Source].Wasm.Offset || got != (value.Value{Lo: 7}) {
						t.Fatalf("op=%x fast=%v arg=%x trap=%v exit=%v state=%+v result=%x", op, fast, arg, trap, exit, state, got)
					}
				} else if exit != native.Return || got.Hi != 0 || !conversionMatches(input, output, arg, got.Lo, want) {
					t.Fatalf("op=%x fast=%v arg=%x result=%x want=%x exit=%v", op, fast, arg, got, want, exit)
				}
			}
			// Exercise normal entry as well as direct conversion continuations.
			out, exit, _ := runFunction(t, compiled, lease, c, []value.Value{{Lo: 0}}, 2, !fast, 1)
			if exit != native.Return || len(out) != 1 || out[0] != (value.Value{}) {
				t.Fatal(op, fast, out, exit)
			}
			if !fast {
				_, exit, state := runFunction(t, compiled, lease, c, []value.Value{{Lo: 0}}, 1, true, 1)
				if exit != native.FuelExhausted || state.Continuation != uint64(pc) {
					t.Fatal("fuel before conversion", op, exit, state)
				}
			}
			c.Close()
			lease.Close()
		}
	}
}
