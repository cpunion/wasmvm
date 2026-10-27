package codegen

// Modified SIMD immediates fill a whole 128-bit register without scalar scratch
// values. Reject every unrepresentable bit pattern rather than rounding a float.
func armVectorImmediate(lo, hi uint64) (uint32, bool) {
	if lo != hi {
		return 0, false
	}
	if lo == uint64(byte(lo))*0x0101010101010101 {
		return armModifiedImmediate(14, false, byte(lo)), true
	}
	for _, invert := range []bool{false, true} {
		bits := lo
		if invert {
			bits = ^bits
		}
		word := uint32(bits)
		if bits == uint64(word)*0x100000001 {
			for shift := uint(0); shift < 4; shift++ {
				if word&^(uint32(255)<<(shift*8)) == 0 {
					return armModifiedImmediate(byte(shift*2), invert, byte(word>>(shift*8))), true
				}
			}
			for shift := uint(1); shift <= 2; shift++ {
				fill := uint32(1)<<(shift*8) - 1
				if word&fill == fill && word>>(shift*8+8) == 0 {
					return armModifiedImmediate(byte(11+shift), invert, byte(word>>(shift*8))), true
				}
			}
		}
		half := uint16(bits)
		if bits == uint64(half)*0x0001000100010001 {
			if half < 256 {
				return armModifiedImmediate(8, invert, byte(half)), true
			}
			if half&255 == 0 {
				return armModifiedImmediate(10, invert, byte(half>>8)), true
			}
		}
	}
	var byteMask byte
	valid := true
	for lane := uint(0); lane < 8; lane++ {
		b := byte(lo >> (lane * 8))
		if b == 255 {
			byteMask |= 1 << lane
		} else if b != 0 {
			valid = false
			break
		}
	}
	if valid {
		return armModifiedImmediate(14, true, byteMask), true
	}
	if word := uint32(lo); lo == uint64(word)*0x100000001 {
		if immediate, ok := armFloatImmediate(uint64(word), 23, 8); ok {
			return armModifiedImmediate(15, false, immediate), true
		}
	}
	if immediate, ok := armFloatImmediate(lo, 52, 11); ok {
		return armModifiedImmediate(15, true, immediate), true
	}
	return 0, false
}

func armModifiedImmediate(mode byte, op bool, immediate byte) uint32 {
	ins := uint32(0x4f000400) | uint32(mode)<<12 | uint32(immediate>>5)<<16 | uint32(immediate&31)<<5
	if op {
		ins |= 1 << 29
	}
	return ins
}

func armFloatImmediate(bits uint64, fraction, exponent uint) (byte, bool) {
	sign := bits >> (fraction + exponent) & 1
	b6 := bits >> (fraction + exponent - 2) & 1
	low := bits >> (fraction - 4) & 63
	immediate := byte(sign<<7 | b6<<6 | low)
	expanded := sign<<(fraction+exponent) | (b6^1)<<(fraction+exponent-1) |
		(b6*((1<<(exponent-3))-1))<<(fraction+2) | low<<(fraction-4)
	return immediate, expanded == bits
}
