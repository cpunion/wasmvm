//go:build ignore

// Run with go run (or the pinned LLGo profile) from the repository root.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
)

func main() {
	if len(os.Args) != 4 {
		panic("expected fixture backend positive-count")
	}
	count, err := strconv.Atoi(os.Args[3])
	if err != nil || count < 1 {
		panic("invalid count")
	}
	start := time.Now()
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	load := time.Since(start).Nanoseconds()
	cfg := component.DefaultConfig()
	cfg.Core.Backend = os.Args[2]
	start = time.Now()
	e, err := component.NewEngine(cfg)
	if err != nil {
		panic(err)
	}
	engine := time.Since(start).Nanoseconds()
	defer e.Close()
	start = time.Now()
	m, err := e.Compile(nil, data)
	compile := time.Since(start).Nanoseconds()
	out := map[string]any{"fixture": os.Args[1], "backend": os.Args[2], "count": count, "load_ns": load, "engine_ns": engine, "compile_ns": compile, "source_compile_error": nil}
	if err != nil {
		out["source_compile_error"] = err.Error()
		json.NewEncoder(os.Stdout).Encode(out)
		return
	}
	defer m.Close()
	imports := component.Imports{}
	var instantiate, value, drop, close int64
	for i := 0; i <= count; i++ {
		start = time.Now()
		in, err := m.Instantiate(nil, imports, component.CallOptions{})
		a := time.Since(start).Nanoseconds()
		if err != nil {
			panic(err)
		}
		start = time.Now()
		v, err := in.Value(nil, "output")
		b := time.Since(start).Nanoseconds()
		if err != nil {
			panic(err)
		}
		start = time.Now()
		_, err = in.DropResource(nil, v.Resource, component.CallOptions{})
		c := time.Since(start).Nanoseconds()
		if err != nil {
			panic(err)
		}
		start = time.Now()
		in.Close()
		d := time.Since(start).Nanoseconds()
		if i == 0 {
			out["first_instantiate_including_store_and_start_ns"] = a
			out["first_value_ns"] = b
			out["first_drop_ns"] = c
			out["first_close_ns"] = d
		} else {
			instantiate += a
			value += b
			drop += c
			close += d
		}
	}
	for k, v := range map[string]int64{"instantiate_including_store_and_start_ns": instantiate, "export_value_ns": value, "explicit_drop_ns": drop, "close_ns": close} {
		out[k] = float64(v) / float64(count)
	}
	out["semantic"] = true
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(fmt.Sprint(err))
	}
}
