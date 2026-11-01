//go:build linux || darwin

package host

import (
	"encoding/binary"
	"net/netip"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

// SocketAddress retains the complete WIT address, including IPv6 flow/scope.
type SocketAddress struct {
	Address           netip.AddrPort
	FlowInfo, ScopeID uint32
}

func SocketWire16(value uint16) uint16 {
	var bytes [2]byte
	binary.BigEndian.PutUint16(bytes[:], value)
	return binary.NativeEndian.Uint16(bytes[:])
}
func SocketWire32(value uint32) uint32 {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], value)
	return binary.NativeEndian.Uint32(bytes[:])
}

// The common unix.SockaddrInet6 representation omits flow-info. Preserve the
// complete WIT address in the native ABI instead of silently discarding it.
func SocketRawAddress(address SocketAddress) (unsafe.Pointer, uintptr, error) {
	if !address.Address.IsValid() || address.Address.Addr().Zone() != "" {
		return nil, 0, unix.EINVAL
	}
	ip := address.Address.Addr()
	if ip.Is4() {
		if address.FlowInfo != 0 || address.ScopeID != 0 {
			return nil, 0, unix.EINVAL
		}
		raw := &unix.RawSockaddrInet4{Family: unix.AF_INET, Port: SocketWire16(address.Address.Port()), Addr: ip.As4()}
		SocketLength4(raw)
		return unsafe.Pointer(raw), unix.SizeofSockaddrInet4, nil
	}
	raw := &unix.RawSockaddrInet6{Family: unix.AF_INET6, Port: SocketWire16(address.Address.Port()), Flowinfo: SocketWire32(address.FlowInfo), Addr: ip.As16(), Scope_id: address.ScopeID}
	SocketLength6(raw)
	return unsafe.Pointer(raw), unix.SizeofSockaddrInet6, nil
}
func SocketBind(fd int, address SocketAddress) error {
	raw, size, err := SocketRawAddress(address)
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
func SocketConnect(fd int, address SocketAddress) error {
	raw, size, err := SocketRawAddress(address)
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
func SocketAddressQuery(fd int, peer bool) (SocketAddress, error) {
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
	return SocketDecodeAddress(&raw, size)
}

func SocketDecodeAddress(raw *unix.RawSockaddrAny, size uint32) (SocketAddress, error) {
	switch raw.Addr.Family {
	case unix.AF_INET:
		if size < unix.SizeofSockaddrInet4 {
			return SocketAddress{}, unix.EINVAL
		}
		address := (*unix.RawSockaddrInet4)(unsafe.Pointer(raw))
		return SocketAddress{Address: netip.AddrPortFrom(netip.AddrFrom4(address.Addr), SocketWire16(address.Port))}, nil
	case unix.AF_INET6:
		if size < unix.SizeofSockaddrInet6 {
			return SocketAddress{}, unix.EINVAL
		}
		address := (*unix.RawSockaddrInet6)(unsafe.Pointer(raw))
		return SocketAddress{Address: netip.AddrPortFrom(netip.AddrFrom16(address.Addr), SocketWire16(address.Port)), FlowInfo: SocketWire32(address.Flowinfo), ScopeID: address.Scope_id}, nil
	}
	return SocketAddress{}, unix.EINVAL
}
