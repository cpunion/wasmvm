package preview2

import (
	"context"
	"errors"
	"io"

	"github.com/cpunion/wasmvm/pkg/wasm"
)

func waitSource(ctx context.Context, source PollSource) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ready, changed := source.Readiness()
		if ready {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}
func (r *ioStream) failLocked(err error) {
	r.closed = true
	r.permit = 0
	r.checked = false
	r.flushing = false
	r.queue = nil
	if err != io.EOF && err != io.ErrClosedPipe {
		r.failure = err
	}
	r.notifyLocked()
}
func (r *ioStream) startWorkerLocked() {
	if r.running || r.closed {
		return
	}
	r.running = true
	r.workerDone = make(chan struct{})
	r.owner.workers.Add(1)
	go r.runOutput()
}
func (r *ioStream) runOutput() {
	defer r.owner.workers.Done()
	r.mu.Lock()
	done := r.workerDone
	r.mu.Unlock()
	defer close(done)
	completed := false
	defer func() {
		if !completed {
			_ = recover()
			r.mu.Lock()
			r.failLocked(errors.New("output source panicked"))
			r.mu.Unlock()
		}
		r.mu.Lock()
		r.running = false
		if !r.dropped && !r.closed && r.ctx.Err() == nil && (len(r.queue) != 0 || r.flushing) {
			r.startWorkerLocked()
		}
		r.notifyLocked()
		dropped := r.dropped
		r.mu.Unlock()
		if dropped {
			r.owner.finalize(r.id)
		}
	}()
	r.pumpOutput()
	completed = true
}
func (r *ioStream) pumpOutput() {
	for {
		r.mu.Lock()
		if r.closed || r.ctx.Err() != nil {
			r.mu.Unlock()
			return
		}
		data := r.queue
		flush := len(data) == 0 && r.flushing
		if len(data) == 0 && !flush {
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()
		if len(data) != 0 {
			n, err := r.source.output.Write(r.ctx, data)
			if n < 0 || n > len(data) {
				n = 0
				err = errors.New("invalid output write count")
			}
			wouldBlock := errors.Is(err, ErrWouldBlock)
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				err = io.ErrClosedPipe
			}
			r.mu.Lock()
			if r.closed {
				r.mu.Unlock()
				return
			}
			if n != 0 {
				r.queue = r.queue[n:]
				r.notifyLocked()
			}
			if err != nil && !wouldBlock {
				r.failLocked(err)
				r.mu.Unlock()
				return
			}
			r.mu.Unlock()
			if n == 0 {
				if err = waitSource(r.ctx, r.source.output); err != nil {
					return
				}
			}
		} else {
			err := r.source.output.Flush(r.ctx)
			wouldBlock := errors.Is(err, ErrWouldBlock)
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				err = io.ErrClosedPipe
			}
			r.mu.Lock()
			if r.closed {
				r.mu.Unlock()
				return
			}
			if err == nil {
				r.flushing = false
				r.notifyLocked()
				r.mu.Unlock()
				continue
			}
			if !wouldBlock {
				r.failLocked(err)
				r.mu.Unlock()
				return
			}
			r.mu.Unlock()
			if err = waitSource(r.ctx, r.source.output); err != nil {
				return
			}
		}
	}
}

