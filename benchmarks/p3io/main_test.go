package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

// Python zlib.crc32(b"0123456789abcdef") is the independent fixture oracle.
const bodyCRC = 1757737011

func p3Fixture(name string) string {
	return filepath.Join("..", "..", "pkg", "wasi", "preview3", "testdata", name+".wasm")
}

func largeOptions() options {
	args := nativeOptions()
	args.payload = bytes.Repeat([]byte("0123456789abcdef"), 4096)
	// Independent Python zlib oracle for the full64KiB payload.
	args.Expected = 2629362236
	return args
}

func nativeOptions() options {
	return options{engine: "native-go", backend: "host", payload: []byte("0123456789abcdef"), Options: service.Options{Workers: 2, Clients: 3, Requests: 2, Samples: 2, Warmups: 1, Expected: bodyCRC}}
}

func checkBatches(t *testing.T, out report, args options) {
	t.Helper()
	if out.SchemaVersion != 1 || out.Engine != args.engine || out.Backend != args.backend || out.Mode != "p3-http-service" || out.PayloadBytes != len(args.payload) || out.Result != args.Expected || len(out.WarmCallNs) != args.Samples || len(out.WarmupCallNs) != args.Warmups || out.Service == nil || out.Phases["cleanup"] <= 0 {
		t.Fatal(out)
	}
	data := out.Service
	if data.Workers != args.Workers || data.Clients != args.Clients || len(data.Batches) != 1+args.Samples+args.Warmups {
		t.Fatal(data)
	}
	for i, batch := range data.Batches {
		if batch.ReadyClients != args.Clients || len(batch.RequestNs) != args.Clients*args.Requests || len(batch.PublicAPINs) != len(batch.RequestNs) || batch.WallNs == 0 {
			t.Fatal(batch)
		}
		if i > args.Warmups && batch.WallNs != uint64(out.WarmCallNs[i-1-args.Warmups]) {
			t.Fatal("batch timing mismatch")
		}
	}
}

func TestNativeCompleteHTTPAndFailureCleanup(t *testing.T) {
	args := nativeOptions()
	out, err := benchmark(nil, args, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkBatches(t, out, args)
	if _, invented := out.Phases["compileValidate"]; invented {
		t.Fatal(out.Phases)
	}
	args.Resources = true
	if out, err := benchmark(context.Background(), args, nil); err != nil || len(out.Service.ResourceSnapshots) != len(out.Service.Batches)+1 {
		t.Fatal(out, err)
	}
	args = nativeOptions()
	args.Expected++
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("invalid expected body accepted")
	}
	args = nativeOptions()
	args.Workers = 0
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("invalid admission accepted")
	}
	args = nativeOptions()
	if _, err := benchmark(nil, args, func(int) error { return io.ErrClosedPipe }); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := benchmark(ctx, args, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	args.engine, args.backend, args.path = "wasmtime", "compiler", "unused"
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("missing reference accepted")
	}
}

type failingBody struct {
	bytes.Reader
	readErr, closeErr error
	closed            bool
}

func (b *failingBody) Read(data []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	return b.Reader.Read(data)
}
func (b *failingBody) Close() error { b.closed = true; return b.closeErr }

func TestEveryHTTPFailureClosesBodyAndChecksAllMetadata(t *testing.T) {
	args := nativeOptions()
	for _, failure := range []string{"nil", "missing-body", "prior", "read", "close", "status", "header", "trailer", "body", "oversize"} {
		body := &failingBody{Reader: *bytes.NewReader(args.payload)}
		response := &http.Response{StatusCode: 201, Body: body, Header: http.Header{"X-Guest": []string{"rust-p3"}}, Trailer: http.Header{"X-Guest-Trailer": []string{"done"}}}
		var prior error
		switch failure {
		case "nil":
			response = nil
		case "missing-body":
			response.Body = nil
		case "prior":
			prior = io.ErrUnexpectedEOF
			body.readErr = errors.New("must not read after prior failure")
		case "read":
			body.readErr = io.ErrUnexpectedEOF
		case "close":
			body.closeErr = io.ErrClosedPipe
		case "status":
			response.StatusCode = 200
		case "header":
			response.Header = nil
		case "trailer":
			response.Trailer = nil
		case "body":
			body.Reader = *bytes.NewReader([]byte("corrupt"))
		case "oversize":
			body.Reader = *bytes.NewReader(append(append([]byte{}, args.payload...), 0))
		}
		call := httpKernel(context.Background(), args, func(context.Context, *http.Request) (*http.Response, error) { return response, prior })
		if _, _, err := call(); err == nil {
			t.Fatal(failure)
		}
		if response != nil && response.Body != nil && !body.closed {
			t.Fatal("body left open", failure)
		}
	}
}

