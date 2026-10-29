//go:build linux || darwin

package preview1

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
)

type programConfig struct {
	Root, GuestRoot, Stdin string
	Args                   []string
	Env                    map[string]string
	ExitCode               uint32 `json:"exit_code"`
	Stdout                 *string
}

func runProgram(t *testing.T, engine *vm.Engine, options vm.CallOptions, path string, metadata programConfig) {
	t.Helper()
	config := DefaultConfig()
	config.Args = append([]string{filepath.Base(path)}, metadata.Args...)
	keys := []string{}
	for key := range metadata.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		config.Env = append(config.Env, key+"="+metadata.Env[key])
	}
	stdout, stderr := NewBuffer(1<<20), NewBuffer(1<<20)
	config.Stdin, config.Stdout, config.Stderr = BytesInput([]byte(metadata.Stdin)), stdout, stderr
	if metadata.Root != "" {
		dir := t.TempDir()
		source := filepath.Join(filepath.Dir(path), metadata.Root)
		if _, err := os.Stat(source); err == nil {
			if err := os.CopyFS(dir, os.DirFS(source)); err != nil {
				t.Fatal(err)
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			t.Fatal(err)
		}
		guest := metadata.GuestRoot
		if guest == "" {
			guest = "/"
		}
		config.Preopens = []Preopen{{GuestPath: guest, HostPath: dir}}
	}
	p, err := NewProcess(config)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	store, err := engine.NewStore(vm.DefaultStoreConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	linker := vm.NewLinker(store)
	if err := p.Install(store, linker); err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(p.Context(), 20*time.Second)
	defer cancel()
	module, err := engine.Compile(ctx, bytes)
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	instance, err := linker.Instantiate(ctx, module, options)
	if err == nil {
		if options.Metered {
			options.Fuel = instance.StartFuelRemaining()
		}
		var start vm.Extern
		start, err = instance.Export("_start")
		if err == nil {
			_, err = start.Call(ctx, nil, options)
		}
	}
	code := uint32(0)
	var exit *ExitError
	if errors.As(err, &exit) {
		code = exit.Code
		err = nil
	}
	if err != nil || code != metadata.ExitCode {
		t.Fatalf("exit %d want %d, error %v\nstdout: %s\nstderr: %s", code, metadata.ExitCode, err, stdout.Bytes(), stderr.Bytes())
	}
	if metadata.Stdout != nil && string(stdout.Bytes()) != *metadata.Stdout {
		t.Fatalf("stdout %q want %q", stdout.Bytes(), *metadata.Stdout)
	}
	p.Close()
	if err := p.WaitClosed(ctx); err != nil {
		t.Fatal("P1 cleanup", err)
	}
	store.Close()
	if err := store.WaitClosed(ctx); err != nil {
		t.Fatal("Store cleanup", err)
	}
}

func TestP1Conformance(t *testing.T) {
	root := "../../../artifacts/preview1/fixtures"
	raw, err := os.ReadFile("../../../testdata/wasi-preview1/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Tests  []struct{ Path string }
		SHA256 map[string]string
	}
	if err = json.Unmarshal(raw, &manifest); err != nil || len(manifest.Tests) != 72 {
		t.Fatal("P1 manifest", err, len(manifest.Tests))
	}
	for name, want := range manifest.SHA256 {
		bytes, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(bytes)
		if hex.EncodeToString(hash[:]) != want {
			t.Fatal("P1 asset hash changed", name)
		}
	}
	sum := sha256.Sum256(raw)
	for _, tier := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("tier-%d", tier), func(t *testing.T) {
			cfg := vm.DefaultEngineConfig()
			cfg.Profile = "core3"
			if tier != 0 {
				cfg.Backend = "native"
			}
			engine, err := vm.NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			options := vm.CallOptions{Metered: tier == 2, Fuel: 1 << 42}
			completed := 0
			for _, test := range manifest.Tests {
				if t.Run(test.Path, func(t *testing.T) {
					path := filepath.Join(root, test.Path)
					bytes, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					hash := sha256.Sum256(bytes)
					if hex.EncodeToString(hash[:]) != manifest.SHA256[test.Path] {
						t.Fatal("P1 input hash changed")
					}
					var metadata programConfig
					if b, err := os.ReadFile(path[:len(path)-5] + ".json"); err == nil {
						if err := json.Unmarshal(b, &metadata); err != nil {
							t.Fatal(err)
						}
					} else if !errors.Is(err, fs.ErrNotExist) {
						t.Fatal(err)
					}
					runProgram(t, engine, options, path, metadata)
				}) {
					completed++
				}
			}
			writeWASIReport(t, fmt.Sprintf("preview1-%d", tier), map[string]any{"Passed": !t.Failed() && completed == 72, "Profile": "wasi_snapshot_preview1", "Tier": tier, "Programs": 72, "CompletedPrograms": completed, "RegisteredOperations": len(functions), "ManifestSHA256": hex.EncodeToString(sum[:]), "SourceRevision": "58b6206a46013fb53d23990e5a650584a5361043", "PrebuiltRevision": "9dfd4867f12498bd6411a57739c24b4880668683"})
			if completed != 72 {
				t.Fatalf("executed %d/72 P1 programs", completed)
			}
		})
	}
}

func TestOfficialGoGuest(t *testing.T) {
	path := "../../../artifacts/preview1/fixtures/integration/go-guest.wasm"
	raw, err := os.ReadFile("../../../testdata/wasi-preview1/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct{ SHA256 map[string]string }
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	input, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(input)
	if hex.EncodeToString(hash[:]) != manifest.SHA256["integration/go-guest.wasm"] {
		t.Fatal("Go guest hash changed")
	}
	expected := "go-p1-ok 42400\n"
	for _, tier := range []int{0, 1, 2} {
		passed := t.Run(fmt.Sprint(tier), func(t *testing.T) {
			cfg := vm.DefaultEngineConfig()
			cfg.Profile = "core3"
			cfg.Native.MaxCodeBytes = 256 << 20
			if tier != 0 {
				cfg.Backend = "native"
			}
			engine, err := vm.NewEngineWithConfig(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			runProgram(t, engine, vm.CallOptions{Metered: tier == 2, Fuel: 1 << 42}, path, programConfig{Root: "empty", GuestRoot: "/sandbox", Stdin: "input", Args: []string{"argument"}, Env: map[string]string{"VM_TEST": "guest"}, Stdout: &expected})
		})
		completed := 0
		if passed {
			completed = 1
		}
		sum := sha256.Sum256(raw)
		writeWASIReport(t, fmt.Sprintf("go-guest-%d", tier), map[string]any{"Passed": passed, "Profile": "go-wasip1-guest", "Tier": tier, "Programs": 1, "CompletedPrograms": completed, "NativeCodeBudget": 256 << 20, "GuestSHA256": hex.EncodeToString(hash[:]), "ManifestSHA256": hex.EncodeToString(sum[:]), "Producer": "go1.27.3"})
	}
}

func writeWASIReport(t *testing.T, name string, report map[string]any) {
	t.Helper()
	base := os.Getenv("WASMVM_CONFORMANCE_REPORT")
	if base == "" {
		return
	}
	report["Compiler"] = runtime.Compiler
	report["RuntimeVersion"] = runtime.Version()
	report["GOOS"] = runtime.GOOS
	report["GOARCH"] = runtime.GOARCH
	dir := filepath.Join(filepath.Dir(base), "wasi")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, name+".json"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}
