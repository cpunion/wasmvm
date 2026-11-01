package decode

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestCustomPagesGateAndEncoding(t *testing.T) {
	for _, exponent := range []byte{0, 16, 63} {
		input := fixture(section(5, 1, 8, 0, exponent))
		if _, err := Decode(input, DefaultLimits()); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal("custom pages enabled by default", err)
		}
		m, err := DecodeWithFeatures(input, DefaultLimits(), wasm.CustomPageSizes)
		if err != nil || m.Memories[0].PageSize != uint64(1)<<exponent {
			t.Fatal(exponent, m, err)
		}
	}
	for _, payload := range [][]byte{
		{1, 8, 0}, {1, 8, 0, 64}, {1, 8, 0, 0xff, 0xff, 0xff, 0xff, 0x0f},
	} {
		if _, err := DecodeWithFeatures(fixture(section(5, payload...)), DefaultLimits(), wasm.CustomPageSizes); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal(payload, err)
		}
	}
	// Limits remain u64 in the proposal encoding even for i32 addressing;
	// semantic validation subsequently rejects counts outside the address space.
	m, err := DecodeWithFeatures(fixture(section(5, 1, 8, 0x80, 0x80, 0x80, 0x80, 0x10, 0)), DefaultLimits(), wasm.CustomPageSizes)
	if err != nil || m.Memories[0].Min != 1<<32 {
		t.Fatal(m, err)
	}
	if _, err := DecodeWithFeatures(fixture(section(4, 1, 0x70, 8, 0, 0)), DefaultLimits(), wasm.CustomPageSizes); !errors.Is(err, wasm.ErrMalformed) {
		t.Fatal("custom page-size flag on a table", err)
	}
}
