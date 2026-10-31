package preview3

import (
	"context"
	"sync"
	"time"

	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/vm"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

const (
	ClocksTypesInterface    = "wasi:clocks/types@0.3.2"
	MonotonicClockInterface = "wasi:clocks/monotonic-clock@0.3.2"
	SystemClockInterface    = "wasi:clocks/system-clock@0.3.2"
)

// Instant preserves the full signed seconds range independently of time.Duration.
type Instant struct {
	Seconds     int64
	Nanoseconds uint32
}
type SystemClockConfig struct {
	Now        func(context.Context) (Instant, error)
	Resolution func(context.Context) (uint64, error)
}
type SystemClock struct{ config SystemClockConfig }

func NewSystemClock(config SystemClockConfig) (*SystemClock, error) {
	if config.Now == nil && config.Resolution == nil {
		config.Now = func(ctx context.Context) (Instant, error) {
			t := time.Now()
			return Instant{t.Unix(), uint32(t.Nanosecond() / 1000 * 1000)}, ctx.Err()
		}
		config.Resolution = func(ctx context.Context) (uint64, error) { return 1000, ctx.Err() }
	}
	if config.Now == nil || config.Resolution == nil {
		return nil, wasm.ErrInvalid
	}
	return &SystemClock{config}, nil
}
func (s *SystemClock) Imports() component.Imports {
	if s == nil || s.config.Now == nil || s.config.Resolution == nil {
		return component.Imports{}
	}
	functions := map[string]component.HostFunction{
		"now": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := c.Context().Err(); err != nil {
				return nil, err
			}
			v, err := s.config.Now(c.Context())
			if err != nil {
				return nil, err
			}
			if v.Nanoseconds >= 1e9 {
				return nil, wasm.ErrInvalid
			}
			return &component.Value{Elements: []component.Value{{Bits: uint64(v.Seconds)}, {Bits: uint64(v.Nanoseconds)}}}, c.Context().Err()
		},
		"get-resolution": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := c.Context().Err(); err != nil {
				return nil, err
			}
			n, err := s.config.Resolution(c.Context())
			if err != nil {
				return nil, err
			}
			if n == 0 {
				return nil, wasm.ErrInvalid
			}
			return &component.Value{Bits: n}, c.Context().Err()
		},
	}
	checks := map[string]component.FunctionCheck{}
	for _, name := range []string{"now", "get-resolution"} {
		checks[name] = func(f component.FunctionInfo) error {
			if f.Async() || f.ParameterCount() != 0 {
				return vm.ErrType
			}
			t := f.Result()
			if name == "get-resolution" {
				if t.Kind() == component.KindU64 {
					return nil
				}
				return vm.ErrType
			}
			sec, a := t.Field(0)
			nano, b := t.Field(1)
			if t.Kind() != component.KindRecord || t.FieldCount() != 2 || sec != "seconds" || a.Kind() != component.KindS64 || nano != "nanoseconds" || b.Kind() != component.KindU32 {
				return vm.ErrType
			}
			return nil
		}
	}
	return component.Imports{Instances: map[string]component.Imports{ClocksTypesInterface: {}, SystemClockInterface: {Functions: functions, FunctionChecks: checks}}}
}

// MonotonicClock owns bounded timer workers, with microsecond ticks and a private
// origin. Close cancels pending operations; WaitClosed joins every timer worker.
// MaxWaits limits outstanding waits across all Stores sharing this provider.
type MonotonicClockConfig struct{ MaxWaits uint32 }
type MonotonicClock struct {
	mu                       sync.Mutex
	origin                   time.Time
	maximum, active, pending uint32
	closed                   bool
	stopping, done           chan struct{}
}

