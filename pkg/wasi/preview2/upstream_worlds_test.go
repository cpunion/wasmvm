//go:build linux || darwin

package preview2_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	w "github.com/cpunion/wasmvm/pkg/wasi/preview2"
	"github.com/cpunion/wasmvm/pkg/wasi/preview3"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type p2WorldPin struct{ SourceSHA256, BinarySHA256 string }
type p2WorldPins struct {
	Revision, Rust, Producer, GeneratorSHA256 string
	Inventory                                 []string
	Cases, RetainedDraftCases                 map[string]p2WorldPin
	Variants                                  map[string]struct{ Original, PatchSHA256, RepairedSourceSHA256, BinarySHA256 string }
	Excluded                                  map[string]string
}

func p2WorldInputs(t *testing.T) (string, p2WorldPins) {
	t.Helper()
	root := os.Getenv("WASMVM_WASI_UPSTREAM_CORPUS")
	if root == "" {
		root = "../../../artifacts/components/wasi-upstream"
	}
	data, err := os.ReadFile("../../../testdata/wasi-preview2/upstream/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins p2WorldPins
	if err := json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	return root, pins
}

func TestUpstreamP2WorldProvenance(t *testing.T) {
	root, pins := p2WorldInputs(t)
	if pins.Revision != "7bac2c2775808aaec5d4aa5627a5e447b51102cf" || pins.Rust != "1.98.0" || pins.Producer != "wasm-tools 1.262.0 (e22db44da 2026-10-13)" || len(pins.Inventory) != 95 || len(pins.Cases) != 92 || len(pins.RetainedDraftCases) != 2 || len(pins.Excluded) != 3 {
		t.Fatal("changed P2 upstream inventory", pins)
	}
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		sum := sha256.Sum256(data)
		if err != nil || hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed P2 input", path, err)
		}
	}
	check("../../../scripts/build_wasi_p2_world_fixtures.py", pins.GeneratorSHA256)
	for _, cases := range []map[string]p2WorldPin{pins.Cases, pins.RetainedDraftCases} {
		for name, pin := range cases {
			check(filepath.Join(root, "source/crates/test-programs/src/bin", name+".rs"), pin.SourceSHA256)
			check(filepath.Join(root, "components", name+".wasm"), pin.BinarySHA256)
		}
	}
	if len(pins.Variants) != 1 {
		t.Fatal("changed executor regression inventory")
	}
	variant := pins.Variants["p2_api_proxy_streaming-executor-lifetimes"]
	if variant.Original != "p2_api_proxy_streaming" || variant.RepairedSourceSHA256 == "" {
		t.Fatal("missing explicit executor provenance")
	}
	check(filepath.Join(root, "patches/p2-streaming-executor.patch"), variant.PatchSHA256)
	check(filepath.Join(root, "components/p2_api_proxy_streaming-executor-lifetimes.wasm"), variant.BinarySHA256)
}

func TestUpstreamP2WorldConsumers(t *testing.T) {
	root, pins := p2WorldInputs(t)
	names := make([]string, 0, len(pins.Cases))
	for name := range pins.Cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, name := range names {
				t.Run(name, func(t *testing.T) { testP2World(t, root, name, backend, "") })
			}
		})
	}
}

func TestUpstreamP2WorldInteractions(t *testing.T) {
	root, _ := p2WorldInputs(t)
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			for _, mode := range []string{"p2-append-empty", "p2-append-same", "p2-append-same-empty", "p3-append", "p3-append-empty", "p3-append-same", "p3-append-same-empty"} {
				t.Run(mode, func(t *testing.T) { testP2World(t, root, "p2_cli_http_headers", backend, mode) })
			}
			for _, mode := range []string{"double-echo", "double-echo-large", "hash-all"} {
				t.Run(mode, func(t *testing.T) { testP2World(t, root, "p2_api_proxy_streaming", backend, mode) })
			}
			for _, mode := range []string{"modify_fields/30", "new_fields/30"} {
				t.Run(mode, func(t *testing.T) { testP2World(t, root, "p2_api_proxy", backend, mode) })
			}
		})
	}
}

type p2DirectoryInput struct{}

func (p2DirectoryInput) Read(context.Context, []byte) (int, error) { return 0, unix.EISDIR }
func (p2DirectoryInput) Readiness() (bool, <-chan struct{})        { return true, nil }
func (p2DirectoryInput) Close() error                              { return nil }

