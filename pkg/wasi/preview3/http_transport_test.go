//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func httpNetworkOwner(t *testing.T, networkConfig SocketsConfig, config HTTPConfig) *HTTP {
	t.Helper()
	network, _ := socketsOwner(t, networkConfig)
	p, err := NewHTTP(network, config)
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
	})
	return p
}
func TestHTTPNetworkTLSResolutionAndUploadTrailers(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, secure := range []bool{false, true} {
			payload := strings.Repeat("0123456789abcdef", 32)
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := io.ReadAll(r.Body)
				if err != nil || string(data) != payload || r.Trailer.Get("X-Upload") != "done" {
					t.Error("secure", secure, "protocol", r.Proto, len(data), r.Trailer, err)
				}
				w.Header().Set("Trailer", "X-End")
				w.Write([]byte(payload))
				w.Header().Set("X-End", "done")
			}))
			if secure {
				server.EnableHTTP2 = true
				server.StartTLS()
			} else {
				server.Start()
			}
			cfg := HTTPConfig{AllowConnect: socketsAllow, AllowRequest: func(context.Context, *url.URL) error { return nil }}
			target, _ := url.Parse(server.URL)
			authority := target.Host
			network := SocketsConfig{}
			if secure {
				roots := x509.NewCertPool()
				roots.AddCert(server.Certificate())
				cfg.TLSConfig = &tls.Config{RootCAs: roots}
			} else {
				_, port, _ := net.SplitHostPort(authority)
				authority = "wasmvm.test:" + port
				network.AllowNameLookup = true
				network.Resolve = func(_ context.Context, name string) ([]netip.Addr, error) {
					if name != "wasmvm.test" {
						t.Error(name)
					}
					return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
				}
			}
			p := httpNetworkOwner(t, network, cfg)
			a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
			fields := a.call("[constructor]fields")
			a.call("[method]fields.append", fields, component.Value{Text: "x-upload"}, component.Value{Text: "done"})
			req, tx := a.message("request", payload, httpOK(httpSome(fields)), component.Value{})
			a.call("[method]request.set-authority", req, httpSome(component.Value{Text: authority}))
			if secure {
				a.call("[method]request.set-scheme", req, httpSome(component.Value{Case: 1}))
			}
			result := a.invoke("client", "send", req)
			if result.Case != 0 {
				t.Fatal(result)
			}
			response := result.Elements[0]
			f := a.function("types", "[static]response.consume-body")
			_, typ := f.Info().Parameter(1)
			pair := a.call("[static]response.consume-body", response, a.source(typ, component.Value{}))
			var data string
			for {
				r := a.read(pair.Elements[0], 3)
				data += r.Bytes
				if r.Closed {
					break
				}
			}
			if data != payload {
				t.Fatal(len(data))
			}
			a.dropAsync(pair.Elements[0])
			trailers := a.read(pair.Elements[1], 1).Values[0]
			if trailers.Case != 0 {
				t.Fatal(trailers)
			}
			a.drop(trailers.Elements[0].Elements[0])
			a.dropAsync(pair.Elements[1])
			if r := a.read(tx, 1); r.Values[0].Case != 0 {
				t.Fatal(r)
			}
			a.dropAsync(tx)
			server.Close()
		}
	})
}
func TestHTTPTransportErrorsBoundsAndURI(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{{&net.DNSError{IsTimeout: true}, "DNS-timeout"}, {&net.DNSError{}, "DNS-error"}, {&tls.CertificateVerificationError{Err: errors.New("bad cert")}, "TLS-certificate-error"}, {x509.UnknownAuthorityError{}, "TLS-certificate-error"}, {tls.RecordHeaderError{}, "TLS-protocol-error"}, {context.DeadlineExceeded, "connection-timeout"}, {io.ErrUnexpectedEOF, "HTTP-response-incomplete"}, {errors.New("unknown"), "HTTP-protocol-error"}}
	for _, test := range cases {
		mapped := httpTransportError(test.err, "HTTP-protocol-error")
		if mapped.(*httpError).value.Case != httpErr(test.want).value.Case || !errors.Is(mapped, test.err) || mapped.Error() == "" {
			t.Fatal(mapped)
		}
	}
	if httpErr("unrecognized").Error() != "internal-error" {
		t.Fatal("unknown error")
	}
	for _, name := range httpErrorNames {
		if httpErrorName(httpErrorResult(httpErr(name))) != name {
			t.Fatal(name)
		}
	}
	httpErrorResult(errors.New("host error"))
	if httpHeaderFailure(errors.New("custom")).Elements[0].Case != 4 {
		t.Fatal("custom header error")
	}
	p := httpOwner(t, HTTPConfig{})
	for _, values := range [][]string{{}, {"0"}, {"3", "3"}, {"3", "4"}, {" 3"}, {"-1"}} {
		fields, _ := host.NewHTTPFields(p.fields, nil, false)
		for _, value := range values {
			fields.Mutate("content-length", []string{value}, true)
		}
		length, err := httpContentLength(fields)
		invalid := len(values) > 1 && values[1] != "3" || len(values) == 1 && (values[0] == " 3" || values[0] == "-1")
		if (err != nil) != invalid || len(values) > 0 && !invalid && length == nil {
			t.Fatal(values, length, err)
		}
	}
	authority := "localhost"
	for _, path := range []string{"?q=1", "/ok", "http://other/", "\n"} {
		message := &httpMessage{authority: &authority, path: &path}
		_, err := p.requestURL(message)
		if (err != nil) != (path == "\n" || strings.HasPrefix(path, "http:")) {
			t.Fatal(path, err)
		}
	}
	for _, value := range []string{"", "user@localhost", "localhost/path", "localhost?x=1", "%"} {
		_, err := p.requestURL(&httpMessage{authority: &value})
		if err == nil {
			t.Fatal(value)
		}
	}
	for i := 1; i <= 10; i++ {
		if _, err := p.transportFor(httpTimeouts{time.Duration(i), time.Second, time.Second}); err != nil {
			t.Fatal(err)
		}
	}
	last := httpTimeouts{10, time.Second, time.Second}
	a, _ := p.transportFor(last)
	b, _ := p.transportFor(last)
	if a != b || len(p.transports) != 8 {
		t.Fatal("transport cache")
	}
	p.Close()
	if _, err := p.transportFor(last); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if httpDuration(0) != time.Nanosecond || httpDuration(^uint64(0)) != time.Duration(1<<63-1) {
		t.Fatal("duration conversion")
	}
}
func TestHTTPDialPoliciesAndResolverFailures(t *testing.T) {
	for _, name := range []string{"no-grant", "bad-address", "bad-port", "lookup-denied", "name-limit", "idna", "empty", "too-many", "resolve-error", "resolve-panic", "invalid-ip", "zone", "deny", "refused"} {
		t.Run(name, func(t *testing.T) {
			n := SocketsConfig{AllowNameLookup: true, Resolve: func(context.Context, string) ([]netip.Addr, error) {
				return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
			}}
			cfg := HTTPConfig{AllowConnect: socketsAllow}
			address := "wasmvm.test:1"
			switch name {
			case "no-grant":
				cfg.AllowConnect = nil
			case "bad-address":
				address = "invalid"
			case "bad-port":
				address = "localhost:invalid"
			case "lookup-denied":
				n.AllowNameLookup = false
			case "name-limit":
				n.MaxNameBytes = 3
			case "idna":
				address = "bad name:1"
			case "empty":
				n.Resolve = func(context.Context, string) ([]netip.Addr, error) { return nil, nil }
			case "too-many":
				n.MaxAddresses = 1
				n.Resolve = func(context.Context, string) ([]netip.Addr, error) {
					return []netip.Addr{netip.IPv4Unspecified(), netip.IPv6Unspecified()}, nil
				}
			case "resolve-error":
				n.Resolve = func(context.Context, string) ([]netip.Addr, error) { return nil, &net.DNSError{} }
			case "resolve-panic":
				n.Resolve = func(context.Context, string) ([]netip.Addr, error) { panic("resolver") }
			case "invalid-ip":
				n.Resolve = func(context.Context, string) ([]netip.Addr, error) { return []netip.Addr{{}}, nil }
			case "zone":
				address = "[fe80::1%wasmvm-missing-interface]:1"
			case "deny":
				cfg.AllowConnect = func(context.Context, SocketAddress) error { return errors.New("denied") }
			case "refused":
				address = "127.0.0.1:1"
			}
			p := httpNetworkOwner(t, n, cfg)
			conn, err := p.dial(context.Background(), "tcp", address, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				t.Fatal("invalid destination admitted", name)
			}
		})
	}
}
func TestHTTPIncomingValidationAndErrors(t *testing.T) {
	p := httpOwner(t, HTTPConfig{Handler: func(context.Context, *http.Request) (*http.Response, error) {
		return nil, httpErr("HTTP-request-denied")
	}})
	in := instantiate(t, compileFixture(t, "interpreter", "http-handler-api"), p.Imports())
	request := httptest.NewRequest("GET", "http://localhost", nil)
	if _, err := p.CallIncoming(nil, nil, request, component.CallOptions{}); err == nil {
		t.Fatal("nil instance")
	}
	if _, err := p.CallIncoming(nil, in, nil, component.CallOptions{}); err == nil {
		t.Fatal("nil request")
	}
	if _, err := p.CallIncoming(nil, in, &http.Request{}, component.CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.CallIncoming(ctx, in, request, component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := p.CallIncoming(nil, in, request, component.CallOptions{}); httpErrorName(httpErrorResult(err)) != "HTTP-request-denied" {
		t.Fatal(err)
	}
	for _, field := range []string{"bad name", strings.Repeat("x", int(p.config.MaxHeaderBytes)+1)} {
		request := httptest.NewRequest("GET", "http://localhost", nil)
		request.Header.Set(field, "x")
		if _, err := p.CallIncoming(nil, in, request, component.CallOptions{}); err == nil {
			t.Fatal(strconv.Itoa(len(field)))
		}
	}
}
