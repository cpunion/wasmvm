package codegen

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestExplicitX86FeatureInputs(t *testing.T) {
	f, _ := vectorFixture(14, 0, 3, 0)
	for _, arch := range []string{"amd64", "arm64"} {
		for _, features := range []X86Features{0, X86SSSE3, X86SSE41, X86SSSE3 | X86SSE41, 4, 0xffffffff} {
			c, err := CompileFunctionWithConfig(arch, f, DefaultLimits(), Config{X86Features: features})
			unsupported := features&^3 != 0 || arch != "amd64" && features != 0
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
