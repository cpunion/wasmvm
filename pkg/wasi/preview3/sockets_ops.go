//go:build linux || darwin

package preview3

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"golang.org/x/sys/unix"
)

func (p *Sockets) execute(ctx context.Context, s *netSocket, name string, args []component.Value) (*component.Value, error) {
	if err := p.check(ctx); err != nil {
		return nil, err
	}
	if name == "bind" || name == "connect" {
		address := netParse(args[1])
		remote := name == "connect"
		if err := netCheckAddress(s.family, address, remote); err != nil {
			return nil, err
		}
		s.mu.Lock()
		allowed := !s.sending && !s.receiving && (s.state == 0 || s.udp && remote)
		s.mu.Unlock()
		if !allowed {
			return nil, netError("invalid-state")
		}
		policy := p.config.AllowBind
		if remote {
			policy = p.config.AllowConnect
		}
		if err := p.authorize(ctx, address, policy); err != nil {
			return nil, err
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.sending || s.receiving || s.state != 0 && !(s.udp && remote) {
			return nil, netError("invalid-state")
		}
		var err error
		if remote {
			err = host.SocketConnect(s.fd, address)
			if err == unix.EINVAL && runtime.GOOS == "linux" && s.peer != nil {
				_ = host.SocketDisconnectDatagram(s.fd)
				err = host.SocketConnect(s.fd, address)
			}
			if err == nil {
				s.state = 3
				s.peer = &address
			}
		} else {
			if !s.udp {
				err = unix.SetsockoptInt(s.fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			}
			if err == nil {
				err = host.SocketBind(s.fd, address)
			}
			if err == nil {
				s.state = 1
			}
		}
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if name == "get-address-family" {
		return &component.Value{Bits: uint64(s.family)}, nil
	}
	if name == "get-is-listening" {
		value := &component.Value{}
		if s.state == 4 {
			value.Bits = 1
		}
		return value, nil
	}
	if s.state == 5 || s.state == 2 {
		return nil, netError("invalid-state")
	}
	switch name {
	case "set-listen-backlog-size":
		if args[1].Bits == 0 {
			return nil, netError("invalid-argument")
		}
		backlog := int(min(args[1].Bits, uint64(1<<31-1)))
		if s.state == 4 {
			if err := unix.Listen(s.fd, backlog); err != nil {
				return nil, err
			}
		}
		s.backlog = backlog
		return nil, nil
	case "get-local-address", "get-remote-address":
		if s.state == 0 || name == "get-remote-address" && s.state != 3 {
			return nil, netError("invalid-state")
		}
		a, err := host.SocketAddressQuery(s.fd, name == "get-remote-address")
		if err != nil {
			return nil, err
		}
		v := netAddress(a)
		return &v, nil
	case "disconnect":
		if s.peer == nil || s.sending || s.receiving {
			return nil, netError("invalid-state")
		}
		local, err := host.SocketAddressQuery(s.fd, false)
		if err != nil {
			return nil, err
		}
		if err = host.SocketDisconnectDatagram(s.fd); err != nil {
			return nil, err
		}
		// Some kernels remove the implicit local binding on AF_UNSPEC disconnect.
		now, err := host.SocketAddressQuery(s.fd, false)
		if err != nil {
			return nil, err
		}
		if now.Address.Port() == 0 {
			err = host.SocketBind(s.fd, local)
			if err != nil {
				return nil, err
			}
		}
		s.peer = nil
		s.state = 1
		return nil, nil
	default:
		option := strings.TrimPrefix(name, "get-")
		option = strings.ReplaceAll(option, "unicast-hop-limit", "hop-limit")
		var input uint64
		if len(args) > 1 {
			input = args[1].Bits
		}
		bits, err := host.SocketOption(s.fd, s.family, option, input)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(name, "set-") {
			if !s.udp {
				if s.inherited == nil {
					s.inherited = map[string]uint64{}
				}
				s.inherited[option] = input
			}
			return nil, nil
		}
		return &component.Value{Bits: bits}, nil
	}
}
func (p *Sockets) executeAsync(ctx context.Context, s *netSocket, name string, args []component.Value) (*component.Value, error) {
	if err := p.check(ctx); err != nil {
		return nil, err
	}
	if name == "connect" {
		address := netParse(args[1])
		if err := netCheckAddress(s.family, address, true); err != nil {
			return nil, err
		}
		s.mu.Lock()
		allowed := s.state <= 1
		s.mu.Unlock()
		if !allowed {
			return nil, netError("invalid-state")
		}
		if err := p.authorize(ctx, address, p.config.AllowConnect); err != nil {
			return nil, err
		}
		s.mu.Lock()
		if s.state > 1 {
			s.mu.Unlock()
			return nil, netError("invalid-state")
		}
		s.state = 2
		err := host.SocketConnect(s.fd, address)
		s.mu.Unlock()
		if err == unix.EINPROGRESS || err == unix.EALREADY {
			_, _, err = host.Poll(ctx, s.fd, false)
			if err == nil {
				var status int
				status, err = unix.GetsockoptInt(s.fd, unix.SOL_SOCKET, unix.SO_ERROR)
				if err == nil && status != 0 {
					err = unix.Errno(status)
				}
			}
		}
		if err == nil {
			err = p.check(ctx)
		}
		s.mu.Lock()
		if err == nil {
			s.state = 3
			s.peer = &address
		} else {
			s.state = 5
			unix.Close(s.fd)
			s.fd = -1
		}
		s.mu.Unlock()
		return nil, err
	}
	sending := name == "send"
	s.mu.Lock()
	if s.state == 5 || sending && s.sending || !sending && (s.receiving || s.state == 0) {
		s.mu.Unlock()
		return nil, netError("invalid-state")
	}
	if sending {
		s.sending = true
	} else {
		s.receiving = true
	}
	var peer *SocketAddress
	if s.peer != nil {
		copy := *s.peer
		peer = &copy
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		if sending {
			s.sending = false
		} else {
			s.receiving = false
		}
		s.mu.Unlock()
	}()
	if sending {
		if len(args[1].Text) > 65536 {
			return nil, netError("datagram-too-large")
		}
		address := peer
		if args[2].Case == 1 {
			a := netParse(args[2].Elements[0])
			address = &a
			if peer != nil && *peer != a {
				return nil, netError("invalid-argument")
			}
		}
		if address == nil {
			return nil, netError("invalid-argument")
		}
		if err := netCheckAddress(s.family, *address, true); err != nil {
			return nil, err
		}
		if err := p.authorize(ctx, *address, p.config.AllowSend); err != nil {
			return nil, err
		}
		for {
			if err := p.check(ctx); err != nil {
				return nil, err
			}
			target := address
			if peer != nil {
				target = nil
			}
			n, err := host.SocketSendDatagram(s.fd, []byte(args[1].Text), target)
			if err == nil {
				if n != len(args[1].Text) {
					return nil, netError("other")
				}
				s.mu.Lock()
				if s.state == 0 {
					s.state = 1
				}
				s.mu.Unlock()
				return nil, nil
			}
			if err == unix.EINTR {
				continue
			}
			if err != unix.EAGAIN && err != unix.EWOULDBLOCK {
				return nil, err
			}
			if _, _, err = host.Poll(ctx, s.fd, false); err != nil {
				return nil, err
			}
		}
	}
	if p.config.AllowReceive == nil {
		return nil, netError("access-denied")
	}
	var err error
	data := make([]byte, 65536)
	for {
		if err = p.check(ctx); err != nil {
			return nil, err
		}
		n, address, failed := host.SocketReceiveDatagram(s.fd, data)
		if failed == nil {
			if err = p.authorize(ctx, address, p.config.AllowReceive); err != nil {
				return nil, err
			}
			v := netAddress(address)
			return &component.Value{Elements: []component.Value{{Text: string(data[:n])}, v}}, nil
		}
		if errors.Is(failed, unix.EINTR) {
			continue
		}
		if failed != unix.EAGAIN && failed != unix.EWOULDBLOCK {
			return nil, failed
		}
		if _, _, err = host.Poll(ctx, s.fd, true); err != nil {
			return nil, err
		}
	}
}