func (r *ioStream) checkWrite() (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	free := uint64(0)
	if !r.flushing {
		free = uint64(r.owner.config.MaxWriteBytes) - uint64(len(r.queue))
	}
	r.checked = true
	r.permit = free
	return free, nil
}
func (r *ioStream) write(data string) error { return r.enqueue(data, 0) }
func (r *ioStream) enqueue(data string, zeroes uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return io.ErrClosedPipe
	}
	if !r.checked || uint64(len(data)) > r.permit || zeroes > r.permit-uint64(len(data)) {
		return wasm.ErrTrap
	}
	r.checked = false
	r.permit = 0
	r.queue = append(r.queue, data...)
	if zeroes != 0 {
		r.queue = append(r.queue, make([]byte, int(zeroes))...)
	}
	if len(data) != 0 || zeroes != 0 {
		r.startWorkerLocked()
	}
	r.notifyLocked()
	return nil
}
func (r *ioStream) flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return io.ErrClosedPipe
	}
	r.flushing = true
	r.checked = false
	r.permit = 0
	r.startWorkerLocked()
	r.notifyLocked()
	return nil
}
func (r *ioStream) waitFlush(ctx context.Context) error {
	for {
		if err := r.owner.check(ctx); err != nil {
			return err
		}
		r.mu.Lock()
		closed, flushing, changed := r.closed, r.flushing, r.changed
		r.mu.Unlock()
		if closed {
			return io.ErrClosedPipe
		}
		if !flushing {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.owner.ctx.Done():
			return wasm.ErrClosed
		case <-changed:
		}
	}
}

// Blocking writes can drive the same prompt backend directly. Queue workers
// remain necessary for nonblocking writes; no goroutine handoff is needed when
// this operation can complete on the active host call.
func (r *ioStream) blockingWrite(ctx context.Context, data string) error {
	r.mu.Lock()
	pending := r.running || len(r.queue) != 0 || r.flushing
	r.checked = false
	r.permit = 0
	r.mu.Unlock()
	if pending {
		if err := r.flush(); err != nil {
			return err
		}
		if err := r.waitFlush(ctx); err != nil {
			return err
		}
	}
	for len(data) != 0 {
		if err := r.owner.check(ctx); err != nil {
			return err
		}
		n, err := r.source.output.Write(ctx, []byte(data))
		if n < 0 || n > len(data) {
			n = 0
			err = errors.New("invalid output write count")
		}
		if failed := r.owner.check(ctx); failed != nil {
			return failed
		}
		data = data[n:]
		if err != nil && !errors.Is(err, ErrWouldBlock) {
			err = normalizeStreamError(err)
			r.mu.Lock()
			r.failLocked(err)
			r.mu.Unlock()
			return err
		}
		if n == 0 {
			if err = r.owner.poller.block(ctx, r.source.output); err != nil {
				return err
			}
		}
	}
	for {
		if err := r.owner.check(ctx); err != nil {
			return err
		}
		err := r.source.output.Flush(ctx)
		if err == nil {
			return r.owner.check(ctx)
		}
		if !errors.Is(err, ErrWouldBlock) {
			err = normalizeStreamError(err)
			r.mu.Lock()
			r.failLocked(err)
			r.mu.Unlock()
			return err
		}
		if err = r.owner.poller.block(ctx, r.source.output); err != nil {
			return err
		}
	}
}

type outputPoll struct{ stream *ioStream }

func (p outputPoll) Readiness() (bool, <-chan struct{}) {
	r := p.stream
	r.mu.Lock()
	defer r.mu.Unlock()
	ready := r.closed || !r.flushing && uint64(len(r.queue)) < uint64(r.owner.config.MaxWriteBytes)
	return ready, r.changed
}

func (r *ioStream) splice(ctx context.Context, src *ioStream, length uint64, blocking bool) (uint64, *ioStream, error) {
	for {
		if err := r.owner.check(ctx); err != nil {
			return 0, r, err
		}
		permit, err := r.checkWrite()
		if err != nil {
			return 0, r, err
		}
		if blocking && length != 0 && permit == 0 {
			if err = r.owner.poller.block(ctx, outputPoll{r}); err != nil {
				return 0, r, err
			}
			continue
		}
		count := length
		if count > permit {
			count = permit
		}
		data, err := src.read(ctx, count, blocking)
		if err != nil {
			return 0, src, err
		}
		if err = r.write(string(data)); err != nil {
			return 0, r, err
		}
		return uint64(len(data)), nil, nil
	}
}
