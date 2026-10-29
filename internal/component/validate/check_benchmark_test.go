package validate

import (
	"os"
	"testing"

	"github.com/cpunion/wasmvm/internal/component/decode"
)

func BenchmarkCanonicalModuleValidation(b *testing.B) {
	benchmarkComponentValidation(b, "testdata/canonical-math.wasm")
}

func BenchmarkComponentCompositionValidation(b *testing.B) {
	benchmarkComponentValidation(b, "../decode/testdata/composition.wasm")
}

func BenchmarkComponentValueStartValidation(b *testing.B) {
	benchmarkComponentValidation(b, "testdata/value-start.wasm")
}

func benchmarkComponentValidation(b *testing.B, path string) {
	input, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	definition, err := decode.Decode(input, decode.DefaultLimits())
	if err != nil {
		b.Fatal(err)
	}
	for _, decoding := range []bool{false, true} {
		name := "ValidatedSyntax"
		if decoding {
			name = "OwnedDecodeAndValidation"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m := definition
				var err error
				if decoding {
					m, err = decode.Decode(input, decode.DefaultLimits())
					if err != nil {
						b.Fatal(err)
					}
				}
				if _, err := Check(m, DefaultLimits()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
