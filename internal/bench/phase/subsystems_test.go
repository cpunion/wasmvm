package phase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const subsystems = "../../../benchmarks/embedded/subsystems"

type subsystemFixture struct {
	ID, Wasm, Outcome                  string
	DefinedFunctions                   uint32
	Imports                            int
	Seed, Iterations, ExpectedChecksum uint32
	ValidationCases                    []struct{ Seed, Iterations, ExpectedChecksum uint32 }
}

func subsystemFixtures(t *testing.T) (host, preparation []subsystemFixture) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(subsystems, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct{ Host, Preparation []subsystemFixture }
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m.Host, m.Preparation
}
func TestHostCallbackBoundaryAndExactCounts(t *testing.T) {
	hosts, _ := subsystemFixtures(t)
	for _, fixture := range hosts {
		args := optionsFor(filepath.Join(subsystems, filepath.Base(fixture.Wasm)), "host")
		for _, v := range fixture.ValidationCases {
			args.seed, args.iterations, args.expected = v.Seed, v.Iterations, v.ExpectedChecksum
			r, err := benchmark(context.Background(), args)
			if err != nil || r.HostCalls != uint64(v.Iterations)*uint64(1+args.warmups+args.samples) {
				t.Fatal(fixture.ID, v, r.HostCalls, err)
			}
		}
		args.mode = "api"
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal("host imports entered pure series")
		}
	}
	for _, name := range []string{"import", "start", "wrong-signature"} {
		args := optionsFor(filepath.Join(negatives, name+".wasm"), "host")
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal(name)
		}
	}
	var out bytes.Buffer
	args := []string{"--mode", "host", "--wasm", filepath.Join(subsystems, "host-callback.wasm"), "--seed", "0", "--iterations", "1", "--expected", "2654435769", "--samples", "2"}
	if code := RunCLI(args, strings.NewReader("run\n"), &out, io.Discard); code != 0 {
		t.Fatal(code, out.String())
	}
}
func TestFreshPreparationShapesAndClassifications(t *testing.T) {
	_, fixtures := subsystemFixtures(t)
	for _, fixture := range fixtures {
		path := filepath.Join(subsystems, filepath.Base(fixture.Wasm))
		input, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for n := 0; n < 2; n++ {
			got, err := prepareOnce(context.Background(), input)
			if err != nil || got.Outcome != fixture.Outcome || got.Functions != fixture.DefinedFunctions || got.Imports != fixture.Imports {
				t.Fatal(fixture.ID, got, err)
			}
			_, instantiated := got.Phases["instantiate"]
			if instantiated != (fixture.Outcome == "valid") {
				t.Fatal(fixture.ID, got)
			}
			for _, ns := range got.Phases {
				if ns < 0 {
					t.Fatal(got)
				}
			}
		}
		var stdout, stderr bytes.Buffer
		args := []string{"prepare", "--wasm", path, "--expect", fixture.Outcome, "--samples", "2"}
		if code := RunCLI(args, strings.NewReader("run\n"), &stdout, &stderr); code != 0 {
			t.Fatal(fixture.ID, code, stderr.String())
		}
		lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		var report preparationReport
		if len(lines) != 2 || json.Unmarshal([]byte(lines[1]), &report) != nil || len(report.Samples) != 2 || report.InputBytes != len(input) || report.Mode != "prepare" {
			t.Fatal(fixture.ID, stdout.String())
		}
	}
	data, _ := os.ReadFile(filepath.Join(subsystems, "small-module.wasm"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := prepareOnce(ctx, data); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, name := range []string{"start", "instantiate-failure"} {
		data, err := os.ReadFile(filepath.Join(negatives, name+".wasm"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := prepareOnce(context.Background(), data); err == nil {
			t.Fatal(name)
		}
	}
}
func TestPreparationProtocolFailures(t *testing.T) {
	path := filepath.Join(subsystems, "small-module.wasm")
	for _, args := range [][]string{{"prepare"}, {"prepare", "--bad"}, {"prepare", "--wasm", path, "extra"}, {"prepare", "--wasm", path, "--samples", "0"}, {"prepare", "--wasm", path, "--samples", "10001"}, {"prepare", "--wasm", path, "--expect", "bad"}, {"prepare", "--wasm", "missing"}, {"prepare", "--wasm", path, "--expect", "malformed"}, {"prepare", "--wasm", filepath.Join(negatives, "start.wasm")}} {
		if code := RunCLI(args, strings.NewReader("run\n"), io.Discard, io.Discard); code != 1 {
			t.Fatal(args, code)
		}
	}
	args := []string{"prepare", "--wasm", path, "--samples", "1"}
	if code := RunCLI(args, strings.NewReader("run\n"), &failingWriter{after: 1}, io.Discard); code != 1 {
		t.Fatal(code)
	}
}
