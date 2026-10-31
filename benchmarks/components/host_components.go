//go:build ignore

package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
)

func main() {
	if len(os.Args) != 5 {
		panic("expected consumer guest backend positive-count")
	}
	count, err := strconv.Atoi(os.Args[4])
	if err != nil || count < 1 {
		panic("invalid count")
	}
	start := time.Now()
	consumer, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	guest, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	load := time.Since(start).Nanoseconds()
	cfg := component.DefaultConfig()
	cfg.Core.Backend = os.Args[3]
	start = time.Now()
	e, err := component.NewEngine(cfg)
	if err != nil {
		panic(err)
	}
	engine := time.Since(start).Nanoseconds()
	defer e.Close()
	start = time.Now()
	code, err := e.Compile(nil, guest)
	if err != nil {
		panic(err)
	}
	guestCompile := time.Since(start).Nanoseconds()
	defer code.Close()
	start = time.Now()
	m, err := e.Compile(nil, consumer)
	if err != nil {
		panic(err)
	}
	consumerCompile := time.Since(start).Nanoseconds()
	defer m.Close()
	imports := component.Imports{Components: map[string]*component.Compiled{"code": code}}
	var setup, closeTime, firstSetup int64
	var in *component.Instance
	for i := 0; i <= count; i++ {
		start = time.Now()
		next, err := m.Instantiate(nil, imports, component.CallOptions{})
		n := time.Since(start).Nanoseconds()
		if err != nil {
			panic(err)
		}
		if i == 0 {
			in = next
			firstSetup = n
			continue
		}
		setup += n
		start = time.Now()
		next.Close()
		closeTime += time.Since(start).Nanoseconds()
	}
	start = time.Now()
	f, err := in.Function("add")
	if err != nil {
		panic(err)
	}
	lookup := time.Since(start).Nanoseconds()
	args := []component.Value{{}, {Bits: 5}}
	start = time.Now()
	got, err := f.Call(nil, args, component.CallOptions{})
	first := time.Since(start).Nanoseconds()
	if err != nil || got.Value.Bits != 5 {
		panic("first result mismatch")
	}
	var checksum uint64
	start = time.Now()
	for i := 0; i < count; i++ {
		args[0].Bits = uint64(uint32(i))
		v, err := f.Call(nil, args, component.CallOptions{})
		if err != nil || v.Value.Bits != uint64(uint32(i)+5) {
			panic("warm result mismatch")
		}
		checksum += v.Value.Bits
	}
	warm := float64(time.Since(start).Nanoseconds()) / float64(count)
	start = time.Now()
	in.Close()
	finalClose := time.Since(start).Nanoseconds()
	out := map[string]any{"load_ns": load, "engine_ns": engine, "guest_compile_ns": guestCompile, "consumer_compile_ns": consumerCompile, "first_store_bind_instantiate_ns": firstSetup, "store_bind_instantiate_ns": float64(setup) / float64(count), "instance_close_ns": float64(closeTime) / float64(count), "lookup_ns": lookup, "first_ns": first, "public_warm_ns": warm, "final_close_ns": finalClose, "count": count, "checksum": checksum, "semantic": true, "backend": os.Args[3]}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}
