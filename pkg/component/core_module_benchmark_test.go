package component

import (
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
)

func BenchmarkHostCoreModule(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/host-module.wasm")
	if err != nil {
		b.Fatal(err)
	}
	core, err := os.ReadFile("testdata/host-core-math.wasm")
	if err != nil {
		b.Fatal(err)
	}
	load := time.Since(start)
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		b.Run(backend, func(b *testing.B) {
			config := DefaultConfig()
			config.Core.Backend = backend
			start := time.Now()
			e, err := NewEngine(config)
			engine := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer e.Close()
			start = time.Now()
			host, err := e.CompileCore(nil, core)
			coreCompile := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer host.Close()
			start = time.Now()
			m, err := e.Compile(nil, input)
			compile := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer m.Close()
			start = time.Now()
			in, err := m.Instantiate(nil, Imports{CoreModules: map[string]*CoreModule{"math": host}, Functions: map[string]HostFunction{"mark": func(*Caller, []Value) (*Value, error) { return nil, nil }}}, CallOptions{})
			instantiate := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			start = time.Now()
			f, err := in.Function("add")
			lookup := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			args := []Value{{Bits: 37}, {Bits: 5}}
			start = time.Now()
			first, err := f.Call(nil, args, CallOptions{})
			firstTime := time.Since(start)
			if err != nil || first.Value.Bits != 42 {
				b.Fatal(first, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				args[0].Bits = uint64(uint32(i))
				v, err := f.Call(nil, args, CallOptions{})
				if err != nil || v.Value.Bits != uint64(uint32(i)+5) {
					b.Fatal(v, err)
				}
			}
			b.StopTimer()
			for name, value := range map[string]time.Duration{"load-ns": load, "engine-ns": engine, "core-compile-ns": coreCompile, "component-compile-ns": compile, "store-bind-instantiate-ns": instantiate, "lookup-ns": lookup, "first-ns": firstTime} {
				b.ReportMetric(float64(value.Nanoseconds()), name)
			}
		})
	}
}
