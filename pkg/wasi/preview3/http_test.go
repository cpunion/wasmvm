//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
)

type httpAPI struct {
	t  *testing.T
	in *component.Instance
}

func (a httpAPI) function(path ...string) *component.Function {
	a.t.Helper()
	f, err := a.in.Function(path...)
	if err != nil {
		a.t.Fatal(err)
	}
	return f
}
func (a httpAPI) call(name string, args ...component.Value) component.Value {
	a.t.Helper()
	return a.invoke("types", name, args...)
}
func (a httpAPI) invoke(iface, name string, args ...component.Value) component.Value {
	a.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := a.function(iface, name).Call(ctx, args, component.CallOptions{})
	if err != nil {
		a.t.Fatal(name, err)
	}
	if r.Value == nil {
		return component.Value{}
	}
	return *r.Value
}
func (a httpAPI) source(typ component.ValueType, value component.Value) component.Value {
	a.t.Helper()
	v, err := a.in.NewSource(nil, typ, component.AsyncSource{Read: func(*component.Caller, uint32) (*component.ReadCompletion, error) {
		p := component.NewReadCompletion()
		p.Resolve(func(*component.Caller) (component.ReadResult, error) {
			return component.ReadResult{Values: []component.Value{value}, Closed: true}, nil
		})
		return p, nil
	}})
	if err != nil {
		a.t.Fatal(err)
	}
	return v
}
func (a httpAPI) read(v component.Value, limit uint32) component.ReadResult {
	a.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := a.in.ReadAsync(ctx, v.Resource, limit, component.CallOptions{})
	if err != nil {
		a.t.Fatal(err)
	}
	return r
}
func (a httpAPI) drop(v component.Value) {
	a.t.Helper()
	if _, err := a.in.DropResource(nil, v.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func (a httpAPI) dropAsync(v component.Value) {
	a.t.Helper()
	if _, err := a.in.DropAsync(nil, v.Resource, component.CallOptions{}); err != nil {
		a.t.Fatal(err)
	}
}
func (a httpAPI) message(kind, text string, trailing component.Value, options component.Value) (component.Value, component.Value) {
	a.t.Helper()
	f := a.function("types", "[static]"+kind+".new")
	_, bt := f.Info().Parameter(1)
	_, tt := f.Info().Parameter(2)
	body := component.Value{}
	if text != "" {
		pending := text
		v, err := a.in.NewSource(nil, bt.Element(), component.AsyncSource{Read: func(_ *component.Caller, limit uint32) (*component.ReadCompletion, error) {
			n := min(int(limit), len(pending))
			data := pending[:n]
			pending = pending[n:]
			p := component.NewReadCompletion()
			p.Resolve(func(*component.Caller) (component.ReadResult, error) {
				return component.ReadResult{Bytes: data, Closed: pending == ""}, nil
			})
			return p, nil
		}})
		if err != nil {
			a.t.Fatal(err)
		}
		body = httpSome(v)
	}
	trailers := a.source(tt, trailing)
	headers := a.call("[constructor]fields")
	args := []component.Value{headers, body, trailers}
	if kind == "request" {
		args = append(args, options)
	}
	r := a.call("[static]"+kind+".new", args...)
	return r.Elements[0], r.Elements[1]
}
func httpOwner(t *testing.T, cfg HTTPConfig) *HTTP {
	t.Helper()
	n, _ := socketsOwner(t, SocketsConfig{})
	p, err := NewHTTP(n, cfg)
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
func TestHTTPMessageMetadataAndConsume(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{})
		in := instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())
		a := httpAPI{t, in}
		for _, kind := range []string{"request", "response"} {
			message, tx := a.message(kind, "abcdef", httpOK(component.Value{}), component.Value{})
			prefix := "[method]" + kind + "."
			if kind == "request" {
				if a.call(prefix+"get-method", message).Case != 0 {
					t.Fatal("default method")
				}
				if a.call(prefix+"set-method", message, component.Value{Case: 2}).Case != 0 {
					t.Fatal("set method")
				}
				for _, name := range []string{"path-with-query", "authority"} {
					v := "/hello?x=1"
					if name == "authority" {
						v = "localhost:8080"
					}
					if a.call(prefix+"set-"+name, message, httpSome(component.Value{Text: v})).Case != 0 || a.call(prefix+"get-"+name, message).Elements[0].Text != v {
						t.Fatal(name)
					}
				}
			} else {
				if a.call(prefix+"get-status-code", message).Bits != 200 {
					t.Fatal("status")
				}
				if a.call(prefix+"set-status-code", message, component.Value{Bits: 99}).Case != 1 {
					t.Fatal("invalid status")
				}
				a.call(prefix+"set-status-code", message, component.Value{Bits: 201})
			}
			fields := a.call(prefix+"get-headers", message)
			if r := a.call("[method]fields.append", fields, component.Value{Text: "x"}, component.Value{Text: "v"}); r.Case != 1 || r.Elements[0].Case != 2 {
				t.Fatal("mutable message headers", r)
			}
			a.drop(fields)
			consume := a.function("types", "[static]"+kind+".consume-body")
			_, rt := consume.Info().Parameter(1)
			res := a.source(rt, component.Value{})
			pair := a.call("[static]"+kind+".consume-body", message, res)
			var data string
			for {
				r := a.read(pair.Elements[0], 2)
				data += r.Bytes
				if r.Closed {
					break
				}
			}
			if data != "abcdef" {
				t.Fatal(data)
			}
			a.dropAsync(pair.Elements[0])
			r := a.read(pair.Elements[1], 1)
			if len(r.Values) != 1 || r.Values[0].Case != 0 {
				t.Fatal(r)
			}
			a.dropAsync(pair.Elements[1])
			r = a.read(tx, 1)
			if len(r.Values) != 1 || r.Values[0].Case != 0 {
				t.Fatal("transmission", r)
			}
			a.dropAsync(tx)
		}
	})
}
func TestHTTPFieldsAndOptions(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{MaxHeaderFields: 3, MaxHeaderBytes: 64, ForbiddenHeaders: []string{"secret"}})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		f := a.call("[constructor]fields")
		for _, s := range []string{"one", "two"} {
			if r := a.call("[method]fields.append", f, component.Value{Text: "X-Test"}, component.Value{Text: s}); r.Case != 0 {
				t.Fatal(r)
			}
		}
		if r := a.call("[method]fields.get", f, component.Value{Text: "x-test"}); len(r.Elements) != 2 {
			t.Fatal(r)
		}
		for _, item := range []struct {
			name string
			code uint32
		}{{"bad name", 0}, {"secret", 1}} {
			r := a.call("[method]fields.append", f, component.Value{Text: item.name}, component.Value{Text: "v"})
			if r.Case != 1 || r.Elements[0].Case != item.code {
				t.Fatal(r)
			}
		}
		cloned := a.call("[method]fields.clone", f)
		r := a.call("[method]fields.get-and-delete", f, component.Value{Text: "x-test"})
		if r.Case != 0 || len(r.Elements[0].Elements) != 2 || a.call("[method]fields.has", f, component.Value{Text: "x-test"}).Bits != 0 {
			t.Fatal(r)
		}
		a.drop(f)
		a.drop(cloned)
		options := a.call("[constructor]request-options")
		for _, name := range []string{"connect", "first-byte", "between-bytes"} {
			a.call("[method]request-options.set-"+name+"-timeout", options, httpSome(component.Value{Bits: 123}))
			if r := a.call("[method]request-options.get-"+name+"-timeout", options); r.Case != 1 || r.Elements[0].Bits != 123 {
				t.Fatal(r)
			}
		}
		clone := a.call("[method]request-options.clone", options)
		a.drop(clone)
		message, tx := a.message("request", "", httpOK(component.Value{}), httpSome(options))
		child := a.call("[method]request.get-options", message).Elements[0]
		if r := a.call("[method]request-options.set-connect-timeout", child, component.Value{}); r.Case != 1 || r.Elements[0].Case != 1 {
			t.Fatal(r)
		}
		a.drop(child)
		a.drop(message)
		a.dropAsync(tx)
	})
}
func TestHTTPClientNetworkBodyAndTrailers(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		received := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			received <- string(data)
			w.Header().Set("Trailer", "X-End")
			w.Header().Set("X-Response", "yes")
			w.WriteHeader(202)
			io.WriteString(w, "response")
			w.Header().Set("X-End", "done")
		}))
		defer server.Close()
		p := httpOwner(t, HTTPConfig{AllowRequest: func(context.Context, *url.URL) error { return nil }, AllowConnect: socketsAllow})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		req, tx := a.message("request", "upload", httpOK(component.Value{}), component.Value{})
		a.call("[method]request.set-authority", req, httpSome(component.Value{Text: strings.TrimPrefix(server.URL, "http://")}))
		result := a.invoke("client", "send", req)
		if result.Case != 0 {
			t.Fatal(result)
		}
		response := result.Elements[0]
		if a.call("[method]response.get-status-code", response).Bits != 202 {
			t.Fatal("status")
		}
		f := a.function("types", "[static]response.consume-body")
		_, typ := f.Info().Parameter(1)
		pair := a.call("[static]response.consume-body", response, a.source(typ, component.Value{}))
		var text string
		for {
			r := a.read(pair.Elements[0], 3)
			text += r.Bytes
			if r.Closed {
				break
			}
		}
		if text != "response" {
			t.Fatal(text)
		}
		a.dropAsync(pair.Elements[0])
		tr := a.read(pair.Elements[1], 1).Values[0]
		if tr.Case != 0 || tr.Elements[0].Case != 1 {
			t.Fatal(tr)
		}
		fields := tr.Elements[0].Elements[0]
		if v := a.call("[method]fields.get", fields, component.Value{Text: "x-end"}); len(v.Elements) != 1 || v.Elements[0].Text != "done" {
			t.Fatal(v)
		}
		a.drop(fields)
		a.dropAsync(pair.Elements[1])
		if r := a.read(tx, 1); r.Values[0].Case != 0 {
			t.Fatal(r)
		}
		a.dropAsync(tx)
		if data := <-received; data != "upload" {
			t.Fatal(data)
		}
	})
}

