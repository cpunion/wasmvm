package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestWideArithmeticTypesGateAndStackLimit(t *testing.T) {
	for op := uint32(0xfc0013); op <= 0xfc0016; op++ {
		arity := 2
		if op <= 0xfc0014 {
			arity = 4
		}
		for _, count := range []int{arity - 1, arity, arity + 1} {
			body := module.Expr{}
			for range count {
				body = append(body, module.Instr{Op: 0x42})
			}
			body = append(body, module.Instr{Op: op}, module.Instr{Op: 0x1a}, module.Instr{Op: 0x1a}, module.Instr{Op: 0xb})
			m := validFunction(body)
			_, err := CheckWithFeatures(m, DefaultLimits(), wasm.WideArithmetic)
			if count == arity && err != nil || count != arity && !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(op, count, err)
			}
			if _, err := Check(m, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal("enabled by default", op, err)
			}
			if count == arity {
				limits := DefaultLimits()
				limits.MaxStack = 1
				if _, err := CheckWithFeatures(m, limits, wasm.WideArithmetic); !errors.Is(err, wasm.ErrLimit) {
					t.Fatal("stack limit bypass", op, err)
				}
				body[0].Op = 0x41
				if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.WideArithmetic); !errors.Is(err, wasm.ErrInvalid) {
					t.Fatal("i32 operand", op, err)
				}
			}
		}
	}
}
