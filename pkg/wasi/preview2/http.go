//go:build linux || darwin

package preview2

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const HTTPTypesInterface = "wasi:http/types@0.2.12"
const HTTPOutgoingInterface = "wasi:http/outgoing-handler@0.2.12"
const HTTPIncomingInterface = "wasi:http/incoming-handler@0.2.12"

// HTTPConfig grants request and resolved connection authority separately.
// Nil callbacks deny. Callbacks run outside provider locks and must observe
// cancellation. TLSConfig is cloned. No environment proxy or credentials are used.
type HTTPConfig struct {
	MaxResources, MaxRequests, MaxHeaderFields, MaxHeaderBytes, BodyBufferBytes uint32
	Timeout                                                                     time.Duration
	AllowRequest                                                                func(context.Context, *url.URL) error
	AllowConnect                                                                func(context.Context, SocketAddress) error
	ForbiddenHeaders                                                            []string
	TLSConfig                                                                   *tls.Config
}

const (
	httpFieldsKind = iota
	httpIncomingRequest
	httpOutgoingRequest
	httpRequestOptions
	httpResponseOutparam
	httpIncomingResponse
	httpIncomingBody
	httpFutureTrailers
	httpOutgoingResponse
	httpOutgoingBody
	httpFutureResponse
)

var httpResourceNames = []string{"fields", "incoming-request", "outgoing-request", "request-options", "response-outparam", "incoming-response", "incoming-body", "future-trailers", "outgoing-response", "outgoing-body", "future-incoming-response"}

type HTTP struct {
	mu                       sync.Mutex
	network                  *Network
	config                   HTTPConfig
	fields                   httpFieldLimits
	clock                    *MonotonicClock
	types                    [11]*component.HostResourceType
	ctx                      context.Context
	cancel                   context.CancelFunc
	stop                     func() bool
	closed                   bool
	next                     uint64
	entries                  map[uint32]*httpEntry
	bodies                   map[*httpBody]bool
	requests                 uint32
	active, workers, cleanup sync.WaitGroup
	done                     chan struct{}
	cleanupErr               error
	transport                *http.Transport
	transports               map[httpTimeouts]*http.Transport
	transportOrder           []httpTimeouts
}
type httpEntry struct {
	mu       sync.Mutex
	owner    *HTTP
	id       uint32
	kind     int
	parent   *httpEntry
	children uint32 // owner.mu
	dropped  bool   // owner.mu
	moved    bool
	fields   *httpFields
	message  *httpMessage
	options  [3]*uint64
	body     *httpBody
	future   *httpFuture
	reply    *httpReply
}
type httpOptionalString struct {
	value string
	set   bool
}
type httpMessage struct {
	headers                 *httpFields
	method                  string
	path, scheme, authority httpOptionalString
	status                  uint16
	body                    *httpBody
	bodyTaken               bool
}
type httpFuture struct {
	mu          sync.Mutex
	ready, used bool
	changed     chan struct{}
	response    *httpMessage
	failure     error
	cancel      context.CancelFunc
}

