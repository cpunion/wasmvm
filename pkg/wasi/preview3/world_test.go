//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestWorldFixtureProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/world-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		Version   string            `json:"wasiVersion"`
		WIT       string            `json:"witArchiveSHA256"`
		Producer  string            `json:"producer"`
		Bindings  string            `json:"bindingGenerator"`
		Generator string            `json:"generatorSHA256"`
		Sources   map[string]string `json:"sourcesSHA256"`
		Binaries  map[string]string `json:"binarySHA256"`
		Unstable  []string          `json:"unstableFeatures"`
	}
	if err = json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	if pins.Version != "0.3.2" || pins.WIT != "2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311" || pins.Producer != "wasm-tools 1.258.0" || pins.Bindings != "wit-bindgen 0.54.0" || len(pins.Binaries) != 8 || len(pins.Unstable) != 0 {
		t.Fatal("changed world producer", pins)
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		sum := sha256.Sum256(data)
		if err != nil || hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed world asset", path, err)
		}
	}
	check("../../../scripts/build_preview3_world_fixtures.py", pins.Generator)
	root := "../../../testdata/wasi-preview3/guests"
	var sources int
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			name, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			check(path, pins.Sources[filepath.ToSlash(name)])
			sources++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if sources != len(pins.Sources) {
		t.Fatal("missing world source", sources, len(pins.Sources))
	}
	for name, want := range pins.Binaries {
		check("testdata/"+name, want)
	}
}

func worldOwnerForTest(t *testing.T, config WorldConfig) *World {
	t.Helper()
	w, err := NewWorld(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		w.Close()
		if err := w.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	})
	return w
}
func TestWorldCompleteProviderAssembly(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		root := t.TempDir()
		if err := os.WriteFile(root+"/data.txt", []byte("abc"), 0600); err != nil {
			t.Fatal(err)
		}
		w := worldOwnerForTest(t, WorldConfig{Environment: EnvironmentConfig{Args: []string{"world"}}, FileSystem: FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/data", HostPath: root}}}})
		imports := w.Imports()
		if len(imports.Instances) != 24 {
			t.Fatal("world interfaces", len(imports.Instances))
		}
		for name := range imports.Instances {
			if !strings.HasSuffix(name, "@0.3.2") {
				t.Fatal("cross-version interface", name)
			}
		}
		if w.IO() == nil || w.HTTP() == nil {
			t.Fatal("missing public owner")
		}
		in := instantiate(t, compileFixture(t, backend, "command-success"), imports)
		if code, err := RunCommand(nil, in, component.CallOptions{}); code != 0 || err != nil {
			t.Fatal(code, err)
		}
		in = instantiate(t, compileFixture(t, backend, "filesystem-guest"), imports)
		if r := call(t, in, "run"); r.Bits != 3 {
			t.Fatal(r)
		}
		in = instantiate(t, compileFixture(t, backend, "http-server"), imports)
		response, err := w.HTTP().CallIncoming(nil, in, httptest.NewRequest("GET", "http://localhost", nil), component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || string(data) != "hello" {
			t.Fatal(string(data), err)
		}
	})
}

type worldInput struct{ closed atomic.Uint32 }

func (s *worldInput) Read(context.Context, []byte) (int, error) { return 0, io.EOF }
func (s *worldInput) Readiness() (bool, <-chan struct{})        { return true, nil }
func (s *worldInput) Close() error                              { s.closed.Add(1); return nil }
func TestWorldConstructionRollbackAndLifecycle(t *testing.T) {
	for _, stage := range []string{"environment", "random", "clock", "io", "filesystem", "sockets", "http"} {
		t.Run(stage, func(t *testing.T) {
			source := &worldInput{}
			config := WorldConfig{IO: IOConfig{Stdin: source}}
			accepted := false
			switch stage {
			case "environment":
				config.Environment = EnvironmentConfig{Env: []string{"bad\x00=value"}}
			case "random":
				config.Random = RandomConfig{MaxBytes: ^uint64(0)}
			case "clock":
				config.SystemClock = SystemClockConfig{Now: func(context.Context) (Instant, error) { return Instant{}, nil }}
			case "io":
				config.IO.MaxReadBytes = 1 << 28
			case "filesystem":
				accepted = true
				config.FileSystem.Preopens = []FileSystemPreopen{{GuestPath: "/data", HostPath: t.TempDir() + "/missing"}}
			case "sockets":
				accepted = true
				config.Sockets.MaxSockets = 1 << 21
			case "http":
				accepted = true
				config.HTTP.ForbiddenHeaders = []string{"bad name"}
			}
			if w, err := NewWorld(config); w != nil || err == nil {
				t.Fatal(w, err)
			}
			want := uint32(0)
			if accepted {
				want = 1
			}
			if source.closed.Load() != want {
				t.Fatal("source ownership", source.closed.Load(), want)
			}
		})
	}
	var w *World
	w.Close()
	if len(w.Imports().Instances) != 0 || w.IO() != nil || w.HTTP() != nil {
		t.Fatal("nil world")
	}
	if err := w.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if err := new(World).WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	owner := worldOwnerForTest(t, WorldConfig{})
	owner.Close()
	owner.Close()
	if err := owner.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.io.operation(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}

func TestWorldFullPinnedCLIServiceMiddleware(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, name := range []string{"command", "service", "middleware"} {
			w := worldOwnerForTest(t, WorldConfig{})
			in := instantiate(t, compileFixture(t, backend, "world-full-"+name), w.Imports())
			if name == "command" {
				if code, err := RunCommand(nil, in, component.CallOptions{}); code != 0 || err != nil {
					t.Fatal(code, err)
				}
			} else {
				if _, err := in.Function(HTTPHandlerInterface, "handle"); err != nil {
					t.Fatal(name, err)
				}
			}
		}
	})
}

