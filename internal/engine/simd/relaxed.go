package simd

import (
	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
)

// Choose a fixed permitted projection: strict swizzle/conversion/min/max/Q15,
// bit selection, separately rounded multiply/add, and signed byte products.
// Native lowering uses the same projection on both targets and every CPU mask.
func relaxed(op uint32, a, b, c Value) (r Value) {
	var fixed uint32
	switch {
	case op == 0x100:
		fixed = 14
	case op >= 0x101 && op <= 0x104:
		fixed = [...]uint32{248, 249, 252, 253}[op-0x101]
	case op >= 0x109 && op <= 0x10c:
		fixed = 82
	case op >= 0x10d && op <= 0x110:
		fixed = [...]uint32{232, 233, 244, 245}[op-0x10d]
	case op == 0x111:
		fixed = 130
	}
	if fixed != 0 {
		in := module.Instr{Op: 0xfd0000 | fixed}
		return Eval(&in, a, b, c)
	}
	if op >= 0x105 && op <= 0x108 {
		width, mul, add, neg := uint(32), uint32(0x94), uint32(0x92), uint32(0x8c)
		if op >= 0x107 {
			width, mul, add, neg = 64, 0xa2, 0xa0, 0x9a
		}
		for i := uint(0); i < 128/width; i++ {
			x, _ := numeric.Eval(mul, lane(a, width, i), lane(b, width, i))
			if op&1 == 0 {
				x, _ = numeric.Eval(neg, x, 0)
			}
			x, _ = numeric.Eval(add, x, lane(c, width, i))
			r = put(r, width, i, x)
		}
		return r
	}
	width, terms := uint(16), uint(2)
	if op == 0x113 {
		width, terms = 32, 4
	}
	for i := uint(0); i < 128/width; i++ {
		var sum int64
		for j := uint(0); j < terms; j++ {
			at := i*terms + j
			sum += signed(lane(a, 8, at), 8) * signed(lane(b, 8, at), 8)
		}
		if terms == 4 {
			sum += int64(lane(c, 32, i))
		}
		r = put(r, width, i, uint64(sum))
	}
	return r
}
