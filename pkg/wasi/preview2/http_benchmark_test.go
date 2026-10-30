//go:build linux || darwin

package preview2

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

// Full requests use an external common server. Its accepted-connection count
// exposes the reference provider's per-request connect policy independently.
func BenchmarkHTTPRequests(b *testing.B) {
	authority := os.Getenv("WASMVM_HTTP_BENCH_AUTHORITY")
	if authority == "" {
		b.Skip("set WASMVM_HTTP_BENCH_AUTHORITY to the common external server")
	}
	benchmarkHTTP(b, "http", "request-batch", []component.Value{{Text: authority}, {Bits: 128}}, true)
}
func BenchmarkHTTPFields(b *testing.B) {
	benchmarkHTTP(b, "http-fields", "fields-batch", []component.Value{{Bits: 128}}, false)
}
func BenchmarkHTTPBodyTransfer(b *testing.B) {
	benchmarkHTTP(b, "http-fields", "body-batch", []component.Value{{Bits: 128}}, false)
}
func benchmarkHTTP(b *testing.B, fixture, export string, args []component.Value, networking bool) {
	start := time.Now()
	input, err := os.ReadFile("testdata/" + fixture + ".wasm")
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
			defer func() { process.Close(); process.WaitClosed(nil) }()
			network, err := NewNetwork(process, NetworkConfig{})
			if err != nil {
				b.Fatal(err)
			}
			defer func() { network.Close(); network.WaitClosed(nil) }()
			p, err := NewHTTP(network, HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				p.Close()
				if err := p.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			imports := p.Imports()
			clock, _ := NewWallClock(WallClockConfig{})
			random, _ := NewRandom(RandomConfig{})
			for _, extra := range []component.Imports{clock.Imports(), random.Imports()} {
				for name, value := range extra.Instances {
					imports.Instances[name] = value
				}
			}
			setup := time.Since(start)
			start = time.Now()
			in, err := module.Instantiate(nil, imports, component.CallOptions{})
			instantiate := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			start = time.Now()
			function, err := in.Function(export)
			lookup := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			batch := func() {
				result, err := function.Call(nil, args, component.CallOptions{})
				if err != nil || result.Value.Bits != 128 {
					b.Fatal(result, err)
				}
			}
			start = time.Now()
			batch()
			first := time.Since(start)
			for i := 0; i < 10; i++ {
				batch()
			}
			b.ReportAllocs()
			if networking {
				b.SetBytes(128 * 8192)
			} else if export == "body-batch" {
				b.SetBytes(128 * 4096)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				batch()
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/lifetime")
			start = time.Now()
			in.Close()
			p.Close()
			network.Close()
			process.Close()
			for _, owner := range []interface{ WaitClosed(context.Context) error }{p, network, process} {
				if err := owner.WaitClosed(nil); err != nil {
					b.Fatal(err)
				}
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
