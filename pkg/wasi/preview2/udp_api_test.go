//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type udpAPI struct{ networkAPI }

func (a udpAPI) new(family uint64) component.Value {
	a.t.Helper()
	v := a.call("create", "create-udp-socket", component.Value{Bits: family})
	if v.Case != 0 {
		a.t.Fatal(v)
	}
	return v.Elements[0]
}
func (a udpAPI) result(kind uint32, name string, self component.Value, args ...component.Value) component.Value {
	a.t.Helper()
	prefix := []string{"udp-socket", "incoming-datagram-stream", "outgoing-datagram-stream"}[kind]
	return a.call("udp", "[method]"+prefix+"."+name, append([]component.Value{self}, args...)...)
}
func (a udpAPI) ok(kind uint32, name string, self component.Value, args ...component.Value) component.Value {
	a.t.Helper()
	v := a.result(kind, name, self, args...)
	if v.Case != 0 {
		a.t.Fatal(name, v)
	}
	if len(v.Elements) == 0 {
		return component.Value{}
	}
	return v.Elements[0]
}
func (a udpAPI) fail(kind uint32, name string, code NetworkErrorCode, self component.Value, args ...component.Value) {
	a.t.Helper()
	v := a.result(kind, name, self, args...)
	if v.Case != 1 || len(v.Elements) != 1 || v.Elements[0].Bits != uint64(code) {
		a.t.Fatal(name, v, code)
	}
}
func (a udpAPI) trap(kind uint32, name string, self component.Value, args ...component.Value) {
	a.t.Helper()
	prefix := []string{"udp-socket", "incoming-datagram-stream", "outgoing-datagram-stream"}[kind]
	f, err := a.in.Function("udp", "[method]"+prefix+"."+name)
	if err != nil {
		a.t.Fatal(err)
	}
	_, err = f.Call(nil, append([]component.Value{self}, args...), component.CallOptions{})
	if !errors.Is(err, wasm.ErrTrap) {
		a.t.Fatal(name, err)
	}
}
func udpSome(v component.Value) component.Value {
	return component.Value{Case: 1, Elements: []component.Value{v}}
}
func udpPacket(text string, peer component.Value) component.Value {
	return component.Value{Elements: []component.Value{{Text: text}, peer}}
}
func udpPackets(packets ...component.Value) component.Value {
	return component.Value{Elements: packets}
}

type udpConnection struct{ network, client, server, ci, co, si, so, clientAddress, serverAddress component.Value }

