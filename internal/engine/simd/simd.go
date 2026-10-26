// Package simd implements the pure fixed-SIMD Core 2.0 operations. Memory
// instructions are handled by the memory subsystem at their checked boundary.
package simd

import (
	"encoding/binary"
	"math/bits"

	"github.com/cpunion/wasmvm/internal/core/module"
	"github.com/cpunion/wasmvm/internal/engine/numeric"
	"github.com/cpunion/wasmvm/internal/engine/value"
)

type Value = value.Value

func lane(v Value, width, index uint) uint64 {
	shift := index * width
	x := v.Lo
	if shift >= 64 {
		x = v.Hi
		shift -= 64
	}
	return (x >> shift) & (^uint64(0) >> (64 - width))
}
func signed(x uint64, width uint) int64 { return int64(x<<(64-width)) >> (64 - width) }
func put(v *Value, width, index uint, x uint64) {
	shift := index * width
	mask := ^uint64(0) >> (64 - width)
	x &= mask
	if shift >= 64 {
		shift -= 64
		v.Hi = v.Hi&^(mask<<shift) | x<<shift
	} else {
		v.Lo = v.Lo&^(mask<<shift) | x<<shift
	}
}
func saturate(x int64, width uint, unsigned bool) uint64 {
	lower, upper := -(int64(1) << (width - 1)), (int64(1)<<(width-1))-1
	if unsigned {
		lower, upper = 0, (int64(1)<<width)-1
	}
	if x < lower {
		x = lower
	}
	if x > upper {
		x = upper
	}
	return uint64(x)
}

