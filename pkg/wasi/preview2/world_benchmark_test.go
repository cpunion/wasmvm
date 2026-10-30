//go:build linux || darwin

package preview2

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func BenchmarkWASICommand(b *testing.B) {
	path := os.Getenv("WASMVM_COMMAND_BENCH_PREOPEN")
	if path == "" {
		path = b.TempDir()
	}
	for _, guest := range []struct{ name, output string }{{"world-c-command", "c-command:ok\n"}, {"world-composed-command", "rust-command:ok\n"}} {
		start := time.Now()
		input, err := os.ReadFile("testdata/" + guest.name + ".wasm")
		load := time.Since(start)
		if err != nil {
			b.Fatal(err)
		}
		for _, backend := range []string{"interpreter", "native"} {
			if backend == "native" && !native.Available {
				continue
			}
			b.Run(guest.name+"/"+backend, func(b *testing.B) {
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
				out := NewOutputBuffer(1 << 20)
				w, err := NewWorld(WorldConfig{Environment: EnvironmentConfig{Args: []string{"wasmvm", "world"}}, IO: IOConfig{Stdout: out}, FileSystem: FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/data", HostPath: path, ReadOnly: true}}}})
				setup := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				defer func() { w.Close(); _ = w.WaitClosed(nil) }()
				start = time.Now()
				in, err := module.Instantiate(nil, w.Imports(), component.CallOptions{})
				instantiate := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				defer in.Close()
				start = time.Now()
				function, err := in.Function(CLICommandInterface, "run")
				lookup := time.Since(start)
				if err != nil {
					b.Fatal(err)
				}
				call := func() {
					result, err := function.Call(nil, nil, component.CallOptions{})
					if err != nil || result.Value.Case != 0 {
						b.Fatal(result, err)
					}
				}
				start = time.Now()
				call()
				first := time.Since(start)
				for i := 0; i < 10; i++ {
					call()
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					call()
				}
				b.StopTimer()
				if expected := bytes.Repeat([]byte(guest.output), b.N+11); !bytes.Equal(out.Drain(), expected) {
					b.Fatal("guest output mismatch")
				}
				start = time.Now()
				in.Close()
				w.Close()
				if err := w.WaitClosed(nil); err != nil {
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
}
