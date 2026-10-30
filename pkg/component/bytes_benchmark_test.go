package component

import (
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
)

// The source baseline can execute the unchanged element-wise row; its packed
// row fails the payload checks. Compilation/binding/lookup/first call are excluded.
func BenchmarkByteListTransfer(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/bytes.wasm")
	loadTime := time.Since(start)
	if err != nil {
		b.Fatal(err)
	}
	const size = 16384
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"elements", "packed"} {
			b.Run(backend+"/"+mode, func(b *testing.B) {
				config := DefaultConfig()
				config.Core.Backend = backend
				start := time.Now()
				e, err := NewEngine(config)
				engineTime := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				defer e.Close()
				start = time.Now()
				m, err := e.Compile(nil, input)
				compileTime := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				defer m.Close()
				start = time.Now()
				in, err := m.Instantiate(nil, Imports{}, CallOptions{})
				instantiateTime := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				defer in.Close()
				start = time.Now()
				f, err := in.Function("echo")
				lookupTime := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				data := make([]byte, size)
				for i := range data {
					data[i] = byte(i)
				}
				v := Value{Text: string(data)}
				if mode == "elements" {
					v = Value{Elements: make([]Value, size)}
					for i := range v.Elements {
						v.Elements[i].Bits = uint64(data[i])
					}
				}
				args := []Value{v}
				verify := func(value *Value) {
					if value == nil {
						b.Fatal("missing byte result")
					}
					if value.Elements == nil {
						if len(value.Text) != size || value.Text[0] != 0 || value.Text[size-1] != 255 {
							b.Fatal("packed byte checksum")
						}
					} else if len(value.Elements) != size || value.Elements[0].Bits != 0 || value.Elements[size-1].Bits != 255 {
						b.Fatal("element byte checksum")
					}
				}
				start = time.Now()
				first, err := f.Call(nil, args, CallOptions{})
				firstTime := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				verify(first.Value)
				for i, expected := range data {
					if first.Value.Elements == nil {
						if first.Value.Text[i] != expected {
							b.Fatal(i)
						}
					} else if first.Value.Elements[i].Bits != uint64(expected) {
						b.Fatal(i)
					}
				}
				b.SetBytes(size)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := f.Call(nil, args, CallOptions{})
					if err != nil {
						b.Fatal(err)
					}
					verify(result.Value)
				}
				b.StopTimer()
				b.ReportMetric(float64(loadTime.Nanoseconds()), "load-ns")
				b.ReportMetric(float64(engineTime.Nanoseconds()), "engine-ns")
				b.ReportMetric(float64(compileTime.Nanoseconds()), "compile-ns")
				b.ReportMetric(float64(instantiateTime.Nanoseconds()), "store-bind-instantiate-ns")
				b.ReportMetric(float64(lookupTime.Nanoseconds()), "lookup-ns")
				b.ReportMetric(float64(firstTime.Nanoseconds()), "first-ns")
			})
		}
	}
}