type p2BrokenOutput struct{}

func (p2BrokenOutput) Write(context.Context, []byte) (int, error) { return 0, unix.EPIPE }
func (p2BrokenOutput) Flush(context.Context) error                { return nil }
func (p2BrokenOutput) Readiness() (bool, <-chan struct{})         { return true, nil }
func (p2BrokenOutput) Close() error                               { return nil }

// Inventory consumers and assertions are unchanged. The separately named large
// stress variant repairs its sample executor's stale pollables. Host inputs
// follow the pinned harness with hermetic filesystem, DNS and HTTP services.
func testP2World(t *testing.T, root, name, backend, mode string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dir, readonly := t.TempDir(), t.TempDir()
	write := func(path string, data []byte) {
		t.Helper()
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(dir, "bar.txt"), []byte("And stood awhile in thought"))
	write(filepath.Join(dir, "1mb"), make([]byte, 1<<20))
	write(filepath.Join(readonly, "test.txt"), []byte("read only test file\n"))
	write(filepath.Join(readonly, "stream-perms.txt"), []byte("stream permission test\n"))
	if name == "p2_cli_directory_list" {
		if err := os.Remove(filepath.Join(dir, "1mb")); err != nil {
			t.Fatal(err)
		}
		for _, filename := range []string{"foo.txt", "baz.txt"} {
			write(filepath.Join(dir, filename), nil)
		}
		if err := os.Mkdir(filepath.Join(dir, "sub"), 0700); err != nil {
			t.Fatal(err)
		}
		for _, filename := range []string{"wow.txt", "yay.txt"} {
			write(filepath.Join(dir, "sub", filename), nil)
		}
	}
	output, diagnostic := w.NewOutputBuffer(16<<20), w.NewOutputBuffer(1<<20)
	defer func() {
		if t.Failed() {
			for label, b := range map[string]*w.OutputBuffer{"stdout": output, "stderr": diagnostic} {
				data := b.Drain()
				if len(data) > 4096 {
					data = data[:4096]
				}
				t.Logf("guest %s: %s", label, data)
			}
		}
	}()
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if name == "p2_http_outbound_request_invalid_version" {
			// Go accepts the CONNECT form rejected by the reference server.
			// Preserve the protocol-failure assertion using a malformed line.
			conn, _, err := resp.(http.Hijacker).Hijack()
			if err == nil {
				_, _ = fmt.Fprint(conn, "HTTP/invalid 200 OK\r\n\r\n")
				_ = conn.Close()
			}
			return
		}
		resp.Header().Set("x-wasmtime-test-method", req.Method)
		resp.Header().Set("x-wasmtime-test-uri", req.URL.RequestURI())
		if name == "p2_api_proxy_streaming" && strings.HasPrefix(mode, "double-echo") {
			// Like the reference Hyper server, publish headers before reading
			// the request body so both directions can make progress together.
			if err := http.NewResponseController(resp).EnableFullDuplex(); err != nil {
				t.Error(err)
				return
			}
			resp.WriteHeader(http.StatusOK)
			resp.(http.Flusher).Flush()
			_, _ = io.Copy(resp, req.Body)
			return
		}
		data, err := io.ReadAll(io.LimitReader(req.Body, 8<<20))
		if err == nil {
			_, _ = resp.Write(data)
		}
	}))
	defer server.Close()
	allow := func(context.Context, w.SocketAddress) error { return nil }
	resolve := func(_ context.Context, name string) ([]netip.Addr, error) {
		switch name {
		case "localhost", "example.com", "api.github.com", "docs.wasmtime.dev", "bytecodealliance.org", "www.rust-lang.org", "xn--mnchen-3ya.de":
			return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
		}
		return nil, &net.DNSError{Err: "not found", Name: name, IsNotFound: true}
	}
	args := []string{"wasmvm"}
	switch name {
	case "p2_cli_args":
		args = append(args, "hello", "this", "", "is an argument", "with 🚩 emoji")
	case "p2_cli_argv0":
		args = nil
	case "p2_cli_much_stdout":
		args = append(args, "stream", "1000")
	case "p2_cli_exit_with_code":
		args = append(args, "42")
	case "p2_cli_hostcall_fuel":
		args = append(args, "read")
	case "p2_cli_http_headers":
		if mode == "" {
			mode = "p2-append"
		}
		args = append(args, mode)
	}
	if strings.Contains(name, "across_perms") {
		args = append(args, "/")
	}
	cwd := "/"
	if name == "p2_cli_initial_cwd" {
		cwd = "/sandbox"
	}
	input := []byte("So rested he by the Tumtum tree")
	if name == "p2_cli_stdin_empty" || name == "p2_cli_stdio_write_flushes" {
		input = nil
	}
	cfg := w.WorldConfig{
		Environment: w.EnvironmentConfig{Args: args, Env: []string{"frabjous=day", "callooh=callay", "HTTP_SERVER=" + strings.TrimPrefix(server.URL, "http://")}, InitialCWD: &cwd},
		IO:          w.IOConfig{Stdin: w.NewBytesInput(input), Stdout: output, Stderr: diagnostic},
		FileSystem:  w.FileSystemConfig{Preopens: []w.FileSystemPreopen{{GuestPath: "/", HostPath: dir, ReadOnly: name == "p2_api_read_only"}}},
		Network:     w.NetworkConfig{AllowNameLookup: true, Resolve: resolve},
		TCP:         w.TCPConfig{AllowBind: allow, AllowConnect: allow},
		UDP:         w.UDPConfig{AllowBind: allow, AllowSend: allow, AllowReceive: allow},
		HTTP:        w.HTTPConfig{AllowRequest: func(context.Context, *url.URL) error { return nil }, AllowConnect: allow, ForbiddenHeaders: []string{"custom-forbidden-header"}},
	}
	if strings.Contains(name, "across_perms") || name == "p2_file_truncation_readonly" || name == "p2_file_stream_not_permitted" {
		cfg.FileSystem.Preopens = append(cfg.FileSystem.Preopens, w.FileSystemPreopen{GuestPath: "readonly", HostPath: readonly, ReadOnly: true})
	}
	switch name {
	case "p2_api_proxy", "p2_api_proxy_forward_request", "p2_api_proxy_streaming":
		cfg.FileSystem.Preopens = nil
	case "p2_api_reactor":
		cfg.Environment.Env = []string{"GOOD_DOG=gussie"}
	case "p2_cli_env":
		cfg.Environment.Env = cfg.Environment.Env[:2]
	case "p2_cli_multiple_preopens":
		cfg.FileSystem.Preopens = []w.FileSystemPreopen{{GuestPath: "/a", HostPath: dir}, {GuestPath: "/b", HostPath: dir}, {GuestPath: "/c", HostPath: dir}}
	case "p2_cli_stdin_eisdir":
		cfg.IO.Stdin = p2DirectoryInput{}
	case "p2_cli_stdout_epipe":
		cfg.IO.Stdout = p2BrokenOutput{}
	case "p2_cli_no_tcp":
		cfg.TCP = w.TCPConfig{DenyCreate: true}
	case "p2_cli_no_udp":
		cfg.UDP = w.UDPConfig{DenyCreate: true}
	case "p2_cli_no_ip_name_lookup":
		cfg.Network.Resolve = func(context.Context, string) ([]netip.Addr, error) {
			return nil, &w.NetworkError{Code: w.NetworkPermanentResolverFailure}
		}
	case "p2_api_time":
		cfg.WallClock = w.WallClockConfig{
			Now: func(context.Context) (w.Datetime, error) {
				return w.Datetime{Seconds: 1431648000, Nanoseconds: 100}, nil
			},
			Resolution: func(context.Context) (w.Datetime, error) { return w.Datetime{Seconds: 1}, nil },
		}
	}
	if mode == "double-echo-large" {
		cfg.HTTP.BodyBufferBytes = 4096
	}
	world, err := w.NewWorld(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		world.Close()
		if err := world.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	}()
	imports := world.Imports()
	if name == "p2_cli_http_headers" {
		p3, err := preview3.NewWorld(preview3.WorldConfig{IO: preview3.IOConfig{Stdout: output, Stderr: diagnostic}})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			p3.Close()
			if err := p3.WaitClosed(nil); err != nil {
				t.Error(err)
			}
		}()
		for name, instance := range p3.Imports().Instances {
			imports.Instances[name] = instance
		}
	}
	if name == "p2_api_time" {
		var tick uint64
		clock := imports.Instances[w.MonotonicClockInterface]
		clock.Functions["now"] = func(*component.Caller, []component.Value) (*component.Value, error) {
			now := tick
			tick += 42e9
			return &component.Value{Bits: now}, nil
		}
		clock.Functions["resolution"] = func(*component.Caller, []component.Value) (*component.Value, error) {
			return &component.Value{Bits: 1e9}, nil
		}
	}
	var waiting atomic.Bool
	if name == "p2_cli_sleep_forever" || name == "p2_cli_serve_sleep" {
		poll := imports.Instances[w.PollInterface]
		for _, key := range []string{"poll", "[method]pollable.block"} {
			original := poll.Functions[key]
			poll.Functions[key] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
				waiting.Store(true)
				cancel()
				return original(c, args)
			}
		}
	}
	var reactorOutput component.Value
	if name == "p2_api_reactor" {
		env := imports.Instances[w.EnvironmentInterface]
		original := env.Functions["get-environment"]
		env.Functions["get-environment"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
			if reactorOutput.Resource == nil {
				var err error
				reactorOutput, err = world.IO().NewOutput(c, output)
				if err != nil {
					return nil, err
				}
			}
			return original(c, args)
		}
	}
	config := component.DefaultConfig()
	config.Core.Backend = backend
	if name == "p2_cli_max_resources" || name == "p2_cli_many_resources" {
		config.MaxResources = 100
	}
	engine, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	binary := name + ".wasm"
	if mode == "double-echo-large" {
		binary = name + "-executor-lifetimes.wasm"
	}
	data, err := os.ReadFile(filepath.Join(root, "components", binary))
	if err != nil {
		t.Fatal(err)
	}
	module, err := engine.Compile(ctx, data)
	if err != nil {
		t.Fatal("compile", err)
	}
	defer module.Close()
	in, err := module.Instantiate(ctx, imports, component.CallOptions{})
	if err != nil {
		t.Fatal("instantiate", err)
	}
	defer func() {
		in.Close()
		if err := in.WaitClosed(nil); err != nil {
			t.Error(err)
		}
	}()
	options := component.CallOptions{Metered: true, Fuel: 500000000}
	if name == "p2_api_reactor" {
		testP2Reactor(t, ctx, in, &reactorOutput, output, options)
		return
	}
	wit, err := os.ReadFile(filepath.Join(root, "components", name+".wit"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(wit), "export wasi:http/incoming-handler@0.2.") {
		testP2Incoming(t, ctx, world, in, name, mode, server.URL, &waiting, options)
		return
	}
	code, err := w.RunCommand(ctx, in, options)
	switch name {
	case "p2_cli_exit_panic":
		if !errors.Is(err, wasm.ErrTrap) || !strings.Contains(string(diagnostic.Drain()), "Curiouser and curiouser!") {
			t.Fatal("wrong panic", code, err)
		}
	case "p2_cli_many_resources":
		if !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("resource quota did not stop guest", err)
		}
	case "p2_pollable_traps", "p2_stream_pollable_traps", "p2_udp_send_too_much":
		if !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("wrong protocol violation", err)
		}
	case "p2_cli_sleep_forever":
		if !waiting.Load() || !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation did not reach guest wait", err)
		}
	case "p2_cli_http_headers":
		if strings.HasPrefix(mode, "p2-") {
			if !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("P2 fields did not enforce quota", err)
			}
		} else if err != nil || code != 0 || !strings.Contains(string(output.Drain()), "error received") {
			t.Fatal("P3 fields lost recoverable error", code, err)
		}
	default:
		want := uint32(0)
		if name == "p2_cli_exit_failure" {
			want = 1
		}
		if name == "p2_cli_exit_with_code" {
			want = 42
		}
		if err != nil || code != want {
			t.Fatal(code, err)
		}
	}
	if name == "p2_cli_hello_stdout" && (string(output.Drain()) != "hello, world\n" || string(diagnostic.Drain()) != "hello, world\n") {
		t.Fatal("lost guest stdio")
	}
	if name == "p2_cli_splice_stdin" {
		want := fmt.Sprintf("before splice\n%s\ncompleted splicing %d bytes\n", input, len(input))
		if got := string(output.Drain()); got != want {
			t.Fatal("splice lost bytes", got)
		}
	}
}

