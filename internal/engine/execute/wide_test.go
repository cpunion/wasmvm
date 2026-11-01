package execute

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// big.Int supplies an independent signed, arbitrary-precision oracle. Results
// wrap modulo 2^128, including negative products and overflowing carry chains.
func TestWideArithmeticOracle(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/wide-arithmetic/wide-arithmetic/script.0.wasm")
	if err != nil {
		t.Fatal(err)
	}
	tiers := []int{0}
	if native.Available {
		tiers = append(tiers, 1, 2)
	}
	for _, tier := range tiers {
		t.Run(fmt.Sprint(tier), func(t *testing.T) {
			r := newSpec(t, context.Background(), tier)
			r.features = wasm.WideArithmetic
			in, err := r.instantiate(data)
			if err != nil {
				t.Fatal(err)
			}
			rng := rand.New(rand.NewSource(128))
			modulus := new(big.Int).Lsh(big.NewInt(1), 128)
			for iteration := 0; iteration < 128; iteration++ {
				args := []value.Value{{Lo: rng.Uint64()}, {Lo: rng.Uint64()}, {Lo: rng.Uint64()}, {Lo: rng.Uint64()}}
				for _, op := range []string{"i64.add128", "i64.sub128", "i64.mul_wide_s", "i64.mul_wide_u"} {
					left, right, want := new(big.Int), new(big.Int), new(big.Int)
					count := 4
					switch op {
					case "i64.add128", "i64.sub128":
						left.Lsh(new(big.Int).SetUint64(args[1].Lo), 64).Add(left, new(big.Int).SetUint64(args[0].Lo))
						right.Lsh(new(big.Int).SetUint64(args[3].Lo), 64).Add(right, new(big.Int).SetUint64(args[2].Lo))
						if op == "i64.add128" {
							want.Add(left, right)
						} else {
							want.Sub(left, right)
						}
					case "i64.mul_wide_s":
						count = 2
						want.Mul(left.SetInt64(int64(args[0].Lo)), right.SetInt64(int64(args[1].Lo)))
					case "i64.mul_wide_u":
						count = 2
						want.Mul(left.SetUint64(args[0].Lo), right.SetUint64(args[1].Lo))
					}
					want.Mod(want, modulus)
					low := want.Uint64()
					high := want.Rsh(want, 64).Uint64()
					got, err := r.call(in.Exports[op].Function, args[:count])
					if err != nil || len(got) != 2 || got[0] != (value.Value{Lo: low}) || got[1] != (value.Value{Lo: high}) {
						t.Fatalf("%s case %d args=%v: got %v err=%v want [%x %x]", op, iteration, args, got, err, low, high)
					}
				}
			}
		})
	}
}
