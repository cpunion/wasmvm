//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

type socketsAPI struct {
	t  *testing.T
	in *component.Instance
}

func (a socketsAPI) call(udp bool, name string, args ...component.Value) *component.Value {
	a.t.Helper()
	resource := "tcp-socket"
	if udp {
		resource = "udp-socket"
	}
	prefix := "[method]"
	if name == "create" {
		prefix = "[static]"
	}
	f, err := a.in.Function("types", prefix+resource+"."+name)
	if err != nil {
		a.t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := f.Call(ctx, args, component.CallOptions{})
	if err != nil {
		a.t.Fatal(name, err)
	}
	return r.Value
}
func (a socketsAPI) ok(udp bool, name string, args ...component.Value) component.Value {
	a.t.Helper()
	r := a.call(udp, name, args...)
	if r.Case != 0 {
		a.t.Fatal(name, "socket error", r)
	}
	if len(r.Elements) > 0 {
		return r.Elements[0]
	}
	return component.Value{}
}
func (a socketsAPI) fail(udp bool, name, want string, args ...component.Value) {
	a.t.Helper()
	v := a.call(udp, name, args...)
	if v.Case != 1 || v.Elements[0].Case != netCases[want] {
		a.t.Fatal(name, want, v)
	}
}
func (a socketsAPI) drop(v component.Value) {
	a.t.Helper()
	if _, err := a.in.DropResource(nil, v.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func (a socketsAPI) dropAsync(v component.Value) {
	a.t.Helper()
	if _, err := a.in.DropAsync(nil, v.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func (a socketsAPI) completion(v component.Value, want string) {
	a.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := a.in.ReadAsync(ctx, v.Resource, 1, component.CallOptions{})
	if err != nil || len(r.Values) != 1 {
		a.t.Fatal(r, err)
	}
	x := r.Values[0]
	if want == "" {
		if x.Case != 0 {
			a.t.Fatal(x)
		}
	} else if x.Case != 1 || x.Elements[0].Case != netCases[want] {
		a.t.Fatal(want, x)
	}
	a.dropAsync(v)
}
func socketsOwner(t *testing.T, config SocketsConfig) (*Sockets, *IO) {
	t.Helper()
	process, err := NewIO(IOConfig{MaxReadBytes: 3, MaxWriteBytes: 3})
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewSockets(process, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
		closeIO(t, process)
	})
	return p, process
}
func socketsAllow(context.Context, SocketAddress) error { return nil }

func TestSocketShutdownDrainsDescriptorLeases(t *testing.T) {
	p, _ := socketsOwner(t, SocketsConfig{})
	s, err := p.create(0, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.acquire(); err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	if err := p.begin(); err != nil {
		t.Fatal(err)
	}
	go func() { close(started); <-release; s.release(); p.end() }()
	<-started
	p.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	wait, stop := context.WithTimeout(ctx, time.Millisecond)
	if err := p.WaitClosed(wait); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("socket shutdown did not join admitted descriptor work", err)
	}
	stop()
	if _, err := unix.FcntlInt(uintptr(s.fd), unix.F_GETFD, 0); err != nil {
		t.Fatal("descriptor closed while admitted work retained its lease", err)
	}
	close(release)
	if err := p.WaitClosed(ctx); err != nil {
		t.Fatal(err)
	}
}
func loopAddress(family uint64, port uint16) component.Value {
	ip := netip.MustParseAddr("127.0.0.1")
	if family == 1 {
		ip = netip.MustParseAddr("::1")
	}
	return netAddress(SocketAddress{Address: netip.AddrPortFrom(ip, port)})
}
func TestSocketsAsyncCoreDatagramGuest(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		listener, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()
		listener.SetReadDeadline(time.Now().Add(3 * time.Second))
		p, _ := socketsOwner(t, SocketsConfig{AllowBind: socketsAllow, AllowSend: socketsAllow})
		in := instantiate(t, compileFixture(t, backend, "sockets-guest"), p.Imports())
		r := call(t, in, "run", component.Value{Bits: uint64(listener.LocalAddr().(*net.UDPAddr).Port)})
		if r.Bits != 0 {
			t.Fatal(r)
		}
		var data [8]byte
		n, _, err := listener.ReadFromUDP(data[:])
		if err != nil || string(data[:n]) != "abc" {
			t.Fatal(n, err)
		}
	})
}
func TestSocketsTCPAcceptStreamsAndOwnership(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _ := socketsOwner(t, SocketsConfig{AllowBind: socketsAllow})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		api := socketsAPI{t, in}
		listener := api.ok(false, "create", component.Value{})
		api.ok(false, "bind", listener, loopAddress(0, 0))
		api.ok(false, "set-hop-limit", listener, component.Value{Bits: 42})
		stream := api.ok(false, "listen", listener)
		local := netParse(api.ok(false, "get-local-address", listener))
		client, err := net.DialTCP("tcp4", nil, net.TCPAddrFromAddrPort(local.Address))
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()
		client.SetDeadline(time.Now().Add(3 * time.Second))
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		readiness, err := in.ReadAsync(ctx, stream.Resource, 0, component.CallOptions{})
		if err != nil || readiness.Count != 0 || readiness.Closed {
			t.Fatal(readiness, err)
		}
		r, err := in.ReadAsync(ctx, stream.Resource, 1, component.CallOptions{})
		if err != nil || len(r.Values) != 1 {
			t.Fatal(r, err)
		}
		socket := r.Values[0]
		if api.ok(false, "get-hop-limit", socket).Bits != 42 {
			t.Fatal("accepted socket lost inherited option")
		}
		api.drop(listener)
		api.dropAsync(stream)
		if api.call(false, "get-address-family", socket).Bits != 0 {
			t.Fatal("family")
		}
		if netParse(api.ok(false, "get-remote-address", socket)).Address.Port() == 0 {
			t.Fatal("peer")
		}
		pair := api.call(false, "receive", socket)
		if _, err := client.Write([]byte("hello")); err != nil {
			t.Fatal(err)
		}
		if err := client.CloseWrite(); err != nil {
			t.Fatal(err)
		}
		var got string
		for {
			batch, err := in.ReadAsync(ctx, pair.Elements[0].Resource, 3, component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			got += batch.Bytes
			if batch.Closed {
				break
			}
		}
		if got != "hello" {
			t.Fatal(got)
		}
		api.dropAsync(pair.Elements[0])
		api.completion(pair.Elements[1], "")
		f, _ := in.Function("types", "[method]tcp-socket.send")
		_, typ := f.Info().Parameter(1)
		pending := "world"
		source, err := in.NewSource(nil, typ, component.AsyncSource{Read: func(_ *component.Caller, limit uint32) (*component.ReadCompletion, error) {
			n := min(len(pending), int(limit))
			text := pending[:n]
			pending = pending[n:]
			closed := pending == ""
			r := component.NewReadCompletion()
			r.Resolve(func(*component.Caller) (component.ReadResult, error) {
				return component.ReadResult{Bytes: text, Closed: closed}, nil
			})
			return r, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		result := api.call(false, "send", socket, source)
		api.drop(socket)
		data, err := io.ReadAll(client)
		if err != nil || string(data) != "world" {
			t.Fatal(string(data), err)
		}
		api.completion(*result, "")
	})
}
func TestSocketsUDPAssociationAndDatagrams(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, family := range []uint64{0, 1} {
			p, _ := socketsOwner(t, SocketsConfig{AllowBind: socketsAllow, AllowConnect: socketsAllow, AllowSend: socketsAllow, AllowReceive: socketsAllow})
			in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
			api := socketsAPI{t, in}
			socket := api.ok(true, "create", component.Value{Bits: family})
			api.fail(true, "get-local-address", "invalid-state", socket)
			api.fail(true, "disconnect", "invalid-state", socket)
			api.ok(true, "bind", socket, loopAddress(family, 0))
			address := netParse(api.ok(true, "get-local-address", socket))
			network := "udp4"
			if family == 1 {
				network = "udp6"
			}
			peer, err := net.ListenUDP(network, net.UDPAddrFromAddrPort(netip.AddrPortFrom(address.Address.Addr(), 0)))
			if err != nil {
				t.Fatal(err)
			}
			defer peer.Close()
			peer.SetDeadline(time.Now().Add(3 * time.Second))
			remote := netAddress(SocketAddress{Address: peer.LocalAddr().(*net.UDPAddr).AddrPort()})
			api.ok(true, "send", socket, component.Value{Text: "a"}, component.Value{Case: 1, Elements: []component.Value{remote}})
			var data [16]byte
			n, from, err := peer.ReadFromUDP(data[:])
			if err != nil || string(data[:n]) != "a" {
				t.Fatal(n, err)
			}
			api.ok(true, "connect", socket, remote)
			if netParse(api.ok(true, "get-remote-address", socket)).Address != peer.LocalAddr().(*net.UDPAddr).AddrPort() {
				t.Fatal("peer")
			}
			api.ok(true, "send", socket, component.Value{Text: ""}, component.Value{})
			n, _, err = peer.ReadFromUDP(data[:])
			if err != nil || n != 0 {
				t.Fatal(n, err)
			}
			if _, err := peer.WriteToUDP([]byte("reply"), from); err != nil {
				t.Fatal(err)
			}
			message := api.ok(true, "receive", socket)
			if message.Elements[0].Text != "reply" || netParse(message.Elements[1]).Address != peer.LocalAddr().(*net.UDPAddr).AddrPort() {
				t.Fatal(message)
			}
			api.ok(true, "disconnect", socket)
			api.fail(true, "get-remote-address", "invalid-state", socket)
			if api.call(true, "get-address-family", socket).Bits != family {
				t.Fatal("family")
			}
			api.drop(socket)
		}
	})
}
func TestSocketsBindingValidation(t *testing.T) {
	p, _ := socketsOwner(t, SocketsConfig{})
	for _, name := range []string{"sockets-wrong-error", "sockets-wrong-async"} {
		m := compileFixture(t, "interpreter", name)
		if _, err := m.Instantiate(nil, p.Imports(), component.CallOptions{}); err == nil {
			t.Fatal(name)
		}
	}
}

func TestSocketsTCPConnectAndOptions(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, family := range []uint64{0, 1} {
			address := netParse(loopAddress(family, 0))
			network := "tcp4"
			if family == 1 {
				network = "tcp6"
			}
			listener, err := net.ListenTCP(network, net.TCPAddrFromAddrPort(address.Address))
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			listener.SetDeadline(time.Now().Add(3 * time.Second))
			p, _ := socketsOwner(t, SocketsConfig{AllowConnect: socketsAllow})
			in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
			api := socketsAPI{t, in}
			socket := api.ok(false, "create", component.Value{Bits: family})
			options := []struct {
				name  string
				value uint64
			}{{"keep-alive-enabled", 1}, {"keep-alive-idle-time", 1e9}, {"keep-alive-interval", 1e9}, {"keep-alive-count", 2}, {"hop-limit", 64}, {"receive-buffer-size", 8192}, {"send-buffer-size", 8192}}
			for _, option := range options {
				api.ok(false, "set-"+option.name, socket, component.Value{Bits: option.value})
				if api.ok(false, "get-"+option.name, socket).Bits == 0 {
					t.Fatal(option.name)
				}
				if option.name != "keep-alive-enabled" {
					api.fail(false, "set-"+option.name, "invalid-argument", socket, component.Value{})
				}
			}
			api.fail(false, "bind", "access-denied", socket, loopAddress(family, 0))
			api.fail(false, "connect", "invalid-argument", socket, loopAddress(family, 0))
			remote := netAddress(SocketAddress{Address: listener.Addr().(*net.TCPAddr).AddrPort()})
			api.ok(false, "connect", socket, remote)
			peer, err := listener.AcceptTCP()
			if err != nil {
				t.Fatal(err)
			}
			defer peer.Close()
			if netParse(api.ok(false, "get-local-address", socket)).Address.Port() == 0 || netParse(api.ok(false, "get-remote-address", socket)).Address != listener.Addr().(*net.TCPAddr).AddrPort() {
				t.Fatal("connected addresses")
			}
			api.fail(false, "connect", "invalid-state", socket, remote)
			api.fail(false, "listen", "invalid-state", socket)
			api.fail(false, "bind", "invalid-state", socket, loopAddress(family, 0))
			pair := api.call(false, "receive", socket)
			duplicate := api.call(false, "receive", socket)
			api.completion(duplicate.Elements[1], "invalid-state")
			api.dropAsync(duplicate.Elements[0])
			// Dropping the invalid duplicate must not shut down the original direction.
			peer.SetDeadline(time.Now().Add(3 * time.Second))
			peer.Write([]byte("ok"))
			peer.CloseWrite()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			var got string
			for {
				r, err := in.ReadAsync(ctx, pair.Elements[0].Resource, 3, component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				got += r.Bytes
				if r.Closed {
					break
				}
			}
			cancel()
			if got != "ok" {
				t.Fatal(got)
			}
			api.dropAsync(pair.Elements[0])
			api.completion(pair.Elements[1], "")
			api.drop(socket)
		}
	})
}
func TestSocketsUDPOptionsPoliciesAndErrors(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _ := socketsOwner(t, SocketsConfig{AllowBind: socketsAllow})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		api := socketsAPI{t, in}
		socket := api.ok(true, "create", component.Value{})
		api.fail(true, "send", "invalid-argument", socket, component.Value{Text: "x"}, component.Value{})
		api.fail(true, "receive", "invalid-state", socket)
		api.fail(true, "bind", "invalid-argument", socket, loopAddress(1, 0))
		api.ok(true, "bind", socket, loopAddress(0, 0))
		api.fail(true, "bind", "invalid-state", socket, loopAddress(0, 0))
		for _, name := range []string{"unicast-hop-limit", "receive-buffer-size", "send-buffer-size"} {
			api.ok(true, "set-"+name, socket, component.Value{Bits: 64})
			if api.ok(true, "get-"+name, socket).Bits == 0 {
				t.Fatal(name)
			}
			api.fail(true, "set-"+name, "invalid-argument", socket, component.Value{})
		}
		api.fail(true, "send", "access-denied", socket, component.Value{Text: "x"}, component.Value{Case: 1, Elements: []component.Value{loopAddress(0, 80)}})
		api.fail(true, "connect", "access-denied", socket, loopAddress(0, 80))
		api.fail(true, "receive", "access-denied", socket)
		api.fail(true, "send", "datagram-too-large", socket, component.Value{Text: string(make([]byte, 65537))}, component.Value{})
		api.drop(socket)
	})
}
func TestSocketsNameLookupContractsAndLimits(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _ := socketsOwner(t, SocketsConfig{AllowNameLookup: true, Resolve: func(_ context.Context, name string) ([]netip.Addr, error) {
			if name != "xn--bcher-kva.example" {
				t.Error(name)
			}
			return []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::ffff:127.0.0.1"), netip.MustParseAddr("::1"), {}, netip.MustParseAddr("fe80::1%en0")}, nil
		}})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		f, _ := in.Function("lookup", "resolve-addresses")
		resolve := func(name string) *component.Value {
			t.Helper()
			v, err := f.Call(nil, []component.Value{{Text: name}}, component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			return v.Value
		}
		for _, name := range []string{"127.0.0.1", "::ffff:127.0.0.1", "::1"} {
			v := resolve(name)
			if v.Case != 0 || len(v.Elements[0].Elements) != 1 {
				t.Fatal(name, v)
			}
		}
		v := resolve("bücher.example")
		if v.Case != 0 || len(v.Elements[0].Elements) != 2 {
			t.Fatal(v)
		}
		for _, name := range []string{"", "a\x00b", "fe80::1%en0"} {
			v := resolve(name)
			if v.Case != 1 || v.Elements[0].Case != 1 {
				t.Fatal(name, v)
			}
		}
		for _, item := range []struct {
			config SocketsConfig
			name   string
			code   uint32
		}{
			{SocketsConfig{}, "denied.example", 0},
			{SocketsConfig{MaxNameBytes: 3}, "long-name", 1},
			{SocketsConfig{AllowNameLookup: true, Resolve: func(context.Context, string) ([]netip.Addr, error) { return nil, nil }}, "empty.example", 2},
			{SocketsConfig{AllowNameLookup: true, Resolve: func(context.Context, string) ([]netip.Addr, error) { return nil, &net.DNSError{IsNotFound: true} }}, "missing.example", 2},
			{SocketsConfig{AllowNameLookup: true, Resolve: func(context.Context, string) ([]netip.Addr, error) { return nil, &net.DNSError{IsTemporary: true} }}, "temporary.example", 3},
			{SocketsConfig{AllowNameLookup: true, Resolve: func(context.Context, string) ([]netip.Addr, error) { return nil, &net.DNSError{} }}, "bad.example", 4},
			{SocketsConfig{AllowNameLookup: true, Resolve: func(context.Context, string) ([]netip.Addr, error) { panic("resolver") }}, "panic.example", 5},
			{SocketsConfig{AllowNameLookup: true, MaxAddresses: 1, Resolve: func(context.Context, string) ([]netip.Addr, error) {
				return []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")}, nil
			}}, "limit.example", 5},
		} {
			owner, _ := socketsOwner(t, item.config)
			other := instantiate(t, compileFixture(t, backend, "sockets-api"), owner.Imports())
			lookup, _ := other.Function("lookup", "resolve-addresses")
			r, err := lookup.Call(nil, []component.Value{{Text: item.name}}, component.CallOptions{})
			if err != nil || r.Value.Case != 1 || r.Value.Elements[0].Case != item.code {
				t.Fatal(item.name, r, err)
			}
		}
	})
}

