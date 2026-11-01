//go:build linux || darwin

package preview3

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
)

// The external reader owns the blocking half of an unbuffered pipe. One
// bounded guest batch is retained by the existing task scheduler per message.
type httpPump struct {
	body      *httpBody
	reader    *io.PipeReader
	writer    *io.PipeWriter
	trailer   http.Header
	fields    *host.HTTPFields
	expected  *uint64
	written   uint64
	finished  bool
	readOnce  sync.Once
	closeOnce sync.Once
}

func (p *httpPump) Read(data []byte) (int, error) {
	n, err := p.reader.Read(data)
	if err == io.EOF {
		p.readOnce.Do(func() {
			if p.fields != nil {
				for _, field := range p.fields.Entries() {
					p.trailer.Add(field.Name, field.Value)
				}
			}
			p.body.transferred()
			p.Close()
		})
	}
	return n, err
}
func (p *httpPump) Close() error {
	p.closeOnce.Do(func() {
		p.body.mu.Lock()
		done := p.body.transmitted
		p.body.mu.Unlock()
		if !done {
			p.body.dispose(nil, httpErr("connection-terminated"))
		}
		p.reader.Close()
		p.body.release()
	})
	return nil
}
func (b *httpBody) pump(c *component.Caller, trailer http.Header, expected *uint64) (io.ReadCloser, error) {
	b.mu.Lock()
	external := b.reader != nil
	b.mu.Unlock()
	if external {
		b.acquire()
		return &httpExternalPump{body: b, trailer: trailer, expected: expected}, nil
	}
	if err := b.watchTrailers(c); err != nil {
		return nil, err
	}
	reader, writer := io.Pipe()
	pump := &httpPump{body: b, reader: reader, writer: writer, trailer: trailer, expected: expected}
	context.AfterFunc(b.ctx, func() { _ = reader.CloseWithError(httpErr("connection-terminated")) })
	b.mu.Lock()
	original, trailing, anchor := b.original, b.trailers, b.anchor
	b.mu.Unlock()
	var refs []*component.Resource
	if original != nil {
		refs = append(refs, original)
	}
	if trailing != nil {
		refs = append(refs, trailing)
	}
	b.acquire()
	var release sync.Once
	cleanup := func() {
		release.Do(func() {
			if !pump.finished {
				b.dispose(nil, httpErr("connection-terminated"))
				writer.CloseWithError(httpErr("connection-terminated"))
			}
			b.release()
		})
	}
	if err := pump.schedule(c, refs, original != nil, cleanup); err != nil {
		cleanup()
		return nil, err
	}
	b.acquire() // The external reader keeps ownership after scheduled writes finish.
	b.mu.Lock()
	b.original, b.trailers, b.anchor = nil, nil, nil
	b.mu.Unlock()
	if anchor != nil {
		_ = c.DropAsync(anchor)
	}
	return pump, nil
}
func (p *httpPump) fail(err error) {
	if err == nil {
		err = httpErr("connection-terminated")
	}
	p.finished = true
	p.body.dispose(nil, err)
	p.writer.CloseWithError(err)
}
func (p *httpPump) schedule(c *component.Caller, refs []*component.Resource, hasBody bool, cleanup func()) error {
	var next []*component.Resource
	var batch component.ReadResult
	var failure error
	transferred := false
	return c.Schedule(component.AsyncWork{Owned: refs, Start: func(c *component.Caller, owned []*component.Resource) (*component.HostCompletion, error) {
		next = owned
		if !hasBody {
			batch.Closed = true
			completion := component.NewHostCompletion()
			completion.Resolve(func(*component.Caller) (*component.Value, error) { return nil, nil })
			return completion, nil
		}
		limit := min(p.body.owner.network.io.config.MaxWriteBytes, p.body.owner.config.BodyBufferBytes)
		return c.AwaitReadChecked(p.body.ctx, owned[0], limit, func(batch component.ReadResult) error {
			if uint64(len(batch.Bytes)) > ^uint64(0)-p.written {
				return httpErr(p.body.sizeFailure)
			}
			size := p.written + uint64(len(batch.Bytes))
			if p.expected != nil && size > *p.expected {
				return httpSizeError(p.body.sizeFailure, size)
			}
			return nil
		}, func(_ *component.Caller, result component.ReadResult, err error) (*component.Value, error) {
			batch, failure = result, err
			return nil, nil
		})
	}, Done: func(c *component.Caller, cancelled bool) error {
		if cancelled || failure != nil {
			p.fail(httpTransportError(failure, "connection-terminated"))
			return nil
		}
		if uint64(len(batch.Bytes)) > ^uint64(0)-p.written {
			p.fail(httpErr(p.body.sizeFailure))
			return nil
		}
		p.written += uint64(len(batch.Bytes))
		if p.expected != nil && (p.written > *p.expected || batch.Closed && p.written != *p.expected) {
			p.fail(httpSizeError(p.body.sizeFailure, p.written))
			return nil
		}
		if err := p.scheduleWrite(c, next, hasBody, batch, cleanup); err != nil {
			return err
		}
		transferred = true
		return nil
	}, Cleanup: func() {
		if !transferred {
			cleanup()
		}
	}})
}
func (p *httpPump) scheduleWrite(c *component.Caller, refs []*component.Resource, hasBody bool, batch component.ReadResult, cleanup func()) error {
	var next []*component.Resource
	var failure error
	transferred := false
	return c.Schedule(component.AsyncWork{Owned: refs, Start: func(c *component.Caller, owned []*component.Resource) (*component.HostCompletion, error) {
		next = owned
		completion := component.NewHostCompletion()
		ctx := c.Context()
		if err := p.body.owner.worker(func() {
			defer func() {
				if recover() != nil {
					failure = httpErr("internal-error")
				}
				if failure != nil {
					p.writer.CloseWithError(failure)
					p.body.dispose(nil, failure)
				}
				if ctx.Err() != nil {
					completion.Cancelled()
				} else {
					completion.Resolve(func(*component.Caller) (*component.Value, error) { return nil, nil })
				}
			}()
			if len(batch.Bytes) != 0 {
				_, failure = p.writer.Write([]byte(batch.Bytes))
			}
			if failure == nil && batch.Closed {
				p.body.end(nil)
				p.body.mu.Lock()
				done := p.body.trailerDone
				p.body.mu.Unlock()
				if done != nil {
					select {
					case <-done:
					case <-p.body.ctx.Done():
						failure = p.body.ctx.Err()
					case <-ctx.Done():
						failure = ctx.Err()
					}
				}
				p.body.mu.Lock()
				p.fields = p.body.trailersValue
				if failure == nil {
					failure = p.body.failure
				}
				p.body.mu.Unlock()
				if failure == nil {
					failure = p.writer.Close()
				}
			}
		}); err != nil {
			return nil, err
		}
		return completion, nil
	}, Done: func(c *component.Caller, cancelled bool) error {
		if cancelled || failure != nil {
			p.fail(httpErr("connection-terminated"))
			return nil
		}
		if batch.Closed {
			p.finished = true
			return nil
		}
		if err := p.schedule(c, next, hasBody, cleanup); err != nil {
			return err
		}
		transferred = true
		return nil
	}, Cleanup: func() {
		if !transferred {
			cleanup()
		}
	}})
}

