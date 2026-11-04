//go:build (cgo || llgo) && ((linux && amd64) || (darwin && arm64))

package native

/*
#cgo CFLAGS: -std=gnu11
#cgo linux LDFLAGS: -lm
#include "adapter.h"
*/
import "C"

import "unsafe"

const Available = true

// ARM64CPUFeatures reports FEAT_LSE in bit zero; an unavailable probe returns zero.
func ARM64CPUFeatures() uint32 { return uint32(C.wasmvm_arm64_features()) }

// CPUFeatures reports SSSE3/SSE4.1/AVX2/AVX512F+VL/AVX512BITALG+VL in bits 0-4.
// AVX-512 requires XMM/YMM/opmask/ZMM OS state even for 128-bit instructions.
// Other targets return zero.
func CPUFeatures() uint32 {
	var ecx1, ebx7, ecx7 C.uint32_t
	var xcr0 C.uint64_t
	C.wasmvm_cpu_probe(&ecx1, &ebx7, &ecx7, &xcr0)
	return x86FeatureMask(uint32(ecx1), uint32(ebx7), uint32(ecx7), uint64(xcr0))
}

func publish(code []byte) unsafe.Pointer {
	return C.wasmvm_publish(unsafe.Pointer(&code[0]), C.size_t(len(code)))
}
func unpublish(p unsafe.Pointer, n uint64) { C.wasmvm_unpublish(p, C.size_t(n)) }
func allocate(n uint64) unsafe.Pointer     { return C.wasmvm_allocate(C.size_t(n)) }
func deallocate(p unsafe.Pointer)          { C.wasmvm_deallocate(p) }
func reserve(n uint64) unsafe.Pointer      { return C.wasmvm_reserve(C.size_t(n)) }
func commit(p unsafe.Pointer, previous, size uint64) bool {
	return C.wasmvm_commit(p, C.size_t(previous), C.size_t(size)) != 0
}
func unreserve(p unsafe.Pointer, n uint64) { C.wasmvm_unreserve(p, C.size_t(n)) }

// Guest addresses below always refer to retained native allocations. Numeric
// C parameters avoid generated pointer-check closures; host-buffer transfers
// keep pointer parameters and their Go liveness/pinning checks.
func sharedTransfer(dst, src unsafe.Pointer, size uint64, mode uint32) {
	C.wasmvm_shared_transfer(dst, src, C.size_t(size), C.uint32_t(mode))
}
func cancel(p unsafe.Pointer) { C.wasmvm_cancel((*C.struct_wasmvm_record)(p)) }
func enter(p, ctx unsafe.Pointer) Exit {
	return Exit(C.wasmvm_enter(p, (*C.struct_wasmvm_record)(ctx)))
}
func layout(field uint32) uint64 { return uint64(C.wasmvm_layout(C.uint32_t(field))) }
func probeFP(entry, record unsafe.Pointer) bool {
	return C.wasmvm_probe_fp(entry, (*C.struct_wasmvm_record)(record)) != 0
}
func probeRegisters(entry, record unsafe.Pointer) bool {
	return C.wasmvm_probe_registers(entry, (*C.struct_wasmvm_record)(record)) != 0
}
