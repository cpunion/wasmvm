//go:build !p3reference && !p3host

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOwnedCompleteGuestBodyTrailerAndComposedOwnership(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		for _, fixture := range []string{"world-rust-http", "world-rust-composed-http"} {
			args := nativeOptions()
			args.engine, args.backend, args.path = "wasmvm", backend, p3Fixture(fixture)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			out, err := benchmark(ctx, args, nil)
			cancel()
			if err != nil {
				t.Fatal(backend, fixture, err)
			}
			checkBatches(t, out, args)
			for _, phase := range []string{"load", "engineSetup", "compileValidate", "providerSetup", "instantiate", "exportLookup"} {
				if _, ok := out.Phases[phase]; !ok {
					t.Fatal(phase, out.Phases)
				}
			}
		}
	}
	for _, backend := range []string{"native", "interpreter"} {
		args := largeOptions()
		args.engine, args.backend, args.path = "wasmvm", backend, p3Fixture("world-rust-http")
		if out, err := benchmark(nil, args, nil); err != nil {
			t.Fatal(err)
		} else {
			checkBatches(t, out, args)
		}
	}
}

func TestOwnedSetupCancellationAndMissingHandlerCleanup(t *testing.T) {
	args := nativeOptions()
	args.engine, args.backend = "wasmvm", "native"
	for _, fixture := range []string{"missing-file", "world-c-command", "command-success"} {
		args.path = p3Fixture(fixture)
		if _, err := benchmark(nil, args, nil); err == nil {
			t.Fatal(fixture)
		}
	}
	args.path = filepath.Join(t.TempDir(), "invalid.wasm")
	if err := os.WriteFile(args.path, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := benchmark(nil, args, nil); err == nil {
		t.Fatal("invalid binary accepted")
	}
	args.path = p3Fixture("world-rust-http")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := benchmark(ctx, args, nil); err == nil {
		t.Fatal("cancelled setup accepted")
	}
}
