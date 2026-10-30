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
)

func TestUDPConstructorsAndConcurrentShutdown(t *testing.T) {
	var zero UDP
	zero.Close()
	if err := zero.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if _, err := NewUDP(nil, UDPConfig{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if v := zero.Imports(); len(v.Instances) != 0 {
		t.Fatal(v)
	}
	p, network, process := udpOwner(t, UDPConfig{})
	if _, err := NewUDP(network, UDPConfig{MaxDatagrams: 65}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.begin(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := p.begin(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.Close()
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
	if _, err := NewUDP(network, UDPConfig{}); err == nil {
		t.Fatal("accepted closed IO")
	}
	network.Close()
	if _, err := NewUDP(network, UDPConfig{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
func TestUDPPolicyCancellationAndPanics(t *testing.T) {
	for _, action := range []string{"panic", "caller-cancel", "provider-close"} {
		t.Run(action, func(t *testing.T) {
			started := make(chan struct{})
			var once sync.Once
			callback := func(ctx context.Context, _ SocketAddress) error {
				if action == "panic" {
					panic("policy failure")
				}
				once.Do(func() { close(started) })
				<-ctx.Done()
				return ctx.Err()
			}
			p, _, _ := udpOwner(t, UDPConfig{AllowBind: callback})
			a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
			net := a.call("instance-network", "instance-network")
			socket := a.new(0)
			if action == "panic" {
				a.fail(0, "start-bind", NetworkAccessDenied, socket, net, tcpAddress(t, "127.0.0.1:0"))
				a.drop(socket)
				a.drop(net)
				return
			}
			f, err := a.in.Function("udp", "[method]udp-socket.start-bind")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			address := tcpAddress(t, "127.0.0.1:0")
			go func() {
				_, err := f.Call(ctx, []component.Value{socket, net, address}, component.CallOptions{})
				done <- err
			}()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("policy never started")
			}
			want := wasm.ErrClosed
			if action == "caller-cancel" {
				want = context.Canceled
				cancel()
			} else {
				p.Close()
			}
			select {
			case err := <-done:
				if !errors.Is(err, want) {
					t.Fatal(err, want)
				}
			case <-time.After(time.Second):
				t.Fatal("policy did not cancel")
			}
			if action == "caller-cancel" {
				a.drop(socket)
				a.drop(net)
			} else if err := p.WaitClosed(nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUDPPollPublicationRollbackAndCleanupErrors(t *testing.T) {
	p, _, process := udpOwnerCleanup(t, UDPConfig{}, os.ErrClosed)
	a := udpAPI{networkAPI{t, udpInstance(t, p, "interpreter", "udp-api")}}
	socket := a.new(0)
	process.poller.mu.Lock()
	process.poller.next = 1 << 32
	process.poller.mu.Unlock()
	f, err := a.in.Function("udp", "[method]udp-socket.subscribe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Call(nil, []component.Value{socket}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	id, err := a.in.ResourceRep(nil, p.socketType, socket.Resource)
	if err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	entry := p.entries[id]
	children := entry.children
	p.mu.Unlock()
	if children != 0 {
		t.Fatal("failed publication leaked", children)
	}
	// External descriptor closure injects a retained physical-cleanup failure.
	if err := entry.file.file.Close(); err != nil {
		t.Fatal(err)
	}
	a.drop(socket)
	p.Close()
	if err := p.WaitClosed(nil); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
}
