//go:build linux || darwin

package preview2

import (
	"errors"
	"runtime"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

func (e *udpEntry) checkAddress(address SocketAddress, remote bool) error {
	return socketCheckAddress(e.family, address, remote)
}
func udpPeer(value component.Value) (*SocketAddress, error) {
	if value.Case == 0 {
		return nil, nil
	}
	if value.Case != 1 || len(value.Elements) != 1 {
		return nil, socketAddressError()
	}
	address, err := parseSocketAddress(value.Elements[0])
	return &address, err
}
func (p *UDP) socketExecute(c *component.Caller, e *udpEntry, name string, args []component.Value) (value component.Value, payload bool, err error) {
	payload = true
	if name == "start-bind" {
		payload = false
		if _, err = c.ResourceRep(p.network.networkType, args[1].Resource); err != nil {
			return
		}
		var address SocketAddress
		address, err = parseSocketAddress(args[2])
		if err != nil {
			return
		}
		if err = e.checkAddress(address, false); err != nil {
			return
		}
		check := func() error {
			if e.pending {
				return socketCode(NetworkConcurrencyConflict)
			}
			if e.bound {
				return socketCode(NetworkInvalidState)
			}
			return nil
		}
		e.mu.Lock()
		err = check()
		e.mu.Unlock()
		if err != nil {
			return
		}
		if err = p.authorize(c.Context(), address, p.config.AllowBind); err != nil {
			return
		}
		e.mu.Lock()
		defer e.mu.Unlock()
		if err = check(); err != nil {
			return
		}
		var fd int
		fd, err = e.file.begin(c.Context())
		if err != nil {
			return
		}
		defer e.file.end()
		err = socketBind(fd, address)
		if err == nil {
			e.bound = true
			e.pending = true
		}
		return
	}
	if name == "stream" {
		value, err = p.openStreams(c, e, args[1])
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	var fd int
	fd, err = e.file.begin(c.Context())
	if err != nil {
		return
	}
	defer e.file.end()
	switch name {
	case "finish-bind":
		payload = false
		if !e.pending {
			err = socketCode(NetworkNotInProgress)
		} else {
			e.pending = false
		}
	case "local-address":
		if !e.bound {
			err = socketCode(NetworkInvalidState)
			return
		}
		var address SocketAddress
		address, err = socketAddressQuery(fd, false)
		if err == nil {
			value, err = socketAddressValue(address)
		}
	case "remote-address":
		if e.remote == nil {
			err = socketCode(NetworkInvalidState)
			return
		}
		value, err = socketAddressValue(*e.remote)
	case "address-family":
		value.Bits = uint64(e.family)
	default:
		option := name
		if name == "unicast-hop-limit" {
			option = "hop-limit"
		}
		if name == "set-unicast-hop-limit" {
			option = "set-hop-limit"
		}
		var input uint64
		if len(args) > 1 {
			input = args[1].Bits
			payload = false
		}
		value, err = socketOption(fd, e.family, option, input)
	}
	return
}
func (p *UDP) openStreams(c *component.Caller, socket *udpEntry, option component.Value) (component.Value, error) {
	check := func() error {
		if socket.pending {
			return socketCode(NetworkConcurrencyConflict)
		}
		if !socket.bound {
			return socketCode(NetworkInvalidState)
		}
		p.mu.Lock()
		hasStreams := socket.streams != 0
		p.mu.Unlock()
		if hasStreams {
			return wasm.ErrTrap
		}
		return nil
	}
	socket.mu.Lock()
	stateError := check()
	socket.mu.Unlock()
	if stateError != nil {
		return component.Value{}, stateError
	}

	peer, err := udpPeer(option)
	if err != nil {
		return component.Value{}, err
	}
	if peer != nil {
		if err = socket.checkAddress(*peer, true); err != nil {
			return component.Value{}, err
		}
		err = p.authorize(c.Context(), *peer, p.config.AllowSend)
		if err != nil {
			if c.Context().Err() != nil || p.ctx.Err() != nil {
				return component.Value{}, err
			}
			if err = p.authorize(c.Context(), *peer, p.config.AllowReceive); err != nil {
				return component.Value{}, err
			}
		}
	}
	socket.mu.Lock()
	defer socket.mu.Unlock()
	if err := check(); err != nil {
		return component.Value{}, err
	}
	fd, err := socket.file.begin(c.Context())
	if err != nil {
		return component.Value{}, err
	}
	defer socket.file.end()
	if peer != nil {
		err = socketConnect(fd, *peer)
		// Linux may require dissolving an old association before changing routes.
		if runtime.GOOS == "linux" && err == unix.EINVAL {
			_ = socketDisconnectDatagram(fd)
			err = socketConnect(fd, *peer)
		}
		if err == unix.EINPROGRESS {
			err = nil
		}
	} else if socket.remote != nil {
		var local SocketAddress
		local, err = socketAddressQuery(fd, false)
		if err == nil {
			err = socketDisconnectDatagram(fd)
		}
		if runtime.GOOS == "darwin" && (err == unix.EINVAL || err == unix.EAFNOSUPPORT) {
			err = nil
		}
		if err == nil {
			current, failed := socketAddressQuery(fd, false)
			if failed != nil {
				err = failed
			} else if current.Address.Port() == 0 {
				err = socketBind(fd, local)
			}
		}
	}
	// A failed connect may still change the kernel association.
	actual, failed := socketAddressQuery(fd, true)
	if failed == nil {
		socket.remote = &actual
	} else {
		socket.remote = nil
	}
	if err != nil {
		return component.Value{}, err
	}
	incoming := &udpEntry{kind: 1, parent: socket, file: socket.file, family: socket.family, poll: newSocketPoll(socket.file, true)}
	in, err := p.publish(c, incoming)
	if err != nil {
		incoming.poll.Close()
		return component.Value{}, err
	}
	outgoing := &udpEntry{kind: 2, parent: socket, file: socket.file, family: socket.family, poll: newSocketPoll(socket.file, false)}
	out, err := p.publish(c, outgoing)
	if err != nil {
		outgoing.poll.Close()
		c.DropResource(in.Resource)
		return component.Value{}, err
	}
	return component.Value{Elements: []component.Value{in, out}}, nil
}
func (p *UDP) receive(c *component.Caller, e *udpEntry, max uint64) (component.Value, error) {
	if max == 0 {
		return component.Value{}, nil
	}
	if max > uint64(p.config.MaxDatagrams) {
		max = uint64(p.config.MaxDatagrams)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	fd, err := e.file.begin(c.Context())
	if err != nil {
		return component.Value{}, err
	}
	defer e.file.end()
	if e.buffer == nil {
		e.buffer = make([]byte, 65535)
	}
	var datagrams []component.Value
	for scan := uint32(0); scan < p.config.MaxDatagrams && uint64(len(datagrams)) < max; scan++ {
		if err = c.Context().Err(); err != nil {
			return component.Value{}, err
		}
		var n int
		var address SocketAddress
		for {
			n, address, err = socketReceiveDatagram(fd, e.buffer)
			if err != unix.EINTR || c.Context().Err() != nil {
				break
			}
		}
		if err == unix.EAGAIN {
			break
		}
		if err != nil {
			if len(datagrams) > 0 {
				break
			}
			return component.Value{}, err
		}
		data := string(e.buffer[:n])
		// Policy callbacks must not run while holding stream/provider locks.
		e.mu.Unlock()
		allowed := p.authorize(c.Context(), address, p.config.AllowReceive)
		e.mu.Lock()
		if allowed != nil {
			if c.Context().Err() != nil {
				return component.Value{}, c.Context().Err()
			}
			if p.ctx.Err() != nil {
				return component.Value{}, wasm.ErrClosed
			}
			continue
		}
		remote, err := socketAddressValue(address)
		if err != nil {
			if len(datagrams) > 0 {
				break
			}
			return component.Value{}, err
		}
		datagrams = append(datagrams, component.Value{Elements: []component.Value{{Text: data}, remote}})
	}
	return component.Value{Elements: datagrams}, nil
}
func (p *UDP) checkSend(c *component.Caller, e *udpEntry) (component.Value, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	fd, err := e.file.begin(c.Context())
	if err != nil {
		return component.Value{}, err
	}
	defer e.file.end()
	_, _, ready, err := host.Ready(fd, false)
	if err != nil {
		return component.Value{}, err
	}
	e.permit = 0
	if ready {
		e.permit = uint64(p.config.MaxDatagrams)
	}
	return component.Value{Bits: e.permit}, nil
}
func (p *UDP) send(c *component.Caller, e *udpEntry, datagrams []component.Value) (component.Value, error) {
	if len(datagrams) == 0 {
		return component.Value{}, nil
	}
	e.mu.Lock()
	if uint64(len(datagrams)) > e.permit {
		e.mu.Unlock()
		return component.Value{}, wasm.ErrTrap
	}
	e.permit = 0
	e.mu.Unlock()
	e.parent.mu.Lock()
	var peer *SocketAddress
	if e.parent.remote != nil {
		copy := *e.parent.remote
		peer = &copy
	}
	e.parent.mu.Unlock()
	var count uint64
	for _, datagram := range datagrams {
		err := func() error {
			if len(datagram.Elements) != 2 {
				return socketAddressError()
			}
			text, err := datagram.Elements[0].ByteString()
			if err != nil {
				return err
			}
			data := []byte(text)
			if len(data) > 65535 {
				return socketCode(NetworkDatagramTooLarge)
			}
			destination, err := udpPeer(datagram.Elements[1])
			if err != nil {
				return err
			}
			if destination == nil {
				destination = peer
			}
			if destination == nil {
				return socketAddressError()
			}
			if err = e.checkAddress(*destination, true); err != nil {
				return err
			}
			if peer != nil && *peer != *destination {
				return socketAddressError()
			}
			if err = p.authorize(c.Context(), *destination, p.config.AllowSend); err != nil {
				return err
			}
			fd, err := e.file.begin(c.Context())
			if err != nil {
				return err
			}
			defer e.file.end()
			target := destination
			if peer != nil {
				target = nil
			}
			for {
				if err = c.Context().Err(); err != nil {
					return err
				}
				var n int
				n, err = socketSendDatagram(fd, data, target)
				if err == unix.EINTR {
					continue
				}
				if err == nil && n != len(data) {
					return socketCode(NetworkUnknown)
				}
				return err
			}
		}()
		if err != nil {
			if count != 0 || errors.Is(err, unix.EAGAIN) {
				break
			}
			return component.Value{}, err
		}
		count++
	}
	return component.Value{Bits: count}, nil
}
