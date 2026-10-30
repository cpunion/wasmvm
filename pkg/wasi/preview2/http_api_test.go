//go:build linux || darwin

package preview2

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpAPI struct{ networkAPI }

func (a httpAPI) raw(name string, args ...component.Value) component.Value {
	a.t.Helper()
	return a.call("types", name, args...)
}
func (a httpAPI) method(resource, name string, self component.Value, args ...component.Value) component.Value {
	a.t.Helper()
	return a.raw("[method]"+resource+"."+name, append([]component.Value{self}, args...)...)
}
func (a httpAPI) ok(resource, name string, self component.Value, args ...component.Value) component.Value {
	a.t.Helper()
	v := a.method(resource, name, self, args...)
	if v.Case != 0 {
		a.t.Fatal(resource, name, v)
	}
	if len(v.Elements) == 0 {
		return component.Value{}
	}
	return v.Elements[0]
}
func (a httpAPI) newFields() component.Value { return a.raw("[constructor]fields") }
func (a httpAPI) newRequest(target string) component.Value {
	a.t.Helper()
	u, err := url.Parse(target)
	if err != nil {
		a.t.Fatal(err)
	}
	request := a.raw("[constructor]outgoing-request", a.newFields())
	a.ok("outgoing-request", "set-authority", request, httpSome(component.Value{Text: u.Host}))
	a.ok("outgoing-request", "set-path-with-query", request, httpSome(component.Value{Text: u.RequestURI()}))
	scheme := component.Value{}
	if u.Scheme == "https" {
		scheme.Case = 1
	}
	a.ok("outgoing-request", "set-scheme", request, httpSome(scheme))
	return request
}
func (a httpAPI) future(request component.Value) component.Value {
	a.t.Helper()
	v := a.call("outgoing", "handle", request, component.Value{})
	if v.Case != 0 {
		a.t.Fatal(v)
	}
	return v.Elements[0]
}
func (a httpAPI) await(resource string, future component.Value) component.Value {
	a.t.Helper()
	// LLGo's pinned netpoll may observe cancellation on its one-second
	// fallback poll tick. This wait checks eventual completion, not latency.
	deadline := time.Now().Add(3 * time.Second)
	for {
		v := a.method(resource, "get", future)
		if v.Case == 1 {
			if v.Elements[0].Case != 0 {
				a.t.Fatal(v)
			}
			return v.Elements[0].Elements[0]
		}
		if time.Now().After(deadline) {
			a.t.Fatal("future remained pending")
		}
		time.Sleep(time.Millisecond)
	}
}
func (a httpAPI) trap(name string, args ...component.Value) {
	a.t.Helper()
	f, err := a.in.Function("types", name)
	if err != nil {
		a.t.Fatal(err)
	}
	if _, err := f.Call(nil, args, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		a.t.Fatal(name, err)
	}
}
func TestHTTPResourceMethodsOwnershipAndOptions(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{ForbiddenHeaders: []string{"x-forbidden"}})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	fields := a.newFields()
	a.ok("fields", "set", fields, component.Value{Text: "x"}, component.Value{Elements: []component.Value{{Text: "one"}, {Text: ""}}})
	if v := a.method("fields", "get", fields, component.Value{Text: "X"}); len(v.Elements) != 2 {
		t.Fatal(v)
	}
	if v := a.method("fields", "append", fields, component.Value{Text: "x-forbidden"}, component.Value{Text: "no"}); v.Case != 1 || v.Elements[0].Case != 1 {
		t.Fatal(v)
	}
	copy := a.method("fields", "clone", fields)
	a.ok("fields", "delete", copy, component.Value{Text: "X"})
	a.drop(copy)
	bad := a.raw("[static]fields.from-list", component.Value{Elements: []component.Value{{Elements: []component.Value{{Text: "bad name"}, {Text: "x"}}}}})
	if bad.Case != 1 || bad.Elements[0].Case != 0 {
		t.Fatal(bad)
	}
	request := a.raw("[constructor]outgoing-request", fields)
	if _, err := a.in.ResourceRep(nil, p.types[httpFieldsKind], fields.Resource); err == nil {
		t.Fatal("owned fields remained usable")
	}
	for _, name := range []string{"path-with-query", "authority", "scheme"} {
		if v := a.method("outgoing-request", name, request); v.Case != 0 {
			t.Fatal(name, v)
		}
	}
	for _, item := range []struct {
		name  string
		value component.Value
	}{{"method", component.Value{Case: 9, Elements: []component.Value{{Text: "bad method"}}}}, {"scheme", httpSome(component.Value{Case: 2, Elements: []component.Value{{Text: "1bad"}}})}, {"authority", httpSome(component.Value{Text: "bad host /"})}, {"path-with-query", httpSome(component.Value{Text: "#fragment"})}} {
		if v := a.method("outgoing-request", "set-"+item.name, request, item.value); v.Case != 1 {
			t.Fatal(item.name, v)
		}
	}
	a.ok("outgoing-request", "set-method", request, component.Value{Case: 9, Elements: []component.Value{{Text: "CUSTOM"}}})
	if v := a.method("outgoing-request", "method", request); v.Case != 9 || v.Elements[0].Text != "CUSTOM" {
		t.Fatal(v)
	}
	a.ok("outgoing-request", "set-scheme", request, httpSome(component.Value{Case: 2, Elements: []component.Value{{Text: "custom+ext"}}}))
	if v := a.method("outgoing-request", "scheme", request); v.Elements[0].Case != 2 {
		t.Fatal(v)
	}
	headers := a.method("outgoing-request", "headers", request)
	if _, err := a.in.DropResource(nil, request.Resource, component.CallOptions{}); !errors.Is(err, wasm.ErrTrap) {
		t.Fatal(err)
	}
	a.drop(headers)

}
func TestHTTPOptionsBodySizeAndChildren(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	options := a.raw("[constructor]request-options")
	for _, name := range []string{"connect-timeout", "first-byte-timeout", "between-bytes-timeout"} {
		if v := a.method("request-options", name, options); v.Case != 0 {
			t.Fatal(v)
		}
		if v := a.method("request-options", "set-"+name, options, httpSome(component.Value{Bits: 1 << 63})); v.Case != 1 {
			t.Fatal(v)
		}
		a.ok("request-options", "set-"+name, options, httpSome(component.Value{Bits: 1234}))
		if v := a.method("request-options", name, options); v.Case != 1 || v.Elements[0].Bits != 1234 {
			t.Fatal(v)
		}
		a.ok("request-options", "set-"+name, options, component.Value{})
	}
	a.drop(options)
	fields := a.newFields()
	a.ok("fields", "append", fields, component.Value{Text: "Content-Length"}, component.Value{Text: "3"})
	response := a.raw("[constructor]outgoing-response", fields)
	for _, code := range []uint64{0, 1000} {
		if v := a.method("outgoing-response", "set-status-code", response, component.Value{Bits: code}); v.Case != 1 {
			t.Fatal(v)
		}
	}
	body := a.ok("outgoing-response", "body", response)
	if v := a.method("outgoing-response", "body", response); v.Case != 1 {
		t.Fatal(v)
	}
	output := a.ok("outgoing-body", "write", body)
	if v := a.method("outgoing-body", "write", body); v.Case != 1 {
		t.Fatal(v)
	}
	// A live stream prevents finishing its parent. Use a separate parent below
	// because the failed owned transfer consumes the capability.
	a.trap("[static]outgoing-body.finish", body, component.Value{})
	a.drop(output)
	a.drop(response)
	fields = a.newFields()
	a.ok("fields", "append", fields, component.Value{Text: "content-length"}, component.Value{Text: "3"})
	response = a.raw("[constructor]outgoing-response", fields)
	body = a.ok("outgoing-response", "body", response)
	result := a.raw("[static]outgoing-body.finish", body, component.Value{})
	if result.Case != 1 || result.Elements[0].Case != 28 {
		t.Fatal(result)
	}
	a.drop(response)
}
func TestHTTPRealFailuresFuturesAndErrorDowncast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/short" {
			w.Header().Set("Content-Length", "10")
			w.Write([]byte("xx"))
			return
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	p, _, _ := httpOwner(t, HTTPConfig{AllowRequest: httpLoopbackURL, AllowConnect: tcpAllowLoopback})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	future := a.future(a.newRequest(server.URL + "/short"))
	result := a.await("future-incoming-response", future)
	if result.Case != 0 {
		t.Fatal(result)
	}
	response := result.Elements[0]
	again := a.method("future-incoming-response", "get", future)
	if again.Case != 1 || again.Elements[0].Case != 1 {
		t.Fatal(again)
	}
	a.drop(future)
	body := a.ok("incoming-response", "consume", response)
	if v := a.method("incoming-response", "consume", response); v.Case != 1 {
		t.Fatal(v)
	}
	a.drop(response)
	input := a.ok("incoming-body", "stream", body)
	if v := a.method("incoming-body", "stream", body); v.Case != 1 {
		t.Fatal(v)
	}
	var cause component.Value
	for i := 0; i < 100; i++ {
		v := a.call("streams", "[method]input-stream.read", input, component.Value{Bits: 16})
		if v.Case == 1 && v.Elements[0].Case == 0 {
			cause = v.Elements[0].Elements[0]
			break
		}
		time.Sleep(time.Millisecond)
	}
	if cause.Resource == nil {
		t.Fatal("missing HTTP IO failure")
	}
	code := a.raw("http-error-code", cause)
	if code.Case != 1 || code.Elements[0].Case != 25 {
		t.Fatal(code)
	}
	a.drop(cause)
	a.drop(input)
	trailers := a.raw("[static]incoming-body.finish", body)
	result = a.await("future-trailers", trailers)
	if result.Case != 1 || result.Elements[0].Case != 25 {
		t.Fatal(result)
	}
	a.drop(trailers)
	p.mu.Lock()
	left := len(p.entries)
	p.mu.Unlock()
	if left != 0 {
		t.Fatal("failure leaked resources", left)
	}
}
func TestHTTPDefaultDenyMalformedTargetsAndProviderClose(t *testing.T) {
	p, _, _ := httpOwner(t, HTTPConfig{})
	a := httpAPI{networkAPI{t, httpInstance(t, p, "interpreter", "http-api")}}
	request := a.raw("[constructor]outgoing-request", a.newFields())
	v := a.call("outgoing", "handle", request, component.Value{})
	if v.Case != 1 || v.Elements[0].Case != 19 {
		t.Fatal(v)
	}
	request = a.newRequest("http://127.0.0.1/echo")
	v = a.call("outgoing", "handle", request, component.Value{})
	if v.Case != 1 || v.Elements[0].Case != 15 {
		t.Fatal(v)
	}
	p.mu.Lock()
	entries, bodies := len(p.entries), len(p.bodies)
	p.mu.Unlock()
	if entries != 0 || bodies != 0 {
		t.Fatal(entries, bodies)
	}
	p.Close()
	f, err := a.in.Function("types", "[constructor]fields")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Call(nil, nil, component.CallOptions{}); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
}
