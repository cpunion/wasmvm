//go:build linux || darwin

package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasi/preview1"
	"github.com/cpunion/wasmvm/pkg/wasi/preview2"
	"golang.org/x/sys/unix"
)

func TestComponentCommands(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, preview := range []string{"2", "3"} {
			for _, language := range []string{"c", "rust"} {
				name := "world-" + language + "-command"
				want := language + "-p3-command:ok\n"
				if preview == "2" {
					want = language + "-command:ok\n"
					if language == "rust" {
						name = "world-composed-command"
					}
				}
				path := "../../pkg/wasi/preview" + preview + "/testdata/" + name + ".wasm"
				code, out, diag := invoke(t, nil, "run", "--wasi-preview", preview, "--backend", backend, "--fuel", "10000000", "--read-only-dir", "/data="+t.TempDir(), path, "world")
				if code != 0 || out != want || diag != "" {
					t.Fatal(backend, preview, language, code, out, diag)
				}
			}
			path := "testdata/component-init-p" + preview + ".wasm"
			for _, fuel := range []string{"4", "5"} {
				code, _, diag := invoke(t, nil, "run", "--wasi-preview", preview, "--backend", backend, "--fuel", fuel, path)
				if fuel == "4" {
					if code != 1 || !strings.Contains(diag, "fuel") {
						t.Fatal("initialization fuel was reset", backend, preview, code, diag)
					}
				} else if code != 0 || diag != "" {
					t.Fatal("exact shared budget", backend, preview, code, diag)
				}
			}
		}
		for _, mode := range []string{"success", "failure", "exit", "trap", "sync"} {
			code, _, diag := invoke(t, nil, "run", "--wasi-preview", "3", "--backend", backend, "../../pkg/wasi/preview3/testdata/command-"+mode+".wasm")
			want := 0
			if mode == "failure" || mode == "trap" || mode == "sync" {
				want = 1
			}
			if mode == "exit" {
				want = 255
			}
			if code != want || (mode == "success" || mode == "failure" || mode == "exit") && diag != "" || (mode == "trap" || mode == "sync") && diag == "" {
				t.Fatal(mode, backend, code, diag)
			}
		}
	}
}

func TestComponentCommandStdin(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "input")
	if err := os.WriteFile(path, []byte("So rested he by the Tumtum tree"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		for _, preview := range []string{"2", "3"} {
			fixture := "testdata/component-stdin-p3.wasm"
			if preview == "2" {
				fixture = "../../pkg/wasi/preview2/testdata/upstream-p2_cli_stdin.wasm"
			}
			code, _, diag := invoke(t, nil, "run", "--wasi-preview", preview, "--backend", backend, "--stdin", path, fixture)
			if code != 0 || diag != "" {
				t.Fatal(backend, preview, code, diag)
			}
		}
	}
}

func TestComponentMixedPreviewImports(t *testing.T) {
	root := os.Getenv("WASMVM_WASI_UPSTREAM_CORPUS")
	if root == "" {
		root = "../../artifacts/components/wasi-upstream"
	}
	input := filepath.Join(t.TempDir(), "input")
	if err := os.WriteFile(input, []byte("hello!"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		for _, fixture := range []struct{ name, output string }{
			{"p3_cli_hello_stdout", "hello, world\n"},
			{"p3_cli_hello_stdout_post_return", "hello, world\nhello again, after return\n"},
			{"p3_cli_read_stdin", ""}, {"p3_cli_stdout_flush", "READY"},
		} {
			// Rust's wasip2 standard library imports P2 alongside native P3
			// bindings and exports both entry points. Select the P3 entry.
			code, out, diag := invoke(t, nil, "run", "--wasi-preview", "3", "--backend", backend, "--fuel", "100000000", "--stdin", input, filepath.Join(root, "components", fixture.name+".wasm"))
			if code != 0 || out != fixture.output || diag != "" {
				t.Fatal(backend, fixture.name, code, out, diag)
			}
		}
	}
}

func TestComponentCommandInvalidConfiguration(t *testing.T) {
	path := "testdata/component-init-p3.wasm"
	for _, args := range [][]string{
		{"--wasi-preview", "0", "--export", "run"}, {"--wasi-preview", "4"},
		{"--wasi-preview", "3", "--wasi-threads"}, {"--wasi-preview", "2", "--export", "run"},
		{"--wasi-preview", "3", "--env", "bad"}, {"--wasi-preview", "3", "--dir", "bad"},
	} {
		code, _, diag := invoke(t, nil, append(append([]string{"run"}, args...), path)...)
		if code != 2 || diag == "" {
			t.Fatal(args, code, diag)
		}
	}
	for _, args := range [][]string{
		{"--profile", "bad"}, {"--backend", "bad"}, {"--stdin", "/definitely/missing"},
		{"--dir", "/data=/definitely/missing"}, {"--env", "K=\x00"}, {"--fuel", "0"},
		{"--backend", "native", "--native-code-bytes", "0"},
	} {
		code, _, diag := invoke(t, nil, append(append([]string{"run", "--wasi-preview", "3"}, args...), path)...)
		if code != 1 || diag == "" {
			t.Fatal(args, code, diag)
		}
	}
	for _, fixture := range []string{"/definitely/missing", "testdata/wasi-stdio.wasm"} {
		code, _, diag := invoke(t, context.Background(), "run", "--wasi-preview", "3", fixture)
		if code != 1 || diag == "" {
			t.Fatal(fixture, code, diag)
		}
	}
}

func TestComponentCommandFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/components.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Producer  string
		Generator string            `json:"generatorSHA256"`
		Sources   map[string]string `json:"sourcesSHA256"`
		Fixtures  map[string]map[string]string
	}
	if err := json.Unmarshal(data, &pins); err != nil || pins.Producer != "wasm-tools 1.258.0" || len(pins.Fixtures) != 3 {
		t.Fatal(pins, err)
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		sum := sha256.Sum256(data)
		if err != nil || hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed fixture input", path, err)
		}
	}
	check("../../scripts/build_component_cli_fixtures.py", pins.Generator)
	for path, want := range pins.Sources {
		check("../../"+path, want)
	}
	for name, files := range pins.Fixtures {
		for ext, want := range files {
			check("testdata/"+name+"."+ext, want)
		}
	}
}

