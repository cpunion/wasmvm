//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func TestHTTPResourceQuotasAndMessageRollback(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{MaxResources: 1})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		fields := a.call("[constructor]fields")
		if _, err := a.function("types", "[constructor]fields").Call(nil, nil, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		a.drop(fields)
		p2 := httpOwner(t, HTTPConfig{MaxMessages: 1})
		a2 := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p2.Imports())}
		req, tx := a2.message("request", "", httpOK(component.Value{}), component.Value{})
		f := a2.function("types", "[static]request.new")
		_, tt := f.Info().Parameter(2)
		trailers := a2.source(tt, httpOK(component.Value{}))
		headers := a2.call("[constructor]fields")
		if _, err := f.Call(nil, []component.Value{headers, {}, trailers, {}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal(err)
		}
		a2.drop(req)
		a2.dropAsync(tx)
		config := component.DefaultConfig()
		config.Core.Backend = backend
		config.MaxResources = 6
		engine, err := component.NewEngine(config)
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		data := fixtureBytes(t, "http-api")
		m, err := engine.Compile(nil, data)
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		owner := httpOwner(t, HTTPConfig{})
		bounded := httpAPI{t, instantiate(t, m, owner.Imports())}
		f = bounded.function("types", "[static]request.new")
		_, tt = f.Info().Parameter(2)
		headers = bounded.call("[constructor]fields")
		trailers = bounded.source(tt, httpOK(component.Value{}))
		if _, err := f.Call(nil, []component.Value{headers, {}, trailers, {}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
			t.Fatal("canonical publication quota", err)
		}
		owner.mu.Lock()
		remaining := len(owner.entries) + len(owner.bodies)
		owner.mu.Unlock()
		if remaining != 0 {
			t.Fatal("failed message retained ownership", remaining)
		}
	})
}
func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name + ".wasm")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type httpFaultReader struct {
	mode   string
	once   atomic.Bool
	closed atomic.Bool
}

