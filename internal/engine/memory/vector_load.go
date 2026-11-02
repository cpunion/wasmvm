package memory

import "github.com/cpunion/wasmvm/internal/engine/value"

// widenLoad spreads an already checked eight-byte input into two words. Each
// input lane occupies the low half of its output lane, leaving the upper half
// free for sign extension without carries between lanes.
func widenLoad(x uint64, sub uint32) (r value.Value) {
	width := uint(8) << ((sub - 1) / 2)
	switch width {
	case 8:
		lo, hi := uint64(uint32(x)), x>>32
		lo = (lo | lo<<16) & 0x0000ffff0000ffff
		hi = (hi | hi<<16) & 0x0000ffff0000ffff
		r.Lo = (lo | lo<<8) & 0x00ff00ff00ff00ff
		r.Hi = (hi | hi<<8) & 0x00ff00ff00ff00ff
	case 16:
		r.Lo = x&0xffff | (x&0xffff0000)<<16
		r.Hi = (x>>32)&0xffff | (x>>16)&0xffff00000000
	case 32:
		r.Lo, r.Hi = uint64(uint32(x)), x>>32
	}
	if sub&1 != 0 {
		mask := uint64(0x80000000)
		if width == 8 {
			mask = 0x0080008000800080
		} else if width == 16 {
			mask = 0x0000800000008000
		}
		lo, hi := r.Lo&mask, r.Hi&mask
		// Each subtraction fills only that lane's empty upper half. The
		// highest lane uses ordinary modulo-64-bit arithmetic.
		r.Lo |= lo<<(width+1) - lo<<1
		r.Hi |= hi<<(width+1) - hi<<1
	}
	return r
}
