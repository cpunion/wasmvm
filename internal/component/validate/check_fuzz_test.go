package validate

import (
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
)

func FuzzComponentCheck(f *testing.F) {
	seed, err := os.ReadFile("testdata/canonical-math.wasm")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte("\x00asm\x0d\x00\x01\x00"))
	f.Fuzz(func(t *testing.T, input []byte) {
		definition, err := decode.Decode(input, decode.Limits{MaxBytes: 1 << 16, MaxNodes: 1 << 14, MaxNameBytes: 1024, MaxNesting: 16})
		if err != nil {
			return
		}
		result, err := Check(definition, Limits{MaxNodes: 1 << 14, MaxDepth: 32})
		if err != nil && result != nil {
			t.Fatal("failed validation published a component")
		}
	})
}