func TestHTTPActualAsyncCoreGuest(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.Copy(io.Discard, r.Body); w.WriteHeader(204) }))
		defer server.Close()
		for _, allowed := range []bool{true, false} {
			cfg := HTTPConfig{}
			if allowed {
				cfg.AllowConnect = socketsAllow
				cfg.AllowRequest = func(context.Context, *url.URL) error { return nil }
			}
			p := httpOwner(t, cfg)
			in := instantiate(t, compileFixture(t, backend, "http-guest"), p.Imports())
			r := call(t, in, "run", component.Value{Text: strings.TrimPrefix(server.URL, "http://")})
			want := uint64(204)
			if !allowed {
				want = 65536 + 15
			}
			if r.Bits != want {
				t.Fatal(allowed, r)
			}
		}
	})
}
func TestHTTPIncomingHandlerAndExternalForwarding(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{Handler: func(ctx context.Context, r *http.Request) (*http.Response, error) {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			if string(data) != "input" {
				t.Error(string(data))
			}
			return &http.Response{StatusCode: 203, Header: http.Header{"X-Reply": []string{"yes"}}, Body: io.NopCloser(strings.NewReader("output")), ContentLength: 6}, nil
		}})
		in := instantiate(t, compileFixture(t, backend, "http-handler-api"), p.Imports())
		r := httptest.NewRequest("POST", "http://localhost/path?q=1", strings.NewReader("input"))
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		response, err := p.CallIncoming(ctx, in, r, component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil || string(data) != "output" || response.StatusCode != 203 {
			t.Fatal(string(data), err, response.StatusCode)
		}
	})
}

