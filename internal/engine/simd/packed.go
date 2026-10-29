package simd

// Clear each lane's high bit before adding, so carries cannot cross lanes.
// Restore its contribution modulo the lane width with XOR. For subtraction,
// set the minuend's high bits first to prevent cross-lane borrows, then adjust.
func packedAddSub(a, b Value, width uint, subtract bool) Value {
	high := uint64(0x8000000000000000)
	switch width {
	case 8:
		high = 0x8080808080808080
	case 16:
		high = 0x8000800080008000
	case 32:
		high = 0x8000000080000000
	}
	if subtract {
		return Value{
			Lo: ((a.Lo | high) - (b.Lo &^ high)) ^ ((^a.Lo ^ b.Lo) & high),
			Hi: ((a.Hi | high) - (b.Hi &^ high)) ^ ((^a.Hi ^ b.Hi) & high),
		}
	}
	return Value{
		Lo: ((a.Lo &^ high) + (b.Lo &^ high)) ^ ((a.Lo ^ b.Lo) & high),
		Hi: ((a.Hi &^ high) + (b.Hi &^ high)) ^ ((a.Hi ^ b.Hi) & high),
	}
}

// Count independently in bit pairs, nibbles, then bytes. Each intermediate
// count fits its group, so shifts and additions cannot mix adjacent bytes.
func popcountBytes(x uint64) uint64 {
	x -= (x >> 1) & 0x5555555555555555
	x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
	return (x + (x >> 4)) & 0x0f0f0f0f0f0f0f0f
}
