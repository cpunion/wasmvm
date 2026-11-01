//go:build (!cgo && !llgo) || !((linux && amd64) || (darwin && arm64))

package native

import "unsafe"

const Available = false

func CPUFeatures() uint32 { return 0 }

func publish([]byte) unsafe.Pointer                                              { return nil }
func unpublish(unsafe.Pointer, uint64)                                           {}
func allocate(uint64) unsafe.Pointer                                             { return nil }
func deallocate(unsafe.Pointer)                                                  {}
func atomicAccess(unsafe.Pointer, uint32, uint32, uint64, uint64) (uint64, bool) { return 0, false }
func sharedLoad(unsafe.Pointer, uint32) uint64                                   { return 0 }
func sharedStore(unsafe.Pointer, uint32, uint64)                                 {}
func sharedTransfer(unsafe.Pointer, unsafe.Pointer, uint64, uint32)              {}
func sharedFill(unsafe.Pointer, uint64, byte)                                    {}
func sharedFence()                                                               {}
func cancel(unsafe.Pointer)                                                      {}
func enter(unsafe.Pointer, unsafe.Pointer) Exit                                  { return 0 }
func layout(uint32) uint64                                                       { return 0 }
func probeFP(unsafe.Pointer, unsafe.Pointer) bool                                { return false }
func probeRegisters(unsafe.Pointer, unsafe.Pointer) bool                         { return false }
