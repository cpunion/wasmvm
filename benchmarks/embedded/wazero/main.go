// Command wazero measures individual embedding operations on a pure Wasm module.
// API call timings include the host-to-guest boundary and guest work; assertions,
// clock calibration, and JSON output are outside those timed calls.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type options struct {
	path       string
	backend    string
	iterations uint32
	seed       uint32
	expected   uint32
	samples    int
	warmups    int
}

type phases struct {
	Load            int64 `json:"load"`
	EngineSetup     int64 `json:"engineSetup"`
	CompileValidate int64 `json:"compileValidate"`
	Instantiate     int64 `json:"instantiate"`
	ExportLookup    int64 `json:"exportLookup"`
	FirstCall       int64 `json:"firstCall"`
	Cleanup         int64 `json:"cleanup"`
}

type report struct {
	SchemaVersion int      `json:"schemaVersion"`
	Backend       string   `json:"backend"`
	Iterations    uint32   `json:"iterations"`
	Seed          uint32   `json:"seed"`
	Samples       int      `json:"samples"`
	Warmups       int      `json:"warmups"`
	Phases        phases   `json:"phasesNs"`
	Result        uint32   `json:"result"`
	WarmupCallNs  []int64  `json:"warmupCallNs"`
	WarmCallNs    []int64  `json:"warmCallNs"`
	WarmupResults []uint32 `json:"warmupResults"`
	WarmResults   []uint32 `json:"warmResults"`
	TimerPairNs   []int64  `json:"timerPairNs"`
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var result options
	flags := flag.NewFlagSet("wazero-phase-runner", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&result.path, "wasm", "", "pure Wasm module exporting run(i32, i32) -> i32")
	flags.StringVar(&result.backend, "backend", "interpreter", "interpreter or compiler")
	iterations := flags.Uint64("iterations", 1, "positive signed-i32 workload iteration count")
	seed := flags.Uint64("seed", 0, "unsigned i32 workload seed")
	expected := flags.Uint64("expected", 0, "required unsigned i32 checksum for every call")
	flags.IntVar(&result.samples, "samples", 7, "number of measured warm calls")
	flags.IntVar(&result.warmups, "warmups", 2, "number of untallied warmup calls after the first call")
	if err := flags.Parse(args); err != nil {
		return result, err
	}
	hasExpected := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "expected" {
			hasExpected = true
		}
	})
	if result.path == "" || !hasExpected || flags.NArg() != 0 {
		return result, errors.New("--wasm and --expected are required; positional arguments are not accepted")
	}
	if result.backend != "interpreter" && result.backend != "compiler" {
		return result, errors.New("--backend must be interpreter or compiler")
	}
	if *iterations == 0 || *iterations > math.MaxInt32 || *seed > math.MaxUint32 || *expected > math.MaxUint32 {
		return result, errors.New("iterations must be 1..2147483647; seed and expected must be unsigned i32 values")
	}
	if result.samples <= 0 || result.warmups < 0 {
		return result, errors.New("samples must be positive and warmups must be nonnegative")
	}
	result.iterations, result.seed, result.expected = uint32(*iterations), uint32(*seed), uint32(*expected)
	return result, nil
}

// pureModule is a contract check after wazero validation, not a Wasm validator.
// Import and start sections would introduce host work or hidden guest execution
// during instantiation, so they are excluded from this benchmark protocol.
func pureModule(wasm []byte) error {
	if len(wasm) < 8 || !bytes.Equal(wasm[:8], []byte{0, 'a', 's', 'm', 1, 0, 0, 0}) {
		return errors.New("expected a core Wasm version 1 module")
	}
	for cursor := 8; cursor < len(wasm); {
		id := wasm[cursor]
		cursor++
		size, width := binary.Uvarint(wasm[cursor:])
		if width <= 0 || size > math.MaxUint32 {
			return errors.New("invalid section length")
		}
		cursor += width
		if size > uint64(len(wasm)-cursor) {
			return errors.New("truncated section")
		}
		if id == 8 {
			return errors.New("pure benchmark modules must not have a start section")
		}
		if id == 2 {
			count, countWidth := binary.Uvarint(wasm[cursor : cursor+int(size)])
			if countWidth <= 0 || count != 0 {
				return errors.New("pure benchmark modules must not import host state or functions")
			}
		}
		cursor += int(size)
	}
	return nil
}

func runSignature(function api.Function) error {
	if function == nil {
		return errors.New("module must export run(i32, i32) -> i32")
	}
	definition := function.Definition()
	params, results := definition.ParamTypes(), definition.ResultTypes()
	if len(params) != 2 || params[0] != api.ValueTypeI32 || params[1] != api.ValueTypeI32 || len(results) != 1 || results[0] != api.ValueTypeI32 {
		return errors.New("run must have signature (i32, i32) -> i32")
	}
	return nil
}