func (r *httpFaultReader) Read(data []byte) (int, error) {
	switch r.mode {
	case "panic":
		panic("reader")
	case "bad-count":
		return len(data) + 1, nil
	case "no-progress":
		return 0, nil
	case "truncated":
		if !r.once.Swap(true) {
			data[0] = 'x'
			return 1, io.ErrUnexpectedEOF
		}
		return 0, io.EOF
	default:
		return 0, io.EOF
	}
}
func (r *httpFaultReader) Close() error {
	r.closed.Store(true)
	if r.mode == "close-panic" {
		panic("close")
	}
	return nil
}
func TestHTTPExternalBodyFailureProgressAndClose(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, mode := range []string{"panic", "bad-count", "no-progress", "truncated", "close-panic", "length"} {
			t.Run(mode, func(t *testing.T) {
				reader := &httpFaultReader{mode: mode}
				p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
					io.Copy(io.Discard, r.Body)
					length := int64(-1)
					if mode == "length" {
						length = 3
					}
					return &http.Response{StatusCode: 200, Body: reader, ContentLength: length}, nil
				}})
				a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
				req, tx := a.message("request", "", httpOK(component.Value{}), component.Value{})
				a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "localhost"}))
				response := a.invoke("handler", "handle", req).Elements[0]
				f := a.function("types", "[static]response.consume-body")
				_, typ := f.Info().Parameter(1)
				pair := a.call("[static]response.consume-body", response, a.source(typ, component.Value{}))
				r := a.read(pair.Elements[0], 1)
				if !r.Closed || mode == "truncated" && r.Bytes != "x" {
					t.Fatal(r)
				}
				a.dropAsync(pair.Elements[0])
				result := a.read(pair.Elements[1], 1)
				want := "internal-error"
				switch mode {
				case "no-progress":
					want = "HTTP-protocol-error"
				case "truncated":
					want = "HTTP-response-incomplete"
				case "length":
					want = "HTTP-response-body-size"
				}
				httpCheckError(t, result.Values[0], want)
				a.dropAsync(pair.Elements[1])
				a.dropAsync(tx)
				if !reader.closed.Load() {
					t.Fatal("reader was not closed")
				}
			})
		}
	})
}
func TestHTTPIncomingCancellationAndFactoryError(t *testing.T) {
	entered := make(chan struct{})
	p := httpOwner(t, HTTPConfig{Handler: func(ctx context.Context, _ *http.Request) (*http.Response, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	}})
	in := instantiate(t, compileFixture(t, "interpreter", "http-handler-api"), p.Imports())
	ctx, cancel := context.WithCancel(context.Background())
	go func() { <-entered; cancel() }()
	if _, err := p.CallIncoming(ctx, in, httptest.NewRequest("GET", "http://localhost", nil), component.CallOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	failed := httpOwner(t, HTTPConfig{})
	other := instantiate(t, compileFixture(t, "interpreter", "http-handler-api"), failed.Imports())
	cleaned := false
	handler := HTTPHandler(func(context.Context) (*component.Instance, *HTTP, func(), error) {
		return other, failed, func() { cleaned = true }, nil
	}, component.CallOptions{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "http://localhost", nil))
	if recorder.Code != 502 || !cleaned {
		t.Fatal(recorder.Code, cleaned)
	}
}
func TestHTTPHandlerReadsExternalRequest(t *testing.T) {
	p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 201, Body: io.NopCloser(strings.NewReader(string(data))), ContentLength: int64(len(data))}, nil
	}})
	in := instantiate(t, compileFixture(t, "interpreter", "http-handler-api"), p.Imports())
	handler := HTTPHandler(func(context.Context) (*component.Instance, *HTTP, func(), error) { return in, p, func() {}, nil }, component.CallOptions{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("POST", "http://localhost", strings.NewReader("input")))
	if recorder.Code != 201 || recorder.Body.String() != "input" {
		t.Fatal(recorder.Code, recorder.Body.String())
	}
}
func TestHTTPLateResponseAndRequestLimit(t *testing.T) {
	reply := &httpReply{done: make(chan struct{})}
	reply.abandon()
	done := make(chan struct{})
	close(done)
	ctx, cancel := context.WithCancel(context.Background())
	closed := false
	body := &httpCallBody{ReadCloser: http.NoBody, done: done, release: func() { closed = true }, cancel: cancel, stop: func() bool { return true }}
	reply.set(&http.Response{Body: body}, nil)
	if !closed || ctx.Err() == nil {
		t.Fatal("late response leaked")
	}
	p := httpOwner(t, HTTPConfig{MaxRequests: 1})
	if err := p.reserveRequest(); err != nil {
		t.Fatal(err)
	}
	if err := p.reserveRequest(); httpErrorName(httpErrorResult(err)) != "connection-limit-reached" {
		t.Fatal(err)
	}
	p.releaseRequest()
	p.Close()
	if err := p.reserveRequest(); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	timerCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := p.WaitClosed(timerCtx); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPAwaitingTrailersDuringProviderClose(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		message, tx := a.message("response", "pending", httpOK(component.Value{}), component.Value{})
		f := a.function("types", "[static]response.consume-body")
		_, typ := f.Info().Parameter(1)
		pair := a.call("[static]response.consume-body", message, a.source(typ, component.Value{}))
		closed := make(chan struct{})
		go func() {
			defer close(closed)
			deadline := time.NewTimer(time.Second)
			defer deadline.Stop()
			tick := time.NewTicker(time.Millisecond)
			defer tick.Stop()
			for {
				p.network.io.mu.Lock()
				waiting := p.network.io.activeWorkers != 0
				p.network.io.mu.Unlock()
				if waiting {
					p.Close()
					return
				}
				select {
				case <-tick.C:
				case <-deadline.C:
					p.Close()
					return
				}
			}
		}()
		r := a.read(pair.Elements[1], 1)
		if len(r.Values) != 1 {
			t.Fatal(r)
		}
		httpCheckError(t, r.Values[0], "connection-terminated")
		<-closed
		a.dropAsync(pair.Elements[0])
		a.dropAsync(pair.Elements[1])
		a.dropAsync(tx)
	})
}
func TestHTTPEmptyResponseReadinessAndPendingTransmission(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		p := httpOwner(t, HTTPConfig{Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
			io.Copy(io.Discard, r.Body)
			return &http.Response{StatusCode: 204}, nil
		}})
		a := httpAPI{t, instantiate(t, compileFixture(t, backend, "http-api"), p.Imports())}
		message, tx := a.message("request", "", httpOK(component.Value{}), component.Value{})
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		_, err := a.in.ReadAsync(ctx, tx.Resource, 1, component.CallOptions{})
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		a.dropAsync(tx)
		a.call("[method]request.set-authority", message, httpSome(component.Value{Text: "localhost"}))
		response := a.invoke("handler", "handle", message).Elements[0]
		f := a.function("types", "[static]response.consume-body")
		_, typ := f.Info().Parameter(1)
		pair := a.call("[static]response.consume-body", response, a.source(typ, component.Value{}))
		if r := a.read(pair.Elements[0], 0); r.Closed || r.Count != 0 {
			t.Fatal(r)
		}
		if r := a.read(pair.Elements[0], 1); !r.Closed {
			t.Fatal(r)
		}
		r := a.read(pair.Elements[1], 1)
		if len(r.Values) != 1 || r.Values[0].Case != 0 || r.Values[0].Elements[0].Case != 0 {
			t.Fatal(r)
		}
		a.dropAsync(pair.Elements[0])
		a.dropAsync(pair.Elements[1])
	})
}