func TestSocketsAdmissionCancellationAndClosedOwners(t *testing.T) {
	if _, err := NewSockets(nil, SocketsConfig{}); err == nil {
		t.Fatal("missing IO")
	}
	ioOwner, _ := NewIO(IOConfig{})
	defer closeIO(t, ioOwner)
	for _, cfg := range []SocketsConfig{{MaxSockets: 1<<20 + 1}, {MaxAddresses: 1<<20 + 1}, {MaxNameBytes: 1<<20 + 1}, {ResolveTimeout: -1}} {
		if _, err := NewSockets(ioOwner, cfg); err == nil {
			t.Fatal(cfg)
		}
	}
	backends(t, func(t *testing.T, backend string) {
		entered := make(chan struct{}, 1)
		p, _ := socketsOwner(t, SocketsConfig{MaxSockets: 1, MaxWorkers: 1, AllowConnect: func(ctx context.Context, _ SocketAddress) error {
			entered <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		}})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		api := socketsAPI{t, in}
		socket := api.ok(false, "create", component.Value{})
		api.fail(true, "create", "out-of-memory", component.Value{})
		fn, _ := in.Function("types", "[method]tcp-socket.connect")
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			_, err := fn.Call(ctx, []component.Value{socket, loopAddress(0, 80)}, component.CallOptions{})
			done <- err
		}()
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("connect policy did not start")
		}
		if err := p.begin(); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("worker admission", err)
		}
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("connect cancellation stranded")
		}
		deadline := time.Now().Add(time.Second)
		for {
			p.mu.Lock()
			active := p.active
			p.mu.Unlock()
			if active == 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("worker reservation leaked")
			}
			time.Sleep(time.Millisecond)
		}
		api.drop(socket)
		p.Close()
		p.Close()
		if err := p.WaitClosed(nil); err != nil {
			t.Fatal(err)
		}
		if _, err := fn.Call(nil, []component.Value{socket, loopAddress(0, 80)}, component.CallOptions{}); err == nil {
			t.Fatal("closed socket accepted")
		}
	})
	ioOwner.Close()
	if _, err := NewSockets(ioOwner, SocketsConfig{}); err == nil {
		t.Fatal("closed IO")
	}
}
func TestSocketsFailedConnectAndImplicitListen(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _ := socketsOwner(t, SocketsConfig{AllowConnect: socketsAllow, AllowBind: socketsAllow})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		api := socketsAPI{t, in}
		server, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		remote := netAddress(SocketAddress{Address: server.Addr().(*net.TCPAddr).AddrPort()})
		server.Close()
		socket := api.ok(false, "create", component.Value{})
		api.fail(false, "connect", "connection-refused", socket, remote)
		api.fail(false, "get-local-address", "invalid-state", socket)
		api.fail(false, "set-hop-limit", "invalid-state", socket, component.Value{Bits: 64})
		api.drop(socket)
		listener := api.ok(false, "create", component.Value{})
		stream := api.ok(false, "listen", listener)
		if netParse(api.ok(false, "get-local-address", listener)).Address.Port() == 0 {
			t.Fatal("implicit bind")
		}
		api.fail(false, "listen", "invalid-state", listener)
		api.dropAsync(stream)
		api.drop(listener)
		v6 := api.ok(false, "create", component.Value{Bits: 1})
		v6stream := api.ok(false, "listen", v6)
		if netParse(api.ok(false, "get-local-address", v6)).Address.Port() == 0 {
			t.Fatal("IPv6 implicit bind")
		}
		api.dropAsync(v6stream)
		api.drop(v6)
		denied, _ := socketsOwner(t, SocketsConfig{})
		deniedInstance := instantiate(t, compileFixture(t, backend, "sockets-api"), denied.Imports())
		deniedAPI := socketsAPI{t, deniedInstance}
		unbound := deniedAPI.ok(false, "create", component.Value{})
		deniedAPI.fail(false, "listen", "access-denied", unbound)
		deniedAPI.drop(unbound)
		panicOwner, _ := socketsOwner(t, SocketsConfig{AllowBind: func(context.Context, SocketAddress) error { panic("policy") }})
		panicInstance := instantiate(t, compileFixture(t, backend, "sockets-api"), panicOwner.Imports())
		panicAPI := socketsAPI{t, panicInstance}
		unbound = panicAPI.ok(true, "create", component.Value{})
		panicAPI.fail(true, "bind", "access-denied", unbound, loopAddress(0, 0))
		panicAPI.drop(unbound)
	})
}
func TestSocketsSuspendedAcceptAndReceiveShutdown(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, kind := range []string{"accept", "udp"} {
			for _, closeProvider := range []bool{false, true} {
				p, _ := socketsOwner(t, SocketsConfig{AllowBind: socketsAllow, AllowReceive: socketsAllow})
				in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
				api := socketsAPI{t, in}
				udp := kind == "udp"
				socket := api.ok(udp, "create", component.Value{})
				api.ok(udp, "bind", socket, loopAddress(0, 0))
				var stream component.Value
				if !udp {
					stream = api.ok(false, "listen", socket)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				done := make(chan error, 1)
				go func() {
					if udp {
						f, _ := in.Function("types", "[method]udp-socket.receive")
						_, err := f.Call(ctx, []component.Value{socket}, component.CallOptions{})
						done <- err
					} else {
						_, err := in.ReadAsync(ctx, stream.Resource, 1, component.CallOptions{})
						done <- err
					}
				}()
				deadline := time.Now().Add(time.Second)
				for {
					p.mu.Lock()
					active := p.active
					p.mu.Unlock()
					if active > 0 {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("read did not suspend")
					}
					time.Sleep(time.Millisecond)
				}
				if closeProvider {
					p.Close()
				} else {
					in.Close()
				}
				select {
				case err := <-done:
					if err == nil && !udp {
						t.Fatal("accept succeeded after shutdown")
					}
				case <-time.After(3 * time.Second):
					t.Fatal("shutdown stranded read")
				}
				in.Close()
				if err := in.WaitClosed(ctx); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}
func TestSocketsLookupCancellationAndPublicationFailure(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		started := make(chan struct{}, 1)
		p, _ := socketsOwner(t, SocketsConfig{AllowNameLookup: true, Resolve: func(ctx context.Context, _ string) ([]netip.Addr, error) {
			started <- struct{}{}
			<-ctx.Done()
			return nil, ctx.Err()
		}})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		f, _ := in.Function("lookup", "resolve-addresses")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		done := make(chan error, 1)
		go func() {
			_, err := f.Call(ctx, []component.Value{{Text: "pending.example"}}, component.CallOptions{})
			done <- err
		}()
		<-started
		p.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("resolver shutdown stranded")
		}
		if err := p.WaitClosed(ctx); err != nil {
			t.Fatal(err)
		}
		owner, _ := socketsOwner(t, SocketsConfig{})
		cfg := component.DefaultConfig()
		cfg.Core.Backend = backend
		cfg.MaxResources = 1
		e, err := component.NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		data, err := os.ReadFile("testdata/sockets-api.wasm")
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.Compile(nil, data)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		other := instantiate(t, m, owner.Imports())
		api := socketsAPI{t, other}
		socket := api.ok(false, "create", component.Value{})
		api.fail(true, "create", "out-of-memory", component.Value{})
		api.drop(socket)
	})
}

