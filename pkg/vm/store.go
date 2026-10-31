package vm

import (
	"context"
	"errors"
	"sync"

	"github.com/cpunion/wasmvm/internal/compiler/slots"
	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/nativeexec"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

var (
	ErrBusy      = errors.New("WebAssembly Store is already executing")
	ErrType      = errors.New("WebAssembly value or export type mismatch")
	ErrReference = state.ErrReference
	ErrLink      = state.ErrLink
)

type StoreLimits = state.Limits
type ExecutionLimits = execute.Limits
type StoreConfig struct {
	Limits                                  StoreLimits
	Execution                               ExecutionLimits
	MaxExternRefs, MaxHostRoots, MaxReentry uint32
	MaxContinuations                        uint32
	CollectEveryAllocation                  bool
}

func DefaultStoreConfig() StoreConfig {
	return StoreConfig{Limits: state.DefaultLimits(), Execution: execute.DefaultLimits(), MaxExternRefs: 1 << 20, MaxHostRoots: 1 << 20, MaxReentry: 32, MaxContinuations: 1024}
}

type CallOptions struct {
	Metered bool
	Fuel    uint64
}
type CallResult struct {
	Values        []Value
	FuelRemaining uint64
}
type Store struct {
	mu            sync.Mutex
	engine        *Engine
	config        StoreConfig
	state         *state.Store
	closing       bool
	closed        chan struct{}
	active        *operation
	roots         map[*reference]struct{}
	externs       map[uint64]any
	nextExtern    uint64
	nativeModules map[*slots.Module]*nativeexec.Lease
	workspace     *callWorkspace
}
type operation struct {
	store         *Store
	ctx           context.Context
	cancel        context.CancelFunc
	options       CallOptions
	fuel          uint64
	frames        []*execute.Execution
	pending       []*state.Instance
	depth         uint32
	collect       bool
	continuations map[*Continuation]struct{}
}

func (e *Engine) NewStore(config StoreConfig) (*Store, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil, wasm.ErrClosed
	}
	core := state.New(config.Limits)
	if e.config.Backend == "native" {
		core = state.NewNative(config.Limits)
	}
	core.Domain = e.domain
	s := &Store{engine: e, config: config, state: core, nativeModules: make(map[*slots.Module]*nativeexec.Lease), closed: make(chan struct{}), roots: make(map[*reference]struct{}), externs: make(map[uint64]any)}
	core.CollectEveryAllocation = config.CollectEveryAllocation
	core.HeapSafepoint = func() { s.mu.Lock(); defer s.mu.Unlock(); s.collectLocked(s.active) }
	return s, nil
}
func resourceLimit(detail string) error { return &wasm.Error{Kind: wasm.ErrLimit, Detail: detail} }
func (s *Store) begin(ctx context.Context, options CallOptions) (*operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return nil, wasm.ErrClosed
	}
	if s.active != nil {
		return nil, ErrBusy
	}
	ctx = contextOrBackground(ctx)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	op := &operation{store: s, ctx: ctx, cancel: cancel, options: options, fuel: options.Fuel}
	s.active = op
	return op, nil
}
func (op *operation) finish() {
	s := op.store
	op.cancel()
	for call := range op.continuations {
		call.mu.Lock()
		call.abort()
		call.mu.Unlock()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = nil
	if s.closing {
		s.closeLocked()
	} else if op.collect {
		s.collectLocked(nil)
	}
}

// Close requests cancellation and returns without waiting for user callbacks.
// Final reclamation happens once the active operation and callbacks leave.
func (s *Store) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return
	}
	s.closing = true
	if s.active != nil {
		s.active.cancel()
	} else {
		s.closeLocked()
	}
}
func (s *Store) closeLocked() {
	if s.workspace != nil {
		s.workspace.execution.Close()
		s.workspace = nil
	}
	s.state.Close()
	for _, lease := range s.nativeModules {
		lease.Close()
	}
	s.nativeModules = nil
	for r := range s.roots {
		r.released = true
	}
	// Retaining the closed Store must not retain the reference maps' capacity.
	s.roots = nil
	s.externs = nil
	close(s.closed)
}

// WaitClosed waits for final reclamation. Do not call it from a callback of
// this Store: that callback itself holds the lease which must first leave.
func (s *Store) WaitClosed(ctx context.Context) error {
	ctx = contextOrBackground(ctx)
	select {
	case <-s.closed:
		return nil
	default:
	}
	select {
	case <-s.closed:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type Usage struct {
	Core                  state.Usage
	ExternRefs, HostRoots uint64
}

func (s *Store) Usage() (Usage, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Usage{}, err
	}
	defer op.finish()
	s.mu.Lock()
	defer s.mu.Unlock()
	return Usage{s.state.Usage(), uint64(len(s.externs)), uint64(len(s.roots))}, nil
}
func (s *Store) Collect() error {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return err
	}
	defer op.finish()
	s.mu.Lock()
	s.collectLocked(op)
	s.mu.Unlock()
	return nil
}
func (s *Store) collectLocked(op *operation) {
	var roots []state.ReferenceValue
	var active []*state.Instance
	marked := make(map[uint64]bool)
	visit := func(t wasm.ValueType, v value.Value) {
		if t.IsReference() && v.Lo != 0 {
			roots = append(roots, state.ReferenceValue{Type: t, Value: v})
		}
	}

	for r := range s.roots {
		visit(r.kind, value.Value{Lo: r.id})
	}
	if op != nil {
		active = append(active, op.pending...)
		for _, frame := range op.frames {
			frame.RetainedValues(visit)
			frame.RetainedInstances(func(in *state.Instance) { active = append(active, in) })
		}
	}
	s.state.CollectReferences(roots, active, func(id uint64) { marked[id] = true })
	if len(s.nativeModules) != 0 {
		live := make(map[*slots.Module]bool)
		s.state.RetainedModules(func(m *slots.Module) { live[m] = true })
		for m, lease := range s.nativeModules {
			if !live[m] {
				lease.Close()
				delete(s.nativeModules, m)
			}
		}
	}
	for id := range s.externs {
		if !marked[id] {
			delete(s.externs, id)
		}
	}
}
