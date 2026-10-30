package preview2

import (
	"context"
	"reflect"
	"sync"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const PollInterface = "wasi:io/poll@0.2.12"

// PollSource returns a nonblocking, coherent readiness snapshot. When not ready,
// changed must close on the next possible readiness change. A nil changed
// channel means no pending notification; cancellation still interrupts polling.
// Implementations must be safe for concurrent snapshots from separate Stores.
// IO errors count as ready; the following IO operation reports the error.
type PollSource interface {
	Readiness() (ready bool, changed <-chan struct{})
}

type PollConfig struct{ MaxPollables, MaxPollItems uint32 }

// Poller owns a bounded host capability domain, independently of guest Stores.
// Close after its bound instances close to reclaim even abandoned subscriptions.
// Dropping a guest resource releases its entry immediately. Sources passed to
// NewPollable are borrowed; their external owner is responsible for closing IO.
type Poller struct {
	mu       sync.Mutex
	config   PollConfig
	resource *component.HostResourceType
	entries  map[uint32]pollEntry
	next     uint64
	done     chan struct{}
	drained  chan struct{}
	closed   bool
	active   sync.WaitGroup
}
type pollEntry struct {
	source  PollSource
	release func()
}

func NewPoller(config PollConfig) (*Poller, error) {
	if config.MaxPollables == 0 {
		config.MaxPollables = 1024
	}
	if config.MaxPollItems == 0 {
		config.MaxPollItems = 1024
	}
	// reflect.Select admits at most 65536 cases, including both cancel channels.
	if config.MaxPollItems > 65534 {
		return nil, wasm.ErrLimit
	}
	p := &Poller{config: config, entries: map[uint32]pollEntry{}, next: 1, done: make(chan struct{}), drained: make(chan struct{})}
	p.resource = component.NewHostResourceType(func(_ *component.Caller, id uint32) error { p.remove(id); return nil })
	return p, nil
}

func (p *Poller) Close() {
	if p == nil || p.done == nil {
		return
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		<-p.drained
		return
	}
	p.closed = true
	close(p.done)
	entries := p.entries
	p.entries = nil
	p.mu.Unlock()
	for _, entry := range entries {
		if entry.release != nil {
			entry.release()
		}
	}
	p.active.Wait()
	close(p.drained)
}

// A creation lease also covers rollback before a resource reaches entries.
func (p *Poller) begin() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return wasm.ErrClosed
	}
	p.active.Add(1)
	return nil
}
func (p *Poller) remove(id uint32) {
	p.mu.Lock()
	entry := p.entries[id]
	delete(p.entries, id)
	if entry.release != nil {
		p.active.Add(1)
	}
	p.mu.Unlock()
	if entry.release != nil {
		defer p.active.Done()
		entry.release()
	}
}
func (p *Poller) publish(c *component.Caller, source PollSource, release func()) (component.Value, error) {
	p.mu.Lock()
	var err error
	if p.closed {
		err = wasm.ErrClosed
	} else if uint64(len(p.entries)) >= uint64(p.config.MaxPollables) || p.next > 1<<32-1 {
		err = wasm.ErrLimit
	} else if source == nil {
		err = vm.ErrReference
	}
	if err != nil {
		p.mu.Unlock()
		if release != nil {
			release()
		}
		return component.Value{}, err
	}
	id := uint32(p.next)
	p.next++
	p.entries[id] = pollEntry{source, release}
	// NewResource executes no user code. Keep publication atomic with Close;
	// failed allocation releases the provisional source after unlocking.
	value, err := c.NewResource(p.resource, id)
	if err != nil {
		delete(p.entries, id)
	}
	p.mu.Unlock()
	if err != nil && release != nil {
		release()
	}
	return value, err
}

