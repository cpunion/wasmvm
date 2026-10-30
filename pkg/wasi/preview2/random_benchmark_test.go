package preview2

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

// Separate stream-cipher cost from provider allocation/leases and the complete
// component/canonical workload. This is a diagnostic, not a WASI throughput row.
func BenchmarkRandomGeneration(b *testing.B) {
	for _, stage := range []string{"cipher", "provider"} {
		b.Run(stage, func(b *testing.B) {
			r, err := NewRandom(RandomConfig{})
			if err != nil {
				b.Fatal(err)
			}
			data := make([]byte, 16384)
			ctx := context.Background()
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if stage == "cipher" {
					r.sources[0].stream.XORKeyStream(data, data)
				} else {
					data, err = r.read(ctx, 0, 16384)
					if err != nil || len(data) != 16384 {
						b.Fatal(err)
					}
				}
			}
			b.StopTimer()
			var checksum byte
			for _, v := range data {
				checksum ^= v
			}
			b.ReportMetric(float64(checksum), "checksum")
		})
	}
}

func BenchmarkRandomBytes(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/random.wasm")
	load := time.Since(start)
	if err != nil {
		b.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		b.Run(backend, func(b *testing.B) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			start := time.Now()
			e, err := component.NewEngine(config)
			engine := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer e.Close()
			start = time.Now()
			m, err := e.Compile(nil, input)
			compile := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			start = time.Now()
			r, err := NewRandom(RandomConfig{})
			entropy := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			start = time.Now()
			in, err := m.Instantiate(nil, r.Imports(), component.CallOptions{})
			instantiate := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			start = time.Now()
			f, err := in.Function("secure-count")
			lookup := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			args := []component.Value{{Bits: 16384}}
			start = time.Now()
			first, err := f.Call(nil, args, component.CallOptions{})
			firstTime := time.Since(start)
			if err != nil || first.Value.Bits != 16384 {
				b.Fatal(first, err)
			}
			b.SetBytes(16384)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := f.Call(nil, args, component.CallOptions{})
				if err != nil || v.Value.Bits != 16384 {
					b.Fatal(v, err)
				}
			}
			b.StopTimer()
			for name, value := range map[string]time.Duration{"load-ns": load, "engine-ns": engine, "compile-ns": compile, "entropy-init-ns": entropy, "store-bind-instantiate-ns": instantiate, "lookup-ns": lookup, "first-ns": firstTime} {
				b.ReportMetric(float64(value.Nanoseconds()), name)
			}
		})
	}
}
