package preview2

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cpunion/wasmvm/internal/platform/native"
	"github.com/cpunion/wasmvm/pkg/component"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type testPollSource struct {
	mu      sync.Mutex
	ready   bool
	changed chan struct{}
}

func (s *testPollSource) Readiness() (bool, <-chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ready, s.changed
}
func (s *testPollSource) set(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = ready
	close(s.changed)
	s.changed = make(chan struct{})
}

func TestPollClockBothBackends(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		t.Run(backend, func(t *testing.T) {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			e, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			m, err := e.Compile(nil, fixture(t, "poll-clock"))
			if err != nil {
				t.Fatal(err)
			}
			defer m.Close()
			p, err := NewPoller(PollConfig{MaxPollables: 8, MaxPollItems: 8})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			clock, err := NewMonotonicClock(p)
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, clock.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			call := func(name string, args ...component.Value) *component.Value {
				t.Helper()
				f, err := in.Function(name)
				if err != nil {
					t.Fatal(err)
				}
				v, err := f.Call(nil, args, component.CallOptions{Metered: true, Fuel: 100000})
				if err != nil {
					t.Fatal(name, err)
				}
				return v.Value
			}
			if result := call("churn", component.Value{Bits: 128}); result.Bits != 128 {
				t.Fatal(result)
			}
			if call("resolution").Bits != 1000 {
				t.Fatal("resolution")
			}
			if result := call("probe"); result.Bits != 2 {
				t.Fatal(result)
			}
			before := call("now").Bits
			if after := call("now").Bits; after < before || after%1000 != 0 {
				t.Fatal(before, after)
			}
			past := *call("instant", component.Value{})
			if call("ready", past).Bits != 1 {
				t.Fatal("past deadline is not ready")
			}
			call("block", past)
			future := *call("instant", component.Value{Bits: ^uint64(0)})
			if call("ready", future).Bits != 0 {
				t.Fatal("u64 instant overflowed")
			}
			overflow := *call("duration", component.Value{Bits: ^uint64(0)})
			if call("ready", overflow).Bits != 0 {
				t.Fatal("u64 duration overflowed")
			}
			indices := call("poll", component.Value{Elements: []component.Value{future, past, past}})
			if len(indices.Elements) != 2 || indices.Elements[0].Bits != 1 || indices.Elements[1].Bits != 2 {
				t.Fatal(indices)
			}
			started := time.Now()
			delayed := *call("duration", component.Value{Bits: uint64(2 * time.Millisecond)})
			call("block", delayed)
			if time.Since(started) < 2*time.Millisecond {
				t.Fatal("duration resolved early", time.Since(started))
			}
			if call("ready", delayed).Bits != 1 {
				t.Fatal("elapsed duration not ready", time.Since(started))
			}
			zero := *call("duration", component.Value{})
			if call("ready", zero).Bits != 1 {
				t.Fatal("zero duration not ready")
			}
			for _, value := range []component.Value{past, future, overflow, delayed, zero} {
				if _, err := in.DropResource(nil, value.Resource, component.CallOptions{}); err != nil {
					t.Fatal(err)
				}
			}
			p.mu.Lock()
			live := len(p.entries)
			p.mu.Unlock()
			if live != 0 {
				t.Fatal("drop leaked subscriptions", live)
			}
		})
	}
}

