package preview2

import (
	"context"
	"errors"
	"io"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

func normalizeStreamError(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
		return io.ErrClosedPipe
	}
	return err
}
func (r *ioStream) read(ctx context.Context, length uint64, blocking bool) ([]byte, error) {
	if length > uint64(r.owner.config.MaxReadBytes) {
		length = uint64(r.owner.config.MaxReadBytes)
	}
	data := make([]byte, int(length))
	for {
		if err := r.owner.check(ctx); err != nil {
			return nil, err
		}
		r.readMu.Lock()
		r.mu.Lock()
		closed := r.closed
		r.mu.Unlock()
		if closed {
			r.readMu.Unlock()
			return nil, io.ErrClosedPipe
		}
		n, err := func() (int, error) { defer r.readMu.Unlock(); return r.source.input.Read(ctx, data) }()
		if n < 0 || n > len(data) {
			n = 0
			err = errors.New("invalid input read count")
		}
		if failed := r.owner.check(ctx); failed != nil {
			return nil, failed
		}
		wouldBlock := errors.Is(err, ErrWouldBlock)
		if err != nil && !wouldBlock {
			err = normalizeStreamError(err)
			r.mu.Lock()
			r.failLocked(err)
			r.mu.Unlock()
		}
		if n > 0 {
			return data[:n], nil
		}
		if err != nil && !wouldBlock {
			return nil, err
		}
		if !blocking || length == 0 {
			return data[:0], nil
		}
		if err = r.owner.poller.block(ctx, r.source.input); err != nil {
			return nil, err
		}
	}
}

func streamOK(value *component.Value) *component.Value {
	out := &component.Value{}
	if value != nil {
		out.Elements = []component.Value{*value}
	}
	return out
}
func (p *IO) streamError(c *component.Caller, r *ioStream, err error) (*component.Value, error) {
	if failed := p.check(c.Context()); failed != nil {
		return nil, failed
	}
	r.mu.Lock()
	cause, reported := r.failure, r.reported
	r.mu.Unlock()
	if cause == nil && (errors.Is(err, wasm.ErrTrap) || errors.Is(err, wasm.ErrLimit)) {
		return nil, err
	}
	failure := component.Value{Case: 1}
	if cause != nil && !reported {
		value, failed := p.errors.NewError(c, cause)
		if failed != nil {
			return nil, failed
		}
		failure = component.Value{Case: 0, Elements: []component.Value{value}}
		r.mu.Lock()
		r.reported = true
		r.mu.Unlock()
	}
	return &component.Value{Case: 1, Elements: []component.Value{failure}}, nil
}

func (p *IO) subscribe(c *component.Caller, r *ioStream) (*component.Value, error) {
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	if err := p.poller.begin(); err != nil {
		return nil, err
	}
	defer p.poller.active.Done()
	r.mu.Lock()
	if r.dropped {
		r.mu.Unlock()
		return nil, wasm.ErrClosed
	}
	r.children++
	r.mu.Unlock()
	var source PollSource = outputPoll{r}
	if r.source.input != nil {
		source = r.source.input
	}
	release := func() {
		r.mu.Lock()
		r.children--
		orphan := r.dropped && r.children == 0
		r.mu.Unlock()
		if orphan {
			p.drop(r.id)
		}
	}
	value, err := p.poller.publish(c, source, release)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
