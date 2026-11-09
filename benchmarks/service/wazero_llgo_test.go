//go:build llgo

package main

import (
	"context"
	"testing"
)

func TestLLGoReferenceRejectsWazeroWithoutSubstitution(t *testing.T) {
	args := nativeOptions()
	args.engine = "wazero"
	if _, err := benchmark(context.Background(), args, nil); err == nil {
		t.Fatal("LLGo adapter silently substituted a runtime for wazero")
	}
}
