package preview2

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type networkAPI struct {
	t  *testing.T
	in *component.Instance
}

func (a networkAPI) call(instance, name string, args ...component.Value) component.Value {
	a.t.Helper()
	f, err := a.in.Function(instance, name)
	if err != nil {
		a.t.Fatal(err)
	}
	result, err := f.Call(nil, args, component.CallOptions{})
	if err != nil {
		a.t.Fatal(name, err)
	}
	if result.Value == nil {
		return component.Value{}
	}
	return *result.Value
}
func (a networkAPI) drop(value component.Value) {
	a.t.Helper()
	if _, err := a.in.DropResource(nil, value.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func (a networkAPI) resolve(network component.Value, name string) component.Value {
	a.t.Helper()
	result := a.call("lookup", "resolve-addresses", network, component.Value{Text: name})
	if result.Case != 0 {
		a.t.Fatal(name, result)
	}
	return result.Elements[0]
}
func (a networkAPI) next(stream component.Value) component.Value {
	return a.call("lookup", "[method]resolve-address-stream.resolve-next-address", stream)
}
func networkOwner(t *testing.T, config NetworkConfig) (*Network, *IO) {
	t.Helper()
	process, err := NewIO(IOConfig{})
	if err != nil {
		t.Fatal(err)
	}
	p, err := NewNetwork(process, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Close()
		process.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := p.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
		if err := process.WaitClosed(ctx); err != nil {
			t.Error(err)
		}
	})
	return p, process
}
func networkInstance(t *testing.T, p *Network, backend, fixtureName string) *component.Instance {
	t.Helper()
	config := component.DefaultConfig()
	config.Core.Backend = backend
	e, err := component.NewEngine(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	m, err := e.Compile(nil, fixture(t, fixtureName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	in, err := m.Instantiate(nil, p.Imports(), component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(in.Close)
	return in
}
func TestNetworkGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: true})
			in := networkInstance(t, p, backend, "network")
			for _, name := range []string{"exercise", "lookup-batch"} {
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				var args []component.Value
				want := uint64(42)
				if name == "lookup-batch" {
					args = []component.Value{{Bits: 3}}
					want = 3
				}
				result, err := f.Call(nil, args, component.CallOptions{Metered: true, Fuel: 100000})
				if err != nil || result.Value.Bits != want {
					t.Fatal(name, result, err)
				}
			}
			p.mu.Lock()
			count := len(p.streams)
			p.mu.Unlock()
			if count != 0 {
				t.Fatal("streams leaked", count)
			}
		})
	}
}
func TestNetworkLookupAsyncAndCancellation(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: true, Resolve: func(ctx context.Context, name string) ([]netip.Addr, error) {
		if name != "xn--bcher-kva.example" {
			return nil, errors.New("name was not IDNA normalized")
		}
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return []netip.Addr{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("2001:db8::1"), netip.MustParseAddr("::ffff:192.0.2.1")}, nil
	}})
	a := networkAPI{t, networkInstance(t, p, "interpreter", "network-api")}
	network := a.call("instance-network", "instance-network")
	stream := a.resolve(network, "bücher.example")
	<-entered
	value := a.next(stream)
	if value.Case != 1 || value.Elements[0].Bits != uint64(NetworkWouldBlock) {
		t.Fatal(value)
	}
	poll := a.call("lookup", "[method]resolve-address-stream.subscribe", stream)
	if ready := a.call("poll", "[method]pollable.ready", poll); ready.Bits != 0 {
		t.Fatal(ready)
	}
	block, err := a.in.Function("poll", "[method]pollable.block")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := block.Call(ctx, []component.Value{poll}, component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	close(release)
	if _, err := block.Call(nil, []component.Value{poll}, component.CallOptions{}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []netip.Addr{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("2001:db8::1")} {
		value := a.next(stream)
		if value.Case != 0 || value.Elements[0].Case != 1 {
			t.Fatal(value)
		}
		got := value.Elements[0].Elements[0]
		expected := ipVariantValue(want)
		if got.Case != expected.Case || len(got.Elements[0].Elements) != len(expected.Elements[0].Elements) {
			t.Fatal(got, expected)
		}
		for i, field := range got.Elements[0].Elements {
			if field.Bits != expected.Elements[0].Elements[i].Bits {
				t.Fatal(got, expected)
			}
		}
	}
	for i := 0; i < 2; i++ {
		value := a.next(stream)
		if value.Case != 0 || value.Elements[0].Case != 0 {
			t.Fatal(value)
		}
	}
	a.drop(poll)
	a.drop(stream)
	a.drop(network)
}
func TestNetworkPoliciesLimitsAndLiterals(t *testing.T) {
	p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: false})
	a := networkAPI{t, networkInstance(t, p, "interpreter", "network-api")}
	network := a.call("instance-network", "instance-network")
	denied := a.call("lookup", "resolve-addresses", network, component.Value{Text: "127.0.0.1"})
	if denied.Case != 1 || denied.Elements[0].Bits != uint64(NetworkAccessDenied) {
		t.Fatal(denied)
	}
	a.drop(network)
	p, _ = networkOwner(t, NetworkConfig{AllowNameLookup: true, MaxAddressStreams: 1, MaxNameBytes: 64})
	a = networkAPI{t, networkInstance(t, p, "interpreter", "network-api")}
	network = a.call("instance-network", "instance-network")
	for _, name := range []string{"", "a\x00b", "[bad", "fe80::1%en0", "name with spaces", "bad/host", "bad:port", strings.Repeat("a", 65)} {
		result := a.call("lookup", "resolve-addresses", network, component.Value{Text: name})
		if result.Case != 1 || result.Elements[0].Bits != uint64(NetworkInvalidArgument) {
			t.Fatal(name, result)
		}
	}
	for _, name := range []string{"127.0.0.1", "::1", "[::1]", "::ffff:127.0.0.1"} {
		stream := a.resolve(network, name)
		next := a.next(stream)
		if next.Case != 0 || next.Elements[0].Case != 1 {
			t.Fatal(name, next)
		}
		ip := next.Elements[0].Elements[0]
		if name == "127.0.0.1" && ip.Case != 0 || name != "127.0.0.1" && ip.Case != 1 {
			t.Fatal(name, ip)
		}
		a.drop(stream)
	}
	stream := a.resolve(network, "127.0.0.1")
	limited := a.call("lookup", "resolve-addresses", network, component.Value{Text: "::1"})
	if limited.Case != 1 || limited.Elements[0].Bits != uint64(NetworkOutOfMemory) {
		t.Fatal(limited)
	}
	a.drop(stream)
	p.next = 1 << 32
	limited = a.call("lookup", "resolve-addresses", network, component.Value{Text: "::1"})
	if limited.Case != 1 || limited.Elements[0].Bits != uint64(NetworkOutOfMemory) {
		t.Fatal(limited)
	}
	a.drop(network)
}
func TestNetworkResolverErrors(t *testing.T) {
	tests := []struct {
		name      string
		addresses []netip.Addr
		err       error
		code      NetworkErrorCode
		panic     bool
	}{
		{name: "not-found", err: &net.DNSError{IsNotFound: true}, code: NetworkNameUnresolvable},
		{name: "timeout", err: &net.DNSError{IsTimeout: true}, code: NetworkTimeout},
		{name: "temporary", err: &net.DNSError{IsTemporary: true}, code: NetworkTemporaryResolverFailure},
		{name: "permanent", err: &net.DNSError{}, code: NetworkPermanentResolverFailure},
		{name: "unknown", err: errors.New("host failure"), code: NetworkUnknown},
		{name: "empty", code: NetworkNameUnresolvable},
		{name: "invalid", addresses: []netip.Addr{{}}, code: NetworkPermanentResolverFailure},
		{name: "zone", addresses: []netip.Addr{netip.MustParseAddr("fe80::1%en0")}, code: NetworkPermanentResolverFailure},
		{name: "too-many", addresses: []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")}, code: NetworkOutOfMemory},
		{name: "panic", panic: true, code: NetworkPermanentResolverFailure},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: true, MaxAddresses: 1, Resolve: func(context.Context, string) ([]netip.Addr, error) {
				if test.panic {
					panic("resolver failed")
				}
				return test.addresses, test.err
			}})
			a := networkAPI{t, networkInstance(t, p, "interpreter", "network-api")}
			network := a.call("instance-network", "instance-network")
			stream := a.resolve(network, "example.test")
			poll := a.call("lookup", "[method]resolve-address-stream.subscribe", stream)
			a.call("poll", "[method]pollable.block", poll)
			for i := 0; i < 2; i++ {
				value := a.next(stream)
				if value.Case != 1 || value.Elements[0].Bits != uint64(test.code) {
					t.Fatal(value, test.code)
				}
			}
			a.drop(poll)
			a.drop(stream)
			a.drop(network)
		})
	}
}
func TestNetworkOwnerDropAndTimeout(t *testing.T) {
	for _, mode := range []string{"owner-close", "stream-drop", "timeout", "parent-with-child"} {
		t.Run(mode, func(t *testing.T) {
			entered, ended := make(chan struct{}), make(chan struct{})
			timeout := time.Second
			if mode == "timeout" {
				timeout = time.Millisecond
			}
			p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: true, ResolveTimeout: timeout, Resolve: func(ctx context.Context, _ string) ([]netip.Addr, error) {
				close(entered)
				defer close(ended)
				<-ctx.Done()
				return nil, ctx.Err()
			}})
			a := networkAPI{t, networkInstance(t, p, "interpreter", "network-api")}
			network := a.call("instance-network", "instance-network")
			stream := a.resolve(network, "example.test")
			<-entered
			var poll component.Value
			if mode != "stream-drop" {
				poll = a.call("lookup", "[method]resolve-address-stream.subscribe", stream)
			}
			switch mode {
			case "owner-close":
				p.Close()
				if err := p.WaitClosed(nil); err != nil {
					t.Fatal(err)
				}
				a.call("poll", "[method]pollable.block", poll)
				a.drop(poll)
				a.drop(stream)
			case "stream-drop":
				a.drop(stream)
			case "timeout":
				a.call("poll", "[method]pollable.block", poll)
				value := a.next(stream)
				if value.Case != 1 || value.Elements[0].Bits != uint64(NetworkTimeout) {
					t.Fatal(value)
				}
				a.drop(poll)
				a.drop(stream)
			case "parent-with-child":
				if _, err := a.in.DropResource(nil, stream.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
					t.Fatal(err)
				}
				a.call("poll", "[method]pollable.block", poll)
				a.drop(poll)
			}
			select {
			case <-ended:
			case <-time.After(time.Second):
				t.Fatal("resolver retained after cancellation")
			}
			a.drop(network)
			var wg sync.WaitGroup
			for i := 0; i < 4; i++ {
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
		})
	}
}
func TestNetworkConstructorAndErrorCodes(t *testing.T) {
	if _, err := NewNetwork(nil, NetworkConfig{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	var zero Network
	zero.Close()
	if err := zero.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if len(zero.Imports().Instances) != 0 {
		t.Fatal("zero provider imports")
	}
	process, err := NewIO(IOConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer process.Close()
	for _, config := range []NetworkConfig{{MaxAddresses: 1 << 21}, {MaxNameBytes: 1 << 21}, {ResolveTimeout: -time.Second}} {
		if _, err := NewNetwork(process, config); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
	}
	process.Close()
	if _, err := NewNetwork(process, NetworkConfig{}); err == nil {
		t.Fatal("accepted closed IO")
	}
	err = &NetworkError{Code: NetworkAccessDenied}
	if err.Error() == "" || networkError(err) != NetworkAccessDenied {
		t.Fatal(err)
	}
	cause := errors.New("network failure")
	err = &NetworkError{Code: NetworkConnectionReset, Cause: cause}
	if err.Error() != cause.Error() || !errors.Is(err, cause) {
		t.Fatal(err)
	}
	if code := networkError(context.DeadlineExceeded); code != NetworkTimeout {
		t.Fatal(code)
	}
}

func TestNetworkCompleteContractsAndCommonSubset(t *testing.T) {
	p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: true})
	in := networkInstance(t, p, "interpreter", "network-common")
	exercise, err := in.Function("exercise")
	if err != nil {
		t.Fatal(err)
	}
	if result, err := exercise.Call(nil, nil, component.CallOptions{}); err != nil || result.Value.Bits != 42 {
		t.Fatal(result, err)
	}
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for _, name := range []string{"error-label", "ip-label", "ip-width", "lookup-name", "subscribe-result", "network-result"} {
		module, err := engine.Compile(nil, fixture(t, "network-contract-"+name))
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
func TestNetworkDefaultResolver(t *testing.T) {
	p, _ := networkOwner(t, NetworkConfig{AllowNameLookup: true})
	a := networkAPI{t, networkInstance(t, p, "interpreter", "network-api")}
	network := a.call("instance-network", "instance-network")
	stream := a.resolve(network, "localhost")
	poll := a.call("lookup", "[method]resolve-address-stream.subscribe", stream)
	function, err := a.in.Function("poll", "[method]pollable.block")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := function.Call(ctx, []component.Value{poll}, component.CallOptions{}); err != nil {
		t.Fatal(err)
	}
	value := a.next(stream)
	if value.Case != 0 || value.Elements[0].Case != 1 {
		t.Fatal(value)
	}
	ip := value.Elements[0].Elements[0]
	if ip.Case == 0 {
		if ip.Elements[0].Elements[0].Bits != 127 {
			t.Fatal("localhost not loopback", ip)
		}
	} else {
		for i, v := range ip.Elements[0].Elements {
			want := uint64(0)
			if i == 7 {
				want = 1
			}
			if v.Bits != want {
				t.Fatal("localhost not loopback", ip)
			}
		}
	}
	a.drop(poll)
	a.drop(stream)
	a.drop(network)
}
func TestNetworkProvenanceAndExpiredPublication(t *testing.T) {
	p, process := networkOwner(t, NetworkConfig{AllowNameLookup: true})
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	module, err := engine.Compile(nil, fixture(t, "network-api"))
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	imports := p.Imports()
	get := imports.Instances[InstanceNetworkInterface].Functions["instance-network"]
	var saved *component.Caller
	imports.Instances[InstanceNetworkInterface].Functions["instance-network"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
		saved = c
		for _, cause := range []error{errors.New("unrelated"), &NetworkError{Code: NetworkConnectionReset}} {
			value, err := process.errors.NewError(c, cause)
			if err != nil {
				return nil, err
			}
			got, err := p.errorCode(c, []component.Value{value})
			if err != nil {
				return nil, err
			}
			_, typed := cause.(*NetworkError)
			if typed && (got.Case != 1 || got.Elements[0].Bits != uint64(NetworkConnectionReset)) || !typed && got.Case != 0 {
				return nil, errors.New("network provenance changed")
			}
			if err = c.DropResource(value.Resource); err != nil {
				return nil, err
			}
		}
		return get(c, args)
	}
	in, err := module.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	a := networkAPI{t, in}
	network := a.call("instance-network", "instance-network")
	if _, err := p.resolve(saved, []component.Value{network, {Text: "127.0.0.1"}}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := p.stream(saved, component.Value{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := p.errorCode(saved, []component.Value{{}}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	// Inject a failed resource publication while the outer Caller remains active.
	old := imports.Instances[IPNameLookupInterface].Functions["resolve-addresses"]
	imports.Instances[IPNameLookupInterface].Functions["resolve-addresses"] = func(c *component.Caller, args []component.Value) (*component.Value, error) {
		original := p.streamType
		p.streamType = component.NewHostResourceType(nil)
		_, err := p.resolve(c, args)
		p.streamType = original
		if err == nil {
			return nil, errors.New("unbound resource was published")
		}
		return old(c, args)
	}
	// The first Instance's imports are immutable snapshots; a second binding
	// exercises the modified host callback in its own nominal Store.
	second, err := module.Instantiate(nil, imports, component.CallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	b := networkAPI{t, second}
	other := b.call("instance-network", "instance-network")
	stream := b.resolve(other, "127.0.0.1")
	b.drop(stream)
	b.drop(other)
	p.mu.Lock()
	count := len(p.streams)
	p.mu.Unlock()
	if count != 0 {
		t.Fatal("failed publication retained stream", count)
	}
	a.drop(network)
	p.Close()
	if err := p.WaitClosed(nil); err != nil {
		t.Fatal(err)
	}
	if err := p.begin(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if err := p.check(context.Background()); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
