package memory

import (
	"unsafe"

	"github.com/cpunion/wasmvm/internal/platform/native"
)

func checkedShared(err error) {
	if err != nil {
		panic(err) // The range/shape was checked before entering the adapter.
	}
}

func (m *Memory) readAccess(b []byte, width uint) uint64 {
	if m.sharedSize == nil {
		return read(b, width)
	}
	return readShared(b, width)
}

func readShared(b []byte, width uint) uint64 {
	v, err := native.SharedLoad(unsafe.Pointer(unsafe.SliceData(b)), uint32(width))
	checkedShared(err)
	return v
}

func (m *Memory) writeAccess(b []byte, width uint, value uint64) {
	if m.sharedSize == nil {
		write(b, width, value)
		return
	}
	checkedShared(native.SharedStore(unsafe.Pointer(unsafe.SliceData(b)), uint32(width), value))
}

func bytePointer(bytes []byte, offset uint64) unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(unsafe.SliceData(bytes)), uintptr(offset))
}
