//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func httpErrorName(v component.Value) string {
	if v.Case != 1 || len(v.Elements) != 1 {
		return ""
	}
	return httpErrorNames[v.Elements[0].Case]
}
func httpCheckError(t *testing.T, v component.Value, want string) {
	t.Helper()
	if got := httpErrorName(v); got != want {
		t.Fatal("HTTP error", got, "want", want, v)
	}
}
func TestHTTPClientFailuresAndEarlyReply(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, name := range []string{"policy", "connect-policy", "missing-authority", "scheme", "missing-handler", "handler-error", "handler-panic", "nil-response", "bad-status", "bad-header", "message-limit", "early-response", "options", "length"} {
			t.Run(name, func(t *testing.T) {
				cfg := HTTPConfig{Handler: func(context.Context, *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: http.NoBody}, nil
				}}
				switch name {
				case "policy":
					cfg.AllowRequest = func(context.Context, *url.URL) error { panic("deny") }
				case "connect-policy":
					cfg.AllowRequest = func(context.Context, *url.URL) error { return nil }
				case "missing-handler":
					cfg.Handler = nil
				case "handler-error":
					cfg.Handler = func(context.Context, *http.Request) (*http.Response, error) { return nil, io.ErrUnexpectedEOF }
				case "handler-panic":
					cfg.Handler = func(context.Context, *http.Request) (*http.Response, error) { panic("failed") }
				case "nil-response":
					cfg.Handler = func(context.Context, *http.Request) (*http.Response, error) { return nil, nil }
				case "bad-status":
					cfg.Handler = func(context.Context, *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: 99}, nil
					}
				case "bad-header":
					cfg.Handler = func(context.Context, *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: 200, Header: http.Header{"bad name": []string{"x"}}}, nil
					}
				case "message-limit":
					cfg.MaxMessages = 1
				}
				p := httpOwner(t, cfg)
				a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
				opts := component.Value{}
				if name == "options" {
					v := a.call("[constructor]request-options")
					for _, kind := range []string{"connect", "first-byte", "between-bytes"} {
						a.call("[method]request-options.set-"+kind+"-timeout", v, httpSome(component.Value{Bits: ^uint64(0)}))
					}
					opts = httpSome(v)
				}
				req, tx := a.message("request", "abcdef", httpOK(component.Value{}), opts)
				if name != "missing-authority" {
					a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "127.0.0.1:1"}))
				}
				if name == "scheme" {
					a.call("[method]request.set-scheme", req, httpSome(component.Value{Case: 2, Elements: []component.Value{{Text: "custom"}}}))
				}
				if name == "length" {
					p.mu.Lock()
					for _, e := range p.entries {
						if e.kind == "request" {
							f, _ := host.NewHTTPFields(p.fields, []host.HTTPField{{Name: "content-length", Value: "not a number"}}, true)
							e.message.headers = f
						}
					}
					p.mu.Unlock()
				}
				iface, method := "handler", "handle"
				if name == "policy" || name == "connect-policy" {
					iface, method = "client", "send"
				}
				result := a.invoke(iface, method, req)
				want := map[string]string{"policy": "HTTP-request-denied", "connect-policy": "destination-IP-prohibited", "missing-authority": "HTTP-request-URI-invalid", "scheme": "HTTP-protocol-error", "missing-handler": "configuration-error", "handler-error": "HTTP-response-incomplete", "handler-panic": "internal-error", "nil-response": "HTTP-protocol-error", "bad-status": "HTTP-protocol-error", "bad-header": "HTTP-response-header-section-size", "message-limit": "connection-limit-reached", "length": "HTTP-protocol-error"}[name]
				if want != "" {
					httpCheckError(t, result, want)
				} else {
					if result.Case != 0 {
						t.Fatal(result)
					}
					a.drop(result.Elements[0])
				}
				r := a.read(tx, 1)
				if len(r.Values) != 1 {
					t.Fatal(r)
				}
				a.dropAsync(tx)
			})
		}
	})
}
func TestHTTPBodyTrailerAndHandlingFailure(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, mode := range []string{"trailers", "trailer-error", "handling-error", "drop-body", "drop-trailers", "drop-message", "provider-close"} {
			t.Run(mode, func(t *testing.T) {
				p := httpOwner(t, HTTPConfig{})
				a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
				trailing := httpOK(component.Value{})
				if mode == "trailers" {
					f := a.call("[constructor]fields")
					a.call("[method]fields.append", f, component.Value{Text: "x-end"}, component.Value{Text: "value"})
					trailing = httpOK(httpSome(f))
				}
				if mode == "trailer-error" {
					trailing = httpErrorResult(httpErr("HTTP-response-incomplete"))
				}
				message, tx := a.message("response", "body", trailing, component.Value{})
				if mode == "drop-message" || mode == "provider-close" {
					if mode == "drop-message" {
						a.drop(message)
					} else {
						p.Close()
					}
					r := a.read(tx, 1)
					if len(r.Values) != 1 || r.Values[0].Case != 1 {
						t.Fatal(r)
					}
					a.dropAsync(tx)
					return
				}
				f := a.function("types", "[static]response.consume-body")
				_, typ := f.Info().Parameter(1)
				handled := component.Value{}
				if mode == "handling-error" {
					handled = httpErrorResult(httpErr("HTTP-response-body-size"))
				}
				pair := a.call("[static]response.consume-body", message, a.source(typ, handled))
				if mode == "drop-body" {
					a.dropAsync(pair.Elements[0])
				} else if mode == "drop-trailers" {
					a.dropAsync(pair.Elements[1])
				}
				if mode != "drop-body" {
					for {
						r := a.read(pair.Elements[0], 3)
						if r.Closed {
							break
						}
					}
					a.dropAsync(pair.Elements[0])
				}
				if mode != "drop-trailers" {
					r := a.read(pair.Elements[1], 1)
					if len(r.Values) != 1 {
						t.Fatal(r)
					}
					if mode == "trailers" {
						v := r.Values[0].Elements[0].Elements[0]
						if a.call("[method]fields.get", v, component.Value{Text: "x-end"}).Elements[0].Text != "value" {
							t.Fatal(r)
						}
						a.drop(v)
					} else if mode != "trailer-error" && mode != "handling-error" && mode != "drop-body" && r.Values[0].Case != 0 {
						t.Fatal(r)
					}
					a.dropAsync(pair.Elements[1])
				}
				r := a.read(tx, 1)
				if len(r.Values) != 1 || mode == "trailers" && r.Values[0].Case != 0 || mode != "trailers" && r.Values[0].Case != 1 {
					t.Fatal(mode, r)
				}
				a.dropAsync(tx)
			})
		}
	})
}
func TestHTTPCancellationAndReadDeadline(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, closeProvider := range []bool{false, true} {
			entered := make(chan struct{})
			p := httpOwner(t, HTTPConfig{Handler: func(ctx context.Context, _ *http.Request) (*http.Response, error) {
				close(entered)
				<-ctx.Done()
				return nil, ctx.Err()
			}})
			a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
			req, tx := a.message("request", "", httpOK(component.Value{}), component.Value{})
			a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "localhost"}))
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() {
				_, err := a.function("handler", "handle").Call(ctx, []component.Value{req}, component.CallOptions{})
				done <- err
			}()
			<-entered
			if closeProvider {
				p.Close()
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if err == nil && !closeProvider {
					t.Fatal("cancelled call succeeded")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("cancelled call stranded")
			}
			cancel()
			a.dropAsync(tx)
		}
		blocked := newHTTPBlockedReader()
		p := httpOwner(t, HTTPConfig{Handler: func(context.Context, *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: blocked, ContentLength: -1}, nil
		}})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		options := a.call("[constructor]request-options")
		a.call("[method]request-options.set-between-bytes-timeout", options, httpSome(component.Value{Bits: uint64(time.Millisecond)}))
		req, tx := a.message("request", "", httpOK(component.Value{}), httpSome(options))
		a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "localhost"}))
		response := a.invoke("handler", "handle", req).Elements[0]
		f := a.function("types", "[static]response.consume-body")
		_, typ := f.Info().Parameter(1)
		pair := a.call("[static]response.consume-body", response, a.source(typ, component.Value{}))
		r := a.read(pair.Elements[0], 1)
		if !r.Closed {
			t.Fatal(r)
		}
		r = a.read(pair.Elements[1], 1)
		httpCheckError(t, r.Values[0], "connection-read-timeout")
		a.dropAsync(pair.Elements[0])
		a.dropAsync(pair.Elements[1])
		a.dropAsync(tx)
	})
}