func testP2Reactor(t *testing.T, ctx context.Context, in *component.Instance, output *component.Value, buffer *w.OutputBuffer, options component.CallOptions) {
	t.Helper()
	call := func(name string, args ...component.Value) *component.Value {
		t.Helper()
		f, err := in.Function(name)
		if err != nil {
			t.Fatal(err)
		}
		r, err := f.Call(ctx, args, options)
		if err != nil {
			t.Fatal(name, err)
		}
		return r.Value
	}
	if result := call("add-strings", component.Value{Elements: []component.Value{{Text: "hello"}, {Text: "$GOOD_DOG"}}}); result.Bits != 2 {
		t.Fatal(result)
	}
	result := call("get-strings")
	if len(result.Elements) != 2 || result.Elements[0].Text != "hello" || result.Elements[1].Text != "gussie" {
		t.Fatal(result)
	}
	if output.Resource == nil {
		t.Fatal("reactor did not use WASI environment")
	}
	if result := call("write-strings-to", *output); result.Case != 0 || string(buffer.Drain()) != "hellogussie" {
		t.Fatal("reactor lost imported stream", result)
	}
	datetime := func(seconds, nanos uint64) component.Value {
		return component.Value{Case: 1, Elements: []component.Value{{Elements: []component.Value{{Bits: seconds}, {Bits: nanos}}}}}
	}
	stat := component.Value{Elements: []component.Value{{}, {}, {}, datetime(45, 123), datetime(10, 789), datetime(1, 0)}}
	want := "DescriptorStat { type: DescriptorType::Unknown, link-count: 0, size: 0, data-access-timestamp: Some(Datetime { seconds: 45, nanoseconds: 123 }), data-modification-timestamp: Some(Datetime { seconds: 10, nanoseconds: 789 }), status-change-timestamp: Some(Datetime { seconds: 1, nanoseconds: 0 }) }"
	if result := call("pass-an-imported-record", stat); result.Text != want {
		t.Fatal(result.Text)
	}
}

