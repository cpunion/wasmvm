// Package cli implements the core command line using only the public VM API.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const usage = `usage: wasmvm <validate|inspect|run> [options] module.wasm [type:value ...]

Options precede the module path. All commands accept:
  --backend name         interpreter (default) or native (supported hosts)
  --profile core2        Base binary profile
  --extended-const       Enable extended constant expressions
  --tail-call            Enable proper direct and indirect tail calls
  --multi-memory         Enable multiple private memories
  --relaxed-simd         Enable all twenty relaxed SIMD instructions
  --function-references  Enable typed function references and non-null references
  --memory64             Enable 64-bit memory and table indices
  --timeout duration     Overall deadline (default 30s; 0 disables)
Run also accepts:
  --export name          Required exported function name
  --fuel count           Shared instruction budget, including the start function

Argument types: i32, i64, f32, f64, f32bits, f64bits, v128, funcref, externref.
Integers accept decimal or 0x hex; v128 uses low64,high64; references accept null.
Output is JSON. Core run installs no WASI or host imports.
`

// Main returns an exit code: 0 success, 1 VM/I/O failure, 2 command misuse.
func Main(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	command := args[0]
	if command != "validate" && command != "inspect" && command != "run" {
		fmt.Fprintf(stderr, "unknown command %q\n%s", command, usage)
		return 2
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	backend := flags.String("backend", "interpreter", "")
	profile := flags.String("profile", "core2", "")
	extendedConst := flags.Bool("extended-const", false, "")
	tailCall := flags.Bool("tail-call", false, "")
	multiMemory := flags.Bool("multi-memory", false, "")
	relaxedSIMD := flags.Bool("relaxed-simd", false, "")
	functionReferences := flags.Bool("function-references", false, "")
	memory64 := flags.Bool("memory64", false, "")
	timeout := flags.Duration("timeout", 30*time.Second, "")
	var exported, fuel string
	if command == "run" {
		flags.StringVar(&exported, "export", "", "")
		flags.StringVar(&fuel, "fuel", "", "")
	}
	flags.Usage = func() { fmt.Fprint(stderr, usage) }
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	positional := flags.Args()
	if len(positional) == 0 || command != "run" && len(positional) != 1 || command == "run" && exported == "" || *timeout < 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	options := vm.CallOptions{}
	if fuel != "" {
		n, err := strconv.ParseUint(fuel, 10, 64)
		if err != nil {
			fmt.Fprintln(stderr, "invalid fuel:", err)
			return 2
		}
		options.Metered = true
		options.Fuel = n
	}
	values := make([]vm.Value, 0, len(positional)-1)
	for _, arg := range positional[1:] {
		v, err := parseValue(arg)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		values = append(values, v)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}
	config := vm.DefaultEngineConfig()
	config.Backend = *backend
	config.Profile = *profile
	if *extendedConst {
		config.Features |= wasm.ExtendedConst
	}
	if *tailCall {
		config.Features |= wasm.TailCall
	}
	if *multiMemory {
		config.Features |= wasm.MultiMemory
	}
	if *relaxedSIMD {
		config.Features |= wasm.RelaxedSIMD
	}
	if *functionReferences {
		config.Features |= wasm.FunctionReferences
	}
	if *memory64 {
		config.Features |= wasm.Memory64
	}
	engine, err := vm.NewEngineWithConfig(config)
	if err != nil {
		return failure(stderr, err)
	}
	defer engine.Close()
	input, err := readModule(positional[0], config.Decode.MaxBytes)
	if err != nil {
		return failure(stderr, err)
	}
	switch command {
	case "validate":
		if err = engine.Validate(ctx, input); err != nil {
			return failure(stderr, err)
		}
		return output(stdout, stderr, map[string]any{"valid": true, "backend": *backend, "profile": *profile, "features": config.Features})
	case "inspect":
		info, err := engine.Inspect(ctx, input)
		if err != nil {
			return failure(stderr, err)
		}
		metadata, err := describe(info)
		if err != nil {
			return failure(stderr, err)
		}
		return output(stdout, stderr, metadata)
	default:
		compiled, err := engine.Compile(ctx, input)
		if err != nil {
			return failure(stderr, err)
		}
		defer compiled.Close()
		store, err := engine.NewStore(vm.DefaultStoreConfig())
		if err != nil {
			return failure(stderr, err)
		}
		defer store.Close()
		instance, err := vm.NewLinker(store).Instantiate(ctx, compiled, options)
		if err != nil {
			return failure(stderr, err)
		}
		fn, err := instance.Export(exported)
		if err != nil {
			return failure(stderr, err)
		}
		// Instantiation and invocation are separate public operations. Account
		// for start fuel in a single CLI budget through the returned receipt.
		if options.Metered {
			options.Fuel = instance.StartFuelRemaining()
		}
		result, err := fn.Call(ctx, values, options)
		if err != nil {
			return failure(stderr, err)
		}
		out := make([]resultValue, len(result.Values))
		for i, v := range result.Values {
			out[i] = formatValue(v)
			v.Release()
		}
		response := map[string]any{"backend": *backend, "profile": *profile, "features": config.Features, "export": exported, "results": out}
		if options.Metered {
			response["fuelRemaining"] = result.FuelRemaining
		}
		return output(stdout, stderr, response)
	}
}
func readModule(path string, maximum uint64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, int64(maximum)+1))
	if err != nil {
		return nil, err
	}
	if uint64(len(data)) > maximum {
		return nil, &wasm.Error{Kind: wasm.ErrLimit, Detail: "module input bytes"}
	}
	return data, nil
}
func failure(stderr io.Writer, err error) int { fmt.Fprintln(stderr, err); return 1 }
func output(stdout, stderr io.Writer, v any) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return failure(stderr, err)
	}
	return 0
}

// Inspection repeats signatures in JSON. Bound that expansion before allocating
// strings or encoding: many exports can share one large binary type definition.
const maxInspectionTypeValues = 1 << 20

func describe(info vm.ModuleInfo) (map[string]any, error) {
	remaining := uint64(maxInspectionTypeValues)
	charge := func(t vm.ExternType) bool {
		if t.Kind != vm.FunctionKind {
			return true
		}
		n := uint64(len(t.Function.Params)) + uint64(len(t.Function.Results))
		if n > remaining {
			return false
		}
		remaining -= n
		return true
	}
	for _, im := range info.Imports {
		if !charge(im.ExternType) {
			return nil, &wasm.Error{Kind: wasm.ErrLimit, Detail: "expanded inspection signature values"}
		}
	}
	for _, ex := range info.Exports {
		if !charge(ex.ExternType) {
			return nil, &wasm.Error{Kind: wasm.ErrLimit, Detail: "expanded inspection signature values"}
		}
	}
	imports := make([]map[string]any, 0, len(info.Imports))
	exports := make([]map[string]any, 0, len(info.Exports))
	for _, im := range info.Imports {
		d := describeType(im.ExternType)
		d["module"] = im.Module
		d["name"] = im.Name
		imports = append(imports, d)
	}
	for _, ex := range info.Exports {
		d := describeType(ex.ExternType)
		d["name"] = ex.Name
		exports = append(exports, d)
	}
	return map[string]any{"backend": info.Backend, "profile": info.Profile, "features": info.Features, "definedFunctions": info.Functions, "hasStart": info.HasStart, "imports": imports, "exports": exports}, nil
}
func describeType(t vm.ExternType) map[string]any {
	result := make(map[string]any)
	limits := func(l vm.MemoryType) {
		result["min"] = l.Min
		if l.HasMax {
			result["max"] = l.Max
		}
	}
	switch t.Kind {
	case vm.FunctionKind:
		result["kind"] = "function"
		params := make([]string, len(t.Function.Params))
		returns := make([]string, len(t.Function.Results))
		for i, v := range t.Function.Params {
			params[i] = v.String()
		}
		for i, v := range t.Function.Results {
			returns[i] = v.String()
		}
		result["params"] = params
		result["results"] = returns
	case vm.TableKind:
		result["kind"] = "table"
		result["element"] = t.Table.Element.String()
		limits(t.Table.Limits)
	case vm.MemoryKind:
		result["kind"] = "memory"
		limits(*t.Memory)
	case vm.GlobalKind:
		result["kind"] = "global"
		result["type"] = t.Global.Value.String()
		result["mutable"] = t.Global.Mutable
	}
	return result
}
