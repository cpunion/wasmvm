//go:build linux || darwin

package preview2

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
)

func httpLoopbackURL(_ context.Context, u *url.URL) error {
	if u.Hostname() == "127.0.0.1" || u.Hostname() == "::1" || u.Hostname() == "localhost" {
		return nil
	}
	return httpErr("HTTP-request-denied")
}
func httpPattern() []byte { return []byte(strings.Repeat("0123456789abcdef", 256)) }
func httpEcho(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil || !bytes.Equal(data, httpPattern()) || r.Trailer.Get("X-Request-Trailer") != "done" {
			t.Error("invalid request", len(data), r.Trailer, err)
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Trailer", "X-Reply-Trailer")
		w.WriteHeader(200)
		w.Write(data)
		w.Header().Set("X-Reply-Trailer", "done")
	})
}
func TestHTTPClientGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, secure := range []bool{false, true} {
			name := backend + "/http"
			if secure {
				name += "s"
			}
			t.Run(name, func(t *testing.T) {
				server := httptest.NewUnstartedServer(httpEcho(t))
				if secure {
					server.StartTLS()
				} else {
					server.Start()
				}
				defer server.Close()
				config := HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback}
				if secure {
					roots := x509.NewCertPool()
					roots.AddCert(server.Certificate())
					config.TLSConfig = &tls.Config{RootCAs: roots}
				}
				p, _, _ := httpOwner(t, config)
				in := httpInstance(t, p, backend, "http")
				u, _ := url.Parse(server.URL)
				f, err := in.Function("probe")
				if err != nil {
					t.Fatal(err)
				}
				secureBit := uint64(0)
				if secure {
					secureBit = 1
				}
				ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
				defer cancel()
				result, err := f.Call(ctx, []component.Value{{Text: u.Host}, {Bits: secureBit}}, component.CallOptions{Metered: true, Fuel: 10000000})
				if err != nil || result.Value.Bits != 42 {
					t.Fatal(result, err)
				}
				p.mu.Lock()
				left := len(p.entries)
				p.mu.Unlock()
				if left != 0 {
					t.Fatal("HTTP resource leak", left)
				}
			})
		}
	}
}
func TestHTTPIncomingGuestBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			p, _, _ := httpOwner(t, HTTPConfig{})
			in := httpInstance(t, p, backend, "http")
			request := httptest.NewRequest("POST", "http://localhost/echo", bytes.NewReader(httpPattern()))
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			response, err := p.CallIncoming(ctx, in, request, component.CallOptions{Metered: true, Fuel: 10000000})
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(response.Body)
			if err != nil || !bytes.Equal(data, httpPattern()) || response.Trailer.Get("X-Reply-Trailer") != "done" {
				t.Fatal(len(data), response.Trailer, err)
			}
			response.Body.Close()
		})
	}
}
