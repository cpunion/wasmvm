package preview2

import (
	"context"
	"errors"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const StreamsInterface = "wasi:io/streams@0.2.12"

type IOConfig struct {
	Stdin                                         InputSource
	Stdout, Stderr                                OutputSource
	MaxStreams, MaxReadBytes, MaxWriteBytes       uint32
	StdinTerminal, StdoutTerminal, StderrTerminal bool
	Errors                                        IOErrorConfig
	Poll                                          PollConfig
}

// IO owns sources, stream/error resources, pollables and queued output workers.
// Zero configuration uses EOF stdin and discarded output, never ambient stdio.
// Successful construction transfers configured source ownership. Close requests
// cancellation; WaitClosed joins workers and source cleanup outside callbacks.
type IO struct {
	mu                                    sync.Mutex
	config                                IOConfig
	ctx                                   context.Context
	cancel                                context.CancelFunc
	done                                  chan struct{}
	closed                                bool
	cleanupErr                            error
	next                                  uint64
	records                               map[uint32]*ioStream
	sources                               map[*ioSource]struct{}
	active, workers                       sync.WaitGroup
	errors                                *IOErrors
	poller                                *Poller
	inputType, outputType                 *component.HostResourceType
	terminalInputType, terminalOutputType *component.HostResourceType
	stdin, stdout, stderr                 *ioSource
}
type ioSource struct {
	input  InputSource
	output OutputSource
	refs   uint32
	once   sync.Once
	err    error
}

func (s *ioSource) close() error {
	s.once.Do(func() {
		complete := false
		defer func() {
			if !complete {
				_ = recover()
				s.err = errors.New("IO source close panicked")
			}
		}()
		if s.input != nil {
			s.err = s.input.Close()
		} else {
			s.err = s.output.Close()
		}
		complete = true
	})
	return s.err
}

type ioStream struct {
	mu, readMu                         sync.Mutex
	owner                              *IO
	id                                 uint32
	source                             *ioSource
	ctx                                context.Context
	cancel                             context.CancelFunc
	changed                            chan struct{}
	workerDone                         chan struct{}
	children                           uint32
	dropped, closed, running, flushing bool
	checked                            bool
	queue                              []byte
	permit                             uint64
	failure                            error
	reported                           bool
}

func (r *ioStream) notifyLocked() { close(r.changed); r.changed = make(chan struct{}) }
func (r *ioStream) requestClose() {
	r.mu.Lock()
	r.closed = true
	r.cancel()
	r.notifyLocked()
	r.mu.Unlock()
}

func NewIO(config IOConfig) (*IO, error) {
	if config.MaxStreams == 0 {
		config.MaxStreams = 1024
	}
	if config.MaxReadBytes == 0 {
		config.MaxReadBytes = 64 << 10
	}
	if config.MaxWriteBytes == 0 {
		config.MaxWriteBytes = 64 << 10
	}
	if config.MaxReadBytes >= 1<<28 || config.MaxWriteBytes >= 1<<28 {
		return nil, wasm.ErrLimit
	}
	errors, err := NewIOErrors(config.Errors)
	if err != nil {
		return nil, err
	}
	poller, err := NewPoller(config.Poll)
	if err != nil {
		errors.Close()
		return nil, err
	}
	if config.Stdin == nil {
		config.Stdin = NewBytesInput(nil)
	}
	if config.Stdout == nil {
		config.Stdout = &discardOutput{}
	}
	if config.Stderr == nil {
		config.Stderr = &discardOutput{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &IO{config: config, ctx: ctx, cancel: cancel, done: make(chan struct{}), next: 1, records: map[uint32]*ioStream{}, sources: map[*ioSource]struct{}{}, errors: errors, poller: poller}
	p.stdin = &ioSource{input: config.Stdin, refs: 1}
	p.stdout = &ioSource{output: config.Stdout, refs: 1}
	p.stderr = &ioSource{output: config.Stderr, refs: 1}
	for _, s := range []*ioSource{p.stdin, p.stdout, p.stderr} {
		p.sources[s] = struct{}{}
	}
	p.terminalInputType = component.NewHostResourceType(nil)
	p.terminalOutputType = component.NewHostResourceType(nil)
	p.inputType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	p.outputType = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { return p.drop(id) })
	return p, nil
}

// Poller is borrowed from this IO owner; its lifetime also bounds clocks built
// with NewMonotonicClock. IO.Close closes this domain and cancels its polls.
func (p *IO) Poller() *Poller {
	if p == nil {
		return nil
	}
	return p.poller
}
func (p *IO) Close() {
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
	records, sources := p.records, p.sources
	p.records = nil
	p.sources = nil
	p.mu.Unlock()
	go func() {
		// Wake poll-based blocking operations before joining active host calls.
		p.poller.Close()
		for _, r := range records {
			r.requestClose()
		}
		for s := range sources {
			p.cleanup(s.close())
		}
		p.active.Wait()
		p.workers.Wait()
		p.errors.Close()
		close(p.done)
	}()
}
func (p *IO) WaitClosed(ctx context.Context) error {
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
func (p *IO) begin() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	p.active.Add(1)
	return nil
}
func (p *IO) check(ctx context.Context) error {
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

func (p *IO) publish(c *component.Caller, s *ioSource) (component.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return component.Value{}, wasm.ErrClosed
	}
	if uint64(len(p.records)) >= uint64(p.config.MaxStreams) || p.next > 1<<32-1 {
		return component.Value{}, wasm.ErrLimit
	}
	id := uint32(p.next)
	p.next++
	ctx, cancel := context.WithCancel(p.ctx)
	r := &ioStream{owner: p, id: id, source: s, ctx: ctx, cancel: cancel, changed: make(chan struct{})}
	typ := p.inputType
	if s.output != nil {
		typ = p.outputType
	}
	value, err := c.NewResource(typ, id)
	if err != nil {
		cancel()
		return component.Value{}, err
	}
	s.refs++
	p.records[id] = r
	p.sources[s] = struct{}{}
	return value, nil
}
func (p *IO) record(c *component.Caller, ref *component.Resource, input bool) (*ioStream, error) {
	typ := p.outputType
	if input {
		typ = p.inputType
	}
	id, err := c.ResourceRep(typ, ref)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	r := p.records[id]
	if r == nil || input != (r.source.input != nil) {
		return nil, vm.ErrReference
	}
	return r, nil
}
func (p *IO) drop(id uint32) error {
	p.mu.Lock()
	r := p.records[id]
	if r == nil {
		p.mu.Unlock()
		return nil
	}
	r.mu.Lock()
	r.dropped = true
	if r.children != 0 {
		r.mu.Unlock()
		p.mu.Unlock()
		return wasm.ErrTrap
	}
	r.closed = true
	r.cancel()
	r.notifyLocked()
	running, done := r.running, r.workerDone
	r.mu.Unlock()
	p.active.Add(1)
	p.mu.Unlock()
	defer p.active.Done()
	if running {
		<-done
	}

	p.finalize(id)
	return nil
}
func (p *IO) finalize(id uint32) {
	p.mu.Lock()
	r := p.records[id]
	if r == nil {
		p.mu.Unlock()
		return
	}
	r.mu.Lock()
	eligible := r.dropped && r.children == 0 && !r.running
	r.mu.Unlock()
	if !eligible {
		p.mu.Unlock()
		return
	}
	delete(p.records, id)
	s := r.source
	s.refs--
	last := s.refs == 0
	if last {
		delete(p.sources, s)
	}
	p.mu.Unlock()
	if last {
		p.cleanup(s.close())
	}
}

func (p *IO) cleanup(err error) {
	if err == nil {
		return
	}
	p.mu.Lock()
	if p.cleanupErr == nil {
		p.cleanupErr = err
	}
	p.mu.Unlock()
}

// NewInput transfers source ownership after accepting an active creation.
// Failed publication closes it; an already closed owner rejects without taking
// ownership. Sources implement idempotent Close. Stdin retains an owner pin.
func (p *IO) NewInput(c *component.Caller, source InputSource) (component.Value, error) {
	if p == nil || p.inputType == nil || source == nil {
		return component.Value{}, vm.ErrReference
	}
	if err := p.begin(); err != nil {
		return component.Value{}, err
	}
	defer p.active.Done()
	s := &ioSource{input: source}
	value, err := p.publish(c, s)
	if err != nil {
		p.cleanup(s.close())
	}
	return value, err
}

// NewOutput transfers source ownership under the same contract as NewInput.
func (p *IO) NewOutput(c *component.Caller, source OutputSource) (component.Value, error) {
	if p == nil || p.outputType == nil || source == nil {
		return component.Value{}, vm.ErrReference
	}
	if err := p.begin(); err != nil {
		return component.Value{}, err
	}
	defer p.active.Done()
	s := &ioSource{output: source}
	value, err := p.publish(c, s)
	if err != nil {
		p.cleanup(s.close())
	}
	return value, err
}
