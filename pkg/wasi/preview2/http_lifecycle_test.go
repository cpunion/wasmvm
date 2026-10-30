//go:build linux || darwin

package preview2

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHTTPConstructorsLimitsAndShutdown(t *testing.T) {
	var zero HTTP
	zero.Close()
	if err := zero.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	if len(zero.Imports().Instances) != 0 {
		t.Fatal("zero provider minted imports")
	}
	if _, err := NewHTTP(nil, HTTPConfig{}); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	p, network, process := httpOwner(t, HTTPConfig{})
	for _, config := range []HTTPConfig{{MaxHeaderFields: 65537}, {MaxHeaderBytes: 17 << 20}, {BodyBufferBytes: 1<<20 + 1}, {Timeout: -1}} {
		if _, err := NewHTTP(network, config); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(config, err)
		}
	}
	if _, err := NewHTTP(network, HTTPConfig{ForbiddenHeaders: []string{"bad name"}}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	q, err := NewHTTP(network, HTTPConfig{TLSConfig: &tls.Config{RootCAs: roots, ClientCAs: roots}})
	if err != nil {
		t.Fatal(err)
	}
	if q.config.TLSConfig.RootCAs == roots || q.config.TLSConfig.ClientCAs == roots {
		t.Fatal("TLS trust pool alias")
	}
	q.Close()
	q.WaitClosed(nil)
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
	if err := p.startWorker(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	process.Close()
	if _, err := NewHTTP(network, HTTPConfig{}); err == nil {
		t.Fatal("closed IO accepted")
	}
	network.Close()
	if _, err := NewHTTP(network, HTTPConfig{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
func TestHTTPResourceQuotaAndPollRollback(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{MaxResources: 1})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	field := a.newFields()
	f, err := a.in.Function("types", "[constructor]fields")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Call(nil, nil, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	a.drop(field)
	p.mu.Lock()
	p.next = 1 << 32
	p.mu.Unlock()
	if _, err := f.Call(nil, nil, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}
func TestHTTPDialAuthorityResolverAndTransportLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	_, port, _ := net.SplitHostPort(u.Host)
	process, err := NewIO(IOConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { process.Close(); process.WaitClosed(nil) }()
	network, err := NewNetwork(process, NetworkConfig{AllowNameLookup: true, Resolve: func(ctx context.Context, name string) ([]netip.Addr, error) {
		switch name {
		case "panic.test":
			panic("resolver")
		case "empty.test":
			return nil, nil
		case "invalid.test":
			return []netip.Addr{{}}, nil
		case "failure.test":
			return nil, &net.DNSError{IsTimeout: true}
		default:
			return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { network.Close(); network.WaitClosed(nil) }()
	p, err := NewHTTP(network, HTTPConfig{AllowConnect: tcpAllowLoopback})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { p.Close(); p.WaitClosed(nil) }()
	conn, err := p.dial(context.Background(), "tcp", net.JoinHostPort("localhost", port), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	for _, name := range []string{"panic.test", "empty.test", "invalid.test", "failure.test"} {
		if _, err := p.dial(context.Background(), "tcp", net.JoinHostPort(name, port), time.Second); err == nil {
			t.Fatal(name)
		}
	}
	for _, name := range []string{"bad", "localhost:bad", "localhost:99999"} {
		if _, err := p.dial(context.Background(), "tcp", name, time.Second); err == nil {
			t.Fatal(name)
		}
	}
	for i := 1; i < 12; i++ {
		if _, err := p.transportFor(httpTimeouts{time.Duration(i) * time.Second, time.Second, time.Second}); err != nil {
			t.Fatal(err)
		}
	}
	if len(p.transports) != 8 {
		t.Fatal("unbounded transport cache", len(p.transports))
	}
	p.Close()
	if _, err := p.transportFor(p.defaultTimeouts()); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
func TestHTTPPolicyCancellationAndFailures(t *testing.T) {
	for _, action := range []string{"panic", "caller", "owner"} {
		t.Run(action, func(t *testing.T) {
			entered := make(chan struct{})
			callback := func(ctx context.Context, u *url.URL) error {
				if action == "panic" {
					panic("policy")
				}
				close(entered)
				<-ctx.Done()
				return ctx.Err()
			}
			p, _, _ := httpOwner(t, HTTPConfig{AllowRequest: callback})
			a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
			request := a.newRequest("http://127.0.0.1/echo")
			f, err := a.in.Function("outgoing", "handle")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				result, err := f.Call(ctx, []component.Value{request, {}}, component.CallOptions{})
				if action == "panic" && err == nil && (result.Value.Case != 1 || result.Value.Elements[0].Case != 15) {
					err = fmt.Errorf("invalid denied result %v", result)
				}
				done <- err
			}()
			if action != "panic" {
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("policy not called")
				}
				if action == "caller" {
					cancel()
				} else {
					p.Close()
				}
			}
			select {
			case err := <-done:
				if action == "caller" && !errors.Is(err, context.Canceled) || action == "owner" && !errors.Is(err, wasm.ErrClosed) || action == "panic" && err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("policy did not cancel")
			}
		})
	}
}
func TestHTTPHandlerStreamsIndependentInstances(t *testing.T) {
	engine, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	module, err := engine.Compile(nil, fixture(t, "http"))
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	factory := func(ctx context.Context) (*component.Instance, *HTTP, func(), error) {
		process, err := NewIO(IOConfig{})
		if err != nil {
			return nil, nil, nil, err
		}
		network, err := NewNetwork(process, NetworkConfig{})
		if err != nil {
			process.Close()
			return nil, nil, nil, err
		}
		p, err := NewHTTP(network, HTTPConfig{BodyBufferBytes: 257})
		if err != nil {
			return nil, nil, nil, err
		}
		imports := p.Imports()
		clock, _ := NewWallClock(WallClockConfig{})
		random, _ := NewRandom(RandomConfig{})
		for _, extra := range []component.Imports{clock.Imports(), random.Imports()} {
			for name, value := range extra.Instances {
				imports.Instances[name] = value
			}
		}
		in, err := module.Instantiate(ctx, imports, component.CallOptions{})
		if err != nil {
			return nil, nil, nil, err
		}
		cleanup := func() {
			p.Close()
			process.Close()
			network.Close()
			p.WaitClosed(nil)
			process.WaitClosed(nil)
			network.WaitClosed(nil)
			in.Close()
		}
		return in, p, cleanup, nil
	}
	server := httptest.NewServer(HTTPHandler(factory, component.CallOptions{Metered: true, Fuel: 10000000}))
	defer server.Close()
	client := &http.Client{Timeout: 3 * time.Second}
	for i := 0; i < 3; i++ {
		response, err := client.Post(server.URL+"/echo", "application/octet-stream", bytes.NewReader(httpPattern()))
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || !bytes.Equal(data, httpPattern()) || response.Trailer.Get("X-Reply-Trailer") != "done" {
			t.Fatal(len(data), err, response.Trailer)
		}
	}
}
func TestHTTPHandlerFactoryFailures(t *testing.T) {
	for _, factory := range []HTTPInstanceFactory{nil, func(context.Context) (*component.Instance, *HTTP, func(), error) {
		return nil, nil, func() {}, errors.New("factory")
	}, func(context.Context) (*component.Instance, *HTTP, func(), error) { return nil, nil, func() {}, nil }} {
		recorder := httptest.NewRecorder()
		HTTPHandler(factory, component.CallOptions{}).ServeHTTP(recorder, httptest.NewRequest("GET", "http://localhost", nil))
		if recorder.Code < 500 {
			t.Fatal(recorder.Code)
		}
	}
}