type httpSlowClose struct {
	entered, allow chan struct{}
	once           atomic.Bool
}

func (r *httpSlowClose) Read([]byte) (int, error) { return 0, io.EOF }
func (r *httpSlowClose) Close() error {
	if !r.once.Swap(true) {
		close(r.entered)
		<-r.allow
	}
	return nil
}
func TestHTTPPhysicalMessageQuotaUntilCloseCompletes(t *testing.T) {
	reader := &httpSlowClose{entered: make(chan struct{}), allow: make(chan struct{})}
	defer close(reader.allow)
	p := httpOwner(t, HTTPConfig{MaxMessages: 2, Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
		io.Copy(io.Discard, r.Body)
		return &http.Response{StatusCode: 200, Body: reader, ContentLength: -1}, nil
	}})
	a := httpAPI{t, instantiate(t, compileFixture(t, "interpreter", "http-api"), p.Imports())}
	req, tx := a.message("request", "", httpOK(component.Value{}), component.Value{})
	a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "localhost"}))
	response := a.invoke("handler", "handle", req).Elements[0]
	a.read(tx, 1)
	a.dropAsync(tx)
	a.drop(response)
	<-reader.entered
	second, transmission := a.message("request", "", httpOK(component.Value{}), component.Value{})
	f := a.function("types", "[static]request.new")
	_, typ := f.Info().Parameter(2)
	trailers := a.source(typ, httpOK(component.Value{}))
	headers := a.call("[constructor]fields")
	if _, err := f.Call(nil, []component.Value{headers, {}, trailers, {}}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal("closing physical message escaped quota", err)
	}
	a.drop(second)
	a.dropAsync(transmission)
	p.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := p.WaitClosed(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("shutdown ignored physical close", err)
	}
}

func TestHTTPUnpublishedReaderPanicAndTrailerQuota(t *testing.T) {
	for _, mode := range []string{"header", "trailer"} {
		p := httpOwner(t, HTTPConfig{MaxHeaderFields: 1, Handler: func(_ context.Context, r *http.Request) (*http.Response, error) {
			io.Copy(io.Discard, r.Body)
			response := &http.Response{StatusCode: 200, Body: &httpFaultReader{mode: "close-panic"}, ContentLength: -1}
			if mode == "header" {
				response.Header = http.Header{"bad header": []string{"x"}}
			} else {
				response.Body = http.NoBody
				response.Trailer = http.Header{"a": []string{"x"}, "b": []string{"y"}}
			}
			return response, nil
		}})
		a := httpAPI{t, instantiate(t, compileFixture(t, "interpreter", "http-api"), p.Imports())}
		req, tx := a.message("request", "", httpOK(component.Value{}), component.Value{})
		a.call("[method]request.set-authority", req, httpSome(component.Value{Text: "localhost"}))
		result := a.invoke("handler", "handle", req)
		if mode == "header" {
			httpCheckError(t, result, "internal-error")
		} else {
			if result.Case != 0 {
				t.Fatal(result)
			}
			f := a.function("types", "[static]response.consume-body")
			_, typ := f.Info().Parameter(1)
			pair := a.call("[static]response.consume-body", result.Elements[0], a.source(typ, component.Value{}))
			a.read(pair.Elements[0], 1)
			httpCheckError(t, a.read(pair.Elements[1], 1).Values[0], "HTTP-response-trailer-section-size")
			a.dropAsync(pair.Elements[0])
			a.dropAsync(pair.Elements[1])
		}
		a.dropAsync(tx)
	}
}

func TestHTTPConsumeBodyPublicationAndTaskRollback(t *testing.T) {
	backends(t, func(t *testing.T, backend string) {
		for _, limit := range []uint32{9, 11, 13} {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			config.MaxResources = limit
			if limit == 13 {
				config.MaxTasks = 1
			}
			engine, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer engine.Close()
			m, err := engine.Compile(nil, fixtureBytes(t, "http-api"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			p := httpOwner(t, HTTPConfig{})
			a := httpAPI{t, instantiate(t, m, p.Imports())}
			f := a.function("types", "[static]response.consume-body")
			_, typ := f.Info().Parameter(1)
			handled := a.source(typ, component.Value{})
			response, tx := a.message("response", "", httpOK(component.Value{}), component.Value{})
			if _, err := f.Call(nil, []component.Value{response, handled}, component.CallOptions{}); !errors.Is(err, wasm.ErrLimit) {
				t.Fatal("consume publication quota", limit, err)
			}
			a.dropAsync(tx)
			p.mu.Lock()
			left := len(p.entries) + len(p.bodies)
			p.mu.Unlock()
			if left != 0 {
				t.Fatal("consume rollback retained message", limit, left)
			}
		}
	})
}
