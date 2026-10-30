package component

import (
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func BenchmarkComponentRuntime(b *testing.B) {
	input, err := os.ReadFile("../../internal/component/decode/testdata/composition.wasm")
	if err != nil {
		b.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		b.Run(backend, func(b *testing.B) {
			config := DefaultConfig()
			config.Core.Backend = backend
			e, err := NewEngine(config)
			if err != nil {
				b.Fatal(err)
			}
			defer e.Close()
			b.Run("Compile", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					m, err := e.Compile(nil, input)
					if err != nil {
						b.Fatal(err)
					}
					m.Close()
				}
			})
			m, err := e.Compile(nil, input)
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			b.Run("CreateStoreInstantiateAndClose", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
					if err != nil {
						b.Fatal(err)
					}
					in.Close()
				}
			})
			for _, session := range []bool{false, true} {
				name := "PublicWarm"
				if session {
					name = "SharedCoreSessionWarm"
				}
				b.Run(name, func(b *testing.B) {
					in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
					if err != nil {
						b.Fatal(err)
					}
					defer in.Close()
					fn, err := in.Function("add")
					if err != nil {
						b.Fatal(err)
					}
					args := []Value{{}, {Bits: 5}}
					start := time.Now()
					if _, err := fn.Call(nil, args, vm.CallOptions{}); err != nil {
						b.Fatal(err)
					}
					first := time.Since(start)
					b.ReportAllocs()
					b.ResetTimer()
					if session {
						_, err = in.store.WithCaller(nil, vm.CallOptions{}, func(caller *vm.Caller) error {
							for i := 0; i < b.N; i++ {
								args[0].Bits = uint64(uint32(i))
								value, err := fn.call(caller, args)
								if err != nil || value.Bits != uint64(uint32(i)+5) {
									b.Fatal(value, err)
								}
							}
							return nil
						})
					} else {
						for i := 0; i < b.N; i++ {
							args[0].Bits = uint64(uint32(i))
							value, err := fn.Call(nil, args, vm.CallOptions{})
							if err != nil || value.Value.Bits != uint64(uint32(i)+5) {
								b.Fatal(value, err)
							}
						}
					}
					if err != nil {
						b.Fatal(err)
					}
					b.StopTimer()
					b.ReportMetric(float64(first.Nanoseconds()), "first-ns")
				})
			}
		})
	}
}
