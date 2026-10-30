//go:build linux || darwin

package preview2

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func fsImports(t *testing.T, p *FileSystem) component.Imports {
	t.Helper()
	imports := p.Imports()
	clock, err := NewWallClock(WallClockConfig{})
	if err != nil {
		t.Fatal(err)
	}
	imports.Instances[WallClockInterface] = clock.Imports().Instances[WallClockInterface]
	return imports
}
func fsOwner(t *testing.T, config FileSystemConfig) (*FileSystem, *IO, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bench.bin"), []byte("0123456789abcdef"), 0600); err != nil {
		t.Fatal(err)
	}
	process, err := NewIO(IOConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if config.Preopens == nil {
		config.Preopens = []FileSystemPreopen{{GuestPath: "/sandbox", HostPath: root}}
	}
	p, err := NewFileSystem(process, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close()
		process.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
		if err := process.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
	})
	return p, process, root
}
func fsInstance(t *testing.T, p *FileSystem, backend, fixtureName string) *component.Instance {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	engine, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(engine.Close)
	module, err := engine.Compile(nil, fixture(t, fixtureName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(module.Close)
	in, err := module.Instantiate(nil, fsImports(t, p), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}
func TestFileSystemGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, _, root := fsOwner(t, FileSystemConfig{})
			in := fsInstance(t, p, backend, "filesystem")
			for _, name := range []string{"exercise", "driver", "stat-batch"} {
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				var args []component.Value
				if name == "stat-batch" {
					args = []component.Value{{Bits: 3}}
				}
				result, err := f.Call(nil, args, component.CallOptions{Metered: true, Fuel: 1000000})
				if err != nil {
					t.Fatal(name, err)
				}
				want := uint64(42)
				if name == "driver" {
					want = 1
				}
				if name == "stat-batch" {
					want = 3
				}
				if result.Value.Bits != want {
					t.Fatal(name, result)
				}
			}
			names, err := os.ReadDir(root)
			if err != nil || len(names) != 1 || names[0].Name() != "bench.bin" {
				t.Fatal(names, err)
			}
			p.mu.Lock()
			desc, entries, backings := len(p.descriptors), len(p.entries), len(p.backings)
			p.mu.Unlock()
			if desc != 0 || entries != 0 || backings != 1 {
				t.Fatal("retained resources", desc, entries, backings)
			}
		})
	}
}
