//go:build !p3reference && !p3host

package main

import (
	"context"
	"testing"
)

// This is a separate CPU/allocation diagnostic, including request construction
// and validation. It does not replace the fixed public-API/service event.
func BenchmarkOwnedP3Complete(b *testing.B) {
	args := largeOptions()
	args.engine, args.backend, args.path = "wasmvm", "native", p3Fixture("world-rust-http")
	args.Workers = 1
	out := report{Phases: map[string]int64{}}
	calls, release, err := ownedKernels(context.Background(), args, &out)
	if release != nil {
		defer func() {
			if err := release(); err != nil {
				b.Error(err)
			}
		}()
	}
	if err != nil {
		b.Fatal(err)
	}
	for range 10 {
		if _, _, err := calls[0](); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, value, err := calls[0](); err != nil || value != args.Expected {
			b.Fatal(value, err)
		}
	}
	b.StopTimer()
}