func timedCall(ctx context.Context, function api.Function, stack []uint64, seed, iterations, expected uint32) (int64, uint32, error) {
	// Reuse argument/result storage, like the C embedding runners. Stack setup
	// and checksum checks are outside the API timer; no overhead is subtracted.
	stack[0], stack[1] = uint64(seed), uint64(iterations)
	start := time.Now()
	err := function.CallWithStack(ctx, stack)
	elapsed := time.Since(start).Nanoseconds()
	// All correctness work occurs after the timer stops, on every invocation.
	if err != nil {
		return elapsed, 0, fmt.Errorf("run trapped: %w", err)
	}
	if uint32(stack[0]) != expected {
		return elapsed, 0, fmt.Errorf("run checksum: got unsigned i32 %d, expected %d", uint32(stack[0]), expected)
	}
	return elapsed, uint32(stack[0]), nil
}

func benchmark(ctx context.Context, args options) (result report, err error) {
	result = report{SchemaVersion: 1, Backend: args.backend, Iterations: args.iterations,
		Seed: args.seed, Samples: args.samples, Warmups: args.warmups,
		WarmupCallNs: make([]int64, 0, args.warmups), WarmCallNs: make([]int64, 0, args.samples),
		WarmupResults: make([]uint32, 0, args.warmups), WarmResults: make([]uint32, 0, args.samples),
		TimerPairNs: make([]int64, 0, args.samples)}
	start := time.Now()
	wasm, err := os.ReadFile(args.path)
	result.Phases.Load = time.Since(start).Nanoseconds()
	if err != nil {
		return result, fmt.Errorf("load: %w", err)
	}
	start = time.Now()
	var config wazero.RuntimeConfig
	switch args.backend {
	case "interpreter":
		config = wazero.NewRuntimeConfigInterpreter()
	case "compiler":
		config = wazero.NewRuntimeConfigCompiler()
	default:
		return result, errors.New("unknown runtime backend")
	}
	runtime := wazero.NewRuntimeWithConfig(ctx, config)
	result.Phases.EngineSetup = time.Since(start).Nanoseconds()
	var compiled wazero.CompiledModule
	defer func() {
		start := time.Now()
		var compiledErr error
		if compiled != nil {
			compiledErr = compiled.Close(context.Background())
		}
		runtimeErr := runtime.Close(context.Background())
		result.Phases.Cleanup = time.Since(start).Nanoseconds()
		err = errors.Join(err, compiledErr, runtimeErr)
	}()
	start = time.Now()
	compiled, err = runtime.CompileModule(ctx, wasm)
	result.Phases.CompileValidate = time.Since(start).Nanoseconds()
	if err != nil {
		return result, fmt.Errorf("compile/validate: %w", err)
	}
	if err = pureModule(wasm); err != nil {
		return result, err
	}
	start = time.Now()
	module, err := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName("benchmark").WithStartFunctions())
	result.Phases.Instantiate = time.Since(start).Nanoseconds()
	if err != nil {
		return result, fmt.Errorf("instantiate: %w", err)
	}
	start = time.Now()
	function := module.ExportedFunction("run")
	result.Phases.ExportLookup = time.Since(start).Nanoseconds()
	if err = runSignature(function); err != nil {
		return result, err
	}
	callStack := make([]uint64, 2)
	result.Phases.FirstCall, result.Result, err = timedCall(ctx, function, callStack, args.seed, args.iterations, args.expected)
	if err != nil {
		return result, fmt.Errorf("first call: %w", err)
	}
	for index := 0; index < args.warmups; index++ {
		duration, value, callErr := timedCall(ctx, function, callStack, args.seed, args.iterations, args.expected)
		if callErr != nil {
			return result, fmt.Errorf("warmup %d: %w", index, callErr)
		}
		result.WarmupCallNs = append(result.WarmupCallNs, duration)
		result.WarmupResults = append(result.WarmupResults, value)
	}
	for index := 0; index < args.samples; index++ {
		duration, value, callErr := timedCall(ctx, function, callStack, args.seed, args.iterations, args.expected)
		if callErr != nil {
			return result, fmt.Errorf("warm sample %d: %w", index, callErr)
		}
		result.WarmCallNs = append(result.WarmCallNs, duration)
		result.WarmResults = append(result.WarmResults, value)
	}
	for index := 0; index < args.samples; index++ {
		start := time.Now()
		result.TimerPairNs = append(result.TimerPairNs, time.Since(start).Nanoseconds())
	}
	return result, nil
}

func runCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// os.Stdout is unbuffered: the parent observes this line as soon as the Go
	// entry point is reached, before flag parsing, file access, or engine setup.
	encoder := json.NewEncoder(stdout)
	if err := encoder.Encode(struct {
		Ready bool `json:"ready"`
	}{true}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	acknowledgement, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil || acknowledgement != "run\n" {
		fmt.Fprintln(stderr, "expected run acknowledgement on stdin after the ready event")
		return 1
	}
	configuration, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err == nil {
		var result report
		result, err = benchmark(context.Background(), configuration)
		if err == nil {
			err = encoder.Encode(result)
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
