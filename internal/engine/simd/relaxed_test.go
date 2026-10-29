package simd

import (
	"testing"

	"github.com/cpunion/wasmvm/internal/core/module"
)

func TestRelaxedProjectionRoundingAndWideDot(t *testing.T) {
	// (1+epsilon)*(1-epsilon) rounds to one before adding minus one.
	// A fused implementation would retain a small negative residual.
	for _, c := range []struct {
		op      uint32
		a, b, c Value
		want    Value
	}{
		{0x105, Value{Lo: 0x3f8000013f800001, Hi: 0x3f8000013f800001}, Value{Lo: 0x3f7ffffe3f7ffffe, Hi: 0x3f7ffffe3f7ffffe}, Value{Lo: 0xbf800000bf800000, Hi: 0xbf800000bf800000}, Value{}},
		{0x106, Value{Lo: 0x3f8000013f800001, Hi: 0x3f8000013f800001}, Value{Lo: 0x3f7ffffe3f7ffffe, Hi: 0x3f7ffffe3f7ffffe}, Value{Lo: 0x3f8000003f800000, Hi: 0x3f8000003f800000}, Value{}},
		{0x107, Value{Lo: 0x3ff0000000000001, Hi: 0x3ff0000000000001}, Value{Lo: 0x3feffffffffffffe, Hi: 0x3feffffffffffffe}, Value{Lo: 0xbff0000000000000, Hi: 0xbff0000000000000}, Value{}},
		{0x108, Value{Lo: 0x3ff0000000000001, Hi: 0x3ff0000000000001}, Value{Lo: 0x3feffffffffffffe, Hi: 0x3feffffffffffffe}, Value{Lo: 0x3ff0000000000000, Hi: 0x3ff0000000000000}, Value{}},
		{0x112, Value{Lo: 0x8080808080808080, Hi: 0x8080808080808080}, Value{Lo: 0x8080808080808080, Hi: 0x8080808080808080}, Value{}, Value{Lo: 0x8000800080008000, Hi: 0x8000800080008000}},
		{0x113, Value{Lo: 0x8080808080808080, Hi: 0x8080808080808080}, Value{Lo: 0x8080808080808080, Hi: 0x8080808080808080}, Value{Lo: 0xffffffffffffffff, Hi: 0xffffffffffffffff}, Value{Lo: 0x0000ffff0000ffff, Hi: 0x0000ffff0000ffff}},
	} {
		in := module.Instr{Op: 0xfd0000 | c.op}
		if got := Eval(&in, c.a, c.b, c.c); got != c.want {
			t.Fatalf("op %x: got %#v want %#v", c.op, got, c.want)
		}
	}
}
