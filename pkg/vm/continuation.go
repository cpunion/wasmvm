package vm

import (
	"sync"

	"github.com/cpunion/wasmvm/internal/engine/execute"
	"github.com/cpunion/wasmvm/internal/engine/value"
	"github.com/cpunion/wasmvm/pkg/wasm"
)

// HostSuspension parks an explicit guest continuation at a host boundary.
// Ready must be nonblocking and must not enter the VM. Resume receives a fresh
// Caller; the Caller which initially returned Pending has already expired.
type HostSuspension struct {
	Ready  func() bool
	Resume func(*Caller) (HostResult, error)
}

type HostResult struct {
	Values  []Value
	Pending *HostSuspension
}

type SuspendableHostFunction func(*Caller, []Value) (HostResult, error)

func (s *Store) NewSuspendableFunction(signature FunctionType, fn SuspendableHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	op, err := s.begin(nil, CallOptions{})
	if err != nil {
		return Extern{}, err
	}
	defer op.finish()
	if fn == nil {
		return Extern{}, ErrType
	}
	return s.newHostFunction(signature, fn, referenceType...)
}

// Continuation belongs to one WithCaller session. Suspended native/interpreter
// frames retain their roots and budgets until Poll completes, Abort runs, or
// the session ends. No Caller or host Go stack is retained as a continuation.
type Continuation struct {
	mu        sync.Mutex
	op        *operation
	work      *callWorkspace
	result    []wasm.ValueType
	pending   *HostSuspension
	args      []Value
	temporary []Value
	running   bool
	closed    bool
}

func (c *Caller) NewSuspendableFunction(signature FunctionType, fn SuspendableHostFunction, referenceType ...wasm.ValueType) (Extern, error) {
	if err := c.enter(); err != nil {
		return Extern{}, err
	}
	defer c.mu.Unlock()
	if fn == nil {
		return Extern{}, ErrType
	}
	return c.op.store.newHostFunction(signature, fn, referenceType...)
}

func (op *operation) removeFrame(e *execute.Execution) {
	for i, frame := range op.frames {
		if frame == e {
			last := len(op.frames) - 1
			op.frames[i], op.frames[last] = op.frames[last], nil
			op.frames = op.frames[:last]
			return
		}
	}
}

func (op *operation) executionLimits(except *execute.Execution) (execute.Limits, error) {
	limits := op.store.config.Execution
	for _, frame := range op.frames {
		if frame == except {
			continue
		}
		used := frame.Footprint()
		if used.MaxFrames > limits.MaxFrames || used.MaxSlots > limits.MaxSlots {
			return limits, resourceLimit("suspended frames")
		}
		limits.MaxFrames -= used.MaxFrames
		limits.MaxSlots -= used.MaxSlots
	}
	return limits, nil
}

func (c *Caller) Start(fn Extern, args []Value) (*Continuation, error) {
	if err := c.enter(); err != nil {
		return nil, err
	}
	defer c.mu.Unlock()
	op, s := c.op, c.op.store
	if uint64(len(op.continuations)) >= uint64(s.config.MaxContinuations) {
		return nil, resourceLimit("continuations")
	}
	if fn.store != s {
		return nil, ErrReference
	}
	if fn.Kind() != FunctionKind {
		return nil, ErrType
	}
	f, ok := s.state.Function(fn.external.Function)
	if !ok {
		return nil, ErrReference
	}
	s.mu.Lock()
	raw, err := s.decodeValuesLocked(args, f.Signature.Params)
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	limits, err := op.executionLimits(nil)
	if err != nil {
		return nil, err
	}
	op.depth++
	w, err := op.acquireWorkspace(s.config.Execution)
	op.depth--
	if err != nil {
		return nil, err
	}
	call := &Continuation{op: op, work: w, result: f.Signature.Results}
	if err := w.execution.SetLimits(limits); err != nil {
		w.execution.Close()
		return nil, err
	}
	if err := w.execution.Start(fn.external.Function, raw); err != nil {
		w.execution.Close()
		return nil, err
	}
	op.frames = append(op.frames, w.execution)
	if op.continuations == nil {
		op.continuations = map[*Continuation]struct{}{}
	}
	op.continuations[call] = struct{}{}
	return call, nil
}

func (call *Continuation) Ready() bool {
	if call == nil {
		return false
	}
	call.mu.Lock()
	defer call.mu.Unlock()
	return !call.closed && !call.running && (call.pending == nil || call.pending.Ready())
}

func (call *Continuation) abort() {
	if call.closed {
		return
	}
	call.closed = true
	call.work.execution.Abort(wasm.ErrClosed)
	call.op.removeFrame(call.work.execution)
	delete(call.op.continuations, call)
	for _, v := range call.args {
		v.Release()
	}
	for _, v := range call.temporary {
		v.Release()
	}
	call.temporary = nil
	call.args = nil
	call.pending = nil
	call.op.releaseWorkspace(call.work)
	call.work = nil
}

