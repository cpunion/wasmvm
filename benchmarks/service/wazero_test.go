//go:build !llgo

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func referenceFixture(name string) string {
	return filepath.Join("..", "embedded", "wazero", "testdata", name+".wasm")
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
