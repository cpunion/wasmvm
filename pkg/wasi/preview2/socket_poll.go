//go:build linux || darwin

package preview2

import (
	"context"
	"golang.org/x/sys/unix"
	"os"
	"sync"

	"github.com/cpunion/wasmvm/internal/wasi/host"
)

// A readiness source starts at most one waiter. The waiter holds a descriptor
// lease, so cancellation/close cannot race descriptor reuse in the host poll.
type socketPoll struct {
	mu      sync.Mutex
	file    *socketFile
	read    bool
	ctx     context.Context
	cancel  context.CancelFunc
	changed chan struct{}
	waiter  chan struct{}
	closed  bool
	failure error
}

func newSocketPoll(file *socketFile, read bool) *socketPoll {
	ctx, cancel := context.WithCancel(file.ctx)
	return &socketPoll{file: file, read: read, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
}
func (s *socketPoll) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file.ctx.Err() != nil {
		s.failure = os.ErrClosed
	}
	if s.closed || s.failure != nil {
		return true, s.changed
	}
	fd, err := s.file.begin(s.ctx)
	if err != nil {
		s.failure = err
		return true, s.changed
	}
	_, _, ready, err := host.Ready(fd, s.read)
	if err == unix.EINTR {
		err = nil
		ready = false
	}
	if err != nil || ready {
		s.file.end()
		if err != nil {
			s.failure = err
			if s.file.ctx.Err() != nil {
				s.failure = os.ErrClosed
			}
		}
		return true, s.changed
	}
	if s.waiter != nil {
		s.file.end()
		return false, s.changed
	}
	done := make(chan struct{})
	s.waiter = done
	go func() {
		defer s.file.end()
		_, _, err := host.Poll(s.ctx, fd, s.read)
		s.mu.Lock()
		if err != nil && !s.closed {
			s.failure = err
			if s.file.ctx.Err() != nil {
				s.failure = os.ErrClosed
			}
		}
		close(s.changed)
		s.changed = make(chan struct{})
		s.waiter = nil
		close(done)
		s.mu.Unlock()
	}()
	return false, s.changed
}
func (s *socketPoll) Close() {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		s.cancel()
		close(s.changed)
		s.changed = make(chan struct{})
	}
	done := s.waiter
	s.mu.Unlock()
	if done != nil {
		<-done
	}
}
func (s *socketPoll) error() error { s.mu.Lock(); defer s.mu.Unlock(); return s.failure }
func (s *tcpSocket) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	state, op := s.state, s.operation
	s.mu.Unlock()
	if op == tcpConnectOperation {
		return s.write.Readiness()
	}
	if state == tcpListening && op == tcpNoOperation {
		return s.read.Readiness()
	}
	return true, nil
}
