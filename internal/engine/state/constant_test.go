package state

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestConstantIntegerWrapAndOperandOrder(t *testing.T) {
	for _, tc := range []struct {
		op         uint32
		a, b, want uint64
	}{
		{0x6a, 0xffffffff, 1, 0}, {0x6b, 0, 1, 0xffffffff},
		{0x6c, 0x80000001, 2, 2}, {0x7c, ^uint64(0), 1, 0},
		{0x7d, 0, 1, ^uint64(0)}, {0x7e, 1<<63 | 1, 2, 2},
	} {
		literal := uint32(0x41)
		if tc.op >= 0x7c {
			literal = 0x42
		}
		expr := module.Expr{{Op: literal, Bits: tc.a}, {Op: literal, Bits: tc.b}, {Op: tc.op}, {Op: 0xb}}
		if got := constant(nil, expr, make([]value.Value, 2)); got.Lo != tc.want || got.Hi != 0 {
			t.Fatalf("%x: %v want %x", tc.op, got, tc.want)
		}
	}
}

func TestConstantDeepStackAndAllocationCleanup(t *testing.T) {
	var expr module.Expr
	for range 64 {
		expr = append(expr, module.Instr{Op: 0x41, Bits: 1})
	}
	for range 63 {
		expr = append(expr, module.Instr{Op: 0x6a})
	}
	expr = append(expr, module.Instr{Op: 0xb})
	d := &module.Module{Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: expr}}}
	v, err := validate.CheckWithFeatures(d, validate.DefaultLimits(), wasm.ExtendedConst)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Compile(v, lower.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	code, err := slots.Compile(ir, slots.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	s := New(DefaultLimits())
	defer s.Close()
	in, err := s.Allocate(code, nil)
	if err != nil || in.Globals[0].Value.Lo != 64 {
		t.Fatal(in, err)
	}
	if err := in.Initialize(); err != nil {
		t.Fatal(err)
	}
	s.CollectFailed(nil, nil)
	if got := s.Usage(); got != (Usage{}) {
		t.Fatal("failed allocation retained state", got)
	}
}