func (a udpAPI) connect(connected bool) udpConnection {
	a.t.Helper()
	c := udpConnection{network: a.call("instance-network", "instance-network"), client: a.new(0), server: a.new(0)}
	for _, socket := range []component.Value{c.client, c.server} {
		a.ok(0, "start-bind", socket, c.network, tcpAddress(a.t, "127.0.0.1:0"))
		a.ok(0, "finish-bind", socket)
	}
	c.clientAddress = a.ok(0, "local-address", c.client)
	c.serverAddress = a.ok(0, "local-address", c.server)
	cp, sp := component.Value{}, component.Value{}
	if connected {
		cp, sp = udpSome(c.serverAddress), udpSome(c.clientAddress)
	}
	streams := a.ok(0, "stream", c.client, cp)
	c.ci, c.co = streams.Elements[0], streams.Elements[1]
	streams = a.ok(0, "stream", c.server, sp)
	c.si, c.so = streams.Elements[0], streams.Elements[1]
	return c
}
func (a udpAPI) close(c udpConnection) {
	a.t.Helper()
	for _, v := range []component.Value{c.ci, c.co, c.si, c.so, c.client, c.server, c.network} {
		a.drop(v)
	}
}
func (a udpAPI) receive(stream component.Value, count uint64) component.Value {
	a.t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		v := a.ok(1, "receive", stream, component.Value{Bits: count})
		if len(v.Elements) > 0 {
			return v
		}
		if time.Now().After(deadline) {
			a.t.Fatal("receive never became ready")
		}
	}
}
func TestUDPPermitsPacketsAndReconfiguration(t *testing.T) {
	p, _, _ := udpOwner(t, UDPConfig{AllowBind: tcpAllowLoopback, AllowSend: tcpAllowLoopback, AllowReceive: tcpAllowLoopback})
	a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
	c := a.connect(false)
	a.trap(0, "stream", c.client, component.Value{})
	a.trap(0, "stream", c.client, udpSome(tcpAddress(t, "127.0.0.1:0")))
	packet := udpPacket("exact\x00bytes", udpSome(c.serverAddress))
	list := udpPackets(packet)
	a.trap(2, "send", c.co, list)
	if v := a.ok(2, "check-send", c.co); v.Bits == 0 {
		t.Fatal(v)
	}
	if v := a.ok(2, "send", c.co, component.Value{}); v.Bits != 0 {
		t.Fatal(v)
	}
	if v := a.ok(2, "send", c.co, list); v.Bits != 1 {
		t.Fatal(v)
	}
	v := a.receive(c.si, 1)
	text, err := v.Elements[0].Elements[0].ByteString()
	if text != "exact\x00bytes" || err != nil {
		t.Fatal(v, err)
	}
	if v := a.ok(1, "receive", c.si, component.Value{}); len(v.Elements) != 0 {
		t.Fatal(v)
	}
	a.trap(2, "send", c.co, list)
	a.ok(2, "check-send", c.co)
	a.fail(2, "send", NetworkInvalidArgument, c.co, udpPackets(udpPacket("missing-peer", component.Value{})))
	a.ok(2, "check-send", c.co)
	a.fail(2, "send", NetworkDatagramTooLarge, c.co, udpPackets(udpPacket(string(make([]byte, 65536)), udpSome(c.serverAddress))))
	a.ok(2, "check-send", c.co)
	if v := a.ok(2, "send", c.co, udpPackets(packet, udpPacket("bad", component.Value{}))); v.Bits != 1 {
		t.Fatal(v)
	}
	a.receive(c.si, 2)
	a.drop(c.ci)
	a.drop(c.co)
	streams := a.ok(0, "stream", c.client, udpSome(c.serverAddress))
	c.ci, c.co = streams.Elements[0], streams.Elements[1]
	peer := a.ok(0, "remote-address", c.client)
	address, err := parseSocketAddress(peer)
	want, _ := parseSocketAddress(c.serverAddress)
	if err != nil || address != want {
		t.Fatal(address, err)
	}
	a.ok(2, "check-send", c.co)
	a.fail(2, "send", NetworkInvalidArgument, c.co, udpPackets(udpPacket("wrong-peer", udpSome(c.clientAddress))))
	a.ok(2, "check-send", c.co)
	if v := a.ok(2, "send", c.co, udpPackets(udpPacket("", component.Value{}))); v.Bits != 1 {
		t.Fatal(v)
	}
	if v := a.receive(c.si, 1); len(v.Elements) != 1 {
		t.Fatal(v)
	}
	a.drop(c.ci)
	a.drop(c.co)
	streams = a.ok(0, "stream", c.client, udpSome(c.clientAddress))
	c.ci, c.co = streams.Elements[0], streams.Elements[1]
	actualPeer, err := parseSocketAddress(a.ok(0, "remote-address", c.client))
	localPeer, _ := parseSocketAddress(c.clientAddress)
	if err != nil || actualPeer != localPeer {
		t.Fatal(actualPeer, err)
	}
	a.drop(c.ci)
	a.drop(c.co)
	streams = a.ok(0, "stream", c.client, component.Value{})
	c.ci, c.co = streams.Elements[0], streams.Elements[1]
	a.fail(0, "remote-address", NetworkInvalidState, c.client)
	a.close(c)
}
func TestUDPStateOptionsAndContractRejection(t *testing.T) {
	p, _, _ := udpOwner(t, UDPConfig{AllowBind: tcpAllowLoopback})
	a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
	net := a.call("instance-network", "instance-network")
	socket := a.new(0)
	a.fail(0, "finish-bind", NetworkNotInProgress, socket)
	a.fail(0, "local-address", NetworkInvalidState, socket)
	a.fail(0, "remote-address", NetworkInvalidState, socket)
	a.fail(0, "stream", NetworkInvalidState, socket, component.Value{})
	a.fail(0, "start-bind", NetworkInvalidArgument, socket, net, tcpAddress(t, "[::1]:0"))
	a.fail(0, "start-bind", NetworkAccessDenied, socket, net, tcpAddress(t, "0.0.0.0:0"))
	for _, name := range []string{"unicast-hop-limit", "receive-buffer-size", "send-buffer-size"} {
		a.fail(0, "set-"+name, NetworkInvalidArgument, socket, component.Value{})
		a.ok(0, name, socket)
	}
	a.ok(0, "set-send-buffer-size", socket, component.Value{Bits: 1<<64 - 1})
	a.ok(0, "start-bind", socket, net, tcpAddress(t, "127.0.0.1:0"))
	a.fail(0, "start-bind", NetworkConcurrencyConflict, socket, net, tcpAddress(t, "127.0.0.1:0"))
	a.fail(0, "stream", NetworkConcurrencyConflict, socket, component.Value{})
	a.ok(0, "finish-bind", socket)
	a.fail(0, "start-bind", NetworkInvalidState, socket, net, tcpAddress(t, "127.0.0.1:0"))
	a.fail(0, "stream", NetworkInvalidArgument, socket, udpSome(tcpAddress(t, "127.0.0.1:0")))
	a.fail(0, "stream", NetworkAccessDenied, socket, udpSome(tcpAddress(t, "127.0.0.1:1234")))
	a.drop(socket)
	a.drop(net)
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for _, name := range []string{"address-field", "stream-order", "receive-width", "send-width", "outgoing-address", "incoming-width", "subscribe-result", "create-result"} {
		module, err := engine.Compile(nil, fixture(t, "udp-contract-"+name))
		if err != nil {
			t.Fatal(name, err)
		}
		in, err := module.Instantiate(nil, p.Imports(), component.CallOptions{})
		if in != nil {
			in.Close()
		}
		module.Close()
		if !errors.Is(err, vm.ErrType) {
			t.Fatal(name, err)
		}
	}
}
func TestUDPQuotasAndOrphanedPollables(t *testing.T) {
	t.Run("socket-limit", func(t *testing.T) {
		p, _, _ := udpOwner(t, UDPConfig{MaxSockets: 1})
		a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
		socket := a.new(0)
		v := a.call("create", "create-udp-socket", component.Value{})
		if v.Case != 1 || v.Elements[0].Bits != uint64(NetworkNewSocketLimit) {
			t.Fatal(v)
		}
		a.drop(socket)
		p.mu.Lock()
		p.next = 1 << 32
		p.mu.Unlock()
		v = a.call("create", "create-udp-socket", component.Value{})
		if v.Case != 1 {
			t.Fatal(v)
		}
	})
	t.Run("partial-stream-publication", func(t *testing.T) {
		p, _, _ := udpOwner(t, UDPConfig{MaxStreams: 1, AllowBind: tcpAllowLoopback})
		a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
		net := a.call("instance-network", "instance-network")
		socket := a.new(0)
		a.ok(0, "start-bind", socket, net, tcpAddress(t, "127.0.0.1:0"))
		a.ok(0, "finish-bind", socket)
		a.fail(0, "stream", NetworkNewSocketLimit, socket, component.Value{})
		p.mu.Lock()
		remaining := p.streams
		p.mu.Unlock()
		if remaining != 0 {
			t.Fatal("partial publication leaked", remaining)
		}
		a.drop(socket)
		a.drop(net)
	})
	t.Run("pollable-parents", func(t *testing.T) {
		p, _, _ := udpOwner(t, UDPConfig{AllowBind: tcpAllowLoopback})
		a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
		c := a.connect(false)
		poll := a.result(1, "subscribe", c.si)
		if v := a.call("poll", "[method]pollable.ready", poll); v.Bits != 0 {
			t.Fatal(v)
		}
		if _, err := a.in.DropResource(nil, c.si.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		if _, err := a.in.DropResource(nil, c.server.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal(err)
		}
		a.drop(poll)
		a.drop(c.so)
		a.drop(c.ci)
		a.drop(c.co)
		a.drop(c.client)
		a.drop(c.network)
		p.mu.Lock()
		remaining := len(p.entries)
		p.mu.Unlock()
		if remaining != 0 {
			t.Fatal(remaining)
		}
	})
}
func TestUDPDefaultDeniedAndReceiveFilter(t *testing.T) {
	p, _, _ := udpOwner(t, UDPConfig{})
	a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
	net := a.call("instance-network", "instance-network")
	socket := a.new(0)
	a.fail(0, "start-bind", NetworkAccessDenied, socket, net, tcpAddress(t, "127.0.0.1:0"))
	a.drop(socket)
	a.drop(net)
	q, _, _ := udpOwner(t, UDPConfig{AllowBind: tcpAllowLoopback, AllowSend: tcpAllowLoopback})
	b := udpAPI{networkAPI{t, udpInstance(t, q, "interpreter", "udp-api")}}
	c := b.connect(true)
	b.ok(2, "check-send", c.co)
	b.ok(2, "send", c.co, udpPackets(udpPacket("denied", component.Value{})))
	if v := b.ok(1, "receive", c.si, component.Value{Bits: 1<<64 - 1}); len(v.Elements) != 0 {
		t.Fatal(v)
	}
	b.close(c)
}
func TestUDPBlockedPollCancellationAndOwnerClose(t *testing.T) {
	p, network, _ := udpOwner(t, UDPConfig{AllowBind: tcpAllowLoopback})
	a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
	c := a.connect(false)
	poll := a.result(1, "subscribe", c.si)
	f, err := a.in.Function("poll", "[method]pollable.block")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.Call(ctx, []component.Value{poll}, component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if v := a.call("poll", "[method]pollable.ready", poll); v.Bits != 0 {
		t.Fatal(v)
	}
	network.Close()
	deadline, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := p.WaitClosed(deadline); err != nil {
		t.Fatal(err)
	}
	if v := a.call("poll", "[method]pollable.ready", poll); v.Bits != 1 {
		t.Fatal(v)
	}
	a.drop(poll)
	f, err = a.in.Function("udp", "[method]incoming-datagram-stream.receive")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Call(nil, []component.Value{c.si, {Bits: 1}}, component.CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
