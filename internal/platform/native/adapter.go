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

// CPUFeatures reports SSSE3, SSE4.1 and OS-enabled AVX2 in bits 0/1/2.
// Other targets return zero; AVX2 requires both XMM and YMM state in XCR0.
func CPUFeatures() uint32 {
	var ecx1, ebx7 C.uint32_t
	var xcr0 C.uint64_t
	C.wasmvm_cpu_probe(&ecx1, &ebx7, &xcr0)
	return x86FeatureMask(uint32(ecx1), uint32(ebx7), uint64(xcr0))
}

func publish(code []byte) unsafe.Pointer {
	return C.wasmvm_publish(unsafe.Pointer(&code[0]), C.size_t(len(code)))
}
func unpublish(p unsafe.Pointer, n uint64) { C.wasmvm_unpublish(p, C.size_t(n)) }
func allocate(n uint64) unsafe.Pointer     { return C.wasmvm_allocate(C.size_t(n)) }
func deallocate(p unsafe.Pointer)          { C.wasmvm_deallocate(p) }
func cancel(p unsafe.Pointer)              { C.wasmvm_cancel((*C.struct_wasmvm_record)(p)) }
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
