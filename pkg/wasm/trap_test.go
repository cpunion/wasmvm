package wasm

import (
	"errors"
	"strings"
	"testing"
)

func TestTrapCodes(t *testing.T) {
	seen := make(map[string]bool)
	for code := NoTrap; code <= TrapNullReference; code++ {
		s := code.String()
		if s == "" || s == "unknown trap" || seen[s] {
			t.Fatal(code, s)
		}
		seen[s] = true
	}
	if TrapCode(255).String() != "unknown trap" {
		t.Fatal("unknown code")
	}
}

func TestTrapMetadata(t *testing.T) {
	var err error = &Trap{Code: TrapMemoryBounds, Function: 7, Offset: 0x42}
	var trap *Trap
	if !errors.Is(err, ErrTrap) || !errors.As(err, &trap) || trap.Code != TrapMemoryBounds || !strings.Contains(err.Error(), "function 7 byte 0x42") {
		t.Fatal(err)
	}
}
