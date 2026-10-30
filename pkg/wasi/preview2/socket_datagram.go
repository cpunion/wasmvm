//go:build linux || darwin

package preview2

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Raw datagram syscalls preserve IPv6 flow-info, which unix.SockaddrInet6 omits.
// The caller holds a socketFile lease and retries EINTR with cancellation checks.
func socketSendDatagram(fd int, data []byte, address *SocketAddress) (int, error) {
	var raw unsafe.Pointer
	var size uintptr
	var err error
	if address != nil {
		raw, size, err = socketRawAddress(*address)
		if err != nil {
			return 0, err
		}
	}
	n, _, errno := unix.Syscall6(unix.SYS_SENDTO, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(data))), uintptr(len(data)), uintptr(unix.MSG_NOSIGNAL), uintptr(raw), size)
	runtime.KeepAlive(data)
	runtime.KeepAlive(raw)
	if errno != 0 {
		return 0, errno
	}
	return int(n), nil
}

func socketReceiveDatagram(fd int, data []byte) (int, SocketAddress, error) {
	var raw unix.RawSockaddrAny
	size := uint32(unsafe.Sizeof(raw))
	n, _, errno := unix.Syscall6(unix.SYS_RECVFROM, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(data))), uintptr(len(data)), 0, uintptr(unsafe.Pointer(&raw)), uintptr(unsafe.Pointer(&size)))
	runtime.KeepAlive(data)
	runtime.KeepAlive(&raw)
	runtime.KeepAlive(&size)
	if errno != 0 {
		return 0, SocketAddress{}, errno
	}
	address, err := socketDecodeAddress(&raw, size)
	return int(n), address, err
}

func socketDisconnectDatagram(fd int) error {
	raw := unix.RawSockaddr{Family: unix.AF_UNSPEC}
	_, _, errno := unix.Syscall(unix.SYS_CONNECT, uintptr(fd), uintptr(unsafe.Pointer(&raw)), unsafe.Sizeof(raw))
	runtime.KeepAlive(&raw)
	if errno != 0 {
		return errno
	}
	return nil
}
