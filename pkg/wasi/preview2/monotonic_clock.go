package preview2

import (
	"sync"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const MonotonicClockInterface = "wasi:clocks/monotonic-clock@0.2.12"

// MonotonicClock shares its Poller's resource domain. Its origin is private to
// this provider, with explicit microsecond ticks. Close the Poller to cancel
// pending polls and reclaim even guest-abandoned timer subscriptions.
type MonotonicClock struct {
	poller *Poller
	origin time.Time
}

func NewMonotonicClock(poller *Poller) (*MonotonicClock, error) {
	if poller == nil || poller.resource == nil {
		return nil, vm.ErrReference
	}
	poller.mu.Lock()
	defer poller.mu.Unlock()
	if poller.closed {
		return nil, wasm.ErrClosed
	}
	return &MonotonicClock{poller: poller, origin: time.Now()}, nil
}
func (m *MonotonicClock) rawNow() (uint64, error) {
	elapsed := time.Since(m.origin)
	if elapsed < 0 {
		return 0, wasm.ErrLimit
	}
	return uint64(elapsed), nil
}

func (m *MonotonicClock) now() (uint64, error) { now, err := m.rawNow(); return now / 1000 * 1000, err }

type clockSubscription struct {
	mu               sync.Mutex
	clock            *MonotonicClock
	when             uint64
	ready            chan struct{}
	timer            *time.Timer
	tasks            sync.WaitGroup
	closed, signaled bool
	never            bool
}

func (s *clockSubscription) armLocked() {
	if s.closed || s.signaled || s.never {
		return
	}
	now, err := s.clock.now()
	if err != nil || now >= s.when {
		s.signaled = true
		close(s.ready)
		return
	}
	delta := s.when - now
	// Chunk waits beyond time.Duration's signed range; never wrap a u64 deadline
	// into an immediate/negative timer. Round sub-tick waits up to one clock tick.
	if delta > 1<<63-1 {
		delta = 1<<63 - 1
	}
	if delta < 1000 {
		delta = 1000
	}
	s.tasks.Add(1)
	s.timer = time.AfterFunc(time.Duration(delta), func() {
		defer s.tasks.Done()
		s.mu.Lock()
		defer s.mu.Unlock()
		s.armLocked()
	})
}
func (s *clockSubscription) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return true, s.ready
	}
	if s.never {
		return false, s.ready
	}
	now, err := s.clock.now()
	return s.signaled || err != nil || now >= s.when, s.ready
}
func (s *clockSubscription) close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.tasks.Wait()
		return
	}
	s.closed = true
	if s.timer != nil {
		if s.timer.Stop() {
			s.tasks.Done()
		}
		s.timer = nil
	}
	if !s.signaled {
		s.signaled = true
		close(s.ready)
	}
	s.mu.Unlock()
	// Stop does not join an already dispatched AfterFunc. The callback runs no
	// user code and cannot schedule another timer after the closed flag is set.
	s.tasks.Wait()
}
func (m *MonotonicClock) subscribe(c *component.Caller, when uint64, relative bool) (*component.Value, error) {
	if err := m.poller.check(c.Context()); err != nil {
		return nil, err
	}
	if err := m.poller.begin(); err != nil {
		return nil, err
	}
	defer m.poller.active.Done()
	never := false
	if relative {
		now, err := m.rawNow()
		if err != nil {
			return nil, err
		}
		if when > ^uint64(0)-now {
			never = true
		} else {
			when += now
		}
	}
	s := &clockSubscription{clock: m, when: when, ready: make(chan struct{}), never: never}
	s.mu.Lock()
	s.armLocked()
	s.mu.Unlock()
	value, err := m.poller.publish(c, s, s.close)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func (m *MonotonicClock) contract(name string) component.FunctionCheck {
	return func(f component.FunctionInfo) error {
		result := f.Result()
		if name == "now" || name == "resolution" {
			if f.ParameterCount() == 0 && result.Kind() == component.KindU64 {
				return nil
			}
		} else {
			_, when := f.Parameter(0)
			if f.ParameterCount() == 1 && when.Kind() == component.KindU64 && result.Kind() == component.KindOwn && result.IsResource(m.poller.resource) {
				return nil
			}
		}
		return vm.ErrType
	}
}

// Imports includes the poll interface and its identical nominal resource type.
// An embedder may merge these Instances with other explicit WASI providers.
func (m *MonotonicClock) Imports() component.Imports {
	if m == nil || m.poller == nil || m.poller.resource == nil {
		return component.Imports{}
	}
	out := m.poller.Imports()
	functions := map[string]component.HostFunction{
		"now": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := m.poller.check(c.Context()); err != nil {
				return nil, err
			}
			now, err := m.now()
			if err != nil {
				return nil, err
			}
			return &component.Value{Bits: now}, m.poller.check(c.Context())
		},
		"resolution": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			return &component.Value{Bits: 1000}, m.poller.check(c.Context())
		},
		"subscribe-instant": func(c *component.Caller, args []component.Value) (*component.Value, error) {
			return m.subscribe(c, args[0].Bits, false)
		},
		"subscribe-duration": func(c *component.Caller, args []component.Value) (*component.Value, error) {
			return m.subscribe(c, args[0].Bits, true)
		},
	}
	checks := map[string]component.FunctionCheck{}
	for name := range functions {
		checks[name] = m.contract(name)
	}
	out.Instances[MonotonicClockInterface] = component.Imports{Resources: map[string]*component.HostResourceType{"pollable": m.poller.resource}, Functions: functions, FunctionChecks: checks}
	return out
}