// Eval executes a validated nonmemory SIMD instruction. Operands use Wasm
// stack order, including scalar lane/shift inputs in Lo. Unknown and memory
// opcodes panic as caller bugs; no fixed pure SIMD instruction traps.
func Eval(in module.Instr, a, b, c Value) (r Value) {
	op := in.Op & 0xffff
	if in.Op>>16 != 0xfd || !pureOpcode(op) {
		panic("simd: nonnumeric or unknown opcode")
	}
	switch {
	case op == 12:
		return Value{Lo: binary.LittleEndian.Uint64(in.Vector[:8]), Hi: binary.LittleEndian.Uint64(in.Vector[8:])}
	case op == 13:
		for i, n := range in.Vector {
			v := a
			if n >= 16 {
				v = b
				n -= 16
			}
			put(&r, 8, uint(i), lane(v, 8, uint(n)))
		}
		return r
	case op == 14:
		for i := uint(0); i < 16; i++ {
			n := lane(b, 8, i)
			if n < 16 {
				put(&r, 8, i, lane(a, 8, uint(n)))
			}
		}
		return r
	case op >= 15 && op <= 20:
		width := [6]uint{8, 16, 32, 64, 32, 64}[op-15]
		for i := uint(0); i < 128/width; i++ {
			put(&r, width, i, a.Lo)
		}
		return r
	case op >= 21 && op <= 34:
		var width uint
		replace, signedExtract := false, false
		switch op {
		case 21, 22, 23:
			width = 8
			replace = op == 23
			signedExtract = op == 21
		case 24, 25, 26:
			width = 16
			replace = op == 26
			signedExtract = op == 24
		case 27, 28:
			width = 32
			replace = op == 28
		case 29, 30:
			width = 64
			replace = op == 30
		case 31, 32:
			width = 32
			replace = op == 32
		case 33, 34:
			width = 64
			replace = op == 34
		}
		if replace {
			r = a
			put(&r, width, uint(in.Lane), b.Lo)
			return r
		}
		x := lane(a, width, uint(in.Lane))
		if signedExtract {
			x = uint64(uint32(signed(x, width)))
		}
		return Value{Lo: x}
	case op >= 35 && op <= 76 || op >= 214 && op <= 219:
		var width, kind uint
		floating := false
		switch {
		case op < 65:
			width = 8 << ((op - 35) / 10)
			kind = uint((op - 35) % 10)
		case op < 71:
			width = 32
			kind = uint(op - 65)
			floating = true
		case op < 77:
			width = 64
			kind = uint(op - 71)
			floating = true
		default:
			width = 64
			kind = [6]uint{0, 1, 2, 4, 6, 8}[op-214]
		}
		for i := uint(0); i < 128/width; i++ {
			x, y := lane(a, width, i), lane(b, width, i)
			yes := false
			if floating {
				base := uint32(0x5b)
				if width == 64 {
					base = 0x61
				}
				z, _ := numeric.Eval(base+uint32(kind), x, y)
				yes = z != 0
			} else {
				switch kind {
				case 0:
					yes = x == y
				case 1:
					yes = x != y
				case 2:
					yes = signed(x, width) < signed(y, width)
				case 3:
					yes = x < y
				case 4:
					yes = signed(x, width) > signed(y, width)
				case 5:
					yes = x > y
				case 6:
					yes = signed(x, width) <= signed(y, width)
				case 7:
					yes = x <= y
				case 8:
					yes = signed(x, width) >= signed(y, width)
				case 9:
					yes = x >= y
				}
			}
			if yes {
				put(&r, width, i, ^uint64(0))
			}
		}
		return r
	case op >= 77 && op <= 83:
		switch op {
		case 77:
			return Value{Lo: ^a.Lo, Hi: ^a.Hi}
		case 78:
			return Value{Lo: a.Lo & b.Lo, Hi: a.Hi & b.Hi}
		case 79:
			return Value{Lo: a.Lo &^ b.Lo, Hi: a.Hi &^ b.Hi}
		case 80:
			return Value{Lo: a.Lo | b.Lo, Hi: a.Hi | b.Hi}
		case 81:
			return Value{Lo: a.Lo ^ b.Lo, Hi: a.Hi ^ b.Hi}
		case 82:
			return Value{Lo: a.Lo&c.Lo | b.Lo&^c.Lo, Hi: a.Hi&c.Hi | b.Hi&^c.Hi}
		case 83:
			if a.Lo|a.Hi != 0 {
				return Value{Lo: 1}
			}
			return r
		}
	case op == 94 || op == 95 || op >= 248 && op <= 255:
		var from, to, n uint
		var scalar uint32
		switch op {
		case 94:
			from, to, n, scalar = 64, 32, 2, 0xb6
		case 95:
			from, to, n, scalar = 32, 64, 2, 0xbb
		case 248, 249:
			from, to, n, scalar = 32, 32, 4, 0xfc0000+op-248
		case 250, 251:
			from, to, n, scalar = 32, 32, 4, 0xb2+op-250
		case 252, 253:
			from, to, n, scalar = 64, 32, 2, 0xfc0002+op-252
		case 254, 255:
			from, to, n, scalar = 32, 64, 2, 0xb7+op-254
		}
		for i := uint(0); i < n; i++ {
			x, _ := numeric.Eval(scalar, lane(a, from, i), 0)
			put(&r, to, i, x)
		}
		return r
	case op >= 103 && op <= 106 || op == 116 || op == 117 || op == 122 || op == 148 || op >= 224 && op <= 247:
		return floating(op, a, b)
	case op >= 124 && op <= 127:
		width := uint(8)
		if op >= 126 {
			width = 16
		}
		unsigned := op%2 != 0
		for i := uint(0); i < 64/width; i++ {
			x, y := lane(a, width, 2*i), lane(a, width, 2*i+1)
			if !unsigned {
				x, y = uint64(signed(x, width)), uint64(signed(y, width))
			}
			put(&r, width*2, i, x+y)
		}
		return r
	case op >= 96 && op <= 223:
		width := uint(8) << ((op - 96) / 32)
		kind := (op - 96) % 32
		switch kind {
		case 3, 4:
			all := true
			for i := uint(0); i < 128/width; i++ {
				x := lane(a, width, i)
				all = all && x != 0
				if kind == 4 {
					r.Lo |= (x >> (width - 1)) << i
				}
			}
			if kind == 3 && all {
				r.Lo = 1
			}
			return r
		case 5, 6:
			if width > 16 {
				panic("simd: invalid narrow opcode")
			}
			n := 64 / width
			for i := uint(0); i < n; i++ {
				put(&r, width, i, saturate(signed(lane(a, 2*width, i), 2*width), width, kind == 6))
				put(&r, width, i+n, saturate(signed(lane(b, 2*width, i), 2*width), width, kind == 6))
			}
			return r
		case 7, 8, 9, 10, 28, 29, 30, 31:
			if width == 8 {
				panic("simd: invalid extend opcode")
			}
			kindBase := uint32(7)
			multiply := kind >= 28
			if multiply {
				kindBase = 28
			}
			high, unsigned := (kind-kindBase)%2 == 1, (kind-kindBase) >= 2
			start := uint(0)
			if high {
				start = 128 / width
			}
			for i := uint(0); i < 128/width; i++ {
				x, y := lane(a, width/2, start+i), lane(b, width/2, start+i)
				if !unsigned {
					x, y = uint64(signed(x, width/2)), uint64(signed(y, width/2))
				}
				if multiply {
					x *= y
				}
				put(&r, width, i, x)
			}
			return r
		case 26:
			if width != 32 {
				panic("simd: invalid dot opcode")
			}
			for i := uint(0); i < 4; i++ {
				x := signed(lane(a, 16, 2*i), 16)*signed(lane(b, 16, 2*i), 16) + signed(lane(a, 16, 2*i+1), 16)*signed(lane(b, 16, 2*i+1), 16)
				put(&r, 32, i, uint64(x))
			}
			return r
		}
		for i := uint(0); i < 128/width; i++ {
			x, y := lane(a, width, i), lane(b, width, i)
			sx, sy := signed(x, width), signed(y, width)
			var z uint64
			switch kind {
			case 0:
				if sx < 0 {
					x = uint64(-sx)
				}
				z = x
			case 1:
				z = 0 - x
			case 2:
				if width == 8 {
					z = uint64(bits.OnesCount8(uint8(x)))
				} else if width == 16 {
					z = saturate((sx*sy+0x4000)>>15, 16, false)
				} else {
					panic("simd: invalid integer opcode")
				}
			case 11:
				z = x << (b.Lo & (uint64(width) - 1))
			case 12:
				z = uint64(sx >> (b.Lo & (uint64(width) - 1)))
			case 13:
				z = x >> (b.Lo & (uint64(width) - 1))
			case 14:
				z = x + y
			case 15:
				z = saturate(sx+sy, width, false)
			case 16:
				z = saturate(int64(x+y), width, true)
			case 17:
				z = x - y
			case 18:
				z = saturate(sx-sy, width, false)
			case 19:
				z = saturate(int64(x)-int64(y), width, true)
			case 21:
				z = x * y
			case 22:
				z = x
				if sx > sy {
					z = y
				}
			case 23:
				z = x
				if x > y {
					z = y
				}
			case 24:
				z = x
				if sx < sy {
					z = y
				}
			case 25:
				z = x
				if x < y {
					z = y
				}
			case 27:
				z = (x + y + 1) >> 1
			default:
				panic("simd: invalid integer opcode")
			}
			put(&r, width, i, z)
		}
		return r
	}
	panic("simd: nonnumeric or unknown opcode")
}

