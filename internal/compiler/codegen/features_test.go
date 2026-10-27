package codegen

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExplicitX86FeatureInputs(t *testing.T) {
	f, _ := vectorFixture(14, 0, 3, 0)
	for _, arch := range []string{"amd64", "arm64"} {
		for _, features := range []X86Features{0, 1, 2, 3, 4, 5, 6, 7, 8, 16, 24, 31, 32, 0xffffffff} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{X86Features: features})
			unsupported := features&^31 != 0 || arch != "amd64" && features != 0
			if unsupported {
				if c != nil || !errors.Is(err, wasm.ErrUnsupported) {
					t.Fatal(arch, features, c, err)
				}
			} else if err != nil || c.Helpers != 0 {
				t.Fatal(arch, features, c, err)
			}
		}
	}
}
