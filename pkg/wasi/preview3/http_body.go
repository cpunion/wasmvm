//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpCell struct {
	mu    sync.Mutex
	done  chan struct{}
	ready bool
	value component.Value
}

func newHTTPCell() *httpCell { return &httpCell{done: make(chan struct{})} }
func (s *httpCell) set(value component.Value) {
	s.mu.Lock()
	if s.ready {
		s.mu.Unlock()
		return
	}
	s.ready = true
	s.value = value
	close(s.done)
	s.mu.Unlock()
}
func (s *httpCell) source(p *HTTP, close func()) component.AsyncSource {
	return component.AsyncSource{Close: close, Read: func(c *component.Caller, _ uint32) (*component.ReadCompletion, error) {
		completion := component.NewReadCompletion()
		ctx := c.Context()
		err := p.network.io.worker(func() {
			select {
			case <-ctx.Done():
				completion.Cancelled()
				return
			case <-s.done:
			}
			s.mu.Lock()
			value := s.value
			s.mu.Unlock()
			completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
				return component.ReadResult{Values: []component.Value{value}, Closed: true}, nil
			})
		})
		if err != nil {
			return nil, err
		}
		return completion, nil
	}}
}

type httpBody struct {
	mu                         sync.Mutex
	owner                      *HTTP
	refs                       uint32
	ctx                        context.Context
	cancel                     context.CancelFunc
	closed, bodyDone, eof      bool
	readerClosed, closeQueued  bool
	transmitted, handlingDone  bool
	failure                    error
	done                       chan struct{}
	tx                         *httpCell
	original, trailers, anchor *component.Resource
	controls                   []*component.Cancellation
	reader                     io.ReadCloser
	requestBody                io.ReadCloser
	readTimeout                time.Duration
	expected                   *uint64
	received                   uint64
	sizeFailure                string
	readerOnce                 sync.Once
	readerMu                   sync.Mutex
	trailerType                component.ValueType
	trailerDone                chan struct{}
	trailerReady               bool
	trailersValue              *host.HTTPFields
	responseHeaders            *host.HTTPFields
	trailerSnapshot            func() (*host.HTTPFields, error)
	networkCancel              context.CancelFunc
	networkStop                func() bool
}

