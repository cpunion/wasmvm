//go:build linux || darwin

package preview2

import (
	"encoding/binary"
	"net/netip"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

func socketWire16(value uint16) uint16 {
	var bytes [2]byte
	binary.BigEndian.PutUint16(bytes[:], value)
	return binary.NativeEndian.Uint16(bytes[:])
}
func socketWire32(value uint32) uint32 {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], value)
	return binary.NativeEndian.Uint32(bytes[:])
}

// The common unix.SockaddrInet6 representation omits flow-info. Preserve the
// complete WIT address in the native ABI instead of silently discarding it.
func socketRawAddress(address SocketAddress) (unsafe.Pointer, uintptr, error) {
	if !address.Address.IsValid() || address.Address.Addr().Zone() != "" {
		return nil, 0, socketAddressError()
	}
	ip := address.Address.Addr()
	if ip.Is4() {
		if address.FlowInfo != 0 || address.ScopeID != 0 {
			return nil, 0, socketAddressError()
		}
		raw := &unix.RawSockaddrInet4{Family: unix.AF_INET, Port: socketWire16(address.Address.Port()), Addr: ip.As4()}
		socketLength4(raw)
		return unsafe.Pointer(raw), unix.SizeofSockaddrInet4, nil
	}
	raw := &unix.RawSockaddrInet6{Family: unix.AF_INET6, Port: socketWire16(address.Address.Port()), Flowinfo: socketWire32(address.FlowInfo), Addr: ip.As16(), Scope_id: address.ScopeID}
	socketLength6(raw)
	return unsafe.Pointer(raw), unix.SizeofSockaddrInet6, nil
}
func socketBind(fd int, address SocketAddress) error {
	raw, size, err := socketRawAddress(address)
	if err != nil {
		return err
	}
	_, _, errno := unix.Syscall(unix.SYS_BIND, uintptr(fd), uintptr(raw), size)
	runtime.KeepAlive(raw)
	if errno != 0 {
		return errno
	}
	return nil
}
func socketConnect(fd int, address SocketAddress) error {
	raw, size, err := socketRawAddress(address)
	if err != nil {
		return err
	}
	_, _, errno := unix.Syscall(unix.SYS_CONNECT, uintptr(fd), uintptr(raw), size)
	runtime.KeepAlive(raw)
	if errno != 0 {
		return errno
	}
	return nil
}
func socketAddressQuery(fd int, peer bool) (SocketAddress, error) {
	var raw unix.RawSockaddrAny
	size := uint32(unsafe.Sizeof(raw))
	query := uintptr(unix.SYS_GETSOCKNAME)
	if peer {
		query = unix.SYS_GETPEERNAME
	}
	_, _, errno := unix.Syscall(query, uintptr(fd), uintptr(unsafe.Pointer(&raw)), uintptr(unsafe.Pointer(&size)))
	runtime.KeepAlive(&raw)
	runtime.KeepAlive(&size)
	if errno != 0 {
		return SocketAddress{}, errno
	}
	switch raw.Addr.Family {
	case unix.AF_INET:
		if size < unix.SizeofSockaddrInet4 {
			return SocketAddress{}, socketAddressError()
		}
		address := (*unix.RawSockaddrInet4)(unsafe.Pointer(&raw))
		return SocketAddress{Address: netip.AddrPortFrom(netip.AddrFrom4(address.Addr), socketWire16(address.Port))}, nil
	case unix.AF_INET6:
		if size < unix.SizeofSockaddrInet6 {
			return SocketAddress{}, socketAddressError()
		}
		address := (*unix.RawSockaddrInet6)(unsafe.Pointer(&raw))
		return SocketAddress{Address: netip.AddrPortFrom(netip.AddrFrom16(address.Addr), socketWire16(address.Port)), FlowInfo: socketWire32(address.Flowinfo), ScopeID: address.Scope_id}, nil
	}
	return SocketAddress{}, socketAddressError()
}