type httpBlockedReader struct {
	done chan struct{}
	once sync.Once
}

func newHTTPBlockedReader() *httpBlockedReader        { return &httpBlockedReader{done: make(chan struct{})} }
func (r *httpBlockedReader) Read([]byte) (int, error) { <-r.done; return 0, io.ErrClosedPipe }
func (r *httpBlockedReader) Close() error             { r.once.Do(func() { close(r.done) }); return nil }
func TestHTTPContractAndProvenance(t *testing.T) {
	data, err := os.ReadFile("testdata/http.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		WASIVersion     string                       `json:"wasiVersion"`
		Producer        string                       `json:"producer"`
		GeneratorSHA256 string                       `json:"generatorSHA256"`
		SHA256          map[string]map[string]string `json:"sha256"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil || manifest.WASIVersion != "0.3.2" || manifest.Producer != "wasm-tools 1.258.0" || len(manifest.SHA256) != 6 {
		t.Fatal(manifest, err)
	}
	generator, _ := os.ReadFile("../../../scripts/build_preview3_http_fixtures.py")
	sum := sha256.Sum256(generator)
	if manifest.GeneratorSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("HTTP generator changed")
	}
	for name, files := range manifest.SHA256 {
		for ext, want := range files {
			data, err := os.ReadFile("testdata/" + name + "." + ext)
			sum := sha256.Sum256(data)
			if err != nil || want != hex.EncodeToString(sum[:]) {
				t.Fatal(name, ext, err)
			}
		}
	}

	var schema struct {
		Version   string `json:"wasiVersion"`
		WIT       string `json:"witArchiveSHA256"`
		Generator string `json:"generatorSHA256"`
		Schema    string `json:"schemaSHA256"`
	}
	schemaData, err := os.ReadFile("testdata/http-schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(schemaData, &schema); err != nil || schema.Version != "0.3.2" || schema.WIT != "2e918e047c6ad251a44b4d1141359495a27219058fd7090f8b8142f6bc0ec311" {
		t.Fatal(schema, err)
	}
	for path, want := range map[string]string{"../../../scripts/build_preview3_http_schema.py": schema.Generator, "http_schema.go": schema.Schema} {
		data, err := os.ReadFile(path)
		sum := sha256.Sum256(data)
		if err != nil || want != hex.EncodeToString(sum[:]) {
			t.Fatal(path, err)
		}
	}
	p := httpOwner(t, HTTPConfig{})
	for _, name := range []string{"http-wrong-async", "http-wrong-header-error"} {
		m := compileFixture(t, "interpreter", name)
		if in, err := m.Instantiate(nil, p.Imports(), component.CallOptions{}); in != nil || !errors.Is(err, vm.ErrType) && !errors.Is(err, vm.ErrLink) {
			t.Fatal(name, in, err)
		}
	}
	if p, err := NewHTTP(nil, HTTPConfig{}); p != nil || !errors.Is(err, vm.ErrReference) {
		t.Fatal(p, err)
	}
	n, _ := socketsOwner(t, SocketsConfig{})
	for _, cfg := range []HTTPConfig{{Timeout: -1}, {MaxMessages: 1 << 21}, {MaxHeaderBytes: 1 << 25}, {MaxHeaderFields: 1 << 17}, {BodyBufferBytes: 1 << 21}, {ForbiddenHeaders: []string{"bad header"}}} {
		if p, err := NewHTTP(n, cfg); p != nil || err == nil {
			t.Fatal(cfg, p, err)
		}
	}
	var invalid HTTP
	if err := invalid.WaitClosed(nil); !errors.Is(err, vm.ErrReference) {
		t.Fatal(err)
	}
	invalid.Close()
	if len(invalid.Imports().Instances) != 0 {
		t.Fatal("zero imports")
	}
	p.Close()
	if err := p.begin(context.Background()); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.WaitClosed(ctx); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestHTTPHandlerFactoryFailures(t *testing.T) {
	for _, test := range []struct {
		factory HTTPInstanceFactory
		status  int
	}{{nil, 503}, {func(context.Context) (*component.Instance, *HTTP, func(), error) {
		return nil, nil, func() {}, errors.New("failed")
	}, 503}, {func(context.Context) (*component.Instance, *HTTP, func(), error) { return nil, nil, nil, nil }, 500}} {
		w := httptest.NewRecorder()
		HTTPHandler(test.factory, component.CallOptions{}).ServeHTTP(w, httptest.NewRequest("GET", "http://localhost", nil))
		if w.Code != test.status {
			t.Fatal(w.Code)
		}
	}
}

func TestHTTPMetadataValidationAndChildOwnership(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{MaxHeaderFields: 2, MaxHeaderBytes: 32})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		for _, entries := range [][]host.HTTPField{{{Name: "A", Value: "one"}, {Name: "B", Value: "two"}}, {{Name: "bad name", Value: "x"}}, {{Name: "x", Value: "one"}, {Name: "y", Value: "two"}, {Name: "z", Value: "three"}}, {{Name: "x", Value: strings.Repeat("a", 32)}}} {
			r := a.call("[static]fields.from-list", httpFieldsValue(entries))
			valid := len(entries) == 2
			if valid {
				fields := r.Elements[0]
				if len(a.call("[method]fields.copy-all", fields).Elements) != 2 {
					t.Fatal(r)
				}
				a.call("[method]fields.set", fields, component.Value{Text: "a"}, httpStringsValue([]string{"new"}))
				a.call("[method]fields.delete", fields, component.Value{Text: "b"})
				if a.call("[method]fields.has", fields, component.Value{Text: "b"}).Bits != 0 {
					t.Fatal("delete")
				}
				a.drop(fields)
			} else if r.Case != 1 {
				t.Fatal(entries, r)
			}
		}
		options := a.call("[constructor]request-options")
		a.call("[method]request-options.set-connect-timeout", options, httpSome(component.Value{Bits: 1}))
		a.call("[method]request-options.set-connect-timeout", options, component.Value{})
		if a.call("[method]request-options.get-connect-timeout", options).Case != 0 {
			t.Fatal("clear timeout")
		}
		req, tx := a.message("request", "", httpOK(component.Value{}), httpSome(options))
		if a.call("[method]request.get-scheme", req).Case != 0 {
			t.Fatal("default scheme")
		}
		for _, text := range []string{"http", "https", "custom"} {
			scheme := component.Value{}
			switch text {
			case "https":
				scheme.Case = 1
			case "custom":
				scheme.Case = 2
				scheme.Elements = []component.Value{{Text: text}}
			}
			if a.call("[method]request.set-scheme", req, httpSome(scheme)).Case != 0 || httpParseScheme(a.call("[method]request.get-scheme", req)) == nil {
				t.Fatal(text)
			}
		}
		for _, key := range []string{"scheme", "authority", "path-with-query"} {
			a.call("[method]request.set-"+key, req, component.Value{})
			if a.call("[method]request.get-"+key, req).Case != 0 {
				t.Fatal(key)
			}
		}
		for _, entry := range []struct {
			key   string
			value component.Value
		}{{"scheme", httpSome(component.Value{Case: 2, Elements: []component.Value{{Text: "1bad"}}})}, {"authority", httpSome(component.Value{Text: "bad/extra"})}, {"path-with-query", httpSome(component.Value{Text: "\n"})}, {"method", component.Value{Case: 9, Elements: []component.Value{{Text: "bad method"}}}}} {
			if a.call("[method]request.set-"+entry.key, req, entry.value).Case != 1 {
				t.Fatal(entry.key)
			}
		}
		a.call("[method]request.set-method", req, component.Value{Case: 9, Elements: []component.Value{{Text: "CUSTOM"}}})
		if r := a.call("[method]request.get-method", req); r.Case != 9 || r.Elements[0].Text != "CUSTOM" {
			t.Fatal(r)
		}
		a.call("[method]request.set-path-with-query", req, httpSome(component.Value{Text: "?q=1"}))
		child := a.call("[method]request.get-options", req).Elements[0]
		if _, err := a.in.DropResource(nil, req.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
			t.Fatal("parent with live child dropped", err)
		}
		a.drop(child)
		r := a.read(tx, 1)
		if r.Values[0].Case != 1 {
			t.Fatal(r)
		}
		a.dropAsync(tx)
	})
}
func TestHTTPTrailerErrorInterruptsPendingBody(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
			_, err := io.Copy(io.Discard, r.Body)
			return nil, err
		}})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		f := a.function("types", "[static]request.new")
		_, bt := f.Info().Parameter(1)
		_, tt := f.Info().Parameter(2)
		body, err := a.in.NewSource(nil, bt.Element(), component.AsyncSource{Read: func(c *component.Caller, _ uint32) (*component.ReadCompletion, error) {
			completion := component.NewReadCompletion()
			ctx := c.Context()
			go func() { <-ctx.Done(); completion.Cancelled() }()
			return completion, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		pair := a.call("[static]request.new", a.call("[constructor]fields"), httpSome(body), a.source(tt, httpErrorResult(httpErr("HTTP-response-incomplete"))), component.Value{})
		a.call("[method]request.set-authority", pair.Elements[0], httpSome(component.Value{Text: "localhost"}))
		a.invoke("handler", "handle", pair.Elements[0])
		httpCheckError(t, a.read(pair.Elements[1], 1).Values[0], "HTTP-response-incomplete")
		a.dropAsync(pair.Elements[1])
	})
}