func (p *HTTP) newBody() (*httpBody, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	if uint64(len(p.bodies)) >= uint64(p.config.MaxMessages) {
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(p.ctx)
	b := &httpBody{owner: p, refs: 1, ctx: ctx, cancel: cancel, done: make(chan struct{}), tx: newHTTPCell()}
	p.bodies[b] = struct{}{}
	return b, nil
}
func (b *httpBody) acquire() { b.mu.Lock(); b.refs++; b.mu.Unlock() }
func (b *httpBody) release() {
	b.mu.Lock()
	b.refs--
	last := b.refs == 0
	b.mu.Unlock()
	if !last {
		return
	}
	b.dispose(nil, httpErr("connection-terminated"))
	b.mu.Lock()
	reader, closed, queued := b.reader, b.readerClosed, b.closeQueued
	b.mu.Unlock()
	if reader == nil || closed && !queued {
		b.owner.mu.Lock()
		delete(b.owner.bodies, b)
		b.owner.mu.Unlock()
	} else {
		b.queueReaderClose()
	}
}
func (b *httpBody) queueReaderClose() {
	b.mu.Lock()
	if b.reader == nil || b.readerClosed || b.closeQueued {
		b.mu.Unlock()
		return
	}
	b.closeQueued = true
	b.mu.Unlock()
	b.owner.cleanup <- func() {
		b.closeReader()
		b.mu.Lock()
		b.closeQueued = false
		last := b.refs == 0
		b.mu.Unlock()
		if last {
			b.owner.mu.Lock()
			delete(b.owner.bodies, b)
			b.owner.mu.Unlock()
		}
	}
}
func (b *httpBody) closeReader() {
	b.readerOnce.Do(func() {
		defer func() {
			if recover() != nil {
				b.end(httpErr("internal-error"))
			}
			b.mu.Lock()
			b.readerClosed = true
			b.mu.Unlock()
		}()
		b.mu.Lock()
		reader, requestBody, cancel, stop := b.reader, b.requestBody, b.networkCancel, b.networkStop
		b.mu.Unlock()
		if stop != nil {
			stop()
		}
		if cancel != nil {
			cancel()
		}
		if requestBody != nil {
			defer requestBody.Close()
		}
		if reader != nil {
			reader.Close()
		}
	})
}
func (b *httpBody) control(signal *component.Cancellation) {
	b.mu.Lock()
	b.controls = append(b.controls, signal)
	closed := b.closed
	b.mu.Unlock()
	if closed {
		signal.Cancel()
	}
}
func (b *httpBody) transferred() {
	b.mu.Lock()
	b.transmitted = true
	ready := b.handlingDone
	failure := b.failure
	b.mu.Unlock()
	if failure != nil {
		b.tx.set(httpErrorResult(failure))
	} else if ready {
		b.tx.set(component.Value{})
	}
}
func (b *httpBody) handled() {
	b.mu.Lock()
	b.handlingDone = true
	ready := b.transmitted
	failure := b.failure
	b.mu.Unlock()
	if failure != nil {
		b.tx.set(httpErrorResult(failure))
	} else if ready {
		b.tx.set(component.Value{})
	}
}
func (b *httpBody) end(err error) {
	b.mu.Lock()
	if err != nil && b.failure == nil {
		b.failure = err
	}
	if !b.bodyDone {
		b.bodyDone = true
		b.eof = err == nil
		close(b.done)
	}
	b.mu.Unlock()
	if err != nil {
		b.tx.set(httpErrorResult(err))
	}
}
func (b *httpBody) dispose(c *component.Caller, err error) {
	b.mu.Lock()
	first := !b.closed
	b.closed = true
	original, trailers, anchor := b.original, b.trailers, b.anchor
	if c != nil {
		b.original, b.trailers, b.anchor = nil, nil, nil
	}
	cancel := b.networkCancel
	controls := append([]*component.Cancellation(nil), b.controls...)
	b.mu.Unlock()
	if first {
		b.end(err)
		b.cancel()
		if cancel != nil {
			cancel()
		}
		for _, signal := range controls {
			signal.Cancel()
		}
		b.queueReaderClose()
	}
	if c != nil {
		for _, r := range []*component.Resource{anchor, original, trailers} {
			if r != nil {
				_ = c.DropAsync(r)
			}
		}
	}
}
func (b *httpBody) txSource(c *component.Caller, t component.ValueType) (component.Value, error) {
	b.acquire()
	v, err := c.NewSource(t, b.tx.source(b.owner, b.release))
	if err != nil {
		b.release()
	}
	return v, err
}
func (p *HTTP) newMessage(c *component.Caller, kind string, args []component.Value) (out *component.Value, err error) {
	var b *httpBody
	published := false
	defer func() {
		if !published {
			if b != nil {
				b.dispose(c, httpErr("connection-terminated"))
			}
			_ = c.DropResource(args[0].Resource)
			if args[1].Case == 1 {
				_ = c.DropAsync(args[1].Elements[0].Resource)
			}
			_ = c.DropAsync(args[2].Resource)
			if kind == "request" && args[3].Case == 1 {
				_ = c.DropResource(args[3].Elements[0].Resource)
			}
		}
	}()
	b, err = p.newBody()
	if err != nil {
		return nil, err
	}
	defer b.release()
	headers, err := p.take(c, args[0], "fields", false)
	if err != nil {
		return nil, err
	}
	headers.fields.Freeze()
	message := &httpMessage{headers: headers.fields, method: "GET", status: 200, body: b}
	b.sizeFailure = "HTTP-" + kind + "-body-size"
	if kind == "request" && args[3].Case == 1 {
		options, err := p.take(c, args[3].Elements[0], "request-options", false)
		if err != nil {
			return nil, err
		}
		copy := httpCopyOptions(options.options)
		message.options = &copy
	}
	_, futureType := c.FunctionInfo().Result().Field(1)
	var owned []*component.Resource
	if args[1].Case == 1 {
		owned = append(owned, args[1].Elements[0].Resource)
	}
	owned = append(owned, args[2].Resource)
	signal := component.NewCancellation()
	anchor, refs, err := c.NewOwnedSource(futureType, component.AsyncSource{Owned: owned, Cancel: signal, Read: func(*component.Caller, uint32) (*component.ReadCompletion, error) { return nil, wasm.ErrTrap }})
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.anchor = anchor.Resource
	if args[1].Case == 1 {
		b.original = refs[0]
	}
	b.trailers = refs[len(refs)-1]
	_, b.trailerType = c.FunctionInfo().Parameter(2)
	b.trailerDone = make(chan struct{})
	b.mu.Unlock()
	b.control(signal)
	future, err := b.txSource(c, futureType)
	if err != nil {
		return nil, err
	}
	value, err := p.publish(c, &httpEntry{kind: kind, message: message})
	if err != nil {
		return nil, err
	}
	published = true
	return &component.Value{Elements: []component.Value{value, future}}, nil
}
func (p *HTTP) consumeBody(c *component.Caller, kind string, args []component.Value) (out *component.Value, err error) {
	entry, err := p.take(c, args[0], kind, true)
	if err != nil {
		_ = c.DropAsync(args[1].Resource)
		return nil, err
	}
	b := entry.message.body
	defer b.release()
	accepted := false
	defer func() {
		if !accepted {
			b.dispose(c, httpErr("connection-terminated"))
			_ = c.DropAsync(args[1].Resource)
		}
	}()
	if err := b.watchTrailers(c); err != nil {
		return nil, err
	}
	_, streamType := c.FunctionInfo().Result().Field(0)
	_, futureType := c.FunctionInfo().Result().Field(1)
	stream, err := b.byteSource(c, streamType)
	if err != nil {
		return nil, err
	}
	future, err := b.trailerSource(c, futureType)
	if err != nil {
		return nil, err
	}
	b.acquire()
	handled := false
	if err := c.Schedule(component.AsyncWork{Owned: []*component.Resource{args[1].Resource}, Start: func(c *component.Caller, refs []*component.Resource) (*component.HostCompletion, error) {
		r, err := c.ReadAsyncContext(b.ctx, refs[0], 1)
		completion := component.NewHostCompletion()
		if err != nil {
			b.dispose(nil, httpErr("connection-terminated"))
			if c.Context().Err() != nil {
				completion.Cancelled()
				return completion, nil
			}
			completion.Resolve(func(*component.Caller) (*component.Value, error) { return nil, nil })
			return completion, nil
		}
		if len(r.Values) != 1 {
			return nil, wasm.ErrTrap
		}
		if r.Values[0].Case == 1 {
			b.dispose(nil, &httpError{value: r.Values[0].Elements[0]})
		} else {
			b.handled()
		}
		handled = true
		completion.Resolve(func(*component.Caller) (*component.Value, error) { return nil, nil })
		return completion, nil
	}, Cleanup: func() {
		if !handled {
			b.dispose(nil, httpErr("connection-terminated"))
		}
		b.release()
	}}); err != nil {
		b.release()
		return nil, err
	}
	b.mu.Lock()
	anchor := b.anchor
	b.anchor = nil
	b.mu.Unlock()
	if anchor != nil {
		_ = c.DropAsync(anchor)
	}
	accepted = true
	return &component.Value{Elements: []component.Value{stream, future}}, nil
}
func (b *httpBody) byteSource(c *component.Caller, t component.ValueType) (component.Value, error) {
	b.mu.Lock()
	original := b.original
	b.mu.Unlock()
	b.acquire()
	signal := component.NewCancellation()
	source := component.AsyncSource{Cancel: signal, Close: func() {
		b.mu.Lock()
		done := b.bodyDone
		b.mu.Unlock()
		if !done {
			b.dispose(nil, httpErr("connection-terminated"))
		}
		b.release()
	}}
	if original != nil {
		source.Owned = []*component.Resource{original}
		source.ReadOwned = func(c *component.Caller, refs []*component.Resource, limit uint32) (*component.ReadCompletion, error) {
			r, err := c.ReadAsyncContext(b.ctx, refs[0], limit)
			completion := component.NewReadCompletion()
			if err != nil {
				if c.Context().Err() != nil {
					completion.Cancelled()
					return completion, nil
				}
				b.end(httpErr("connection-terminated"))
				r = component.ReadResult{Closed: true}
			}
			if r.Closed {
				b.end(nil)
			}
			completion.Resolve(func(*component.Caller) (component.ReadResult, error) { return r, nil })
			return completion, nil
		}
	} else {
		source.Read = func(c *component.Caller, limit uint32) (*component.ReadCompletion, error) {
			completion := component.NewReadCompletion()
			ctx := c.Context()
			b.mu.Lock()
			reader, failure := b.reader, b.failure
			b.mu.Unlock()
			if reader == nil || failure != nil {
				b.end(failure)
				completion.Resolve(func(*component.Caller) (component.ReadResult, error) { return component.ReadResult{Closed: true}, nil })
				return completion, nil
			}
			joint, cancel := context.WithCancel(ctx)
			stop := context.AfterFunc(b.ctx, cancel)
			err := b.owner.worker(func() {
				defer cancel()
				defer stop()
				b.readerMu.Lock()
				defer b.readerMu.Unlock()
				var data []byte
				var err error
				defer func() {
					if recover() != nil {
						err = httpErr("internal-error")
					}
					if ctx.Err() != nil && len(data) == 0 {
						completion.Cancelled()
						return
					}
					closed := err != nil

					r := component.ReadResult{Bytes: string(data), Closed: closed}
					completion.Resolve(func(*component.Caller) (component.ReadResult, error) { return r, nil })
				}()
				if limit == 0 {
					return
				}
				if joint.Err() != nil {
					err = joint.Err()
					return
				}
				buffer := make([]byte, min(limit, b.owner.network.io.config.MaxReadBytes))
				n, failed := b.readExternal(buffer)
				if n < 0 || n > len(buffer) {
					err = httpErr("internal-error")
					return
				}
				data = buffer[:n]
				err = failed
				if n == 0 && err == nil {
					err = io.ErrNoProgress
				}
			})
			if err != nil {
				stop()
				cancel()
				return nil, err
			}
			return completion, nil
		}
	}
	value, err := c.NewSource(t, source)
	if err != nil {
		b.release()
		return value, err
	}
	b.mu.Lock()
	b.original = nil
	b.mu.Unlock()
	b.control(signal)
	return value, nil
}
func (b *httpBody) trailerSource(c *component.Caller, t component.ValueType) (component.Value, error) {
	b.acquire()
	var resolved atomic.Bool
	source := component.AsyncSource{Close: func() {
		if !resolved.Load() {
			b.dispose(nil, httpErr("connection-terminated"))
		}
		b.release()
	}}
	source.ReadOwned = func(c *component.Caller, refs []*component.Resource, _ uint32) (*component.ReadCompletion, error) {
		completion := component.NewReadCompletion()
		b.mu.Lock()
		done := b.bodyDone
		b.mu.Unlock()
		if !done {
			waiter, err := c.NewSource(t, component.AsyncSource{Read: func(c *component.Caller, _ uint32) (*component.ReadCompletion, error) {
				p := component.NewReadCompletion()
				ctx := c.Context()
				err := b.owner.network.io.worker(func() {
					select {
					case <-ctx.Done():
						p.Cancelled()
					case <-b.done:
						p.Resolve(func(*component.Caller) (component.ReadResult, error) {
							return component.ReadResult{Values: []component.Value{httpOK(component.Value{})}, Closed: true}, nil
						})
					}
				})
				return p, err
			}})
			if err != nil {
				return nil, err
			}
			_, err = c.ReadAsyncContext(b.ctx, waiter.Resource, 1)
			_ = c.DropAsync(waiter.Resource)
			if err != nil && c.Context().Err() != nil {
				completion.Cancelled()
				return completion, nil
			}
		}
		if err := b.waitTrailers(c); err != nil {
			if c.Context().Err() != nil {
				completion.Cancelled()
				return completion, nil
			}
			return nil, err
		}
		b.mu.Lock()
		failure, fields := b.failure, b.trailersValue
		// A parent context can cancel before the provider's close callback
		// publishes the body failure. Cancellation cannot become a success.
		if failure == nil && b.ctx.Err() != nil {
			failure = httpErr("connection-terminated")
		}
		b.mu.Unlock()
		var value component.Value
		if failure != nil {
			value = httpErrorResult(failure)
		} else if fields != nil {
			v, err := b.owner.publish(c, &httpEntry{kind: "fields", fields: fields})
			if err != nil {
				return nil, err
			}
			value = httpOK(httpSome(v))
		} else {
			value = httpOK(component.Value{})
		}
		if value.Case == 1 {
			b.tx.set(value)
		} else {
			b.transferred()
		}
		resolved.Store(true)
		completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
			return component.ReadResult{Values: []component.Value{value}, Closed: true}, nil
		})
		return completion, nil
	}
	value, err := c.NewSource(t, source)
	if err != nil {
		b.release()
		return value, err
	}
	b.mu.Lock()
	b.trailers = nil
	b.mu.Unlock()
	return value, nil
}
func (p *HTTP) worker(fn func()) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return wasm.ErrClosed
	}
	p.workers.Add(1)
	p.mu.Unlock()
	err := p.network.io.worker(func() { defer p.workers.Done(); fn() })
	if err != nil {
		p.workers.Done()
	}
	return err
}

