//go:build linux || darwin

package preview3

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasi/preview2"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type upstreamWASIPins struct {
	Revision, Version, Rust, Bindings, Producer, GeneratorSHA256 string
	Inventory                                                    []string
	Excluded                                                     map[string]string
	Cases                                                        map[string]struct{ SourceSHA256, BinarySHA256, ComposedSHA256 string }
}

func upstreamWASIInputs(t *testing.T) (string, upstreamWASIPins) {
	t.Helper()
	root := os.Getenv("WASMVM_WASI_UPSTREAM_CORPUS")
	if root == "" {
		root = "../../../artifacts/components/wasi-upstream"
	}
	data, err := os.ReadFile("../../../testdata/wasi-preview3/upstream/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins upstreamWASIPins
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	return root, pins
}

func TestUpstreamWASIProvenance(t *testing.T) {
	root, pins := upstreamWASIInputs(t)
	if pins.Revision != "7bac2c2775808aaec5d4aa5627a5e447b51102cf" || pins.Version != "48.0.1" || pins.Rust != "1.98.0" || pins.Bindings != "wit-bindgen 0.60.0" || pins.Producer != "wasm-tools 1.262.0 (e22db44da 2026-10-13)" || len(pins.Inventory) != 61 || len(pins.Cases) != 60 || len(pins.Excluded) != 1 || pins.Excluded["p3_tls_sample_application"] == "" {
		t.Fatal("changed upstream WASI profile", pins)
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		sum := sha256.Sum256(data)
		if err != nil || hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed upstream WASI asset", path, err)
		}
	}
	check("../../../scripts/build_wasi_upstream_fixtures.py", pins.GeneratorSHA256)
	for name, pin := range pins.Cases {
		check(filepath.Join(root, "source/crates/test-programs/src/bin", name+".rs"), pin.SourceSHA256)
		check(filepath.Join(root, "components", name+".wasm"), pin.BinarySHA256)
		if pin.ComposedSHA256 != "" {
			check(filepath.Join(root, "components", name+"-composed.wasm"), pin.ComposedSHA256)
		}
	}
}

// All guest sources and assertions are unchanged. Host inputs are hermetic:
// loopback networking, fixed DNS, explicit preopens, and an HTTP echo server.
func TestUpstreamWASIConsumers(t *testing.T) {
	root, pins := upstreamWASIInputs(t)
	names := make([]string, 0, len(pins.Cases))
	for name := range pins.Cases {
		names = append(names, name)
	}
	sort.Strings(names)
	backends(t, func(t *testing.T, backend string) {
		for _, name := range names {
			t.Run(name, func(t *testing.T) { testUpstreamWASIConsumer(t, root, name, backend) })
		}
	})
}

func TestUpstreamWASIStreamingInteractions(t *testing.T) {
	root, _ := upstreamWASIInputs(t)
	backends(t, func(t *testing.T, backend string) {
		for _, name := range []string{"p3_http_echo", "p3_http_middleware", "p3_http_middleware_with_chain"} {
			for _, mode := range []string{"trailers", "host-to-host", "compressed"} {
				if mode == "compressed" && name != "p3_http_middleware" {
					continue
				}
				t.Run(name+"/"+mode, func(t *testing.T) { testUpstreamWASIConsumer(t, root, name, backend, mode) })
			}
		}
	})
}

