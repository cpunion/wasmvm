// Package constant implements the total integer operations allowed in extended
// initializers. Compilation and instance-dependent evaluation share these rules.
package constant

// Binary applies a validated constant binary instruction with Wasm wrapping.
func Binary(op uint32, a, b uint64) uint64 {
	switch op {
	case 0x6a, 0x7c:
		a += b
	case 0x6b, 0x7d:
		a -= b
	case 0x6c, 0x7e:
		a *= b
	default:
		panic("constant: invalid binary instruction")
	}
	if op <= 0x6c {
		return uint64(uint32(a))
	}
	return a
}