func TestComponentInitializationBudgetReceipt(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		config := component.DefaultConfig()
		config.Core.Backend = backend
		engine, err := component.NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		data, err := os.ReadFile("testdata/component-init-p3.wasm")
		if err != nil {
			t.Fatal(err)
		}
		module, err := engine.Compile(nil, data)
		if err != nil {
			t.Fatal(err)
		}
		defer module.Close()
		in, err := module.Instantiate(nil, component.Imports{}, component.CallOptions{Metered: true, Fuel: 5})
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()
		if in.StartFuelRemaining() != 2 {
			t.Fatal("initialization not charged", in.StartFuelRemaining())
		}
		fn, _ := in.Function("wasi:cli/run@0.3.2", "run")
		result, err := fn.Call(nil, nil, component.CallOptions{Metered: true, Fuel: in.StartFuelRemaining()})
		if err != nil || result.FuelRemaining != 0 || in.StartFuelRemaining() != 2 {
			t.Fatal("mutable or reset initialization receipt", result, err)
		}
	}
}

func TestComponentCommandPipeCancellationAndFlagRestore(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	before, err := unix.FcntlInt(uintptr(host.Descriptor(r)), unix.F_GETFL, 0)
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = original }()
	for _, backend := range []string{"interpreter", "native"} {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		code, _, diag := invoke(t, ctx, "run", "--wasi-preview", "3", "--backend", backend, "testdata/component-stdin-p3.wasm")
		cancel()
		if code != 1 || !strings.Contains(diag, "deadline") {
			t.Fatal(backend, code, diag)
		}
		after, err := unix.FcntlInt(uintptr(host.Descriptor(r)), unix.F_GETFL, 0)
		if err != nil || before != after {
			t.Fatal("stdin flags changed", before, after, err)
		}
	}
}

func TestCommandSourceReadinessAndShutdown(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	var out bytes.Buffer
	original := os.Stdin
	os.Stdin = r
	stdio, err := cliStreams("", &out, &out)
	os.Stdin = original
	if err != nil {
		t.Fatal(err)
	}
	defer stdio.Close()
	if err := stdio.streams[0].(interface{ SetFlags(uint16) error }).SetFlags(4); err != nil {
		t.Fatal(err)
	}
	source := newCommandSource(stdio.streams[0], true)
	defer func() { source.Close(); source.waiters.Wait() }()
	buffer := make([]byte, 8)
	if _, err := source.Read(context.Background(), buffer); !errors.Is(err, preview2.ErrWouldBlock) {
		t.Fatal("blocking empty pipe", err)
	}
	ready, changed := source.Readiness()
	if ready || changed == nil {
		t.Fatal("unreadable pipe reported ready")
	}
	if ready, _ := source.Readiness(); ready {
		t.Fatal("duplicate readiness changed state")
	}
	if _, err := w.Write([]byte("input")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("lost readiness notification")
	}
	if ready, _ := source.Readiness(); !ready {
		t.Fatal("readable pipe reported blocked")
	}
	if n, err := source.Read(context.Background(), buffer); err != nil || string(buffer[:n]) != "input" {
		t.Fatal(n, err)
	}
	w.Close()
	if _, err := source.Read(context.Background(), buffer); !errors.Is(err, io.EOF) {
		t.Fatal("pipe EOF", err)
	}
	source.Close()
	if _, err := source.Read(context.Background(), buffer); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if _, err := source.Write(context.Background(), buffer); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if err := source.Flush(context.Background()); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.Read(ctx, buffer); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if ready, _ := source.Readiness(); !ready {
		t.Fatal("closed stream was not ready")
	}
}

type commandReadinessFailure struct {
	preview1.Stream
	interrupted bool
	readyError  bool
}

func (s commandReadinessFailure) Ready(context.Context, bool) (uint64, bool, bool, error) {
	if s.interrupted {
		return 0, false, false, unix.EINTR
	}
	if s.readyError {
		return 0, false, false, io.ErrUnexpectedEOF
	}
	return 0, false, false, nil
}
func (s commandReadinessFailure) Poll(context.Context, bool) (uint64, bool, error) {
	return 0, false, io.ErrUnexpectedEOF
}

func TestCommandSourceReadinessFailures(t *testing.T) {
	for _, mode := range []string{"ready-error", "poll-error", "interrupted"} {
		s := newCommandSource(commandReadinessFailure{Stream: preview1.BytesInput(nil), readyError: mode == "ready-error", interrupted: mode == "interrupted"}, true)
		ready, changed := s.Readiness()
		if !ready {
			select {
			case <-changed:
			case <-time.After(time.Second):
				t.Fatal("readiness failure was not published", mode)
			}
		}
		if ready, _ := s.Readiness(); !ready {
			t.Fatal("failure was not ready", mode)
		}
		if _, err := s.Read(context.Background(), make([]byte, 1)); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatal("readiness error lost", mode, err)
		}
		s.Close()
		s.waiters.Wait()
	}
}