// External HTTP messages can be forwarded without entering a component callback
// from a network goroutine. Guest-backed messages use the scheduled pump above.
type httpExternalPump struct {
	body     *httpBody
	trailer  http.Header
	expected *uint64
	seen     uint64
	once     sync.Once
}

func (p *httpExternalPump) Read(data []byte) (int, error) {
	p.body.readerMu.Lock()
	defer p.body.readerMu.Unlock()
	n, err := p.body.readExternal(data)
	if uint64(n) > ^uint64(0)-p.seen {
		err = httpErr(p.body.sizeFailure)
	} else {
		p.seen += uint64(n)
	}
	if p.expected != nil && (p.seen > *p.expected || err == io.EOF && p.seen != *p.expected) {
		err = httpSizeError(p.body.sizeFailure, p.seen)
	}
	if err == io.EOF {
		p.body.mu.Lock()
		fields := p.body.trailersValue
		p.body.mu.Unlock()
		if fields != nil {
			for _, field := range fields.Entries() {
				p.trailer.Add(field.Name, field.Value)
			}
		}
		p.body.transferred()
		p.Close()
	} else if err != nil {
		p.body.end(err)
	}
	return n, err
}
func (p *httpExternalPump) Close() error {
	p.once.Do(func() {
		p.body.mu.Lock()
		done := p.body.eof
		p.body.mu.Unlock()
		if !done {
			p.body.dispose(nil, httpErr("connection-terminated"))
		}
		p.body.release()
	})
	return nil
}
