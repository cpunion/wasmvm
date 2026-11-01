//go:build linux || darwin

package preview3

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/internal/wasi/host"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const HTTPTypesInterface = "wasi:http/types@0.3.2"
const HTTPHandlerInterface = "wasi:http/handler@0.3.2"
const HTTPClientInterface = "wasi:http/client@0.3.2"

// Request and resolved-connection grants are independent. Nil grants deny;
// Handler explicitly supplies an upstream handler without inheriting network
// authority. Policy/handler callbacks must observe cancellation.
type HTTPConfig struct {
	MaxResources, MaxMessages, MaxRequests, MaxHeaderFields, MaxHeaderBytes, BodyBufferBytes uint32
	Timeout                                                                                  time.Duration
	AllowRequest                                                                             func(context.Context, *url.URL) error
	AllowConnect                                                                             func(context.Context, SocketAddress) error
	Handler                                                                                  func(context.Context, *http.Request) (*http.Response, error)
	ForbiddenHeaders                                                                         []string
	TLSConfig                                                                                *tls.Config
}
type HTTP struct {
	mu              sync.Mutex
	config          HTTPConfig
	network         *Sockets
	fields          host.HTTPFieldLimits
	ctx             context.Context
	cancel          context.CancelFunc
	stop            func() bool
	closed          bool
	done            chan struct{}
	active, workers sync.WaitGroup
	next            uint64
	types           map[string]*component.HostResourceType
	entries         map[uint32]*httpEntry
	bodies          map[*httpBody]struct{}
	requests        uint32
	cleanup         chan func()
	cleanupDone     chan struct{}
	cleanupStopping bool
	transports      map[httpTimeouts]*http.Transport
	transportOrder  []httpTimeouts
}
type httpEntry struct {
	id             uint32
	kind           string
	owner          *HTTP
	parent         *httpEntry
	children       uint32
	dropped, moved bool
	mu             sync.Mutex
	fields         *host.HTTPFields
	options        [3]*uint64
	immutable      bool
	message        *httpMessage
}
type httpMessage struct {
	headers                 *host.HTTPFields
	method                  string
	path, scheme, authority *string
	status                  uint16
	options                 *[3]*uint64
	body                    *httpBody
}

