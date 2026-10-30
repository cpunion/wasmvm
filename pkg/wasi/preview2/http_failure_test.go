//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHTTPAllErrorPayloadsAndResponseOutparam(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	set, err := a.in.Function("types", "[static]response-outparam.set")
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range httpErrorNames {
		reply := &httpReply{done: make(chan struct{})}
		value, err := p.publishExternal(context.Background(), a.in, &httpEntry{kind: httpResponseOutparam, reply: reply})
		if err != nil {
			t.Fatal(err)
		}
		code := httpErr(name)
		if code.Error() != name {
			t.Fatal(name, code.Error())
		}
		result, err := set.Call(nil, []component.Value{value, {Case: 1, Elements: []component.Value{code.value}}}, component.CallOptions{})
		if err != nil || result.Value != nil {
			t.Fatal(name, result, err)
		}
		<-reply.done
		var failure *httpError
		if !errors.As(reply.failure, &failure) || failure.value.Case != uint32(i) {
			t.Fatal(name, reply.failure)
		}
	}
	for _, mode := range []string{"empty", "length", "informational"} {
		reply := &httpReply{done: make(chan struct{})}
		out, err := p.publishExternal(context.Background(), a.in, &httpEntry{kind: httpResponseOutparam, reply: reply})
		if err != nil {
			t.Fatal(err)
		}
		fields := a.newFields()
		if mode == "length" {
			a.ok("fields", "append", fields, component.Value{Text: "content-length"}, component.Value{Text: "2"})
		}
		response := a.raw("[constructor]outgoing-response", fields)
		if mode == "informational" {
			a.ok("outgoing-response", "set-status-code", response, component.Value{Bits: 103})
		}
		if _, err := set.Call(nil, []component.Value{out, httpOK(response)}, component.CallOptions{}); err != nil {
			t.Fatal(mode, err)
		}
		<-reply.done
		if mode == "empty" {
			if reply.response == nil || reply.response.ContentLength != 0 {
				t.Fatal(reply)
			}
			reply.response.Body.Close()
		} else if reply.failure == nil {
			t.Fatal("accepted invalid final response", mode)
		}
	}
}
func TestHTTPPendingFutureCancellationAndParentPollable(t *testing.T) {
	entered := make(chan struct{})
	cancelled := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-r.Context().Done():
			close(cancelled)
		case <-release:
		}
	}))
	defer func() { close(release); server.Close() }()
	p, _, process := httpOwner(t, HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	future := a.future(a.newRequest(server.URL))
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	poll := a.method("future-incoming-response", "subscribe", future)
	if v := a.call("poll", "[method]pollable.ready", poll); v.Bits != 0 {
		t.Fatal(v)
	}
	if _, err := a.in.DropResource(nil, future.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	a.drop(poll)
	select {
	case <-cancelled:
	case <-time.After(3 * time.Second):
		t.Fatal("dropping future did not cancel network request")
	}
	process.poller.mu.Lock()
	process.poller.next = 1 << 32
	process.poller.mu.Unlock()
	// Poll publication failure must release its parent lease.
	reply := &httpFuture{changed: make(chan struct{})}
	value, err := p.publishExternal(context.Background(), a.in, &httpEntry{kind: httpFutureResponse, future: reply})
	if err != nil {
		t.Fatal(err)
	}
	f, err := a.in.Function("types", "[method]future-incoming-response.subscribe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Call(nil, []component.Value{value}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	a.drop(value)
}
func TestHTTPTLSRejectionConnectDeniedAndFirstByteTimeout(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	defer tlsServer.Close()
	p, _, _ := httpOwner(t, HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	future := a.future(a.newRequest(tlsServer.URL))
	result := a.await("future-incoming-response", future)
	if result.Case != 1 || result.Elements[0].Case != 13 {
		t.Fatal(result)
	}
	a.drop(future)
	q, _, _ := httpOwner(t, HTTPConfig{AllowRequest: httpLoopbackURL})
	b := httpAPI{networkAPI{t, httpInstance(t, q, "interpreter", "http-api")}}
	future = b.future(b.newRequest("http://127.0.0.1:1/"))
	result = b.await("future-incoming-response", future)
	if result.Case != 1 || result.Elements[0].Case != 4 {
		t.Fatal(result)
	}
	b.drop(future)
	entered := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer func() { close(release); server.Close() }()
	request := a.newRequest(server.URL)
	options := a.raw("[constructor]request-options")
	a.ok("request-options", "set-first-byte-timeout", options, httpSome(component.Value{Bits: 10000000}))
	value := a.call("outgoing", "handle", request, httpSome(options))
	if value.Case != 0 {
		t.Fatal(value)
	}
	future = value.Elements[0]
	result = a.await("future-incoming-response", future)
	if result.Case != 1 {
		t.Fatal("timeout succeeded", result)
	}
	a.drop(future)
}
func TestHTTPInvalidURIContentLengthAndTrailers(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{AllowRequest: httpLoopbackURL})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	for _, name := range []string{"path-with-query", "authority"} {
		if v := a.method("outgoing-request", "set-"+name, a.newRequest("http://localhost/"), httpSome(component.Value{Text: "/é"})); v.Case != 1 {
			t.Fatal(name, v)
		}
	}
	fields := a.newFields()
	a.ok("fields", "append", fields, component.Value{Text: "Content-Length"}, component.Value{Text: "bad"})
	request := a.raw("[constructor]outgoing-request", fields)
	a.ok("outgoing-request", "set-authority", request, httpSome(component.Value{Text: "localhost"}))
	if v := a.call("outgoing", "handle", request, component.Value{}); v.Case != 1 || v.Elements[0].Case != 35 {
		t.Fatal(v)
	}
	response := a.raw("[constructor]outgoing-response", a.newFields())
	body := a.ok("outgoing-response", "body", response)
	trailers := a.newFields()
	a.ok("fields", "append", trailers, component.Value{Text: "content-length"}, component.Value{Text: "0"})
	result := a.raw("[static]outgoing-body.finish", body, httpSome(trailers))
	if result.Case != 1 || result.Elements[0].Case != 35 {
		t.Fatal(result)
	}
	a.drop(response)
}

func TestHTTPRedirectEmptyAndDiscardedBodyTrailers(t *testing.T) {
	var redirected atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			w.Header().Set("Location", "/target")
			w.WriteHeader(302)
		case "/target":
			redirected.Store(true)
		case "/empty":
			w.WriteHeader(204)
		default:
			w.Header().Set("Trailer", "X-Completed")
			_, _ = w.Write([]byte(strings.Repeat("x", 32768)))
			w.Header().Set("X-Completed", "yes")
		}
	}))
	defer server.Close()
	p, _, _ := httpOwner(t, HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback, BodyBufferBytes: 257})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	for _, item := range []struct {
		path   string
		status uint64
	}{{"/redirect", 302}, {"/empty", 204}, {"/large", 200}} {
		future := a.future(a.newRequest(server.URL + item.path))
		result := a.await("future-incoming-response", future)
		if result.Case != 0 {
			t.Fatal(item.path, result)
		}
		response := result.Elements[0]
		a.drop(future)
		if v := a.method("incoming-response", "status", response); v.Bits != item.status {
			t.Fatal(item.path, v)
		}
		body := a.ok("incoming-response", "consume", response)
		a.drop(response)
		// Finish without reading must release producer backpressure and retain
		// the final trailers without retaining the discarded body bytes.
		trailers := a.raw("[static]incoming-body.finish", body)
		result = a.await("future-trailers", trailers)
		if result.Case != 0 {
			t.Fatal(item.path, result)
		}
		if fields := result.Elements[0]; fields.Case == 1 {
			value := fields.Elements[0]
			if item.path == "/large" {
				values := a.method("fields", "get", value, component.Value{Text: "x-completed"})
				if len(values.Elements) != 1 || values.Elements[0].Text != "yes" {
					t.Fatal(values)
				}
			}
			a.drop(value)
		} else if item.path == "/large" {
			t.Fatal("missing discarded-body trailers")
		}
		a.drop(trailers)
	}
	if redirected.Load() {
		t.Fatal("redirect was followed implicitly")
	}
}

