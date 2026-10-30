package preview2

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
	"golang.org/x/net/idna"
)

const NetworkInterface = "wasi:sockets/network@0.2.12"
const InstanceNetworkInterface = "wasi:sockets/instance-network@0.2.12"
const IPNameLookupInterface = "wasi:sockets/ip-name-lookup@0.2.12"

type NetworkErrorCode uint32

const (
	NetworkUnknown NetworkErrorCode = iota
	NetworkAccessDenied
	NetworkNotSupported
	NetworkInvalidArgument
	NetworkOutOfMemory
	NetworkTimeout
	NetworkConcurrencyConflict
	NetworkNotInProgress
	NetworkWouldBlock
	NetworkInvalidState
	NetworkNewSocketLimit
	NetworkAddressNotBindable
	NetworkAddressInUse
	NetworkRemoteUnreachable
	NetworkConnectionRefused
	NetworkConnectionReset
	NetworkConnectionAborted
	NetworkDatagramTooLarge
	NetworkNameUnresolvable
	NetworkTemporaryResolverFailure
	NetworkPermanentResolverFailure
)

var networkErrorLabels = []string{"unknown", "access-denied", "not-supported", "invalid-argument", "out-of-memory", "timeout", "concurrency-conflict", "not-in-progress", "would-block", "invalid-state", "new-socket-limit", "address-not-bindable", "address-in-use", "remote-unreachable", "connection-refused", "connection-reset", "connection-aborted", "datagram-too-large", "name-unresolvable", "temporary-resolver-failure", "permanent-resolver-failure"}

// NetworkError retains typed provenance for network-error-code, independently
// of human-readable IO error diagnostics.
type NetworkError struct {
	Code  NetworkErrorCode
	Cause error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "WASI network error"
}
func (e *NetworkError) Unwrap() error { return e.Cause }
func networkError(err error) NetworkErrorCode {
	var typed *NetworkError
	if errors.As(err, &typed) {
		return typed.Code
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return NetworkTimeout
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		if dns.IsTimeout {
			return NetworkTimeout
		}
		if dns.IsNotFound {
			return NetworkNameUnresolvable
		}
		if dns.IsTemporary {
			return NetworkTemporaryResolverFailure
		}
		return NetworkPermanentResolverFailure
	}
	return NetworkUnknown
}
func networkFailure(code NetworkErrorCode) *component.Value {
	return &component.Value{Case: 1, Elements: []component.Value{{Bits: uint64(code)}}}
}

// NetworkConfig grants name lookup explicitly. Resolve must observe cancellation,
// return promptly after cancellation, and return immutable addresses. A nil
// callback uses the Go resolver only when AllowNameLookup is enabled. Literal
// IP addresses are parsed locally. No socket authority is implied by this API.
type NetworkConfig struct {
	AllowNameLookup                               bool
	Resolve                                       func(context.Context, string) ([]netip.Addr, error)
	MaxAddressStreams, MaxAddresses, MaxNameBytes uint32
	ResolveTimeout                                time.Duration
}
type Network struct {
	mu                      sync.Mutex
	config                  NetworkConfig
	io                      *IO
	networkType, streamType *component.HostResourceType
	ctx                     context.Context
	cancel                  context.CancelFunc
	closed                  bool
	done                    chan struct{}
	active, workers         sync.WaitGroup
	next                    uint64
	streams                 map[uint32]*resolveStream
}
type resolveStream struct {
	id            uint32
	mu            sync.Mutex
	addresses     []netip.Addr
	at            int
	done, dropped bool
	code          NetworkErrorCode
	failed        bool
	children      uint32
	changed       chan struct{}
	cancel        context.CancelFunc
	workerDone    chan struct{}
}

