//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type httpBody struct {
	mu             sync.Mutex
	owner          *HTTP
	buffer         *httpBodyBuffer
	ctx            context.Context
	cancel         context.CancelFunc
	externalCancel context.CancelFunc
	closeOnce      sync.Once
	done           chan struct{}
	complete       bool
	portTaken      bool
	outgoing       bool
	expected       *uint64
	errorName      string
	reader         io.ReadCloser
	trailers       *httpFields
	trailerMap     http.Header
	failure        error
}

// A transport timeout, explicit body drop and pump completion may race to
// release the same input. io.Closer does not require repeated Close to succeed.
type httpOwnedReader struct {
	io.ReadCloser
	once sync.Once
	err  error
}

func (r *httpOwnedReader) Close() error {
	r.once.Do(func() { r.err = r.ReadCloser.Close() })
	return r.err
}

func (p *HTTP) newBody(outgoing bool) (*httpBody, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	if uint64(len(p.bodies)) >= uint64(p.config.MaxResources) {
		return nil, wasm.ErrLimit
	}
	buffer, err := newHTTPBodyBuffer(p.config.BodyBufferBytes)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(p.ctx)
	b := &httpBody{owner: p, buffer: buffer, ctx: ctx, cancel: cancel, done: make(chan struct{}), outgoing: outgoing, errorName: "HTTP-request-body-size"}
	p.bodies[b] = true
	return b, nil
}
func (b *httpBody) Readiness() (bool, <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.complete, b.done
}
func (b *httpBody) finish(err error, trailers *httpFields) {
	b.mu.Lock()
	if b.complete {
		b.mu.Unlock()
		return
	}
	b.complete = true
	b.failure = err
	b.trailers = trailers
	b.buffer.finish(err)
	close(b.done)
	b.mu.Unlock()
}
func (b *httpBody) close() {
	b.closeOnce.Do(func() {
		b.cancel()
		b.buffer.close()
		b.mu.Lock()
		reader, cancel := b.reader, b.externalCancel
		b.reader = nil
		b.externalCancel = nil
		b.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		if reader != nil {
			b.owner.recordCleanup(reader.Close())
		}
		b.finish(httpErr("HTTP-response-incomplete"), nil)
		b.owner.mu.Lock()
		delete(b.owner.bodies, b)
		b.owner.mu.Unlock()
	})
}
func (b *httpBody) pump(reader io.ReadCloser, trailers http.Header, first, between time.Duration) error {
	if err := b.owner.startWorker(); err != nil {
		return err
	}
	reader = &httpOwnedReader{ReadCloser: reader}
	b.mu.Lock()
	b.reader = reader
	b.mu.Unlock()
	go func() {
		defer b.owner.workers.Done()
		defer func() {
			b.mu.Lock()
			r := b.reader
			b.reader = nil
			b.mu.Unlock()
			if r != nil {
				b.owner.recordCleanup(r.Close())
			}
		}()
		data := make([]byte, 32768)
		timeout := first
		for {
			if b.ctx.Err() != nil {
				return
			}
			var timedOut atomic.Bool
			var timer *time.Timer
			var timerDone chan struct{}
			if timeout >= 0 {
				timerDone = make(chan struct{})
				timer = time.AfterFunc(timeout, func() {
					timedOut.Store(true)
					b.mu.Lock()
					cancel := b.externalCancel
					b.mu.Unlock()
					if cancel != nil {
						cancel()
					}
					reader.Close()
					close(timerDone)
				})
			}
			n, err := reader.Read(data)
			if timer != nil && !timer.Stop() {
				<-timerDone
			}
			if timedOut.Load() {
				err = httpErr("HTTP-response-timeout")
			}
			timeout = between
			pending := data[:n]
			for len(pending) != 0 {
				wrote, failed := b.buffer.write(b.ctx, pending)
				pending = pending[wrote:]
				if failed != nil && !errors.Is(failed, ErrWouldBlock) {
					return
				}
				if len(pending) != 0 {
					ready, changed := b.buffer.readiness(false)
					if !ready {
						select {
						case <-changed:
						case <-b.ctx.Done():
							return
						}
					}
				}
			}
			if err != nil {
				var fields *httpFields
				if errors.Is(err, io.EOF) {
					err = nil
					fields, err = b.owner.fieldsFromHeader(trailers)
				} else {
					err = httpTransportError(err, "HTTP-response-incomplete")
				}
				b.finish(err, fields)
				return
			}
			if n == 0 {
				select {
				case <-b.ctx.Done():
					return
				case <-time.After(time.Millisecond):
				}
			}
		}
	}()
	return nil
}

type httpBodyPort struct {
	mu            sync.Mutex
	once          sync.Once
	body          *httpBody
	parent        *httpEntry
	input, closed bool
}

func (s *httpBodyPort) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return true, nil
	}
	return s.body.buffer.readiness(s.input)
}
func (s *httpBodyPort) Read(ctx context.Context, out []byte) (int, error) {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return 0, io.EOF
	}
	if err := s.body.ctx.Err(); err != nil {
		return 0, io.ErrClosedPipe
	}
	return s.body.buffer.read(ctx, out, false)
}
func (s *httpBodyPort) Write(ctx context.Context, data []byte) (int, error) {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed || s.body.ctx.Err() != nil {
		return 0, io.ErrClosedPipe
	}
	return s.body.buffer.write(ctx, data)
}
func (s *httpBodyPort) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.body.ctx.Err() != nil {
		return io.ErrClosedPipe
	}
	// Bytes have already transferred into the bounded HTTP queue. Waiting for
	// network consumption would deadlock small bodies written before handle.
	s.body.buffer.mu.Lock()
	defer s.body.buffer.mu.Unlock()
	if s.body.buffer.closed {
		return io.ErrClosedPipe
	}
	return s.body.buffer.failure
}
func (s *httpBodyPort) Close() error {
	s.once.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
		s.body.buffer.mu.Lock()
		s.body.buffer.notify()
		s.body.buffer.mu.Unlock()
		s.parent.owner.release(s.parent)
	})
	return nil
}
func (p *HTTP) bodyPort(c *component.Caller, e *httpEntry) (component.Value, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	b := e.body
	b.mu.Lock()
	if b.portTaken {
		b.mu.Unlock()
		return httpUnitResult(false), nil
	}
	b.portTaken = true
	b.mu.Unlock()
	if err := p.child(e); err != nil {
		return component.Value{}, err
	}
	port := &httpBodyPort{body: b, parent: e, input: e.kind == httpIncomingBody}
	var value component.Value
	var err error
	if port.input {
		value, err = p.network.io.NewInput(c, port)
	} else {
		value, err = p.network.io.NewOutput(c, port)
	}
	if err != nil {
		port.Close()
		return component.Value{}, err
	}
	return httpOK(value), nil
}

type httpBodyReader struct {
	body     *httpBody
	ctx      context.Context
	trailers http.Header
}

func (r *httpBodyReader) Read(out []byte) (int, error) {
	n, err := r.body.buffer.read(r.ctx, out, true)
	if errors.Is(err, io.EOF) && r.trailers != nil {
		r.body.mu.Lock()
		fields := r.body.trailers
		r.body.mu.Unlock()
		if fields != nil {
			for _, field := range fields.entries() {
				r.trailers.Add(field.name, field.value)
			}
		}
		r.trailers = nil
	}
	return n, err
}
func (r *httpBodyReader) Close() error { r.body.close(); return nil }
