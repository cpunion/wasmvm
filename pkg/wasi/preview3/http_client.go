//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpSendJob struct {
	mu                sync.Mutex
	closed            bool
	request, response *httpBody
	cancel            context.CancelFunc
	stop              func() bool
}

func (j *httpSendJob) result(body *httpBody) {
	j.mu.Lock()
	if j.closed {
		j.mu.Unlock()
		body.release()
		return
	}
	j.response = body
	j.mu.Unlock()
}
func (j *httpSendJob) close() {
	j.mu.Lock()
	if j.closed {
		j.mu.Unlock()
		return
	}
	j.closed = true
	response := j.response
	j.response = nil
	j.mu.Unlock()
	j.stop()
	if response == nil {
		j.cancel()
	} else {
		response.release()
	}
	j.request.release()
}
func httpCompletion(value component.Value) *component.HostCompletion {
	p := component.NewHostCompletion()
	p.Resolve(func(*component.Caller) (*component.Value, error) { return &value, nil })
	return p
}
func httpDuration(value uint64) time.Duration {
	if value == 0 {
		return time.Nanosecond
	}
	if value > 1<<63-1 {
		return time.Duration(1<<63 - 1)
	}
	return time.Duration(value)
}
func (p *HTTP) reserveRequest() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	if p.requests >= p.config.MaxRequests {
		return httpErr("connection-limit-reached")
	}
	p.requests++
	return nil
}
func (p *HTTP) releaseRequest() { p.mu.Lock(); p.requests--; p.mu.Unlock() }
func (p *HTTP) send(handler bool) component.AsyncHostFunction {
	return func(c *component.Caller, args []component.Value) (completion *component.HostCompletion, err error) {
		if err = p.begin(c.Context()); err != nil {
			return nil, err
		}
		defer p.active.Done()
		entry, err := p.take(c, args[0], "request", false)
		if err != nil {
			return nil, err
		}
		message := entry.message
		body := message.body
		accepted := false
		defer func() {
			if !accepted {
				body.dispose(c, httpErr("connection-terminated"))
				body.release()
			}
		}()
		target, failed := p.requestURL(message)
		if failed != nil {
			return httpCompletion(httpErrorResult(failed)), nil
		}
		if handler {
			if p.config.Handler == nil {
				return httpCompletion(httpErrorResult(httpErr("configuration-error"))), nil
			}
		} else if failed = p.authorizeRequest(c.Context(), target); failed != nil {
			return httpCompletion(httpErrorResult(failed)), nil
		}
		if failed = p.reserveRequest(); failed != nil {
			return httpCompletion(httpErrorResult(failed)), nil
		}
		requestHeld := true
		defer func() {
			if requestHeld {
				p.releaseRequest()
			}
		}()
		ctx, cancel := context.WithCancel(p.ctx)
		callCtx := c.Context()
		stop := context.AfterFunc(callCtx, cancel)
		bodyStop := context.AfterFunc(body.ctx, cancel)

		job := &httpSendJob{request: body, cancel: cancel, stop: stop}
		startAccepted := false
		defer func() {
			if !startAccepted {
				stop()
				bodyStop()
				cancel()
			}
		}()
		expected, failed := httpContentLength(message.headers)
		if failed != nil {
			return httpCompletion(httpErrorResult(failed)), nil
		}
		request := &http.Request{Method: message.method, URL: target, Header: httpHeader(message.headers), Host: target.Host, Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1, ContentLength: -1, TransferEncoding: []string{"chunked"}, Trailer: http.Header{}}
		pump, failed := body.pump(c, request.Trailer, expected)
		if failed != nil {
			return nil, failed
		}
		request.Body = pump
		request = request.WithContext(ctx)
		completion = component.NewHostCompletionWithCleanup(job.close)
		if failed = p.worker(func() {
			defer p.releaseRequest()
			var response *http.Response
			var result *httpBody
			var failure error
			defer func() {
				if recover() != nil {
					failure = httpErr("internal-error")
				}
				if response != nil && result == nil && response.Body != nil {
					func() {
						defer func() {
							if recover() != nil {
								failure = httpErr("internal-error")
							}
						}()
						response.Body.Close()
					}()
				}
				if result == nil {
					request.Body.Close()
				}
				if result != nil {
					job.result(result)
				}
				if failure != nil {
					body.dispose(nil, failure)
				} else {
					body.handled()
				}
				if callCtx.Err() != nil {
					completion.Cancelled()
					return
				}
				completion.Resolve(func(fresh *component.Caller) (*component.Value, error) {
					if failure != nil {
						v := httpErrorResult(failure)
						return &v, nil
					}
					stop()
					value, err := p.publish(fresh, &httpEntry{kind: "response", message: &httpMessage{headers: result.responseHeaders, method: message.method, status: uint16(response.StatusCode), body: result}})
					if err != nil {
						return nil, err
					}
					v := httpOK(value)
					return &v, nil
				})
			}()
			timeouts := p.defaultTimeouts()
			if message.options != nil {
				for i, value := range *message.options {
					if value != nil {
						timeouts[i] = httpDuration(*value)
					}
				}
			}
			var failed error
			if handler {
				response, failed = p.config.Handler(ctx, request)
			} else {

				var transport *http.Transport
				transport, failed = p.transportFor(timeouts)
				if failed == nil {
					response, failed = transport.RoundTrip(request)
				}
			}
			if failed != nil {
				// A local body rejection can cancel the transport before its
				// generic network error arrives. Preserve the known cause.
				body.mu.Lock()
				local := body.failure
				body.mu.Unlock()
				failure = httpTransportError(failed, "HTTP-protocol-error")
				var size *httpError
				if errors.As(local, &size) && (httpErrorNames[size.value.Case] == "HTTP-request-body-size" || httpErrorNames[size.value.Case] == "HTTP-response-body-size") {
					failure = size
				}
				return
			}
			if response == nil || response.StatusCode < 100 || response.StatusCode > 999 {
				failure = httpErr("HTTP-protocol-error")
				return
			}
			bodyStop()
			if response.Body == nil || response.Body == http.NoBody {
				request.Body.Close()
			}
			if response.Body == nil {
				response.Body = io.NopCloser(&emptyHTTPReader{})
			}
			headers, failed := p.fieldsFromHeader(response.Header)
			if failed != nil {
				failure = httpErr("HTTP-response-header-section-size")
				return
			}
			result, failed = p.newBody()
			if failed != nil {
				failure = httpErr("connection-limit-reached")
				return
			}
			result.mu.Lock()
			result.reader = response.Body
			result.requestBody = request.Body
			result.responseHeaders = headers
			result.readTimeout = timeouts[2]
			result.sizeFailure = "HTTP-response-body-size"
			if response.ContentLength >= 0 {
				length := uint64(response.ContentLength)
				result.expected = &length
			}
			result.networkCancel = cancel
			result.networkStop = bodyStop
			result.trailerSnapshot = func() (*host.HTTPFields, error) {
				if len(response.Trailer) == 0 {
					return nil, nil
				}
				fields, err := p.fieldsFromHeader(response.Trailer)
				if err != nil {
					return nil, httpErr("HTTP-response-trailer-section-size")
				}
				return fields, nil
			}
			closed := result.closed
			result.mu.Unlock()
			if closed {
				result.queueReaderClose()
			}
		}); failed != nil {
			pump.Close()
			body.dispose(c, failed)
			return nil, failed
		}
		requestHeld = false
		accepted = true
		startAccepted = true
		return completion, nil
	}
}

type emptyHTTPReader struct{}

func (*emptyHTTPReader) Read([]byte) (int, error) { return 0, io.EOF }
