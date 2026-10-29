package constant

import "testing"

func TestBinaryWrapping(t *testing.T) {
	for _, tc := range []struct {
		op         uint32
		a, b, want uint64
	}{
		{0x6a, 0xffffffff, 1, 0}, {0x6b, 0, 1, 0xffffffff},
		{0x6c, 0x80000001, 2, 2}, {0x7c, ^uint64(0), 1, 0},
		{0x7d, 0, 1, ^uint64(0)}, {0x7e, 1<<63 | 1, 2, 2},
	} {
		if got := Binary(tc.op, tc.a, tc.b); got != tc.want {
			t.Fatalf("%x: %x want %x", tc.op, got, tc.want)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("invalid instruction accepted")
		}
	}()
	Binary(0x6d, 1, 1)
}
