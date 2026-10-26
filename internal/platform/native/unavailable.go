//go:build (!cgo && !llgo) || !((linux && amd64) || (darwin && arm64))

package native

import "unsafe"

const Available = false

func publish([]byte) unsafe.Pointer                      { return nil }
func unpublish(unsafe.Pointer, uint64)                   {}
func allocate(uint64) unsafe.Pointer                     { return nil }
func deallocate(unsafe.Pointer)                          {}
func cancel(unsafe.Pointer)                              {}
func enter(unsafe.Pointer, unsafe.Pointer) Exit          { return 0 }
func layout(uint32) uint64                               { return 0 }
func probeFP(unsafe.Pointer, unsafe.Pointer) bool        { return false }
func probeRegisters(unsafe.Pointer, unsafe.Pointer) bool { return false }
