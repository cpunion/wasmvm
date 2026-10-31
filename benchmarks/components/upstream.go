//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	p2 "github.com/cpunion/wasmvm/pkg/wasi/preview2"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// Standard-library command entries may exit or trap and are one-shot. Report
// first execution separately from compilation, binding, lookup and teardown.
func main() {
	if len(os.Args) < 4 {
		panic("usage: upstream COMPONENT BACKEND PREOPEN [GUEST_ARGS...]")
	}
	start := time.Now()
	data, err := os.ReadFile(os.Args[1])
	must(err)
	load := time.Since(start)
	config := component.DefaultConfig()
	config.Core.Backend = os.Args[2]
	start = time.Now()
	engine, err := component.NewEngine(config)
	must(err)
	engineTime := time.Since(start)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start = time.Now()
	compiled, err := engine.Compile(ctx, data)
	must(err)
	compile := time.Since(start)
	start = time.Now()
	out, stderr := p2.NewOutputBuffer(1<<20), p2.NewOutputBuffer(1<<20)
	cwd := "/"
	world, err := p2.NewWorld(p2.WorldConfig{
		Environment: p2.EnvironmentConfig{Args: os.Args[4:], Env: []string{"frabjous=day", "callooh=callay"}, InitialCWD: &cwd},
		IO:          p2.IOConfig{Stdin: p2.NewBytesInput([]byte("So rested he by the Tumtum tree")), Stdout: out, Stderr: stderr},
		FileSystem:  p2.FileSystemConfig{Preopens: []p2.FileSystemPreopen{{GuestPath: "/", HostPath: os.Args[3]}}},
	})
	must(err)
	provider := time.Since(start)
	start = time.Now()
	in, err := compiled.Instantiate(ctx, world.Imports(), component.CallOptions{})
	must(err)
	instantiate := time.Since(start)
	start = time.Now()
	function, err := in.Function(p2.CLICommandInterface, "run")
	must(err)
	info := function.Info()
	resultType := info.Result()
	if info.ParameterCount() != 0 || resultType.Kind() != component.KindResult || resultType.CaseCount() != 2 {
		panic("incorrect standard command signature")
	}
	for i := 0; i < 2; i++ {
		if _, payload := resultType.Case(i); payload.Valid() {
			panic("command result payload")
		}
	}
	lookup := time.Since(start)
	start = time.Now()
	result, err := function.Call(ctx, nil, component.CallOptions{})
	first := time.Since(start)
	code := uint32(0)
	trapped := errors.Is(err, wasm.ErrTrap)
	var exit *p2.ExitError
	if errors.As(err, &exit) {
		code, err = exit.Code, nil
	} else if err == nil {
		code = result.Value.Case
	}
	stdout, stderrText := string(out.Drain()), string(stderr.Drain())
	start = time.Now()
	in.Close()
	world.Close()
	must(world.WaitClosed(nil))
	compiled.Close()
	engine.Close()
	closeTime := time.Since(start)
	row := map[string]any{"load_ns": load.Nanoseconds(), "engine_ns": engineTime.Nanoseconds(), "compile_ns": compile.Nanoseconds(), "provider_ns": provider.Nanoseconds(), "store_bind_instantiate_ns": instantiate.Nanoseconds(), "lookup_ns": lookup.Nanoseconds(), "first_ns": first.Nanoseconds(), "close_ns": closeTime.Nanoseconds(), "code": code, "trapped": trapped, "stdout": stdout, "stderr": stderrText}
	if err != nil {
		row["error"] = err.Error()
	}
	must(json.NewEncoder(os.Stdout).Encode(row))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
