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

func udpOwner(t *testing.T, config UDPConfig) (*UDP, *Network, *IO) {
	t.Helper()
	return udpOwnerCleanup(t, config, nil)
}
func udpOwnerCleanup(t *testing.T, config UDPConfig, wantCleanup error) (*UDP, *Network, *IO) {
	t.Helper()
	network, process := networkOwner(t, NetworkConfig{})
	provider, err := NewUDP(network, config)
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
func udpInstance(t *testing.T, p *UDP, backend, name string) *component.Instance {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	engine, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(engine.Close)
	module, err := engine.Compile(nil, fixture(t, name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(module.Close)
	in, err := module.Instantiate(nil, p.Imports(), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}
func TestUDPGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, _, _ := udpOwner(t, UDPConfig{AllowBind: tcpAllowLoopback, AllowSend: tcpAllowLoopback, AllowReceive: tcpAllowLoopback})
			in := udpInstance(t, p, backend, "udp")
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
					args = []component.Value{{}, {Bits: 1}}
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
			remaining := len(p.entries)
			p.mu.Unlock()
			if remaining != 0 {
				t.Fatal("leaked resources", remaining)
			}
		})
	}
}