func awaitNetPoll(t *testing.T, p *Sockets) *netStream {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		p.mu.Lock()
		streams := make([]*netStream, 0, len(p.streams))
		for s := range p.streams {
			streams = append(streams, s)
		}
		p.mu.Unlock()
		for _, s := range streams {
			s.mu.Lock()
			watching := s.watching
			s.mu.Unlock()
			if watching {
				return s
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("socket never suspended for readiness")
		}
		time.Sleep(time.Millisecond)
	}
}
func TestSocketsTCPReadinessAndInputCancellation(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, cancelInstance := range []bool{false, true} {
			listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			p, _ := socketsOwner(t, SocketsConfig{AllowConnect: socketsAllow, MaxWorkers: 1})
			in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
			api := socketsAPI{t, in}
			socket := api.ok(false, "create", component.Value{})
			api.ok(false, "connect", socket, netAddress(SocketAddress{Address: listener.Addr().(*net.TCPAddr).AddrPort()}))
			peer, err := listener.AcceptTCP()
			if err != nil {
				t.Fatal(err)
			}
			defer peer.Close()
			peer.SetDeadline(time.Now().Add(3 * time.Second))
			pair := api.call(false, "receive", socket)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			type outcome struct {
				r   component.ReadResult
				err error
			}
			done := make(chan outcome, 1)
			go func() {
				r, err := in.ReadAsync(ctx, pair.Elements[0].Resource, 3, component.CallOptions{})
				done <- outcome{r, err}
			}()
			stream := awaitNetPoll(t, p)
			ready, changed := stream.Readiness()
			again, second := stream.Readiness()
			if ready || again || changed != second {
				t.Fatal("incoherent readiness")
			}
			if cancelInstance {
				in.Close()
			} else {
				if _, err := peer.Write([]byte("ab")); err != nil {
					t.Fatal(err)
				}
				peer.CloseWrite()
			}
			select {
			case v := <-done:
				if cancelInstance {
					if v.err == nil {
						t.Fatal(v)
					}
				} else if v.err != nil || v.r.Bytes != "ab" {
					t.Fatal(v)
				}
			case <-ctx.Done():
				t.Fatal("pending read", ctx.Err())
			}
			if cancelInstance {
				if err := in.WaitClosed(ctx); err != nil {
					t.Fatal(err)
				}
				continue
			}
			api.drop(socket)
			if ready, changed := stream.Readiness(); !ready {
				select {
				case <-changed:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				if ready, _ := stream.Readiness(); !ready {
					t.Fatal("FIN did not become ready")
				}
			}
			r, err := in.ReadAsync(ctx, pair.Elements[0].Resource, 3, component.CallOptions{})
			if err != nil || !r.Closed {
				t.Fatal(r, err)
			}
			api.dropAsync(pair.Elements[0])
			api.completion(pair.Elements[1], "")
			if ready, _ := stream.Readiness(); !ready {
				t.Fatal("closed source not ready")
			}
			if err := stream.Close(); err != nil {
				t.Fatal("repeated stream close", err)
			}
		}
	})
}
func TestSocketsTCPOutputBackpressureAndFIN(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()
		p, process := socketsOwner(t, SocketsConfig{AllowConnect: socketsAllow, MaxWorkers: 1})
		process.config.MaxWriteBytes = 64 << 10
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		api := socketsAPI{t, in}
		socket := api.ok(false, "create", component.Value{})
		api.ok(false, "set-send-buffer-size", socket, component.Value{Bits: 1024})
		api.ok(false, "connect", socket, netAddress(SocketAddress{Address: listener.Addr().(*net.TCPAddr).AddrPort()}))
		peer, err := listener.AcceptTCP()
		if err != nil {
			t.Fatal(err)
		}
		defer peer.Close()
		peer.SetDeadline(time.Now().Add(5 * time.Second))
		f, _ := in.Function("types", "[method]tcp-socket.send")
		_, typ := f.Info().Parameter(1)
		payload := string(make([]byte, 512<<10))
		pending := payload
		source, err := in.NewSource(nil, typ, component.AsyncSource{Read: func(_ *component.Caller, limit uint32) (*component.ReadCompletion, error) {
			n := min(len(pending), int(limit))
			data := pending[:n]
			pending = pending[n:]
			closed := pending == ""
			r := component.NewReadCompletion()
			r.Resolve(func(*component.Caller) (component.ReadResult, error) {
				return component.ReadResult{Bytes: data, Closed: closed}, nil
			})
			return r, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var result component.CallResult
		done := make(chan error, 1)
		go func() {
			var err error
			result, err = f.Call(ctx, []component.Value{socket, source}, component.CallOptions{})
			done <- err
		}()
		awaitNetPoll(t, p)
		data, err := io.ReadAll(peer)
		if err != nil || string(data) != payload {
			t.Fatal("partial TCP output", len(data), err)
		}
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		api.completion(*result.Value, "")
		api.drop(socket)
	})
}

func TestSocketsLateResultsAndPhysicalClose(t *testing.T) {
	p, _ := socketsOwner(t, SocketsConfig{MaxSockets: 2})
	first, err := p.create(0, true)
	if err != nil {
		t.Fatal(err)
	}
	batch := &netAcceptBatch{}
	batch.close()
	batch.add(first)
	batch.close()
	// A descriptor obtained after shutdown is still registered for physical
	// cleanup while its admitted worker keeps the owner join open.
	if err := p.begin(); err != nil {
		t.Fatal(err)
	}
	if err := p.reserve(); err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	p.Close()
	waitCtx, waitCancel := context.WithTimeout(context.Background(), time.Millisecond)
	if err := p.WaitClosed(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("owner did not wait for admitted work", err)
	}
	waitCancel()
	late := p.register(fd, 0, true, 0)
	in := instantiate(t, compileFixture(t, "interpreter", "sockets-api"), p.Imports())
	_, err = in.WithCaller(nil, component.CallOptions{}, func(c *component.Caller) error {
		_, err := p.publish(c, late)
		if !errors.Is(err, os.ErrClosed) {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	late.release()
	p.end()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE); !errors.Is(err, unix.EBADF) {
		t.Fatal("late FD retained", err)
	}
	if err := p.begin(); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
	if err := p.reserve(); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
}
func TestSocketsAcceptPublicationRollbackAndReadinessAdmission(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p, _ := socketsOwner(t, SocketsConfig{AllowBind: socketsAllow})
		cfg := component.DefaultConfig()
		cfg.Core.Backend = backend
		cfg.MaxResources = 3
		e, err := component.NewEngine(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		data, _ := os.ReadFile("testdata/sockets-api.wasm")
		m, err := e.Compile(nil, data)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		in := instantiate(t, m, p.Imports())
		api := socketsAPI{t, in}
		listener := api.ok(false, "create", component.Value{})
		api.ok(false, "bind", listener, loopAddress(0, 0))
		stream := api.ok(false, "listen", listener)
		address := netParse(api.ok(false, "get-local-address", listener))
		client, err := net.DialTCP("tcp4", nil, net.TCPAddrFromAddrPort(address.Address))
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()
		if _, err := in.ReadAsync(nil, stream.Resource, 1, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("accepted socket publication quota", err)
		}
		api.dropAsync(stream)
		api.drop(listener)
		deadline := time.Now().Add(time.Second)
		for {
			p.mu.Lock()
			objects := len(p.objects)
			p.mu.Unlock()
			if objects == 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("unpublished accepted FD retained", objects)
			}
			time.Sleep(time.Millisecond)
		}
		peer, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		defer peer.Close()
		owner, _ := socketsOwner(t, SocketsConfig{AllowConnect: socketsAllow, MaxWorkers: 1})
		other := instantiate(t, compileFixture(t, backend, "sockets-api"), owner.Imports())
		api = socketsAPI{t, other}
		socket := api.ok(false, "create", component.Value{})
		api.ok(false, "connect", socket, netAddress(SocketAddress{Address: peer.Addr().(*net.TCPAddr).AddrPort()}))
		connection, err := peer.AcceptTCP()
		if err != nil {
			t.Fatal(err)
		}
		defer connection.Close()
		pair := api.call(false, "receive", socket)
		owner.mu.Lock()
		var source *netStream
		for s := range owner.streams {
			source = s
		}
		owner.mu.Unlock()
		if err := owner.begin(); err != nil {
			t.Fatal(err)
		}
		ready, _ := source.Readiness()
		owner.end()
		if !ready {
			t.Fatal("failed readiness admission did not finish")
		}
		r, err := other.ReadAsync(nil, pair.Elements[0].Resource, 1, component.CallOptions{})
		if err != nil || !r.Closed {
			t.Fatal(r, err)
		}
		api.dropAsync(pair.Elements[0])
		api.completion(pair.Elements[1], "out-of-memory")
		api.drop(socket)
	})
}
func TestSocketsLookupNormalizationAndCancellation(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, item := range []struct {
			config  SocketsConfig
			name    string
			failure bool
		}{
			{SocketsConfig{AllowNameLookup: true}, "localhost", false},
			{SocketsConfig{AllowNameLookup: true}, "%bad.example", true},
			{SocketsConfig{AllowNameLookup: true, MaxNameBytes: 8}, "é.é", true},
		} {
			p, _ := socketsOwner(t, item.config)
			in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
			f, _ := in.Function("lookup", "resolve-addresses")
			v, err := f.Call(nil, []component.Value{{Text: item.name}}, component.CallOptions{})
			if err != nil || v.Value.Case == 1 != item.failure {
				t.Fatal(item.name, v, err)
			}
		}
		started := make(chan struct{}, 1)
		p, _ := socketsOwner(t, SocketsConfig{AllowNameLookup: true, Resolve: func(ctx context.Context, _ string) ([]netip.Addr, error) {
			started <- struct{}{}
			<-ctx.Done()
			return nil, ctx.Err()
		}})
		in := instantiate(t, compileFixture(t, backend, "sockets-api"), p.Imports())
		f, _ := in.Function("lookup", "resolve-addresses")
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			_, err := f.Call(ctx, []component.Value{{Text: "cancel.example"}}, component.CallOptions{})
			done <- err
		}()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("lookup start")
		}
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("lookup cancellation")
		}
	})
}