func TestHTTPIncomingActualStreamingGuest(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{})
		in := instantiate(t, compileFixture(t, backend, "http-server"), p.Imports())
		for _, early := range []bool{false, false, true} {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			response, err := p.CallIncoming(ctx, in, httptest.NewRequest("GET", "http://localhost/path", nil), component.CallOptions{})
			if err != nil {
				cancel()
				t.Fatal(err)
			}
			if !early {
				data, err := io.ReadAll(response.Body)
				if err != nil || string(data) != "hello" {
					t.Fatal(string(data), err)
				}
			}
			response.Body.Close()
			cancel()
			if early {
				in.Close()
			}
		}
	})
}
func TestHTTPHandlerActualGuest(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		compiled := compileFixture(t, backend, "http-server")
		handler := HTTPHandler(func(context.Context) (*component.Instance, *HTTP, func(), error) {
			process, err := NewIO(IOConfig{MaxReadBytes: 3, MaxWriteBytes: 3})
			if err != nil {
				return nil, nil, nil, err
			}
			network, err := NewSockets(process, SocketsConfig{})
			if err != nil {
				process.Close()
				return nil, nil, nil, err
			}
			p, err := NewHTTP(network, HTTPConfig{})
			if err != nil {
				network.Close()
				process.Close()
				return nil, nil, nil, err
			}
			in, err := compiled.Instantiate(nil, p.Imports(), component.CallOptions{})
			cleanup := func() {
				p.Close()
				if in != nil {
					in.Close()
					in.WaitClosed(nil)
				}
				p.WaitClosed(nil)
				network.Close()
				network.WaitClosed(nil)
				process.Close()
				process.WaitClosed(nil)
			}
			return in, p, cleanup, err
		}, component.CallOptions{})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", "http://localhost/", nil))
		if recorder.Code != 200 || recorder.Body.String() != "hello" {
			t.Fatal(recorder.Code, recorder.Body.String())
		}
	})
}

