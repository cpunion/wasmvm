package codegen

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestLeafConstantFoldingMatchesIntegerSemantics(t *testing.T) {
	values := []uint64{0, 1, 2, 0x7fffffff, 0x80000000, 0xffffffff, 1 << 32, 1 << 63, ^uint64(0)}
	for _, op := range []uint32{0x46, 0x47, 0x49, 0x4b, 0x4d, 0x4f, 0x51, 0x52, 0x54, 0x56, 0x58, 0x5a, 0x6a, 0x6b, 0x6c, 0x7c, 0x7d, 0x7e} {
		literal := uint32(0x41)
		if integerWide(op) {
			literal = 0x42
		}
		for _, a := range values {
			for _, b := range values {
				if literal == 0x41 {
					a, b = uint64(uint32(a)), uint64(uint32(b))
				}
				left, right := &leafExpression{op: literal, bits: a}, &leafExpression{op: literal, bits: b}
				n := &leafExpression{op: op, left: left, right: right}
				folded := foldLeafExpression(n)
				want, trap := numeric.Eval(op, a, b)
				if trap != wasm.NoTrap || folded.bits != want || folded.left != nil || folded.right != nil || n.left != left || n.right != right {
					t.Fatal(op, a, b, folded, want, trap)
				}
			}
		}
	}
}

func TestLeafAffineCompositionKeepsWrappingAndSharedInputs(t *testing.T) {
	for _, wide := range []bool{false, true} {
		literal, add := uint32(0x41), uint32(0x6a)
		if wide {
			literal, add = 0x42, 0x7c
		}
		c := func(n uint64) *leafExpression {
			if !wide {
				n = uint64(uint32(n))
			}
			return &leafExpression{op: literal, bits: n}
		}
		var evaluate func(*leafExpression, uint64) uint64
		evaluate = func(n *leafExpression, x uint64) uint64 {
			if n.op == 0 {
				return x
			}
			if n.op == literal {
				return n.bits
			}
			v, trap := numeric.Eval(n.op, evaluate(n.left, x), evaluate(n.right, x))
			if trap != wasm.NoTrap {
				t.Fatal(trap)
			}
			return v
		}
		input := &leafExpression{}
		for _, factor := range []uint64{0, 1, 3, 1664525, ^uint64(0)} {
			for _, bias := range []uint64{0, 1, 1013904223, ^uint64(0)} {
				expression := input
				for count := 1; count <= 16; count++ {
					previous := expression
					product := foldLeafExpression(&leafExpression{op: add + 2, left: c(factor), right: expression})
					expression = foldLeafExpression(&leafExpression{op: add, left: c(bias), right: product})
					for _, seed := range []uint64{0, 1, 1 << 31, 1 << 63, ^uint64(0)} {
						if !wide {
							seed = uint64(uint32(seed))
						}
						want := seed
						for i := 0; i < count; i++ {
							if i == count-1 && evaluate(previous, seed) != want {
								t.Fatal("composition mutated shared input", count)
							}
							want = want*factor + bias
							if !wide {
								want = uint64(uint32(want))
							}
						}
						if got := evaluate(expression, seed); got != want {
							t.Fatal(wide, factor, bias, count, seed, got, want)
						}
					}
				}
			}
		}
	}
}
