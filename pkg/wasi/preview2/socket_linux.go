//go:build linux

package preview2

import "golang.org/x/sys/unix"

func socketLength4(*unix.RawSockaddrInet4) {}
func socketLength6(*unix.RawSockaddrInet6) {}

const tcpKeepIdle = unix.TCP_KEEPIDLE
