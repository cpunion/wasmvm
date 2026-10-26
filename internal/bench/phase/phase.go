// Package phase measures the owned engine through public API and internal
// dispatch boundaries. Diagnostics never replace the public API comparison.
package phase

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

	"github.com/cpunion/wasmvm/internal/compiler/lower"
	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/core/decode"
	"github.com/cpunion/wasmvm/internal/core/validate"
	"github.com/cpunion/wasmvm/internal/engine/interp"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// The build orchestrator generates a separate harness-only init file. Engine
// sources are archived from the exact selected commit and never overlaid.
var sourceRevision = "development"
var compilerRevision = "development"

type options struct {
	path, backend, mode        string
	iterations, seed, expected uint32
	samples, warmups           int
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
	SchemaVersion    int              `json:"schemaVersion"`
	Engine           string           `json:"engine"`
	SourceRevision   string           `json:"sourceRevision"`
	Compiler         string           `json:"compiler"`
	CompilerRevision string           `json:"compilerRevision"`
	RuntimeVersion   string           `json:"runtimeVersion"`
	GOOS             string           `json:"goos"`
	GOARCH           string           `json:"goarch"`
	Backend          string           `json:"backend"`
	Mode             string           `json:"mode"`
	Policy           string           `json:"policy"`
	Iterations       uint32           `json:"iterations"`
	Seed             uint32           `json:"seed"`
	Samples          int              `json:"samples"`
	Warmups          int              `json:"warmups"`
	Phases           phases           `json:"phasesNs"`
	PreparationParts map[string]int64 `json:"preparationPartsNs,omitempty"`
	Result           uint32           `json:"result"`
	WarmupCallNs     []int64          `json:"warmupCallNs"`
	WarmCallNs       []int64          `json:"warmCallNs"`
	WarmupResults    []uint32         `json:"warmupResults"`
	WarmResults      []uint32         `json:"warmResults"`
	TimerPairNs      []int64          `json:"timerPairNs"`
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var out options
	flags := flag.NewFlagSet("wasmvm-phase", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&out.path, "wasm", "", "")
	flags.StringVar(&out.backend, "backend", "interpreter", "")
	flags.StringVar(&out.mode, "mode", "api", "")
	iterations := flags.Uint64("iterations", 1, "")
	seed := flags.Uint64("seed", 0, "")
	expected := flags.Uint64("expected", 0, "")
	flags.IntVar(&out.samples, "samples", 7, "")
	flags.IntVar(&out.warmups, "warmups", 2, "")
	if err := flags.Parse(args); err != nil {
		return out, err
	}
	hasExpected := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "expected" {
			hasExpected = true
		}
	})
	if out.path == "" || !hasExpected || flags.NArg() != 0 {
		return out, errors.New("--wasm and --expected are required; no positional arguments")
	}
	if out.backend != "interpreter" {
		return out, wasm.ErrUnsupported
	}
	if out.mode != "api" && out.mode != "dispatch" {
		return out, errors.New("mode must be api or dispatch")
	}
	if *iterations == 0 || *iterations > math.MaxInt32 || *seed > math.MaxUint32 || *expected > math.MaxUint32 {
		return out, errors.New("invalid i32 workload arguments")
	}
	if out.samples < 1 || out.samples > 10000 || out.warmups < 0 || out.warmups > 10000 {
		return out, errors.New("samples must be 1..10000 and warmups 0..10000")
	}
	out.iterations, out.seed, out.expected = uint32(*iterations), uint32(*seed), uint32(*expected)
	return out, nil
}

// A kernel times the actual call and stops its clock before inspecting results.
// Public API marshalling remains inside its timer; dispatch is a separate row.
type kernel func() (elapsed int64, result uint32, err error)

func checkResult(result, expected uint32) error {
	if result != expected {
		return fmt.Errorf("run checksum: got %d, expected %d", result, expected)
	}
	return nil
}
func signature(t vm.FunctionType) error {
	if len(t.Params) != 2 || t.Params[0] != wasm.I32 || t.Params[1] != wasm.I32 || len(t.Results) != 1 || t.Results[0] != wasm.I32 {
		return errors.New("run must have signature (i32, i32) -> i32")
	}
	return nil
}
func pure(imports int, start bool) error {
	if imports != 0 || start {
		return errors.New("pure phase modules must not import or have start functions")
	}
	return nil
}

func publicKernel(ctx context.Context, input []byte, args options, out *report) (call kernel, close func(), err error) {
	start := time.Now()
	engine := vm.NewEngine()
	store, err := engine.NewStore(vm.DefaultStoreConfig())
	out.Phases.EngineSetup = time.Since(start).Nanoseconds()
	var compiled *vm.CompiledModule
	close = func() {
		start := time.Now()
		if store != nil {
			store.Close()
			_ = store.WaitClosed(context.Background())
		}
		if compiled != nil {
			compiled.Close()
		}
		engine.Close()
		out.Phases.Cleanup = time.Since(start).Nanoseconds()
	}
	if err != nil {
		return nil, close, err
	}
	start = time.Now()
	compiled, err = engine.Compile(ctx, input)
	out.Phases.CompileValidate = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	info, err := compiled.Info()
	if err != nil {
		return nil, close, err
	}
	if err = pure(len(info.Imports), info.HasStart); err != nil {
		return nil, close, err
	}
	linker := vm.NewLinker(store)
	start = time.Now()
	instance, err := linker.Instantiate(ctx, compiled, vm.CallOptions{})
	out.Phases.Instantiate = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	start = time.Now()
	fn, err := instance.Export("run")
	out.Phases.ExportLookup = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	typ, err := fn.FunctionType()
	if err != nil {
		return nil, close, err
	}
	if err = signature(typ); err != nil {
		return nil, close, err
	}
	arguments := []vm.Value{vm.I32(args.seed), vm.I32(args.iterations)}
	return func() (int64, uint32, error) {
		start := time.Now()
		result, err := fn.Call(ctx, arguments, vm.CallOptions{})
		elapsed := time.Since(start).Nanoseconds()
		if err != nil {
			return elapsed, 0, err
		}
		lo, _, err := result.Values[0].Bits()
		if err != nil {
			return elapsed, 0, err
		}
		return elapsed, uint32(lo), checkResult(uint32(lo), args.expected)
	}, close, nil
}

