//go:build linux || darwin

package preview3

import (
	"context"
	"net"
	"net/netip"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/sys/unix"
)

const SocketsTypesInterface = "wasi:sockets/types@0.3.2"
const IPNameLookupInterface = "wasi:sockets/ip-name-lookup@0.3.2"

type SocketAddress = host.SocketAddress

// Policy callbacks grant individual operations, run without provider locks and
// must observe cancellation. Creation alone grants no network access.
type SocketsConfig struct {
	MaxSockets, MaxWorkers, MaxAddresses, MaxNameBytes uint32
	AllowBind, AllowConnect, AllowSend, AllowReceive   func(context.Context, SocketAddress) error
	AllowNameLookup                                    bool
	Resolve                                            func(context.Context, string) ([]netip.Addr, error)
	ResolveTimeout                                     time.Duration
}
type Sockets struct {
	mu              sync.Mutex
	io              *IO
	config          SocketsConfig
	ctx             context.Context
	cancel          context.CancelFunc
	closed          bool
	done            chan struct{}
	workers         sync.WaitGroup
	active, opening uint32
	next            uint64
	resources       map[string]*component.HostResourceType
	handles         map[uint32]*netSocket
	objects         map[*netSocket]struct{}
	streams         map[*netStream]struct{}
}
type netSocket struct {
	mu                                 sync.Mutex
	owner                              *Sockets
	fd                                 int
	family                             uint32
	backlog                            int
	udp                                bool
	state                              uint8  // 0 unbound, 1 bound, 2 connecting, 3 connected, 4 listening, 5 failed
	refs                               uint32 // guarded by owner.mu
	closed                             bool   // guarded by owner.mu
	peer                               *SocketAddress
	sent, received, sending, receiving bool
	inherited                          map[string]uint64
}

