//go:build linux || darwin

package preview2

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

type tcpSource struct {
	mu            sync.Mutex
	closeOnce     sync.Once
	socket        *tcpSocket
	poll          *socketPoll
	input, closed bool
}

func tcpStreamError(err error) error {
	if err == nil || errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, unix.EAGAIN) {
		return ErrWouldBlock
	}
	return &NetworkError{Code: socketError(err), Cause: err}
}
func (s *tcpSource) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return true, nil
	}
	return s.poll.Readiness()
}
func (s *tcpSource) Read(ctx context.Context, out []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, io.EOF
	}
	if failed := s.poll.error(); failed != nil {
		return 0, tcpStreamError(failed)
	}
	fd, err := s.socket.file.begin(ctx)
	if err != nil {
		return 0, tcpStreamError(err)
	}
	defer s.socket.file.end()
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		n, err := unix.Read(fd, out)
		if err == unix.EINTR {
			continue
		}
		if n < 0 {
			n = 0
		}
		if n == 0 && err == nil && len(out) > 0 {
			err = io.EOF
		}
		return n, tcpStreamError(err)
	}
}
func (s *tcpSource) Write(ctx context.Context, data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, io.ErrClosedPipe
	}
	if failed := s.poll.error(); failed != nil {
		return 0, tcpStreamError(failed)
	}
	fd, err := s.socket.file.begin(ctx)
	if err != nil {
		return 0, tcpStreamError(err)
	}
	defer s.socket.file.end()
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		n, err := unix.SendmsgN(fd, data, nil, nil, unix.MSG_NOSIGNAL)
		if err == unix.EINTR {
			continue
		}
		if n < 0 {
			n = 0
		}
		return n, tcpStreamError(err)
	}
}
func (s *tcpSource) Flush(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.closed {
		return io.ErrClosedPipe
	}
	fd, err := s.socket.file.begin(ctx)
	if err != nil {
		return tcpStreamError(err)
	}
	_ = fd
	s.socket.file.end()
	return tcpStreamError(s.poll.error())
}
func (s *tcpSource) shutdown() { s.mu.Lock(); s.closed = true; s.mu.Unlock(); s.poll.Close() }
func (s *tcpSource) Close() error {
	// once protects the parent lease independently of shutdown, which may happen
	// earlier through the socket's explicit shutdown method.
	s.closeOnce.Do(func() {
		s.shutdown()
		fd, err := s.socket.file.begin(context.Background())
		if err == nil {
			how := unix.SHUT_WR
			if s.input {
				how = unix.SHUT_RD
			}
			unix.Shutdown(fd, how)
			s.socket.file.end()
		}
		s.socket.owner.release(s.socket)
	})
	return nil
}
func (p *TCP) streams(c *component.Caller, s *tcpSocket) (component.Value, error) {
	if s.streams {
		return component.Value{}, socketCode(NetworkInvalidState)
	}
	if err := p.child(s); err != nil {
		return component.Value{}, err
	}
	input := &tcpSource{socket: s, poll: s.read, input: true}
	if err := p.child(s); err != nil {
		input.Close()
		return component.Value{}, err
	}
	output := &tcpSource{socket: s, poll: s.write}
	in, err := p.network.io.NewInput(c, input)
	if err != nil {
		input.Close()
		output.Close()
		return component.Value{}, err
	}
	out, err := p.network.io.NewOutput(c, output)
	if err != nil {
		output.Close()
		c.DropResource(in.Resource)
		return component.Value{}, err
	}
	s.streams = true
	s.input, s.output = input, output
	return component.Value{Elements: []component.Value{in, out}}, nil
}