// NewPollable grants ownership in the Caller's bound Store. A resource from a
// different Poller or Store cannot gain authority through its numeric ID.
func (p *Poller) NewPollable(c *component.Caller, source PollSource) (component.Value, error) {
	if p == nil || p.resource == nil {
		return component.Value{}, vm.ErrReference
	}
	if err := p.begin(); err != nil {
		return component.Value{}, err
	}
	defer p.active.Done()
	return p.publish(c, source, nil)
}
func (p *Poller) source(c *component.Caller, value component.Value) (PollSource, error) {
	id, err := c.ResourceRep(p.resource, value.Resource)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, wasm.ErrClosed
	}
	entry, ok := p.entries[id]
	if !ok {
		return nil, vm.ErrReference
	}
	return entry.source, nil
}
func (p *Poller) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-p.done:
		return wasm.ErrClosed
	default:
		return nil
	}
}

// block needs no indexed result list or reflective multi-source selection.
func (p *Poller) block(ctx context.Context, source PollSource) error {
	for {
		if err := p.check(ctx); err != nil {
			return err
		}
		ready, changed := source.Readiness()
		if ready {
			return p.check(ctx)
		}
		select {
		case <-ctx.Done():
		case <-p.done:
		case <-changed:
		}
	}
}
func (p *Poller) wait(ctx context.Context, sources []PollSource) ([]component.Value, error) {
	if len(sources) == 0 || uint64(len(sources)) > uint64(p.config.MaxPollItems) {
		return nil, wasm.ErrLimit
	}
	cases := make([]reflect.SelectCase, len(sources)+2)
	cases[0] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}
	cases[1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(p.done)}
	for {
		if err := p.check(ctx); err != nil {
			return nil, err
		}
		ready := []component.Value{}
		for i, source := range sources {
			at, changed := source.Readiness()
			if at {
				ready = append(ready, component.Value{Bits: uint64(i)})
			}
			cases[i+2] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(changed)}
		}
		if err := p.check(ctx); err != nil {
			return nil, err
		}
		if len(ready) != 0 {
			return ready, nil
		}
		reflect.Select(cases)
	}
}
func (p *Poller) contract(name string) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		if f.ParameterCount() != 1 {
			return vm.ErrType
		}
		_, arg := f.Parameter(0)
		result := f.Result()
		if name == "poll" {
			if arg.Kind() == component.KindList && arg.Element().Kind() == component.KindBorrow && arg.Element().IsResource(p.resource) && result.Kind() == component.KindList && result.Element().Kind() == component.KindU32 {
				return nil
			}
		} else if arg.Kind() == component.KindBorrow && arg.IsResource(p.resource) {
			if name == "[method]pollable.ready" && result.Kind() == component.KindBool || name == "[method]pollable.block" && !result.Valid() {
				return nil
			}
		}
		return vm.ErrType
	}
}
func (p *Poller) Imports() component.Imports {
	if p == nil || p.resource == nil {
		return component.Imports{}
	}
	functions := map[string]component.HostFunction{
		"[method]pollable.ready": func(c *component.Caller, args []component.Value) (*component.Value, error) {
			source, err := p.source(c, args[0])
			if err != nil {
				return nil, err
			}
			if err = p.check(c.Context()); err != nil {
				return nil, err
			}
			ready, _ := source.Readiness()
			value := component.Value{}
			if ready {
				value.Bits = 1
			}
			return &value, p.check(c.Context())
		},
		"[method]pollable.block": func(c *component.Caller, args []component.Value) (*component.Value, error) {
			source, err := p.source(c, args[0])
			if err != nil {
				return nil, err
			}
			return nil, p.block(c.Context(), source)
		},
		"poll": func(c *component.Caller, args []component.Value) (*component.Value, error) {
			input := args[0].Elements
			if len(input) == 0 || uint64(len(input)) > uint64(p.config.MaxPollItems) {
				return nil, wasm.ErrLimit
			}
			sources := make([]PollSource, len(input))
			for i, v := range input {
				var err error
				sources[i], err = p.source(c, v)
				if err != nil {
					return nil, err
				}
			}
			ready, err := p.wait(c.Context(), sources)
			if err != nil {
				return nil, err
			}
			return &component.Value{Elements: ready}, nil
		},
	}
	checks := map[string]component.FunctionCheck{}
	for name := range functions {
		checks[name] = p.contract(name)
	}
	return component.Imports{Instances: map[string]component.Imports{PollInterface: {Resources: map[string]*component.HostResourceType{"pollable": p.resource}, Functions: functions, FunctionChecks: checks}}}
}
