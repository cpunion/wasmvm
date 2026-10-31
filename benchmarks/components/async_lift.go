//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"github.com/cpunion/wasmvm/pkg/component"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 6 {
		panic("usage: async_lift COMPONENT BACKEND CALLS BATCH ENTRY")
	}
	count, err := strconv.ParseUint(os.Args[3], 10, 32)
	must(err)
	batch, err := strconv.ParseUint(os.Args[4], 10, 32)
	must(err)
	if count == 0 || batch == 0 || count > 100000 || batch > 100000 {
		panic("count outside bounds")
	}
	start := time.Now()
	bytes, err := os.ReadFile(os.Args[1])
	must(err)
	load := time.Since(start)
	cfg := component.DefaultConfig()
	cfg.Core.Backend = os.Args[2]
	start = time.Now()
	engine, err := component.NewEngine(cfg)
	must(err)
	engineTime := time.Since(start)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	start = time.Now()
	module, err := engine.Compile(ctx, bytes)
	must(err)
	compile := time.Since(start)
	start = time.Now()
	in, err := module.Instantiate(ctx, component.Imports{}, component.CallOptions{})
	must(err)
	instantiate := time.Since(start)
	start = time.Now()
	f, err := in.Function(os.Args[5])
	must(err)
	lookup := time.Since(start)
	args := []component.Value{{Bits: batch}}
	expected := batch + 42
	if os.Args[5] == "churn" {
		expected = batch
	}
	callbacks := uint64(0)
	if os.Args[5] == "run" {
		expected = 42
		callbacks = batch + 1
	}
	if os.Args[5] == "stackful" {
		callbacks = 1
	}
	call := func() {
		v, err := f.Call(ctx, args, component.CallOptions{})
		must(err)
		if v.Value.Bits != expected {
			panic("wrong async task result")
		}
	}
	start = time.Now()
	call()
	first := time.Since(start)
	for range 10 {
		call()
	}
	start = time.Now()
	for n := uint64(0); n < count; n++ {
		call()
	}
	warm := float64(time.Since(start).Nanoseconds()) / float64(count)
	start = time.Now()
	in.Close()
	module.Close()
	engine.Close()
	closeTime := time.Since(start)
	must(json.NewEncoder(os.Stdout).Encode(map[string]any{"load_ns": load.Nanoseconds(), "engine_ns": engineTime.Nanoseconds(), "compile_ns": compile.Nanoseconds(), "store_instantiate_ns": instantiate.Nanoseconds(), "lookup_ns": lookup.Nanoseconds(), "first_ns": first.Nanoseconds(), "warm_call_ns": warm, "callbacks_per_call": callbacks, "calls": count, "executor_ns": 0, "close_ns": closeTime.Nanoseconds()}))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
