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

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const fixtures = "../../../benchmarks/embedded/fixtures"
const negatives = "../../../benchmarks/embedded/wazero/testdata"

func optionsFor(path, mode string) options {
	return options{path: path, backend: "interpreter", mode: mode, iterations: 1, seed: 7, expected: 7, samples: 2, warmups: 1}
}

func TestBothBoundariesAndEveryOracleVector(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(fixtures, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Workloads []struct {
			Wasm            string
			ValidationCases []struct{ Seed, Iterations, ExpectedChecksum uint32 }
		}
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, mode := range []string{"api", "dispatch"} {
			for _, workload := range manifest.Workloads {
				for _, vector := range workload.ValidationCases {
					args := optionsFor(filepath.Join(fixtures, filepath.Base(workload.Wasm)), mode)
					args.backend = backend
					args.seed, args.iterations, args.expected = vector.Seed, vector.Iterations, vector.ExpectedChecksum
					out, err := benchmark(context.Background(), args)
					if err != nil {
						t.Fatal(mode, workload.Wasm, vector, err)
					}
					if out.Engine != "wasmvm" || out.Backend != backend || out.Mode != mode || out.Result != vector.ExpectedChecksum || len(out.WarmResults) != 2 || len(out.WarmupResults) != 1 || len(out.TimerPairNs) != 2 {
						t.Fatal(out)
					}
					for _, v := range append(out.WarmResults, out.WarmupResults...) {
						if v != vector.ExpectedChecksum {
							t.Fatal(v)
						}
					}
					if mode == "dispatch" {
						parts := 4
						if backend == "native" {
							parts++
						}
						if len(out.PreparationParts) != parts {
							t.Fatal(out.PreparationParts)
						}
						var sum int64
						for _, ns := range out.PreparationParts {
							if ns < 0 {
								t.Fatal(ns)
							}
							sum += ns
						}
						if sum > out.Phases.CompileValidate {
							t.Fatal(sum, out.Phases)
						}
					} else if out.PreparationParts != nil {
						t.Fatal("invented public subphases")
					}
				}
			}
		}
	}
}

func TestFailuresDoNotProduceSuccessfulMeasurements(t *testing.T) {
	for _, mode := range []string{"api", "dispatch"} {
		for _, name := range []string{"trap.wasm", "import.wasm", "start.wasm", "instantiate-failure.wasm", "wrong-signature.wasm", "mutating.wasm"} {
			args := optionsFor(filepath.Join(negatives, name), mode)
			if _, err := benchmark(context.Background(), args); err == nil {
				t.Fatal(mode, name)
			}
		}
		args := optionsFor(filepath.Join(fixtures, "call-control.wasm"), mode)
		args.expected = 0
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal("wrong checksum")
		}
		args.expected = 7
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := benchmark(ctx, args); !errors.Is(err, context.Canceled) {
			t.Fatal(mode, err)
		}
		args.path = filepath.Join(t.TempDir(), "missing")
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal("missing input")
		}
		args.path = filepath.Join(t.TempDir(), "bad.wasm")
		if err := os.WriteFile(args.path, []byte{0}, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := benchmark(context.Background(), args); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal(err)
		}
		// A complete binary with an invalid stack, then a valid empty module
		// without run, exercises validation and export failures independently.
		invalid := []byte{0, 97, 115, 109, 1, 0, 0, 0, 1, 4, 1, 96, 0, 0, 3, 2, 1, 0, 10, 5, 1, 3, 0, 0x1a, 0xb}
		os.WriteFile(args.path, invalid, 0600)
		if _, err := benchmark(context.Background(), args); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal(err)
		}
		os.WriteFile(args.path, invalid[:8], 0600)
		if _, err := benchmark(context.Background(), args); err == nil {
			t.Fatal("missing run")
		}
	}
	if err := signature(vm.FunctionType{}); err == nil {
		t.Fatal("signature")
	}
}

func TestMutableGuestFailureIsCheckedAfterEachCall(t *testing.T) {
	// The authored module returns the seed and increments a global each time.
	// Its first call succeeds; subsequent calls must fail checksum validation.
	for _, mode := range []string{"api", "dispatch"} {
		for _, warmups := range []int{0, 1} {
			args := optionsFor(filepath.Join(negatives, "mutating.wasm"), mode)
			args.warmups = warmups
			args.seed = 0
			args.expected = 0
			_, err := benchmark(context.Background(), args)
			if err == nil {
				t.Fatal(mode, warmups)
			}
		}
	}
}

func TestOptionsAndReadyProtocol(t *testing.T) {
	valid := []string{"--wasm", filepath.Join(fixtures, "call-control.wasm"), "--expected", "7", "--seed", "7", "--samples", "2", "--warmups", "1"}
	config, err := parseOptions(valid, io.Discard)
	if err != nil || config.expected != 7 || config.mode != "api" {
		t.Fatal(config, err)
	}
	for _, args := range [][]string{nil, {"--bad"}, {"--wasm", "x"}, {"--wasm", "x", "--expected", "0", "extra"}, append(append([]string{}, valid...), "--backend", "compiler"), append(append([]string{}, valid...), "--mode", "bad"), append(append([]string{}, valid...), "--samples", "0"), append(append([]string{}, valid...), "--samples", "10001"), append(append([]string{}, valid...), "--warmups", "-1"), append(append([]string{}, valid...), "--iterations", "0"), append(append([]string{}, valid...), "--iterations", "2147483648"), append(append([]string{}, valid...), "--seed", "4294967296"), append(append([]string{}, valid...), "--expected", "4294967296")} {
		if _, err := parseOptions(args, io.Discard); err == nil {
			t.Fatal(args)
		}
	}
	for _, mode := range []string{"api", "dispatch"} {
		var out, stderr bytes.Buffer
		args := append(append([]string{}, valid...), "--mode", mode)
		if code := RunCLI(args, strings.NewReader("run\n"), &out, &stderr); code != 0 {
			t.Fatal(code, stderr.String())
		}
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(lines) != 2 || lines[0] != `{"ready":true}` {
			t.Fatal(out.String())
		}
		var report report
		if err := json.Unmarshal([]byte(lines[1]), &report); err != nil || report.Result != 7 {
			t.Fatal(err, report)
		}
	}
	for _, ack := range []string{"", "no\n", "run"} {
		if code := RunCLI(valid, strings.NewReader(ack), io.Discard, io.Discard); code != 1 {
			t.Fatal(code)
		}
	}
	if code := RunCLI([]string{"--help"}, strings.NewReader("run\n"), io.Discard, io.Discard); code != 0 {
		t.Fatal(code)
	}
	if code := RunCLI(nil, strings.NewReader("run\n"), io.Discard, io.Discard); code != 1 {
		t.Fatal(code)
	}
	wrong := append(append([]string{}, valid...), "--expected", "0")
	if code := RunCLI(wrong, strings.NewReader("run\n"), io.Discard, io.Discard); code != 1 {
		t.Fatal(code)
	}
	for _, after := range []int{0, 1} {
		if code := RunCLI(valid, strings.NewReader("run\n"), &failingWriter{after: after}, io.Discard); code != 1 {
			t.Fatal(code)
		}
	}
}

type failingWriter struct{ after, writes int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.writes >= w.after {
		return 0, io.ErrClosedPipe
	}
	w.writes++
	return len(p), nil
}
