//go:build linux || darwin

package preview1

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"golang.org/x/sys/unix"
)

// Buffer is a finite in-memory Stream. Bytes remains usable after Close.
type Buffer struct {
	mu       sync.Mutex
	data     []byte
	position uint64
	maximum  uint64
	closed   bool
}

func NewBuffer(maximum uint64) *Buffer { return &Buffer{maximum: maximum} }
func BytesInput(data []byte) Stream {
	return &Buffer{data: append([]byte(nil), data...), maximum: uint64(len(data))}
}
func (b *Buffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.data...)
}
func (b *Buffer) Read(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, ErrBadf
	}
	n := copy(p, b.data[b.position:])
	b.position += uint64(n)
	if n == 0 && len(p) != 0 {
		return 0, io.EOF
	}
	return n, nil
}
func (b *Buffer) Write(ctx context.Context, p []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, ErrBadf
	}
	n := uint64(len(p))
	if n > b.maximum-uint64(len(b.data)) {
		n = b.maximum - uint64(len(b.data))
	}
	b.data = append(b.data, p[:int(n)]...)
	if n < uint64(len(p)) {
		return int(n), ErrNospc
	}
	return int(n), nil
}
func (b *Buffer) Poll(ctx context.Context, read bool) (uint64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0, false, ErrBadf
	}
	if read {
		n := uint64(len(b.data)) - b.position
		return n, n == 0, nil
	}
	return 0, false, nil
}
func (b *Buffer) Close() error { b.mu.Lock(); defer b.mu.Unlock(); b.closed = true; return nil }

// OutputFunc injects a context-aware output. It must be safe for concurrent
// calls and promptly honor cancellation; readiness is always reported writable.
type OutputFunc func(context.Context, []byte) (int, error)

func (f OutputFunc) Read(context.Context, []byte) (int, error)        { return 0, ErrBadf }
func (f OutputFunc) Write(ctx context.Context, p []byte) (int, error) { return f(ctx, p) }
func (f OutputFunc) Poll(ctx context.Context, read bool) (uint64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}
	if read {
		return 0, false, ErrBadf
	}
	return 0, false, nil
}
func (f OutputFunc) Close() error { return nil }

// FileStream takes exclusive ownership of f and makes its open-file description
// nonblocking. Callers must not keep aliases to that description. This is also
// the CLI adapter for pipes/stdio; no background read goroutine can outlive Close.
func FileStream(f *os.File) (Stream, error) {
	if f == nil {
		return nil, ErrBadf
	}
	if err := unix.SetNonblock(host.Descriptor(f), true); err != nil {
		f.Close()
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &fileStream{file: f, ctx: ctx, cancel: cancel}, nil
}

type fileStream struct {
	file    *os.File
	mu      sync.Mutex
	flags   uint16
	active  uint64
	closing bool
	ctx     context.Context
	cancel  context.CancelFunc
}

func (s *fileStream) enter(parent context.Context) (context.Context, uint16, func(), error) {
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return nil, 0, nil, ErrBadf
	}
	s.active++
	flags := s.flags
	s.mu.Unlock()
	ctx, cancel := context.WithCancel(parent)
	stop := context.AfterFunc(s.ctx, cancel)
	return ctx, flags, func() {
		stop()
		cancel()
		s.mu.Lock()
		s.active--
		closeNow := s.closing && s.active == 0
		s.mu.Unlock()
		if closeNow {
			s.file.Close()
		}
	}, nil
}
func (s *fileStream) Read(parent context.Context, b []byte) (int, error) {
	ctx, flags, finish, err := s.enter(parent)
	if err != nil {
		return 0, err
	}
	defer finish()
	if flags&8 != 0 {
		if err := s.file.Sync(); err != nil {
			return 0, err
		}
	}
	return fileIO(ctx, s.file, b, true, flags&4 != 0)
}
func (s *fileStream) Write(parent context.Context, b []byte) (int, error) {
	ctx, flags, finish, err := s.enter(parent)
	if err != nil {
		return 0, err
	}
	defer finish()
	n, err := fileIO(ctx, s.file, b, false, flags&4 != 0)
	if n > 0 && flags&(2|16) != 0 {
		if syncErr := s.file.Sync(); syncErr != nil {
			return 0, syncErr
		}
	}
	return n, err
}
func (s *fileStream) Poll(parent context.Context, read bool) (uint64, bool, error) {
	ctx, _, finish, err := s.enter(parent)
	if err != nil {
		return 0, false, err
	}
	defer finish()
	return host.Poll(ctx, host.Descriptor(s.file), read)
}
func (s *fileStream) Ready(parent context.Context, read bool) (uint64, bool, bool, error) {
	ctx, _, finish, err := s.enter(parent)
	if err != nil {
		return 0, false, false, err
	}
	defer finish()
	if err := ctx.Err(); err != nil {
		return 0, false, false, err
	}
	return host.Ready(host.Descriptor(s.file), read)
}
func (s *fileStream) Close() error {
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return nil
	}
	s.closing = true
	s.cancel()
	closeNow := s.active == 0
	s.mu.Unlock()
	if closeNow {
		return s.file.Close()
	}
	return nil
}

func (b *Buffer) SetFlags(flags uint16) error {
	if flags&^uint16(1|4) != 0 {
		return ErrNotsup
	}
	return nil
}
func (f OutputFunc) SetFlags(flags uint16) error {
	if flags&^uint16(1|4) != 0 {
		return ErrNotsup
	}
	return nil
}
func (s *fileStream) SetFlags(flags uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return ErrBadf
	}
	if flags&(2|8|16) != 0 {
		st, err := host.Fstat(s.file)
		if err != nil {
			return err
		}
		if st.Mode&unix.S_IFMT != unix.S_IFREG {
			return ErrNotsup
		}
	}
	fd := host.Descriptor(s.file)
	old, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		return err
	}
	old &^= unix.O_APPEND
	old |= unix.O_NONBLOCK
	if flags&1 != 0 {
		old |= unix.O_APPEND
	}
	if _, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, old); err != nil {
		return err
	}
	s.flags = flags
	return nil
}
func fileIO(ctx context.Context, f *os.File, b []byte, read, nonblock bool) (int, error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		var n int
		var err error
		if read {
			n, err = unix.Read(host.Descriptor(f), b)
		} else {
			n, err = unix.Write(host.Descriptor(f), b)
		}
		if err == unix.EINTR {
			continue
		}
		if err != unix.EAGAIN || nonblock {
			if n < 0 {
				if err == nil {
					return 0, ErrIo
				}
				n = 0
			}
			return n, err
		}
		if _, _, err = host.Poll(ctx, host.Descriptor(f), read); err != nil {
			return 0, err
		}
	}
}

func (b *Buffer) Ready(ctx context.Context, read bool) (uint64, bool, bool, error) {
	n, h, err := b.Poll(ctx, read)
	return n, h, err == nil, err
}
func (f OutputFunc) Ready(ctx context.Context, read bool) (uint64, bool, bool, error) {
	n, h, err := f.Poll(ctx, read)
	return n, h, err == nil, err
}