func NewSockets(ioOwner *IO, config SocketsConfig) (*Sockets, error) {
	if ioOwner == nil || ioOwner.inputType == nil {
		return nil, vm.ErrReference
	}
	ioOwner.mu.Lock()
	closed := ioOwner.closed
	ioOwner.mu.Unlock()
	if closed {
		return nil, wasm.ErrClosed
	}
	if config.MaxSockets == 0 {
		config.MaxSockets = 128
	}
	if config.MaxWorkers == 0 {
		config.MaxWorkers = 256
	}
	if config.MaxAddresses == 0 {
		config.MaxAddresses = 256
	}
	if config.MaxNameBytes == 0 {
		config.MaxNameBytes = 1024
	}
	if config.ResolveTimeout == 0 {
		config.ResolveTimeout = 30 * time.Second
	}
	if config.MaxSockets > 1<<20 || config.MaxAddresses > 1<<20 || config.MaxNameBytes > 1<<20 || config.ResolveTimeout < 0 {
		return nil, wasm.ErrLimit
	}
	if config.Resolve == nil {
		r := &net.Resolver{PreferGo: true}
		config.Resolve = func(ctx context.Context, name string) ([]netip.Addr, error) { return r.LookupNetIP(ctx, "ip", name) }
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Sockets{io: ioOwner, config: config, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, handles: map[uint32]*netSocket{}, objects: map[*netSocket]struct{}{}, streams: map[*netStream]struct{}{}, resources: map[string]*component.HostResourceType{}}
	for _, name := range []string{"tcp-socket", "udp-socket"} {
		p.resources[name] = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { p.drop(id); return nil })
	}
	return p, nil
}
func (p *Sockets) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.ctx.Err() != nil {
		return os.ErrClosed
	}
	return nil
}
func (p *Sockets) begin() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return os.ErrClosed
	}
	if p.active >= p.config.MaxWorkers {
		return wasm.ErrLimit
	}
	p.active++
	p.workers.Add(1)
	return nil
}
func (p *Sockets) releaseSlot() { p.mu.Lock(); p.active--; p.mu.Unlock() }
func (p *Sockets) end()         { p.releaseSlot(); p.workers.Done() }
func (p *Sockets) reserve() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return os.ErrClosed
	}
	if uint64(len(p.objects))+uint64(p.opening) >= uint64(p.config.MaxSockets) {
		return wasm.ErrLimit
	}
	p.opening++
	return nil
}
func (p *Sockets) unreserve() { p.mu.Lock(); p.opening--; p.mu.Unlock() }
func (p *Sockets) register(fd int, family uint32, udp bool, state uint8) *netSocket {
	s := &netSocket{owner: p, fd: fd, family: family, udp: udp, state: state, refs: 1}
	p.mu.Lock()
	p.opening--
	p.objects[s] = struct{}{}
	closed := p.closed
	p.mu.Unlock()
	if closed {
		s.close()
	}
	return s
}
func (p *Sockets) create(family uint32, udp bool) (*netSocket, error) {
	if family > 1 {
		return nil, netError("not-supported")
	}
	if err := p.reserve(); err != nil {
		return nil, err
	}
	domain, kind := unix.AF_INET, unix.SOCK_STREAM
	if family == 1 {
		domain = unix.AF_INET6
	}
	if udp {
		kind = unix.SOCK_DGRAM
	}
	syscall.ForkLock.RLock()
	fd, err := unix.Socket(domain, kind, 0)
	if err == nil {
		unix.CloseOnExec(fd)
		err = unix.SetNonblock(fd, true)
		if err == nil && family == 1 {
			err = host.ConfigureIPv6Socket(fd)
		}
		if err != nil {
			unix.Close(fd)
		}
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		p.unreserve()
		return nil, err
	}
	return p.register(fd, family, udp, 0), nil
}
func (s *netSocket) acquire() error {
	p := s.owner
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || s.closed {
		return os.ErrClosed
	}
	if s.refs == 1<<32-1 {
		return wasm.ErrLimit
	}
	s.refs++
	return nil
}
func (s *netSocket) release() {
	p := s.owner
	p.mu.Lock()
	s.refs--
	last := s.refs == 0
	p.mu.Unlock()
	if last {
		s.close()
	}
}
func (s *netSocket) close() {
	p := s.owner
	p.mu.Lock()
	if s.closed {
		p.mu.Unlock()
		return
	}
	s.closed = true
	// The last lease must release the binding before a subsequent guest bind
	// or send. These nonblocking sockets never enable positive SO_LINGER.
	_ = unix.Close(s.fd)
	delete(p.objects, s)
	p.mu.Unlock()
}
func (p *Sockets) publish(c *component.Caller, s *netSocket) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || s.closed {
		return component.Value{}, os.ErrClosed
	}
	if uint64(len(p.handles)) >= uint64(p.config.MaxSockets) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	id := uint32(p.next)
	p.next++
	name := "tcp-socket"
	if s.udp {
		name = "udp-socket"
	}
	value, err := c.NewResource(p.resources[name], id)
	if err != nil {
		return component.Value{}, err
	}
	s.refs++
	p.handles[id] = s
	return value, nil
}
func (p *Sockets) socket(c *component.Caller, value component.Value, udp bool) (*netSocket, error) {
	name := "tcp-socket"
	if udp {
		name = "udp-socket"
	}
	id, err := c.ResourceRep(p.resources[name], value.Resource)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	s := p.handles[id]
	if p.closed || s == nil || s.closed {
		p.mu.Unlock()
		return nil, vm.ErrReference
	}
	s.refs++
	p.mu.Unlock()
	return s, nil
}
func (p *Sockets) drop(id uint32) {
	p.mu.Lock()
	s := p.handles[id]
	delete(p.handles, id)
	p.mu.Unlock()
	if s != nil {
		s.release()
	}
}
func (p *Sockets) Close() {
	if p == nil || p.done == nil {
		return
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	streams := make([]*netStream, 0, len(p.streams))
	for s := range p.streams {
		streams = append(streams, s)
	}
	p.handles = nil
	p.mu.Unlock()
	for _, s := range streams {
		s.op.stop(netCode(netError("invalid-state")) + 1)
	}
	p.cancel()
	go func() {
		p.workers.Wait()
		p.mu.Lock()
		objects := make([]*netSocket, 0, len(p.objects))
		for s := range p.objects {
			objects = append(objects, s)
		}
		p.mu.Unlock()
		for _, s := range objects {
			s.close()
		}
		close(p.done)
	}()
}
func (p *Sockets) WaitClosed(ctx context.Context) error {
	if p == nil || p.done == nil {
		return vm.ErrReference
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return nil
	}
}
