//go:build linux || darwin

package preview2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestWorldIndependentLanguageCommands(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, guest := range []struct{ name, output string }{{"world-c-command", "c-command:ok\n"}, {"world-composed-command", "rust-command:ok\n"}} {
			t.Run(backend+"/"+guest.name, func(t *testing.T) {
				out := NewOutputBuffer(4096)
				w, err := NewWorld(WorldConfig{
					Environment: EnvironmentConfig{Args: []string{"wasmvm", "world"}},
					IO:          IOConfig{Stdout: out},
					FileSystem:  FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/data", HostPath: t.TempDir()}}},
				})
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					w.Close()
					if err := w.WaitClosed(context.Background()); err != nil {
						t.Error(err)
					}
				}()
				config := component.DefaultConfig()
				config.Core.Backend = backend
				engine, err := component.NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				defer engine.Close()
				module, err := engine.Compile(nil, fixture(t, guest.name))
				if err != nil {
					t.Fatal(err)
				}
				defer module.Close()
				in, err := module.Instantiate(nil, w.Imports(), component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				if code, err := RunCommand(nil, in, component.CallOptions{Metered: true, Fuel: 10000000}); err != nil || code != 0 {
					t.Fatal(code, err)
				}
				if got := string(out.Drain()); got != guest.output {
					t.Fatalf("output %q; want %q", got, guest.output)
				}
			})
		}
	}
}

