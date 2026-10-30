//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"net/netip"
	"runtime"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

func TestDatagramSyscalls(t *testing.T) {
	for _, family := range []uint32{0, 1} {
		t.Run([]string{"ipv4", "ipv6"}[family], func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			client, err := newSocketFile(ctx, family, unix.SOCK_DGRAM)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				client.Close()
				if err := client.WaitClosed(nil); err != nil {
					t.Error(err)
				}
			}()
			server, err := newSocketFile(ctx, family, unix.SOCK_DGRAM)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				server.Close()
				if err := server.WaitClosed(nil); err != nil {
					t.Error(err)
				}
			}()
			ip := netip.MustParseAddr("127.0.0.1")
			if family == 1 {
				ip = netip.IPv6Loopback()
			}
			address := SocketAddress{Address: netip.AddrPortFrom(ip, 0)}
			for _, file := range []*socketFile{client, server} {
				if err := socketBind(file.fd, address); err != nil {
					t.Fatal(err)
				}
			}
			remote, err := socketAddressQuery(server.fd, false)
			if err != nil {
				t.Fatal(err)
			}
			local, err := socketAddressQuery(client.fd, false)
			if err != nil {
				t.Fatal(err)
			}
			buffer := make([]byte, 65535)
			if _, _, err := socketReceiveDatagram(server.fd, buffer); !errors.Is(err, unix.EAGAIN) {
				t.Fatal(err)
			}
			for _, payload := range [][]byte{nil, []byte("exact\x00datagram"), make([]byte, 4096)} {
				if n, err := socketSendDatagram(client.fd, payload, &remote); n != len(payload) || err != nil {
					t.Fatal(n, err)
				}
				if _, _, err := host.Poll(ctx, server.fd, true); err != nil {
					t.Fatal(err)
				}
				n, from, err := socketReceiveDatagram(server.fd, buffer)
				if err != nil || n != len(payload) || from.Address != local.Address || string(buffer[:n]) != string(payload) {
					t.Fatal(n, from, err)
				}
			}
			if err := socketConnect(client.fd, remote); err != nil {
				t.Fatal(err)
			}
			if n, err := socketSendDatagram(client.fd, []byte("connected"), nil); n != 9 || err != nil {
				t.Fatal(n, err)
			}
			if _, _, err := host.Poll(ctx, server.fd, true); err != nil {
				t.Fatal(err)
			}
			if n, _, err := socketReceiveDatagram(server.fd, buffer); n != 9 || err != nil {
				t.Fatal(n, err)
			}
			err = socketDisconnectDatagram(client.fd)
			if err != nil && !(runtime.GOOS == "darwin" && (errors.Is(err, unix.EAFNOSUPPORT) || errors.Is(err, unix.EINVAL))) {
				t.Fatal(err)
			}
			if _, err := socketAddressQuery(client.fd, true); !errors.Is(err, unix.ENOTCONN) {
				t.Fatal(err)
			}
			if _, err := socketSendDatagram(server.fd, make([]byte, 65536), &local); !errors.Is(err, unix.EMSGSIZE) {
				t.Fatal(err)
			}
			if _, err := socketSendDatagram(server.fd, nil, &SocketAddress{}); err == nil {
				t.Fatal("invalid destination accepted")
			}
		})
	}
}