func testP2Incoming(t *testing.T, ctx context.Context, world *w.World, in *component.Instance, name, mode, target string, waiting *atomic.Bool, options component.CallOptions) {
	t.Helper()
	payload := []byte("hello, world!")
	if mode == "double-echo-large" {
		payload = bytes.Repeat([]byte("bounded duplex body\n"), 8192)
	}
	method, path := http.MethodPost, "/test"
	if name == "p2_api_proxy_streaming" {
		if mode == "" {
			mode = "echo"
		}
		path = "/" + mode
		if mode == "double-echo-large" {
			path = "/double-echo"
		}
		if mode == "hash-all" {
			method = http.MethodGet
		}
	}
	if name == "p2_api_proxy" && mode != "" {
		method = http.MethodGet
		path = "/" + mode
	}
	req := httptest.NewRequest(method, target+path, bytes.NewReader(payload))
	req.Header.Set("url", target+"/test")
	if name == "p2_cli_serve_authority_and_scheme" {
		req = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	}
	if name == "p2_cli_serve_echo_env" {
		req.Header.Set("env", "frabjous")
	}
	response, err := world.HTTP().CallIncoming(ctx, in, req, options)
	switch name {
	case "p2_cli_serve_sleep":
		if !waiting.Load() || !errors.Is(err, context.Canceled) {
			t.Fatal("service cancellation missed guest wait", err)
		}
		return
	case "p2_cli_serve_trap_before_set", "p2_cli_serve_return_before_set":
		if err == nil || err.Error() != "HTTP-response-incomplete" {
			t.Fatal("service lost incomplete response", err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil || response.StatusCode != 200 {
		t.Fatal(response.StatusCode, len(body), err)
	}
	switch name {
	case "p2_api_proxy":
		if mode == "" && !bytes.Equal(body, payload) {
			t.Fatal("proxy response", string(body))
		}
	case "p2_api_proxy_streaming":
		if mode == "hash-all" {
			sum := sha256.Sum256(nil)
			want := target + "/test: " + base64.RawStdEncoding.EncodeToString(sum[:]) + "\n"
			if string(body) != want {
				t.Fatal("hash fanout", string(body), want)
			}
		} else if !bytes.Equal(body, payload) {
			t.Fatal("streaming proxy lost bytes", len(body))
		}
	case "p2_cli_serve_echo_env":
		if response.Header.Get("env") != "day" {
			t.Fatal("service environment", response.Header)
		}
	}
}
