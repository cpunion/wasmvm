//go:build linux || darwin

package preview2

import (
	"context"
	"net/netip"

	"github.com/cpunion/wasmvm/pkg/component"
)

func socketCode(code NetworkErrorCode) error   { return &NetworkError{Code: code} }
func socketFailure(err error) *component.Value { return networkFailure(socketError(err)) }
func socketCheckAddress(family uint32, address SocketAddress, remote bool) error {
	ip := address.Address.Addr()
	if family == 0 && !ip.Is4() || family == 1 && !ip.Is6() {
		return socketCode(NetworkInvalidArgument)
	}
	if ip.IsMulticast() || ip == netip.AddrFrom4([4]byte{255, 255, 255, 255}) || ip.Is4In6() || remote && (address.Address.Port() == 0 || ip.IsUnspecified()) {
		return socketCode(NetworkInvalidArgument)
	}
	return nil
}
func socketAuthorize(ctx, owner context.Context, address SocketAddress, callback func(context.Context, SocketAddress) error) (err error) {
	if callback == nil {
		return socketCode(NetworkAccessDenied)
	}
	joint, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(owner, cancel)
	defer cancel()
	defer stop()
	complete := false
	defer func() {
		if !complete {
			_ = recover()
			err = socketCode(NetworkAccessDenied)
		}
	}()
	err = callback(joint, address)
	complete = true
	if joint.Err() != nil {
		return joint.Err()
	}
	if err != nil {
		return &NetworkError{Code: NetworkAccessDenied, Cause: err}
	}
	return nil
}
