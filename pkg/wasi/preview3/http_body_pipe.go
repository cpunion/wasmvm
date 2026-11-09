//go:build linux || darwin

package preview3

import (
	"io"
	"sync"
)

// httpBodyPipe retains one immutable, already bounded guest batch. The writer
// waits until the entire batch is consumed; partial external reads only copy
// bytes and do not wake the writer for each fragment. No guest-memory view or
// extra mutable batch copy is retained.
type httpBodyPipe struct {
	mu, writeMu sync.Mutex
	available   chan struct{}
	done        chan struct{}
	chunk       *httpBodyChunk
	closed      bool
	readErr     error
	writeErr    error
}

type httpBodyChunk struct {
	data     string
	read     int
	consumed bool
	done     chan struct{}
}

func newHTTPBodyPipe() *httpBodyPipe {
	return &httpBodyPipe{available: make(chan struct{}), done: make(chan struct{})}
}

func (p *httpBodyPipe) Read(out []byte) (int, error) {
	p.mu.Lock()
	for {
		if p.closed {
			err := error(io.ErrClosedPipe)
			if p.readErr == nil && p.writeErr != nil {
				err = p.writeErr
			}
			p.mu.Unlock()
			return 0, err
		}
		if chunk := p.chunk; chunk != nil {
			n := copy(out, chunk.data[chunk.read:])
			chunk.read += n
			if chunk.read == len(chunk.data) {
				chunk.data, chunk.consumed = "", true
				p.chunk = nil
				p.available = make(chan struct{})
				close(chunk.done)
			}
			p.mu.Unlock()
			return n, nil
		}
		available := p.available
		p.mu.Unlock()
		select {
		case <-available:
		case <-p.done:
		}
		p.mu.Lock()
	}
}

func (p *httpBodyPipe) WriteString(data string) (int, error) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	p.mu.Lock()
	if p.closed {
		err := p.writeCloseError()
		p.mu.Unlock()
		return 0, err
	}
	chunk := &httpBodyChunk{data: data, done: make(chan struct{})}
	p.chunk = chunk
	close(p.available)
	p.mu.Unlock()
	select {
	case <-chunk.done:
	case <-p.done:
	}
	p.mu.Lock()
	n := chunk.read
	var err error
	if !chunk.consumed {
		err = p.writeCloseError()
	}
	p.mu.Unlock()
	return n, err
}

// The caller holds mu. As with io.Pipe, a side's own close prevents further
// operations on that side; a peer's first close error propagates across it.
func (p *httpBodyPipe) writeCloseError() error {
	if p.writeErr == nil && p.readErr != nil {
		return p.readErr
	}
	return io.ErrClosedPipe
}

func (p *httpBodyPipe) CloseRead(err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil {
		err = io.ErrClosedPipe
	}
	if p.readErr == nil {
		p.readErr = err
	}
	p.close()
	return nil
}

func (p *httpBodyPipe) CloseWrite(err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil {
		err = io.EOF
	}
	if p.writeErr == nil {
		p.writeErr = err
	}
	p.close()
	return nil
}

func (p *httpBodyPipe) close() {
	if !p.closed {
		p.closed = true
		if p.chunk != nil {
			p.chunk.data = ""
			p.chunk = nil
		}
		close(p.done)
	}
}