func TestPollSnapshotsNotificationsAndCancellation(t *testing.T) {
	p, err := NewPoller(PollConfig{MaxPollItems: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	a := &testPollSource{changed: make(chan struct{})}
	b := &testPollSource{changed: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	type result struct {
		values []component.Value
		err    error
	}
	done := make(chan result, 1)
	go func() { v, err := p.wait(ctx, []PollSource{a, b, b}); done <- result{v, err} }()
	b.set(true)
	got := <-done
	if got.err != nil || !reflect.DeepEqual(got.values, []component.Value{{Bits: 1}, {Bits: 2}}) {
		t.Fatal(got)
	}
	b.set(false)
	cancelCtx, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	if _, err := p.wait(cancelCtx, []PollSource{b}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := p.wait(ctx, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	if _, err := p.wait(ctx, []PollSource{a, a, a, a, a}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	go func() { v, err := p.wait(ctx, []PollSource{a}); done <- result{v, err} }()
	p.Close()
	p.Close()
	got = <-done
	if !errors.Is(got.err, wasm.ErrClosed) {
		t.Fatal(got)
	}
}

func TestPollProviderBoundsAndBadBindings(t *testing.T) {
	if _, err := NewPoller(PollConfig{MaxPollItems: 65535}); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	for _, p := range []*Poller{nil, {}} {
		p.Close()
		if len(p.Imports().Instances) != 0 {
			t.Fatal("empty poller exports imports")
		}
		if _, err := p.NewPollable(nil, nil); err == nil {
			t.Fatal("uninitialized poller")
		}
		if _, err := NewMonotonicClock(p); err == nil {
			t.Fatal("uninitialized clock")
		}
	}
	if len((*MonotonicClock)(nil).Imports().Instances) != 0 || len((&MonotonicClock{}).Imports().Instances) != 0 {
		t.Fatal("uninitialized clock exports imports")
	}
	p, err := NewPoller(PollConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	source := &testPollSource{changed: make(chan struct{})}
	if _, err := p.NewPollable(nil, source); err == nil {
		t.Fatal("nil Caller")
	}
	released := 0
	if _, err := p.publish(nil, nil, func() { released++ }); err == nil || released != 1 {
		t.Fatal(err, released)
	}
	p.mu.Lock()
	p.next = 1 << 32
	p.mu.Unlock()
	if _, err := p.publish(nil, source, nil); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
	p.mu.Lock()
	live := len(p.entries)
	p.mu.Unlock()
	if live != 0 {
		t.Fatal("failed publication leaked", live)
	}
	e, err := component.NewEngine(component.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	for _, name := range []string{"wrong-ready", "wrong-block", "wrong-own", "wrong-poll-result"} {
		m, err := e.Compile(nil, fixture(t, "poll-"+name))
		if err != nil {
			t.Fatal(name, err)
		}
		if in, err := m.Instantiate(nil, p.Imports(), component.CallOptions{}); in != nil || err == nil {
			t.Fatal(name, in, err)
		}
		m.Close()
	}
	m, err := e.Compile(nil, fixture(t, "poll-clock"))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	p2, err := NewPoller(PollConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()
	clock, err := NewMonotonicClock(p2)
	if err != nil {
		t.Fatal(err)
	}
	imports := clock.Imports()
	imports.Instances[PollInterface] = p.Imports().Instances[PollInterface]
	if in, err := m.Instantiate(nil, imports, component.CallOptions{}); in != nil || err == nil {
		t.Fatal("foreign nominal pollable", in, err)
	}
	for _, name := range []string{"wrong-now", "wrong-width", "wrong-when", "wrong-resource"} {
		bad, err := e.Compile(nil, fixture(t, "monotonic-"+name))
		if err != nil {
			t.Fatal(name, err)
		}
		if in, err := bad.Instantiate(nil, clock.Imports(), component.CallOptions{}); in != nil || err == nil {
			t.Fatal(name, in, err)
		}
		bad.Close()
	}
	p.Close()
	if _, err := NewMonotonicClock(p); !errors.Is(err, wasm.ErrClosed) {
		t.Fatal(err)
	}
	if _, err := p.publish(nil, source, func() { released++ }); !errors.Is(err, wasm.ErrClosed) || released != 2 {
		t.Fatal(err, released)
	}
}

func TestPollPendingStoreAndProviderClose(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, operation := range []string{"block", "poll"} {
			for _, closeStore := range []bool{false, true} {
				config := component.DefaultConfig()
				config.Core.Backend = backend
				e, err := component.NewEngine(config)
				if err != nil {
					t.Fatal(err)
				}
				m, err := e.Compile(nil, fixture(t, "poll-source"))
				if err != nil {
					t.Fatal(err)
				}
				p, err := NewPoller(PollConfig{})
				if err != nil {
					t.Fatal(err)
				}
				clock, err := NewMonotonicClock(p)
				if err != nil {
					t.Fatal(err)
				}
				entered := make(chan struct{}, 1)
				source := pollNotifySource{source: &testPollSource{changed: make(chan struct{})}, entered: entered}
				imports := clock.Imports()
				imports.Functions = map[string]component.HostFunction{"create": func(c *component.Caller, _ []component.Value) (*component.Value, error) {
					value, err := p.NewPollable(c, source)
					return &value, err
				}}
				in, err := m.Instantiate(nil, imports, component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				create, err := in.Function("create")
				if err != nil {
					t.Fatal(err)
				}
				value, err := create.Call(nil, nil, component.CallOptions{})
				if err != nil {
					t.Fatal(err)
				}
				f, err := in.Function(operation)
				if err != nil {
					t.Fatal(err)
				}
				argument := *value.Value
				if operation == "poll" {
					argument = component.Value{Elements: []component.Value{argument}}
				}
				done := make(chan error, 1)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				go func() { _, err := f.Call(ctx, []component.Value{argument}, component.CallOptions{}); done <- err }()
				select {
				case <-entered:
				case <-ctx.Done():
					t.Fatal("poll not entered")
				}
				want := wasm.ErrClosed
				if closeStore {
					in.Close()
					want = context.Canceled
				} else {
					p.Close()
				}
				select {
				case err := <-done:
					if !errors.Is(err, want) {
						t.Fatal(operation, closeStore, err)
					}
				case <-ctx.Done():
					t.Fatal("poll did not leave")
				}
				cancel()
				in.Close()
				p.Close()
				m.Close()
				e.Close()
				p.mu.Lock()
				remaining := len(p.entries)
				p.mu.Unlock()
				if remaining != 0 {
					t.Fatal("owner Close leaked entries", remaining)
				}
			}
		}
	}
}

type pollNotifySource struct {
	source  PollSource
	entered chan<- struct{}
}

func (s pollNotifySource) Readiness() (bool, <-chan struct{}) {
	select {
	case s.entered <- struct{}{}:
	default:
	}
	return s.source.Readiness()
}

func TestPollQuotasAndEmptyInput(t *testing.T) {
	for _, backend := range []string{"interpreter", "native"} {
		if backend == "native" && !native.Available {
			continue
		}
		for _, kind := range []string{"empty", "items", "entries"} {
			config := component.DefaultConfig()
			config.Core.Backend = backend
			e, err := component.NewEngine(config)
			if err != nil {
				t.Fatal(err)
			}
			m, err := e.Compile(nil, fixture(t, "poll-clock"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := NewPoller(PollConfig{MaxPollables: 2, MaxPollItems: 1})
			if err != nil {
				t.Fatal(err)
			}
			clock, err := NewMonotonicClock(p)
			if err != nil {
				t.Fatal(err)
			}
			in, err := m.Instantiate(nil, clock.Imports(), component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			f, err := in.Function("instant")
			if err != nil {
				t.Fatal(err)
			}
			value, err := f.Call(nil, []component.Value{{}}, component.CallOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var expected error
			if kind == "entries" {
				if _, err = f.Call(nil, []component.Value{{}}, component.CallOptions{}); err != nil {
					t.Fatal(err)
				}
				_, expected = f.Call(nil, []component.Value{{}}, component.CallOptions{})
			} else {
				f, err = in.Function("poll")
				if err != nil {
					t.Fatal(err)
				}
				list := component.Value{}
				if kind == "items" {
					list.Elements = []component.Value{*value.Value, *value.Value}
				}
				_, expected = f.Call(nil, []component.Value{list}, component.CallOptions{})
			}
			if !errors.Is(expected, wasm.ErrLimit) {
				t.Fatal(kind, expected)
			}
			in.Close()
			p.Close()
			m.Close()
			e.Close()
		}
	}
}

func TestMonotonicSubscriptionLifecycleAndClockGuard(t *testing.T) {
	p, err := NewPoller(PollConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	clock, err := NewMonotonicClock(p)
	if err != nil {
		t.Fatal(err)
	}
	s := &clockSubscription{clock: clock, when: 0, ready: make(chan struct{})}
	s.mu.Lock()
	s.armLocked()
	s.armLocked()
	s.mu.Unlock()
	if ready, _ := s.Readiness(); !ready {
		t.Fatal("past subscription")
	}
	s.close()
	s.close()
	if ready, _ := s.Readiness(); !ready {
		t.Fatal("closed subscription")
	}
	s = &clockSubscription{clock: clock, when: ^uint64(0), ready: make(chan struct{})}
	s.mu.Lock()
	s.armLocked()
	s.mu.Unlock()
	if ready, _ := s.Readiness(); ready {
		t.Fatal("large deadline became ready")
	}
	s.close()
	clock.origin = time.Now().Add(time.Hour)
	if _, err := clock.now(); !errors.Is(err, wasm.ErrLimit) {
		t.Fatal(err)
	}
}

func TestPollCloseJoinsDispatchedTimers(t *testing.T) {
	p, err := NewPoller(PollConfig{MaxPollables: 128})
	if err != nil {
		t.Fatal(err)
	}
	clock, err := NewMonotonicClock(p)
	if err != nil {
		t.Fatal(err)
	}
	subscriptions := make([]*clockSubscription, 128)
	for i := range subscriptions {
		now, err := clock.rawNow()
		if err != nil {
			t.Fatal(err)
		}
		s := &clockSubscription{clock: clock, when: now + 1000, ready: make(chan struct{})}
		s.mu.Lock()
		s.armLocked()
		s.mu.Unlock()
		p.entries[uint32(i+1)] = pollEntry{s, s.close}
		subscriptions[i] = s
	}
	var joined sync.WaitGroup
	for i := 0; i < 4; i++ {
		joined.Add(1)
		go func() { defer joined.Done(); p.Close() }()
	}
	joined.Wait()
	for _, s := range subscriptions {
		s.mu.Lock()
		if !s.closed || s.timer != nil {
			t.Error("timer retained after concurrent Close")
		}
		s.mu.Unlock()
		s.close()
	}
}

func TestPollCloseWaitsForProvisionalAndDetachedCleanup(t *testing.T) {
	for _, provisional := range []bool{false, true} {
		p, err := NewPoller(PollConfig{})
		if err != nil {
			t.Fatal(err)
		}
		cleanupStarted, resume := make(chan struct{}), make(chan struct{})
		release := func() { close(cleanupStarted); <-resume }
		cleanupDone := make(chan struct{})
		if provisional {
			if err := p.begin(); err != nil {
				t.Fatal(err)
			}
		} else {
			p.entries[1] = pollEntry{release: release}
			go func() { p.remove(1); close(cleanupDone) }()
			<-cleanupStarted
		}
		closed := make(chan struct{})
		go func() { p.Close(); close(closed) }()
		<-p.done
		select {
		case <-closed:
			t.Fatal("provider did not join pending cleanup")
		default:
		}
		if provisional {
			go func() {
				_, err := p.publish(nil, &testPollSource{}, release)
				if !errors.Is(err, wasm.ErrClosed) {
					t.Error(err)
				}
				p.active.Done()
				close(cleanupDone)
			}()
			<-cleanupStarted
		}
		close(resume)
		<-cleanupDone
		select {
		case <-closed:
		case <-time.After(time.Second):
			t.Fatal("provider cleanup did not finish")
		}
		if err := p.begin(); !errors.Is(err, wasm.ErrClosed) {
			t.Fatal(err)
		}
	}
}
