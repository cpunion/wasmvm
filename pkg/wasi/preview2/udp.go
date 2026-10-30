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

const UDPInterface = "wasi:sockets/udp@0.2.12"
const UDPCreateInterface = "wasi:sockets/udp-create-socket@0.2.12"

// UDPConfig grants bind, send and receive independently; nil callbacks deny.
// Callbacks run without provider/resource locks, may run concurrently and must
// observe cancellation. MaxDatagrams bounds each send permit and receive scan.
type UDPConfig struct {
	MaxSockets, MaxStreams, MaxDatagrams uint32
	AllowBind, AllowSend, AllowReceive   func(context.Context, SocketAddress) error
}

type UDP struct {
	mu                                sync.Mutex
	network                           *Network
	config                            UDPConfig
	ctx                               context.Context
	cancel                            context.CancelFunc
	stop                              func() bool
	closed                            bool
	done                              chan struct{}
	active, cleanup                   sync.WaitGroup
	cleanupErr                        error
	next                              uint64
	creating                          uint32
	sockets, streams                  uint32
	entries                           map[uint32]*udpEntry
	socketType, inputType, outputType *component.HostResourceType
}

type udpEntry struct {
	mu                sync.Mutex
	owner             *UDP
	id, kind          uint32 // 0 socket, 1 incoming, 2 outgoing
	parent            *udpEntry
	file              *socketFile
	family            uint32
	bound, pending    bool
	remote            *SocketAddress
	poll              *socketPoll
	permit            uint64
	buffer            []byte
	children, streams uint32 // guarded by owner.mu
	dropped           bool   // guarded by owner.mu
}

func NewUDP(network *Network, config UDPConfig) (*UDP, error) {
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
	if config.MaxStreams == 0 {
		config.MaxStreams = 256
	}
	if config.MaxDatagrams == 0 {
		config.MaxDatagrams = 16
	}
	if config.MaxDatagrams > 64 {
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(network.ctx)
	p := &UDP{network: network, config: config, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, entries: map[uint32]*udpEntry{}}
	p.socketType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	p.inputType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	p.outputType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	stop := context.AfterFunc(network.ctx, p.Close)
	p.mu.Lock()
	p.stop = stop
	p.mu.Unlock()
	return p, nil
}
func (p *UDP) begin(ctx context.Context) error {
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
func (p *UDP) Close() {
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
	entries, stop := p.entries, p.stop
	p.entries = nil
	p.mu.Unlock()
	if stop != nil {
		stop()
	}
	for _, e := range entries {
		if e.kind == 0 {
			e.file.Close()
		}
	}
	go func() {
		p.active.Wait()
		for _, e := range entries {
			if e.kind == 0 {
				p.recordCleanup(e.file.WaitClosed(nil))
			}
		}
		p.cleanup.Wait()
		close(p.done)
	}()
}
func (p *UDP) recordCleanup(err error) {
	if err != nil {
		p.mu.Lock()
		if p.cleanupErr == nil {
			p.cleanupErr = err
		}
		p.mu.Unlock()
	}
}
func (p *UDP) WaitClosed(ctx context.Context) error {
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
func (p *UDP) resourceType(kind uint32) *component.HostResourceType {
	switch kind {
	case 0:
		return p.socketType
	case 1:
		return p.inputType
	default:
		return p.outputType
	}
}
func (p *UDP) entry(c *component.Caller, v component.Value, kind uint32) (*udpEntry, error) {
	id, err := c.ResourceRep(p.resourceType(kind), v.Resource)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	e := p.entries[id]
	if p.closed {
		return nil, wasm.ErrClosed
	}
	if e == nil || e.dropped || e.kind != kind {
		return nil, vm.ErrReference
	}
	return e, nil
}
func (p *UDP) publish(c *component.Caller, e *udpEntry) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.ctx.Err() != nil {
		return component.Value{}, wasm.ErrClosed
	}
	if p.next > 1<<32-1 {
		return component.Value{}, socketCode(NetworkNewSocketLimit)
	}
	if e.kind != 0 && (p.streams >= p.config.MaxStreams || e.parent.dropped || p.entries[e.parent.id] != e.parent) {
		return component.Value{}, socketCode(NetworkNewSocketLimit)
	}
	e.owner = p
	e.id = uint32(p.next)
	p.next++
	v, err := c.NewResource(p.resourceType(e.kind), e.id)
	if err != nil {
		return component.Value{}, err
	}
	p.entries[e.id] = e
	if e.kind == 0 {
		p.sockets++
	} else {
		p.streams++
		e.parent.children++
		e.parent.streams++
	}
	return v, nil
}
func (p *UDP) create(c *component.Caller, family uint32) (component.Value, error) {
	p.mu.Lock()
	if p.closed || p.ctx.Err() != nil {
		p.mu.Unlock()
		return component.Value{}, wasm.ErrClosed
	}
	if uint64(p.sockets)+uint64(p.creating) >= uint64(p.config.MaxSockets) || p.next > 1<<32-1 {
		p.mu.Unlock()
		return component.Value{}, socketCode(NetworkNewSocketLimit)
	}
	p.creating++
	p.mu.Unlock()
	defer func() { p.mu.Lock(); p.creating--; p.mu.Unlock() }()
	file, err := newSocketFile(p.ctx, family, unix.SOCK_DGRAM)
	if err != nil {
		return component.Value{}, err
	}
	v, err := p.publish(c, &udpEntry{file: file, family: family})
	if err != nil {
		file.Close()
		file.WaitClosed(nil)
	}
	return v, err
}
func (p *UDP) retire(e *udpEntry) {
	delete(p.entries, e.id)
	if e.kind == 0 {
		p.sockets--
		p.cleanup.Add(1)
		e.file.Close()
		go func() { defer p.cleanup.Done(); p.recordCleanup(e.file.WaitClosed(nil)) }()
	} else {
		p.streams--
		e.poll.Close()
		e.parent.children--
		e.parent.streams--
		if e.parent.dropped && e.parent.children == 0 {
			p.retire(e.parent)
		}
	}
}
func (p *UDP) drop(id uint32) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	e := p.entries[id]
	if e == nil {
		return nil
	}
	e.dropped = true
	if e.children != 0 {
		return wasm.ErrTrap
	}
	p.retire(e)
	return nil
}
func (p *UDP) child(e *udpEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || e.dropped || p.entries[e.id] != e {
		return wasm.ErrClosed
	}
	e.children++
	return nil
}
func (p *UDP) release(e *udpEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e.children--
	if !p.closed && e.dropped && e.children == 0 && p.entries[e.id] == e {
		p.retire(e)
	}
}
func (e *udpEntry) Readiness() (bool, <-chan struct{}) {
	if e.kind == 0 {
		return true, nil
	}
	return e.poll.Readiness()
}
func (p *UDP) subscribe(c *component.Caller, e *udpEntry) (*component.Value, error) {
	poller := p.network.io.poller
	if err := poller.begin(); err != nil {
		return nil, err
	}
	defer poller.active.Done()
	if err := p.child(e); err != nil {
		return nil, err
	}
	value, err := poller.publish(c, e, func() { p.release(e) })
	return &value, err
}
func (p *UDP) authorize(ctx context.Context, address SocketAddress, callback func(context.Context, SocketAddress) error) error {
	return socketAuthorize(ctx, p.ctx, address, callback)
}