func NewNetwork(ioOwner *IO, config NetworkConfig) (*Network, error) {
	if ioOwner == nil || ioOwner.inputType == nil {
		return nil, vm.ErrReference
	}
	if err := ioOwner.check(context.Background()); err != nil {
		return nil, err
	}
	if config.MaxAddressStreams == 0 {
		config.MaxAddressStreams = 128
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
	if config.MaxAddresses > 1<<20 || config.MaxNameBytes > 1<<20 || config.ResolveTimeout < 0 {
		return nil, wasm.ErrLimit
	}
	if config.Resolve == nil {
		resolver := &net.Resolver{PreferGo: true}
		config.Resolve = func(ctx context.Context, name string) ([]netip.Addr, error) {
			return resolver.LookupNetIP(ctx, "ip", name)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Network{config: config, io: ioOwner, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, streams: map[uint32]*resolveStream{}}
	p.networkType = component.NewHostResourceType(nil)
	p.streamType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	return p, nil
}
func (p *Network) begin() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	p.active.Add(1)
	return nil
}
func (p *Network) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-p.ctx.Done():
		return wasm.ErrClosed
	default:
		return nil
	}
}
func (p *Network) Close() {
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
	streams := p.streams
	p.streams = nil
	p.mu.Unlock()
	for _, s := range streams {
		s.mu.Lock()
		s.dropped = true
		s.cancel()
		s.mu.Unlock()
	}
	go func() { p.active.Wait(); p.workers.Wait(); close(p.done) }()
}
func (p *Network) WaitClosed(ctx context.Context) error {
	if p == nil || p.done == nil {
		return vm.ErrReference
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (p *Network) stream(c *component.Caller, value component.Value) (*resolveStream, error) {
	id, err := c.ResourceRep(p.streamType, value.Resource)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	s := p.streams[id]
	if s == nil {
		return nil, vm.ErrReference
	}
	return s, nil
}
func (p *Network) drop(id uint32) error {
	p.mu.Lock()
	s := p.streams[id]
	if s == nil {
		p.mu.Unlock()
		return nil
	}
	s.mu.Lock()
	s.dropped = true
	s.cancel()
	children := s.children
	if children == 0 {
		delete(p.streams, id)
	}
	s.mu.Unlock()
	p.mu.Unlock()
	<-s.workerDone
	if children != 0 {
		return wasm.ErrTrap
	}
	return nil
}
func (p *Network) resolve(c *component.Caller, args []component.Value) (*component.Value, error) {
	if err := p.begin(); err != nil {
		return nil, err
	}
	defer p.active.Done()
	if err := p.check(c.Context()); err != nil {
		return nil, err
	}
	if _, err := c.ResourceRep(p.networkType, args[0].Resource); err != nil {
		return nil, err
	}
	if !p.config.AllowNameLookup {
		return networkFailure(NetworkAccessDenied), nil
	}
	name := args[1].Text
	if name == "" || strings.IndexByte(name, 0) >= 0 || uint64(len(name)) > uint64(p.config.MaxNameBytes) {
		return networkFailure(NetworkInvalidArgument), nil
	}
	literal := name
	if strings.HasPrefix(literal, "[") && strings.HasSuffix(literal, "]") {
		literal = literal[1 : len(literal)-1]
	}
	address, literalErr := netip.ParseAddr(literal)
	if literalErr == nil && address.Zone() != "" {
		return networkFailure(NetworkInvalidArgument), nil
	}
	if literalErr != nil {
		var err error
		name, err = idna.Lookup.ToASCII(name)
		if err != nil || name == "" || len(name) > 253 {
			return networkFailure(NetworkInvalidArgument), nil
		}
	}
	ctx, cancel := p.ctx, context.CancelFunc(func() {})
	if literalErr != nil {
		ctx, cancel = context.WithTimeout(p.ctx, p.config.ResolveTimeout)
	}
	s := &resolveStream{changed: make(chan struct{}), cancel: cancel, workerDone: make(chan struct{})}
	if literalErr == nil {
		s.done = true
		s.addresses = []netip.Addr{address}
		close(s.workerDone)
		cancel()
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		cancel()
		return nil, wasm.ErrClosed
	}
	if uint64(len(p.streams)) >= uint64(p.config.MaxAddressStreams) || p.next > 1<<32-1 {
		p.mu.Unlock()
		cancel()
		return networkFailure(NetworkOutOfMemory), nil
	}
	id := uint32(p.next)
	s.id = id
	p.next++
	p.streams[id] = s
	value, err := c.NewResource(p.streamType, id)
	if err != nil {
		delete(p.streams, id)
		p.mu.Unlock()
		cancel()
		return nil, err
	}
	if !s.done {
		p.workers.Add(1)
	}
	p.mu.Unlock()
	if !s.done {
		go p.lookup(ctx, s, name)
	}
	return streamOK(&value), nil
}
func (p *Network) lookup(ctx context.Context, s *resolveStream, name string) {
	defer p.workers.Done()
	defer close(s.workerDone)
	defer s.cancel()
	var addresses []netip.Addr
	var err error
	func() {
		defer func() {
			if recover() != nil {
				err = &NetworkError{Code: NetworkPermanentResolverFailure}
			}
		}()
		addresses, err = p.config.Resolve(ctx, name)
	}()
	if err == nil {
		err = ctx.Err()
	}
	var result []netip.Addr
	if err == nil {
		if uint64(len(addresses)) > uint64(p.config.MaxAddresses) {
			err = &NetworkError{Code: NetworkOutOfMemory}
		} else {
			seen := map[netip.Addr]bool{}
			for _, address := range addresses {
				if !address.IsValid() || address.Zone() != "" {
					err = &NetworkError{Code: NetworkPermanentResolverFailure}
					break
				}
				address = address.Unmap()
				if !seen[address] {
					result = append(result, address)
					seen[address] = true
				}
			}
		}
	}
	if err == nil && len(result) == 0 {
		err = &NetworkError{Code: NetworkNameUnresolvable}
	}
	s.mu.Lock()
	s.done = true
	s.addresses = result
	if err != nil {
		s.failed = true
		s.code = networkError(err)
	}
	close(s.changed)
	s.changed = make(chan struct{})
	s.mu.Unlock()
}
func (s *resolveStream) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done || s.dropped, s.changed
}
