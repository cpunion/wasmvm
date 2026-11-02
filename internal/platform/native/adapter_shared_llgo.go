//go:build llgo && ((linux && amd64) || (darwin && arm64))

package native

import "unsafe"

// These calls contain only numeric arguments for retained C-owned memory.
// Direct C declarations avoid the pinned LLGo Cgo wrapper's unused allocation
// (#2958). Host-buffer calls keep the ordinary Cgo checks in adapter.go.
type atomicAccessResult struct{ value, retry uint64 }

//go:linkname atomicAccessNative C.wasmvm_atomic_access
func atomicAccessNative(address uintptr, width, operation uint32, a, b uint64) atomicAccessResult

//go:linkname sharedLoadNative C.wasmvm_shared_load
func sharedLoadNative(address uintptr, width uint32) uint64

//go:linkname sharedStoreNative C.wasmvm_shared_store
func sharedStoreNative(address uintptr, width uint32, value uint64)

//go:linkname sharedFillNative C.wasmvm_shared_fill
func sharedFillNative(address uintptr, size uint64, value uint8)

//go:linkname sharedFence C.wasmvm_shared_fence
func sharedFence()

func atomicAccess(p unsafe.Pointer, width, operation uint32, a, b uint64) (uint64, bool) {
	result := atomicAccessNative(uintptr(p), width, operation, a, b)
	return result.value, result.retry != 0
}
func sharedLoad(p unsafe.Pointer, width uint32) uint64 { return sharedLoadNative(uintptr(p), width) }
func sharedStore(p unsafe.Pointer, width uint32, value uint64) {
	sharedStoreNative(uintptr(p), width, value)
}
func sharedFill(p unsafe.Pointer, size uint64, value byte) { sharedFillNative(uintptr(p), size, value) }