func TestWorldIndependentLanguageNativeAsyncCommands(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, language := range []string{"rust", "c"} {
			t.Run(language, func(t *testing.T) {
				output := NewOutputBuffer(4096)
				w := worldOwnerForTest(t, WorldConfig{Environment: EnvironmentConfig{Args: []string{"wasmvm", "world"}}, IO: IOConfig{Stdout: output, MaxReadBytes: 3, MaxWriteBytes: 3}, FileSystem: FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: "/data", HostPath: t.TempDir()}}}})
				in := instantiate(t, compileFixture(t, backend, "world-"+language+"-command"), w.Imports())
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				for range 3 {
					if code, err := RunCommand(ctx, in, component.CallOptions{Metered: true, Fuel: 10000000}); code != 0 || err != nil {
						t.Fatal(code, err)
					}
				}
				if got := string(output.Drain()); got != strings.Repeat(language+"-p3-command:ok\n", 3) {
					t.Fatal(got)
				}
			})
		}
	})
}

func TestWorldIndependentRustHTTPService(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, fixture := range []string{"world-rust-http", "world-rust-composed-http"} {
			w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxReadBytes: 3, MaxWriteBytes: 3}})
			in := instantiate(t, compileFixture(t, backend, fixture), w.Imports())
			payload := strings.Repeat("rust-http-body", 8)
			for range 2 {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				response, err := w.HTTP().CallIncoming(ctx, in, httptest.NewRequest("POST", "http://localhost/echo", strings.NewReader(payload)), component.CallOptions{Metered: true, Fuel: 10000000})
				if err != nil {
					cancel()
					t.Fatal(err)
				}
				data, err := io.ReadAll(response.Body)
				response.Body.Close()
				cancel()
				if err != nil || string(data) != payload || response.StatusCode != 201 || response.Header.Get("X-Guest") != "rust-p3" || response.Trailer.Get("X-Guest-Trailer") != "done" {
					t.Fatal(string(data), response.StatusCode, response.Header, response.Trailer, err)
				}
			}
		}
	})
}

func TestWorldIndependentRustMiddlewareForwarding(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		w := worldOwnerForTest(t, WorldConfig{IO: IOConfig{MaxReadBytes: 3, MaxWriteBytes: 3}, HTTP: HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 202, Body: r.Body, ContentLength: -1}, nil
		}}})
		in := instantiate(t, compileFixture(t, backend, "world-rust-middleware"), w.Imports())
		payload := strings.Repeat("forward-body", 8)
		for range 2 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			response, err := w.HTTP().CallIncoming(ctx, in, httptest.NewRequest("POST", "http://localhost/", strings.NewReader(payload)), component.CallOptions{Metered: true, Fuel: 10000000})
			if err != nil {
				cancel()
				t.Fatal(err)
			}
			<-response.Body.(*httpCallBody).done
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			cancel()
			if err != nil || response.StatusCode != 202 || string(data) != payload {
				t.Fatal(string(data), response.StatusCode, err)
			}
		}
	})
}
