package decode

import (
	"os"
	"testing"
)

// BenchmarkDecodeComposition measures owned binary decoding, including the
// embedded Core module grammar. It does not include semantic validation,
// native compilation, instantiation or execution.
func BenchmarkDecodeComposition(b *testing.B) {
	input, err := os.ReadFile("testdata/composition.wasm")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, err := Decode(input, DefaultLimits())
		if err != nil || len(m.Definitions) == 0 {
			b.Fatal(err)
		}
	}
}
