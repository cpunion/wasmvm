package preview2

import (
	"context"
	"errors"
	"io"
	"sync"
)

// ErrWouldBlock reports unavailable immediate progress. Readiness must then
// supply a coherent notification; it is not a permanent stream error.
var ErrWouldBlock = errors.New("WASI IO temporarily unavailable")

// InputSource methods are thread safe. Read returns promptly with available
// bytes, EOF, an error, or ErrWouldBlock; it never waits for future input.
// Close is idempotent, unblocks operations and releases owned host resources.
type InputSource interface {
	PollSource
	Read(context.Context, []byte) (int, error)
	Close() error
}

// OutputSource supports prompt partial writes. Flush returns ErrWouldBlock
// while output is pending. Readiness is writable/flush readiness, including
// permanent errors. Implementations observe cancellation and concurrent Close.
// Write must not retain the supplied byte slice after returning. During a
// pending Flush, Readiness must wait for flush completion before reporting ready.
type OutputSource interface {
	PollSource
	Write(context.Context, []byte) (int, error)
	Flush(context.Context) error
	Close() error
}

type BytesInput struct {
	mu     sync.Mutex
	data   []byte
	at     int
	closed bool
}

func NewBytesInput(data []byte) *BytesInput { return &BytesInput{data: append([]byte(nil), data...)} }
func (b *BytesInput) Read(ctx context.Context, out []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, io.ErrClosedPipe
	}
	if b.at == len(b.data) {
		return 0, io.EOF
	}
	n := copy(out, b.data[b.at:])
	b.at += n
	return n, nil
}
func (b *BytesInput) Readiness() (bool, <-chan struct{}) { return true, nil }
func (b *BytesInput) Close() error {
	b.mu.Lock()
	b.closed = true
	b.data = nil
	b.mu.Unlock()
	return nil
}

// OutputBuffer is a bounded host-drained byte sink. Draining signals writers;
// its returned bytes are detached from future writes. It never inherits stdio.
type OutputBuffer struct {
	mu      sync.Mutex
	data    []byte
	maximum uint32
	changed chan struct{}
	closed  bool
}

func NewOutputBuffer(maximum uint32) *OutputBuffer {
	if maximum == 0 {
		maximum = 1 << 20
	}
	return &OutputBuffer{maximum: maximum, changed: make(chan struct{})}
}
func (b *OutputBuffer) initializeLocked() {
	if b.changed == nil {
		b.changed = make(chan struct{})
		if b.maximum == 0 {
			b.maximum = 1 << 20
		}
	}
}
func (b *OutputBuffer) notify() {
	b.initializeLocked()
	close(b.changed)
	b.changed = make(chan struct{})
}
func (b *OutputBuffer) Write(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initializeLocked()
	if b.closed {
		return 0, io.ErrClosedPipe
	}
	free := uint64(b.maximum) - uint64(len(b.data))
	if len(data) > 0 && free == 0 {
		return 0, ErrWouldBlock
	}
	if uint64(len(data)) > free {
		data = data[:free]
	}
	b.data = append(b.data, data...)
	return len(data), nil
}
func (b *OutputBuffer) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initializeLocked()
	if b.closed {
		return io.ErrClosedPipe
	}
	return nil
}
func (b *OutputBuffer) Readiness() (bool, <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initializeLocked()
	return b.closed || uint64(len(b.data)) < uint64(b.maximum), b.changed
}
func (b *OutputBuffer) Drain() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := b.data
	b.data = nil
	b.notify()
	return out
}
func (b *OutputBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		b.notify()
	}
	return nil
}

type discardOutput struct {
	mu     sync.Mutex
	closed bool
}

func (d *discardOutput) Write(ctx context.Context, data []byte) (int, error) {
	if err := d.Flush(ctx); err != nil {
		return 0, err
	}
	return len(data), nil
}
func (d *discardOutput) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return io.ErrClosedPipe
	}
	return nil
}
func (d *discardOutput) Readiness() (bool, <-chan struct{}) { return true, nil }
func (d *discardOutput) Close() error                       { d.mu.Lock(); d.closed = true; d.mu.Unlock(); return nil }