// readExternal serializes reads through readerMu at its call sites. The fixed
// cleanup worker interrupts a blocked reader when its per-read timer expires.
func (b *httpBody) readExternal(data []byte) (n int, err error) {
	b.mu.Lock()
	reader, timeout := b.reader, b.readTimeout
	b.mu.Unlock()
	var timer *time.Timer
	var expired atomic.Bool
	if timeout > 0 {
		timer = time.AfterFunc(timeout, func() { expired.Store(true); b.dispose(nil, httpErr("connection-read-timeout")) })
	}
	defer func() {
		if timer != nil {
			timer.Stop()
		}
		if expired.Load() {
			err = httpErr("connection-read-timeout")
		}
		if recover() != nil {
			n = 0
			err = httpErr("internal-error")
		}
		if n < 0 || n > len(data) {
			n = 0
			err = httpErr("internal-error")
		}
		if uint64(n) > ^uint64(0)-b.received {
			err = httpErr(b.sizeFailure)
		} else {
			b.received += uint64(n)
		}
		if b.expected != nil && (b.received > *b.expected || err == io.EOF && b.received != *b.expected) {
			err = httpErr(b.sizeFailure)
		}
		if n == 0 && err == nil {
			err = io.ErrNoProgress
		}
		if err != nil {
			failure := err
			if errors.Is(err, io.EOF) {
				failure = nil
				if b.trailerSnapshot != nil {
					fields, failed := b.trailerSnapshot()
					failure = failed
					b.mu.Lock()
					b.trailersValue = fields
					b.mu.Unlock()
				}
			}
			if failure != nil {
				failure = httpTransportError(failure, "HTTP-protocol-error")
			}
			b.end(failure)
			b.closeReader()
			if failure != nil {
				err = failure
			}
		}
	}()
	return reader.Read(data)
}

