//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"os"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type netError string

func (e netError) Error() string { return string(e) }

var netCases = func() map[string]uint32 {
	m := map[string]uint32{}
	for i, f := range socketsWITNamedTypes["error-code"].fields {
		m[f.name] = uint32(i)
	}
	return m
}()

func netCode(err error) uint32 {
	var named netError
	if errors.As(err, &named) {
		if n, ok := netCases[string(named)]; ok {
			return n
		}
	}
	if errors.Is(err, wasm.ErrLimit) {
		return netCases["out-of-memory"]
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return netCases["timeout"]
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, os.ErrClosed) || errors.Is(err, io.ErrClosedPipe) {
		return netCases["invalid-state"]
	}
	for _, e := range netErrnos {
		if errors.Is(err, e.err) {
			return netCases[e.name]
		}
	}
	return netCases["other"]
}
func netCompletion(err error) uint32 {
	if err == nil || errors.Is(err, io.EOF) {
		return 0
	}
	return netCode(err) + 1
}
func netCompletionValue(code uint32) component.Value {
	if code == 0 {
		return component.Value{}
	}
	v := component.Value{Case: code - 1}
	if v.Case == netCases["other"] {
		v.Elements = []component.Value{{}}
	}
	return component.Value{Case: 1, Elements: []component.Value{v}}
}
func netResult(v *component.Value, err error) *component.Value {
	if err != nil {
		r := netCompletionValue(netCompletion(err))
		return &r
	}
	r := &component.Value{}
	if v != nil {
		r.Elements = []component.Value{*v}
	}
	return r
}

var netErrnos = []struct {
	err  error
	name string
}{
	{unix.EACCES, "access-denied"}, {unix.EPERM, "access-denied"}, {unix.EAFNOSUPPORT, "invalid-argument"}, {unix.EPROTONOSUPPORT, "not-supported"}, {unix.ENOPROTOOPT, "not-supported"}, {unix.EOPNOTSUPP, "not-supported"}, {unix.ENOSYS, "not-supported"},
	{unix.EINVAL, "invalid-argument"}, {unix.EDESTADDRREQ, "invalid-argument"}, {unix.ENOMEM, "out-of-memory"}, {unix.ENOBUFS, "out-of-memory"}, {unix.EMFILE, "out-of-memory"}, {unix.ENFILE, "out-of-memory"}, {unix.ETIMEDOUT, "timeout"},
	{unix.EBADF, "invalid-state"}, {unix.ENOTSOCK, "invalid-state"}, {unix.ENOTCONN, "invalid-state"}, {unix.EISCONN, "invalid-state"}, {unix.EALREADY, "invalid-state"},
	{unix.EADDRNOTAVAIL, "address-not-bindable"}, {unix.EADDRINUSE, "address-in-use"}, {unix.ENETDOWN, "remote-unreachable"}, {unix.ENETUNREACH, "remote-unreachable"}, {unix.EHOSTUNREACH, "remote-unreachable"},
	{unix.ECONNREFUSED, "connection-refused"}, {unix.ECONNRESET, "connection-reset"}, {unix.ECONNABORTED, "connection-aborted"}, {unix.EPIPE, "connection-broken"}, {unix.EMSGSIZE, "datagram-too-large"},
}

func netIP(ip netip.Addr) component.Value {
	v := component.Value{}
	if ip.Is4() {
		for _, b := range ip.As4() {
			v.Elements = append(v.Elements, component.Value{Bits: uint64(b)})
		}
	} else {
		v.Case = 1
		a := ip.As16()
		for i := 0; i < 16; i += 2 {
			v.Elements = append(v.Elements, component.Value{Bits: uint64(a[i])<<8 | uint64(a[i+1])})
		}
	}
	return component.Value{Case: v.Case, Elements: []component.Value{{Elements: v.Elements}}}
}
func netAddress(a SocketAddress) component.Value {
	ip := netIP(a.Address.Addr())
	f := []component.Value{{Bits: uint64(a.Address.Port())}, ip.Elements[0]}
	if ip.Case == 1 {
		f = []component.Value{f[0], {Bits: uint64(a.FlowInfo)}, f[1], {Bits: uint64(a.ScopeID)}}
	}
	return component.Value{Case: ip.Case, Elements: []component.Value{{Elements: f}}}
}
func netParse(v component.Value) SocketAddress {
	f := v.Elements[0].Elements
	var a SocketAddress
	if v.Case == 0 {
		b := f[1].Elements
		a.Address = netip.AddrPortFrom(netip.AddrFrom4([4]byte{byte(b[0].Bits), byte(b[1].Bits), byte(b[2].Bits), byte(b[3].Bits)}), uint16(f[0].Bits))
	} else {
		var raw [16]byte
		for i, b := range f[2].Elements {
			raw[i*2], raw[i*2+1] = byte(b.Bits>>8), byte(b.Bits)
		}
		a.Address = netip.AddrPortFrom(netip.AddrFrom16(raw), uint16(f[0].Bits))
		a.FlowInfo = uint32(f[1].Bits)
		a.ScopeID = uint32(f[3].Bits)
	}
	return a
}
func netCheckAddress(family uint32, a SocketAddress, remote bool) error {
	ip := a.Address.Addr()
	if family == 0 && !ip.Is4() || family == 1 && !ip.Is6() || ip.Is4In6() || ip.IsMulticast() || ip == netip.AddrFrom4([4]byte{255, 255, 255, 255}) || remote && (ip.IsUnspecified() || a.Address.Port() == 0) {
		return netError("invalid-argument")
	}
	return nil
}
func lookupResult(addresses []netip.Addr, err error) *component.Value {
	if err == nil {
		v := &component.Value{}
		for _, a := range addresses {
			v.Elements = append(v.Elements, netIP(a))
		}
		return &component.Value{Elements: []component.Value{*v}}
	}
	code := uint32(5)
	var named netError
	var dns *net.DNSError
	if errors.As(err, &named) {
		switch named {
		case "access-denied":
			code = 0
		case "invalid-argument":
			code = 1
		case "name-unresolvable":
			code = 2
		}
	} else if errors.As(err, &dns) {
		code = 4
		if dns.IsNotFound {
			code = 2
		} else if dns.IsTemporary || dns.IsTimeout {
			code = 3
		}
	}
	v := component.Value{Case: code}
	if code == 5 {
		v.Elements = []component.Value{{}}
	}
	return &component.Value{Case: 1, Elements: []component.Value{v}}
}