func TestSocketsErrorVariants(t *testing.T) {
	for _, item := range []struct {
		err  error
		code uint32
	}{
		{unix.EPIPE, 10}, {unix.ECONNREFUSED, 9}, {unix.ECONNRESET, 11}, {unix.ENOTCONN, 5}, {unix.EAFNOSUPPORT, 2}, {context.DeadlineExceeded, 4}, {os.ErrClosed, 5}, {wasm.ErrLimit, 3}, {netError("unrecognized"), 14}, {errors.New("external failure"), 14},
	} {
		result := netResult(nil, fmt.Errorf("operation: %w", item.err))
		if result.Case != 1 || result.Elements[0].Case != item.code {
			t.Fatal(item.err, result)
		}
		if item.code == 14 && (len(result.Elements[0].Elements) != 1 || result.Elements[0].Elements[0].Case != 0) {
			t.Fatal("other payload", result)
		}
	}
	if netCompletion(io.EOF) != 0 || netCompletion(nil) != 0 {
		t.Fatal("EOF is successful completion")
	}
}

func TestSocketsFixtureProvenance(t *testing.T) {
	check := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("changed producer input", path)
		}
	}
	var fixtures struct {
		Version   string                       `json:"wasiVersion"`
		WIT       string                       `json:"witArchiveSHA256"`
		Producer  string                       `json:"producer"`
		Generator string                       `json:"generatorSHA256"`
		Files     map[string]map[string]string `json:"sha256"`
	}
	data, err := os.ReadFile("testdata/sockets.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	if fixtures.Version != "0.3.2" || fixtures.WIT != "2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311" || fixtures.Producer != "wasm-tools 1.258.0" || len(fixtures.Files) != 4 {
		t.Fatal("changed socket fixture provenance")
	}
	check("../../../scripts/build_preview3_sockets_fixtures.py", fixtures.Generator)
	for name, files := range fixtures.Files {
		for ext, hash := range files {
			check("testdata/"+name+"."+ext, hash)
		}
	}
	var schema struct {
		Version   string `json:"wasiVersion"`
		WIT       string `json:"witArchiveSHA256"`
		Generator string `json:"generatorSHA256"`
		Schema    string `json:"schemaSHA256"`
	}
	data, err = os.ReadFile("testdata/sockets-schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.Version != fixtures.Version || schema.WIT != fixtures.WIT {
		t.Fatal("changed schema provenance")
	}
	check("../../../scripts/build_preview3_sockets_schema.py", schema.Generator)
	check("sockets_schema.go", schema.Schema)
}
