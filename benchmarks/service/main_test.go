package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

func nativeOptions() options {
	return options{engine: "native-c", backend: "native", seed: 17, iterations: 64,
		Options: service.Options{Workers: 3, Clients: 7, Requests: 3, Samples: 2, Warmups: 1, Expected: 3381501185}}
}

func checkBatches(t *testing.T, out report, args options) {
	t.Helper()
	if out.Result != args.Expected || out.Engine != args.engine || out.Mode != "service" || len(out.WarmCallNs) != args.Samples || len(out.WarmupCallNs) != args.Warmups || out.Service == nil {
		t.Fatal(out)
	}
	data := out.Service
	if data.Workers != args.Workers || data.Clients != args.Clients || len(data.Batches) != 1+args.Warmups+args.Samples {
		t.Fatal(data)
	}
	for i, batch := range data.Batches {
		if batch.ReadyClients != args.Clients || len(batch.RequestNs) != args.Clients*args.Requests || len(batch.PublicAPINs) != len(batch.RequestNs) || batch.WallNs == 0 {
			t.Fatal(batch)
		}
		if i > args.Warmups && batch.WallNs != uint64(out.WarmCallNs[i-1-args.Warmups]) {
			t.Fatal("warm sample uses a different wall interval")
		}
	}
}

func TestNativeConcurrentSHAAndEveryReply(t *testing.T) {
	// Independent hashlib oracle: benchmarks/applications/oracle.py.
	for _, v := range []struct{ seed, count, want uint32 }{
		{0, 1, 3091200397}, {17, 64, 3381501185}, {0xffffffff, 3, 2854132502},
	} {
		args := nativeOptions()
		args.seed, args.iterations, args.Expected = v.seed, v.count, v.want
		out, err := benchmark(context.Background(), args, nil)
		if err != nil {
			t.Fatal(err)
		}
		checkBatches(t, out, args)
		if _, invented := out.Phases["compileValidate"]; invented {
			t.Fatal("native C reports a Wasm compilation phase")
		}
		args.Expected++
		if _, err := benchmark(context.Background(), args, nil); err == nil {
			t.Fatal("corrupt SHA accepted")
		}
	}
	args := nativeOptions()
	args.Resources = true
	if out, err := benchmark(context.Background(), args, nil); err != nil || len(out.Service.ResourceSnapshots) != len(out.Service.Batches)+1 {
		t.Fatal(out, err)
	}
	args.Workers = 0
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("invalid admission accepted")
	}
	args = nativeOptions()
	if _, err := benchmark(context.Background(), args, func(int) error { return io.ErrClosedPipe }); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	args.Resources = true
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("missing resource collector accepted")
	}
}

func TestFlagsKeepReferencePoliciesExplicit(t *testing.T) {
	valid := []string{"--engine", "native-c", "--backend", "native", "--expected", "3381501185"}
	if _, err := parse(valid, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, extra := range [][]string{
		{"--engine", "other"}, {"--backend", "compiler"}, {"--termination-checks"}, {"--wasm", "unused"},
		{"--iterations", "65537"}, {"--iterations", "0"}, {"--iterations", "2147483648"},
		{"--seed", "4294967296"}, {"--expected", "4294967296"}, {"--workers", "0"}, {"--clients", "257"},
		{"--samples", "0"}, {"--warmups", "-1"}, {"--requests", "1048576"}, {"unexpected"}, {"--unknown"},
	} {
		if _, err := parse(append(append([]string{}, valid...), extra...), io.Discard); err == nil {
			t.Fatal(extra)
		}
	}
	for _, invalid := range [][]string{nil, {"--expected", "1"}, {"--expected", "1", "--wasm", "unused", "--backend", "native"}} {
		if _, err := parse(invalid, io.Discard); err == nil {
			t.Fatal(invalid)
		}
	}
	if _, err := parse([]string{"--expected", "1", "--wasm", "unused"}, io.Discard); err != nil {
		t.Fatal(err)
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestCLIHandshakeAndInterleavedBatches(t *testing.T) {
	args := []string{"--engine", "native-c", "--backend", "native", "--expected", "3381501185", "--clients", "2", "--requests", "2", "--samples", "1", "--warmups", "0"}
	t.Setenv("WASMVM_INTERLEAVE", "1")
	var out, diagnostic bytes.Buffer
	if code := run(args, strings.NewReader("run\nsample\n"), &out, &diagnostic); code != 0 {
		t.Fatal(code, &diagnostic)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 || lines[0] != `{"ready":true}` || lines[1] != `{"sampleReady":0}` {
		t.Fatal(&out)
	}
	var result report
	if err := json.Unmarshal([]byte(lines[2]), &result); err != nil || result.Result != 3381501185 {
		t.Fatal(err, result)
	}
	if result.WarmupCallNs == nil || result.WarmupResults == nil || len(result.WarmupCallNs) != 0 || len(result.WarmupResults) != 0 {
		t.Fatal("zero warmups must use empty JSON arrays", lines[2])
	}
	for _, input := range []string{"", "wrong\n", "run\n", "run\nwrong\n"} {
		if code := run(args, strings.NewReader(input), io.Discard, io.Discard); code != 1 {
			t.Fatal(input, code)
		}
	}
	if run(args, strings.NewReader("run\n"), failedWriter{}, io.Discard) != 1 {
		t.Fatal("writer failure hidden")
	}
	if run([]string{"--help"}, strings.NewReader("run\n"), io.Discard, io.Discard) != 0 {
		t.Fatal("help failed")
	}
	if run([]string{"--bad"}, strings.NewReader("run\n"), io.Discard, io.Discard) != 1 {
		t.Fatal("invalid flags accepted")
	}
	t.Setenv("WASMVM_INTERLEAVE", "")
	if run(args, strings.NewReader("run\n"), io.Discard, io.Discard) != 0 {
		t.Fatal("noninterleaved protocol failed")
	}
	if result.Service.GOMAXPROCS <= 0 {
		t.Fatal(result.Service)
	}
}
