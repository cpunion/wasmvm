//go:build linux

package host

import "golang.org/x/sys/unix"

func SocketLength4(*unix.RawSockaddrInet4) {}
func SocketLength6(*unix.RawSockaddrInet6) {}

const TCPKeepIdle = unix.TCP_KEEPIDLE

func socketLengthUnspec(*unix.RawSockaddr) {}
