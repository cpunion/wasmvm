package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func TestEvaluation(t *testing.T) {
	results, err := Evaluate(context.Background(), Corpus)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 13 {
		t.Fatalf("got %d observations, want 13", len(results))
	}
	supported, unsupported := 0, 0
	for _, result := range results {
		if !result.MatchesExpected {
			t.Errorf("unexpected observation: %+v", result)
		}
		switch result.Outcome {
		case "supported":
			supported++
		case "unsupported":
			unsupported++
		}
	}
	if supported != 7 || unsupported != 6 {
		t.Errorf("supported=%d unsupported=%d; want 7 and 6", supported, unsupported)
	}
	encoded, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("EVALUATION_REPORT %s", encoded)
}

func copyCorpus(t *testing.T) fstest.MapFS {
	t.Helper()
	result := fstest.MapFS{}
	err := fs.WalkDir(Corpus, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := Corpus.ReadFile(name)
		if err == nil {
			result[name] = &fstest.MapFile{Data: data}
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestManifestRejectsInvalidInput(t *testing.T) {
	mutations := map[string]func(*Manifest, fstest.MapFS){
		"schema":                func(m *Manifest, _ fstest.MapFS) { m.SchemaVersion = 2 },
		"origin":                func(m *Manifest, _ fstest.MapFS) { m.Origin = "" },
		"empty":                 func(m *Manifest, _ fstest.MapFS) { m.Fixtures = nil },
		"candidate-version":     func(m *Manifest, _ fstest.MapFS) { m.Candidate.Version = "main" },
		"candidate-module":      func(m *Manifest, _ fstest.MapFS) { m.Candidate.Module = "other" },
		"candidate-commit":      func(m *Manifest, _ fstest.MapFS) { m.Candidate.Commit = "unfixed" },
		"candidate-backend":     func(m *Manifest, _ fstest.MapFS) { m.Candidate.Backend = "compiler" },
		"feature-configuration": func(m *Manifest, _ fstest.MapFS) { m.Candidate.FeatureConfiguration = "" },
		"producer":              func(m *Manifest, _ fstest.MapFS) { m.Producer.Version = "latest" },
		"duplicate-id":          func(m *Manifest, _ fstest.MapFS) { m.Fixtures[1].ID = m.Fixtures[0].ID },
		"missing-id":            func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].ID = "" },
		"feature-scope":         func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].Features = nil },
		"notes":                 func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].Notes = "" },
		"wasm-path":             func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].Wasm = "../core.wasm" },
		"wat-path":              func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].WAT = "/core.wat" },
		"missing-wat":           func(m *Manifest, f fstest.MapFS) { delete(f, "testdata/"+m.Fixtures[0].WAT) },
		"missing-wasm":          func(m *Manifest, f fstest.MapFS) { delete(f, "testdata/"+m.Fixtures[0].Wasm) },
		"corrupt-wasm":          func(m *Manifest, f fstest.MapFS) { f["testdata/"+m.Fixtures[0].Wasm].Data = []byte("corrupt") },
		"invalid-hash":          func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].SHA256 = "not-hex" },
		"corrupt-wat":           func(m *Manifest, f fstest.MapFS) { f["testdata/"+m.Fixtures[0].WAT].Data = []byte("(module)") },
		"invalid-wat-hash":      func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].WATSHA256 = "not-hex" },
		"invalid-expectation":   func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].Expected = Expectation{"execute", "unsupported"} },
		"missing-export":        func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].Calls[0].Export = "" },
		"contradictory-trap":    func(m *Manifest, _ fstest.MapFS) { m.Fixtures[0].Calls[0].Trap = true },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			files := copyCorpus(t)
			manifest, err := LoadManifest(files)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&manifest, files)
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			files["testdata/manifest.json"].Data = data
			if _, err = LoadManifest(files); err == nil {
				t.Fatal("invalid manifest was accepted")
			}
		})
	}
	for name, data := range map[string][]byte{"syntax": []byte("{"), "unknown-field": []byte(`{"unknown":true}`), "extra-value": []byte(`{} {}`)} {
		t.Run(name, func(t *testing.T) {
			files := copyCorpus(t)
			files["testdata/manifest.json"].Data = data
			if _, err := LoadManifest(files); err == nil {
				t.Fatal("invalid JSON was accepted")
			}
		})
	}
	if _, err := Evaluate(context.Background(), fstest.MapFS{}); err == nil {
		t.Fatal("missing manifest was accepted")
	}
}

func fixtureByID(t *testing.T, id string) Fixture {
	t.Helper()
	manifest, err := LoadManifest(Corpus)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range manifest.Fixtures {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("unknown fixture %q", id)
	return Fixture{}
}

func TestUnexpectedFailuresAreNotCapabilityPasses(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name, id string
		mutate   func(*Fixture, fstest.MapFS)
	}{
		{"missing-binary", "core", func(f *Fixture, files fstest.MapFS) { delete(files, "testdata/"+f.Wasm) }},
		{"corrupt-binary", "gc", func(f *Fixture, files fstest.MapFS) { files["testdata/"+f.Wasm].Data = []byte("corrupt") }},
		{"missing-import-provider", "thread-worker", func(_ *Fixture, files fstest.MapFS) { delete(files, "testdata/shared-memory.wasm") }},
		{"invalid-import-provider", "thread-worker", func(_ *Fixture, files fstest.MapFS) { files["testdata/shared-memory.wasm"].Data = []byte("corrupt") }},
		{"unlinked-host", "host-callback", func(f *Fixture, _ fstest.MapFS) { f.ID = "unlinked" }},
		{"wrong-result", "core", func(f *Fixture, _ fstest.MapFS) { f.Calls[0].Results = []uint64{99} }},
		{"no-execution", "core", func(f *Fixture, _ fstest.MapFS) { f.Calls = nil }},
		{"wrong-stdout", "p1", func(f *Fixture, _ fstest.MapFS) { f.Stdout = "wrong\n" }},
		{"wrong-p1-entry", "p1", func(f *Fixture, files fstest.MapFS) {
			binary, _ := Corpus.ReadFile("testdata/core.wasm")
			files["testdata/"+f.Wasm].Data = binary
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixtureByID(t, tt.id)
			files := copyCorpus(t)
			tt.mutate(&f, files)
			got := evaluateFixture(ctx, files, f)
			if got.Outcome != "error" || got.MatchesExpected || got.Error == "" {
				t.Fatalf("invalid success: %+v", got)
			}
		})
	}
	if knownRejection("gc", "compile", nil) || knownRejection("gc", "execute", errors.New("x")) || knownRejection("component", "instantiate", errors.New("invalid version header")) {
		t.Fatal("arbitrary failure classified as unsupported")
	}
}

