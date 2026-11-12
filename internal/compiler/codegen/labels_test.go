package codegen

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestDenseLabelsKeepUnboundAndZeroTargetsDistinct(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		a := assembler{arch: arch}
		zero, forward := a.label(), a.label()
		a.mark(zero)
		// Grow beyond the initial backing allocation without binding forward.
		for range 1024 {
			a.label()
		}
		if position, ok := a.labelAt(zero); !ok || position != 0 {
			t.Fatal("lost a bound target at offset zero", position, ok)
		}
		for _, id := range []int{-1, forward, 1 << 24} {
			if position, ok := a.labelAt(id); ok || position != 0 {
				t.Fatal("unbound label became a valid target", id, position, ok)
			}
		}
		a.branch(forward, false, 0)
		if _, err := a.finish(); !errors.Is(err, wasm.ErrInvalid) {
			t.Fatal("unbound forward branch was accepted", err)
		}
		a.mark(forward)
		if _, err := a.finish(); err != nil {
			t.Fatal("binding a forward label did not resolve the branch", err)
		}
	}
}
