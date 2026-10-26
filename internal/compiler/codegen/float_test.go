package codegen

import (
	"errors"
	"math"
	"math/rand"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func floatOps() []uint32 {
	var ops []uint32
	for op := uint32(0); op < 0xc0; op++ {
		if isFloat(op) {
			ops = append(ops, op)
		}
	}
	return ops
}
func floatFixture(op uint32, dst slots.Slot) *slots.Function {
	return &slots.Function{
		Slots:   []wasm.ValueType{wasm.I64, wasm.I64, wasm.I64},
		Sources: []ir.Source{{Wasm: module.Instr{Op: op}}, {Wasm: module.Instr{Op: 0xb}}},
		Args:    []slots.Slot{dst},
		Code: []slots.Instruction{
			{Op: op, A: 0, B: 1, C: slots.NoSlot, Dst: dst},
			{Op: slots.ControlBase + uint32(ir.Return), Source: 1, A: slots.NoSlot, B: slots.NoSlot, C: slots.NoSlot, Dst: slots.NoSlot, Count: 1},
		},
	}
}
func TestFloatTargetsAndBudgets(t *testing.T) {
	if len(floatOps()) != 44 {
		t.Fatal("floating-point instruction inventory")
	}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, compile := range []func(string, *slots.Function, Limits) (*CompiledFunction, error){CompileFunction, CompileUnmeteredFunction} {
			for _, op := range floatOps() {
				f := floatFixture(op, 2)
				compiled, err := compile(arch, f, DefaultLimits())
				if err != nil || compiled.Helpers != 0 {
					t.Fatal(arch, op, err)
				}
				limits := DefaultLimits()
				limits.MaxCodeBytes = uint64(len(compiled.Entry.Code))
				if _, err := compile(arch, f, limits); err != nil {
					t.Fatal(arch, op, err)
				}
				limits.MaxCodeBytes--
				if _, err := compile(arch, f, limits); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("code budget", arch, op, err)
				}
			}
		}
	}
}

func floatResultMatches(op uint32, a, b, got, want uint64) bool {
	if got == want {
		return true
	}
	if op < 0x8d || op > 0xa5 || op == 0x98 || op == 0x99 || op == 0x9a {
		return false // Comparisons, signs and reinterpretation require exact bits.
	}
	exponent, mantissa, quiet := uint64(0x7f800000), uint64(0x007fffff), uint64(0x00400000)
	if op >= 0x99 {
		exponent, mantissa, quiet = 0x7ff0000000000000, 0x000fffffffffffff, 0x0008000000000000
	}
	if want&exponent != exponent || want&mantissa == 0 || got&exponent != exponent || got&quiet == 0 {
		return false
	}
	noncanonical := func(v uint64) bool { return v&exponent == exponent && v&mantissa != 0 && v&mantissa != quiet }
	binary := op >= 0x92 && op <= 0x97 || op >= 0xa0 && op <= 0xa5
	return noncanonical(a) || binary && noncanonical(b) || got&mantissa == quiet
}

