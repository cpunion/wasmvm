package preview2

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

// httpBodyBuffer provides bounded streaming without a mandatory worker per IO
// port. A transport or incoming-body pump owns the blocking side. The HTTP
// owner closes buffers explicitly when revoking requests or shutting down.
type httpBodyBuffer struct {
	mu                           sync.Mutex
	data                         []byte
	head, count                  int
	changed                      chan struct{}
	finished, closed, discarding bool
	failure                      error
	written                      uint64
}

func newHTTPBodyBuffer(capacity uint32) (*httpBodyBuffer, error) {
	if capacity == 0 || capacity > 1<<20 {
		return nil, wasm.ErrLimit
	}
	return &httpBodyBuffer{data: make([]byte, int(capacity)), changed: make(chan struct{})}, nil
}
func (b *httpBodyBuffer) notify() { close(b.changed); b.changed = make(chan struct{}) }
func (b *httpBodyBuffer) write(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.finished || b.closed {
		return 0, io.ErrClosedPipe
	}
	if len(data) == 0 {
		return 0, nil
	}
	n := len(data)
	if !b.discarding {
		n = min(n, len(b.data)-b.count)
		if n == 0 {
			return 0, ErrWouldBlock
		}
	}
	if uint64(n) > ^uint64(0)-b.written {
		return 0, wasm.ErrLimit
	}
	if !b.discarding {
		tail := (b.head + b.count) % len(b.data)
		first := min(n, len(b.data)-tail)
		copy(b.data[tail:tail+first], data[:first])
		copy(b.data[:n-first], data[first:n])
		b.count += n
	}
	b.written += uint64(n)
	b.notify()
	return n, nil
}
func (b *httpBodyBuffer) read(ctx context.Context, out []byte, blocking bool) (int, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		b.mu.Lock()
		var n int
		var err error
		switch {
		case len(out) == 0:
		case b.closed:
			err = io.EOF
		case b.count != 0:
			n = min(len(out), b.count)
			first := min(n, len(b.data)-b.head)
			copy(out[:first], b.data[b.head:b.head+first])
			copy(out[first:n], b.data[:n-first])
			b.head = (b.head + n) % len(b.data)
			b.count -= n
			b.notify()
		case b.failure != nil:
			err = b.failure
		case b.finished:
			err = io.EOF
		default:
			err = ErrWouldBlock
		}
		changed := b.changed
		b.mu.Unlock()
		if !blocking || !errors.Is(err, ErrWouldBlock) {
			return n, err
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}
func (b *httpBodyBuffer) readiness(read, flushing bool) (bool, <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ready := b.closed || b.failure != nil
	if read {
		ready = ready || b.finished || b.count != 0
	} else if flushing {
		ready = ready || b.count == 0
	} else {
		ready = ready || b.finished || b.discarding || b.count < len(b.data)
	}
	return ready, b.changed
}
func (b *httpBodyBuffer) flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return io.ErrClosedPipe
	}
	if b.failure != nil {
		return b.failure
	}
	if b.count != 0 {
		return ErrWouldBlock
	}
	return nil
}
func (b *httpBodyBuffer) finish(err error) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.finished || b.closed {
		return io.ErrClosedPipe
	}
	b.finished = true
	b.failure = err
	b.notify()
	return nil
}

// Discarding permits an incoming-body owner to drain unread bytes for trailers
// after its IO child has dropped. No data is retained once draining begins.
func (b *httpBodyBuffer) discard() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.discarding {
		b.discarding = true
		b.head = 0
		b.count = 0
		b.notify()
	}
}
func (b *httpBodyBuffer) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		b.head = 0
		b.count = 0
		b.notify()
	}
}
