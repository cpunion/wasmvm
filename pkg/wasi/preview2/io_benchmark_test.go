package preview2

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func BenchmarkIOWriteBatch(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/io-streams.wasm")
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
			sink := &benchmarkOutput{}
			r, err := NewIO(IOConfig{Stdout: sink})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				r.Close()
				if err := r.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()

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
			f, err := in.Function("write-batch")
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
			sink.mu.Lock()
			sink.bytes = 0
			sink.mu.Unlock()
			b.SetBytes(128 * 4096)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := f.Call(nil, args, component.CallOptions{})
				if err != nil || v.Value.Bits != 128 {
					b.Fatal(v, err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/write")
			sink.mu.Lock()
			written := sink.bytes
			sink.mu.Unlock()
			if written != uint64(b.N)*128*4096 {
				b.Fatal("written bytes", written)
			}

			for name, value := range map[string]time.Duration{"load-ns": load, "engine-ns": engine, "compile-ns": compile, "provider-init-ns": provider, "store-bind-instantiate-ns": instantiate, "lookup-ns": lookup, "first-ns": firstTime} {
				b.ReportMetric(float64(value.Nanoseconds()), name)
			}
		})
	}
}

type benchmarkOutput struct {
	mu    sync.Mutex
	bytes uint64
}

func (s *benchmarkOutput) Readiness() (bool, <-chan struct{}) { return true, nil }
func (s *benchmarkOutput) Write(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	const pattern = "0123456789abcdef"
	if len(data) > 0 && (data[0] != pattern[s.bytes%16] || data[len(data)-1] != pattern[(s.bytes+uint64(len(data))-1)%16]) {
		return 0, errors.New("output contents mismatch")
	}
	s.bytes += uint64(len(data))
	return len(data), nil
}
func (s *benchmarkOutput) Flush(ctx context.Context) error { return ctx.Err() }
func (s *benchmarkOutput) Close() error                    { return nil }

// Separate raw source, synchronous backend and queued worker costs from the
// complete guest/canonical workload. These diagnostic scopes are not additive.
func BenchmarkIOOutput(b *testing.B) {
	const payload = "0123456789abcdef"
	text := strings.Repeat(payload, 256)
	bytes := []byte(text)
	for _, stage := range []string{"sink", "blocking-backend", "queued-backend"} {
		b.Run(stage, func(b *testing.B) {
			sink := &benchmarkOutput{}
			owner, err := NewIO(IOConfig{Stdout: sink})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				owner.Close()
				if err := owner.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			ctx, cancel := context.WithCancel(owner.ctx)
			defer cancel()
			stream := &ioStream{owner: owner, id: 1, source: owner.stdout, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
			owner.records[1] = stream
			owner.stdout.refs++
			b.ReportAllocs()
			b.SetBytes(4096)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				switch stage {
				case "sink":
					if n, err := sink.Write(ctx, bytes); err != nil || n != 4096 {
						b.Fatal(n, err)
					}
				case "blocking-backend":
					if err := stream.blockingWrite(ctx, text); err != nil {
						b.Fatal(err)
					}
				default:
					if _, err := stream.checkWrite(); err != nil {
						b.Fatal(err)
					}
					if err := stream.write(text); err != nil {
						b.Fatal(err)
					}
					if err := stream.flush(); err != nil {
						b.Fatal(err)
					}
					if err := stream.waitFlush(ctx); err != nil {
						b.Fatal(err)
					}
				}
			}
			b.StopTimer()
			sink.mu.Lock()
			written := sink.bytes
			sink.mu.Unlock()
			if written != uint64(b.N)*4096 {
				b.Fatal(written)
			}
		})
	}
}
