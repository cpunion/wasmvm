package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExtendedConstantTyping(t *testing.T) {
	for _, op := range []uint32{0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e} {
		typ, literal := wasm.I32, uint32(0x41)
		if op >= 0x7c {
			typ, literal = wasm.I64, 0x42
		}
		expr := module.Expr{{Op: literal}, {Op: literal}, {Op: op, Offset: 42}, {Op: 0xb}}
		m := &module.Module{Globals: []module.Global{{Type: module.GlobalType{Value: typ}, Init: expr}}}
		if _, err := Check(m, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatalf("%x enabled by default: %v", op, err)
		}
		v, err := CheckWithFeatures(m, DefaultLimits(), wasm.ExtendedConst)
		if err != nil || v.ConstantStack() != 2 {
			t.Fatalf("%x: %v, %v", op, v, err)
		}
		for _, bad := range []module.Expr{
			{{Op: op}, {Op: 0xb}},
			{{Op: literal}, {Op: op}, {Op: 0xb}},
			{{Op: 0x43}, {Op: literal}, {Op: op}, {Op: 0xb}},
			{{Op: literal}, {Op: 0x43}, {Op: op}, {Op: 0xb}},
			{{Op: literal}, {Op: literal}, {Op: 0xb}},
			{{Op: literal}, {Op: 0x1a}, {Op: literal}, {Op: 0xb}},
			{{Op: literal}, {Op: literal}, {Op: op}},
			{{Op: 0xb}},
		} {
			m.Globals[0].Init = bad
			if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.ExtendedConst); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatalf("%x accepted %v: %v", op, bad, err)
			}
		}
	}
}

func TestExtendedConstantLimitsAndGlobals(t *testing.T) {
	var expr module.Expr
	for range 64 {
		expr = append(expr, module.Instr{Op: 0x41})
	}
	for range 63 {
		expr = append(expr, module.Instr{Op: 0x6a})
	}
	expr = append(expr, module.Instr{Op: 0xb})
	m := &module.Module{Globals: []module.Global{{Type: module.GlobalType{Value: wasm.I32}, Init: expr}}}
	limits := DefaultLimits()
	limits.MaxStack = 63
	if _, err := CheckWithFeatures(m, limits, wasm.ExtendedConst); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	limits.MaxStack = 64
	v, err := CheckWithFeatures(m, limits, wasm.ExtendedConst)
	if err != nil || v.ConstantStack() != 64 {
		t.Fatal(v, err)
	}
	if _, err := CheckWithFeatures(m, limits, wasm.Features(1<<63)); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
	// Extended const still cannot read a mutable or module-defined global.
	m.Globals[0].Init = module.Expr{{Op: 0x23}, {Op: 0x41}, {Op: 0x6a}, {Op: 0xb}}
	for _, imports := range [][]module.Import{nil, {{Kind: 3, Global: module.GlobalType{Value: wasm.I32, Mutable: true}}}} {
		m.Imports = imports
		if _, err := CheckWithFeatures(m, limits, wasm.ExtendedConst); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
	}
	limits.MaxStack = 0
	m.Globals[0].Init = module.Expr{{Op: 0x41}, {Op: 0xb}}
	if _, err := CheckWithFeatures(m, limits, wasm.ExtendedConst); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
