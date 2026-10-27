package vm

import (
	"context"
	"sync"

	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/nativeexec"
	"github.com/cpunion/wasmvm/internal/engine/state"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

type HostFunction func(caller *Caller, args []Value) ([]Value, error)

// Caller is a capability for the duration of one synchronous host callback.
// Reentry shares fuel and frame quotas with its suspended parent. Concurrent
// use returns ErrBusy; retaining Caller past the callback returns ErrClosed.
type Caller struct {
	mu        sync.Mutex
	active    bool
	op        *operation
	instance  *state.Instance
	temporary []Value
}

func (c *Caller) enter() error {
	if !c.mu.TryLock() {
		return ErrBusy
	}
	if !c.active {
		c.mu.Unlock()
		return wasm.ErrClosed
	}
	if err := c.op.ctx.Err(); err != nil {
		c.mu.Unlock()
		return err
	}
	return nil
}
func (c *Caller) Context() context.Context { return c.op.ctx }
func (c *Caller) Store() *Store            { return c.op.store }

// NewExternRef creates a callback-scoped root, valid through result handoff.
// Use Store.Root to keep a separate root after the callback returns.
func (c *Caller) NewExternRef(object any) (Value, error) {
	if err := c.enter(); err != nil {
		return Value{}, err
	}
	defer c.mu.Unlock()
	v, err := c.op.store.NewExternRef(object)
	if err == nil {
		c.temporary = append(c.temporary, v)
	}
	return v, err
}
func (c *Caller) Call(fn Extern, args []Value) (CallResult, error) {
	if err := c.enter(); err != nil {
		return CallResult{}, err
	}
	defer c.mu.Unlock()
	return c.op.callPublic(fn, args)
}
func (c *Caller) CallRef(fn Value, args []Value) (CallResult, error) {
	if err := c.enter(); err != nil {
		return CallResult{}, err
	}
	defer c.mu.Unlock()
	s := c.op.store
	s.mu.Lock()
	raw, err := s.rawLocked(fn, wasm.FuncRef)
	s.mu.Unlock()
	if err != nil {
		return CallResult{}, err
	}
	return c.op.callPublic(Extern{store: s, external: state.Extern{Owner: s.state, Kind: 0, Function: raw.Lo}}, args)
}
func (c *Caller) Collect() error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	s := c.op.store
	s.mu.Lock()
	s.collectLocked(c.op)
	s.mu.Unlock()
	return nil
}
func (c *Caller) ReadMemory(index uint32, offset uint64, dst []byte) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	if c.instance == nil || uint64(index) >= uint64(len(c.instance.Memories)) {
		return ErrReference
	}
	if !c.instance.Memories[index].Read(offset, dst) {
		return &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return nil
}
func (c *Caller) WriteMemory(index uint32, offset uint64, src []byte) error {
	if err := c.enter(); err != nil {
		return err
	}
	defer c.mu.Unlock()
	if c.instance == nil || uint64(index) >= uint64(len(c.instance.Memories)) {
		return ErrReference
	}
	if !c.instance.Memories[index].Write(offset, src) {
		return &wasm.Trap{Code: wasm.TrapMemoryBounds}
	}
	return nil
}
func (s *Store) Call(ctx context.Context, fn Extern, args []Value, options CallOptions) (CallResult, error) {
	op, err := s.begin(ctx, options)
	if err != nil {
		return CallResult{}, err
	}
	defer op.finish()
	return op.callPublic(fn, args)
}
func (s *Store) CallRef(ctx context.Context, fn Value, args []Value, options CallOptions) (CallResult, error) {
	op, err := s.begin(ctx, options)
	if err != nil {
		return CallResult{}, err
	}
	defer op.finish()
	s.mu.Lock()
	raw, err := s.rawLocked(fn, wasm.FuncRef)
	s.mu.Unlock()
	if err != nil {
		return CallResult{}, err
	}
	return op.callPublic(Extern{store: s, external: state.Extern{Owner: s.state, Kind: 0, Function: raw.Lo}}, args)
}
func (op *operation) callPublic(fn Extern, args []Value) (CallResult, error) {
	s := op.store
	if fn.store != s {
		return CallResult{}, ErrReference
	}
	if fn.external.Kind != 0 {
		return CallResult{}, ErrType
	}
	f, ok := s.state.Function(fn.external.Function)
	if !ok {
		return CallResult{}, ErrReference
	}
	s.mu.Lock()
	raw, err := s.decodeValuesLocked(args, f.Signature.Params)
	s.mu.Unlock()
	if err != nil {
		return CallResult{}, err
	}
	result, err := op.call(fn.external.Function, raw)
	if err != nil {
		return CallResult{FuelRemaining: op.fuel}, err
	}
	s.mu.Lock()
	out, err := s.encodeValuesLocked(result, f.Signature.Results)
	s.mu.Unlock()
	return CallResult{Values: out, FuelRemaining: op.fuel}, err
}
func (s *Store) decodeValuesLocked(values []Value, types []wasm.ValueType) ([]value.Value, error) {
	if len(values) != len(types) {
		return nil, ErrType
	}
	out := make([]value.Value, len(values))
	for i, v := range values {
		raw, err := s.rawLocked(v, types[i])
		if err != nil {
			return nil, err
		}
		out[i] = raw
	}
	return out, nil
}
func (s *Store) encodeValuesLocked(values []value.Value, types []wasm.ValueType) ([]Value, error) {
	out := make([]Value, len(values))
	for i, v := range values {
		root, err := s.rootLocked(types[i], v)
		if err != nil {
			for _, p := range out {
				if p.ref != nil {
					s.releaseLocked(p.ref)
				}
			}
			return nil, err
		}
		out[i] = root
	}
	return out, nil
}
func (op *operation) call(ref uint64, args []value.Value) ([]value.Value, error) {
	if err := op.ctx.Err(); err != nil {
		return nil, err
	}
	if op.depth > op.store.config.MaxReentry {
		return nil, resourceLimit("host reentry depth")
	}
	op.depth++
	defer func() { op.depth-- }()
	limits := op.store.config.Execution
	for _, frame := range op.frames {
		used := frame.Footprint()
		limits.MaxFrames -= used.MaxFrames
		limits.MaxSlots -= used.MaxSlots
	}
	config := execute.Config{Limits: limits, Metered: op.options.Metered, Fuel: op.fuel}
	if s := op.store; s.engine.config.Backend == "native" {
		backend, err := nativeexec.New(op.ctx, s.nativeModules, limits.MaxSlots, limits.MaxFrames, op.options.Metered)
		if err != nil {
			return nil, err
		}
		config.Backend = backend
	}
	e := execute.New(op.store.state, op.ctx, config)
	op.frames = append(op.frames, e)
	defer func() {
		e.Close()
		op.frames[len(op.frames)-1] = nil
		op.frames = op.frames[:len(op.frames)-1]
	}()
	if err := e.Start(ref, args); err != nil {
		return nil, err
	}
	for {
		status := e.Run()
		op.fuel = e.Fuel()
		switch status {
		case execute.Done:
			return e.Results(), nil
		case execute.Failed:
			return nil, e.Err()
		case execute.HostRequest:
			if err := op.ctx.Err(); err != nil {
				return nil, err
			}
			request := e.Host()
			var callerInstance *state.Instance
			e.RetainedInstances(func(in *state.Instance) { callerInstance = in })
			results, err := op.host(request, callerInstance)
			if err == nil {
				err = op.ctx.Err()
			}
			e.SetFuel(op.fuel)
			if err = e.ResumeHost(results, err); err != nil {
				return nil, err
			}
		}
	}
}
func (op *operation) host(request *execute.HostCall, in *state.Instance) ([]value.Value, error) {
	s := op.store
	s.mu.Lock()
	args, err := s.encodeValuesLocked(request.Args, request.Function.Signature.Params)
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	caller := &Caller{active: true, op: op, instance: in}
	panicked := true
	// Invalidate the capability and drain an in-flight capability operation
	// before touching the suspended continuation, including on host panic.
	defer func() {
		if panicked {
			op.cancel()
		}
		caller.mu.Lock()
		caller.active = false
		caller.mu.Unlock()
		for _, v := range args {
			v.Release()
		}
		for _, v := range caller.temporary {
			v.Release()
		}
	}()
	results, err := request.Function.Host.(HostFunction)(caller, args)
	panicked = false
	if err != nil {
		return nil, err
	}
	caller.mu.Lock()
	caller.active = false
	caller.mu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.decodeValuesLocked(results, request.Function.Signature.Results)
}