func TestHTTPForeignInstancesLimitsAndNonHTTPErrors(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{MaxHeaderFields: 1})
	in := httpInstance(t, p, "interpreter", "http-api")
	api := httpAPI{networkAPI{t, in}}
	f, err := in.Function("types", "[static]fields.from-list")
	if err != nil {
		t.Fatal(err)
	}
	pair := component.Value{Elements: []component.Value{{Text: "x"}, {Text: "value"}}}
	if _, err := f.Call(nil, []component.Value{{Elements: []component.Value{pair, pair}}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	fields := api.newFields()
	f, err = in.Function("types", "[method]fields.set")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Call(nil, []component.Value{fields, {Text: "x"}, {Elements: []component.Value{{Text: "a"}, {Text: "b"}}}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	api.drop(fields)
	if _, err := p.CallIncoming(nil, nil, nil, component.CallOptions{}); err == nil {
		t.Fatal("nil incoming call accepted")
	}
	if _, err := p.CallIncoming(nil, in, &http.Request{}, component.CallOptions{}); !errors.Is(err, wasm.ErrInvalid) {
		t.Fatal(err)
	}
	q, _, _ := httpOwner(t, HTTPConfig{})
	foreign := httpInstance(t, q, "interpreter", "http")
	if _, err := p.CallIncoming(context.Background(), foreign, httptest.NewRequest("GET", "http://localhost", nil), component.CallOptions{}); err == nil {
		t.Fatal("foreign HTTP resource types accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.CallIncoming(ctx, in, httptest.NewRequest("GET", "http://localhost", nil), component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	process, err := NewIO(IOConfig{Stdout: failingOutput{cause: errors.New("unrelated IO failure")}})
	if err != nil {
		t.Fatal(err)
	}
	network, err := NewNetwork(process, NetworkConfig{})
	if err != nil {
		t.Fatal(err)
	}
	own, err := NewHTTP(network, HTTPConfig{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		own.Close()
		process.Close()
		network.Close()
		own.WaitClosed(nil)
		process.WaitClosed(nil)
		network.WaitClosed(nil)
	})
	a := httpAPI{networkAPI{t, httpInstance(t, own, "interpreter", "http-api")}}
	output := a.call("stdout", "get-stdout")
	result := a.call("streams", "[method]output-stream.blocking-write-and-flush", output, component.Value{Text: "x"})
	if result.Case != 1 || result.Elements[0].Case != 0 {
		t.Fatal(result)
	}
	cause := result.Elements[0].Elements[0]
	if v := a.raw("http-error-code", cause); v.Case != 0 {
		t.Fatal("unrelated error downcast", v)
	}
	a.drop(cause)
	a.drop(output)
}
