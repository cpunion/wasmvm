package memory

import (
	"bytes"
	"testing"

	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestPackedVectorLoadsMixedLanes(t *testing.T) {
	m := &Memory{bytes: make([]byte, 17)}
	for _, tc := range loads() {
		if tc.op < 0xfd0001 || tc.op > 0xfd000a {
			continue
		}
		check := func() {
			t.Helper()
			got, trap := m.Load(tc.op, 1, 0, 0, value.Value{})
			want := expectedLoad(tc, m.bytes[1:1+tc.size], nil, 0)
			if trap != wasm.NoTrap || !bytes.Equal(vectorBytes(got), want) {
				t.Fatalf("op=%x input=%x got=%x want=%x trap=%v", tc.op, m.bytes, vectorBytes(got), want, trap)
			}
		}
		// Exercise every byte value independently at each input position.
		for at := 0; at < tc.size; at++ {
			for n := 0; n < 256; n++ {
				clear(m.bytes)
				m.bytes[1+at] = byte(n)
				check()
			}
		}
		// Mixed signs and nonzero neighbors expose cross-lane carry/borrow.
		x := uint32(0x9e3779b9)
		for range 256 {
			for i := range m.bytes {
				x ^= x << 13
				x ^= x >> 17
				x ^= x << 5
				m.bytes[i] = byte(x)
			}
			check()
		}
	}
}
