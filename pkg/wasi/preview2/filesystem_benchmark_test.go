//go:build linux || darwin

package preview2

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
)

func BenchmarkFileSystemStatBatch(b *testing.B) {
	start := time.Now()
	input, err := os.ReadFile("testdata/filesystem.wasm")
	load := time.Since(start)
	if err != nil {
		b.Fatal(err)
	}
	root := b.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bench.bin"), []byte("0123456789abcdef"), 0600); err != nil {
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
			fs, err := NewFileSystem(process, FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/sandbox", HostPath: root}}})
			if err != nil {
				b.Fatal(err)
			}
			defer func() {
				fs.Close()
				if err := fs.WaitClosed(nil); err != nil {
					b.Error(err)
				}
			}()
			imports := fs.Imports()
			clock, err := NewWallClock(WallClockConfig{})
			if err != nil {
				b.Fatal(err)
			}
			imports.Instances[WallClockInterface] = clock.Imports().Instances[WallClockInterface]
			setup := time.Since(start)
			start = time.Now()
			in, err := module.Instantiate(nil, imports, component.CallOptions{})
			instantiate := time.Since(start)
			if err != nil {
				b.Fatal(err)
			}
			defer in.Close()
			start = time.Now()
			function, err := in.Function("stat-batch")
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
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/128, "ns/stat")
			start = time.Now()
			in.Close()
			fs.Close()
			process.Close()
			if err := fs.WaitClosed(nil); err != nil {
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
func BenchmarkFileSystemStatStages(b *testing.B) {
	root := b.TempDir()
	path := filepath.Join(root, "bench.bin")
	if err := os.WriteFile(path, []byte("0123456789abcdef"), 0600); err != nil {
		b.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	st, err := host.Fstat(file)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("host-fstat", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			stat, err := host.Fstat(file)
			if err != nil || stat.Size != 16 {
				b.Fatal(stat, err)
			}
		}
	})
	b.Run("stat-value", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			value, err := fsStatValue(st)
			if err != nil || value.Elements[2].Bits != 16 {
				b.Fatal(value, err)
			}
		}
	})
}
