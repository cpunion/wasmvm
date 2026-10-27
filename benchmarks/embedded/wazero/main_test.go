package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Unit fixtures are project-authored WAT compiled with wasm-tools 1.258.0:
// wasm-tools parse testdata/NAME.wat -o testdata/NAME.wasm
// wasm-tools validate testdata/NAME.wasm

func fixture(name string) string {
	return filepath.Join("testdata", name+".wasm")
}

func configuration(name, backend string) options {
	return options{path: fixture(name), backend: backend, seed: math.MaxUint32,
		iterations: 1, expected: math.MaxUint32, samples: 7, warmups: 2}
}

func TestBenchmarkBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "compiler"} {
		t.Run(backend, func(t *testing.T) {
			for _, input := range []string{"control", "reset-memory"} {
				t.Run(input, func(t *testing.T) {
					result, err := benchmark(context.Background(), configuration(input, backend))
					if err != nil {
						t.Fatal(err)
					}
					if result.SchemaVersion != 1 || result.Backend != backend || result.Result != math.MaxUint32 || result.Seed != math.MaxUint32 || result.Iterations != 1 {
						t.Fatalf("incorrect identity or unsigned result: %+v", result)
					}
					if len(result.WarmCallNs) != 7 || len(result.WarmResults) != 7 || len(result.TimerPairNs) != 7 || len(result.WarmupCallNs) != 2 || len(result.WarmupResults) != 2 {
						t.Fatalf("incomplete measurement/result arrays: %+v", result)
					}
					for _, values := range [][]uint32{result.WarmupResults, result.WarmResults} {
						for _, value := range values {
							if value != math.MaxUint32 {
								t.Fatalf("repeated call changed state/result: %d", value)
							}
						}
					}
					encoded, err := json.Marshal(result)
					if err != nil {
						t.Fatal(err)
					}
					var wire struct {
						Phases map[string]int64 `json:"phasesNs"`
					}
					if err = json.Unmarshal(encoded, &wire); err != nil {
						t.Fatal(err)
					}
					if len(wire.Phases) != 7 {
						t.Fatalf("phase schema mismatch: %s", encoded)
					}
					for _, key := range []string{"load", "engineSetup", "compileValidate", "instantiate", "exportLookup", "firstCall", "cleanup"} {
						if value, ok := wire.Phases[key]; !ok || value < 0 {
							t.Errorf("missing or negative phase %q: %s", key, encoded)
						}
					}
					for _, timings := range [][]int64{result.WarmupCallNs, result.WarmCallNs, result.TimerPairNs} {
						for _, duration := range timings {
							if duration < 0 {
								t.Fatal("negative duration")
							}
						}
					}
				})
			}
		})
	}
}

