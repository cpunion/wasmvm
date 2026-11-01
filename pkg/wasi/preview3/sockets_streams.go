//go:build linux || darwin

package preview3

import (
	"context"
	"io"
	"net/netip"
	"os"
	"sync"
	"syscall"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

type netStream struct {
	mu                                           sync.Mutex
	owner                                        *Sockets
	socket                                       *netSocket
	op                                           *ioOperation
	ctx                                          context.Context
	cancel                                       context.CancelFunc
	changed                                      chan struct{}
	input, listening, closed, watching, shutdown bool
	failure                                      error
}

func (p *Sockets) newStream(s *netSocket, input, listening bool, failure error) (*netStream, error) {
	op, err := p.io.operationFor(netCompletion, netCompletionValue, netCases["invalid-state"]+1)
	if err != nil {
		return nil, err
	}
	if err = s.acquire(); err != nil {
		op.finish(netCompletion(err))
		return nil, err
	}
	ctx, cancel := context.WithCancel(p.ctx)
	stream := &netStream{owner: p, socket: s, op: op, ctx: ctx, cancel: cancel, changed: make(chan struct{}), input: input, listening: listening, failure: failure, shutdown: !listening && failure == nil}
	p.mu.Lock()
	closed := p.closed
	if !closed {
		p.streams[stream] = struct{}{}
	}
	p.mu.Unlock()
	op.setRelease(func() { stream.Close() })
	if closed {
		op.finish(netCases["invalid-state"] + 1)
		return nil, os.ErrClosed
	}
	return stream, nil
}
func (s *netStream) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.cancel()
	close(s.changed)
	s.changed = make(chan struct{})
	s.mu.Unlock()
	p := s.owner
	p.mu.Lock()
	delete(p.streams, s)
	closed := p.closed
	p.mu.Unlock()
	if closed || !s.shutdown {
		s.socket.release()
		return nil
	}
	// Half-close outside the VM, retaining the object until this queued work has
	// run. At most two data streams exist per physical TCP socket.
	p.jobs <- func() {
		p.mu.Lock()
		closed := s.socket.closed
		p.mu.Unlock()
		if !closed {
			how := unix.SHUT_WR
			if s.input {
				how = unix.SHUT_RD
			}
			_ = unix.Shutdown(s.socket.fd, how)
		}
		s.socket.release()
	}
	return nil
}
func (s *netStream) check(ctx context.Context) error {
	if err := s.owner.check(ctx); err != nil {
		return err
	}
	if s.ctx.Err() != nil {
		return os.ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failure
}
func (s *netStream) begin(ctx context.Context) error {
	if err := s.check(ctx); err != nil {
		return err
	}
	if err := s.owner.begin(); err != nil {
		return err
	}
	if err := s.socket.acquire(); err != nil {
		s.owner.end()
		return err
	}
	return nil
}
func (s *netStream) end() { s.socket.release(); s.owner.end() }
func (s *netStream) Read(ctx context.Context, data []byte) (int, error) {
	if err := s.begin(ctx); err != nil {
		return 0, err
	}
	defer s.end()
	for {
		if err := s.check(ctx); err != nil {
			return 0, err
		}
		n, err := unix.Read(s.socket.fd, data)
		if err == unix.EINTR {
			continue
		}
		if n < 0 {
			n = 0
		}
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			err = ErrWouldBlock
		}
		if n == 0 && err == nil && len(data) > 0 {
			err = io.EOF
		}
		return n, err
	}
}
func (s *netStream) Write(ctx context.Context, data []byte) (int, error) {
	if err := s.begin(ctx); err != nil {
		return 0, err
	}
	defer s.end()
	for {
		if err := s.check(ctx); err != nil {
			return 0, err
		}
		n, err := unix.SendmsgN(s.socket.fd, data, nil, nil, unix.MSG_NOSIGNAL)
		if err == unix.EINTR {
			continue
		}
		if n < 0 {
			n = 0
		}
		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			err = ErrWouldBlock
		}
		return n, err
	}
}
func (s *netStream) Flush(ctx context.Context) error {
	if err := s.begin(ctx); err != nil {
		return err
	}
	defer s.end()
	err := unix.Shutdown(s.socket.fd, unix.SHUT_WR)
	if err == unix.ENOTCONN {
		return nil
	}
	return err
}
func (s *netStream) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.failure != nil {
		return true, s.changed
	}
	if s.watching {
		return false, s.changed
	}
	if err := s.owner.begin(); err != nil {
		s.failure = err
		return true, s.changed
	}
	if err := s.socket.acquire(); err != nil {
		s.owner.end()
		s.failure = err
		return true, s.changed
	}
	_, _, ready, err := host.Ready(s.socket.fd, s.input)
	if err == unix.EINTR {
		err = nil
		ready = false
	}
	if ready || err != nil {
		s.socket.release()
		s.owner.end()
		s.failure = err
		return true, s.changed
	}
	s.watching = true
	go func() {
		defer s.owner.workers.Done()
		_, _, err := host.Poll(s.ctx, s.socket.fd, s.input)
		s.socket.release()
		s.owner.releaseSlot()
		s.mu.Lock()
		s.watching = false
		if err != nil && !s.closed {
			s.failure = err
		}
		close(s.changed)
		s.changed = make(chan struct{})
		s.mu.Unlock()
	}()
	return false, s.changed
}
func (p *Sockets) tcpStream(c *component.Caller, s *netSocket, name string, args []component.Value) (*component.Value, error) {
	input := name == "receive"
	s.mu.Lock()
	var failure error
	if s.state != 3 || input && s.received || !input && s.sent {
		failure = netError("invalid-state")
	} else if input {
		s.received = true
	} else {
		s.sent = true
	}
	s.mu.Unlock()
	stream, err := p.newStream(s, input, false, failure)
	if err != nil {
		return nil, err
	}
	if input {
		_, st := c.FunctionInfo().Result().Field(0)
		_, ft := c.FunctionInfo().Result().Field(1)
		return p.io.inputWithOperation(c, stream.op, stream, st, ft, failure)
	}
	return p.io.outputWithOperation(c, []component.Value{args[1]}, stream, stream.op, failure)
}
func (p *Sockets) listen(c *component.Caller, s *netSocket) (*component.Value, error) {
	s.mu.Lock()
	initial := s.state == 0
	allowed := s.state <= 1
	s.mu.Unlock()
	if !allowed {
		return netResult(nil, netError("invalid-state")), nil
	}
	if initial {
		ip := netip.IPv4Unspecified()
		if s.family == 1 {
			ip = netip.IPv6Unspecified()
		}
		a := SocketAddress{Address: netip.AddrPortFrom(ip, 0)}
		if err := p.authorize(c.Context(), a, p.config.AllowBind); err != nil {
			return netResult(nil, err), nil
		}
	}
	s.mu.Lock()
	if s.state > 1 {
		s.mu.Unlock()
		return netResult(nil, netError("invalid-state")), nil
	}
	var err error
	if s.state == 0 {
		ip := netip.IPv4Unspecified()
		if s.family == 1 {
			ip = netip.IPv6Unspecified()
		}
		err = unix.SetsockoptInt(s.fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err == nil {
			err = host.SocketBind(s.fd, SocketAddress{Address: netip.AddrPortFrom(ip, 0)})
		}
		if err == nil {
			s.state = 1
		}
	}
	if err == nil {
		err = unix.Listen(s.fd, 128)
	}
	if err == nil {
		s.state = 4
	}
	s.mu.Unlock()
	if err != nil {
		return netResult(nil, err), nil
	}
	stream, err := p.newStream(s, true, true, nil)
	if err != nil {
		return nil, err
	}
	_, st := c.FunctionInfo().Result().Case(0)
	v, err := c.NewSource(st, component.AsyncSource{Read: stream.accept, Close: func() { stream.op.stop(); stream.op.finish(0) }})
	if err != nil {
		stream.op.finish(netCompletion(err))
		return nil, err
	}
	return netResult(&v, nil), nil
}

type netAcceptBatch struct {
	mu      sync.Mutex
	closed  bool
	sockets []*netSocket
}

func (b *netAcceptBatch) close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	sockets := b.sockets
	b.sockets = nil
	b.mu.Unlock()
	for _, s := range sockets {
		s.release()
	}
}
func (b *netAcceptBatch) add(s *netSocket) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		s.release()
		return
	}
	b.sockets = append(b.sockets, s)
	b.mu.Unlock()
}
func (s *netStream) accept(c *component.Caller, limit uint32) (*component.ReadCompletion, error) {
	if err := s.begin(c.Context()); err != nil {
		return nil, err
	}
	callCtx := c.Context()
	ctx, cancel := context.WithCancel(callCtx)
	stop := context.AfterFunc(s.ctx, cancel)
	batch := &netAcceptBatch{}
	completion := component.NewReadCompletionWithCleanup(func() { cancel(); stop(); batch.close() })
	go func() {
		defer s.owner.workers.Done()
		var accepted *netSocket
		var err error
		defer func() {
			panicked := recover()
			if accepted != nil {
				batch.add(accepted)
			}
			s.socket.release()
			s.owner.releaseSlot()
			if callCtx.Err() != nil && accepted == nil {
				completion.Cancelled()
				return
			}
			completion.Resolve(func(fresh *component.Caller) (component.ReadResult, error) {
				if panicked != nil {
					panic(panicked)
				}
				if err != nil {
					return component.ReadResult{}, err
				}
				if accepted == nil {
					return component.ReadResult{}, nil
				}
				v, err := s.owner.publish(fresh, accepted)
				if err != nil {
					return component.ReadResult{}, err
				}
				return component.ReadResult{Values: []component.Value{v}}, nil
			})
		}()
		for {
			if err = s.check(ctx); err != nil {
				return
			}
			if limit == 0 {
				_, _, err = host.Poll(ctx, s.socket.fd, true)
				return
			}
			if err = s.owner.reserve(); err != nil {
				return
			}
			syscall.ForkLock.RLock()
			fd, _, failed := unix.Accept(s.socket.fd)
			if failed == nil {
				unix.CloseOnExec(fd)
				failed = unix.SetNonblock(fd, true)
				if failed != nil {
					unix.Close(fd)
				}
			}
			syscall.ForkLock.RUnlock()
			if failed == nil {
				s.socket.mu.Lock()
				for name, value := range s.socket.inherited {
					if _, failed = host.SocketOption(fd, s.socket.family, name, value); failed != nil {
						break
					}
				}
				s.socket.mu.Unlock()
				if failed != nil {
					unix.Close(fd)
					s.owner.unreserve()
					err = failed
					return
				}
				accepted = s.owner.register(fd, s.socket.family, false, 3)
				return
			}
			s.owner.unreserve()
			if failed == unix.EINTR {
				continue
			}
			if failed != unix.EAGAIN && failed != unix.EWOULDBLOCK && failed != unix.ECONNABORTED && failed != unix.ENETDOWN && failed != unix.EPROTO && failed != unix.ENOPROTOOPT && failed != unix.EHOSTUNREACH && failed != unix.EOPNOTSUPP && failed != unix.ENETUNREACH {
				err = failed
				return
			}
			if _, _, err = host.Poll(ctx, s.socket.fd, true); err != nil {
				return
			}
		}
	}()
	return completion, nil
}
