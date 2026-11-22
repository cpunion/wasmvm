//go:build cgo && !llgo && ((linux && amd64) || (darwin && arm64))

package native

/*
#include "adapter.h"
*/
import "C"

import (
	"runtime"
	"sync/atomic"
	"unsafe"

	"github.com/cpunion/wasmvm/internal/platform/hostatomic"
)

var wordAddAvailable = runtime.GOARCH == "amd64" || ARM64CPUFeatures()&1 != 0

func atomicAccess(p unsafe.Pointer, width, operation uint32, a, b uint64) (uint64, bool) {
	// SharedAtomicContext checks natural alignment and polls cancellation.
	// These operations complete without a retry and use SC word accesses.
	if width >= 4 {
		op := AtomicOperation(operation)
		switch op {
		case AtomicLoad:
			if width == 4 {
				return uint64(atomic.LoadUint32((*uint32)(p))), false
			}
			return atomic.LoadUint64((*uint64)(p)), false
		case AtomicStore:
			if width == 4 {
				atomic.StoreUint32((*uint32)(p), uint32(a))
			} else {
				atomic.StoreUint64((*uint64)(p), a)
			}
			return 0, false
		case AtomicAdd, AtomicSub:
			if wordAddAvailable {
				if op == AtomicSub {
					a = -a
				}
				if width == 4 {
					return uint64(hostatomic.FetchAdd32((*uint32)(p), uint32(a))), false
				}
				return hostatomic.FetchAdd64((*uint64)(p), a), false
			}
		}
	}
	result := C.wasmvm_atomic_access(C.uintptr_t(uintptr(p)), C.uint32_t(width), C.uint32_t(operation), C.uint64_t(a), C.uint64_t(b))
	return uint64(result.value), result.retry != 0
}

// Naturally aligned words can use the Go compiler's atomic instructions.
func sharedLoad(p unsafe.Pointer, width uint32) uint64 {
	switch width {
	case 4:
		if uintptr(p)&3 == 0 {
			return uint64(atomic.LoadUint32((*uint32)(p)))
		}
	case 8:
		if uintptr(p)&7 == 0 {
			return atomic.LoadUint64((*uint64)(p))
		}
	}
	return uint64(C.wasmvm_shared_load(C.uintptr_t(uintptr(p)), C.uint32_t(width)))
}
func sharedStore(p unsafe.Pointer, width uint32, value uint64) {
	switch width {
	case 4:
		if uintptr(p)&3 == 0 {
			atomic.StoreUint32((*uint32)(p), uint32(value))
			return
		}
	case 8:
		if uintptr(p)&7 == 0 {
			atomic.StoreUint64((*uint64)(p), value)
			return
		}
	}
	C.wasmvm_shared_store(C.uintptr_t(uintptr(p)), C.uint32_t(width), C.uint64_t(value))
}
func sharedFill(dst unsafe.Pointer, size uint64, byte byte) {
	C.wasmvm_shared_fill(C.uintptr_t(uintptr(dst)), C.size_t(size), C.uint8_t(byte))
}
func sharedFence() { C.wasmvm_shared_fence() }