func TestFloatNativeSemantics(t *testing.T) {
	if !native.Available {
		return
	}
	edges32 := []uint64{0, 1, 0x007fffff, 0x00800000, 0x7f7fffff, 0x7f800000, 0xff800000,
		0x7fc00000, 0x7fa12345, 0xffc12345, 0x80000000}
	edges64 := []uint64{0, 1, 0x000fffffffffffff, 0x0010000000000000, 0x7fefffffffffffff,
		0x7ff0000000000000, 0xfff0000000000000, 0x7ff8000000000000, 0x7ff123456789abcd,
		0xfff8123456789abc, 0x8000000000000000}
	for _, f := range []float64{-4.5, -3.5, -2.5, -1.5, -0.5, -0.25, 0.25, 0.5, 1.5, 2.5, 3.5, 4.5, 8388607.5, 8388608, 4503599627370495.5, 4503599627370496} {
		edges32 = append(edges32, uint64(math.Float32bits(float32(f))))
		edges64 = append(edges64, math.Float64bits(f))
	}
	for _, op := range floatOps() {
		edges := edges32
		if op >= 0x61 && op <= 0x66 || op >= 0x99 && op <= 0xa6 || op == 0xbd || op == 0xbf {
			edges = edges64
		}
		var pairs [][2]uint64
		for _, a := range edges {
			for _, b := range edges {
				pairs = append(pairs, [2]uint64{a, b})
			}
		}
		rng := rand.New(rand.NewSource(20261027))
		for i := 0; i < 64; i++ {
			pairs = append(pairs, [2]uint64{rng.Uint64(), rng.Uint64()})
		}
		for _, fast := range []bool{false, true} {
			for _, dst := range []slots.Slot{0, 1, 2} {
				compiled, lease, c := functionContext(t, floatFixture(op, dst), fast)
				for _, pair := range pairs {
					a, b := pair[0], pair[1]
					want, trap := numeric.Eval(op, a, b)
					out, exit, state := runFunction(t, compiled, lease, c, []value.Value{{Lo: a}, {Lo: b}}, 1, !fast, 1)
					if trap != wasm.NoTrap || exit != native.Return || len(out) != 1 || out[0].Hi != 0 || !floatResultMatches(op, a, b, out[0].Lo, want) || !fast && state.Fuel != 0 {
						t.Fatalf("op=%x fast=%v dst=%d a=%x b=%x: result=%x want=%x exit=%v fuel=%d", op, fast, dst, a, b, out, want, exit, state.Fuel)
					}
				}
				c.Close()
				lease.Close()
			}
		}
	}
}

func TestFloatValidatedFunctionsAndFuelBoundaries(t *testing.T) {
	for _, op := range floatOps() {
		input, output := wasm.F32, wasm.F32
		binary := op >= 0x5b && op <= 0x66 || op >= 0x92 && op <= 0x98 || op >= 0xa0 && op <= 0xa6
		if op >= 0x61 && op <= 0x66 || op >= 0x99 && op <= 0xa6 || op == 0xbd || op == 0xbf {
			input, output = wasm.F64, wasm.F64
		}
		if op <= 0x66 {
			output = wasm.I32
		} else if op == 0xbc {
			output = wasm.I32
		} else if op == 0xbd {
			output = wasm.I64
		} else if op == 0xbe {
			input = wasm.I32
		} else if op == 0xbf {
			input = wasm.I64
		}
		params := []wasm.ValueType{input}
		body := module.Expr{{Op: 0x20}}
		args := []value.Value{{Lo: 0xbfc00000}}
		if input == wasm.F64 || input == wasm.I64 {
			args[0].Lo = 0xbff8000000000000
		}
		b := uint64(0)
		if binary {
			params = append(params, input)
			body = append(body, module.Instr{Op: 0x20, A: 1})
			b = args[0].Lo
			args = append(args, value.Value{Lo: b})
		}
		body = append(body, module.Instr{Op: op}, module.Instr{Op: 0xb})
		f := functionModule(t, functionFixture(module.FuncType{Params: params, Results: []wasm.ValueType{output}}, 0, body)).Functions[0]
		if !native.Available {
			continue // Lowering/validation still runs in compiler-only builds.
		}
		for _, fast := range []bool{false, true} {
			compiled, lease, c := functionContext(t, f, fast)
			want, _ := numeric.Eval(op, args[0].Lo, b)
			out, exit, _ := runFunction(t, compiled, lease, c, args, uint64(len(args)+1), !fast, 1)
			if exit != native.Return || len(out) != 1 || !floatResultMatches(op, args[0].Lo, b, out[0].Lo, want) {
				t.Fatal(op, fast, out, exit, want)
			}
			if !fast {
				_, exit, state := runFunction(t, compiled, lease, c, args, uint64(len(args)), true, 1)
				if exit != native.FuelExhausted || state.Fuel != 0 || f.Sources[f.Code[state.Continuation].Source].Wasm.Op != op {
					t.Fatal("fuel must stop before floating-point instruction", op, exit, state)
				}
			}
			c.Close()
			lease.Close()
		}
	}
}
