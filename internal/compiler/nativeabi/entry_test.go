package nativeabi

import (
	"bytes"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"testing"
)

func TestEntryLayoutAndLimits(t *testing.T) {
	for _, arch := range []string{"amd64", "arm64"} {
		body := []byte{1, 2, 3, 4}
		entry, err := Wrap(arch, body, 4096)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(entry.Code[entry.BodyOffset:entry.ExitOffset], body) {
			t.Fatal("body relocated incorrectly")
		}
		body[0] = 99
		if entry.Code[entry.BodyOffset] != 1 {
			t.Fatal("body not copied")
		}
		if arch == "amd64" && (entry.StackBytes != 56 || entry.Code[len(entry.Code)-1] != 0xc3) {
			t.Fatal(entry)
		}
		if arch == "arm64" && (entry.StackBytes != 224 || !bytes.Equal(entry.Code[len(entry.Code)-4:], []byte{0xc0, 3, 0x5f, 0xd6})) {
			t.Fatal(entry)
		}
		if _, err := Wrap(arch, body, uint64(len(entry.Code))); err != nil {
			t.Fatal(err)
		}
		if _, err := Wrap(arch, body, uint64(len(entry.Code)-1)); err != wasm.ErrLimit {
			t.Fatal(err)
		}
	}
	if _, err := Wrap("arm64", []byte{1}, 4096); err != wasm.ErrInvalid {
		t.Fatal(err)
	}
	if _, err := Wrap("riscv64", nil, 4096); err != wasm.ErrUnsupported {
		t.Fatal(err)
	}
}
