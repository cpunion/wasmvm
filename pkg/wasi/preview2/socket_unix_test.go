//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

func TestSocketRawAddressWireFormat(t *testing.T) {
	address := SocketAddress{Address: netip.MustParseAddrPort("[2001:db8::1]:4660"), FlowInfo: 0x12345678, ScopeID: 42}
	raw, size, err := socketRawAddress(address)
	if err != nil || size != unix.SizeofSockaddrInet6 {
		t.Fatal(size, err)
	}
	ip := (*unix.RawSockaddrInet6)(raw)
	if ip.Family != unix.AF_INET6 || socketWire16(ip.Port) != 4660 || socketWire32(ip.Flowinfo) != address.FlowInfo || ip.Scope_id != 42 || ip.Addr != address.Address.Addr().As16() {
		t.Fatal(ip)
	}
	if _, _, err := socketRawAddress(SocketAddress{}); networkError(err) != NetworkInvalidArgument {
		t.Fatal(err)
	}
	if _, _, err := socketRawAddress(SocketAddress{Address: netip.MustParseAddrPort("127.0.0.1:80"), ScopeID: 1}); networkError(err) != NetworkInvalidArgument {
		t.Fatal(err)
	}
	if _, err := socketAddressQuery(-1, false); !errors.Is(err, unix.EBADF) {
		t.Fatal(err)
	}
	if err := socketBind(-1, address); !errors.Is(err, unix.EBADF) {
		t.Fatal(err)
	}
	if err := socketConnect(-1, address); !errors.Is(err, unix.EBADF) {
		t.Fatal(err)
	}
}
func TestSocketRawLoopbackConnectAndQueries(t *testing.T) {
	for _, family := range []uint32{0, 1} {
		server, err := newSocketFile(nil, family, unix.SOCK_STREAM)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			server.Close()
			if err := server.WaitClosed(nil); err != nil {
				t.Error(err)
			}
		}()
		listen, err := server.begin(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer server.end()
		address := SocketAddress{Address: netip.MustParseAddrPort("127.0.0.1:0")}
		if family == 1 {
			address.Address = netip.MustParseAddrPort("[::1]:0")
		}
		if err := socketBind(listen, address); err != nil {
			t.Fatal(err)
		}
		if err := unix.Listen(listen, 2); err != nil {
			t.Fatal(err)
		}
		bound, err := socketAddressQuery(listen, false)
		if err != nil || bound.Address.Port() == 0 || bound.Address.Addr() != address.Address.Addr() {
			t.Fatal(bound, err)
		}
		client, err := newSocketFile(nil, family, unix.SOCK_STREAM)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			client.Close()
			if err := client.WaitClosed(nil); err != nil {
				t.Error(err)
			}
		}()
		fd, err := client.begin(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer client.end()
		if _, err := socketAddressQuery(fd, true); err == nil {
			t.Fatal("unconnected socket has peer")
		}
		if err := socketConnect(fd, bound); err != nil && !errors.Is(err, unix.EINPROGRESS) {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, _, err = host.Poll(ctx, fd, false)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		status, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err != nil || status != 0 {
			t.Fatal(status, err)
		}
		peer, err := socketAddressQuery(fd, true)
		if err != nil || peer != bound {
			t.Fatal(peer, bound, err)
		}
		local, err := socketAddressQuery(fd, false)
		if err != nil || local.Address.Port() == 0 || local.Address.Addr() != address.Address.Addr() {
			t.Fatal(local, err)
		}
	}
}