// Abort releases a parked call even after cancellation. It does not execute
// guest code and cannot interrupt a concurrently running Poll.
func (c *Caller) Abort(call *Continuation) error {
	if !c.mu.TryLock() {
		return ErrBusy
	}
	defer c.mu.Unlock()
	if !c.active {
		return wasm.ErrClosed
	}
	if call == nil || call.op != c.op {
		return ErrReference
	}
	call.mu.Lock()
	defer call.mu.Unlock()
	if call.running {
		return ErrBusy
	}
	call.abort()
	return nil
}

func (c *Caller) Poll(call *Continuation) (result CallResult, done bool, err error) {
	if err := c.enter(); err != nil {
		return result, false, err
	}
	defer c.mu.Unlock()
	if call == nil || call.op != c.op {
		return result, false, ErrReference
	}
	call.mu.Lock()
	if call.running {
		call.mu.Unlock()
		return result, false, ErrBusy
	}
	if call.closed {
		call.mu.Unlock()
		return result, false, wasm.ErrClosed
	}
	call.running = true
	call.mu.Unlock()
	complete := false
	defer func() {
		call.mu.Lock()
		call.running = false
		if !complete {
			call.abort()
		}
		call.mu.Unlock()
	}()
	op, s := c.op, c.op.store
	if op.depth > s.config.MaxReentry {
		return result, false, resourceLimit("host reentry depth")
	}
	op.depth++
	defer func() { op.depth-- }()
	e := call.work.execution
	for {
		limits, failed := op.executionLimits(e)
		if failed != nil {
			return result, false, failed
		}
		if failed = e.SetLimits(limits); failed != nil {
			return result, false, failed
		}
		if err = op.ctx.Err(); err != nil {
			return result, false, err
		}
		if call.pending != nil {
			if !call.pending.Ready() {
				complete = true
				return result, false, nil
			}
			var step HostResult
			step, err = call.resumeHost(nil, call.pending.Resume)
			if step.Pending != nil && err == nil {
				call.pending = step.Pending
				complete = true
				return result, false, nil
			}
			call.pending = nil
			if err = call.resumeValues(step.Values, err); err != nil {
				return result, false, err
			}
		}
		e.SetFuel(op.fuel)
		switch e.Run() {
		case execute.Done:
			op.fuel = e.Fuel()
			s.mu.Lock()
			result.Values, err = s.encodeValuesLocked(e.Results(), call.result)
			s.mu.Unlock()
			result.FuelRemaining = op.fuel
			if err != nil {
				return result, false, err
			}
			call.mu.Lock()
			call.abort()
			call.mu.Unlock()
			complete = true
			return result, true, nil
		case execute.Failed:
			op.fuel = e.Fuel()
			return result, false, op.exceptionError(e.Err())
		case execute.HostRequest:
			op.fuel = e.Fuel()
			request := e.Host()
			if host, ok := request.Function.Host.(SuspendableHostFunction); ok {
				var step HostResult
				step, err = call.resumeHost(request, func(current *Caller) (HostResult, error) { return host(current, call.args) })
				if step.Pending != nil && err == nil {
					call.pending = step.Pending
					complete = true
					return result, false, nil
				}
				if err = call.resumeValues(step.Values, err); err != nil {
					return result, false, err
				}
			} else {
				var values []value.Value
				values, err = op.host(request, request.Caller)
				e.SetFuel(op.fuel)
				if err = e.ResumeHost(values, err); err != nil {
					return result, false, op.exceptionError(err)
				}
			}
		default:
			return result, false, ErrType
		}
	}
}

func (call *Continuation) resumeHost(request *execute.HostCall, fn func(*Caller) (HostResult, error)) (step HostResult, err error) {
	op, s := call.op, call.op.store
	if request != nil {
		s.mu.Lock()
		call.args, err = s.encodeValuesLocked(request.Args, request.Function.Signature.Params)
		s.mu.Unlock()
		if err != nil {
			return step, err
		}
	}
	in := call.work.execution.Host().Caller
	caller := &Caller{active: true, op: op, instance: in}
	panicked := true
	defer func() {
		if panicked {
			op.cancel()
		}
		caller.mu.Lock()
		caller.active = false
		caller.mu.Unlock()
		call.temporary = append(call.temporary, caller.temporary...)
	}()
	step, err = fn(caller)
	panicked = false
	if step.Pending != nil && (step.Pending.Ready == nil || step.Pending.Resume == nil || len(step.Values) != 0) {
		return HostResult{}, ErrType
	}
	return step, err
}

func (call *Continuation) resumeValues(values []Value, failed error) error {
	op, s := call.op, call.op.store
	var raw []value.Value
	err := failed
	if err == nil {
		s.mu.Lock()
		raw, err = s.decodeValuesLocked(values, call.work.execution.Host().Function.Signature.Results)
		s.mu.Unlock()
	} else {
		err = op.hostError(err)
	}
	defer func() {
		for _, v := range call.args {
			v.Release()
		}
		call.args = nil
		for _, v := range call.temporary {
			v.Release()
		}
		call.temporary = nil
	}()
	if err == nil {
		err = op.ctx.Err()
	}
	call.work.execution.SetFuel(op.fuel)
	if failed := call.work.execution.ResumeHost(raw, err); failed != nil {
		return op.exceptionError(failed)
	}
	return nil
}
