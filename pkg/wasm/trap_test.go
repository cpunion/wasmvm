package wasm

import "testing"

func TestTrapCodes(t *testing.T) {
	seen := make(map[string]bool)
	for code := NoTrap; code <= TrapInvalidConversion; code++ {
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