func (f *httpFuture) Readiness() (bool, <-chan struct{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ready, f.changed
}
func NewHTTP(network *Network, config HTTPConfig) (*HTTP, error) {
	if network == nil || network.networkType == nil {
		return nil, vm.ErrReference
	}
	if err := network.check(context.Background()); err != nil {
		return nil, err
	}
	if err := network.io.check(context.Background()); err != nil {
		return nil, err
	}
	if config.MaxResources == 0 {
		config.MaxResources = 4096
	}
	if config.MaxRequests == 0 {
		config.MaxRequests = 128
	}
	if config.MaxHeaderFields == 0 {
		config.MaxHeaderFields = 1024
	}
	if config.MaxHeaderBytes == 0 {
		config.MaxHeaderBytes = 65536
	}
	if config.BodyBufferBytes == 0 {
		config.BodyBufferBytes = 65536
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxHeaderFields > 65536 || config.MaxHeaderBytes > 16<<20 || config.BodyBufferBytes > 1<<20 || config.Timeout < 0 {
		return nil, wasm.ErrLimit
	}
	clock, err := NewMonotonicClock(network.io.poller)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(network.ctx)
	p := &HTTP{network: network, config: config, clock: clock, ctx: ctx, cancel: cancel, next: 1, entries: map[uint32]*httpEntry{}, bodies: map[*httpBody]bool{}, done: make(chan struct{})}
	p.fields = httpFieldLimits{count: config.MaxHeaderFields, bytes: uint64(config.MaxHeaderBytes), forbidden: map[string]bool{}}
	for _, name := range config.ForbiddenHeaders {
		if !httpToken(name) {
			cancel()
			return nil, wasm.ErrInvalid
		}
		p.fields.forbidden[httpLower(name)] = true
	}
	p.config.ForbiddenHeaders = append([]string(nil), config.ForbiddenHeaders...)
	if config.TLSConfig != nil {
		p.config.TLSConfig = config.TLSConfig.Clone()
		if config.TLSConfig.RootCAs != nil {
			p.config.TLSConfig.RootCAs = config.TLSConfig.RootCAs.Clone()
		}
		if config.TLSConfig.ClientCAs != nil {
			p.config.TLSConfig.ClientCAs = config.TLSConfig.ClientCAs.Clone()
		}
	}
	for kind := range p.types {
		p.types[kind] = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	}
	p.transport = p.newTransport()
	p.transports = map[httpTimeouts]*http.Transport{p.defaultTimeouts(): p.transport}
	p.transportOrder = []httpTimeouts{p.defaultTimeouts()}
	stop := context.AfterFunc(network.ctx, p.Close)
	p.mu.Lock()
	p.stop = stop
	p.mu.Unlock()
	return p, nil
}
func (p *HTTP) begin(ctx context.Context) error {
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
func (p *HTTP) startWorker() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	p.workers.Add(1)
	return nil
}
func (p *HTTP) Close() {
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
	entries, bodies, stop, transports := p.entries, p.bodies, p.stop, p.transports
	p.entries = nil
	p.bodies = nil
	p.transports = nil
	p.mu.Unlock()
	if stop != nil {
		stop()
	}
	for _, transport := range transports {
		transport.CloseIdleConnections()
	}
	for body := range bodies {
		body.cancel()
		body.buffer.close()
	}
	for _, entry := range entries {
		p.destroy(entry)
	}
	for body := range bodies {
		body.close()
	}
	go func() {
		p.active.Wait()
		p.workers.Wait()
		p.cleanup.Wait()
		p.transport.CloseIdleConnections()
		close(p.done)
	}()
}
func (p *HTTP) WaitClosed(ctx context.Context) error {
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
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.cleanupErr
	}
}
func (p *HTTP) recordCleanup(err error) {
	if err != nil {
		p.mu.Lock()
		if p.cleanupErr == nil {
			p.cleanupErr = err
		}
		p.mu.Unlock()
	}
}
func (p *HTTP) publish(c *component.Caller, e *httpEntry) (component.Value, error) {
	return p.publishWith(e, func(t *component.HostResourceType, id uint32) (component.Value, error) { return c.NewResource(t, id) })
}
func (p *HTTP) publishWith(e *httpEntry, mint func(*component.HostResourceType, uint32) (component.Value, error)) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.ctx.Err() != nil {
		return component.Value{}, wasm.ErrClosed
	}
	if uint64(len(p.entries)) >= uint64(p.config.MaxResources) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	if e.parent != nil && (e.parent.dropped || p.entries[e.parent.id] != e.parent) {
		return component.Value{}, vm.ErrReference
	}
	e.owner = p
	e.id = uint32(p.next)
	p.next++
	v, err := mint(p.types[e.kind], e.id)
	if err != nil {
		return component.Value{}, err
	}
	p.entries[e.id] = e
	if e.parent != nil {
		e.parent.children++
	}
	return v, nil
}
func (p *HTTP) entry(c *component.Caller, value component.Value, kind int) (*httpEntry, error) {
	id, err := c.ResourceRep(p.types[kind], value.Resource)
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
func (p *HTTP) checkChildren(e *httpEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if e.children != 0 {
		return wasm.ErrTrap
	}
	return nil
}
func (p *HTTP) take(c *component.Caller, value component.Value, kind int) (*httpEntry, error) {
	e, err := p.entry(c, value, kind)
	if err != nil {
		return nil, err
	}
	if err = p.checkChildren(e); err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.moved = true
	e.mu.Unlock()
	if err = c.DropResource(value.Resource); err != nil {
		return nil, err
	}
	return e, nil
}
func (p *HTTP) detach(e *httpEntry, list *[]*httpEntry) {
	delete(p.entries, e.id)
	p.cleanup.Add(1)
	*list = append(*list, e)
	if e.parent != nil {
		e.parent.children--
		if e.parent.dropped && e.parent.children == 0 {
			p.detach(e.parent, list)
		}
	}
}
func (p *HTTP) drop(id uint32) error {
	p.mu.Lock()
	e := p.entries[id]
	if e == nil {
		p.mu.Unlock()
		return nil
	}
	e.dropped = true
	if e.children != 0 {
		p.mu.Unlock()
		return wasm.ErrTrap
	}
	var entries []*httpEntry
	p.detach(e, &entries)
	p.mu.Unlock()
	for _, entry := range entries {
		p.destroy(entry)
		p.cleanup.Done()
	}
	return nil
}
func (p *HTTP) child(e *httpEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || e.dropped || p.entries[e.id] != e {
		return wasm.ErrClosed
	}
	e.children++
	return nil
}
func (p *HTTP) release(e *httpEntry) {
	p.mu.Lock()
	e.children--
	var entries []*httpEntry
	if !p.closed && e.dropped && e.children == 0 && p.entries[e.id] == e {
		p.detach(e, &entries)
	}
	p.mu.Unlock()
	for _, entry := range entries {
		p.destroy(entry)
		p.cleanup.Done()
	}
}
func (p *HTTP) destroy(e *httpEntry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.moved && e.kind != httpFutureTrailers {
		return
	}
	if e.body != nil {
		e.body.close()
	}
	if e.message != nil && e.message.body != nil && (!e.message.bodyTaken || e.kind == httpOutgoingRequest || e.kind == httpOutgoingResponse) {
		e.message.body.close()
	}
	if e.future != nil {
		e.future.mu.Lock()
		if e.future.cancel != nil {
			e.future.cancel()
		}
		if e.future.response != nil && !e.future.used {
			e.future.response.body.close()
		}
		e.future.ready = true
		select {
		case <-e.future.changed:
		default:
			close(e.future.changed)
		}
		e.future.mu.Unlock()
	}
	if e.reply != nil {
		e.reply.fail(httpErr("HTTP-response-incomplete"))
	}
}
func (p *HTTP) subscribe(c *component.Caller, e *httpEntry, source PollSource) (*component.Value, error) {
	poller := p.network.io.poller
	if err := poller.begin(); err != nil {
		return nil, err
	}
	defer poller.active.Done()
	if err := p.child(e); err != nil {
		return nil, err
	}
	v, err := poller.publish(c, source, func() { p.release(e) })
	return &v, err
}
