package validate

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestRelaxedSIMDSignaturesAndFeatureGate(t *testing.T) {
	for op := uint32(0x100); op <= 0x113; op++ {
		arity := 2
		if op >= 0x101 && op <= 0x104 {
			arity = 1
		} else if op >= 0x105 && op <= 0x10c || op == 0x113 {
			arity = 3
		}
		for _, count := range []int{arity - 1, arity, arity + 1} {
			body := module.Expr{}
			for range count {
				body = append(body, module.Instr{Op: 0xfd000c})
			}
			body = append(body, module.Instr{Op: 0xfd0000 | op}, module.Instr{Op: 0x1a}, module.Instr{Op: 0xb})
			m := validFunction(body)
			_, err := CheckWithFeatures(m, DefaultLimits(), wasm.RelaxedSIMD)
			if count == arity && err != nil || count != arity && !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal(op, count, err)
			}
			if _, err := Check(m, DefaultLimits()); !errors.Is(err, wasm.ErrInvalid) {
				t.Fatal("validator enabled extension", op, err)
			}
			if count == arity {
				body[0].Op = 0x41
				if _, err := CheckWithFeatures(m, DefaultLimits(), wasm.RelaxedSIMD); !errors.Is(err, wasm.ErrInvalid) {
					t.Fatal("scalar operand accepted", op, err)
				}
			}
		}
		polymorphic := validFunction(module.Expr{{Op: 0}, {Op: 0xfd0000 | op}, {Op: 0x1a}, {Op: 0xb}})
		if _, err := CheckWithFeatures(polymorphic, DefaultLimits(), wasm.RelaxedSIMD); err != nil {
			t.Fatal("unreachable operand stack", op, err)
		}
	}
}
