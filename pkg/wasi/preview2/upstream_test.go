//go:build linux || darwin

package preview2

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestUpstreamRustStandardCommands(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			engine, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			for _, name := range []string{"args", "argv0", "env", "hello_stdout", "stdin", "default_clocks", "sleep", "exit_default", "exit_success", "exit_failure", "exit_panic", "file_read", "file_append", "file_dir_sync"} {
				t.Run(name, func(t *testing.T) {
					dir := t.TempDir()
					const initial = "And stood awhile in thought"
					path := filepath.Join(dir, "bar.txt")
					if err := os.WriteFile(path, []byte(initial), 0600); err != nil {
						t.Fatal(err)
					}
					args := []string{"wasmvm"}
					if name == "args" {
						args = append(args, "hello", "this", "", "is an argument", "with 🚩 emoji")
					}
					if name == "argv0" {
						args = nil
					}
					out, stderr := NewOutputBuffer(4096), NewOutputBuffer(4096)
					cwd := "/"
					world, err := NewWorld(WorldConfig{
						Environment: EnvironmentConfig{Args: args, Env: []string{"frabjous=day", "callooh=callay"}, InitialCWD: &cwd},
						IO:          IOConfig{Stdin: NewBytesInput([]byte("So rested he by the Tumtum tree")), Stdout: out, Stderr: stderr},
						FileSystem:  FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/", HostPath: dir}}},
					})
					if err != nil {
						t.Fatal(err)
					}
					defer func() {
						world.Close()
						if err := world.WaitClosed(nil); err != nil {
							t.Error(err)
						}
					}()
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					compiled, err := engine.Compile(ctx, fixture(t, "upstream-p2_cli_"+name))
					if err != nil {
						t.Fatal(err)
					}
					defer compiled.Close()
					in, err := compiled.Instantiate(ctx, world.Imports(), component.CallOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer in.Close()
					code, err := RunCommand(ctx, in, component.CallOptions{Metered: true, Fuel: 10000000})
					want := uint32(0)
					if name == "exit_failure" {
						want = 1
					}
					if name == "exit_panic" {
						if !errors.Is(err, wasm.ErrTrap) || !strings.Contains(string(stderr.Drain()), "Curiouser and curiouser!") {
							t.Fatal("wrong guest panic", code, err)
						}
					} else if err != nil || code != want {
						t.Fatal(code, err)
					}
					if name == "hello_stdout" && (string(out.Drain()) != "hello, world\n" || string(stderr.Drain()) != "hello, world\n") {
						t.Fatal("lost guest stdio")
					}
					if name == "file_append" {
						data, err := os.ReadFile(path)
						want := initial + "Did gyre and gimble in the wabe;\nAll mimsy were the borogoves,\nAnd the mome raths outgrabe.\n"
						if err != nil || string(data) != want {
							t.Fatal("seek changed append semantics", string(data), err)
						}
					}
				})
			}
		})
	}
}

func TestUpstreamCommandProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/upstream-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Commit    string            `json:"sourceCommit"`
		Rust      string            `json:"rustVersion"`
		Target    string            `json:"target"`
		Generator string            `json:"generatorSha256"`
		Sources   map[string]string `json:"sourcesSha256"`
		Binaries  map[string]string `json:"binarySha256"`
	}
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Commit != "7bac2c2775808aaec5d4aa5627a5e447b51102cf" || pins.Rust != "1.98.0" || pins.Target != "wasm32-wasip2" || len(pins.Binaries) != 14 || len(pins.Sources) != 15 {
		t.Fatal("wrong upstream profile", pins)
	}
	groups := []struct {
		root   string
		hashes map[string]string
	}{
		{"testdata", pins.Binaries}, {"../../../testdata/wasi-preview2/upstream-wasmtime", pins.Sources},
		{"../../../scripts", map[string]string{"build_upstream_wasi_fixtures.py": pins.Generator}},
	}
	for _, group := range groups {
		for name, expected := range group.hashes {
			data, err := os.ReadFile(filepath.Join(group.root, name))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data)
			if hex.EncodeToString(sum[:]) != expected {
				t.Fatal("upstream fixture checksum changed", name)
			}
		}
	}
}
