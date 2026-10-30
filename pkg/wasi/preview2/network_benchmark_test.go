package preview2

import (
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func BenchmarkNetworkLiteralBatch(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/network-common.wasm")
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
			engine, err := component.NewEngine(config)
			engineTime := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer engine.Close()
			start = time.Now()
			module, err := engine.Compile(nil, input)
			compile := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer module.Close()
			start = time.Now()
			process, err := NewIO(IOConfig{})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				process.Close()
				if err := process.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			network, err := NewNetwork(process, NetworkConfig{AllowNameLookup: true})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				network.Close()
				if err := network.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			imports := network.Imports()
			setup := time.Since(start)
			start = time.Now()
			in, err := module.Instantiate(nil, imports, component.CallOptions{})
			instantiate := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			start = time.Now()
			function, err := in.Function("lookup-batch")
			lookup := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			args := []component.Value{{Bits: 128}}
			start = time.Now()
			result, err := function.Call(nil, args, component.CallOptions{})
			first := time.Since(start)
			if err != nil || result.Value.Bits != 128 {
				b.Fatal(result, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := function.Call(nil, args, component.CallOptions{})
				if err != nil || result.Value.Bits != 128 {
					b.Fatal(result, err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/lifetime")
			start = time.Now()
			in.Close()
			network.Close()
			process.Close()
			if err := network.WaitClosed(nil); err != nil {
				b.Fatal(err)
			}
			if err := process.WaitClosed(nil); err != nil {
				b.Fatal(err)
			}
			module.Close()
			engine.Close()
			closeTime := time.Since(start)
			for name, value := range map[string]time.Duration{"load-ns": load, "engine-ns": engineTime, "compile-ns": compile, "provider-init-ns": setup, "store-bind-instantiate-ns": instantiate, "lookup-ns": lookup, "first-ns": first, "close-ns": closeTime} {
				b.ReportMetric(float64(value.Nanoseconds()), name)
			}
		})
	}
}
func BenchmarkNetworkLiteralStages(b *testing.B) {
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			address, err := netip.ParseAddr("127.0.0.1")
			if err != nil || !address.Is4() {
				b.Fatal(address, err)
			}
		}
	})
	address := netip.MustParseAddr("127.0.0.1")
	b.Run("value", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			value := ipVariantValue(address)
			if value.Case != 0 || value.Elements[0].Elements[0].Bits != 127 {
				b.Fatal(value)
			}
		}
	})
}