func dispatchKernel(ctx context.Context, input []byte, args options, out *report) (call kernel, close func(), err error) {
	start := time.Now()
	store := state.New(state.DefaultLimits())
	ctx, cancel := context.WithCancel(ctx)
	out.Phases.EngineSetup = time.Since(start).Nanoseconds()
	close = func() {
		start := time.Now()
		cancel()
		store.Close()
		out.Phases.Cleanup = time.Since(start).Nanoseconds()
	}
	out.PreparationParts = make(map[string]int64)
	preparationStart := time.Now()
	start = time.Now()
	decoded, err := decode.Decode(input, decode.DefaultLimits())
	out.PreparationParts["decode"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	start = time.Now()
	validated, err := validate.Check(decoded, validate.DefaultLimits())
	out.PreparationParts["validate"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	start = time.Now()
	ir, err := lower.Compile(validated, lower.DefaultLimits())
	out.PreparationParts["lower"] = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	start = time.Now()
	compiled, err := slots.Compile(ir, slots.DefaultLimits())
	out.PreparationParts["slots"] = time.Since(start).Nanoseconds()
	out.Phases.CompileValidate = time.Since(preparationStart).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	if err = pure(len(decoded.Imports), decoded.HasStart); err != nil {
		return nil, close, err
	}
	start = time.Now()
	instance, err := store.Allocate(compiled, nil)
	if err == nil {
		err = instance.Initialize()
	}
	if err == nil {
		err = store.Publish(instance)
	}
	out.Phases.Instantiate = time.Since(start).Nanoseconds()
	if err != nil {
		return nil, close, err
	}
	start = time.Now()
	export, ok := instance.Exports["run"]
	out.Phases.ExportLookup = time.Since(start).Nanoseconds()
	if !ok || export.Kind != 0 {
		return nil, close, errors.New("missing function export run")
	}
	fn, _ := store.Function(export.Function)
	if err = signature(fn.Signature); err != nil {
		return nil, close, err
	}
	execution := interp.New(store, ctx, interp.Config{Limits: interp.DefaultLimits()})
	arguments := []value.Value{{Lo: uint64(args.seed)}, {Lo: uint64(args.iterations)}}
	return func() (int64, uint32, error) {
		start := time.Now()
		err := execution.Start(export.Function, arguments)
		status := execution.Run()
		elapsed := time.Since(start).Nanoseconds()
		if err != nil {
			return elapsed, 0, err
		}
		if status != interp.Done {
			err := execution.Err()
			if err == nil {
				err = errors.New("unexpected interpreter suspension")
			}
			return elapsed, 0, err
		}
		result := uint32(execution.Results()[0].Lo)
		return elapsed, result, checkResult(result, args.expected)
	}, close, nil
}

func benchmark(ctx context.Context, args options) (out report, err error) {
	out = report{SchemaVersion: 1, Engine: "wasmvm", SourceRevision: sourceRevision, Compiler: runtime.Compiler, CompilerRevision: compilerRevision, RuntimeVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Backend: args.backend, Mode: args.mode, Policy: "unmetered; bounded cancellation; checked Core 2.0", Iterations: args.iterations, Seed: args.seed, Samples: args.samples, Warmups: args.warmups, WarmupCallNs: []int64{}, WarmCallNs: []int64{}, WarmupResults: []uint32{}, WarmResults: []uint32{}, TimerPairNs: []int64{}}
	start := time.Now()
	input, err := os.ReadFile(args.path)
	out.Phases.Load = time.Since(start).Nanoseconds()
	if err != nil {
		return out, err
	}
	var call kernel
	var close func()
	if args.mode == "api" {
		call, close, err = publicKernel(ctx, input, args, &out)
	} else {
		call, close, err = dispatchKernel(ctx, input, args, &out)
	}
	defer close()
	if err != nil {
		return out, err
	}
	out.Phases.FirstCall, out.Result, err = call()
	if err != nil {
		return out, err
	}
	for i := 0; i < args.warmups; i++ {
		ns, v, e := call()
		if e != nil {
			return out, fmt.Errorf("warmup %d: %w", i, e)
		}
		out.WarmupCallNs = append(out.WarmupCallNs, ns)
		out.WarmupResults = append(out.WarmupResults, v)
	}
	for i := 0; i < args.samples; i++ {
		ns, v, e := call()
		if e != nil {
			return out, fmt.Errorf("sample %d: %w", i, e)
		}
		out.WarmCallNs = append(out.WarmCallNs, ns)
		out.WarmResults = append(out.WarmResults, v)
	}
	for i := 0; i < args.samples; i++ {
		start := time.Now()
		out.TimerPairNs = append(out.TimerPairNs, time.Since(start).Nanoseconds())
	}
	return out, nil
}

func RunCLI(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	encoder := json.NewEncoder(stdout)
	if err := encoder.Encode(struct {
		Ready bool `json:"ready"`
	}{true}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	ack, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil || ack != "run\n" {
		fmt.Fprintln(stderr, "expected run acknowledgement after ready")
		return 1
	}
	config, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err == nil {
		var out report
		out, err = benchmark(context.Background(), config)
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
