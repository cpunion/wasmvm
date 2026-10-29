package decode

import (
	"errors"
	"testing"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestTailGrammarIsExplicitAndBounded(t *testing.T) {
	for _, body := range [][]byte{{0x12, 0, 0xb}, {0x13, 0, 0, 0xb}} {
		input := function(body...)
		if _, err := Decode(input, DefaultLimits()); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal(err)
		}
		m, err := DecodeWithFeatures(input, DefaultLimits(), wasm.TailCall)
		if err != nil || m.Functions[0].Body[0].Op != uint32(body[0]) {
			t.Fatal(m, err)
		}
		limits := DefaultLimits()
		limits.MaxNodes = 0
		if _, err := DecodeWithFeatures(input, limits, wasm.TailCall); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	for _, body := range [][]byte{{0x12}, {0x12, 0x80}, {0x13, 0}, {0x13, 0, 0x80}} {
		if _, err := DecodeWithFeatures(function(body...), DefaultLimits(), wasm.TailCall); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal(body, err)
		}
	}
	if _, err := DecodeWithFeatures(magic, DefaultLimits(), wasm.Features(1<<63)); !errors.Is(err, wasm.ErrUnsupported) {
		t.Fatal(err)
	}
}
