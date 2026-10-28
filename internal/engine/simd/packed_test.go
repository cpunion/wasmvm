package simd

import (
	"math/bits"
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
)

func scalarPackedArithmetic(a, b Value, width uint, subtract bool) (r Value) {
	mask := ^uint64(0) >> (64 - width)
	for word := 0; word < 2; word++ {
		x, y := a.Lo, b.Lo
		if word == 1 {
			x, y = a.Hi, b.Hi
		}
		var out uint64
		for shift := uint(0); shift < 64; shift += width {
			left, right := (x>>shift)&mask, (y>>shift)&mask
			z := left + right
			if subtract {
				z = left - right
			}
			out |= (z & mask) << shift
		}
		if word == 0 {
			r.Lo = out
		} else {
			r.Hi = out
		}
	}
	return r
}

func checkPackedArithmetic(t *testing.T, a, b Value, index uint32) {
	t.Helper()
	width := uint(8) << index
	for _, subtract := range []bool{false, true} {
		op := uint32(110) + index*32
		if subtract {
			op += 3
		}
		in := module.Instr{Op: 0xfd0000 | op}
		want := scalarPackedArithmetic(a, b, width, subtract)
		if got := Eval(&in, a, b, Value{}); got != want {
			t.Fatalf("width %d subtract %v (%#v, %#v): %#v != %#v", width, subtract, a, b, got, want)
		}
	}
}

func TestPackedBytePairs(t *testing.T) {
	for x := uint64(0); x < 256; x++ {
		for y := uint64(0); y < 256; y++ {
			a := Value{Lo: x * 0x0101010101010101, Hi: ^(x * 0x0100010001000100)}
			b := Value{Lo: y * 0x0101010101010101, Hi: y * 0x0001000100010001}
			checkPackedArithmetic(t, a, b, 0)
			in := module.Instr{Op: 0xfd0062}
			input := Value{Lo: x | y<<8 | 0xffffffffffff0000, Hi: ^(x | y<<8)}
			var want Value
			for shift := uint(0); shift < 64; shift += 8 {
				want.Lo |= uint64(bits.OnesCount8(byte(input.Lo>>shift))) << shift
				want.Hi |= uint64(bits.OnesCount8(byte(input.Hi>>shift))) << shift
			}
			if got := Eval(&in, input, Value{}, Value{}); got != want {
				t.Fatalf("popcount %#v: %#v != %#v", input, got, want)
			}
		}
	}
}

func TestPackedWideArithmetic(t *testing.T) {
	seed := uint64(0x123456789abcdef)
	for index := uint32(1); index < 4; index++ {
		for _, x := range []uint64{0, 1, ^uint64(0), 0x8000800080008000, 0x8000000080000000, 1 << 63} {
			for _, y := range []uint64{0, 1, ^uint64(0), x, ^x} {
				checkPackedArithmetic(t, Value{Lo: x, Hi: ^y}, Value{Lo: y, Hi: ^x}, index)
			}
		}
		for n := 0; n < 4096; n++ {
			seed ^= seed << 13
			seed ^= seed >> 7
			seed ^= seed << 17
			checkPackedArithmetic(t, Value{Lo: seed, Hi: ^seed}, Value{Lo: bits.RotateLeft64(seed, 17), Hi: bits.RotateLeft64(seed, -29)}, index)
		}
	}
}
