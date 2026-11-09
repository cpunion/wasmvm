//go:build !llgo

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cpunion/wasmvm/internal/bench/service"
)

func referenceFixture(name string) string {
	return filepath.Join("..", "embedded", "wazero", "testdata", name+".wasm")
}

func checkIndependentWorkerState(t *testing.T, calls []service.Call) {
	t.Helper()
	// Every worker starts with zero. Later calls must retain its own state;
	// the fixed checksum callback deliberately reports the changed value.
	for round := range 3 {
		for _, call := range calls {
			_, value, err := call()
			if value != uint32(round*2) || (err == nil) != (round == 0) {
				t.Fatal("worker state is shared or reset", round, value, err)
			}
		}
	}
}

func TestWazeroWorkersOwnMemoryAndGlobals(t *testing.T) {
	for _, backend := range []string{"compiler", "interpreter"} {
		args := nativeOptions()
		args.engine, args.backend, args.path, args.Expected = "wazero", backend, filepath.Join("testdata", "worker-state.wasm"), 0
		out := report{Phases: make(map[string]int64)}
		calls, release, err := wazeroKernels(context.Background(), args, &out)
		if release != nil {
			defer release()
		}
		if err != nil {
			t.Fatal(err)
		}
		checkIndependentWorkerState(t, calls)
	}
}

func TestWazeroSharedCompilationIndependentInstances(t *testing.T) {
	for _, backend := range []string{"interpreter", "compiler"} {
		for _, termination := range []bool{false, true} {
			for _, fixture := range []string{"control", "reset-memory"} {
				args := nativeOptions()
				args.engine, args.backend, args.path, args.termination = "wazero", backend, referenceFixture(fixture), termination
				args.seed, args.iterations, args.Expected = 0xffffffff, 1, 0xffffffff
				out, err := benchmark(context.Background(), args, nil)
				if err != nil {
					t.Fatal(backend, termination, err)
				}
				checkBatches(t, out, args)
				args.Expected = 0
				if _, err := benchmark(context.Background(), args, nil); err == nil {
					t.Fatal("corrupt result accepted")
				}
			}
		}
	}
}

func TestWazeroRejectsImportsStartsSignaturesTrapsAndInvalidModules(t *testing.T) {
	args := nativeOptions()
	args.engine, args.backend = "wazero", "compiler"
	for _, name := range []string{"missing", "import", "start", "wrong-signature", "trap", "instantiate-failure"} {
		args.path = referenceFixture(name)
		if _, err := benchmark(context.Background(), args, nil); err == nil {
			t.Fatal(name)
		}
	}
	args.path = filepath.Join(t.TempDir(), "invalid.wasm")
	if err := os.WriteFile(args.path, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("invalid module accepted")
	}
	args.path, args.termination = referenceFixture("control"), true
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := benchmark(ctx, args, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, input := range [][]byte{
		nil, []byte("invalid"), append([]byte{0, 'a', 's', 'm', 1, 0, 0, 0}, 0, 0x80),
		append([]byte{0, 'a', 's', 'm', 1, 0, 0, 0}, 0, 4),
		append([]byte{0, 'a', 's', 'm', 1, 0, 0, 0}, 2, 1, 0x80),
	} {
		if pureModule(input) == nil {
			t.Fatal("invalid pure-module contract accepted")
		}
	}
}