func NewHTTP(network *Sockets, config HTTPConfig) (*HTTP, error) {
	if network == nil || network.resources == nil {
		return nil, vm.ErrReference
	}
	if err := network.check(context.Background()); err != nil {
		return nil, err
	}
	if config.MaxResources == 0 {
		config.MaxResources = 4096
	}
	if config.MaxMessages == 0 {
		config.MaxMessages = 128
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
	if config.MaxResources > 1<<20 || config.MaxRequests > 1<<20 || config.MaxMessages > 1<<20 || config.MaxHeaderFields > 65536 || config.MaxHeaderBytes > 16<<20 || config.BodyBufferBytes > 1<<20 || config.Timeout < 0 {
		return nil, wasm.ErrLimit
	}
	ctx, cancel := context.WithCancel(network.ctx)
	p := &HTTP{network: network, config: config, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, types: map[string]*component.HostResourceType{}, entries: map[uint32]*httpEntry{}, bodies: map[*httpBody]struct{}{}, cleanup: make(chan func(), config.MaxMessages+1), cleanupDone: make(chan struct{}), transports: map[httpTimeouts]*http.Transport{}}
	p.fields = host.HTTPFieldLimits{Count: config.MaxHeaderFields, Bytes: uint64(config.MaxHeaderBytes), Forbidden: map[string]bool{}}
	for _, name := range config.ForbiddenHeaders {
		if !host.HTTPToken(name) {
			cancel()
			return nil, wasm.ErrInvalid
		}
		p.fields.Forbidden[httpLower(name)] = true
	}
	p.config.ForbiddenHeaders = nil
	if config.TLSConfig != nil {
		p.config.TLSConfig = config.TLSConfig.Clone()
		if config.TLSConfig.RootCAs != nil {
			p.config.TLSConfig.RootCAs = config.TLSConfig.RootCAs.Clone()
		}
		if config.TLSConfig.ClientCAs != nil {
			p.config.TLSConfig.ClientCAs = config.TLSConfig.ClientCAs.Clone()
		}
	}
	for _, name := range []string{"fields", "request", "request-options", "response"} {
		p.types[name] = component.NewHostResourceType(func(c *component.Caller, id uint32) error { return p.drop(c, id) })
	}
	p.types["headers"], p.types["trailers"] = p.types["fields"], p.types["fields"]
	go p.reap()
	p.mu.Lock()
	p.stop = context.AfterFunc(network.ctx, p.Close)
	p.mu.Unlock()
	return p, nil
}
func (p *HTTP) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.ctx.Err() != nil {
		return wasm.ErrClosed
	}
	return nil
}
func (p *HTTP) begin(ctx context.Context) error {
	if err := p.check(ctx); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	p.active.Add(1)
	return nil
}
func (p *HTTP) publish(c *component.Caller, e *httpEntry) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return component.Value{}, wasm.ErrClosed
	}
	if uint64(len(p.entries)) >= uint64(p.config.MaxResources) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	id := uint32(p.next)
	p.next++
	v, err := c.NewResource(p.types[e.kind], id)
	if err != nil {
		return component.Value{}, err
	}
	e.owner, e.id = p, id
	p.entries[id] = e
	if e.parent != nil {
		e.parent.children++
	}
	if e.message != nil {
		e.message.body.acquire()
	}
	return v, nil
}
func (p *HTTP) entry(c *component.Caller, v component.Value, kind string) (*httpEntry, error) {
	id, err := c.ResourceRep(p.types[kind], v.Resource)
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
func (p *HTTP) take(c *component.Caller, v component.Value, kind string, detachChildren bool) (*httpEntry, error) {
	e, err := p.entry(c, v, kind)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	if e.children != 0 && !detachChildren {
		p.mu.Unlock()
		return nil, wasm.ErrTrap
	}
	if detachChildren {
		for _, child := range p.entries {
			if child.parent == e {
				child.parent = nil
			}
		}
		e.children = 0
	}
	e.moved = true
	p.mu.Unlock()
	if err := c.DropResource(v.Resource); err != nil {
		return nil, err
	}
	return e, nil
}
func (p *HTTP) drop(c *component.Caller, id uint32) error {
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
	delete(p.entries, id)
	parent := e.parent
	if parent != nil {
		parent.children--
	}
	p.mu.Unlock()
	if e.message != nil && !e.moved {
		e.message.body.dispose(c, httpErr("connection-terminated"))
		e.message.body.release()
	}
	if parent != nil && parent.dropped && parent.children == 0 {
		return p.drop(c, parent.id)
	}
	return nil
}
func (p *HTTP) reap() {
	defer close(p.cleanupDone)
	for {
		fn := <-p.cleanup
		if fn != nil {
			fn()
		}
		p.mu.Lock()
		stop := p.cleanupStopping
		p.mu.Unlock()
		if stop && len(p.cleanup) == 0 {
			return
		}
	}
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
	entries := p.entries
	p.entries = nil
	bodies := make([]*httpBody, 0, len(p.bodies))
	for b := range p.bodies {
		bodies = append(bodies, b)
	}
	p.bodies = nil
	transports := p.transports
	p.transports = nil
	p.mu.Unlock()
	if p.stop != nil {
		p.stop()
	}
	for _, b := range bodies {
		b.dispose(nil, httpErr("connection-terminated"))
	}
	for _, t := range transports {
		t.CloseIdleConnections()
	}
	for _, e := range entries {
		if e.message != nil && !e.moved {
			e.message.body.release()
		}
	}
	go func() {
		p.active.Wait()
		p.workers.Wait()
		p.mu.Lock()
		p.cleanupStopping = true
		p.mu.Unlock()
		p.cleanup <- nil
		<-p.cleanupDone
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
		return nil
	}
}
func (p *HTTP) Imports() component.Imports {
	if p == nil || p.types == nil {
		return component.Imports{}
	}
	types := component.Imports{Resources: p.types, Functions: map[string]component.HostFunction{}, FunctionChecks: map[string]component.FunctionCheck{}}
	for name, schema := range httpWITFunctions["types"] {
		types.FunctionChecks[name] = schema.check(p.types)
		types.Functions[name] = p.method(name)
	}
	instances := map[string]component.Imports{HTTPTypesInterface: types}
	for _, name := range []string{"handler", "client"} {
		key := "handle"
		if name == "client" {
			key = "send"
		}
		instances["wasi:http/"+name+"@0.3.2"] = component.Imports{Resources: p.types, AsyncCompletions: map[string]component.AsyncHostFunction{key: p.send(name == "handler")}, FunctionChecks: map[string]component.FunctionCheck{key: httpWITFunctions[name][key].check(p.types)}}
	}
	return component.Imports{Instances: instances}
}
