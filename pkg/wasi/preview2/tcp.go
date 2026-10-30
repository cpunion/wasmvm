//go:build linux || darwin

package preview2

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

const TCPInterface = "wasi:sockets/tcp@0.2.12"
const TCPCreateInterface = "wasi:sockets/tcp-create-socket@0.2.12"

// TCPConfig authorizes bind/connect explicitly. Nil callbacks deny access.
// Callbacks run without provider/socket locks, must observe cancellation and
// may run concurrently across Stores. An accepted connection inherits its
// listener's grant. Socket creation alone grants no network access.
type TCPConfig struct {
	MaxSockets              uint32
	AllowBind, AllowConnect func(context.Context, SocketAddress) error
}
type TCP struct {
	clock           *MonotonicClock
	mu              sync.Mutex
	network         *Network
	config          TCPConfig
	resource        *component.HostResourceType
	ctx             context.Context
	cancel          context.CancelFunc
	stop            func() bool
	closed          bool
	done            chan struct{}
	active, cleanup sync.WaitGroup
	next            uint64
	creating        uint32
	sockets         map[uint32]*tcpSocket
	cleanupErr      error
}
type tcpState uint8

const (
	tcpInitial tcpState = iota
	tcpBound
	tcpConnecting
	tcpConnected
	tcpListening
	tcpFailed
)

type tcpOperation uint8

const (
	tcpNoOperation tcpOperation = iota
	tcpBindOperation
	tcpConnectOperation
	tcpListenOperation
)

type tcpSocket struct {
	mu            sync.Mutex
	owner         *TCP
	id            uint32
	file          *socketFile
	family        uint32
	state         tcpState
	operation     tcpOperation
	failure       error
	backlog       int
	streams       bool
	children      uint32 // guarded by provider.mu, including IO sources and pollables
	dropped       bool   // guarded by provider.mu
	read, write   *socketPoll
	input, output *tcpSource
	inherited     map[string]uint64
}

func NewTCP(network *Network, config TCPConfig) (*TCP, error) {
	if network == nil || network.networkType == nil {
		return nil, vm.ErrReference
	}
	if err := network.check(context.Background()); err != nil {
		return nil, err
	}
	if err := network.io.check(context.Background()); err != nil {
		return nil, err
	}
	if config.MaxSockets == 0 {
		config.MaxSockets = 128
	}
	ctx, cancel := context.WithCancel(network.ctx)
	p := &TCP{network: network, config: config, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, sockets: map[uint32]*tcpSocket{}}
	clock, err := NewMonotonicClock(network.io.poller)
	if err != nil {
		cancel()
		return nil, err
	}
	p.clock = clock
	p.resource = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	stop := context.AfterFunc(network.ctx, p.Close)
	p.mu.Lock()
	p.stop = stop
	p.mu.Unlock()
	return p, nil
}
func (p *TCP) begin(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.ctx.Err() != nil {
		return wasm.ErrClosed
	}
	p.active.Add(1)
	return nil
}
func (p *TCP) Close() {
	if p == nil || p.done == nil {
		return
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.cancel()
	sockets := p.sockets
	stop := p.stop
	p.sockets = nil
	p.mu.Unlock()
	if stop != nil {
		stop()
	}
	for _, s := range sockets {
		s.file.Close()
	}
	go func() {
		p.active.Wait()
		for _, s := range sockets {
			p.recordCleanup(s.file.WaitClosed(nil))
		}
		p.cleanup.Wait()
		close(p.done)
	}()
}
func (p *TCP) recordCleanup(err error) {
	if err == nil {
		return
	}
	p.mu.Lock()
	if p.cleanupErr == nil {
		p.cleanupErr = err
	}
	p.mu.Unlock()
}
func (p *TCP) WaitClosed(ctx context.Context) error {
	if p == nil || p.done == nil {
		return vm.ErrReference
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-p.done:
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.cleanupErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (p *TCP) reserve() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.ctx.Err() != nil {
		return wasm.ErrClosed
	}
	if uint64(len(p.sockets))+uint64(p.creating) >= uint64(p.config.MaxSockets) || p.next > 1<<32-1 {
		return &NetworkError{Code: NetworkNewSocketLimit}
	}
	p.creating++
	return nil
}
func (p *TCP) unreserve() { p.mu.Lock(); p.creating--; p.mu.Unlock() }
func (p *TCP) publish(c *component.Caller, file *socketFile, family uint32, state tcpState) (component.Value, *tcpSocket, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.ctx.Err() != nil {
		return component.Value{}, nil, wasm.ErrClosed
	}
	if p.next > 1<<32-1 {
		return component.Value{}, nil, &NetworkError{Code: NetworkNewSocketLimit}
	}
	id := uint32(p.next)
	p.next++
	s := &tcpSocket{owner: p, id: id, file: file, family: family, state: state, backlog: 128}
	s.read = newSocketPoll(file, true)
	s.write = newSocketPoll(file, false)
	p.sockets[id] = s
	value, err := c.NewResource(p.resource, id)
	if err != nil {
		delete(p.sockets, id)
		return component.Value{}, nil, err
	}
	return value, s, nil
}
func (p *TCP) socket(c *component.Caller, value component.Value) (*tcpSocket, error) {
	id, err := c.ResourceRep(p.resource, value.Resource)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	s := p.sockets[id]
	if s == nil || s.dropped {
		return nil, vm.ErrReference
	}
	return s, nil
}

// retire is called with provider.mu held. Close drains syscall/poll leases;
// cleanup joins detached closes as well as sockets still present at Close.
func (p *TCP) retire(s *tcpSocket) {
	delete(p.sockets, s.id)
	p.cleanup.Add(1)
	s.file.Close()
	go func() { defer p.cleanup.Done(); p.recordCleanup(s.file.WaitClosed(nil)) }()
}
func (p *TCP) drop(id uint32) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	s := p.sockets[id]
	if s == nil {
		return nil
	}
	s.dropped = true
	if s.children != 0 {
		return wasm.ErrTrap
	}
	p.retire(s)
	return nil
}
func (p *TCP) child(s *tcpSocket) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || s.dropped || p.sockets[s.id] != s {
		return wasm.ErrClosed
	}
	s.children++
	return nil
}
func (p *TCP) release(s *tcpSocket) {
	p.mu.Lock()
	defer p.mu.Unlock()
	s.children--
	if !p.closed && s.dropped && s.children == 0 && p.sockets[s.id] == s {
		p.retire(s)
	}
}
func (p *TCP) authorize(ctx context.Context, address SocketAddress, connect bool) error {
	callback := p.config.AllowBind
	if connect {
		callback = p.config.AllowConnect
	}
	return socketAuthorize(ctx, p.ctx, address, callback)
}
func (s *tcpSocket) checkAddress(address SocketAddress, connect bool) error {
	return socketCheckAddress(s.family, address, connect)
}

func tcpCreate(c *component.Caller, p *TCP, family uint32) (component.Value, error) {
	if err := p.reserve(); err != nil {
		return component.Value{}, err
	}
	defer p.unreserve()
	file, err := newSocketFile(p.ctx, family, unix.SOCK_STREAM)
	if err != nil {
		return component.Value{}, err
	}
	value, _, err := p.publish(c, file, family, tcpInitial)
	if err != nil {
		file.Close()
		file.WaitClosed(nil)
	}
	return value, err
}