// Observe trailers concurrently with the body. A producer can signal an error
// before body EOF; that error must interrupt the transfer immediately.
func (b *httpBody) watchTrailers(c *component.Caller) error {
	b.mu.Lock()
	ref := b.trailers
	b.mu.Unlock()
	if ref == nil {
		return nil
	}
	b.acquire()
	observed := false
	err := c.Schedule(component.AsyncWork{Owned: []*component.Resource{ref}, Start: func(c *component.Caller, refs []*component.Resource) (*component.HostCompletion, error) {
		return c.AwaitRead(b.ctx, refs[0], 1, func(c *component.Caller, result component.ReadResult, err error) (*component.Value, error) {
			var fields *host.HTTPFields
			if err == nil {
				if len(result.Values) != 1 {
					err = wasm.ErrTrap
				} else {
					value := result.Values[0]
					if value.Case == 1 {
						err = &httpError{value: value.Elements[0]}
					} else if value.Elements[0].Case == 1 {
						var entry *httpEntry
						entry, err = b.owner.take(c, value.Elements[0].Elements[0], "fields", false)
						if err == nil {
							entry.fields.Freeze()
							fields = entry.fields
						}
					}
				}
			}
			if err != nil {
				err = httpTransportError(err, "connection-terminated")
			}
			b.finishTrailers(fields, err)
			observed = true
			return nil, nil
		})
	}, Cleanup: func() {
		if !observed {
			b.finishTrailers(nil, httpErr("connection-terminated"))
		}
		b.release()
	}})
	if err != nil {
		b.release()
		return err
	}
	b.mu.Lock()
	b.trailers = nil
	b.mu.Unlock()
	return nil
}
func (b *httpBody) finishTrailers(fields *host.HTTPFields, err error) {
	b.mu.Lock()
	if !b.trailerReady {
		b.trailerReady = true
		b.trailersValue = fields
		close(b.trailerDone)
	}
	b.mu.Unlock()
	if err != nil {
		b.dispose(nil, err)
	}
}
func (b *httpBody) waitTrailers(c *component.Caller) error {
	b.mu.Lock()
	done, ready, typ := b.trailerDone, b.trailerReady, b.trailerType
	b.mu.Unlock()
	if done == nil || ready {
		return nil
	}
	waiter, err := c.NewSource(typ, component.AsyncSource{Read: func(c *component.Caller, _ uint32) (*component.ReadCompletion, error) {
		completion := component.NewReadCompletion()
		ctx := c.Context()
		err := b.owner.network.io.worker(func() {
			select {
			case <-ctx.Done():
				completion.Cancelled()
			case <-done:
				completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
					return component.ReadResult{Values: []component.Value{httpOK(component.Value{})}, Closed: true}, nil
				})
			}
		})
		return completion, err
	}})
	if err != nil {
		return err
	}
	defer c.DropAsync(waiter.Resource)
	_, err = c.ReadAsyncContext(b.ctx, waiter.Resource, 1)
	if c.Context().Err() == nil && b.ctx.Err() != nil {
		return nil
	}
	return err
}