func TestHTTPDuplexHandlerKeepsUploadAfterHeaders(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: r.Body, ContentLength: -1}, nil
		}})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		req, tx := a.message("request", "duplex-body", httpOK(component.Value{}), component.Value{})
		a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "localhost"}))
		consume := a.function("types", "[static]response.consume-body")
		_, typ := consume.Info().Parameter(1)
		send := a.function("handler", "handle")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var text string
		_, err := send.CallWithResult(ctx, []component.Value{req}, component.CallOptions{}, func(c *component.Caller, v *component.Value) error {
			if v.Case != 0 {
				t.Error(v)
				return errors.New("handler failed")
			}
			handled, err := c.NewSource(typ, component.AsyncSource{Read: func(*component.Caller, uint32) (*component.ReadCompletion, error) {
				completion := component.NewReadCompletion()
				completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
					return component.ReadResult{Values: []component.Value{{}}, Closed: true}, nil
				})
				return completion, nil
			}})
			if err != nil {
				return err
			}
			pair, err := c.Call(consume, []component.Value{v.Elements[0], handled})
			if err != nil {
				return err
			}
			for {
				r, err := c.ReadAsync(pair.Elements[0].Resource, 3)
				if err != nil {
					return err
				}
				text += r.Bytes
				if r.Closed {
					break
				}
			}
			if err = c.DropAsync(pair.Elements[0].Resource); err != nil {
				return err
			}
			r, err := c.ReadAsync(pair.Elements[1].Resource, 1)
			if err != nil {
				return err
			}
			if len(r.Values) != 1 || r.Values[0].Case != 0 {
				t.Error(r)
			}
			return c.DropAsync(pair.Elements[1].Resource)
		})
		if err != nil || text != "duplex-body" {
			t.Fatal(text, err)
		}
		if r := a.read(tx, 1); r.Values[0].Case != 0 {
			t.Fatal(r)
		}
		a.dropAsync(tx)
	})
}

func TestHTTPIncomingForwardedBodyOutlivesHandler(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: r.Body, ContentLength: -1}, nil
		}})
		in := instantiate(t, compileFixture(t, backend, "http-handler-api"), p.Imports())
		input := io.NopCloser(strings.NewReader("forwarded"))
		request := httptest.NewRequest("POST", "http://localhost/", input)
		request.ContentLength = -1
		response, err := p.CallIncoming(nil, in, request, component.CallOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		<-response.Body.(*httpCallBody).done
		p.mu.Lock()
		var source *httpBody
		for body := range p.bodies {
			body.mu.Lock()
			same := body.reader == input
			body.mu.Unlock()
			if same {
				source = body
				break
			}
		}
		p.mu.Unlock()
		if source == nil {
			t.Fatal("forwarded input lost its owner")
		}
		source.mu.Lock()
		revoked := source.closed
		source.mu.Unlock()
		if revoked {
			t.Fatal("forwarded input revoked before the response reader finished")
		}
		data, err := io.ReadAll(response.Body)
		if err != nil || string(data) != "forwarded" {
			t.Fatal(string(data), err)
		}
	})
}

func TestHTTPIncomingContextCancelsForwardedResponseAfterGuestExit(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		reader := newHTTPBlockedReader()
		p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, _ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: reader, ContentLength: -1}, nil
		}})
		in := instantiate(t, compileFixture(t, backend, "http-handler-api"), p.Imports())
		ctx, cancel := context.WithCancel(context.Background())
		response, err := p.CallIncoming(ctx, in, httptest.NewRequest("GET", "http://localhost/", nil), component.CallOptions{})
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		defer response.Body.Close()
		<-response.Body.(*httpCallBody).done
		cancel()
		done := make(chan error, 1)
		go func() { _, err := io.ReadAll(response.Body); done <- err }()
		select {
		case err := <-done:
			if err == nil {
				t.Fatal("cancelled response read succeeded")
			}
		case <-time.After(time.Second):
			t.Fatal("forwarded response ignored parent cancellation")
		}
	})
}