func TestFlagsPreserveWorkloadAndPolicy(t *testing.T) {
	base := []string{"--engine", "native-go", "--backend", "host", "--payload-bytes", "16", "--expected", "1757737011"}
	if _, err := parse(base, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, extra := range [][]string{{"--payload-bytes", "0"}, {"--payload-bytes", "17"}, {"--payload-bytes", "1048592"}, {"--expected", "4294967296"}, {"--expected", "1"}, {"--wasm", "unused"}, {"--backend", "native"}, {"--engine", "unknown"}, {"--workers", "0"}, {"unexpected"}, {"--invalid"}} {
		if _, err := parse(append(append([]string{}, base...), extra...), io.Discard); err == nil {
			t.Fatal(extra)
		}
	}
	for _, invalid := range [][]string{nil, {"--expected", "1"}, {"--engine", "wasmvm", "--backend", "host", "--wasm", "unused", "--expected", "1"}, {"--engine", "wasmtime", "--backend", "host", "--expected", "1"}} {
		if _, err := parse(invalid, io.Discard); err == nil {
			t.Fatal(invalid)
		}
	}
	for _, engine := range []string{"wasmvm", "wasmtime"} {
		args := append(append([]string{}, base...), "--engine", engine, "--wasm", "unused", "--backend", "native")
		if engine == "wasmtime" {
			args[len(args)-1] = "compiler"
		}
		if _, err := parse(args, io.Discard); err != nil {
			t.Fatal(err)
		}
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestProtocolAndZeroWarmupArrays(t *testing.T) {
	args := []string{"--engine", "native-go", "--backend", "host", "--payload-bytes", "16", "--expected", "1757737011", "--clients", "2", "--requests", "1", "--samples", "1", "--warmups", "0"}
	t.Setenv("WASMVM_INTERLEAVE", "1")
	var output, diagnostic bytes.Buffer
	if code := run(args, strings.NewReader("run\nsample\n"), &output, &diagnostic); code != 0 {
		t.Fatal(code, &diagnostic)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 || lines[0] != `{"ready":true}` || lines[1] != `{"sampleReady":0}` {
		t.Fatal(&output)
	}
	var r map[string]json.RawMessage
	if err := json.Unmarshal([]byte(lines[2]), &r); err != nil {
		t.Fatal(err)
	}
	if string(r["warmupCallNs"]) != "[]" || string(r["warmupResults"]) != "[]" {
		t.Fatal(lines[2])
	}
	for _, in := range []string{"", "wrong\n", "run\n", "run\nwrong\n"} {
		if run(args, strings.NewReader(in), io.Discard, io.Discard) != 1 {
			t.Fatal(in)
		}
	}
	if run(args, strings.NewReader("run\n"), failedWriter{}, io.Discard) != 1 {
		t.Fatal("writer error ignored")
	}
	if run([]string{"--help"}, strings.NewReader("run\n"), io.Discard, io.Discard) != 0 {
		t.Fatal("help failed")
	}
	if run([]string{"--invalid"}, strings.NewReader("run\n"), io.Discard, io.Discard) != 1 {
		t.Fatal("bad flag accepted")
	}
	t.Setenv("WASMVM_INTERLEAVE", "")
	if run(args, strings.NewReader("run\n"), io.Discard, io.Discard) != 0 {
		t.Fatal("ordinary protocol failed")
	}
}
