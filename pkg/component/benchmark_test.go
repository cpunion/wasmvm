package component

import (
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
)

func BenchmarkComponentRuntime(b *testing.B) {
	path := os.Getenv("WASMVM_COMPONENT_BENCH_FIXTURE")
	if path == "" {
		path = "../../internal/component/decode/testdata/composition.wasm"
	}
	input, err := os.ReadFile(path)
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

// The batch includes the guest loop and new/rep/drop/destructor execution. It
// amortizes the public call boundary; compilation/startup stay outside timing.
func BenchmarkResourceInstructions(b *testing.B) {
	input, err := os.ReadFile("testdata/resources.wasm")
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
			m, err := e.Compile(nil, input)
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{}, vm.CallOptions{})
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			f, err := in.Function("churn")
			if err != nil {
				b.Fatal(err)
			}
			args := []Value{{Bits: 128}}
			if _, err := f.Call(nil, args, vm.CallOptions{}); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := f.Call(nil, args, vm.CallOptions{})
				if err != nil || v.Value.Bits != 128 {
					b.Fatal(v, err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/resource")
		})
	}
}

// Each lifetime crosses the public boundary four times: make/get/pass/consume.
// Startup, export lookup and the first lifetime are outside the timed region.
func BenchmarkOwnedResourceLifetime(b *testing.B) {
	input, err := os.ReadFile("testdata/ownership.wasm")
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
			m, err := e.Compile(nil, input)
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			in, err := m.Instantiate(nil, Imports{}, CallOptions{})
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			makeResource, _ := in.Function("make")
			get, _ := in.Function("get")
			pass, _ := in.Function("pass")
			consume, _ := in.Function("consume")
			total, _ := in.Function("total")
			n := []Value{{Bits: 1}}
			lifetime := func() {
				v, err := makeResource.Call(nil, n, CallOptions{})
				if err != nil {
					b.Fatal(err)
				}
				value, err := get.Call(nil, []Value{*v.Value}, CallOptions{})
				if err != nil || value.Value.Bits != 1 {
					b.Fatal(value, err)
				}
				moved, err := pass.Call(nil, []Value{*v.Value}, CallOptions{})
				if err != nil {
					b.Fatal(err)
				}
				value, err = consume.Call(nil, []Value{*moved.Value}, CallOptions{})
				if err != nil || value.Value.Bits != 1 {
					b.Fatal(value, err)
				}
			}
			lifetime()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lifetime()
			}
			b.StopTimer()
			value, err := total.Call(nil, nil, CallOptions{})
			if err != nil || value.Value.Bits != uint64(uint32(b.N+1)) {
				b.Fatal(value, err)
			}
		})
	}
}
