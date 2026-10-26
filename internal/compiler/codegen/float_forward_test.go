package codegen

import (
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/ir"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestFloatForwardingPreservesEveryContinuation(t *testing.T) {
	for _, wide := range []bool{false, true} {
		for _, directResult := range []bool{false, true} {
			typ, constant, mul, sqrt := wasm.F32, uint32(0x43), uint32(0x94), uint32(0x91)
			bits := func(v float64) uint64 { return uint64(math.Float32bits(float32(v))) }
			if wide {
				typ, constant, mul, sqrt = wasm.F64, 0x44, 0xa2, 0x9f
				bits = math.Float64bits
			}
			f := &slots.Function{LocalCount: 2, Slots: []wasm.ValueType{typ, typ, typ, typ, typ}, Args: []slots.Slot{0, 1, 2, 3, 4}}
			add := func(op uint32, dst, a, b slots.Slot, local uint32, literal uint64) {
				f.Code = append(f.Code, slots.Instruction{Op: op, Source: uint32(len(f.Sources)), A: a, B: b, C: slots.NoSlot, Dst: dst})
				f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, A: local, Bits: literal}})
			}
			if !directResult {
				add(0x20, 2, slots.NoSlot, slots.NoSlot, 0, 0)
				add(constant, 3, slots.NoSlot, slots.NoSlot, 0, bits(1.25))
				add(mul, 4, 2, 3, 0, 0)
			} else {
				add(mul, 4, 0, 1, 0, 0)
			}
			add(0x21, slots.NoSlot, 4, slots.NoSlot, 1, 0)
			add(sqrt, 4, 1, slots.NoSlot, 0, 0)
			add(0x22, 2, 4, slots.NoSlot, 0, 0)
			add(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot, 0, 0)
			f.Code[len(f.Code)-1].Count = 5
			for _, arch := range []string{"amd64", "arm64"} {
				compiled, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
				if err != nil || compiled.ForwardedScalars == 0 {
					t.Fatal(arch, err, compiled)
				}
			}
			if !native.Available {
				continue
			}
			_, lease, c := functionContext(t, f, true)
			for start := range f.Code {
				want := make([]value.Value, 5)
				for i := range want {
					want[i].Lo = bits(float64(i+2) / 4)
					writeSlot(t, c, 16, slots.Slot(i), want[i])
				}
				for _, in := range f.Code[start : len(f.Code)-1] {
					source := f.Sources[in.Source].Wasm
					switch in.Op {
					case 0x20:
						want[in.Dst] = want[source.A]
					case 0x21, 0x22:
						want[source.A] = want[in.A]
						if in.Dst != slots.NoSlot {
							want[in.Dst] = want[in.A]
						}
					case 0x43, 0x44:
						want[in.Dst].Lo = source.Bits
					default:
						b := uint64(0)
						if in.B != slots.NoSlot {
							b = want[in.B].Lo
						}
						want[in.Dst].Lo, _ = numeric.Eval(in.Op, want[in.A].Lo, b)
					}
				}
				if err := c.SetState(native.State{Used: 16, Depth: 1, Continuation: uint64(start)}); err != nil {
					t.Fatal(err)
				}
				if exit, err := c.Enter(lease, 0); err != nil || exit != native.Return {
					t.Fatal(exit, err)
				}
				for i, v := range want {
					if got := readSlot(t, c, 16, slots.Slot(i)); got != v {
						t.Fatal("forwarded continuation", wide, directResult, start, i, got, v)
					}
				}
			}
			c.Close()
			lease.Close()
		}
	}
}

func TestFloatCopyProofBoundaries(t *testing.T) {
	for _, kind := range []string{"ordinary", "source-overwrite", "copy-overwrite", "helper", "poll", "scan-limit"} {
		f := &slots.Function{LocalCount: 1, Slots: []wasm.ValueType{wasm.F64, wasm.F64, wasm.F64, wasm.F64}, Args: []slots.Slot{3}}
		appendOp := func(op uint32, dst, a, b slots.Slot) {
			f.Code = append(f.Code, slots.Instruction{Op: op, Source: uint32(len(f.Sources)), A: a, B: b, C: slots.NoSlot, Dst: dst})
			f.Sources = append(f.Sources, ir.Source{Wasm: module.Instr{Op: op, Bits: 0x3ff0000000000000}})
		}
		if kind == "poll" {
			for i := 0; i < 255; i++ {
				appendOp(1, slots.NoSlot, slots.NoSlot, slots.NoSlot)
			}
		}
		appendOp(0x20, 1, slots.NoSlot, slots.NoSlot)
		dst := slots.Slot(2)
		if kind == "source-overwrite" {
			dst = 0
		}
		if kind == "copy-overwrite" {
			dst = 1
		}
		appendOp(0x44, dst, slots.NoSlot, slots.NoSlot)
		if kind == "helper" {
			appendOp(0x3f, slots.NoSlot, slots.NoSlot, slots.NoSlot)
		}
		if kind == "scan-limit" {
			for i := 0; i < 7; i++ {
				appendOp(0x44, 2, slots.NoSlot, slots.NoSlot)
			}
		}
		appendOp(0xa2, 3, 1, 2)
		appendOp(slots.ControlBase+uint32(ir.Return), slots.NoSlot, slots.NoSlot, slots.NoSlot)
		f.Code[len(f.Code)-1].Count = 1
		for _, arch := range []string{"amd64", "arm64"} {
			compiled, err := CompileUnmeteredFunction(arch, f, DefaultLimits())
			if err != nil || (compiled.ForwardedScalars != 0) != (kind == "ordinary") {
				t.Fatal(arch, kind, compiled, err)
			}
			precise, err := CompileFunction(arch, f, DefaultLimits())
			if err != nil || precise.ForwardedScalars != 0 {
				t.Fatal("precise fuel must not combine", kind, err)
			}
		}
	}
}
