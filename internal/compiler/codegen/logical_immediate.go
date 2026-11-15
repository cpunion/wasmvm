package codegen

import "math/bits"

// A logical immediate repeats a rotated, nonempty, nonfull run of ones in a
// power-of-two element. Return the architectural N:immr:imms fields.
func armLogicalImmediate(value uint64, width int) (uint32, bool) {
	if width == 32 {
		value = uint64(uint32(value))
	}
	for size := 2; size <= width; size *= 2 {
		mask := ^uint64(0) >> (64 - size)
		pattern := value & mask
		if pattern == 0 || pattern == mask {
			continue
		}
		repeated := pattern
		for filled := size; filled < width; filled *= 2 {
			repeated |= repeated << filled
		}
		if repeated != value {
			continue
		}
		ones := bits.OnesCount64(pattern)
		run := ^uint64(0) >> (64 - ones)
		for rotation := 0; rotation < size; rotation++ {
			rotated := (run>>rotation | run<<(size-rotation)) & mask
			if rotated == pattern {
				n := uint32(0)
				if size == 64 {
					n = 1 << 22
				}
				return n | uint32(rotation)<<16 | uint32((-size*2|(ones-1))&63)<<10, true
			}
		}
	}
	return 0, false
}
