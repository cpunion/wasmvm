//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpCell struct {
	mu        sync.Mutex
	ready     bool
	usedAsync bool
	readyFlag atomic.Bool
	value     component.Value
	resultVal [1]component.Value
	pending   *component.ReadCompletion
	stop      func() bool
}

func (s *httpCell) isReady() bool { return s.readyFlag.Load() }

func (s *httpCell) set(value component.Value) {
	s.mu.Lock()
	if s.ready {
		s.mu.Unlock()
		return
	}
	s.ready = true
	s.value = value
	s.resultVal[0] = value
	s.readyFlag.Store(true)
	pending, stop := s.pending, s.stop
	s.pending, s.stop = nil, nil
	s.mu.Unlock()
	if pending != nil && stop() {
		s.resolveCompletion(pending)
	}
}

func (s *httpCell) readReady(*component.Caller, uint32) (component.ReadResult, bool, error) {
	if !s.readyFlag.Load() {
		return component.ReadResult{}, false, nil
	}
	return component.ReadResult{Values: s.resultVal[:1], Closed: true}, true, nil
}

func (s *httpCell) resolveCompletion(completion *component.ReadCompletion) {
	_ = completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
		return component.ReadResult{Values: s.resultVal[:1], Closed: true}, nil
	})
}

// A message publishes one transmit future. Its read consumes a subscription
// under the Store lease, and the producer may publish from any goroutine.
func (s *httpCell) read(c *component.Caller, _ uint32) (*component.ReadCompletion, error) {
	s.mu.Lock()
	if s.ready {
		s.mu.Unlock()
		completion := component.NewReadCompletion()
		s.resolveCompletion(completion)
		return completion, nil
	}
	if s.pending != nil {
		s.mu.Unlock()
		return nil, wasm.ErrTrap
	}
	var completion *component.ReadCompletion
	completion = component.NewReadCompletionWithCleanup(func() {
		s.mu.Lock()
		if s.pending != completion {
			s.mu.Unlock()
			return
		}
		stop := s.stop
		s.pending, s.stop = nil, nil
		s.mu.Unlock()
		stop()
	})
	s.usedAsync = true
	s.pending = completion
	s.stop = context.AfterFunc(c.Context(), func() { _ = completion.Cancelled() })
	s.mu.Unlock()
	return completion, nil
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
	hasTrailers                bool
	consumeHandled             bool
	trailerObserved            bool
	unpooled                   bool
	failure                    error
	done                       chan struct{}
	tx                         httpCell
	original, trailers, anchor *component.Resource
	inlineControls             [2]*component.Cancellation
	controls                   []*component.Cancellation
	anchorCancel               component.Cancellation
	byteCancel                 component.Cancellation
	readBuf                    httpReadBuffer
	trailerResolved            atomic.Bool
	trailerResultBuf           [1]component.Value
	pairBuf                    [2]component.Value
	pairVal                    component.Value
	reader                     io.ReadCloser
	requestBody                io.ReadCloser
	readTimeout                time.Duration
	expectedValue              uint64
	expected                   *uint64
	received                   uint64
	sizeFailure                string
	readerOnce                 sync.Once
	readerMu                   sync.Mutex
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
	if p.closed {
		p.mu.Unlock()
		return nil, wasm.ErrClosed
	}
	if uint64(len(p.bodies)) >= uint64(p.config.MaxMessages) {
		p.mu.Unlock()
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(p.ctx)
	var b *httpBody
	if n := len(p.freeBodies); n != 0 {
		b = p.freeBodies[n-1]
		p.freeBodies[n-1] = nil
		p.freeBodies = p.freeBodies[:n-1]
		*b = httpBody{owner: p, refs: 1, ctx: ctx, cancel: cancel}
	} else {
		b = &httpBody{owner: p, refs: 1, ctx: ctx, cancel: cancel}
	}
	p.bodies[b] = struct{}{}
	p.mu.Unlock()
	return b, nil
}
func (b *httpBody) acquire() { b.mu.Lock(); b.refs++; b.mu.Unlock() }
func (b *httpBody) release() {
	b.mu.Lock()
	b.refs--
	last := b.refs == 0
	alreadyDone := b.closed || b.bodyDone && b.tx.isReady()
	b.mu.Unlock()
	if !last {
		return
	}
	if alreadyDone {
		b.dispose(nil, nil)
	} else {
		b.dispose(nil, errConnectionTerminated)
	}
	b.mu.Lock()
	reader, requestBody, closed, queued, netStop, unpooled := b.reader, b.requestBody, b.readerClosed, b.closeQueued, b.networkStop, b.unpooled
	b.mu.Unlock()
	if reader != nil && !closed && !queued && requestBody == nil && isInlineCloseReader(reader) {
		b.closeReader()
		closed = true
	}
	if reader == nil || closed && !queued {
		b.tx.mu.Lock()
		usedAsync := b.tx.usedAsync
		b.tx.mu.Unlock()
		owner := b.owner
		owner.mu.Lock()
		delete(owner.bodies, b)
		if !owner.closed && netStop == nil && !unpooled && !usedAsync && len(owner.freeBodies) < 8 {
			*b = httpBody{}
			owner.freeBodies = append(owner.freeBodies, b)
		}
		owner.mu.Unlock()
	} else {
		b.queueReaderClose()
	}
}
func isInlineCloseReader(reader io.ReadCloser) bool {
	if reader == http.NoBody {
		return true
	}
	_, ok := reader.(*snapshotReader)
	return ok
}
func (b *httpBody) queueReaderClose() {
	b.mu.Lock()
	if b.reader == nil || b.readerClosed || b.closeQueued {
		b.mu.Unlock()
		return
	}
	if b.requestBody == nil && isInlineCloseReader(b.reader) {
		b.mu.Unlock()
		b.closeReader()
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
	if b.controls == nil {
		b.controls = b.inlineControls[:0]
	}
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
		if !b.tx.isReady() {
			b.tx.set(httpErrorResult(failure))
		}
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
		if !b.tx.isReady() {
			b.tx.set(httpErrorResult(failure))
		}
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
		if b.done != nil {
			close(b.done)
		}
	}
	b.mu.Unlock()
	if err != nil && !b.tx.isReady() {
		b.tx.set(httpErrorResult(err))
	}
}
func (b *httpBody) bodyDoneChan() (chan struct{}, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.bodyDone {
		return nil, true
	}
	if b.done == nil {
		b.done = make(chan struct{})
	}
	return b.done, false
}
func (b *httpBody) dispose(c *component.Caller, err error) {
	var localControls [2]*component.Cancellation
	var controls []*component.Cancellation
	b.mu.Lock()
	first := !b.closed
	b.closed = true
	original, trailers, anchor := b.original, b.trailers, b.anchor
	if c != nil {
		b.original, b.trailers, b.anchor = nil, nil, nil
	}
	cancel := b.networkCancel
	if n := len(b.controls); n != 0 {
		if n <= len(localControls) {
			copy(localControls[:n], b.controls)
			controls = localControls[:n]
		} else {
			controls = append([]*component.Cancellation(nil), b.controls...)
		}
	}
	b.mu.Unlock()
	if first {
		if err != nil {
			b.end(err)
		}
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
		if anchor != nil {
			_ = c.DropAsync(anchor)
		}
		if original != nil {
			_ = c.DropAsync(original)
		}
		if trailers != nil {
			_ = c.DropAsync(trailers)
		}
	}
}
func (b *httpBody) txSource(c *component.Caller, t component.ValueType) (component.Value, error) {
	b.acquire()
	v, err := c.NewSource(t, component.AsyncSource{ReadReady: b.tx.readReady, Read: b.tx.read, Close: b.release})
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
	entryAlloc := &httpMessageEntry{
		entry:   httpEntry{kind: kind},
		message: httpMessage{headers: headers.fields, method: "GET", status: 200, body: b},
	}
	entryAlloc.entry.message = &entryAlloc.message
	message := &entryAlloc.message
	if kind == "request" {
		b.sizeFailure = "HTTP-request-body-size"
	} else {
		b.sizeFailure = "HTTP-response-body-size"
	}
	if kind == "request" && args[3].Case == 1 {
		options, err := p.take(c, args[3].Elements[0], "request-options", false)
		if err != nil {
			return nil, err
		}
		copy := httpCopyOptions(options.options)
		message.options = &copy
	}
	_, futureType := c.FunctionInfo().Result().Field(1)
	var ownedBuf [2]*component.Resource
	owned := ownedBuf[:0]
	if args[1].Case == 1 {
		owned = append(owned, args[1].Elements[0].Resource)
	}
	owned = append(owned, args[2].Resource)
	signal := &b.anchorCancel
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
	b.hasTrailers = true
	b.mu.Unlock()
	b.control(signal)
	future, err := b.txSource(c, futureType)
	if err != nil {
		return nil, err
	}
	value, err := p.publish(c, &entryAlloc.entry)
	if err != nil {
		return nil, err
	}
	published = true
	b.pairBuf[0], b.pairBuf[1] = value, future
	b.pairVal = component.Value{Elements: b.pairBuf[:2]}
	return &b.pairVal, nil
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
			b.dispose(c, errConnectionTerminated)
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
	b.consumeHandled = false
	owned := [1]*component.Resource{args[1].Resource}
	if err := c.Schedule(component.AsyncWork{Owned: owned[:], Start: func(c *component.Caller, refs []*component.Resource) (*component.HostCompletion, error) {
		r, err := c.ReadAsyncContext(b.ctx, refs[0], 1)
		completion := component.NewHostCompletion()
		if err != nil {
			b.dispose(nil, errConnectionTerminated)
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
		b.consumeHandled = true
		completion.Resolve(func(*component.Caller) (*component.Value, error) { return nil, nil })
		return completion, nil
	}, Cleanup: func() {
		if !b.consumeHandled {
			b.dispose(nil, errConnectionTerminated)
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
	b.pairBuf[0], b.pairBuf[1] = stream, future
	b.pairVal = component.Value{Elements: b.pairBuf[:2]}
	return &b.pairVal, nil
}
func (b *httpBody) byteSource(c *component.Caller, t component.ValueType) (component.Value, error) {
	b.mu.Lock()
	original := b.original
	b.mu.Unlock()
	b.acquire()
	signal := &b.byteCancel
	buffered := &b.readBuf
	source := component.AsyncSource{Cancel: signal, Close: func() {
		buffered.close()
		b.mu.Lock()
		done := b.bodyDone
		b.mu.Unlock()
		if !done {
			b.dispose(nil, errConnectionTerminated)
		}
		b.release()
	}}
	var ownedBuf [1]*component.Resource
	if original != nil {
		ownedBuf[0] = original
		source.Owned = ownedBuf[:]
		source.ReadOwned = func(c *component.Caller, refs []*component.Resource, limit uint32) (*component.ReadCompletion, error) {
			return c.AwaitSourceRead(b.ctx, refs[0], limit, func(c *component.Caller, r component.ReadResult, err error) (component.ReadResult, error) {
				if err != nil {
					if c.Context().Err() != nil {
						return component.ReadResult{}, err
					}
					b.end(httpErr("connection-terminated"))
					r = component.ReadResult{Closed: true}
				}
				if r.Closed {
					b.end(nil)
				}
				return r, nil
			})
		}
	} else {
		source.Prompt = true
		source.ReadBytes = func(limit uint32) (component.ReadResult, bool, error) {
			if result, ready := buffered.take(limit); ready {
				return result, true, nil
			}
			b.mu.Lock()
			reader, failure := b.reader, b.failure
			b.mu.Unlock()
			if reader == nil || failure != nil {
				b.end(failure)
				return component.ReadResult{Closed: true}, true, nil
			}
			if limit != 0 && b.ctx.Err() == nil {
				if data, failed, ready := b.readSnapshot(reader); ready {
					return buffered.deliver(data, failed != nil, limit), true, nil
				}
			}
			return component.ReadResult{}, false, nil
		}
		source.Read = func(c *component.Caller, limit uint32) (*component.ReadCompletion, error) {
			completion := component.NewReadCompletion()
			if result, ready := buffered.take(limit); ready {
				completion.Resolve(func(*component.Caller) (component.ReadResult, error) { return result, nil })
				return completion, nil
			}
			ctx := c.Context()
			b.mu.Lock()
			reader, failure := b.reader, b.failure
			b.mu.Unlock()
			if reader == nil || failure != nil {
				b.end(failure)
				completion.Resolve(func(*component.Caller) (component.ReadResult, error) { return component.ReadResult{Closed: true}, nil })
				return completion, nil
			}
			if limit != 0 && ctx.Err() == nil && b.ctx.Err() == nil {
				if data, failed, ready := b.readSnapshot(reader); ready {
					completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
						return buffered.deliver(data, failed != nil, limit), nil
					})
					return completion, nil
				}
			}
			joint, cancel := context.WithCancel(ctx)
			stop := context.AfterFunc(b.ctx, cancel)
			err := b.owner.worker(func() {
				defer cancel()
				defer stop()
				b.readerMu.Lock()
				defer b.readerMu.Unlock()
				var data string
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

					completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
						return buffered.deliver(data, closed, limit), nil
					})
				}()
				if limit == 0 {
					return
				}
				if joint.Err() != nil {
					err = joint.Err()
					return
				}
				buffer := b.owner.network.io.acquireTransferBuffer(int(b.owner.network.io.config.MaxReadBytes))
				defer b.owner.network.io.releaseTransferBuffer(buffer)
				n, failed := b.readExternal(buffer[:len(buffer):len(buffer)])
				if n < 0 || n > len(buffer) {
					err = httpErr("internal-error")
					return
				}
				// Snapshot before returning the mutable buffer to its owner;
				// completion publication may run after another worker reuses it.
				data = string(buffer[:n])
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
	source := component.AsyncSource{Close: func() {
		if !b.trailerResolved.Load() {
			b.dispose(nil, errConnectionTerminated)
		}
		b.release()
	}}
	source.ReadReady = func(c *component.Caller, _ uint32) (component.ReadResult, bool, error) {
		b.mu.Lock()
		ready := b.bodyDone && (!b.hasTrailers || b.trailerReady)
		b.mu.Unlock()
		if !ready {
			return component.ReadResult{}, false, nil
		}
		value, err := b.trailerResult(c, &b.trailerResolved)
		if err != nil {
			return component.ReadResult{}, true, err
		}
		b.trailerResultBuf[0] = value
		return component.ReadResult{Values: b.trailerResultBuf[:1], Closed: true}, true, nil
	}
	source.ReadOwned = func(c *component.Caller, refs []*component.Resource, _ uint32) (*component.ReadCompletion, error) {
		completion := component.NewReadCompletion()
		if doneChan, done := b.bodyDoneChan(); !done {
			if err := c.WaitSignal(b.ctx, doneChan); err != nil {
				if c.Context().Err() != nil {
					completion.Cancelled()
					return completion, nil
				}
				if b.ctx.Err() == nil {
					return nil, err
				}
			}
		}
		if err := b.waitTrailers(c); err != nil {
			if c.Context().Err() != nil {
				completion.Cancelled()
				return completion, nil
			}
			return nil, err
		}
		value, err := b.trailerResult(c, &b.trailerResolved)
		if err != nil {
			return nil, err
		}
		b.trailerResultBuf[0] = value
		completion.Resolve(func(*component.Caller) (component.ReadResult, error) {
			return component.ReadResult{Values: b.trailerResultBuf[:1], Closed: true}, nil
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

// Called only after both body completion and trailer publication. Resource
// creation uses the current callback's transaction, including on the ready path.
func (b *httpBody) trailerResult(c *component.Caller, resolved *atomic.Bool) (component.Value, error) {
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
			return component.Value{}, err
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
	if timeout > 0 {
		b.unpooled = true
	}
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
		n, err = b.finishExternalRead(n, len(data), err)
	}()
	return reader.Read(data)
}

// Both scheduled external reads and immutable snapshots share byte accounting,
// EOF/trailer publication and exactly-once reader closure. readerMu is held.
func (b *httpBody) finishExternalRead(n, limit int, err error) (int, error) {
	if n < 0 || n > limit {
		n = 0
		err = httpErr("internal-error")
	}
	if uint64(n) > ^uint64(0)-b.received {
		err = httpErr(b.sizeFailure)
	} else {
		b.received += uint64(n)
	}
	if b.expected != nil && (b.received > *b.expected || err == io.EOF && b.received != *b.expected) {
		err = httpSizeError(b.sizeFailure, b.received)
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
	return n, err
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
	b.trailerObserved = false
	owned := [1]*component.Resource{ref}
	err := c.Schedule(component.AsyncWork{Owned: owned[:], Start: func(c *component.Caller, refs []*component.Resource) (*component.HostCompletion, error) {
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
			b.trailerObserved = true
			return nil, nil
		})
	}, Cleanup: func() {
		if !b.trailerObserved {
			b.finishTrailers(nil, errConnectionTerminated)
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
		if err != nil && b.failure == nil {
			b.failure = err
		}
		if b.trailerDone != nil {
			close(b.trailerDone)
		}
	}
	b.mu.Unlock()
	if err != nil {
		b.dispose(nil, err)
	}
}
func (b *httpBody) trailerDoneChan() chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.hasTrailers || b.trailerReady {
		return nil
	}
	if b.trailerDone == nil {
		b.trailerDone = make(chan struct{})
	}
	return b.trailerDone
}
func (b *httpBody) waitTrailers(c *component.Caller) error {
	done := b.trailerDoneChan()
	if done == nil {
		return nil
	}
	err := c.WaitSignal(b.ctx, done)
	if c.Context().Err() == nil && b.ctx.Err() != nil {
		return nil
	}
	return err
}