func testUpstreamWASIConsumer(t *testing.T, root, name, backend string, scenario ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	output, diagnostic := NewOutputBuffer(16<<20), NewOutputBuffer(1<<20)
	defer func() {
		if t.Failed() {
			t.Logf("guest stdout: %s\nguest stderr: %s", output.Drain(), diagnostic.Drain())
		}
	}()
	// The invalid-version consumer requires a protocol failure, whereas Go's
	// HTTP server accepts CONNECT. Supply an explicitly malformed response line.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if name == "p3_http_outbound_request_invalid_version" {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				_, _ = fmt.Fprint(conn, "HTTP/invalid 200 OK\r\n\r\n")
				_ = conn.Close()
			}
			return
		}
		w.Header().Set("x-wasmtime-test-method", r.Method)
		w.Header().Set("x-wasmtime-test-uri", r.URL.RequestURI())
		data, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
		if err != nil {
			return
		} // Deliberately short request bodies are negative tests.
		_, _ = w.Write(data)
	}))
	defer server.Close()
	allow := func(context.Context, SocketAddress) error { return nil }
	resolve := func(_ context.Context, name string) ([]netip.Addr, error) {
		switch name {
		case "localhost", "example.com", "api.github.com", "docs.wasmtime.dev", "bytecodealliance.org", "www.rust-lang.org", "xn--mnchen-3ya.de":
			return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
		}
		return nil, &net.DNSError{Err: "not found", Name: name, IsNotFound: true}
	}
	args := []string{name + ".component", "."}
	if name == "p3_cli_much_stdout" {
		args = []string{name, "stream", "1000"}
	}
	if name == "p3_cli_random_limits" {
		args = []string{name, "random", "4096"}
	}
	var input []byte
	if name == "p3_cli_read_stdin" || name == "p3_cli_stdout_flush" {
		input = []byte("hello!")
	}
	env := EnvironmentConfig{Args: args, Env: []string{"WASMVM_TEST=1", "HTTP_SERVER=" + strings.TrimPrefix(server.URL, "http://")}}
	dir, readonly := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(readonly, "test.txt"), []byte("read only test file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := WorldConfig{Environment: env, IO: IOConfig{Stdin: NewBytesInput(input), Stdout: output, Stderr: diagnostic}, FileSystem: FileSystemConfig{Preopens: []FileSystemPreopen{{GuestPath: ".", HostPath: dir}}}, Sockets: SocketsConfig{AllowBind: allow, AllowConnect: allow, AllowSend: allow, AllowReceive: allow, AllowNameLookup: true, Resolve: resolve}, HTTP: HTTPConfig{AllowRequest: func(context.Context, *url.URL) error { return nil }, AllowConnect: allow, ForbiddenHeaders: []string{"custom-forbidden-header"}}}
	if name == "p3_sockets_tcp_busy_poll" {
		// Sequential connections must reclaim their physical sockets without
		// depending on a separately scheduled cleanup worker.
		cfg.Sockets.MaxSockets = 8
		previous := runtime.GOMAXPROCS(1)
		defer runtime.GOMAXPROCS(previous)
	}
	if name == "p3_cli_deny_listen" {
		cfg.Sockets = SocketsConfig{}
	}
	if strings.Contains(name, "across_perms") || name == "p3_file_truncation_readonly" {
		cfg.FileSystem.Preopens = append(cfg.FileSystem.Preopens, FileSystemPreopen{GuestPath: "readonly", HostPath: readonly, ReadOnly: true})
	}
	var intercepted atomic.Bool
	cfg.HTTP.Handler = func(_ context.Context, r *http.Request) (*http.Response, error) {
		if name == "p3_http_forbidden_headers" {
			if r.URL.Host != "p3-test" || r.Header.Get("connection") != "" || r.Header.Get("custom-forbidden-header") != "" || r.Header.Get("host") != "" || r.Header.Get("foo") != "bar" {
				return nil, errors.New("forbidden inbound header leaked")
			}
			intercepted.Store(true)
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), ContentLength: int64(len(data)), Trailer: r.Trailer, Body: io.NopCloser(bytes.NewReader(data))}, nil
	}
	w2, err := preview2.NewWorld(preview2.WorldConfig{Environment: env, IO: preview2.IOConfig{Stdin: NewBytesInput(input), Stdout: output, Stderr: diagnostic}, FileSystem: preview2.FileSystemConfig{Preopens: []preview2.FileSystemPreopen{{GuestPath: ".", HostPath: dir}}}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		w2.Close()
		if err := w2.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	}()
	w3, err := NewWorld(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		w3.Close()
		if err := w3.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	}()
	imports := w2.Imports()
	for key, instance := range w3.Imports().Instances {
		imports.Instances[key] = instance
	}
	if name == "p3_http_forbidden_headers" {
		// Match upstream TestHooks' explicit outbound interceptor. The same WIT
		// signature permits an embedding-supplied handler as the client transport.
		imports.Instances[HTTPClientInterface].AsyncCompletions["send"] = imports.Instances[HTTPHandlerInterface].AsyncCompletions["handle"]
	}
	config := component.DefaultConfig()
	config.Core.Backend = backend
	if name == "p3_cli_many_tasks" || name == "p3_cli_many_stream" {
		config.MaxResources = 100
	}
	engine, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	binary := name + ".wasm"
	if name == "p3_http_middleware_with_chain" {
		binary = name + "-composed.wasm"
	}
	data, err := os.ReadFile(filepath.Join(root, "components", binary))
	if err != nil {
		t.Fatal(err)
	}
	module, err := engine.Compile(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	in, err := module.Instantiate(ctx, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		in.Close()
		if err := in.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	}()
	options := component.CallOptions{Metered: true, Fuel: 500000000}
	wit, err := os.ReadFile(filepath.Join(root, "components", name+".wit"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wit), "export wasi:http/handler@0.3.") {
		code, err := RunCommand(ctx, in, options)
		if name == "p3_cli_many_tasks" {
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("missing expected resource limit", code, err)
			}
			return
		}
		if err != nil || code != 0 {
			t.Fatal(code, err)
		}
		want := map[string]string{"p3_cli": "hello stdout\n", "p3_cli_hello_stdout": "hello, world\n", "p3_cli_hello_stdout_post_return": "hello, world\nhello again, after return\n", "p3_cli_stdout_flush": "READY"}[name]
		if want != "" && string(output.Drain()) != want {
			t.Fatal("incorrect or missing command output", want)
		}
		return
	}
	if name == "p3_cli_serve_sleep" {
		// Cancel only after the guest has actually entered its indefinite clock
		// wait. A timeout during compilation/instantiation cannot satisfy this test.
		var clock *MonotonicClock
		for _, owner := range w3.owners {
			if c, ok := owner.(*MonotonicClock); ok {
				clock = c
			}
		}
		if clock == nil {
			t.Fatal("missing clock")
		}
		done := make(chan struct{})
		var observed atomic.Bool
		go func() {
			defer close(done)
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for {
				clock.mu.Lock()
				active := clock.active
				clock.mu.Unlock()
				if active != 0 {
					observed.Store(true)
					cancel()
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
		_, err := w3.HTTP().CallIncoming(ctx, in, httptest.NewRequest("GET", server.URL, nil), options)
		cancel()
		<-done
		if !observed.Load() || !errors.Is(err, context.Canceled) {
			t.Fatal("guest wait was not cancelled", err)
		}
		return
	}
	payload := "hello, world!"
	raw := []byte(payload)
	mode := ""
	if len(scenario) != 0 {
		mode = scenario[0]
		payload = strings.Repeat("forwarded body, with trailers; ", 4096)
		raw = []byte(payload)
	}
	if mode == "compressed" {
		var compressed bytes.Buffer
		writer, err := flate.NewWriter(&compressed, flate.BestSpeed)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(raw); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		raw = compressed.Bytes()
	}
	request := httptest.NewRequest("POST", server.URL+"/test", bytes.NewReader(raw))
	if mode != "" {
		request.Trailer = http.Header{"X-Tail": []string{"complete"}}
	}
	if mode == "host-to-host" {
		request.Header.Set("x-host-to-host", "true")
	}
	if mode == "compressed" {
		request.Header.Set("content-encoding", "deflate")
		request.Header.Set("accept-encoding", "deflate")
	}
	request.Header.Set("url", server.URL+"/test")
	request.Header.Set("foo", "bar")
	if name == "p3_http_forbidden_headers" {
		request = httptest.NewRequest("GET", server.URL, nil)
		request.Header = http.Header{"Host": []string{"victim.internal"}, "Connection": []string{"close"}, "Custom-Forbidden-Header": []string{"hidden"}, "Foo": []string{"bar"}}
	}
	response, err := w3.HTTP().CallIncoming(ctx, in, request, options)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err == nil && mode == "compressed" {
		if response.Header.Get("content-encoding") != "deflate" {
			t.Fatal("missing encoded response", response.Header)
		}
		reader := flate.NewReader(bytes.NewReader(body))
		body, err = io.ReadAll(reader)
		_ = reader.Close()
	}
	closeErr := response.Body.Close()
	if err != nil || closeErr != nil || response.StatusCode != 200 {
		t.Fatal(response.StatusCode, err, closeErr)
	}
	want := payload
	switch name {
	case "p3_cli_serve_hello_world":
		want = "Hello, WASI!"
	case "p3_cli_serve_post_return", "p3_http_forbidden_headers":
		want = ""
	}
	if mode != "" && response.Trailer.Get("x-tail") != "complete" {
		t.Fatal("lost forwarded trailer", response.Trailer)
	}
	if string(body) != want {
		t.Fatalf("service body %q, want %q", body, want)
	}
	if name == "p3_cli_serve_post_return" && string(output.Drain()) != "please see me\n" {
		t.Fatal("post-return task was not drained")
	}
	if name == "p3_http_forbidden_headers" && !intercepted.Load() {
		t.Fatal("outbound interceptor was not called")
	}
}
