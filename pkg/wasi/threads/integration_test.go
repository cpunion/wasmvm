//go:build linux || darwin

package threads

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
)

type corpusManifest struct {
	Official  []string
	SHA256    map[string]string
	Producers struct{ Revision, StrippedSHA256 string }
}

func corpus(t *testing.T) corpusManifest {
	t.Helper()
	b, err := os.ReadFile("../../../testdata/wasi-threads/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var m corpusManifest
	if err := json.Unmarshal(b, &m); err != nil || len(m.Official) != 15 {
		t.Fatal("thread manifest", err)
	}
	for name, want := range m.SHA256 {
		b, err := os.ReadFile(filepath.Join("../../../artifacts/wasi-threads", name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("thread asset changed", name)
		}
	}
	return m
}

func executionReport(t *testing.T, name string, programs, completed int) {
	t.Helper()
	path := os.Getenv("WASMVM_CONFORMANCE_REPORT")
	if path == "" {
		return
	}
	tier := 0
	if strings.HasSuffix(t.Name(), "/native") {
		tier = 1
	}
	if strings.HasSuffix(t.Name(), "/fuel") {
		tier = 2
	}
	b, err := os.ReadFile("../../../testdata/wasi-threads/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	record := map[string]any{"Profile": name, "Tier": tier, "Programs": programs, "CompletedPrograms": completed, "Passed": completed == programs && !t.Failed(), "ManifestSHA256": hex.EncodeToString(sum[:]), "Compiler": runtime.Compiler, "RuntimeVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "SourceRevision": "f8ea03cc0ffbde143415e945497de112f6f9fd53", "GuestProducer": "8e45ed5e8bbd080995b0516ef40ec1f5218d2e1f", "NativeCodeBudget": 256 << 20}
	b, err = json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(filepath.Dir(path), "wasi", name+"-"+string(rune('0'+tier))+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestOfficialWASIThreads(t *testing.T) {
	m := corpus(t)
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		completed := 0
		for _, name := range m.Official {
			path := filepath.Join("../../../artifacts/wasi-threads/official", name)
			if t.Run(name, func(t *testing.T) {
				b, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				m, err := e.Compile(ctx, b)
				if err != nil {
					t.Fatal(err)
				}
				defer m.Close()
				read, write, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer write.Close()
				pc := preview1.DefaultConfig()
				pc.Stdin, err = preview1.FileStream(read)
				if err != nil {
					t.Fatal(err)
				}
				p, err := preview1.NewProcess(pc)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				cfg := DefaultConfig()
				cfg.Execution = options
				g, err := New(ctx, e, m, p, cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer g.Close()
				var expected struct {
					ExitCode uint32 `json:"exit_code"`
				}
				if metadata, err := os.ReadFile(path[:len(path)-5] + ".json"); err == nil {
					if err := json.Unmarshal(metadata, &expected); err != nil {
						t.Fatal(err)
					}
				} else if !errors.Is(err, os.ErrNotExist) {
					t.Fatal(err)
				}
				err = g.Run()
				code := uint32(0)
				var exit *preview1.ExitError
				if errors.As(err, &exit) {
					code, err = exit.Code, nil
				}
				if err != nil || code != expected.ExitCode {
					t.Fatal("thread result", code, expected.ExitCode, err)
				}
				if err := g.WaitClosed(ctx); err != nil {
					t.Fatal("group cleanup", err)
				}
			}) {
				completed++
			}
		}
		executionReport(t, "legacy-wasi-threads", len(m.Official), completed)
	})
}

func TestLLGoThreadedGuest(t *testing.T) {
	corpus(t)
	b, err := os.ReadFile("../../../artifacts/wasi-threads/guest-llgo.wasm")
	if err != nil {
		t.Fatal(err)
	}
	engines(t, func(t *testing.T, e *vm.Engine, options vm.CallOptions) {
		// The full guest compiles once per tier; child Stores retain this code.
		m, err := e.Compile(nil, b)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		completed := 0
		for _, mode := range []string{"all", "exit", "panic"} {
			if t.Run(mode, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				pc := preview1.DefaultConfig()
				pc.Args = []string{"guest", mode}
				dir := t.TempDir()
				pc.Preopens = []preview1.Preopen{{GuestPath: "/work", HostPath: dir}}
				out, diagnostic := preview1.NewBuffer(1<<20), preview1.NewBuffer(1<<20)
				pc.Stdout, pc.Stderr = out, diagnostic
				p, err := preview1.NewProcess(pc)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				cfg := DefaultConfig()
				cfg.Execution = options
				g, err := New(ctx, e, m, p, cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer g.Close()
				observer, err := e.NewStore(vm.DefaultStoreConfig())
				if err != nil {
					t.Fatal(err)
				}
				defer observer.Close()
				memory, err := observer.ImportMemory(g.memories[[2]string{"env", "memory"}])
				if err != nil {
					t.Fatal(err)
				}
				initialPages, err := memory.MemoryPages()
				if err != nil {
					t.Fatal(err)
				}
				err = g.Run()
				code, expected := uint32(0), uint32(0)
				var exit *preview1.ExitError
				if errors.As(err, &exit) {
					code, err = exit.Code, nil
				}
				if mode == "exit" {
					expected = 23
				}
				if mode == "panic" {
					expected = 2
				}
				if err != nil || code != expected {
					t.Fatalf("exit %d want %d: %v\n%s\n%s", code, expected, err, out.Bytes(), diagnostic.Bytes())
				}
				if mode == "all" && string(out.Bytes()) != "llgo-threads-ok 828\n" {
					t.Fatal("guest output", string(out.Bytes()), string(diagnostic.Bytes()))
				}
				if mode == "panic" && !bytes.Contains(diagnostic.Bytes(), []byte("expected-worker-panic")) {
					t.Fatal("guest panic missing", string(diagnostic.Bytes()))
				}
				if mode == "all" {
					pages, err := memory.MemoryPages()
					if err != nil || pages <= initialPages {
						t.Fatal("guest did not grow shared memory", initialPages, pages, err)
					}
					entries, err := os.ReadDir(dir)
					if err != nil || len(entries) != 8 {
						t.Fatal("thread files", len(entries), err)
					}
				}
				if err := g.WaitClosed(ctx); err != nil {
					t.Fatal(err)
				}
			}) {
				completed++
			}
		}
		executionReport(t, "llgo-threaded-guest", 3, completed)
	})
}