func TestWorldCompleteCLIImportsAndCommandOutcomes(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"full-command", "command-success", "command-failure", "command-trap", "command-wrong-result", "command-wrong-parameter", "command-no-result", "command-exit"} {
			t.Run(backend+"/"+mode, func(t *testing.T) {
				w, err := NewWorld(WorldConfig{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { w.Close(); _ = w.WaitClosed(nil) }()
				config := component.DefaultConfig()
				config.Core.Backend = backend
				engine, err := component.NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				defer engine.Close()
				module, err := engine.Compile(nil, fixture(t, "world-"+mode))
				if err != nil {
					t.Fatal(err)
				}
				defer module.Close()
				marks := 0
				imports := w.Imports()
				imports.Functions = map[string]component.HostFunction{"mark": func(*component.Caller, []component.Value) (*component.Value, error) { marks++; return nil, nil }}
				in, err := module.Instantiate(nil, imports, component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer in.Close()
				code, err := RunCommand(nil, in, component.CallOptions{})
				switch mode {
				case "command-wrong-result", "command-wrong-parameter", "command-no-result":
					if !errors.Is(err, vm.ErrType) || marks != 0 {
						t.Fatal("invalid export executed guest effects", marks, err)
					}
				case "command-trap":
					if !errors.Is(err, wasm.ErrTrap) || marks != 1 {
						t.Fatal(marks, err)
					}
				default:
					want := uint32(0)
					if mode == "command-failure" {
						want = 1
					}
					if mode == "command-exit" {
						want = 255
					}
					if err != nil || code != want {
						t.Fatal(code, want, err)
					}
				}
				if mode == "command-success" {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					if _, err := RunCommand(ctx, in, component.CallOptions{}); !errors.Is(err, context.Canceled) || marks != 1 {
						t.Fatal("canceled run executed effects", marks, err)
					}
				}
			})
		}
	}
}

type worldInput struct {
	BytesInput
	closes int
	err    error
}

func (s *worldInput) Close() error { s.closes++; _ = s.BytesInput.Close(); return s.err }

func TestWorldConstructorRollbackAndOwnerShutdown(t *testing.T) {
	for _, config := range []WorldConfig{
		{Environment: EnvironmentConfig{Env: []string{"bad"}}},
		{Random: RandomConfig{MaxBytes: 1}},
		{WallClock: WallClockConfig{Now: systemWallNow}},
		{IO: IOConfig{MaxReadBytes: 1 << 28}},
		{FileSystem: FileSystemConfig{Preopens: []FileSystemPreopen{{HostPath: t.TempDir() + "/absent"}}}},
		{Network: NetworkConfig{ResolveTimeout: -1}},
		{UDP: UDPConfig{MaxDatagrams: 65}},
		{HTTP: HTTPConfig{Timeout: -1}},
	} {
		if w, err := NewWorld(config); err == nil || w != nil {
			t.Fatal("invalid world constructed", w, err)
		}
	}
	input := &worldInput{}
	if _, err := NewWorld(WorldConfig{IO: IOConfig{Stdin: input}, HTTP: HTTPConfig{Timeout: -1}}); err == nil || input.closes != 1 {
		t.Fatal("rollback did not reclaim accepted IO source", input.closes, err)
	}
	failure := errors.New("source close failed")
	input = &worldInput{err: failure}
	if _, err := NewWorld(WorldConfig{IO: IOConfig{Stdin: input}, HTTP: HTTPConfig{Timeout: -1}}); !errors.Is(err, failure) || !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("rollback discarded primary or cleanup failure", err)
	}
	input = &worldInput{err: failure}
	w, err := NewWorld(WorldConfig{IO: IOConfig{Stdin: input}})
	if err != nil {
		t.Fatal(err)
	}
	if w.IO() == nil || w.HTTP() == nil {
		t.Fatal("missing shared owners")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Add(1)
		go func() { defer group.Done(); w.Close() }()
	}
	group.Wait()
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := w.WaitClosed(ctx); !errors.Is(err, failure) || input.closes != 1 {
		t.Fatal(input.closes, err)
	}
	var zero *World
	zero.Close()
	if zero.IO() != nil || zero.HTTP() != nil || len(zero.Imports().Instances) != 0 || !errors.Is(zero.WaitClosed(nil), vm.ErrReference) {
		t.Fatal("invalid zero-world behavior")
	}
	if _, err := RunCommand(nil, nil, component.CallOptions{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
}

func TestWorldHTTPImportsAndSharedNominalDomains(t *testing.T) {
	server := httptest.NewServer(httpEcho(t))
	defer server.Close()
	target, _ := url.Parse(server.URL)
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			w, err := NewWorld(WorldConfig{HTTP: HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback}})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { w.Close(); _ = w.WaitClosed(nil) }()
			config := component.DefaultConfig()
			config.Core.Backend = backend
			engine, _ := component.NewEngine(config)
			defer engine.Close()
			module, err := engine.Compile(nil, fixture(t, "http"))
			if err != nil {
				t.Fatal(err)
			}
			defer module.Close()
			in, err := module.Instantiate(nil, w.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			probe, _ := in.Function("probe")
			if value, err := probe.Call(nil, []component.Value{{Text: target.Host}, {}}, component.CallOptions{}); err != nil || value.Value.Bits != 42 {
				t.Fatal(value, err)
			}
			request := httptest.NewRequest("POST", "http://localhost/echo", bytes.NewReader(httpPattern()))
			response, err := w.HTTP().CallIncoming(nil, in, request, component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil || !bytes.Equal(data, httpPattern()) || response.Trailer.Get("X-Reply-Trailer") != "done" {
				t.Fatal(len(data), response.Trailer, err)
			}
		})
	}
}

func TestWorldInvalidIncomingExportRejectsBeforeEffects(t *testing.T) {
	w, err := NewWorld(WorldConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { w.Close(); _ = w.WaitClosed(nil) }()
	engine, _ := component.NewEngine(component.DefaultConfig())
	defer engine.Close()
	module, err := engine.Compile(nil, fixture(t, "world-invalid-incoming-result"))
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	called := false
	imports := w.Imports()
	imports.Functions = map[string]component.HostFunction{"mark-incoming": func(*component.Caller, []component.Value) (*component.Value, error) {
		called = true
		return &component.Value{Bits: 42}, nil
	}}
	in, err := module.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	if _, err := w.HTTP().CallIncoming(nil, in, httptest.NewRequest("GET", "http://localhost/", nil), component.CallOptions{}); !errors.Is(err, vm.ErrType) || called {
		t.Fatal("invalid standard handler executed guest effects", called, err)
	}
}
