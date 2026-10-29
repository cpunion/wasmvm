package decode

import (
	"encoding/binary"
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

func TestMemory64LimitsAndDisplacementGrammar(t *testing.T) {
	for _, limits := range [][]byte{{4, 0}, {5, 0, 0}, {4, 0x80, 0x80, 0x80, 0x80, 0x10}} {
		input := fixture(section(5, append([]byte{1}, limits...)...))
		if _, err := Decode(input, DefaultLimits()); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal("64-bit limits enabled by default", err)
		}
		m, err := DecodeWithFeatures(input, DefaultLimits(), wasm.Memory64)
		if err != nil || !m.Memories[0].Index64 {
			t.Fatal(m, err)
		}
	}
	for _, limits := range [][]byte{{2, 0}, {3, 0, 1}, {6, 0}, {7, 0, 1}, {8, 0}, {4, 0x80}} {
		input := fixture(section(5, append([]byte{1}, limits...)...))
		if _, err := DecodeWithFeatures(input, DefaultLimits(), wasm.Memory64); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal("reserved/shared/truncated limits", limits, err)
		}
	}
	for _, offset := range []uint64{0, 1 << 32, 1 << 63, ^uint64(0)} {
		var b [10]byte
		n := binary.PutUvarint(b[:], offset)
		body := append([]byte{0x28, 2}, b[:n]...)
		m, err := DecodeWithFeatures(function(append(body, 0xb)...), DefaultLimits(), wasm.Memory64)
		if err != nil || m.Functions[0].Body[0].MemoryOffset() != offset {
			t.Fatal(offset, m, err)
		}
	}
}

func TestRelaxedSIMDGrammarIsExplicitAndBounded(t *testing.T) {
	for op := uint32(0x100); op <= 0x12f; op++ {
		body := append([]byte{0xfd}, uleb(op)...)
		input := function(append(body, 0xb)...)
		if _, err := Decode(input, DefaultLimits()); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal("extension enabled by default", op, err)
		}
		m, err := DecodeWithFeatures(input, DefaultLimits(), wasm.RelaxedSIMD)
		if op > 0x113 {
			if !errors.Is(err, wasm.ErrMalformed) {
				t.Fatal("reserved relaxed opcode", op, err)
			}
		} else if err != nil || m.Functions[0].Body[0].Op != 0xfd0000|op {
			t.Fatal(op, err)
		}
	}
	for _, body := range [][]byte{{0xfd}, {0xfd, 0x80}, {0xfd, 0x80, 0x80, 0x80, 0x80, 0x10}} {
		if _, err := DecodeWithFeatures(function(body...), DefaultLimits(), wasm.RelaxedSIMD); !errors.Is(err, wasm.ErrMalformed) {
			t.Fatal(body, err)
		}
	}
}
