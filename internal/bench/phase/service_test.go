package phase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestServiceOptionsBoundResourcesAndKeepSerialModesExplicit(t *testing.T) {
	valid := []string{"--wasm", "unused", "--expected", "7", "--mode", "service", "--workers", "4", "--clients", "64", "--requests", "16"}
	if _, err := parseOptions(valid, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := parseOptions([]string{"--wasm", "unused", "--expected", "7", "--service-resources"}, &bytes.Buffer{}); err == nil {
		t.Fatal("resource instrumentation silently changed a serial mode")
	}
	for _, extra := range [][]string{
		{"--workers", "0"}, {"--workers", "33"}, {"--clients", "0"}, {"--clients", "257"},
		{"--requests", "0"}, {"--requests", "1048577"}, {"--requests", "1048576"},
		{"--mode", "api"}, {"--mode", "dispatch"},
	} {
		if _, err := parseOptions(append(append([]string{}, valid...), extra...), &bytes.Buffer{}); err == nil {
			t.Fatal(extra)
		}
	}
}

func TestServiceSharedModuleIndependentStoresAndEveryReply(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		args := optionsFor(filepath.Join(fixtures, "call-control.wasm"), "service")
		args.backend, args.workers, args.clients, args.requests = backend, 3, 7, 5
		out, err := benchmark(context.Background(), args)
		if err != nil {
			t.Fatal(backend, err)
		}
		service := out.Service
		if service.Workers != 3 || service.Clients != 7 || service.RequestsPerClient != 5 || len(service.Batches) != 4 || out.Result != 7 || !strings.Contains(service.AllocationScope, "not peak RSS") {
			t.Fatal(service)
		}
		for i, batch := range service.Batches {
			if len(batch.RequestNs) != 35 || len(batch.PublicAPINs) != 35 || batch.ReadyClients != 7 || batch.ReadinessNs < 0 || (batch.Mallocs == nil) != (runtime.Compiler == "llgo") {
				t.Fatal(i, batch)
			}
			if i >= 2 && batch.WallNs != uint64(out.WarmCallNs[i-2]) {
				t.Fatal("batch wall interval differs from warm record", batch)
			}
		}
		args.expected = 0
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal("corrupt service result accepted")
		}
		args.workers = 0
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal("invalid service capacity accepted")
		}
	}
}

func TestServiceCancellationDrainsAllClientsAndJoinsWorkers(t *testing.T) {
	args := options{mode: "service", workers: 4, clients: 8, requests: 16, samples: 1, expected: 7}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var count, active atomic.Int32
	calls := make([]kernel, args.workers)
	for i := range calls {
		var entered atomic.Bool
		calls[i] = func() (int64, uint32, error) {
			if !entered.CompareAndSwap(false, true) {
				t.Error("one Store invoked concurrently")
			}
			if active.Add(1) > int32(args.workers) {
				t.Error("admission limit exceeded")
			}
			if count.Add(1) == 7 {
				cancel()
			}
			runtime.Gosched()
			active.Add(-1)
			entered.Store(false)
			return 17, 7, ctx.Err()
		}
	}
	var out report
	call, stop := boundedService(calls, args, &out)
	_, _, err := call()
	stop()
	if !errors.Is(err, context.Canceled) || count.Load() != 128 || active.Load() != 0 || len(out.Service.Batches[0].RequestNs) != 128 {
		t.Fatal(err, count.Load(), active.Load(), out.Service)
	}
}

func TestServiceCLIUsesTheExistingInterleavedProtocol(t *testing.T) {
	t.Setenv("WASMVM_INTERLEAVE", "1")
	args := []string{"--wasm", filepath.Join(fixtures, "call-control.wasm"), "--expected", "7", "--seed", "7", "--mode", "service", "--workers", "2", "--clients", "4", "--requests", "3", "--samples", "1", "--warmups", "0"}
	var output, diagnostic bytes.Buffer
	if code := RunCLI(args, strings.NewReader("run\nsample\n"), &output, &diagnostic); code != 0 {
		t.Fatal(code, &diagnostic)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 || lines[1] != `{"sampleReady":0}` {
		t.Fatal(&output)
	}
	var out report
	if err := json.Unmarshal([]byte(lines[2]), &out); err != nil || out.Mode != "service" || out.Service == nil || len(out.Service.Batches) != 2 {
		t.Fatal(err, out)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	config := optionsFor(filepath.Join(fixtures, "call-control.wasm"), "service")
	config.workers, config.clients, config.requests = 2, 4, 3
	if _, err := benchmark(ctx, config); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestServiceResourcesAreExplicitAndUseActualOSCounters(t *testing.T) {
	args := optionsFor(filepath.Join(fixtures, "call-control.wasm"), "service")
	args.workers, args.clients, args.requests, args.resources = 2, 4, 3, true
	out, err := benchmark(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !out.Service.ResourceInstrumented || len(out.Service.ResourceSnapshots) != len(out.Service.Batches)+1 {
		t.Fatal(out.Service)
	}
	for _, snapshot := range out.Service.ResourceSnapshots {
		if snapshot.RSSBytes == 0 || snapshot.PeakRSSBytes == 0 || snapshot.OSThreads <= 0 {
			t.Fatal(snapshot)
		}
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := benchmark(context.Background(), args); err == nil {
		t.Fatal("missing OS collector produced success")
	}
}
