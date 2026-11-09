//go:build llgo || (!wasmtime && !wasmer && !wasmedge)

package main

import (
	"context"
	"strings"
	"testing"
)

func TestUnavailableSDKNeverFallsBack(t *testing.T) {
	args := nativeOptions()
	args.engine, args.backend, args.path = "wasmtime", "compiler", "unused"
	if _, err := benchmark(context.Background(), args, nil); err == nil || !strings.Contains(err.Error(), "requires a Go build") {
		t.Fatal(err)
	}
}