func newRuntime(t *testing.T) wazero.Runtime {
	t.Helper()
	r := wazero.NewRuntimeWithConfig(context.Background(), runtimeConfig())
	t.Cleanup(func() {
		if err := r.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	return r
}

func instantiate(t *testing.T, r wazero.Runtime, file, name string) api.Module {
	t.Helper()
	binary, err := Corpus.ReadFile("testdata/" + file)
	if err != nil {
		t.Fatal(err)
	}
	m, err := r.InstantiateWithConfig(context.Background(), binary, wazero.NewModuleConfig().WithName(name).WithStartFunctions())
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestCallAssertionsAndImportConflicts(t *testing.T) {
	ctx := context.Background()
	r := newRuntime(t)
	m := instantiate(t, r, "core.wasm", "core")
	cases := []Call{
		{Export: "missing"},
		{Export: "add", Args: []uint64{1}},
		{Export: "add", Args: []uint64{1, 2}},
		{Export: "add", Args: []uint64{1, 2}, Results: []uint64{4}},
		{Export: "add", Args: []uint64{1, 2}, Trap: true},
		{Export: "trap", Args: []uint64{1}, Trap: true},
		{Export: "trap"},
	}
	for _, call := range cases {
		if err := checkCall(ctx, m, call); err == nil {
			t.Errorf("accepted invalid assertion: %+v", call)
		}
	}
	for _, id := range []string{"host-callback", "p1"} {
		if err := installImports(ctx, Corpus, r, id); err != nil {
			t.Fatal(err)
		}
		if err := installImports(ctx, Corpus, r, id); err == nil {
			t.Fatalf("duplicate %s host registration accepted", id)
		}
	}
}

// Different instances and Function values share the same imported memory.
// This tests actual execution, not merely the threads feature flag.
func TestSharedMemoryConcurrencyAndWaitNotify(t *testing.T) {
	ctx := context.Background()
	r := newRuntime(t)
	shared := instantiate(t, r, "shared-memory.wasm", "shared")
	const workers, iterations = 4, 128
	modules := make([]api.Module, workers)
	for i := range modules {
		modules[i] = instantiate(t, r, "thread-worker.wasm", fmt.Sprintf("worker-%d", i))
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	errorsCh := make(chan error, workers)
	for _, m := range modules {
		wg.Add(1)
		go func(m api.Module) {
			defer wg.Done()
			<-start
			fn := m.ExportedFunction("add")
			for j := 0; j < iterations; j++ {
				if _, err := fn.Call(ctx, 1); err != nil {
					errorsCh <- err
					return
				}
			}
		}(m)
	}
	close(start)
	if err := checkCall(ctx, shared, Call{Export: "grow", Args: []uint64{1}, Results: []uint64{1}}); err != nil {
		t.Error(err)
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Error(err)
	}
	if err := checkCall(ctx, modules[0], Call{Export: "load", Results: []uint64{workers * iterations}}); err != nil {
		t.Fatal(err)
	}
	if err := checkCall(ctx, shared, Call{Export: "size", Results: []uint64{2}}); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- checkCall(ctx, modules[0], Call{Export: "wait", Args: []uint64{workers * iterations, uint64(time.Second)}, Results: []uint64{0}})
	}()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			t.Log(`CONCURRENCY_REPORT {"atomicWorkers":4,"increments":512,"memoryGrowth":true,"waitNotify":"passed"}`)
			return
		case <-tick.C:
			if _, err := modules[1].ExportedFunction("notify").Call(ctx, 1); err != nil {
				t.Fatal(err)
			}
		case <-deadline.C:
			t.Fatal("bounded wait did not complete")
		}
	}
}

// A finite wait bounds this known cancellation limitation. Its duration is
// observed rather than asserted against a fragile scheduler timing threshold.
func TestAtomicWaitCancellationObservation(t *testing.T) {
	r := newRuntime(t)
	shared := instantiate(t, r, "shared-memory.wasm", "shared")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := shared.ExportedFunction("wait").Call(ctx, 0, uint64(time.Second))
	elapsed := time.Since(start)
	if err == nil || ctx.Err() != context.DeadlineExceeded || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("unexpected cancellation outcome: err=%v context=%v", err, ctx.Err())
	}
	data, _ := json.Marshal(map[string]any{"id": "atomic-wait-cancellation", "guestWaitMillis": 1000, "contextTimeoutMillis": 20, "elapsedMillis": elapsed.Milliseconds(), "error": err.Error(), "classification": "known cancellation limitation; timing is observational"})
	t.Logf("CANCELLATION_REPORT %s", data)
}
