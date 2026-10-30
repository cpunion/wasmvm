//go:build darwin

package preview2

import "golang.org/x/sys/unix"

func socketLength4(address *unix.RawSockaddrInet4) { address.Len = unix.SizeofSockaddrInet4 }
func socketLength6(address *unix.RawSockaddrInet6) { address.Len = unix.SizeofSockaddrInet6 }

const tcpKeepIdle = unix.TCP_KEEPALIVE
