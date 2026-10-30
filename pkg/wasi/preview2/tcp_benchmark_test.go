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

// BenchmarkTCPRoundtripBatch excludes connection setup and teardown from the
// timed 128-roundtrip batch. Each roundtrip transfers 4 KiB in each direction.
func BenchmarkTCPRoundtripBatch(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/tcp-common.wasm")
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
			network, err := NewNetwork(process, NetworkConfig{})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				network.Close()
				if err := network.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			provider, err := NewTCP(network, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				provider.Close()
				if err := provider.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			imports := provider.Imports()
			setup := time.Since(start)
			start = time.Now()
			in, err := module.Instantiate(nil, imports, component.CallOptions{})
			instantiate := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			call := func(name string, args []component.Value, want uint64) {
				f, err := in.Function(name)
				if err != nil {
					b.Fatal(err)
				}
				result, err := f.Call(context.Background(), args, component.CallOptions{})
				if err != nil || result.Value.Bits != want {
					b.Fatal(name, result, err)
				}
			}
			for _, family := range []uint64{0, 1} {
				call("exercise", []component.Value{{Bits: family}}, 42)
			}
			start = time.Now()
			call("setup", []component.Value{{}}, 1)
			sessionSetup := time.Since(start)
			start = time.Now()
			transfer, err := in.Function("transfer")
			lookup := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			args := []component.Value{{Bits: 128}}
			batch := func() {
				result, err := transfer.Call(nil, args, component.CallOptions{})
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
			b.SetBytes(128 * 8192)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				batch()
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/roundtrip")
			start = time.Now()
			call("teardown", nil, 1)
			sessionClose := time.Since(start)
			start = time.Now()
			in.Close()
			provider.Close()
			network.Close()
			process.Close()
			for _, owner := range []interface{ WaitClosed(context.Context) error }{provider, network, process} {
				if err := owner.WaitClosed(nil); err != nil {
					b.Fatal(err)
				}
			}
			module.Close()
			engine.Close()
			closeTime := time.Since(start)
			for name, value := range map[string]time.Duration{"load-ns": load, "engine-ns": engineTime, "compile-ns": compile, "provider-init-ns": setup, "store-bind-instantiate-ns": instantiate, "lookup-ns": lookup, "session-setup-ns": sessionSetup, "first-ns": first, "session-close-ns": sessionClose, "close-ns": closeTime} {
				b.ReportMetric(float64(value.Nanoseconds()), name)
			}
		})
	}
}
