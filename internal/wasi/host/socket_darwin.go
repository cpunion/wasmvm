//go:build darwin

package host

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

func SocketLength4(address *unix.RawSockaddrInet4) { address.Len = unix.SizeofSockaddrInet4 }
func SocketLength6(address *unix.RawSockaddrInet6) { address.Len = unix.SizeofSockaddrInet6 }

const TCPKeepIdle = unix.TCP_KEEPALIVE

func socketLengthUnspec(address *unix.RawSockaddr) { address.Len = uint8(unsafe.Sizeof(*address)) }
