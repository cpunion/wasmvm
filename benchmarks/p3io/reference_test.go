//go:build p3reference && !llgo

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReferenceCompleteP3BodyTrailersAndIndependentStores(t *testing.T) {
	for _, fixture := range []string{"world-rust-http", "world-rust-composed-http"} {
		for _, args := range []options{nativeOptions(), largeOptions()} {
			args.engine, args.backend, args.path = "wasmtime", "compiler", p3Fixture(fixture)
			out, err := benchmark(nil, args, nil)
			if err != nil {
				t.Fatal(fixture, len(args.payload), err)
			}
			checkBatches(t, out, args)
			for _, phase := range []string{"load", "engineSetup", "compileValidate", "providerSetup", "instantiate", "exportLookup", "executorSetup"} {
				if _, ok := out.Phases[phase]; !ok {
					t.Fatal(phase, out.Phases)
				}
			}
		}
	}
}

func TestReferenceRejectsSetupAndResponseFailuresWithoutFallback(t *testing.T) {
	args := nativeOptions()
	args.engine, args.backend = "wasmtime", "compiler"
	for _, fixture := range []string{"missing-file", "world-c-command", "command-success", "http-server"} {
		args.path = p3Fixture(fixture)
		if _, err := benchmark(nil, args, nil); err == nil {
			t.Fatal(fixture)
		}
	}
	args.path = filepath.Join(t.TempDir(), "invalid.wasm")
	for _, data := range [][]byte{nil, []byte("invalid")} {
		if err := os.WriteFile(args.path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := benchmark(nil, args, nil); err == nil {
			t.Fatal("invalid binary accepted")
		}
	}
	args.engine, args.backend, args.path = "wasmvm", "native", p3Fixture("world-rust-http")
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("owned VM silently included in reference")
	}
}
