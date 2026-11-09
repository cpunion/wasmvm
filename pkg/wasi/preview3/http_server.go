//go:build linux || darwin

package preview3

import (
	"context"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpReply struct {
	mu               sync.Mutex
	ready, abandoned bool
	done             chan struct{}
	response         *http.Response
	failure          error
}

func (r *httpReply) set(response *http.Response, err error) {
	r.mu.Lock()
	if r.ready || r.abandoned {
		r.mu.Unlock()
		if response != nil {
			response.Body.(*httpCallBody).discard()
		}
		return
	}
	r.ready = true
	r.response = response
	r.failure = err
	close(r.done)
	r.mu.Unlock()
}
func (r *httpReply) abandon() {
	r.mu.Lock()
	r.abandoned = true
	response := r.response
	r.response = nil
	r.mu.Unlock()
	if response != nil {
		response.Body.(*httpCallBody).discard()
	}
}

// CallIncoming invokes the pinned async handler on an idle Store. Response
// reads drive no guest code on the network goroutine: the ordinary component
// scheduler supplies bounded batches while the invocation owns the Store.
// Close the returned Body before closing the request-scoped provider/instance.
func (p *HTTP) CallIncoming(ctx context.Context, in *component.Instance, request *http.Request, options component.CallOptions) (*http.Response, error) {
	return p.callIncoming(ctx, in, request, options, nil)
}

func (p *HTTP) callIncoming(ctx context.Context, in *component.Instance, request *http.Request, options component.CallOptions, collector *httpCollector) (*http.Response, error) {
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
	function, err := in.Function(HTTPHandlerInterface, "handle")
	if err != nil {
		return nil, err
	}
	if err = httpWITFunctions["handler"]["handle"].check(p.types)(function.Info()); err != nil {
		return nil, err
	}
	fields, err := p.fieldsFromHeader(request.Header)
	if err != nil {
		return nil, err
	}
	body, err := p.newBody()
	if err != nil {
		return nil, err
	}
	defer body.release()
	callCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	stop := context.AfterFunc(p.ctx, cancel)
	accepted := false
	defer func() {
		if !accepted {
			stop()
			cancel()
			body.dispose(nil, httpErr("connection-terminated"))
		}
	}()
	reader := request.Body
	if reader == nil {
		reader = http.NoBody
	}
	body.mu.Lock()
	body.reader = reader
	body.networkStop = context.AfterFunc(callCtx, func() { body.dispose(nil, callCtx.Err()) })
	body.readTimeout = p.config.Timeout
	body.sizeFailure = "HTTP-request-body-size"
	body.trailerSnapshot = func() (*host.HTTPFields, error) {
		if len(request.Trailer) == 0 {
			return nil, nil
		}
		fields, err := p.fieldsFromHeader(request.Trailer)
		if err != nil {
			return nil, httpErr("HTTP-request-trailer-section-size")
		}
		return fields, nil
	}
	if request.ContentLength >= 0 {
		n := uint64(request.ContentLength)
		body.expected = &n
	}
	body.mu.Unlock()
	if err = p.reserveRequest(); err != nil {
		return nil, err
	}
	reserved := true
	defer func() {
		if reserved {
			p.releaseRequest()
		}
	}()
	authority := request.Host
	if authority == "" {
		authority = request.URL.Host
	}
	scheme := request.URL.Scheme
	if scheme == "" {
		scheme = "http"
		if request.TLS != nil {
			scheme = "https"
		}
	}
	path := request.URL.RequestURI()
	reply := &httpReply{done: make(chan struct{})}
	callDone := make(chan struct{})
	run := func() {
		defer close(callDone)
		defer func() {
			if recover() != nil {
				body.dispose(nil, httpErr("internal-error"))
				reply.set(nil, httpErr("internal-error"))
				if collector != nil {
					collector.failure = httpErr("internal-error")
				}
				cancel()
			}
		}()
		defer p.releaseRequest()
		fuel, failed := function.CallWithResultPrepared(callCtx, options, func(c *component.Caller) ([]component.Value, error) {
			input, failed := p.publish(c, &httpEntry{kind: "request", message: &httpMessage{headers: fields, method: request.Method, path: &path, scheme: &scheme, authority: &authority, body: body}})
			return []component.Value{input}, failed
		}, func(c *component.Caller, v *component.Value) error {
			if v == nil {
				return vm.ErrType
			}
			if v.Case == 1 {
				reply.set(nil, &httpError{value: v.Elements[0]})
				return nil
			}
			entry, failed := p.take(c, v.Elements[0], "response", false)
			if failed != nil {
				return failed
			}
			message := entry.message
			owned := true
			defer func() {
				if owned {
					message.body.dispose(c, httpErr("connection-terminated"))
					message.body.release()
				}
			}()
			if message.status < 200 {
				return httpErr("HTTP-protocol-error")
			}
			length, failed := httpContentLength(message.headers)
			if failed != nil {
				return failed
			}
			response := &http.Response{StatusCode: int(message.status), Header: httpHeader(message.headers), Trailer: http.Header{}, ContentLength: -1}
			var stream io.ReadCloser
			if collector == nil {
				stream, failed = message.body.pump(c, response.Trailer, length)
			} else {
				collector.data, failed = message.body.collect(c, response.Trailer, length, collector.limit)
				stream = http.NoBody
			}
			if failed != nil {
				return failed
			}
			message.body.handled()
			response.Body = &httpCallBody{ReadCloser: stream, done: callDone, interrupt: context.AfterFunc(callCtx, func() { message.body.dispose(nil, callCtx.Err()) }), release: message.body.release, cancel: cancel, stop: stop}
			owned = false
			reply.set(response, nil)
			return nil
		})
		if collector != nil {
			collector.fuel, collector.failure = fuel, failed
		}
		if failed != nil {
			body.dispose(nil, failed)
			reply.set(nil, failed)
			cancel()
		} else {
			reply.set(nil, httpErr("HTTP-response-incomplete"))
		}
	}
	if collector != nil {
		run()
	} else {
		if err = p.worker(run); err != nil {
			return nil, err
		}
	}
	reserved = false
	accepted = true
	if collector != nil && collector.failure != nil {
		reply.abandon()
		stop()
		cancel()
		return nil, collector.failure
	}
	select {
	case <-reply.done:
		reply.mu.Lock()
		response, failure := reply.response, reply.failure
		reply.mu.Unlock()
		if failure != nil {
			stop()
			cancel()
			return nil, failure
		}
		return response, nil
	case <-callCtx.Done():
		reply.abandon()
		stop()
		cancel()
		return nil, callCtx.Err()
	}
}

type httpCallBody struct {
	io.ReadCloser
	once      sync.Once
	release   func()
	cancel    context.CancelFunc
	stop      func() bool
	done      <-chan struct{}
	eof       atomic.Bool
	interrupt func() bool
}

func (b *httpCallBody) discard() (err error) {
	b.once.Do(func() {
		if b.interrupt != nil {
			b.interrupt()
		}
		err = b.ReadCloser.Close()
		b.release()
		b.stop()
		b.cancel()
	})
	return err
}

func (b *httpCallBody) Read(out []byte) (int, error) {
	n, err := b.ReadCloser.Read(out)
	if err == io.EOF {
		b.eof.Store(true)
	}
	return n, err
}
func (b *httpCallBody) Close() error {
	if b.eof.Load() {
		<-b.done
	}
	err := b.discard()
	<-b.done
	return err
}

// HTTPInstanceFactory supplies a fresh Store and request-scoped owners. Cleanup
// closes and joins the HTTP/network/IO owners and then the component instance.
type HTTPInstanceFactory func(context.Context) (*component.Instance, *HTTP, func(), error)

func HTTPHandler(factory HTTPInstanceFactory, options component.CallOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if factory == nil {
			http.Error(w, "HTTP component unavailable", 503)
			return
		}
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		in, p, cleanup, err := factory(ctx)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			http.Error(w, "HTTP component unavailable", 503)
			return
		}
		if in == nil || p == nil || cleanup == nil {
			http.Error(w, "invalid HTTP component factory", 500)
			return
		}
		controller := http.NewResponseController(w)
		request := r.Clone(ctx)
		request.Body = &httpServerBody{ReadCloser: r.Body, interrupt: func() { _ = controller.SetReadDeadline(time.Now()) }}
		response, err := p.CallIncoming(ctx, in, request, options)
		if err != nil {
			p.Close()
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
		if _, err = io.Copy(w, response.Body); err != nil {
			panic(http.ErrAbortHandler)
		}
		for name, values := range response.Trailer {
			for _, value := range values {
				w.Header().Add(http.TrailerPrefix+name, value)
			}
		}
	})
}

// The deadline interrupts an unfinished server-side read before Body.Close.
// Some net/http bodies otherwise wait for a peer indefinitely during Close.
type httpServerBody struct {
	io.ReadCloser
	mu        sync.Mutex
	done      bool
	interrupt func()
}

func (b *httpServerBody) Read(out []byte) (int, error) {
	if b.ReadCloser == nil {
		return 0, io.EOF
	}
	n, err := b.ReadCloser.Read(out)
	if err != nil {
		b.mu.Lock()
		b.done = true
		b.mu.Unlock()
	}
	return n, err
}
func (b *httpServerBody) Close() error {
	b.mu.Lock()
	done := b.done
	b.mu.Unlock()
	if !done && b.interrupt != nil {
		b.interrupt()
	}
	if b.ReadCloser != nil {
		return b.ReadCloser.Close()
	}
	return nil
}
