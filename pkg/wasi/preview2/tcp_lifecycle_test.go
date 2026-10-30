//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

func TestTCPResourceLimitsAndPollRollback(t *testing.T) {
	p, _, ioOwner := tcpOwner(t, TCPConfig{MaxSockets: 1})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	socket := a.new(0)
	limited := a.call("create", "create-tcp-socket", component.Value{})
	if limited.Case != 1 || limited.Elements[0].Bits != uint64(NetworkNewSocketLimit) {
		t.Fatal(limited)
	}
	ioOwner.poller.next = 1 << 32
	f, err := a.in.Function("tcp", "[method]tcp-socket.subscribe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Call(nil, []component.Value{socket}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	p.mu.Lock()
	for _, s := range p.sockets {
		if s.children != 0 {
			t.Fatal("failed subscribe retained child", s.children)
		}
	}
	p.mu.Unlock()
	a.drop(socket)
	p.next = 1 << 32
	limited = a.call("create", "create-tcp-socket", component.Value{})
	if limited.Case != 1 || limited.Elements[0].Bits != uint64(NetworkNewSocketLimit) {
		t.Fatal(limited)
	}
}
func TestTCPStreamPublicationRollback(t *testing.T) {
	p, _, process := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	net := a.call("instance-network", "instance-network")
	server := a.new(0)
	client := a.new(0)
	a.ok("start-bind", server, net, tcpAddress(t, "127.0.0.1:0"))
	a.ok("finish-bind", server)
	a.ok("start-listen", server)
	a.ok("finish-listen", server)
	address := a.ok("local-address", server)
	a.ok("start-connect", client, net, address)
	a.block(client)
	process.config.MaxStreams = 1
	// Output publication fails after input publication; both child leases and
	// the already published owned input must roll back.
	result := a.result("finish-connect", client)
	if result.Case != 1 {
		t.Fatal(result)
	}
	process.mu.Lock()
	remaining := len(process.records)
	process.mu.Unlock()
	id, err := a.in.ResourceRep(nil, p.resource, client.Resource)
	if err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	children := p.sockets[id].children
	p.mu.Unlock()
	if remaining != 0 || children != 0 {
		t.Fatal("partial stream publication retained resources", remaining, children)
	}
	// Accepted-socket publication must also reclaim its descriptor on failure.
	result = a.result("accept", server)
	if result.Case != 1 {
		t.Fatal(result)
	}
	p.mu.Lock()
	sockets := len(p.sockets)
	p.mu.Unlock()
	if sockets != 2 {
		t.Fatal("partial accept retained socket", sockets)
	}
	a.drop(client)
	a.drop(server)
	a.drop(net)
}
func TestTCPAuthorityCancellationAndPanic(t *testing.T) {
	for _, mode := range []string{"denied", "panic", "caller-cancel", "provider-close"} {
		t.Run(mode, func(t *testing.T) {
			entered := make(chan struct{})
			var p *TCP
			allow := func(ctx context.Context, _ SocketAddress) error {
				close(entered)
				switch mode {
				case "denied":
					return os.ErrPermission
				case "panic":
					panic("policy failed")
				default:
					<-ctx.Done()
					return ctx.Err()
				}
			}
			p, _, _ = tcpOwner(t, TCPConfig{AllowBind: allow})
			a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
			net := a.call("instance-network", "instance-network")
			socket := a.new(0)
			bind, err := a.in.Function("tcp", "[method]tcp-socket.start-bind")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				v, err := bind.Call(ctx, []component.Value{socket, net, tcpAddress(t, "127.0.0.1:0")}, component.CallOptions{})
				if err == nil && (v.Value.Case != 1 || v.Value.Elements[0].Bits != uint64(NetworkAccessDenied)) {
					err = errors.New("policy failure was not access-denied")
				}
				done <- err
			}()
			<-entered
			if mode == "caller-cancel" {
				cancel()
			}
			if mode == "provider-close" {
				p.Close()
			}
			select {
			case err := <-done:
				switch mode {
				case "caller-cancel":
					if !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
				case "provider-close":
					if !errors.Is(err, wasm.ErrClosed) {
						t.Fatal(err)
					}
				default:
					if err != nil {
						t.Fatal(err)
					}
				}
			case <-time.After(time.Second):
				t.Fatal("policy callback not cancelled")
			}
			if mode != "provider-close" {
				a.drop(socket)
				a.drop(net)
			}
		})
	}
}
func TestTCPCloseWakesReadsAndPreservesErrorProvenance(t *testing.T) {
	p, _, _ := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	c := a.connect()
	read, err := a.in.Function("streams", "[method]input-stream.blocking-read")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan component.CallResult, 1)
	failed := make(chan error, 1)
	go func() {
		value, err := read.Call(nil, []component.Value{c.si, {Bits: 1}}, component.CallOptions{})
		if err != nil {
			failed <- err
		} else {
			done <- value
		}
	}()
	// Closing is valid before or after the read creates its readiness waiter.
	p.Close()
	select {
	case err := <-failed:
		t.Fatal(err)
	case result := <-done:
		if result.Value.Case != 1 || result.Value.Elements[0].Case != 0 {
			t.Fatal(result)
		}
		cause := result.Value.Elements[0].Elements[0]
		code := a.call("network", "network-error-code", cause)
		if code.Case != 1 || code.Elements[0].Bits != uint64(NetworkInvalidState) {
			t.Fatal(code)
		}
		a.drop(cause)
	case <-time.After(time.Second):
		t.Fatal("TCP close stranded a read")
	}
	a.close(c)
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
}
func TestTCPConstructorCloseAndExpiredCallers(t *testing.T) {
	if _, err := NewTCP(nil, TCPConfig{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	var zero TCP
	zero.Close()
	if len(zero.Imports().Instances) != 0 {
		t.Fatal("zero imports")
	}
	if err := zero.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	p, network, process := tcpOwner(t, TCPConfig{})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	module, err := engine.Compile(nil, fixture(t, "tcp-api"))
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	imports := p.Imports()
	create := imports.Instances[TCPCreateInterface].Functions["create-tcp-socket"]
	var saved *component.Caller
	imports.Instances[TCPCreateInterface].Functions["create-tcp-socket"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
		saved = c
		return create(c, args)
	}
	in, err := module.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	a := tcpAPI{networkAPI{t, in}}
	socket := a.new(0)
	a.drop(socket)
	file, err := newSocketFile(p.ctx, 0, unix.SOCK_STREAM)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := p.publish(saved, file, 0, tcpInitial); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	file.Close()
	file.WaitClosed(nil)
	p.mu.Lock()
	remaining := len(p.sockets)
	p.mu.Unlock()
	if remaining != 0 {
		t.Fatal("expired publication retained socket", remaining)
	}
	// Active creations and detached closes are joined, with caller-controlled wait.
	if err := p.begin(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	p.active.Done()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Close()
			if err := p.WaitClosed(nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if err := p.begin(context.Background()); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	process.Close()
	if _, err := NewTCP(network, TCPConfig{}); err == nil {
		t.Fatal("accepted closed IO")
	}
	network.Close()
	if _, err := NewTCP(network, TCPConfig{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
func TestTCPContractRejection(t *testing.T) {
	p, _, _ := tcpOwner(t, TCPConfig{})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for _, name := range []string{"address-field", "address-width", "stream-order", "keep-count", "shutdown-label", "family-result", "subscribe-result", "create-result"} {
		module, err := engine.Compile(nil, fixture(t, "tcp-contract-"+name))
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

func TestTCPListenerOptionsAndParentOwnerClose(t *testing.T) {
	p, network, _ := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	net := a.call("instance-network", "instance-network")
	server := a.new(1)
	for _, option := range []struct {
		name  string
		value uint64
	}{{"keep-alive-idle-time", 2000000000}, {"hop-limit", 31}, {"receive-buffer-size", 16384}, {"send-buffer-size", 16384}} {
		a.ok("set-"+option.name, server, component.Value{Bits: option.value})
	}
	a.ok("start-bind", server, net, tcpAddress(t, "[::1]:0"))
	a.ok("finish-bind", server)
	a.ok("start-listen", server)
	a.ok("finish-listen", server)
	client := a.new(1)
	a.ok("start-connect", client, net, a.ok("local-address", server))
	a.block(client)
	streams := a.ok("finish-connect", client)
	a.block(server)
	accepted := a.ok("accept", server)
	child := accepted.Elements[0]
	if v := a.ok("keep-alive-idle-time", child); v.Bits != 2000000000 {
		t.Fatal(v)
	}
	if v := a.ok("hop-limit", child); v.Bits != 31 {
		t.Fatal(v)
	}
	for _, name := range []string{"receive-buffer-size", "send-buffer-size"} {
		if v := a.ok(name, child); v.Bits < 16384 {
			t.Fatal(name, v)
		}
	}
	for _, value := range []component.Value{streams.Elements[0], streams.Elements[1], accepted.Elements[1], accepted.Elements[2], client, child, server, net} {
		a.drop(value)
	}
	network.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.WaitClosed(ctx); err != nil {
		t.Fatal("parent network did not close TCP", err)
	}
}
func TestTCPCleanupErrorAndBlockedAcceptCancellation(t *testing.T) {
	p, _, _ := tcpOwnerCleanup(t, TCPConfig{AllowBind: tcpAllowLoopback}, os.ErrClosed)
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	net := a.call("instance-network", "instance-network")
	server := a.new(0)
	a.ok("start-bind", server, net, tcpAddress(t, "127.0.0.1:0"))
	a.ok("finish-bind", server)
	a.ok("start-listen", server)
	a.ok("finish-listen", server)
	poll := a.call("tcp", "[method]tcp-socket.subscribe", server)
	if ready := a.call("poll", "[method]pollable.ready", poll); ready.Bits != 0 {
		t.Fatal(ready)
	}
	f, err := a.in.Function("poll", "[method]pollable.block")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.Call(ctx, []component.Value{poll}, component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	a.drop(poll)
	id, err := a.in.ResourceRep(nil, p.resource, server.Resource)
	if err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	file := p.sockets[id].file
	p.mu.Unlock()
	// Inject an external descriptor close to verify retained cleanup errors.
	if err := file.file.Close(); err != nil {
		t.Fatal(err)
	}
	a.drop(server)
	a.drop(net)
	p.Close()
	if err := p.WaitClosed(nil); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
}

func TestTCPDefaultDeniesNetworkUse(t *testing.T) {
	p, _, _ := tcpOwner(t, TCPConfig{})
	a := tcpAPI{networkAPI{t, tcpInstance(t, p, "interpreter", "tcp-api")}}
	net := a.call("instance-network", "instance-network")
	socket := a.new(0)
	a.fail("start-bind", NetworkAccessDenied, socket, net, tcpAddress(t, "127.0.0.1:0"))
	a.fail("start-connect", NetworkAccessDenied, socket, net, tcpAddress(t, "127.0.0.1:80"))
	a.drop(socket)
	a.drop(net)
}
