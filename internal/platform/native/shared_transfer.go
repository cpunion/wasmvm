package native

import "unsafe"

// The caller retains the owner and validates guest ranges before every access.
// Unaligned scalar accesses preserve little-endian bits and may tear by byte.
func SharedLoad(address unsafe.Pointer, width uint32) (uint64, error) {
	if !Available {
		return 0, ErrUnavailable
	}
	if address == nil || width != 1 && width != 2 && width != 4 && width != 8 {
		return 0, ErrInvalid
	}
	return sharedLoad(address, width), nil
}

func SharedStore(address unsafe.Pointer, width uint32, value uint64) error {
	if !Available {
		return ErrUnavailable
	}
	if address == nil || width != 1 && width != 2 && width != 4 && width != 8 {
		return ErrInvalid
	}
	sharedStore(address, width, value)
	return nil
}

// SharedTransfer preserves memmove overlap semantics. Either endpoint may be
// an ordinary host/private buffer; shared endpoints use relaxed atomic bytes
// or aligned words. No Go pointer is retained by the synchronous adapter.
func SharedTransfer(dst, src unsafe.Pointer, size uint64, dstShared, srcShared bool) error {
	if !Available {
		return ErrUnavailable
	}
	if size > uint64(^uint(0)>>1) || size != 0 && (dst == nil || src == nil) {
		return ErrInvalid
	}
	var mode uint32
	if dstShared {
		mode |= 1
	}
	if srcShared {
		mode |= 2
	}
	sharedTransfer(dst, src, size, mode)
	return nil
}

func SharedFill(dst unsafe.Pointer, size uint64, value byte) error {
	if !Available {
		return ErrUnavailable
	}
	if size > uint64(^uint(0)>>1) || size != 0 && dst == nil {
		return ErrInvalid
	}
	sharedFill(dst, size, value)
	return nil
}

func SharedFence() error {
	if !Available {
		return ErrUnavailable
	}
	sharedFence()
	return nil
}
