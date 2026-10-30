//go:build linux || darwin

package preview2

import (
	"context"
	"os"
	"syscall"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func (p *TCP) execute(c *component.Caller, s *tcpSocket, name string, args []component.Value) (value component.Value, hasValue bool, err error) {
	hasValue = true
	if name == "start-bind" || name == "start-connect" {
		if _, err = c.ResourceRep(p.network.networkType, args[1].Resource); err != nil {
			return
		}
		var address SocketAddress
		address, err = parseSocketAddress(args[2])
		if err != nil {
			return
		}
		err = s.checkAddress(address, name == "start-connect")
		if err != nil {
			return
		}
		// Validate state before calling embedder policy, then recheck it afterwards.
		s.mu.Lock()
		err = s.canStart(name)
		s.mu.Unlock()
		if err != nil {
			return
		}
		if err = p.authorize(c.Context(), address, name == "start-connect"); err != nil {
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if err = s.canStart(name); err != nil {
			return
		}
		var fd int
		fd, err = s.file.begin(c.Context())
		if err != nil {
			return
		}
		defer s.file.end()
		if name == "start-bind" {
			err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			if err == nil {
				err = socketBind(fd, address)
			}
			if err == nil {
				s.state = tcpBound
				s.operation = tcpBindOperation
			}
		} else {
			err = socketConnect(fd, address)
			if err == nil || err == unix.EINPROGRESS || err == unix.EALREADY {
				s.state = tcpConnecting
				s.operation = tcpConnectOperation
				err = nil
			} else {
				s.state = tcpFailed
				s.failure = err
			}
		}
		hasValue = false
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var fd int
	fd, err = s.file.begin(c.Context())
	if err != nil {
		return
	}
	defer s.file.end()
	switch name {
	case "finish-bind", "finish-listen":
		operation := tcpBindOperation
		if name == "finish-listen" {
			operation = tcpListenOperation
		}
		if s.operation != operation {
			err = tcpCode(NetworkNotInProgress)
		} else {
			s.operation = tcpNoOperation
		}
		hasValue = false
	case "finish-connect":
		if s.operation != tcpConnectOperation {
			err = tcpCode(NetworkNotInProgress)
			return
		}
		var ready bool
		_, _, ready, err = host.Ready(fd, false)
		if err != nil {
			return
		}
		if !ready {
			err = tcpCode(NetworkWouldBlock)
			return
		}
		var status int
		status, err = unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
		if err == nil && status != 0 {
			err = unix.Errno(status)
		}
		s.operation = tcpNoOperation
		if err != nil {
			s.state = tcpFailed
			s.failure = err
			return
		}
		s.state = tcpConnected
		value, err = p.streams(c, s)
	case "start-listen":
		hasValue = false
		if s.operation != tcpNoOperation {
			err = tcpCode(NetworkConcurrencyConflict)
			return
		}
		if s.state != tcpBound {
			err = tcpCode(NetworkInvalidState)
			return
		}
		err = unix.Listen(fd, s.backlog)
		if err == nil {
			s.state = tcpListening
			s.operation = tcpListenOperation
		}
	case "accept":
		if s.state != tcpListening || s.operation != tcpNoOperation {
			err = tcpCode(NetworkInvalidState)
			return
		}
		value, err = p.accept(c, s, fd)
	case "local-address", "remote-address":
		if s.state == tcpFailed {
			err = s.failure
			return
		}
		if s.state == tcpInitial || s.state == tcpConnecting || name == "remote-address" && s.state != tcpConnected {
			err = tcpCode(NetworkInvalidState)
			return
		}
		var address SocketAddress
		address, err = socketAddressQuery(fd, name == "remote-address")
		if err == nil {
			value, err = socketAddressValue(address)
		}
	case "address-family":
		value.Bits = uint64(s.family)
	case "is-listening":
		if s.state == tcpListening {
			value.Bits = 1
		}
	case "set-listen-backlog-size":
		hasValue = false
		n := args[1].Bits
		if n == 0 {
			err = tcpCode(NetworkInvalidArgument)
			return
		}
		if n > 1<<31-1 {
			n = 1<<31 - 1
		}
		if s.state != tcpInitial && s.state != tcpBound && s.state != tcpListening {
			err = tcpCode(NetworkInvalidState)
			return
		}
		if s.state == tcpListening {
			err = unix.Listen(fd, int(n))
		}
		if err == nil {
			s.backlog = int(n)
		}
	case "shutdown":
		hasValue = false
		if s.state != tcpConnected || !s.streams {
			err = tcpCode(NetworkInvalidState)
			return
		}
		how := unix.SHUT_RD
		if args[1].Bits == 1 {
			how = unix.SHUT_WR
		} else if args[1].Bits == 2 {
			how = unix.SHUT_RDWR
		}
		s.input.mu.Lock()
		readClosed := s.input.closed
		s.input.mu.Unlock()
		s.output.mu.Lock()
		writeClosed := s.output.closed
		s.output.mu.Unlock()
		needRead := args[1].Bits != 1 && !readClosed
		needWrite := args[1].Bits != 0 && !writeClosed
		if needRead || needWrite {
			how = unix.SHUT_RDWR
			if !needRead {
				how = unix.SHUT_WR
			}
			if !needWrite {
				how = unix.SHUT_RD
			}
			err = unix.Shutdown(fd, how)
			if err == unix.ENOTCONN {
				err = nil
			}
		}
		if err == nil {
			if args[1].Bits != 1 {
				s.input.shutdown()
			}
			if args[1].Bits != 0 {
				s.output.shutdown()
			}
		}
	default:
		if s.state == tcpConnecting {
			err = tcpCode(NetworkInvalidState)
			return
		}
		if s.state == tcpFailed {
			err = s.failure
			return
		}
		var input uint64
		if len(args) > 1 {
			input = args[1].Bits
			hasValue = false
		}
		value, err = tcpOption(fd, s.family, name, input)
		if err == nil {
			switch name {
			case "set-keep-alive-idle-time", "set-hop-limit", "set-receive-buffer-size", "set-send-buffer-size":
				if s.inherited == nil {
					s.inherited = map[string]uint64{}
				}
				s.inherited[name] = input
			}
		}
	}
	return
}
func (s *tcpSocket) canStart(name string) error {
	if s.operation != tcpNoOperation {
		return tcpCode(NetworkConcurrencyConflict)
	}
	if name == "start-bind" && s.state != tcpInitial || name == "start-connect" && s.state != tcpInitial && s.state != tcpBound {
		return tcpCode(NetworkInvalidState)
	}
	return nil
}
func (p *TCP) accept(c *component.Caller, parent *tcpSocket, fd int) (component.Value, error) {
	if err := p.reserve(); err != nil {
		return component.Value{}, err
	}
	defer p.unreserve()
	syscall.ForkLock.RLock()
	child, _, err := unix.Accept(fd)
	if err == nil {
		unix.CloseOnExec(child)
		err = unix.SetNonblock(child, true)
		if err != nil {
			unix.Close(child)
		}
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		return component.Value{}, err
	}
	for name, value := range parent.inherited {
		if _, err = tcpOption(child, parent.family, name, value); err != nil {
			unix.Close(child)
			return component.Value{}, err
		}
	}
	ctx, cancel := context.WithCancel(p.ctx)
	file := &socketFile{file: os.NewFile(uintptr(child), "WASI accepted TCP socket"), fd: child, ctx: ctx, cancel: cancel, done: make(chan struct{})}
	socket, s, err := p.publish(c, file, parent.family, tcpConnected)
	if err != nil {
		file.Close()
		file.WaitClosed(nil)
		return component.Value{}, err
	}
	streams, err := p.streams(c, s)
	if err != nil {
		c.DropResource(socket.Resource)
		return component.Value{}, err
	}
	return component.Value{Elements: append([]component.Value{socket}, streams.Elements...)}, nil
}