func NewMonotonicClock(config MonotonicClockConfig) (*MonotonicClock, error) {
	if config.MaxWaits == 0 {
		config.MaxWaits = 1024
	}
	return &MonotonicClock{origin: time.Now(), maximum: config.MaxWaits, stopping: make(chan struct{}), done: make(chan struct{})}, nil
}
func (m *MonotonicClock) now() (uint64, error) {
	n := time.Since(m.origin)
	if n < 0 {
		return 0, wasm.ErrLimit
	}
	return uint64(n) / 1000 * 1000, nil
}
func (m *MonotonicClock) check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return wasm.ErrClosed
	}
	return nil
}
func (m *MonotonicClock) Close() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		close(m.stopping)
		if m.active == 0 {
			close(m.done)
		}
	}
}
func (m *MonotonicClock) WaitClosed(ctx context.Context) error {
	if m == nil {
		return vm.ErrReference
	}
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-m.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (m *MonotonicClock) begin() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return wasm.ErrClosed
	}
	if m.pending >= m.maximum {
		return wasm.ErrLimit
	}
	m.active++
	m.pending++
	return nil
}
func (m *MonotonicClock) finish() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active--
	if m.closed && m.active == 0 {
		close(m.done)
	}
}
func (m *MonotonicClock) wait(c *component.Caller, when uint64, relative bool) (*component.HostCompletion, error) {
	if err := m.check(c.Context()); err != nil {
		return nil, err
	}
	never := false
	if relative {
		now, err := m.now()
		if err != nil {
			return nil, err
		}
		if when > ^uint64(0)-now {
			never = true
		} else {
			when += now
		}
	}
	if err := m.begin(); err != nil {
		return nil, err
	}
	p := component.NewHostCompletion()
	ctx := c.Context()
	go func() {
		defer m.finish()
		var err error
		for {
			now, failed := m.now()
			if failed != nil {
				err = failed
				break
			}
			if !never && now >= when {
				break
			}
			delta := uint64(1<<63 - 1)
			if !never {
				delta = min(delta, when-now)
			}
			delta = max(delta, 1000)
			timer := time.NewTimer(time.Duration(delta))
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-m.stopping:
				err = wasm.ErrClosed
			case <-timer.C:
			}
			timer.Stop()
			if err != nil {
				break
			}
		}
		m.mu.Lock()
		m.pending--
		m.mu.Unlock()
		if ctx.Err() != nil {
			p.Cancelled()
			return
		}
		p.Resolve(func(*component.Caller) (*component.Value, error) { return nil, err })
	}()
	return p, nil
}
func (m *MonotonicClock) Imports() component.Imports {
	if m == nil || m.done == nil {
		return component.Imports{}
	}
	functions := map[string]component.HostFunction{
		"now": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			if err := m.check(c.Context()); err != nil {
				return nil, err
			}
			n, err := m.now()
			return &component.Value{Bits: n}, err
		},
		"get-resolution": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
			return &component.Value{Bits: 1000}, m.check(c.Context())
		},
	}
	async := map[string]component.AsyncHostFunction{}
	checks := map[string]component.FunctionCheck{}
	for _, name := range []string{"now", "get-resolution", "wait-for", "wait-until"} {
		waiting := name == "wait-for" || name == "wait-until"
		checks[name] = func(f component.FunctionInfo) error {
			if f.Async() != waiting {
				return vm.ErrType
			}
			if waiting {
				_, p := f.Parameter(0)
				if f.ParameterCount() != 1 || p.Kind() != component.KindU64 || f.Result().Valid() {
					return vm.ErrType
				}
			} else if f.ParameterCount() != 0 || f.Result().Kind() != component.KindU64 {
				return vm.ErrType
			}
			return nil
		}
		if waiting {
			async[name] = func(c *component.Caller, args []component.Value) (*component.HostCompletion, error) {
				return m.wait(c, args[0].Bits, name == "wait-for")
			}
		}
	}
	return component.Imports{Instances: map[string]component.Imports{ClocksTypesInterface: {}, MonotonicClockInterface: {Functions: functions, AsyncCompletions: async, FunctionChecks: checks}}}
}
