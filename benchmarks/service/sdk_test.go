//go:build !llgo && (wasmtime || wasmer || wasmedge)

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func sdkOptions() options {
	args := nativeOptions()
	args.engine, _ = sdkIdentity()
	args.backend = "compiler"
	if args.engine == "wasmedge" {
		args.backend = "interpreter"
	}
	return args
}

func TestSDKIndependentWorkersChecksumsAndLifecycleFailures(t *testing.T) {
	args := sdkOptions()
	for _, name := range []string{"control", "reset-memory"} {
		args.path = referenceFixture(name)
		args.seed, args.iterations, args.Expected = 0xffffffff, 1, 0xffffffff
		out, err := benchmark(context.Background(), args, nil)
		if err != nil {
			t.Fatal(args.engine, name, err)
		}
		checkBatches(t, out, args)
		for _, phase := range []string{"load", "engineSetup", "compileValidate", "instantiate", "exportLookup", "firstCall", "cleanup"} {
			// Sub-microsecond export lookup may round to zero in the SDK clock.
			if ns, ok := out.Phases[phase]; !ok || ns < 0 || (phase != "exportLookup" && ns == 0) {
				t.Fatal(phase, out.Phases)
			}
		}
		args.Expected++
		if _, err := benchmark(context.Background(), args, nil); err == nil {
			t.Fatal("corrupt checksum accepted")
		}
	}
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
	// This passes the pure-module envelope but fails independent SDK validation.
	if err := os.WriteFile(args.path, []byte{0, 'a', 's', 'm', 1, 0, 0, 0, 1, 1, 0xff}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("invalid type section accepted")
	}
	args.path += ".missing"
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("missing path accepted")
	}
	args.engine = "different-sdk"
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("different SDK accepted")
	}
}

func TestSDKBothIndependentSHAProducers(t *testing.T) {
	for _, producer := range []string{"clang", "rust"} {
		for _, v := range []struct{ seed, count, want uint32 }{{17, 64, 3381501185}, {0xffffffff, 3, 2854132502}} {
			args := sdkOptions()
			args.path = filepath.Join("..", "applications", "fixtures", "sha256-"+producer+".wasm")
			args.seed, args.iterations, args.Expected = v.seed, v.count, v.want
			out, err := benchmark(context.Background(), args, nil)
			if err != nil {
				t.Fatal(producer, err)
			}
			checkBatches(t, out, args)
		}
	}
}

func TestSDKWorkersOwnMemoryAndGlobals(t *testing.T) {
	args := sdkOptions()
	args.path, args.Expected = filepath.Join("testdata", "worker-state.wasm"), 0
	out := report{Phases: make(map[string]int64)}
	calls, release, err := sdkKernels(args, &out)
	if release != nil {
		defer release()
	}
	if err != nil {
		t.Fatal(err)
	}
	checkIndependentWorkerState(t, calls)
}
