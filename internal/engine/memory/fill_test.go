package memory

import (
	"bytes"
	"math"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/platform/native"
)

func TestFillPreservesEveryByteAndRejectedRanges(t *testing.T) {
	for _, kind := range []string{"private", "native", "shared"} {
		t.Run(kind, func(t *testing.T) {
			limits := module.Limits{Min: 1, Max: 1, HasMax: true}
			var m *Memory
			var err error
			switch kind {
			case "native":
				if !native.Available {
					t.Skip("native platform unavailable")
				}
				m, err = NewNative(limits, 1)
			case "shared":
				limits.Shared = true
				m, err = NewShared(limits, 1, NewDomain(1, native.Available))
			default:
				m, err = New(limits, 1)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer m.Release()
			for _, n := range []uint64{0, 1, 2, 3, 7, 15, 16, 17, 31, 32, 33, 257, 4095, 4096, 4097, 8192, 8193, 32767, 65534} {
				for _, value := range []byte{0, 1, 0x80, 0xff} {
					for i := range m.bytes {
						m.bytes[i] = 0xa5
					}
					if !m.Fill(1, value, n) {
						t.Fatal("valid fill rejected", n, value)
					}
					for i, got := range m.bytes {
						want := byte(0xa5)
						if i >= 1 && uint64(i) < n+1 {
							want = value
						}
						if got != want {
							t.Fatal("fill changed an incorrect byte", n, value, i, got, want)
						}
					}
				}
			}
			before := append([]byte(nil), m.bytes...)
			for _, bad := range [][2]uint64{{PageSize, 1}, {PageSize - 1, 2}, {PageSize + 1, 0}, {math.MaxUint64, 0}, {0, math.MaxUint64}} {
				if m.Fill(bad[0], 0, bad[1]) || !bytes.Equal(m.bytes, before) {
					t.Fatal("rejected fill modified memory", bad)
				}
			}
			if !m.Fill(PageSize, 0, 0) || !bytes.Equal(m.bytes, before) {
				t.Fatal("empty end range")
			}
		})
	}
}