func TestPureCorpusBackends(t *testing.T) {
	// This is a repository integration test: all runners consume this shared
	// corpus. Validation cases stay small; the performance runner uses the main
	// iteration counts. Repeated calls validate per-invocation state reset.
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Workloads []struct {
			ID              string `json:"id"`
			Wasm            string `json:"wasm"`
			ValidationCases []struct {
				Seed       uint32 `json:"seed"`
				Iterations uint32 `json:"iterations"`
				Expected   uint32 `json:"expectedChecksum"`
			} `json:"validationCases"`
		} `json:"workloads"`
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Workloads) != 6 {
		t.Fatalf("expected six shared corpus workloads, got %d", len(corpus.Workloads))
	}
	for _, backend := range []string{"interpreter", "compiler"} {
		t.Run(backend, func(t *testing.T) {
			seen := map[string]bool{}
			for _, workload := range corpus.Workloads {
				if seen[workload.Wasm] {
					continue // Small/sustained workloads share the same validation cases.
				}
				seen[workload.Wasm] = true
				if len(workload.ValidationCases) == 0 {
					t.Fatalf("%s lacks validation cases", workload.ID)
				}
				for _, input := range workload.ValidationCases {
					t.Run(fmt.Sprintf("%s/seed-%d/iterations-%d", workload.ID, input.Seed, input.Iterations), func(t *testing.T) {
						args := options{path: filepath.Join("..", workload.Wasm), backend: backend,
							seed: input.Seed, iterations: input.Iterations, expected: input.Expected, samples: 2, warmups: 1}
						if _, err := benchmark(context.Background(), args); err != nil {
							t.Fatal(err)
						}
					})
				}
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	valid := []string{"--wasm", fixture("control"), "--seed", "4294967295", "--expected", "4294967295"}
	actual, err := parseOptions(valid, io.Discard)
	if err != nil || actual != configuration("control", "interpreter") {
		t.Fatalf("options=%+v error=%v", actual, err)
	}
	for _, invalid := range [][]string{
		{}, {"--wasm", fixture("control")}, {"--unknown"},
		append(append([]string{}, valid...), "position"),
		append(append([]string{}, valid...), "--backend", "native"),
		append(append([]string{}, valid...), "--iterations", "0"),
		append(append([]string{}, valid...), "--iterations", "2147483648"),
		append(append([]string{}, valid...), "--seed", "4294967296"),
		append(append([]string{}, valid...), "--expected", "4294967296"),
		append(append([]string{}, valid...), "--seed", "-1"),
		append(append([]string{}, valid...), "--samples", "0"),
		append(append([]string{}, valid...), "--warmups", "-1"),
	} {
		if _, err = parseOptions(invalid, io.Discard); err == nil {
			t.Errorf("accepted invalid arguments: %v", invalid)
		}
	}
}

func TestBenchmarkRejectsFailures(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
		mutate            func(*options)
	}{
		{"load", "absent", "load:", nil},
		{"backend", "control", "unknown runtime backend", func(args *options) { args.backend = "absent" }},
		{"checksum", "control", "first call: run checksum", func(args *options) { args.expected = 0 }},
		{"trap", "trap", "first call: run trapped", nil},
		{"signature", "wrong-signature", "run must have signature", nil},
		{"imports", "import", "must not import", nil},
		{"start", "start", "must not have a start", nil},
		{"instantiate", "instantiate-failure", "instantiate:", nil},
		{"warmup-checksum", "mutating", "warmup 0: run checksum", func(args *options) { args.expected = 1 }},
		{"sample-checksum", "mutating", "warm sample 0: run checksum", func(args *options) { args.expected, args.warmups = 1, 0 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			args := configuration(test.input, "interpreter")
			if test.mutate != nil {
				test.mutate(&args)
			}
			if _, err := benchmark(context.Background(), args); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("got %v, want %q", err, test.want)
			}
		})
	}
	for _, test := range []struct {
		data []byte
		want string
	}{
		{[]byte("invalid"), "compile/validate:"},
		{[]byte{0, 'a', 's', 'm', 1, 0, 0, 0}, "module must export run"},
	} {
		args := configuration("control", "interpreter")
		args.path = filepath.Join(t.TempDir(), "fixture.wasm")
		if err := os.WriteFile(args.path, test.data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := benchmark(context.Background(), args); err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("got %v, want %q", err, test.want)
		}
	}
}

func TestPureModuleBounds(t *testing.T) {
	header := []byte{0, 'a', 's', 'm', 1, 0, 0, 0}
	for _, suffix := range [][]byte{
		{0},                             // Missing section size.
		{0, 2},                          // Missing section body.
		{0, 0x80, 0x80, 0x80, 0x80, 16}, // Length exceeds uint32.
		{2, 0},                          // Missing import vector length.
		{8, 0},                          // Start section is forbidden.
	} {
		if err := pureModule(append(append([]byte{}, header...), suffix...)); err == nil {
			t.Errorf("accepted malformed or impure module: %v", suffix)
		}
	}
	if err := pureModule([]byte("wrong header")); err == nil {
		t.Fatal("accepted wrong header")
	}
	if err := pureModule(append(append([]byte{}, header...), 2, 1, 0)); err != nil {
		t.Fatalf("empty import section should be pure: %v", err)
	}
}

func TestCLIProtocol(t *testing.T) {
	args := []string{"--wasm", fixture("control"), "--seed", "42", "--expected", "42"}
	var stdout, stderr bytes.Buffer
	if code := runCLI(args, strings.NewReader("run\n"), &stdout, &stderr); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, &stderr)
	}
	decoder := json.NewDecoder(&stdout)
	var ready map[string]bool
	if err := decoder.Decode(&ready); err != nil || !reflect.DeepEqual(ready, map[string]bool{"ready": true}) {
		t.Fatalf("ready=%v err=%v", ready, err)
	}
	var result report
	if err := decoder.Decode(&result); err != nil || result.Result != 42 || len(result.WarmResults) != 7 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF || stderr.Len() != 0 {
		t.Fatalf("unexpected extra protocol output: err=%v stderr=%s", err, &stderr)
	}
	for _, input := range []string{"", "run", "invalid\n", "run\r\n"} {
		stdout.Reset()
		stderr.Reset()
		if code := runCLI(args, strings.NewReader(input), &stdout, &stderr); code != 1 || stdout.String() != "{\"ready\":true}\n" || !strings.Contains(stderr.String(), "acknowledgement") {
			t.Fatalf("invalid acknowledgement accepted: %q, code=%d, stdout=%s, stderr=%s", input, code, &stdout, &stderr)
		}
	}
	for _, args := range [][]string{{"--expected", "0"}, {"--wasm", "missing.wasm", "--expected", "0"}} {
		stdout.Reset()
		stderr.Reset()
		if code := runCLI(args, strings.NewReader("run\n"), &stdout, &stderr); code != 1 || stdout.String() != "{\"ready\":true}\n" || stderr.Len() == 0 {
			t.Fatalf("failed command emitted a success report: code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := runCLI([]string{"--help"}, strings.NewReader("run\n"), &stdout, &stderr); code != 0 || !strings.Contains(stderr.String(), "Usage") {
		t.Fatalf("help failed: code=%d stderr=%s", code, &stderr)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failure") }

type readyOnlyWriter struct {
	bytes.Buffer
}

func (writer *readyOnlyWriter) Write(data []byte) (int, error) {
	if writer.Len() > 0 {
		return 0, errors.New("report write failure")
	}
	return writer.Buffer.Write(data)
}

func TestCLIWriteFailures(t *testing.T) {
	var stderr bytes.Buffer
	if code := runCLI(nil, strings.NewReader("run\n"), failWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "write failure") {
		t.Fatalf("ready write failure was ignored: %d %s", code, &stderr)
	}
	stderr.Reset()
	args := []string{"--wasm", fixture("control"), "--seed", "42", "--expected", "42"}
	if code := runCLI(args, strings.NewReader("run\n"), &readyOnlyWriter{}, &stderr); code != 1 || !strings.Contains(stderr.String(), "report write failure") {
		t.Fatalf("report write failure was ignored: %d %s", code, &stderr)
	}
}

func TestInterleavedCalls(t *testing.T) {
	t.Setenv("WASMVM_INTERLEAVE", "1")
	args := []string{"--wasm", fixture("control"), "--seed", "42", "--expected", "42", "--samples", "2"}
	for _, test := range []struct {
		input string
		code  int
	}{
		{"run\nsample\nsample\n", 0},
		{"run\n", 1},
		{"run\nsample\n", 1},
		{"run\ninvalid\n", 1},
	} {
		var out, diagnostic bytes.Buffer
		code := runCLI(args, strings.NewReader(test.input), &out, &diagnostic)
		if code != test.code {
			t.Fatalf("input %q: exit %d: %s", test.input, code, &diagnostic)
		}
		if code == 0 {
			lines := strings.Split(strings.TrimSpace(out.String()), "\n")
			if len(lines) != 4 || lines[1] != `{"sampleReady":0}` || lines[2] != `{"sampleReady":1}` {
				t.Fatalf("invalid sample protocol: %s", &out)
			}
			var result report
			if err := json.Unmarshal([]byte(lines[3]), &result); err != nil || len(result.WarmResults) != 2 || result.WarmResults[0] != 42 || result.WarmResults[1] != 42 {
				t.Fatalf("invalid interleaved result: %+v, %v", result, err)
			}
		} else if !strings.Contains(diagnostic.String(), "sample acknowledgement") {
			t.Fatal(diagnostic.String())
		}
	}
	var diagnostic bytes.Buffer
	if code := runCLI(args, strings.NewReader("run\nsample\nsample\n"), &readyOnlyWriter{}, &diagnostic); code != 1 || !strings.Contains(diagnostic.String(), "write failure") {
		t.Fatalf("ignored sample event write failure: %d %s", code, &diagnostic)
	}
}
