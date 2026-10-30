//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func tcpAllowLoopback(_ context.Context, address SocketAddress) error {
	if address.Address.Addr().IsLoopback() {
		return nil
	}
	return errors.New("outside loopback grant")
}
func tcpOwner(t *testing.T, config TCPConfig) (*TCP, *Network, *IO) {
	t.Helper()
	return tcpOwnerCleanup(t, config, nil)
}
func tcpOwnerCleanup(t *testing.T, config TCPConfig, wantCleanup error) (*TCP, *Network, *IO) {
	t.Helper()
	network, process := networkOwner(t, NetworkConfig{})
	provider, err := NewTCP(network, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		provider.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := provider.WaitClosed(ctx); !errors.Is(err, wantCleanup) {
			t.Error(err)
		}
	})
	return provider, network, process
}
func tcpInstance(t *testing.T, provider *TCP, backend, fixtureName string) *component.Instance {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	engine, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(engine.Close)
	module, err := engine.Compile(nil, fixture(t, fixtureName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(module.Close)
	in, err := module.Instantiate(nil, provider.Imports(), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}
func TestTCPGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, _, process := tcpOwner(t, TCPConfig{AllowBind: tcpAllowLoopback, AllowConnect: tcpAllowLoopback})
			in := tcpInstance(t, p, backend, "tcp")
			for _, family := range []uint64{0, 1} {
				f, err := in.Function("exercise")
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				result, err := f.Call(ctx, []component.Value{{Bits: family}}, component.CallOptions{Metered: true, Fuel: 10000000})
				cancel()
				if err != nil || result.Value.Bits != 42 {
					t.Fatal(family, result, err)
				}
			}
			for _, name := range []string{"setup", "transfer", "teardown"} {
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				var args []component.Value
				want := uint64(1)
				if name == "setup" {
					args = []component.Value{{}}
				}
				if name == "transfer" {
					args = []component.Value{{Bits: 128}}
					want = 128
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				result, err := f.Call(ctx, args, component.CallOptions{Metered: true, Fuel: 10000000})
				cancel()
				if err != nil || result.Value.Bits != want {
					t.Fatal(name, result, err)
				}
			}
			p.mu.Lock()
			remaining := len(p.sockets)
			p.mu.Unlock()
			process.mu.Lock()
			streams := len(process.records)
			process.mu.Unlock()
			if remaining != 0 || streams != 0 {
				t.Fatal("retained socket/stream resources", remaining, streams)
			}
		})
	}
}
