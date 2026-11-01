package vm

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestSharedConcurrentEntriesObserveGrowthAndCancel(t *testing.T) {
	input, err := os.ReadFile("testdata/shared-concurrent.wasm")
	if err != nil {
		t.Fatal(err)
	}
	withBackends(t, func(t *testing.T, newEngine func() *Engine) {
		cfg := newEngine().config
		cfg.Profile, cfg.Features = "core3", wasm.Threads
		engine, err := NewEngineWithConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		compiled, err := engine.Compile(nil, input)
		if err != nil {
			t.Fatal(err)
		}
		defer compiled.Close()
		owner, err := engine.NewStore(DefaultStoreConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer owner.Close()
		memory, err := owner.NewMemory(MemoryType{Min: 1, Max: 3, HasMax: true, Shared: true})
		if err != nil {
			t.Fatal(err)
		}
		capability, err := memory.ShareMemory()
		if err != nil {
			t.Fatal(err)
		}
		defer capability.Close()
		instances := make([]*Instance, 2)
		memories := make([]Extern, 2)
		for i := range instances {
			store, err := engine.NewStore(DefaultStoreConfig())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { store.Close(); store.WaitClosed(nil) })
			ex, err := store.ImportMemory(capability)
			if err != nil {
				t.Fatal(err)
			}
			linker := NewLinker(store)
			if err := linker.Define("env", "memory", ex); err != nil {
				t.Fatal(err)
			}
			instances[i], memories[i] = instantiate(t, linker, compiled), ex
		}
		type outcome struct {
			result CallResult
			err    error
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		observer := export(t, instances[0], "observe-growth")
		done := make(chan outcome, 1)
		go func() {
			r, err := observer.Call(ctx, nil, CallOptions{})
			done <- outcome{r, err}
		}()
		waitWord := func(at uint64, predicate func(uint32) bool) {
			t.Helper()
			var bytes [4]byte
			for {
				if err := memory.ReadMemory(at, bytes[:]); err != nil {
					t.Fatal(err)
				}
				if predicate(binary.LittleEndian.Uint32(bytes[:])) {
					return
				}
				if err := ctx.Err(); err != nil {
					t.Fatal("shared guest did not make progress", err)
				}
				time.Sleep(100 * time.Microsecond)
			}
		}
		waitWord(0, func(v uint32) bool { return v == 1 })
		if old, ok, err := memory.GrowMemory(1); err != nil || !ok || old != 1 {
			t.Fatal("growth while native entry is active", old, ok, err)
		}
		if err := memory.WriteMemory(65536, []byte{123, 0, 0, 0}); err != nil {
			t.Fatal(err)
		}
		if _, err := export(t, instances[1], "signal").Call(ctx, []Value{I32(4), I32(1)}, CallOptions{}); err != nil {
			t.Fatal("simultaneous shared entry", err)
		}
		observed := <-done
		if observed.err != nil || bits(t, observed.result.Values[0]) != 125 {
			t.Fatal("entry retained stale size/data after synchronized growth", observed)
		}
		// Existing imports retain the backing after both creator roots close.
		owner.Close()
		if err := owner.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		capability.Close()
		memory = memories[1]
		var workers sync.WaitGroup
		errorsOut := make(chan error, 2)
		for _, instance := range instances {
			counter := export(t, instance, "count")
			workers.Add(1)
			go func() {
				defer workers.Done()
				_, err := counter.Call(ctx, []Value{I32(10000)}, CallOptions{})
				errorsOut <- err
			}()
		}
		workers.Wait()
		for range instances {
			if err := <-errorsOut; err != nil {
				t.Fatal(err)
			}
		}
		waitWord(8, func(v uint32) bool { return v == 20000 })
		counter := export(t, instances[0], "count")
		childCtx, stop := context.WithCancel(ctx)
		go func() {
			r, err := counter.Call(childCtx, []Value{I32(^uint32(0))}, CallOptions{})
			done <- outcome{r, err}
		}()
		waitWord(8, func(v uint32) bool { return v > 20000 })
		stop()
		if stopped := <-done; !errors.Is(stopped.err, context.Canceled) {
			t.Fatal("atomic loop cancellation", stopped.err)
		}
	})
}