func pureOpcode(op uint32) bool {
	if op < 12 || op > 255 || op >= 84 && op <= 93 {
		return false
	}
	// Core 2.0's reserved SIMD holes; the same pinned binary grammar is checked
	// by the decoder. This guard keeps internal dispatch mistakes visible.
	mask := [4]uint64{0xffffffffffffffff, 0xffffffffffffffff, 0xf7e27f9bfbffffff, 0xffffbffbffe27f9b}
	return mask[op/64]&(uint64(1)<<(op%64)) != 0
}

func floating(op uint32, a, b Value) (r Value) {
	width := uint(32)
	var scalar uint32
	switch op {
	case 103, 104, 105, 106:
		scalar = 0x8d + op - 103
	case 116:
		width, scalar = 64, 0x9b
	case 117:
		width, scalar = 64, 0x9c
	case 122:
		width, scalar = 64, 0x9d
	case 148:
		width, scalar = 64, 0x9e
	case 224, 225:
		scalar = 0x8b + op - 224
	case 227:
		scalar = 0x91
	case 228, 229, 230, 231, 232, 233:
		scalar = 0x92 + op - 228
	case 234, 235:
		scalar = 0
	case 236, 237:
		width, scalar = 64, 0x99+op-236
	case 239:
		width, scalar = 64, 0x9f
	case 240, 241, 242, 243, 244, 245:
		width, scalar = 64, 0xa0+op-240
	case 246, 247:
		width, scalar = 64, 0
	default:
		panic("simd: invalid float opcode")
	}
	for i := uint(0); i < 128/width; i++ {
		x, y := lane(a, width, i), lane(b, width, i)
		var z uint64
		if scalar != 0 {
			z, _ = numeric.Eval(scalar, x, y)
		} else {
			// Pseudo-min/max are ordered selections, preserving the selected operand's
			// raw NaN/zero bits. For pmin use y<x ? y:x; for pmax use x<y ? y:x.
			cmp := uint32(0x5d)
			if width == 64 {
				cmp = 0x63
			}
			left, right := y, x
			if op%2 == 1 {
				left, right = x, y
			}
			yes, _ := numeric.Eval(cmp, left, right)
			z = x
			if yes != 0 {
				z = y
			}
		}
		put(&r, width, i, z)
	}
	return r
}
