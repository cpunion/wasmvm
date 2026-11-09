// Command service measures independent runtimes and native C using the same
// bounded admission driver as the owned engine. It does not import the VM.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

type options struct {
	engine, backend, path string
	seed, iterations      uint32
	termination           bool
	service.Options
}

type report struct {
	SchemaVersion int              `json:"schemaVersion"`
	Engine        string           `json:"engine"`
	Backend       string           `json:"backend"`
	HostCompiler  string           `json:"hostCompiler"`
	Mode          string           `json:"mode"`
	Policy        string           `json:"policy"`
	Seed          uint32           `json:"seed"`
	Iterations    uint32           `json:"iterations"`
	Samples       int              `json:"samples"`
	Warmups       int              `json:"warmups"`
	Phases        map[string]int64 `json:"phasesNs"`
	Result        uint32           `json:"result"`
	WarmCallNs    []int64          `json:"warmCallNs"`
	WarmupCallNs  []int64          `json:"warmupCallNs"`
	WarmResults   []uint32         `json:"warmResults"`
	WarmupResults []uint32         `json:"warmupResults"`
	Service       *service.Report  `json:"service"`
}

func parse(args []string, stderr io.Writer) (out options, err error) {
	f := flag.NewFlagSet("service-reference", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&out.engine, "engine", "wazero", "wazero or native-c")
	f.StringVar(&out.backend, "backend", "compiler", "wazero: compiler/interpreter; native-c: native")
	f.StringVar(&out.path, "wasm", "", "pure Core module for wazero")
	f.BoolVar(&out.termination, "termination-checks", false, "enable wazero context termination checks")
	f.IntVar(&out.Workers, "workers", 1, "independent workers")
	f.IntVar(&out.Clients, "clients", 64, "clients with one outstanding request each")
	f.IntVar(&out.Requests, "requests", 16, "requests per client per batch")
	f.IntVar(&out.Samples, "samples", 7, "measured batches")
	f.IntVar(&out.Warmups, "warmups", 2, "untallied warmup batches")
	f.BoolVar(&out.Resources, "service-resources", false, "separate OS resource cohort")
	seed := f.Uint64("seed", 17, "unsigned i32 seed")
	iterations := f.Uint64("iterations", 64, "positive signed i32 iterations")
	expected := f.Uint64("expected", 0, "required unsigned i32 checksum")
	if err = f.Parse(args); err != nil {
		return out, err
	}
	hasExpected := false
	f.Visit(func(v *flag.Flag) { hasExpected = hasExpected || v.Name == "expected" })
	if !hasExpected || f.NArg() != 0 || *seed > math.MaxUint32 || *expected > math.MaxUint32 || *iterations == 0 || *iterations > math.MaxInt32 {
		return out, errors.New("expected checksum and valid i32 seed/iterations are required; no positional arguments")
	}
	if out.engine == "wazero" {
		if out.path == "" || (out.backend != "compiler" && out.backend != "interpreter") {
			return out, errors.New("wazero requires --wasm and backend compiler/interpreter")
		}
	} else if out.engine != "native-c" || out.backend != "native" || out.termination || out.path != "" || *iterations > 65536 {
		return out, errors.New("native-c requires backend native, iterations <=65536, no Wasm path or termination policy")
	}
	out.seed, out.iterations, out.Expected = uint32(*seed), uint32(*iterations), uint32(*expected)
	return out, out.Options.Validate()
}

func benchmark(ctx context.Context, args options, beforeSample func(int) error) (out report, err error) {
	if err = args.Options.Validate(); err != nil {
		return out, err
	}
	out = report{SchemaVersion: 1, Engine: args.engine, Backend: args.backend, HostCompiler: runtime.Compiler,
		Mode: "service", Seed: args.seed, Iterations: args.iterations, Samples: args.Samples, Warmups: args.Warmups,
		Phases: make(map[string]int64), WarmCallNs: make([]int64, 0, args.Samples),
		WarmupCallNs: make([]int64, 0, args.Warmups), WarmResults: make([]uint32, 0, args.Samples),
		WarmupResults: make([]uint32, 0, args.Warmups)}
	var calls []service.Call
	var release func() error
	var stop func()
	if args.engine == "native-c" {
		out.Policy = "ordinary scalar C SHA-256 through the host C bridge; shared admission driver; no VM safety, cancellation, fuel, JIT or Wasm lifecycle"
		calls, release, err = nativeKernels(args)
	} else {
		out.Policy = fmt.Sprintf("wazero independent instances sharing one compiled module; context termination checks=%t", args.termination)
		calls, release, err = wazeroKernels(ctx, args, &out)
	}
	if release != nil {
		defer func() {
			start := time.Now()
			if stop != nil {
				stop()
			}
			err = errors.Join(err, release())
			out.Phases["cleanup"] = time.Since(start).Nanoseconds()
		}()
	}
	if err != nil {
		return out, err
	}
	start := time.Now()
	call, joinedStop, data, err := service.Start(calls, args.Options)
	if err != nil {
		return out, err
	}
	out.Service = data
	stop = joinedStop
	data.WorkerSetupNs = time.Since(start).Nanoseconds()
	out.Phases["firstCall"], out.Result, err = call()
	if err != nil {
		return out, err
	}
	for range args.Warmups {
		ns, value, callErr := call()
		if callErr != nil {
			return out, callErr
		}
		out.WarmupCallNs, out.WarmupResults = append(out.WarmupCallNs, ns), append(out.WarmupResults, value)
	}
	for i := range args.Samples {
		if beforeSample != nil {
			if err = beforeSample(i); err != nil {
				return out, err
			}
		}
		ns, value, callErr := call()
		if callErr != nil {
			return out, callErr
		}
		out.WarmCallNs, out.WarmResults = append(out.WarmCallNs, ns), append(out.WarmResults, value)
	}
	return out, nil
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	encoder := json.NewEncoder(stdout)
	err := encoder.Encode(struct {
		Ready bool `json:"ready"`
	}{true})
	reader := bufio.NewReader(stdin)
	if err == nil {
		line, readErr := reader.ReadString('\n')
		if readErr != nil || line != "run\n" {
			err = errors.New("expected run acknowledgement")
		}
	}
	var config options
	if err == nil {
		config, err = parse(args, stderr)
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err == nil {
		var wait func(int) error
		if os.Getenv("WASMVM_INTERLEAVE") == "1" {
			wait = func(i int) error {
				if err := encoder.Encode(struct {
					SampleReady int `json:"sampleReady"`
				}{i}); err != nil {
					return err
				}
				line, err := reader.ReadString('\n')
				if err != nil || line != "sample\n" {
					return errors.New("expected sample acknowledgement")
				}
				return nil
			}
		}
		var out report
		out, err = benchmark(context.Background(), config, wait)
		if err == nil {
			err = encoder.Encode(out)
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
