//go:build linux || darwin

package preview2

import (
	"context"
	"os"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

// socketFile grants syscall leases while open. Close cancels readiness work and
// joins leases before physically closing, preventing descriptor-number reuse
// from redirecting an in-flight poll/read/write to an unrelated host object.
type socketFile struct {
	mu       sync.Mutex
	file     *os.File
	fd       int
	ctx      context.Context
	cancel   context.CancelFunc
	closed   bool
	active   sync.WaitGroup
	done     chan struct{}
	closeErr error
}

func newSocketFile(ctx context.Context, family uint32, kind int) (*socketFile, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	domain := unix.AF_INET
	switch family {
	case 0:
	case 1:
		domain = unix.AF_INET6
	default:
		return nil, &NetworkError{Code: NetworkNotSupported}
	}
	if kind != unix.SOCK_STREAM && kind != unix.SOCK_DGRAM {
		return nil, &NetworkError{Code: NetworkInvalidArgument}
	}
	// Darwin cannot add SOCK_CLOEXEC at socket creation. ForkLock covers the
	// creation/fcntl window on both platforms, without inheritable publication.
	syscall.ForkLock.RLock()
	fd, err := unix.Socket(domain, kind, 0)
	if err == nil {
		unix.CloseOnExec(fd)
		err = unix.SetNonblock(fd, true)
		if err == nil && family == 1 {
			err = unix.SetsockoptInt(fd, unix.IPPROTO_IPV6, unix.IPV6_V6ONLY, 1)
		}
		if err != nil {
			unix.Close(fd)
		}
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		return nil, &NetworkError{Code: socketError(err), Cause: err}
	}
	file := os.NewFile(uintptr(fd), "WASI socket")
	child, cancel := context.WithCancel(ctx)
	return &socketFile{file: file, fd: fd, ctx: child, cancel: cancel, done: make(chan struct{})}, nil
}
func (s *socketFile) begin(ctx context.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.ctx.Err() != nil {
		return 0, os.ErrClosed
	}
	s.active.Add(1)
	return s.fd, nil
}
func (s *socketFile) end() { s.active.Done() }
func (s *socketFile) Close() {
	if s == nil || s.done == nil {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.cancel()
	s.mu.Unlock()
	go func() { s.active.Wait(); s.closeErr = s.file.Close(); close(s.done) }()
}
func (s *socketFile) WaitClosed(ctx context.Context) error {
	if s == nil || s.done == nil {
		return os.ErrInvalid
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-s.done:
		return s.closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
