package preview2

import (
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func BenchmarkPollClockSubscriptions(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/poll-clock.wasm")
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
			poller, err := NewPoller(PollConfig{})
			if err != nil {
				b.Fatal(err)
			}
			defer poller.Close()
			r, err := NewMonotonicClock(poller)
			provider := time.Since(start)
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
			f, err := in.Function("churn")
			lookup := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			args := []component.Value{{Bits: 128}}
			start = time.Now()
			first, err := f.Call(nil, args, component.CallOptions{})
			firstTime := time.Since(start)
			if err != nil || first.Value.Bits != 128 {
				b.Fatal(first, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := f.Call(nil, args, component.CallOptions{})
				if err != nil || v.Value.Bits != 128 {
					b.Fatal(v, err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/subscription")
			poller.mu.Lock()
			live := len(poller.entries)
			poller.mu.Unlock()
			if live != 0 {
				b.Fatal("subscriptions retained", live)
			}
			for name, value := range map[string]time.Duration{"load-ns": load, "engine-ns": engine, "compile-ns": compile, "provider-init-ns": provider, "store-bind-instantiate-ns": instantiate, "lookup-ns": lookup, "first-ns": firstTime} {
				b.ReportMetric(float64(value.Nanoseconds()), name)
			}
		})
	}
}
