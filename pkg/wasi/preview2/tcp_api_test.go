//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type tcpAPI struct{ networkAPI }

func (a tcpAPI) new(family uint64) component.Value {
	a.t.Helper()
	v := a.call("create", "create-tcp-socket", component.Value{Bits: family})
	if v.Case != 0 {
		a.t.Fatal(v)
	}
	return v.Elements[0]
}
func (a tcpAPI) result(name string, self component.Value, args ...component.Value) component.Value {
	a.t.Helper()
	return a.call("tcp", "[method]tcp-socket."+name, append([]component.Value{self}, args...)...)
}
func (a tcpAPI) ok(name string, self component.Value, args ...component.Value) component.Value {
	a.t.Helper()
	v := a.result(name, self, args...)
	if v.Case != 0 {
		a.t.Fatal(name, v)
	}
	if len(v.Elements) == 0 {
		return component.Value{}
	}
	return v.Elements[0]
}
func (a tcpAPI) fail(name string, code NetworkErrorCode, self component.Value, args ...component.Value) {
	a.t.Helper()
	v := a.result(name, self, args...)
	if v.Case != 1 || v.Elements[0].Bits != uint64(code) {
		a.t.Fatal(name, v, code)
	}
}
func (a tcpAPI) block(socket component.Value) {
	a.t.Helper()
	poll := a.call("tcp", "[method]tcp-socket.subscribe", socket)
	f, err := a.in.Function("poll", "[method]pollable.block")
	if err != nil {
		a.t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := f.Call(ctx, []component.Value{poll}, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
	a.drop(poll)
}
func tcpAddress(t *testing.T, address string) component.Value {
	t.Helper()
	v, err := socketAddressValue(SocketAddress{Address: netip.MustParseAddrPort(address)})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

type tcpConnection struct{ network, server, client, child, ci, co, si, so component.Value }

func (a tcpAPI) connect() tcpConnection {
	a.t.Helper()
	var c tcpConnection
	c.network = a.call("instance-network", "instance-network")
	c.server = a.new(0)
	a.ok("start-bind", c.server, c.network, tcpAddress(a.t, "127.0.0.1:0"))
	a.ok("finish-bind", c.server)
	a.ok("start-listen", c.server)
	a.ok("finish-listen", c.server)
	address := a.ok("local-address", c.server)
	c.client = a.new(0)
	a.ok("start-connect", c.client, c.network, address)
	for {
		result := a.result("finish-connect", c.client)
		if result.Case == 0 {
			c.ci, c.co = result.Elements[0].Elements[0], result.Elements[0].Elements[1]
			break
		}
		if result.Elements[0].Bits != uint64(NetworkWouldBlock) {
			a.t.Fatal(result)
		}
		a.block(c.client)
	}
	for {
		result := a.result("accept", c.server)
		if result.Case == 0 {
			values := result.Elements[0].Elements
			c.child, c.si, c.so = values[0], values[1], values[2]
			break
		}
		if result.Elements[0].Bits != uint64(NetworkWouldBlock) {
			a.t.Fatal(result)
		}
		a.block(c.server)
	}
	return c
}
func (a tcpAPI) close(c tcpConnection) {
	a.t.Helper()
	for _, v := range []component.Value{c.ci, c.co, c.si, c.so, c.client, c.child, c.server, c.network} {
		a.drop(v)
	}
}
func TestTCPStatesOptionsAndPermissions(t *testing.T) {
	p, _, _ := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	network := a.call("instance-network", "instance-network")
	socket := a.new(0)
	for _, name := range []string{"finish-bind", "finish-connect", "finish-listen"} {
		a.fail(name, NetworkNotInProgress, socket)
	}
	for _, name := range []string{"start-listen", "accept", "local-address", "remote-address"} {
		a.fail(name, NetworkInvalidState, socket)
	}
	a.fail("shutdown", NetworkInvalidState, socket, component.Value{Bits: 2})
	a.fail("start-bind", NetworkAccessDenied, socket, network, tcpAddress(t, "0.0.0.0:0"))
	a.fail("start-connect", NetworkInvalidArgument, socket, network, tcpAddress(t, "127.0.0.1:0"))
	a.fail("start-connect", NetworkInvalidArgument, socket, network, tcpAddress(t, "0.0.0.0:80"))
	a.fail("start-connect", NetworkInvalidArgument, socket, network, tcpAddress(t, "[::1]:80"))
	for _, address := range []string{"224.0.0.1:80", "255.255.255.255:80"} {
		a.fail("start-bind", NetworkInvalidArgument, socket, network, tcpAddress(t, address))
	}
	for _, name := range []string{"set-listen-backlog-size", "set-keep-alive-idle-time", "set-keep-alive-interval", "set-keep-alive-count", "set-hop-limit", "set-receive-buffer-size", "set-send-buffer-size"} {
		a.fail(name, NetworkInvalidArgument, socket, component.Value{})
	}
	a.ok("set-keep-alive-enabled", socket, component.Value{Bits: 1})
	if v := a.ok("keep-alive-enabled", socket); v.Bits != 1 {
		t.Fatal(v)
	}
	a.ok("set-keep-alive-enabled", socket, component.Value{})
	if v := a.ok("keep-alive-enabled", socket); v.Bits != 0 {
		t.Fatal(v)
	}
	a.ok("set-keep-alive-idle-time", socket, component.Value{Bits: ^uint64(0)})
	if v := a.ok("keep-alive-idle-time", socket); v.Bits != 32767*1000000000 {
		t.Fatal(v)
	}
	a.ok("set-keep-alive-interval", socket, component.Value{Bits: 1})
	if v := a.ok("keep-alive-interval", socket); v.Bits != 1000000000 {
		t.Fatal(v)
	}
	a.ok("set-keep-alive-count", socket, component.Value{Bits: 10000})
	if v := a.ok("keep-alive-count", socket); v.Bits != 127 {
		t.Fatal(v)
	}
	a.ok("set-listen-backlog-size", socket, component.Value{Bits: ^uint64(0)})
	a.ok("start-bind", socket, network, tcpAddress(t, "127.0.0.1:0"))
	a.fail("start-bind", NetworkConcurrencyConflict, socket, network, tcpAddress(t, "127.0.0.1:0"))
	a.fail("start-listen", NetworkConcurrencyConflict, socket)
	a.ok("finish-bind", socket)
	a.fail("start-bind", NetworkInvalidState, socket, network, tcpAddress(t, "127.0.0.1:0"))
	a.ok("start-listen", socket)
	a.fail("start-connect", NetworkConcurrencyConflict, socket, network, tcpAddress(t, "127.0.0.1:80"))
	a.ok("finish-listen", socket)
	a.fail("accept", NetworkWouldBlock, socket)
	a.ok("set-listen-backlog-size", socket, component.Value{Bits: 8})
	a.fail("start-listen", NetworkInvalidState, socket)
	a.drop(socket)
	a.drop(network)
}
func TestTCPChildrenStreamsAndShutdown(t *testing.T) {
	p, _, _ := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	backend := "interpreter"
	if native.Available {
		backend = "native"
	}
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, backend, "tcp-api")}}
	c := a.connect()
	payload := []byte("socket data\x00界🙂")
	result := a.call("streams", "[method]output-stream.blocking-write-and-flush", c.co, component.ByteList(payload))
	if result.Case != 0 {
		t.Fatal(result)
	}
	got := []byte{}
	for len(got) < len(payload) {
		value := a.call("streams", "[method]input-stream.blocking-read", c.si, component.Value{Bits: uint64(len(payload) - len(got))})
		if value.Case != 0 {
			t.Fatal(value)
		}
		text, err := value.Elements[0].ByteString()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, text...)
	}
	if string(got) != string(payload) {
		t.Fatal(string(got))
	}
	a.fail("finish-connect", NetworkNotInProgress, c.client)
	a.fail("set-listen-backlog-size", NetworkInvalidState, c.client, component.Value{Bits: 8})
	for _, how := range []uint64{0, 1, 2} {
		a.ok("shutdown", c.client, component.Value{Bits: how})
	}
	poll := a.call("tcp", "[method]tcp-socket.subscribe", c.child)
	if _, err := a.in.DropResource(nil, c.child.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal("parent drop with children", err)
	}
	a.drop(poll)
	a.drop(c.si)
	a.drop(c.so)
	for _, value := range []component.Value{c.ci, c.co, c.client, c.server, c.network} {
		a.drop(value)
	}
	p.mu.Lock()
	remaining := len(p.sockets)
	p.mu.Unlock()
	if remaining != 0 {
		t.Fatal("orphan socket retained", remaining)
	}
}
func TestTCPRefusedConnectionAndClosedState(t *testing.T) {
	p, _, _ := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	network := a.call("instance-network", "instance-network")
	listener := a.new(0)
	a.ok("start-bind", listener, network, tcpAddress(t, "127.0.0.1:0"))
	a.ok("finish-bind", listener)
	address := a.ok("local-address", listener)
	id, err := a.in.ResourceRep(nil, p.resource, listener.Resource)
	if err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	root := p.sockets[id].file
	p.mu.Unlock()
	a.drop(listener)
	if err := root.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	// Close an unused bound port before connecting, avoiding platform-specific
	// treatment of SYN packets addressed to a bound but non-listening socket.
	client := a.new(0)
	start := a.result("start-connect", client, network, address)
	if start.Case == 0 {
		a.block(client)
		a.fail("finish-connect", NetworkConnectionRefused, client)
	} else if start.Elements[0].Bits != uint64(NetworkConnectionRefused) {
		t.Fatal(start)
	}
	for _, name := range []string{"local-address", "remote-address", "keep-alive-enabled"} {
		a.fail(name, NetworkConnectionRefused, client)
	}
	a.fail("start-connect", NetworkInvalidState, client, network, address)
	a.drop(client)
	a.drop(network)
}
