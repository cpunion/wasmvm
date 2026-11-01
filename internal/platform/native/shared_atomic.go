package native

import "unsafe"

type AtomicOperation uint32

const (
	AtomicLoad AtomicOperation = iota
	AtomicStore
	AtomicAdd
	AtomicSub
	AtomicAnd
	AtomicOr
	AtomicXor
	AtomicExchange
	AtomicCompareExchange
)

// SharedAtomic accesses retained native linear memory. The trusted caller must
// check the complete guest range against the atomically observed logical size
// and keep the owner alive through this call. All widths use sequentially consistent
// operations; subword inputs/results are zero-extended bit patterns.
func SharedAtomic(address unsafe.Pointer, width uint32, operation AtomicOperation, a, b uint64) (uint64, error) {
	if !Available {
		return 0, ErrUnavailable
	}
	if address == nil || width != 1 && width != 2 && width != 4 && width != 8 || uintptr(address)&uintptr(width-1) != 0 || operation > AtomicCompareExchange {
		return 0, ErrInvalid
	}
	return atomicAccess(address, width, uint32(operation), a, b), nil
}
