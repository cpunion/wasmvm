//go:build linux || darwin

package preview2

import (
	"context"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpReply struct {
	once     sync.Once
	done     chan struct{}
	response *http.Response
	body     *httpBody
	failure  error
}

func (r *httpReply) fail(err error) { r.once.Do(func() { r.failure = err; close(r.done) }) }
func (r *httpReply) set(response *http.Response) {
	r.once.Do(func() {
		r.response = response
		if reader, ok := response.Body.(*httpBodyReader); ok {
			r.body = reader.body
		}
		close(r.done)
	})
}
func (p *HTTP) setResponse(c *component.Caller, args []component.Value) error {
	outparam, err := p.take(c, args[0], httpResponseOutparam)
	if err != nil {
		return err
	}
	reply := outparam.reply
	if args[1].Case == 1 {
		reply.fail(&httpError{value: args[1].Elements[0]})
		return nil
	}
	response, err := p.take(c, args[1].Elements[0], httpOutgoingResponse)
	if err != nil {
		reply.fail(err)
		return err
	}
	message := response.message
	length, err := httpContentLength(message.headers)
	if err != nil {
		message.body.close()
		reply.fail(err)
		return nil
	}
	if message.status < 200 {
		message.body.close()
		reply.fail(httpErr("HTTP-protocol-error"))
		return nil
	}
	result := &http.Response{StatusCode: int(message.status), Header: httpHeader(message.headers), ContentLength: -1, Body: http.NoBody}
	if message.bodyTaken {
		trailers := http.Header{}
		result.Trailer = trailers
		result.Body = &httpBodyReader{body: message.body, ctx: message.body.ctx, trailers: trailers}
		if length != nil {
			result.ContentLength = int64(*length)
		}
	} else {
		if length != nil && *length != 0 {
			message.body.close()
			reply.fail(httpErr("HTTP-response-body-size"))
			return nil
		}
		result.ContentLength = 0
		message.body.finish(nil, nil)
		message.body.close()
	}
	reply.set(result)
	return nil
}

// publishExternal reserves before acquiring the Store. No provider lock is
// held while Instance.NewResource waits for an idle Store.
func (p *HTTP) publishExternal(ctx context.Context, in *component.Instance, e *httpEntry) (component.Value, error) {
	p.mu.Lock()
	if p.closed || uint64(len(p.entries)) >= uint64(p.config.MaxResources) || p.next > 1<<32-1 {
		p.mu.Unlock()
		return component.Value{}, wasm.ErrLimit
	}
	e.owner = p
	e.id = uint32(p.next)
	p.next++
	p.entries[e.id] = e
	p.mu.Unlock()
	value, err := in.NewResource(ctx, p.types[e.kind], e.id)
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if err != nil {
		p.drop(e.id)
		return component.Value{}, err
	}
	if closed {
		in.DropResource(context.Background(), value.Resource, component.CallOptions{})
		return component.Value{}, wasm.ErrClosed
	}
	return value, nil
}

// CallIncoming invokes the standard guest export on an idle instance. The
// caller must consume/close the returned Body and close the request-scoped
// provider after guest execution. Handler below supplies that ownership pattern.
func (p *HTTP) CallIncoming(ctx context.Context, in *component.Instance, request *http.Request, options component.CallOptions) (*http.Response, error) {
	if in == nil || request == nil {
		return nil, vm.ErrReference
	}
	if request.URL == nil {
		return nil, wasm.ErrInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := p.begin(ctx); err != nil {
		return nil, err
	}
	defer p.active.Done()
	callCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	stop := context.AfterFunc(p.ctx, cancel)
	handedOff := false
	defer func() {
		if !handedOff {
			stop()
			cancel()
		}
	}()
	function, err := in.Function(HTTPIncomingInterface, "handle")
	if err != nil {
		return nil, err
	}
	fields, err := p.fieldsFromHeader(request.Header)
	if err != nil {
		return nil, err
	}
	body, err := p.newBody(false)
	if err != nil {
		return nil, err
	}
	source := request.Body
	if source == nil {
		source = http.NoBody
	}
	if err = body.pump(source, request.Trailer, p.config.Timeout, p.config.Timeout); err != nil {
		body.close()
		return nil, err
	}
	authority := request.Host
	if authority == "" {
		authority = request.URL.Host
	}
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	message := &httpMessage{headers: fields, method: request.Method, path: httpOptionalString{request.URL.RequestURI(), true}, scheme: httpOptionalString{scheme, true}, authority: httpOptionalString{authority, true}, body: body}
	input, err := p.publishExternal(callCtx, in, &httpEntry{kind: httpIncomingRequest, message: message})
	if err != nil {
		body.close()
		return nil, err
	}
	reply := &httpReply{done: make(chan struct{})}
	output, err := p.publishExternal(callCtx, in, &httpEntry{kind: httpResponseOutparam, reply: reply})
	if err != nil {
		in.DropResource(ctx, input.Resource, component.CallOptions{})
		return nil, err
	}
	if err = p.startWorker(); err != nil {
		stop()
		cancel()
		in.DropResource(ctx, input.Resource, component.CallOptions{})
		in.DropResource(ctx, output.Resource, component.CallOptions{})
		return nil, err
	}
	go func() {
		defer p.workers.Done()
		_, err := function.Call(callCtx, []component.Value{input, output}, options)
		if err != nil {
			reply.fail(err)
		} else {
			reply.fail(httpErr("HTTP-response-incomplete"))
		}
		<-reply.done
		if reply.body != nil {
			reply.body.mu.Lock()
			complete := reply.body.complete
			reply.body.mu.Unlock()
			if !complete {
				reply.body.finish(httpErr("HTTP-response-incomplete"), nil)
			}
		}
	}()
	select {
	case <-reply.done:
		if reply.failure != nil {
			stop()
			cancel()
			return nil, reply.failure
		}
		if reader, ok := reply.response.Body.(*httpBodyReader); ok {
			reader.ctx = callCtx
		}
		reply.response.Body = &httpCallBody{ReadCloser: reply.response.Body, cancel: cancel, stop: stop}
		handedOff = true
		return reply.response, nil
	case <-callCtx.Done():
		return nil, callCtx.Err()
	case <-ctx.Done():
		stop()
		cancel()
		return nil, ctx.Err()
	case <-p.ctx.Done():
		stop()
		cancel()
		return nil, wasm.ErrClosed
	}
}

// HTTPInstanceFactory supplies a fresh Store and request-scoped owners. Cleanup
// must close/wait HTTP, IO and network owners and then close the instance.
type HTTPInstanceFactory func(context.Context) (*component.Instance, *HTTP, func(), error)

// HTTPHandler runs a proxy component with an independent Store per request.
// It streams the response concurrently with the guest to preserve backpressure.
func HTTPHandler(factory HTTPInstanceFactory, options component.CallOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if factory == nil {
			http.Error(w, "HTTP component unavailable", 503)
			return
		}
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		in, provider, cleanup, err := factory(ctx)
		if err != nil {
			if cleanup != nil {
				cleanup()
			}
			http.Error(w, "HTTP component unavailable", 503)
			return
		}
		if cleanup == nil || provider == nil || in == nil {
			if cleanup != nil {
				cleanup()
			}
			http.Error(w, "invalid HTTP component factory", 500)
			return
		}
		defer cleanup()
		defer cancel()
		// A server body may be waiting on a peer that stopped producing bytes.
		// Apply a deadline before closing it during provider cancellation.
		controller := http.NewResponseController(w)
		copy := r.Clone(ctx)
		copy.Body = &httpServerBody{ReadCloser: r.Body, interrupt: func() { _ = controller.SetReadDeadline(time.Now()) }}
		response, err := provider.CallIncoming(ctx, in, copy, options)
		if err != nil {
			provider.Close()
			http.Error(w, "HTTP component failed", 502)
			return
		}
		defer response.Body.Close()
		for name, values := range response.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(response.StatusCode)
		_, err = io.Copy(w, response.Body)
		if err != nil {
			panic(http.ErrAbortHandler)
		}
		if err == nil {
			for name, values := range response.Trailer {
				for _, value := range values {
					w.Header().Add(http.TrailerPrefix+name, value)
				}
			}
		}
	})
}

type httpCallBody struct {
	io.ReadCloser
	once   sync.Once
	cancel context.CancelFunc
	stop   func() bool
}

func (b *httpCallBody) Close() error {
	var err error
	b.once.Do(func() { b.stop(); b.cancel(); err = b.ReadCloser.Close() })
	return err
}

type httpServerBody struct {
	io.ReadCloser
	interrupt func()
	complete  atomic.Bool
}

func (b *httpServerBody) Read(out []byte) (int, error) {
	n, err := b.ReadCloser.Read(out)
	if err != nil {
		b.complete.Store(true)
	}
	return n, err
}

func (b *httpServerBody) Close() error {
	if b.interrupt != nil && !b.complete.Load() {
		b.interrupt()
	}
	if b.ReadCloser == nil {
		return nil
	}
	return b.ReadCloser.Close()
}
