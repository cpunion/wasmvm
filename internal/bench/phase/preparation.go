package phase

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type preparationSample struct {
	Phases    map[string]int64 `json:"phasesNs"`
	Outcome   string           `json:"outcome"`
	Functions uint32           `json:"definedFunctions"`
	Imports   int              `json:"imports"`
	Usage     vm.Usage         `json:"usage"`
}
type preparationReport struct {
	SchemaVersion    int                 `json:"schemaVersion"`
	Engine           string              `json:"engine"`
	SourceRevision   string              `json:"sourceRevision"`
	Compiler         string              `json:"compiler"`
	CompilerRevision string              `json:"compilerRevision"`
	RuntimeVersion   string              `json:"runtimeVersion"`
	GOOS             string              `json:"goos"`
	GOARCH           string              `json:"goarch"`
	Mode             string              `json:"mode"`
	InputBytes       int                 `json:"inputBytes"`
	LoadNs           int64               `json:"loadNs"`
	Samples          []preparationSample `json:"samples"`
}

// Each sample constructs and closes a fresh Engine and Store. No guest export
// is invoked; rejected compilation is a distinct outcome, not fast execution.
func prepareOnce(ctx context.Context, input []byte) (out preparationSample, err error) {
	out.Phases = make(map[string]int64)
	out.Outcome = "valid"
	start := time.Now()
	engine := vm.NewEngine()
	store, err := engine.NewStore(vm.DefaultStoreConfig())
	out.Phases["engineSetup"] = time.Since(start).Nanoseconds()
	var compiled *vm.CompiledModule
	defer func() {
		start := time.Now()
		if store != nil {
			store.Close()
			_ = store.WaitClosed(context.Background())
		}
		if compiled != nil {
			compiled.Close()
		}
		engine.Close()
		out.Phases["cleanup"] = time.Since(start).Nanoseconds()
	}()
	if err != nil {
		return out, err
	}
	start = time.Now()
	compiled, err = engine.Compile(ctx, input)
	out.Phases["compileValidate"] = time.Since(start).Nanoseconds()
	if err != nil {
		for _, kind := range []struct {
			name  string
			value error
		}{{"malformed", wasm.ErrMalformed}, {"invalid", wasm.ErrInvalid}, {"limit", wasm.ErrLimit}} {
			if errors.Is(err, kind.value) {
				out.Outcome = kind.name
				return out, nil
			}
		}
		return out, err
	}
	info, err := compiled.Info()
	if err != nil {
		return out, err
	}
	if info.HasStart {
		return out, errors.New("preparation fixtures cannot execute start functions")
	}
	out.Functions, out.Imports = info.Functions, len(info.Imports)
	linker := vm.NewLinker(store)
	for _, im := range info.Imports {
		if im.Function == nil {
			return out, errors.New("preparation fixture imports must be functions")
		}
	}
	start = time.Now()
	for _, im := range info.Imports {
		fn, err := store.NewFunction(*im.Function, func(*vm.Caller, []vm.Value) ([]vm.Value, error) {
			return nil, errors.New("unexpected guest execution in preparation fixture")
		})
		if err != nil {
			return out, err
		}
		if err := linker.Define(im.Module, im.Name, fn); err != nil {
			return out, err
		}
	}
	out.Phases["hostBinding"] = time.Since(start).Nanoseconds()
	start = time.Now()
	_, err = linker.Instantiate(ctx, compiled, vm.CallOptions{})
	out.Phases["instantiate"] = time.Since(start).Nanoseconds()
	if err != nil {
		return out, err
	}
	out.Usage, err = store.Usage()
	return out, err
}

func preparationCLI(args []string, encoder *json.Encoder, stderr io.Writer) int {
	flags := flag.NewFlagSet("wasmvm-prepare", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("wasm", "", "")
	expected := flags.String("expect", "valid", "")
	samples := flags.Int("samples", 7, "")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if *path == "" || flags.NArg() != 0 || *samples < 1 || *samples > 10000 || (*expected != "valid" && *expected != "malformed" && *expected != "invalid" && *expected != "limit") {
		fmt.Fprintln(stderr, "invalid preparation fixture options")
		return 1
	}
	start := time.Now()
	input, err := os.ReadFile(*path)
	elapsed := time.Since(start).Nanoseconds()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	out := preparationReport{SchemaVersion: 1, Engine: "wasmvm", SourceRevision: sourceRevision, Compiler: runtime.Compiler, CompilerRevision: compilerRevision, RuntimeVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Mode: "prepare", InputBytes: len(input), LoadNs: elapsed, Samples: []preparationSample{}}
	for i := 0; i < *samples; i++ {
		sample, err := prepareOnce(context.Background(), input)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if sample.Outcome != *expected {
			fmt.Fprintf(stderr, "preparation outcome: got %s, expected %s\n", sample.Outcome, *expected)
			return 1
		}
		out.Samples = append(out.Samples, sample)
	}
	if err := encoder.Encode(out); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
