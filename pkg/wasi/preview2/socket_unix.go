//go:build linux || darwin

package preview2

import (
	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
	"unsafe"
)

var socketWire16 = host.SocketWire16
var socketWire32 = host.SocketWire32

func socketRawAddress(address SocketAddress) (unsafe.Pointer, uintptr, error) {
	p, n, err := host.SocketRawAddress(address)
	if err == unix.EINVAL {
		err = socketAddressError()
	}
	return p, n, err
}

var socketBind = host.SocketBind
var socketConnect = host.SocketConnect
var socketAddressQuery = host.SocketAddressQuery

func socketDecodeAddress(raw *unix.RawSockaddrAny, size uint32) (SocketAddress, error) {
	address, err := host.SocketDecodeAddress(raw, size)
	if err == unix.EINVAL {
		err = socketAddressError()
	}
	return address, err
}
